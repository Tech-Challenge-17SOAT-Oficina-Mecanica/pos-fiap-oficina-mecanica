package peca

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	pecaApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/peca"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/peca"
)

const colunas = `i.id, i.codigo, i.nome, i.descricao, i.categoria_id, c.nome,
	i.fabricante, i.unidade_medida, i.preco_venda::text,
	i.saldo_fisico, i.saldo_reservado, i.estoque_minimo, i.ativo, i.version,
	EXISTS (
		SELECT 1 FROM pedido_compra_item pci
		JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
		WHERE pci.item_estoque_id = i.id AND pc.status IN ('ABERTO', 'PARCIAL')
	)`

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (repository PostgresRepository) BuscarPorFiltro(ctx context.Context, filtros pecaApplication.Filtros, limite, deslocamento int) ([]peca.Peca, int, error) {
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

	var pecas []peca.Peca
	for rows.Next() {
		item, err := ler(rows)
		if err != nil {
			return nil, 0, err
		}
		pecas = append(pecas, item)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return pecas, total, nil
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (peca.Peca, error) {
	consulta := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE i.id = $1 AND i.tipo = 'PECA'`, colunas)

	item, err := ler(repository.db.QueryRow(ctx, consulta, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrNaoEncontrada
	}
	return item, err
}

type scanner interface {
	Scan(destinos ...any) error
}

func ler(linha scanner) (peca.Peca, error) {
	var item peca.Peca
	err := linha.Scan(
		&item.ID, &item.Codigo, &item.Nome, &item.Descricao, &item.CategoriaID, &item.Categoria,
		&item.Fabricante, &item.UnidadeMedida, &item.PrecoVenda,
		&item.SaldoFisico, &item.SaldoReservado, &item.EstoqueMinimo, &item.Ativo, &item.Version,
		&item.PossuiPedidoEmAberto,
	)
	return item, err
}

func montarCondicoes(filtros pecaApplication.Filtros) (string, []any) {
	condicoes := []string{"i.tipo = 'PECA'"}
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
	if filtros.Fabricante != "" {
		adicionar("i.fabricante ILIKE $%d", "%"+filtros.Fabricante+"%")
	}
	if filtros.SomenteDisponiveis {
		condicoes = append(condicoes, "(i.saldo_fisico - i.saldo_reservado) > 0")
	}
	return strings.Join(condicoes, " AND "), args
}

func (repository PostgresRepository) OrdensComReservaAtiva(ctx context.Context, itemID string) ([]string, error) {
	const consulta = `SELECT DISTINCT osi.ordem_servico_id::text
		FROM reserva_estoque r
		JOIN ordem_servico_item osi ON osi.id = r.ordem_servico_item_id
		WHERE r.item_estoque_id = $1 AND r.status = 'ATIVA'
		ORDER BY 1`

	rows, err := repository.db.Query(ctx, consulta, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ordens []string
	for rows.Next() {
		var ordemID string
		if err := rows.Scan(&ordemID); err != nil {
			return nil, err
		}
		ordens = append(ordens, ordemID)
	}
	return ordens, rows.Err()
}

func (repository PostgresRepository) EmOrcamentoCriado(ctx context.Context, itemID string) (bool, error) {
	const consulta = `SELECT EXISTS (
		SELECT 1 FROM orcamento_item oi
		JOIN orcamento o ON o.id = oi.orcamento_id
		WHERE oi.item_estoque_id = $1 AND o.status = 'CRIADO'
	)`

	var existe bool
	err := repository.db.QueryRow(ctx, consulta, itemID).Scan(&existe)
	return existe, err
}

func (repository PostgresRepository) Desativar(ctx context.Context, item peca.Peca) error {
	const comando = `UPDATE item_estoque
		SET ativo = FALSE, data_desativacao = $2, usuario_desativacao = $3
		WHERE id = $1 AND ativo AND tipo = 'PECA'`

	etiqueta, err := repository.db.Exec(ctx, comando, item.ID, item.DataDesativacao, item.UsuarioDesativacao)
	if err != nil {
		return err
	}
	if etiqueta.RowsAffected() == 0 {
		return peca.ErrJaInativa
	}
	return nil
}

// Cadastrar insere a peca e devolve a entidade ja montada pelo BuscarPorID, que traz o
// nome da categoria e a flag de pedido em aberto sem duplicar o SELECT.
func (repository PostgresRepository) Cadastrar(ctx context.Context, cadastro peca.Cadastro) (peca.Peca, error) {
	var ativa bool
	err := repository.db.QueryRow(ctx,
		`SELECT ativa FROM categoria WHERE id = $1`, cadastro.CategoriaID).Scan(&ativa)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if !ativa {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}

	var id string
	var criadoEm time.Time
	err = repository.db.QueryRow(ctx, `
		INSERT INTO item_estoque (
			categoria_id, tipo, codigo, nome, descricao, descricao_normalizada,
			fabricante, unidade_medida, preco_venda, estoque_minimo
		) VALUES (
			$1, 'PECA', 'PEC-' || LPAD(nextval('seq_peca_codigo')::TEXT, 6, '0'),
			$2, $3, $4, $5, $6, $7, $8
		)
		RETURNING id, criado_em`,
		cadastro.CategoriaID, cadastro.Nome, cadastro.Descricao, cadastro.DescricaoNormalizada,
		cadastro.Fabricante, cadastro.UnidadeMedida, cadastro.PrecoVenda, cadastro.EstoqueMinimo,
	).Scan(&id, &criadoEm)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return peca.Peca{}, pecaApplication.ErrDescricaoDuplicada
		}
		return peca.Peca{}, err
	}

	cadastrada, err := repository.BuscarPorID(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	cadastrada.DataCriacao = &criadoEm
	return cadastrada, nil
}

type itemProcessamento struct {
	id             string
	tipo           string
	ativo          bool
	saldoFisico    int64
	saldoReservado int64
	precoVenda     *string
	osItemID       *string
	vinculado      bool
	jaProcessado   bool
	quantidade     int64
}

func (repository PostgresRepository) SolicitarCompraEReservar(ctx context.Context, solicitacao pecaApplication.SolicitacaoCompraReserva) (pecaApplication.ResultadoCompraReserva, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if resultado, encontrado, err := buscarRespostaIdempotente(ctx, tx, solicitacao); err != nil || encontrado {
		if err != nil {
			return pecaApplication.ResultadoCompraReserva{}, err
		}
		if err = tx.Commit(ctx); err != nil {
			return pecaApplication.ResultadoCompraReserva{}, err
		}
		return resultado, nil
	}

	resultado := pecaApplication.ResultadoCompraReserva{OrdemServicoID: solicitacao.OrdemServicoID}
	if err = carregarFornecedor(ctx, tx, solicitacao.FornecedorID, &resultado); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	if err = validarOrdemComOrcamentoAprovado(ctx, tx, solicitacao.OrdemServicoID); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}

	itens, err := carregarItensProcessamento(ctx, tx, solicitacao)
	if err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}

	var pedidoID string
	for _, item := range itens {
		if item.tipo != peca.TipoPeca || !item.ativo || !item.vinculado {
			return pecaApplication.ResultadoCompraReserva{}, pecaApplication.ErrItemProcessamentoInvalido
		}
		if item.jaProcessado {
			return pecaApplication.ResultadoCompraReserva{}, pecaApplication.ErrProcessamentoDuplicado
		}

		processamento := peca.NovoProcessamento(item.id, item.quantidade, item.saldoFisico-item.saldoReservado)
		osItemID, err := garantirItemOrdemServico(ctx, tx, solicitacao.OrdemServicoID, item)
		if err != nil {
			return pecaApplication.ResultadoCompraReserva{}, err
		}
		if processamento.QuantidadeReservada > 0 {
			reservaID, err := reservarSaldoDisponivel(ctx, tx, solicitacao.OrdemServicoID, osItemID, processamento)
			if err != nil {
				return pecaApplication.ResultadoCompraReserva{}, err
			}
			if err = registrarMovimentacaoReserva(ctx, tx, solicitacao.OrdemServicoID, reservaID, processamento); err != nil {
				return pecaApplication.ResultadoCompraReserva{}, err
			}
			resultado.PecasReservadas = append(resultado.PecasReservadas, pecaApplication.ItemReservado{
				ItemID:              item.id,
				Quantidade:          processamento.QuantidadeReservada,
				SaldoDisponivelApos: processamento.SaldoDisponivelApos,
			})
		}
		if processamento.QuantidadeCompra > 0 {
			if pedidoID == "" {
				pedidoID, err = criarPedidoCompra(ctx, tx, solicitacao.FornecedorID)
				if err != nil {
					return pecaApplication.ResultadoCompraReserva{}, err
				}
			}
			if err = solicitarCompra(ctx, tx, pedidoID, osItemID, processamento); err != nil {
				return pecaApplication.ResultadoCompraReserva{}, err
			}
			valorParcial := valorParcial(item.precoVenda, processamento.QuantidadeCompra)
			resultado.ValorTotalCompraParcial += valorParcial
			resultado.PecasCompraSolicitada = append(resultado.PecasCompraSolicitada, pecaApplication.ItemCompraSolicitada{
				ItemID:       item.id,
				Quantidade:   processamento.QuantidadeCompra,
				ValorParcial: valorParcial,
			})
		}
	}

	resultado.StatusOrdemServico = "AGUARDANDO_EXECUCAO"
	if len(resultado.PecasCompraSolicitada) > 0 {
		resultado.StatusOrdemServico = "AGUARDANDO_RECURSOS"
	}
	if _, err = tx.Exec(ctx, `UPDATE ordem_servico SET status = $2 WHERE id = $1`, solicitacao.OrdemServicoID, resultado.StatusOrdemServico); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	if err = registrarAuditoriaProcessamento(ctx, tx, solicitacao, resultado); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	if err = gravarRespostaIdempotente(ctx, tx, solicitacao, resultado); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, err
	}
	return resultado, nil
}

func buscarRespostaIdempotente(ctx context.Context, tx pgx.Tx, solicitacao pecaApplication.SolicitacaoCompraReserva) (pecaApplication.ResultadoCompraReserva, bool, error) {
	var hash string
	var resposta []byte
	err := tx.QueryRow(ctx, `
		SELECT hash_requisicao, resposta
		FROM chave_idempotencia
		WHERE operacao = $1 AND chave = $2
		FOR UPDATE`, pecaApplication.OperacaoSolicitarCompraReservarPecas, solicitacao.IdempotencyKey,
	).Scan(&hash, &resposta)
	if errors.Is(err, pgx.ErrNoRows) {
		return pecaApplication.ResultadoCompraReserva{}, false, nil
	}
	if err != nil {
		return pecaApplication.ResultadoCompraReserva{}, false, err
	}
	if hash != solicitacao.HashRequisicao {
		return pecaApplication.ResultadoCompraReserva{}, true, pecaApplication.ErrIdempotencyKeyEmUso
	}
	var resultado pecaApplication.ResultadoCompraReserva
	if err = json.Unmarshal(resposta, &resultado); err != nil {
		return pecaApplication.ResultadoCompraReserva{}, false, err
	}
	resultado.Reprocessado = true
	return resultado, true, nil
}

func carregarFornecedor(ctx context.Context, tx pgx.Tx, fornecedorID string, resultado *pecaApplication.ResultadoCompraReserva) error {
	var ativo bool
	err := tx.QueryRow(ctx, `
		SELECT id, COALESCE(nome_fantasia, razao_social), ativo
		FROM fornecedor WHERE id = $1`, fornecedorID,
	).Scan(&resultado.Fornecedor.ID, &resultado.Fornecedor.Nome, &ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return pecaApplication.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return err
	}
	if !ativo {
		return pecaApplication.ErrFornecedorInativo
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
		return pecaApplication.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return err
	}
	if !possuiOrcamento {
		return pecaApplication.ErrOrdemServicoInvalida
	}
	return nil
}

func carregarItensProcessamento(ctx context.Context, tx pgx.Tx, solicitacao pecaApplication.SolicitacaoCompraReserva) ([]itemProcessamento, error) {
	quantidades := make(map[string]int64, len(solicitacao.Itens))
	ids := make([]string, 0, len(solicitacao.Itens))
	for _, item := range solicitacao.Itens {
		quantidades[item.ItemID] = item.Quantidade
		ids = append(ids, item.ItemID)
	}

	rows, err := tx.Query(ctx, `
		SELECT i.id, i.tipo, i.ativo, i.saldo_fisico, i.saldo_reservado, i.preco_venda::text,
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
			&item.precoVenda, &item.osItemID, &item.vinculado, &item.jaProcessado); err != nil {
			return nil, err
		}
		item.quantidade = quantidades[item.id]
		itens = append(itens, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(itens) != len(ids) {
		return nil, pecaApplication.ErrItemNaoEncontrado
	}
	return itens, nil
}

func garantirItemOrdemServico(ctx context.Context, tx pgx.Tx, ordemServicoID string, item itemProcessamento) (string, error) {
	if item.osItemID != nil {
		if _, err := tx.Exec(ctx, `
			UPDATE ordem_servico_item
			SET quantidade_necessaria = GREATEST(quantidade_necessaria, $2)
			WHERE id = $1`, *item.osItemID, item.quantidade); err != nil {
			return "", err
		}
		return *item.osItemID, nil
	}

	var osItemID string
	err := tx.QueryRow(ctx, `
		INSERT INTO ordem_servico_item (ordem_servico_id, item_estoque_id, quantidade_necessaria, valor_unitario)
		VALUES ($1, $2, $3, $4::NUMERIC)
		RETURNING id`, ordemServicoID, item.id, item.quantidade, valorTexto(item.precoVenda),
	).Scan(&osItemID)
	return osItemID, err
}

func reservarSaldoDisponivel(ctx context.Context, tx pgx.Tx, ordemServicoID, osItemID string, processamento peca.Processamento) (string, error) {
	if _, err := tx.Exec(ctx, `
		UPDATE item_estoque
		SET saldo_reservado = saldo_reservado + $2
		WHERE id = $1`, processamento.ItemID, processamento.QuantidadeReservada); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE ordem_servico_item
		SET quantidade_reservada = quantidade_reservada + $2
		WHERE id = $1`, osItemID, processamento.QuantidadeReservada); err != nil {
		return "", err
	}

	var reservaID string
	err := tx.QueryRow(ctx, `
		INSERT INTO reserva_estoque (ordem_servico_item_id, item_estoque_id, quantidade, status)
		VALUES ($1, $2, $3, 'ATIVA')
		RETURNING id`, osItemID, processamento.ItemID, processamento.QuantidadeReservada,
	).Scan(&reservaID)
	return reservaID, err
}

func registrarMovimentacaoReserva(ctx context.Context, tx pgx.Tx, ordemServicoID, reservaID string, processamento peca.Processamento) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade)
		VALUES ($1, $2, $3, 'RESERVA', $4)`,
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

func solicitarCompra(ctx context.Context, tx pgx.Tx, pedidoID, osItemID string, processamento peca.Processamento) error {
	var pedidoItemID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO pedido_compra_item (
			pedido_compra_id, item_estoque_id, quantidade_necessaria, quantidade_pedida, quantidade_reservada
		) VALUES ($1, $2, $3, $3, 0)
		RETURNING id`, pedidoID, processamento.ItemID, processamento.QuantidadeCompra,
	).Scan(&pedidoItemID); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO pedido_compra_item_os (pedido_compra_item_id, ordem_servico_item_id, quantidade_atendida)
		VALUES ($1, $2, $3)`, pedidoItemID, osItemID, processamento.QuantidadeCompra)
	return err
}

func registrarAuditoriaProcessamento(ctx context.Context, tx pgx.Tx, solicitacao pecaApplication.SolicitacaoCompraReserva, resultado pecaApplication.ResultadoCompraReserva) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, ocorrido_em)
		VALUES ($1, NULL, 'ESTOQUE', $1, 'PECAS_RESERVA_COMPRA_PROCESSADA', $2::jsonb, CURRENT_TIMESTAMP)`,
		solicitacao.OrdemServicoID,
		fmt.Sprintf(`{"fornecedorId":"%s","statusOrdemServico":"%s","pecasReservadas":%d,"pecasCompraSolicitada":%d}`,
			solicitacao.FornecedorID, resultado.StatusOrdemServico, len(resultado.PecasReservadas), len(resultado.PecasCompraSolicitada)))
	return err
}

func gravarRespostaIdempotente(ctx context.Context, tx pgx.Tx, solicitacao pecaApplication.SolicitacaoCompraReserva, resultado pecaApplication.ResultadoCompraReserva) error {
	resposta, err := json.Marshal(resultado)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO chave_idempotencia (chave, operacao, hash_requisicao, status_resposta, resposta)
		VALUES ($1, $2, $3, 201, $4::jsonb)`,
		solicitacao.IdempotencyKey, pecaApplication.OperacaoSolicitarCompraReservarPecas, solicitacao.HashRequisicao, string(resposta))
	return err
}

func valorTexto(valor *string) string {
	if valor == nil || strings.TrimSpace(*valor) == "" {
		return "0"
	}
	return strings.TrimSpace(*valor)
}

func valorParcial(preco *string, quantidade int64) float64 {
	valor, err := strconv.ParseFloat(valorTexto(preco), 64)
	if err != nil {
		return 0
	}
	return valor * float64(quantidade)
}

// Atualizar aplica as alteracoes sob lock otimista, em transacao: le a linha com
// FOR UPDATE para saber o preco anterior e a version corrente, atualiza, e grava o
// historico quando o preco muda — sem que exista preco novo sem o registro.
func (repository PostgresRepository) Atualizar(ctx context.Context, id string, version int, atualizacao peca.Atualizacao, usuarioID string) (peca.Peca, error) {
	transacao, err := repository.db.Begin(ctx)
	if err != nil {
		return peca.Peca{}, err
	}
	defer func() { _ = transacao.Rollback(ctx) }()

	var ativa bool
	err = transacao.QueryRow(ctx, `SELECT ativa FROM categoria WHERE id = $1`, atualizacao.CategoriaID).Scan(&ativa)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if !ativa {
		return peca.Peca{}, pecaApplication.ErrCategoriaInvalida
	}

	var precoAnterior *string
	var versionAtual int
	err = transacao.QueryRow(ctx, `
		SELECT preco_venda::text, version FROM item_estoque
		WHERE id = $1 AND tipo = 'PECA' AND ativo
		FOR UPDATE`, id).Scan(&precoAnterior, &versionAtual)
	if errors.Is(err, pgx.ErrNoRows) {
		return peca.Peca{}, pecaApplication.ErrNaoEncontrada
	}
	if err != nil {
		return peca.Peca{}, err
	}
	if versionAtual != version {
		return peca.Peca{}, pecaApplication.ErrVersaoDivergente
	}

	var dataAtualizacao time.Time
	var usuarioAtualizacao *string
	err = transacao.QueryRow(ctx, `
		UPDATE item_estoque
		SET nome = $2,
			descricao = $3,
			descricao_normalizada = $4,
			categoria_id = $5,
			fabricante = $6,
			preco_venda = $7::NUMERIC,
			estoque_minimo = $8,
			data_atualizacao = CURRENT_TIMESTAMP,
			usuario_atualizacao = NULLIF($9, '')::UUID,
			version = version + 1
		WHERE id = $1
		RETURNING data_atualizacao, usuario_atualizacao::text`,
		id, atualizacao.Nome, atualizacao.Descricao, atualizacao.DescricaoNormalizada,
		atualizacao.CategoriaID, atualizacao.Fabricante, atualizacao.PrecoVenda,
		atualizacao.EstoqueMinimo, usuarioID).Scan(&dataAtualizacao, &usuarioAtualizacao)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return peca.Peca{}, pecaApplication.ErrDescricaoDuplicada
		}
		return peca.Peca{}, err
	}

	if precoAnterior == nil || !mesmoPreco(*precoAnterior, atualizacao.PrecoVenda) {
		if _, err = transacao.Exec(ctx, `
			INSERT INTO historico_preco_item (item_estoque_id, preco_anterior, preco_novo, usuario_id)
			VALUES ($1, $2::NUMERIC, $3::NUMERIC, NULLIF($4, '')::UUID)`,
			id, precoAnterior, atualizacao.PrecoVenda, usuarioID); err != nil {
			return peca.Peca{}, err
		}
	}

	if err = transacao.Commit(ctx); err != nil {
		return peca.Peca{}, err
	}

	atualizada, err := repository.BuscarPorID(ctx, id)
	if err != nil {
		return peca.Peca{}, err
	}
	atualizada.DataAtualizacao = &dataAtualizacao
	atualizada.UsuarioAtualizacao = usuarioAtualizacao
	return atualizada, nil
}

// mesmoPreco compara valor, nao texto: o banco devolve "180.00" onde o cliente mandou
// "180", e um historico por diferenca de formatacao seria ruido.
func mesmoPreco(anterior, novo string) bool {
	valorAnterior, erroAnterior := strconv.ParseFloat(anterior, 64)
	valorNovo, erroNovo := strconv.ParseFloat(novo, 64)
	if erroAnterior != nil || erroNovo != nil {
		return anterior == novo
	}
	return valorAnterior == valorNovo
}
