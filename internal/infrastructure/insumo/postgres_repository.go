package insumo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	insumoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/insumo"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/insumo"
)

// Os saldos saem como texto porque insumo e fracionario: converter para float aqui
// perderia a precisao que o NUMERIC(14,3) garante.
const colunas = `i.id, i.codigo, i.nome, i.descricao, i.categoria_id, c.nome,
	COALESCE(i.fornecedor_id::text, ''), i.unidade_medida, i.custo_unitario::text,
	i.saldo_fisico::text, i.saldo_reservado::text, i.estoque_minimo::text,
	i.ativo, i.version,
	EXISTS (
		SELECT 1 FROM pedido_compra_item pci
		JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		WHERE pci.item_estoque_id = i.id AND pc.status IN ('ABERTO', 'PARCIAL')
	)`

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (repository PostgresRepository) Desativar(ctx context.Context, insumoID, usuarioID string) (insumo.Insumo, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return insumo.Insumo{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var tipo string
	var ativo bool
	err = tx.QueryRow(ctx, `SELECT tipo, ativo FROM item_estoque WHERE id = $1 FOR UPDATE`, insumoID).Scan(&tipo, &ativo)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && tipo != insumo.Tipo) {
		return insumo.Insumo{}, insumoApplication.ErrInsumoNaoEncontrado
	}
	if err != nil {
		return insumo.Insumo{}, err
	}
	if !ativo {
		return insumo.Insumo{}, insumo.ErrInsumoJaInativo
	}

	var raw []byte
	err = tx.QueryRow(ctx, `SELECT COALESCE(jsonb_agg(DISTINCT bloqueio.ordem_servico_id), '[]'::jsonb)
		FROM (
			SELECT ordem_servico_id FROM ordem_servico_item
			WHERE item_estoque_id = $1 AND quantidade_reservada > 0
			UNION
			SELECT o.ordem_servico_id FROM orcamento_item oi
			JOIN orcamento o ON o.id = oi.orcamento_id
			WHERE oi.item_estoque_id = $1 AND o.status = 'CRIADO'
		) bloqueio`, insumoID).Scan(&raw)
	if err != nil {
		return insumo.Insumo{}, err
	}
	var ordens []string
	if err := json.Unmarshal(raw, &ordens); err != nil {
		return insumo.Insumo{}, err
	}
	if len(ordens) > 0 {
		return insumo.Insumo{}, insumoApplication.InsumoEmUsoError{OrdensServico: ordens}
	}

	var item insumo.Insumo
	var dataDesativacao time.Time
	var usuarioDesativacao string
	err = tx.QueryRow(ctx, `UPDATE item_estoque
		SET ativo = FALSE, data_desativacao = CURRENT_TIMESTAMP, usuario_desativacao = $2, version = version + 1
		WHERE id = $1 AND tipo = 'INSUMO' AND ativo
		RETURNING id, codigo, nome, unidade_medida, saldo_fisico::text, ativo, version, data_desativacao, usuario_desativacao::text`, insumoID, usuarioID).
		Scan(&item.ID, &item.Codigo, &item.Nome, &item.UnidadeMedida, &item.SaldoFisico, &item.Ativo, &item.Version, &dataDesativacao, &usuarioDesativacao)
	if err != nil {
		return insumo.Insumo{}, err
	}
	item.DataDesativacao = &dataDesativacao
	item.UsuarioDesativacao = &usuarioDesativacao
	if err := tx.Commit(ctx); err != nil {
		return insumo.Insumo{}, err
	}
	return item, nil
}

func (repository PostgresRepository) BuscarPorFiltro(ctx context.Context, filtros insumoApplication.FiltrosConsulta, limite, deslocamento int) ([]insumo.Insumo, int, error) {
	condicoes, args := montarCondicoes(filtros)

	var total int
	contagem := "SELECT COUNT(*) FROM item_estoque i WHERE " + condicoes
	if err := repository.db.QueryRow(ctx, contagem, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return nil, 0, nil
	}

	listagem := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE %s ORDER BY i.codigo LIMIT $%d OFFSET $%d`,
		colunas, condicoes, len(args)+1, len(args)+2)

	rows, err := repository.db.Query(ctx, listagem, append(args, limite, deslocamento)...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	itens := []insumo.Insumo{}
	for rows.Next() {
		item, err := ler(rows)
		if err != nil {
			return nil, 0, err
		}
		itens = append(itens, item)
	}
	return itens, total, rows.Err()
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (insumo.Insumo, error) {
	consulta := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE i.id = $1 AND i.tipo = 'INSUMO'`, colunas)

	item, err := ler(repository.db.QueryRow(ctx, consulta, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return insumo.Insumo{}, insumoApplication.ErrInsumoNaoEncontrado
	}
	return item, err
}

// Cadastrar insere o insumo e devolve a entidade montada pelo BuscarPorID, que ja traz o
// nome da categoria e a flag de pedido em aberto sem duplicar o SELECT.
func (repository PostgresRepository) Cadastrar(ctx context.Context, cadastro insumo.Cadastro) (insumo.Insumo, error) {
	var ativa bool
	err := repository.db.QueryRow(ctx,
		`SELECT ativa FROM categoria WHERE id = $1`, cadastro.CategoriaID).Scan(&ativa)
	if errors.Is(err, pgx.ErrNoRows) {
		return insumo.Insumo{}, insumoApplication.ErrCategoriaInvalida
	}
	if err != nil {
		return insumo.Insumo{}, err
	}
	if !ativa {
		return insumo.Insumo{}, insumoApplication.ErrCategoriaInvalida
	}
	if err := validarFornecedor(ctx, repository.db, cadastro.FornecedorID); err != nil {
		return insumo.Insumo{}, err
	}

	var id string
	var criadoEm time.Time
	err = repository.db.QueryRow(ctx, `
		INSERT INTO item_estoque (
			categoria_id, tipo, codigo, nome, descricao, descricao_normalizada,
			fornecedor_id, unidade_medida, custo_unitario, estoque_minimo
		) VALUES (
			$1, 'INSUMO', 'INS-' || LPAD(nextval('seq_insumo_codigo')::TEXT, 6, '0'),
			$2, $3, $4, $5, $6, $7::NUMERIC, $8::NUMERIC
		)
		RETURNING id, criado_em`,
		cadastro.CategoriaID, cadastro.Nome, cadastro.Descricao, cadastro.DescricaoNormalizada,
		valorFornecedor(cadastro.FornecedorID), cadastro.UnidadeMedida, cadastro.CustoUnitario, cadastro.EstoqueMinimo,
	).Scan(&id, &criadoEm)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return insumo.Insumo{}, insumoApplication.ErrDescricaoDuplicada
		}
		return insumo.Insumo{}, err
	}

	cadastrado, err := repository.BuscarPorID(ctx, id)
	if err != nil {
		return insumo.Insumo{}, err
	}
	cadastrado.DataCriacao = &criadoEm
	return cadastrado, nil
}

type itemProcessamento struct {
	id             string
	tipo           string
	ativo          bool
	saldoFisico    string
	saldoReservado string
	custoUnitario  *string
	osItemID       *string
	vinculado      bool
	jaProcessado   bool
	quantidade     string
}

func (repository PostgresRepository) SolicitarCompraEReservar(ctx context.Context, solicitacao insumoApplication.SolicitacaoCompraReserva) (insumoApplication.ResultadoCompraReserva, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if resultado, encontrado, err := buscarRespostaIdempotente(ctx, tx, solicitacao); err != nil || encontrado {
		if err != nil {
			return insumoApplication.ResultadoCompraReserva{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return insumoApplication.ResultadoCompraReserva{}, err
		}
		return resultado, nil
	}

	resultado := insumoApplication.ResultadoCompraReserva{
		OrdemServicoID:          solicitacao.OrdemServicoID,
		ValorTotalCompraParcial: json.Number("0"),
	}
	if err = carregarFornecedor(ctx, tx, solicitacao.FornecedorID, &resultado); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	if err = validarOrdemComOrcamentoAprovado(ctx, tx, solicitacao.OrdemServicoID); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}

	itens, err := carregarItensProcessamento(ctx, tx, solicitacao)
	if err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}

	var pedidoID string
	totalCompra := new(big.Rat)
	for _, item := range itens {
		if item.tipo != insumo.Tipo || !item.ativo || !item.vinculado {
			return insumoApplication.ResultadoCompraReserva{}, insumoApplication.ErrItemProcessamentoInvalido
		}
		if item.jaProcessado {
			return insumoApplication.ResultadoCompraReserva{}, insumoApplication.ErrProcessamentoDuplicado
		}

		processamento := insumo.NovoProcessamento(item.id, item.quantidade, insumo.Insumo{
			SaldoFisico:    item.saldoFisico,
			SaldoReservado: item.saldoReservado,
		}.SaldoDisponivel())
		osItemID, err := garantirItemOrdemServico(ctx, tx, solicitacao.OrdemServicoID, item)
		if err != nil {
			return insumoApplication.ResultadoCompraReserva{}, err
		}
		if processamento.QuantidadeReservada != "0" {
			reservaID, err := reservarSaldoDisponivel(ctx, tx, solicitacao.OrdemServicoID, osItemID, processamento)
			if err != nil {
				return insumoApplication.ResultadoCompraReserva{}, err
			}
			if err = registrarMovimentacaoReserva(ctx, tx, solicitacao.OrdemServicoID, reservaID, processamento); err != nil {
				return insumoApplication.ResultadoCompraReserva{}, err
			}
			resultado.InsumosReservados = append(resultado.InsumosReservados, insumoApplication.ItemReservado{
				ItemID:              item.id,
				Quantidade:          json.Number(processamento.QuantidadeReservada),
				SaldoDisponivelApos: json.Number(processamento.SaldoDisponivelApos),
			})
		}
		if processamento.QuantidadeCompra != "0" {
			if pedidoID == "" {
				pedidoID, err = criarPedidoCompra(ctx, tx, solicitacao.FornecedorID)
				if err != nil {
					return insumoApplication.ResultadoCompraReserva{}, err
				}
			}
			if err = solicitarCompra(ctx, tx, pedidoID, osItemID, processamento, item.custoUnitario); err != nil {
				return insumoApplication.ResultadoCompraReserva{}, err
			}
			valorParcial := valorParcial(item.custoUnitario, processamento.QuantidadeCompra)
			totalCompra.Add(totalCompra, valorParcial)
			resultado.InsumosCompraSolicitada = append(resultado.InsumosCompraSolicitada, insumoApplication.ItemCompraSolicitada{
				ItemID:       item.id,
				Quantidade:   json.Number(processamento.QuantidadeCompra),
				ValorParcial: json.Number(formatarValor(valorParcial)),
			})
		}
	}

	resultado.ValorTotalCompraParcial = json.Number(formatarValor(totalCompra))
	resultado.StatusOrdemServico = "AGUARDANDO_EXECUCAO"
	if len(resultado.InsumosCompraSolicitada) > 0 {
		resultado.StatusOrdemServico = "AGUARDANDO_RECURSOS"
	}
	if _, err = tx.Exec(ctx, `UPDATE ordem_servico SET status = $2,
		data_entrada_fila = CASE WHEN $2 = 'AGUARDANDO_EXECUCAO' THEN CURRENT_TIMESTAMP ELSE data_entrada_fila END
		WHERE id = $1`, solicitacao.OrdemServicoID, resultado.StatusOrdemServico); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	if err = registrarAuditoriaProcessamento(ctx, tx, solicitacao, resultado); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	if err = gravarRespostaIdempotente(ctx, tx, solicitacao, resultado); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, err
	}
	return resultado, nil
}

func buscarRespostaIdempotente(ctx context.Context, tx pgx.Tx, solicitacao insumoApplication.SolicitacaoCompraReserva) (insumoApplication.ResultadoCompraReserva, bool, error) {
	var hash string
	var resposta []byte
	err := tx.QueryRow(ctx, `
		SELECT hash_requisicao, resposta
		FROM chave_idempotencia
		WHERE operacao = $1 AND chave = $2
		FOR UPDATE`, insumoApplication.OperacaoSolicitarCompraReservarInsumos, solicitacao.IdempotencyKey,
	).Scan(&hash, &resposta)
	if errors.Is(err, pgx.ErrNoRows) {
		return insumoApplication.ResultadoCompraReserva{}, false, nil
	}
	if err != nil {
		return insumoApplication.ResultadoCompraReserva{}, false, err
	}
	if hash != solicitacao.HashRequisicao {
		return insumoApplication.ResultadoCompraReserva{}, true, insumoApplication.ErrIdempotencyKeyEmUso
	}
	var resultado insumoApplication.ResultadoCompraReserva
	if err = json.Unmarshal(resposta, &resultado); err != nil {
		return insumoApplication.ResultadoCompraReserva{}, false, err
	}
	resultado.Reprocessado = true
	return resultado, true, nil
}

func carregarFornecedor(ctx context.Context, tx pgx.Tx, fornecedorID string, resultado *insumoApplication.ResultadoCompraReserva) error {
	var ativo bool
	err := tx.QueryRow(ctx, `
		SELECT id, COALESCE(nome_fantasia, razao_social), ativo
		FROM fornecedor WHERE id = $1`, fornecedorID,
	).Scan(&resultado.Fornecedor.ID, &resultado.Fornecedor.Nome, &ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return insumoApplication.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return err
	}
	if !ativo {
		return insumoApplication.ErrFornecedorInativo
	}
	return nil
}

func validarOrdemComOrcamentoAprovado(ctx context.Context, tx pgx.Tx, ordemServicoID string) error {
	var possuiOrcamento bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM orcamento
			WHERE ordem_servico_id = os.id AND status = 'APROVADO'
		)
		FROM ordem_servico os
		WHERE os.id = $1`, ordemServicoID,
	).Scan(&possuiOrcamento)
	if errors.Is(err, pgx.ErrNoRows) {
		return insumoApplication.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return err
	}
	if !possuiOrcamento {
		return insumoApplication.ErrOrdemServicoInvalida
	}
	return nil
}

func carregarItensProcessamento(ctx context.Context, tx pgx.Tx, solicitacao insumoApplication.SolicitacaoCompraReserva) ([]itemProcessamento, error) {
	quantidades := make(map[string]string, len(solicitacao.Itens))
	ids := make([]string, 0, len(solicitacao.Itens))
	for _, item := range solicitacao.Itens {
		quantidades[item.ItemID] = item.Quantidade.String()
		ids = append(ids, item.ItemID)
	}

	rows, err := tx.Query(ctx, `
		SELECT i.id, i.tipo, i.ativo, i.saldo_fisico::text, i.saldo_reservado::text, i.custo_unitario::text,
		       osi.id::text,
		       EXISTS (
		           SELECT 1 FROM orcamento o
		           JOIN orcamento_item oi ON oi.orcamento_id = o.id
		           WHERE o.ordem_servico_id = $1
		             AND o.status = 'APROVADO'
		             AND oi.item_estoque_id = i.id
		       ) OR osi.id IS NOT NULL AS vinculado,
		       EXISTS (
		           SELECT 1 FROM reserva_estoque r
		           JOIN ordem_servico_item osi2 ON osi2.id = r.ordem_servico_item_id
		           WHERE osi2.ordem_servico_id = $1
		             AND r.item_estoque_id = i.id
		             AND r.status = 'ATIVA'
		       ) OR EXISTS (
		           SELECT 1 FROM pedido_compra_item_os pcios
		           JOIN pedido_compra_item pci ON pci.id = pcios.pedido_compra_item_id
		           JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		           JOIN ordem_servico_item osi3 ON osi3.id = pcios.ordem_servico_item_id
		           WHERE osi3.ordem_servico_id = $1
		             AND pci.item_estoque_id = i.id
		             AND pc.status IN ('ABERTO', 'PARCIAL')
		       ) AS ja_processado
		FROM item_estoque i
		LEFT JOIN ordem_servico_item osi ON osi.ordem_servico_id = $1 AND osi.item_estoque_id = i.id
		WHERE i.id = ANY($2::uuid[])
		ORDER BY i.id
		FOR UPDATE OF i`, solicitacao.OrdemServicoID, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	itens := make([]itemProcessamento, 0, len(ids))
	for rows.Next() {
		var item itemProcessamento
		if err := rows.Scan(&item.id, &item.tipo, &item.ativo, &item.saldoFisico, &item.saldoReservado,
			&item.custoUnitario, &item.osItemID, &item.vinculado, &item.jaProcessado); err != nil {
			return nil, err
		}
		item.quantidade = quantidades[item.id]
		itens = append(itens, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(itens) != len(ids) {
		return nil, insumoApplication.ErrInsumoNaoEncontrado
	}
	return itens, nil
}

func garantirItemOrdemServico(ctx context.Context, tx pgx.Tx, ordemServicoID string, item itemProcessamento) (string, error) {
	if item.osItemID != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE ordem_servico_item
			SET quantidade_necessaria = GREATEST(quantidade_necessaria, $2::NUMERIC)
			WHERE id = $1`, *item.osItemID, item.quantidade); err != nil {
			return "", err
		}
		return *item.osItemID, nil
	}

	var osItemID string
	err := tx.QueryRow(ctx, `
		INSERT INTO ordem_servico_item (ordem_servico_id, item_estoque_id, quantidade_necessaria, valor_unitario)
		VALUES ($1, $2, $3::NUMERIC, $4::NUMERIC)
		RETURNING id`, ordemServicoID, item.id, item.quantidade, valorTexto(item.custoUnitario),
	).Scan(&osItemID)
	return osItemID, err
}

func reservarSaldoDisponivel(ctx context.Context, tx pgx.Tx, ordemServicoID, osItemID string, processamento insumo.Processamento) (string, error) {
	if _, err := tx.Exec(ctx, `
		UPDATE item_estoque
		SET saldo_reservado = saldo_reservado + $2::NUMERIC
		WHERE id = $1`, processamento.ItemID, processamento.QuantidadeReservada); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE ordem_servico_item
		SET quantidade_reservada = quantidade_reservada + $2::NUMERIC
		WHERE id = $1`, osItemID, processamento.QuantidadeReservada); err != nil {
		return "", err
	}

	var reservaID string
	err := tx.QueryRow(ctx, `
		INSERT INTO reserva_estoque (ordem_servico_item_id, item_estoque_id, quantidade, status)
		VALUES ($1, $2, $3::NUMERIC, 'ATIVA')
		RETURNING id`, osItemID, processamento.ItemID, processamento.QuantidadeReservada,
	).Scan(&reservaID)
	return reservaID, err
}

func registrarMovimentacaoReserva(ctx context.Context, tx pgx.Tx, ordemServicoID, reservaID string, processamento insumo.Processamento) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade)
		VALUES ($1, $2, $3, 'RESERVA', $4::NUMERIC)`,
		processamento.ItemID, ordemServicoID, reservaID, processamento.QuantidadeReservada)
	return err
}

func criarPedidoCompra(ctx context.Context, tx pgx.Tx, fornecedorID string) (string, error) {
	var pedidoID string
	err := tx.QueryRow(ctx, `
		INSERT INTO pedido_compra (fornecedor_id, numero, status)
		VALUES ($1, to_char(CURRENT_DATE, 'YYYY') || '/' || LPAD(nextval('seq_pedido_compra_numero')::TEXT, 4, '0'), 'ABERTO')
		RETURNING id`, fornecedorID,
	).Scan(&pedidoID)
	return pedidoID, err
}

func solicitarCompra(ctx context.Context, tx pgx.Tx, pedidoID, osItemID string, processamento insumo.Processamento, custoUnitario *string) error {
	var pedidoItemID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO pedido_compra_item (
			pedido_compra_id, item_estoque_id, quantidade_necessaria, quantidade_pedida, quantidade_reservada, custo_unitario
		) VALUES ($1, $2, $3::NUMERIC, $3::NUMERIC, 0, $4::NUMERIC)
		RETURNING id`, pedidoID, processamento.ItemID, processamento.QuantidadeCompra, valorTexto(custoUnitario),
	).Scan(&pedidoItemID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO pedido_compra_item_os (pedido_compra_item_id, ordem_servico_item_id, quantidade_atendida)
		VALUES ($1, $2, $3::NUMERIC)`, pedidoItemID, osItemID, processamento.QuantidadeCompra)
	return err
}

func registrarAuditoriaProcessamento(ctx context.Context, tx pgx.Tx, solicitacao insumoApplication.SolicitacaoCompraReserva, resultado insumoApplication.ResultadoCompraReserva) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, ocorrido_em)
		VALUES ($1, NULL, 'ESTOQUE', $1, 'INSUMOS_RESERVA_COMPRA_PROCESSADA', $2::jsonb, CURRENT_TIMESTAMP)`,
		solicitacao.OrdemServicoID,
		fmt.Sprintf(`{"fornecedorId":"%s","statusOrdemServico":"%s","insumosReservados":%d,"insumosCompraSolicitada":%d}`,
			solicitacao.FornecedorID, resultado.StatusOrdemServico, len(resultado.InsumosReservados), len(resultado.InsumosCompraSolicitada)))
	return err
}

func gravarRespostaIdempotente(ctx context.Context, tx pgx.Tx, solicitacao insumoApplication.SolicitacaoCompraReserva, resultado insumoApplication.ResultadoCompraReserva) error {
	resposta, err := json.Marshal(resultado)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO chave_idempotencia (chave, operacao, hash_requisicao, status_resposta, resposta)
		VALUES ($1, $2, $3, 201, $4::jsonb)`,
		solicitacao.IdempotencyKey, insumoApplication.OperacaoSolicitarCompraReservarInsumos, solicitacao.HashRequisicao, string(resposta))
	return err
}

func valorTexto(valor *string) string {
	if valor == nil || strings.TrimSpace(*valor) == "" {
		return "0"
	}
	return strings.TrimSpace(*valor)
}

func valorParcial(custo *string, quantidade string) *big.Rat {
	return new(big.Rat).Mul(decimal(valorTexto(custo)), decimal(quantidade))
}

func decimal(valor string) *big.Rat {
	resultado, ok := new(big.Rat).SetString(valor)
	if !ok {
		return new(big.Rat)
	}
	return resultado
}

func formatarValor(valor *big.Rat) string {
	return valor.FloatString(2)
}

type scanner interface {
	Scan(destinos ...any) error
}

func ler(linha scanner) (insumo.Insumo, error) {
	var item insumo.Insumo
	var fornecedorID string
	err := linha.Scan(
		&item.ID, &item.Codigo, &item.Nome, &item.Descricao, &item.CategoriaID, &item.Categoria,
		&fornecedorID, &item.UnidadeMedida, &item.CustoUnitario,
		&item.SaldoFisico, &item.SaldoReservado, &item.EstoqueMinimo,
		&item.Ativo, &item.Version, &item.PossuiPedidoEmAberto,
	)
	if fornecedorID != "" {
		item.FornecedorID = &fornecedorID
	}
	return item, err
}

func montarCondicoes(filtros insumoApplication.FiltrosConsulta) (string, []any) {
	condicoes := []string{"i.tipo = 'INSUMO'"}
	var args []any

	adicionar := func(formato string, valor any) {
		args = append(args, valor)
		condicoes = append(condicoes, fmt.Sprintf(formato, len(args)))
	}

	if !filtros.IncluirInativos {
		condicoes = append(condicoes, "i.ativo")
	}
	if filtros.Codigo != "" {
		adicionar("i.codigo = $%d", filtros.Codigo)
	}
	if filtros.Descricao != "" {
		adicionar("i.descricao ILIKE $%d", "%"+filtros.Descricao+"%")
	}
	if filtros.CategoriaID != "" {
		adicionar("i.categoria_id = $%d", filtros.CategoriaID)
	}
	if filtros.SomenteDisponiveis {
		adicionar("(i.saldo_fisico - i.saldo_reservado) >= $%d::NUMERIC", *filtros.QuantidadeDesejada)
	}
	return strings.Join(condicoes, " AND "), args
}

type consultor interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func validarFornecedor(ctx context.Context, consultor consultor, fornecedorID *string) error {
	if fornecedorID == nil {
		return nil
	}
	var ativo bool
	err := consultor.QueryRow(ctx, `SELECT ativo FROM fornecedor WHERE id = $1`, *fornecedorID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && !ativo) {
		return insumoApplication.ErrFornecedorInvalido
	}
	return err
}

func valorFornecedor(fornecedorID *string) any {
	if fornecedorID == nil {
		return nil
	}
	return *fornecedorID
}
