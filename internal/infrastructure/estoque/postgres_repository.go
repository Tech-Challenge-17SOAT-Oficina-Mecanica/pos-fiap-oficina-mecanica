package estoque

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/estoque"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
)

const operacaoRegistrarEntrada = "ENTRADA_ESTOQUE"

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

type itemRow struct {
	id, codigo, tipo, unidadeMedida string
	ativo                           bool
	saldoFisico, saldoReservado     float64
}

type pedidoItemRow struct {
	id                                                        string
	quantidadePedida, quantidadeRecebida, quantidadeReservada float64
}

func (repository PostgresRepository) RegistrarEntrada(ctx context.Context, input application.RegistrarEntradaInput, cadastro domain.EntradaCadastro) (application.Resultado, error) {
	hash := hashRequisicao(input)
	if resultado, encontrada, err := buscarRespostaIdempotente(ctx, repository.db, input.IdempotencyKey); err != nil {
		return application.Resultado{}, err
	} else if encontrada {
		return application.Resultado{Entrada: resultado, JaProcessada: true}, nil
	}

	resultado, err := repository.processarEntrada(ctx, input, cadastro)
	if err != nil {
		return application.Resultado{}, err
	}

	if err = gravarChaveIdempotencia(ctx, repository.db, input.IdempotencyKey, hash, resultado); err != nil {
		if isUniqueViolation(err) {
			existente, encontrada, ferr := carregarRespostaIdempotente(ctx, repository.db, input.IdempotencyKey)
			if ferr != nil {
				return application.Resultado{}, ferr
			}
			if encontrada {
				return application.Resultado{Entrada: existente, JaProcessada: true}, nil
			}
		}
		return application.Resultado{}, err
	}
	return application.Resultado{Entrada: resultado, JaProcessada: false}, nil
}

func (repository PostgresRepository) processarEntrada(ctx context.Context, input application.RegistrarEntradaInput, cadastro domain.EntradaCadastro) (domain.ResultadoEntrada, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.ResultadoEntrada{}, err
	}
	defer tx.Rollback(ctx)

	if cadastro.FornecedorID != "" {
		if err = validarFornecedorEntrada(ctx, tx, cadastro.FornecedorID); err != nil {
			return domain.ResultadoEntrada{}, err
		}
	}

	itens, err := carregarItens(ctx, tx, cadastro.Itens)
	if err != nil {
		return domain.ResultadoEntrada{}, err
	}

	var pedido *pedidoItensCarregados
	if cadastro.PedidoCompraID != "" {
		pedido, err = carregarPedido(ctx, tx, cadastro.PedidoCompraID, cadastro.Itens, cadastro.ConfirmarDivergencia)
		if err != nil {
			return domain.ResultadoEntrada{}, err
		}
		if cadastro.FornecedorID == "" {
			cadastro.FornecedorID = pedido.fornecedorID
		} else if cadastro.FornecedorID != pedido.fornecedorID {
			return domain.ResultadoEntrada{}, application.ErrFornecedorDivergente
		}
	}

	resultado := domain.ResultadoEntrada{
		DocumentoOrigem: cadastro.DocumentoOrigem, RegistradoEm: time.Now(), RegistradoPor: input.UsuarioID,
	}
	ordensAfetadas := map[string]struct{}{}
	for _, requisitado := range cadastro.Itens {
		item := itens[requisitado.ItemID]
		saldoFisicoAnterior := item.saldoFisico
		novoSaldoFisico := item.saldoFisico + requisitado.Quantidade
		if _, err = tx.Exec(ctx, "UPDATE item_estoque SET saldo_fisico = $2, custo_unitario = $3 WHERE id = $1",
			item.id, novoSaldoFisico, requisitado.CustoUnitario); err != nil {
			return domain.ResultadoEntrada{}, fmt.Errorf("atualizar saldo fisico: %w", err)
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO movimentacao_estoque (item_estoque_id, tipo, quantidade, custo_unitario, documento_origem, fornecedor_id)
			VALUES ($1, $2, $3, $4, $5, NULLIF($6, '')::uuid)`, item.id, domain.MovimentacaoEntrada, requisitado.Quantidade, requisitado.CustoUnitario, cadastro.DocumentoOrigem, cadastro.FornecedorID,
		); err != nil {
			if isUniqueViolation(err) {
				return domain.ResultadoEntrada{}, application.ErrDocumentoOrigemDuplicado
			}
			return domain.ResultadoEntrada{}, fmt.Errorf("registrar movimentacao: %w", err)
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO auditoria_estoque (item_estoque_id, fornecedor_id, pedido_compra_id, usuario_id, tipo_evento, documento_origem, dados, ocorrido_em)
			VALUES ($1, NULLIF($2, '')::uuid, NULLIF($3, '')::uuid, NULLIF($4, '')::uuid, 'ENTRADA_ESTOQUE', $5, jsonb_build_object('quantidade', $6::numeric, 'custoUnitario', $7::numeric), CURRENT_TIMESTAMP)`,
			item.id, cadastro.FornecedorID, cadastro.PedidoCompraID, input.UsuarioID, cadastro.DocumentoOrigem, requisitado.Quantidade, requisitado.CustoUnitario,
		); err != nil {
			return domain.ResultadoEntrada{}, fmt.Errorf("registrar auditoria de entrada: %w", err)
		}

		saldoReservado := item.saldoReservado
		if pedido != nil {
			if pedidoItem, ok := pedido.itens[requisitado.ItemID]; ok {
				efetivado, afetadas, err := efetivarReservas(ctx, tx, pedidoItem.id, item.id, requisitado.Quantidade)
				if err != nil {
					return domain.ResultadoEntrada{}, err
				}
				if efetivado > 0 {
					saldoReservado += efetivado
					if _, err = tx.Exec(ctx, "UPDATE item_estoque SET saldo_reservado = $2 WHERE id = $1", item.id, saldoReservado); err != nil {
						return domain.ResultadoEntrada{}, err
					}
				}
				if _, err = tx.Exec(ctx, "UPDATE pedido_compra_item SET quantidade_recebida = quantidade_recebida + $2, quantidade_reservada = quantidade_reservada + $3 WHERE id = $1",
					pedidoItem.id, requisitado.Quantidade, efetivado); err != nil {
					return domain.ResultadoEntrada{}, err
				}
				for _, osID := range afetadas {
					ordensAfetadas[osID] = struct{}{}
				}
			}
		}

		resultado.Itens = append(resultado.Itens, domain.ItemEntradaResultado{
			ItemID: item.id, Codigo: item.codigo, UnidadeMedida: item.unidadeMedida, Quantidade: requisitado.Quantidade,
			SaldoFisicoAnterior: saldoFisicoAnterior, SaldoFisicoAtual: novoSaldoFisico,
			SaldoReservado: saldoReservado, SaldoDisponivel: novoSaldoFisico - saldoReservado,
		})
	}

	if pedido != nil {
		statusPedido, err := recalcularStatusPedido(ctx, tx, cadastro.PedidoCompraID)
		if err != nil {
			return domain.ResultadoEntrada{}, err
		}
		resultado.PedidoCompra = &domain.PedidoCompraResultado{ID: cadastro.PedidoCompraID, Numero: pedido.numero, Status: statusPedido}
		resultado.OrdensServico, err = liberarOrdensServico(ctx, tx, ordensAfetadas)
		if err != nil {
			return domain.ResultadoEntrada{}, err
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return domain.ResultadoEntrada{}, err
	}
	return resultado, nil
}

func validarFornecedorEntrada(ctx context.Context, tx pgx.Tx, fornecedorID string) error {
	var ativo bool
	if err := tx.QueryRow(ctx, "SELECT ativo FROM fornecedor WHERE id = $1 FOR UPDATE", fornecedorID).Scan(&ativo); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return application.ErrFornecedorNaoEncontrado
		}
		return err
	}
	if !ativo {
		return application.ErrFornecedorInativo
	}
	return nil
}

func carregarItens(ctx context.Context, tx pgx.Tx, itens []domain.ItemEntrada) (map[string]itemRow, error) {
	ids := make([]string, len(itens))
	for index, item := range itens {
		ids[index] = item.ItemID
	}
	rows, err := tx.Query(ctx, `
		SELECT id, codigo, tipo, unidade_medida, ativo, saldo_fisico, saldo_reservado
		FROM item_estoque WHERE id = ANY($1) ORDER BY id FOR UPDATE`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	carregados := make(map[string]itemRow, len(itens))
	for rows.Next() {
		var row itemRow
		if err = rows.Scan(&row.id, &row.codigo, &row.tipo, &row.unidadeMedida, &row.ativo, &row.saldoFisico, &row.saldoReservado); err != nil {
			return nil, err
		}
		carregados[row.id] = row
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for _, item := range itens {
		row, ok := carregados[item.ItemID]
		if !ok {
			return nil, application.ErrItemNaoEncontrado
		}
		if !row.ativo {
			return nil, application.ErrItemInativo
		}
		if err = domain.QuantidadeValida(row.tipo, item.Quantidade); err != nil {
			return nil, err
		}
		if row.tipo == domain.TipoInsumo {
			if err = domain.QuantidadeCompativelComUnidade(item.Quantidade, row.unidadeMedida); err != nil {
				return nil, err
			}
		}
	}
	return carregados, nil
}

type pedidoItensCarregados struct {
	numero       string
	fornecedorID string
	itens        map[string]pedidoItemRow
}

func carregarPedido(ctx context.Context, tx pgx.Tx, pedidoID string, itens []domain.ItemEntrada, confirmarDivergencia bool) (*pedidoItensCarregados, error) {
	var numero, fornecedorID string
	if err := tx.QueryRow(ctx, "SELECT numero, fornecedor_id FROM pedido_compra WHERE id = $1 FOR UPDATE", pedidoID).Scan(&numero, &fornecedorID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, application.ErrPedidoCompraNaoEncontrado
		}
		return nil, err
	}
	rows, err := tx.Query(ctx, `
		SELECT id, item_estoque_id, quantidade_pedida, quantidade_recebida, quantidade_reservada
		FROM pedido_compra_item WHERE pedido_compra_id = $1 ORDER BY item_estoque_id FOR UPDATE`, pedidoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	pedido := &pedidoItensCarregados{numero: numero, fornecedorID: fornecedorID, itens: map[string]pedidoItemRow{}}
	for rows.Next() {
		var itemID string
		var row pedidoItemRow
		if err = rows.Scan(&row.id, &itemID, &row.quantidadePedida, &row.quantidadeRecebida, &row.quantidadeReservada); err != nil {
			return nil, err
		}
		pedido.itens[itemID] = row
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	for _, requisitado := range itens {
		pedidoItem, ok := pedido.itens[requisitado.ItemID]
		if !ok {
			return nil, application.ErrItemForaDoPedido
		}
		totalRecebido := pedidoItem.quantidadeRecebida + requisitado.Quantidade
		if totalRecebido > pedidoItem.quantidadePedida && !confirmarDivergencia {
			return nil, application.ErrDivergenciaQuantidade
		}
	}
	return pedido, nil
}

// efetivarReservas cria uma reserva ATIVA por vinculo (pedido_compra_item, ordem_servico_item)
// ainda nao reservado, na ordem dos vinculos, ate esgotar a quantidade recebida ou os vinculos.
func efetivarReservas(ctx context.Context, tx pgx.Tx, pedidoCompraItemID, itemEstoqueID string, quantidadeRecebida float64) (float64, []string, error) {
	rows, err := tx.Query(ctx, `
		SELECT pcio.ordem_servico_item_id,
		       pcio.quantidade_atendida,
		       osi.ordem_servico_id,
		       reserva.ja_reservado
		FROM pedido_compra_item_os pcio
		JOIN ordem_servico_item osi ON osi.id = pcio.ordem_servico_item_id
		CROSS JOIN LATERAL (
			SELECT COALESCE(SUM(r.quantidade), 0) AS ja_reservado
			FROM reserva_estoque r
			WHERE r.ordem_servico_item_id = pcio.ordem_servico_item_id
				AND r.pedido_compra_item_id = pcio.pedido_compra_item_id
				AND r.status = $2
		) reserva
		WHERE pcio.pedido_compra_item_id = $1
			AND reserva.ja_reservado < pcio.quantidade_atendida
		ORDER BY pcio.ordem_servico_item_id
		FOR UPDATE OF pcio, osi`, pedidoCompraItemID, domain.ReservaAtiva)
	if err != nil {
		return 0, nil, err
	}
	type link struct {
		osItemID, ordemServicoID string
		quantidadeAtendida       float64
		jaReservado              float64
	}
	var links []link
	for rows.Next() {
		var l link
		if err = rows.Scan(&l.osItemID, &l.quantidadeAtendida, &l.ordemServicoID, &l.jaReservado); err != nil {
			rows.Close()
			return 0, nil, err
		}
		links = append(links, l)
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return 0, nil, err
	}

	var totalEfetivado float64
	restante := quantidadeRecebida
	var afetadas []string
	for _, l := range links {
		if restante <= 0 {
			break
		}
		faltante := l.quantidadeAtendida - l.jaReservado
		if faltante <= 0 {
			continue
		}
		alocado := faltante
		if alocado > restante {
			alocado = restante
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO reserva_estoque (ordem_servico_item_id, item_estoque_id, pedido_compra_item_id, quantidade, status)
			VALUES ($1, $2, $3, $4, $5)`, l.osItemID, itemEstoqueID, pedidoCompraItemID, alocado, domain.ReservaAtiva,
		); err != nil {
			return 0, nil, fmt.Errorf("efetivar reserva: %w", err)
		}
		if _, err = tx.Exec(ctx, "UPDATE ordem_servico_item SET quantidade_reservada = quantidade_reservada + $2 WHERE id = $1", l.osItemID, alocado); err != nil {
			return 0, nil, fmt.Errorf("atualizar reserva da os: %w", err)
		}
		totalEfetivado += alocado
		restante -= alocado
		afetadas = append(afetadas, l.ordemServicoID)
	}
	return totalEfetivado, afetadas, nil
}

func recalcularStatusPedido(ctx context.Context, tx pgx.Tx, pedidoID string) (string, error) {
	var pendentes int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM pedido_compra_item WHERE pedido_compra_id = $1 AND quantidade_recebida < quantidade_pedida", pedidoID).Scan(&pendentes); err != nil {
		return "", err
	}
	status := "PARCIAL"
	if pendentes == 0 {
		status = "CONCLUIDO"
	}
	if _, err := tx.Exec(ctx, "UPDATE pedido_compra SET status = $2, recebido_em = CASE WHEN $3 = 'CONCLUIDO' THEN CURRENT_TIMESTAMP ELSE recebido_em END WHERE id = $1", pedidoID, status, status); err != nil {
		return "", err
	}
	return status, nil
}

// liberarOrdensServico move para AGUARDANDO_EXECUCAO as OS afetadas que nao tem mais vinculo
// pendente (pedido_compra_item_os cujo pedido ainda nao esta CONCLUIDO).
func liberarOrdensServico(ctx context.Context, tx pgx.Tx, ordensAfetadas map[string]struct{}) ([]domain.OrdemServicoLiberada, error) {
	var resultado []domain.OrdemServicoLiberada
	for ordemServicoID := range ordensAfetadas {
		var statusAtual string
		if err := tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", ordemServicoID).Scan(&statusAtual); err != nil {
			return nil, err
		}
		if statusAtual != "AGUARDANDO_RECURSOS" {
			continue
		}
		var itensPendentes int
		if err := tx.QueryRow(ctx, `
			SELECT COUNT(*) FROM pedido_compra_item_os pcio
			JOIN pedido_compra_item pci ON pci.id = pcio.pedido_compra_item_id
			JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
			JOIN ordem_servico_item osi ON osi.id = pcio.ordem_servico_item_id
			WHERE osi.ordem_servico_id = $1 AND pc.status <> 'CONCLUIDO'`, ordemServicoID).Scan(&itensPendentes); err != nil {
			return nil, err
		}
		novoStatus := statusAtual
		if itensPendentes == 0 {
			novoStatus = "AGUARDANDO_EXECUCAO"
			if _, err := tx.Exec(ctx, "UPDATE ordem_servico SET status = $2, data_entrada_fila = CURRENT_TIMESTAMP WHERE id = $1", ordemServicoID, novoStatus); err != nil {
				return nil, err
			}
		}
		resultado = append(resultado, domain.OrdemServicoLiberada{
			OrdemServicoID: ordemServicoID, StatusAnterior: statusAtual, Status: novoStatus, ItensPendentes: itensPendentes,
		})
	}
	return resultado, nil
}

func hashRequisicao(input application.RegistrarEntradaInput) string {
	soma := sha256.Sum256([]byte(fmt.Sprintf("%s|%s|%s|%v|%v", input.DocumentoOrigem, input.FornecedorID, input.PedidoCompraID, input.ConfirmarDivergencia, input.Itens)))
	return hex.EncodeToString(soma[:])
}

func buscarRespostaIdempotente(ctx context.Context, db *pgxpool.Pool, chave string) (domain.ResultadoEntrada, bool, error) {
	return carregarRespostaIdempotente(ctx, db, chave)
}

func carregarRespostaIdempotente(ctx context.Context, db *pgxpool.Pool, chave string) (domain.ResultadoEntrada, bool, error) {
	var resposta []byte
	err := db.QueryRow(ctx, "SELECT resposta FROM chave_idempotencia WHERE operacao = $1 AND chave = $2", operacaoRegistrarEntrada, chave).Scan(&resposta)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ResultadoEntrada{}, false, nil
	}
	if err != nil {
		return domain.ResultadoEntrada{}, false, err
	}
	var resultado domain.ResultadoEntrada
	if err = json.Unmarshal(resposta, &resultado); err != nil {
		return domain.ResultadoEntrada{}, false, err
	}
	return resultado, true, nil
}

func gravarChaveIdempotencia(ctx context.Context, db *pgxpool.Pool, chave, hash string, resultado domain.ResultadoEntrada) error {
	payload, err := json.Marshal(resultado)
	if err != nil {
		return err
	}
	_, err = db.Exec(ctx, `
		INSERT INTO chave_idempotencia (chave, operacao, hash_requisicao, status_resposta, resposta)
		VALUES ($1, $2, $3, 201, $4)`, chave, operacaoRegistrarEntrada, hash, payload)
	return err
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
