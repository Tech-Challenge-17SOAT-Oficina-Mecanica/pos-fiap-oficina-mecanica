package orcamento

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

type itemAprovacao struct {
	id             string
	tipo           string
	ativo          bool
	quantidade     string
	saldoFisico    string
	saldoReservado string
	valorUnitario  string
	osItemID       *string
}

func (repository PostgresRepository) Aprovar(ctx context.Context, input application.AprovarInput) (domain.Aprovacao, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.Aprovacao{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var resultado domain.Aprovacao
	var statusOrcamento, statusOS string
	err = tx.QueryRow(ctx, `
		SELECT o.id, o.ordem_servico_id, o.tipo_orcamento, COALESCE(o.orcamento_original_id::text, ''),
		       o.status, os.status, os.cliente_id::text
		FROM orcamento o
		JOIN ordem_servico os ON os.id = o.ordem_servico_id
		WHERE o.id = $1
		FOR UPDATE OF o, os`, input.OrcamentoID,
	).Scan(&resultado.OrcamentoID, &resultado.OrdemServicoID, &resultado.TipoOrcamento,
		&resultado.OrcamentoOriginalID, &statusOrcamento, &statusOS, &resultado.ClienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Aprovacao{}, application.ErrOrcamentoNaoEncontrado
	}
	if err != nil {
		return domain.Aprovacao{}, err
	}
	if resultado.ClienteID != input.ClienteID || (input.OrdemServicoID != "" && input.OrdemServicoID != resultado.OrdemServicoID) {
		return domain.Aprovacao{}, application.ErrAcessoNegado
	}
	if statusOS != "AGUARDANDO_APROVACAO" {
		return domain.Aprovacao{}, application.ErrOrdemServicoStatusInvalido
	}
	if statusOrcamento != "CRIADO" {
		return domain.Aprovacao{}, application.ErrOrcamentoJaDecidido
	}
	if resultado.TipoOrcamento != "PRINCIPAL" && resultado.TipoOrcamento != "COMPLEMENTAR" {
		return domain.Aprovacao{}, application.ErrOrcamentoJaDecidido
	}
	if resultado.TipoOrcamento == "COMPLEMENTAR" && !orcamentoOriginalValido(ctx, tx, resultado.OrdemServicoID, resultado.OrcamentoOriginalID) {
		return domain.Aprovacao{}, application.ErrOrcamentoComplementarSemPai
	}
	if err = validarFornecedor(ctx, tx, input.FornecedorID); err != nil {
		return domain.Aprovacao{}, err
	}

	itens, err := carregarItensAprovacao(ctx, tx, resultado.OrcamentoID, resultado.OrdemServicoID)
	if err != nil {
		return domain.Aprovacao{}, err
	}
	compraSolicitada, err := processarItensAprovados(ctx, tx, resultado.OrdemServicoID, input.FornecedorID, itens)
	if err != nil {
		return domain.Aprovacao{}, err
	}

	resultado.StatusOrcamento = "APROVADO"
	resultado.StatusOrdemServico = "AGUARDANDO_EXECUCAO"
	if compraSolicitada {
		resultado.StatusOrdemServico = "AGUARDANDO_RECURSOS"
	}
	resultado.FornecedorID = input.FornecedorID
	resultado.PossuiCompraSolicitada = compraSolicitada
	err = tx.QueryRow(ctx, `
		UPDATE orcamento
		SET status = 'APROVADO',
		    aprovado_em = CURRENT_TIMESTAMP,
		    cliente_aprovador_id = $2,
		    data_atualizacao = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING aprovado_em`, resultado.OrcamentoID, input.ClienteID,
	).Scan(&resultado.DataAprovacao)
	if err != nil {
		return domain.Aprovacao{}, err
	}
	if _, err = tx.Exec(ctx, `UPDATE ordem_servico SET status = $2 WHERE id = $1`, resultado.OrdemServicoID, resultado.StatusOrdemServico); err != nil {
		return domain.Aprovacao{}, err
	}
	if _, err = tx.Exec(ctx, `
		INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, ocorrido_em)
		VALUES ($1, NULL, 'ORCAMENTO', $2, 'ORCAMENTO_APROVADO', $3::jsonb, CURRENT_TIMESTAMP)`,
		resultado.OrdemServicoID, resultado.OrcamentoID,
		fmt.Sprintf(`{"clienteId":"%s","statusOrdemServico":"%s"}`, input.ClienteID, resultado.StatusOrdemServico)); err != nil {
		return domain.Aprovacao{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.Aprovacao{}, err
	}
	return resultado, nil
}

func orcamentoOriginalValido(ctx context.Context, tx pgx.Tx, ordemServicoID, originalID string) bool {
	if originalID == "" {
		return false
	}
	var existe bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM orcamento
			WHERE id = $1 AND ordem_servico_id = $2 AND tipo_orcamento = 'PRINCIPAL'
		)`, originalID, ordemServicoID,
	).Scan(&existe)
	return err == nil && existe
}

func validarFornecedor(ctx context.Context, tx pgx.Tx, fornecedorID string) error {
	var ativo bool
	err := tx.QueryRow(ctx, `SELECT ativo FROM fornecedor WHERE id = $1`, fornecedorID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return err
	}
	if !ativo {
		return application.ErrFornecedorInativo
	}
	return nil
}

func carregarItensAprovacao(ctx context.Context, tx pgx.Tx, orcamentoID, ordemServicoID string) ([]itemAprovacao, error) {
	rows, err := tx.Query(ctx, `
		SELECT i.id::text, i.tipo, i.ativo, oi.quantidade::text, i.saldo_fisico::text,
		       i.saldo_reservado::text, oi.valor_unitario::text, osi.id::text
		FROM orcamento_item oi
		JOIN item_estoque i ON i.id = oi.item_estoque_id
		LEFT JOIN ordem_servico_item osi ON osi.ordem_servico_id = $2 AND osi.item_estoque_id = i.id
		WHERE oi.orcamento_id = $1 AND oi.item_estoque_id IS NOT NULL
		ORDER BY i.id
		FOR UPDATE OF i`, orcamentoID, ordemServicoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var itens []itemAprovacao
	for rows.Next() {
		var item itemAprovacao
		if err := rows.Scan(&item.id, &item.tipo, &item.ativo, &item.quantidade, &item.saldoFisico,
			&item.saldoReservado, &item.valorUnitario, &item.osItemID); err != nil {
			return nil, err
		}
		if !item.ativo || (item.tipo != "PECA" && item.tipo != "INSUMO") {
			return nil, application.ErrOrcamentoJaDecidido
		}
		itens = append(itens, item)
	}
	return itens, rows.Err()
}

func processarItensAprovados(ctx context.Context, tx pgx.Tx, ordemServicoID, fornecedorID string, itens []itemAprovacao) (bool, error) {
	var pedidoID string
	compraSolicitada := false
	for _, item := range itens {
		osItemID, err := garantirItemOrdemServico(ctx, tx, ordemServicoID, item)
		if err != nil {
			return false, err
		}
		disponivel := subtrairDecimal(item.saldoFisico, item.saldoReservado)
		reservar := menorDecimal(item.quantidade, disponivel)
		if compararDecimal(reservar, "0") < 0 {
			reservar = "0"
		}
		comprar := subtrairDecimal(item.quantidade, reservar)
		if compararDecimal(reservar, "0") > 0 {
			reservaID, err := reservarItem(ctx, tx, ordemServicoID, osItemID, item.id, reservar)
			if err != nil {
				return false, err
			}
			if err = registrarMovimentacaoReserva(ctx, tx, ordemServicoID, reservaID, item.id, reservar); err != nil {
				return false, err
			}
		}
		if compararDecimal(comprar, "0") > 0 {
			if pedidoID == "" {
				pedidoID, err = criarPedidoCompra(ctx, tx, fornecedorID)
				if err != nil {
					return false, err
				}
			}
			if err = solicitarCompra(ctx, tx, pedidoID, osItemID, item, comprar); err != nil {
				return false, err
			}
			compraSolicitada = true
		}
	}
	return compraSolicitada, nil
}

func garantirItemOrdemServico(ctx context.Context, tx pgx.Tx, ordemServicoID string, item itemAprovacao) (string, error) {
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
		RETURNING id`, ordemServicoID, item.id, item.quantidade, item.valorUnitario,
	).Scan(&osItemID)
	return osItemID, err
}

func reservarItem(ctx context.Context, tx pgx.Tx, ordemServicoID, osItemID, itemID, quantidade string) (string, error) {
	if _, err := tx.Exec(ctx, `
		UPDATE item_estoque SET saldo_reservado = saldo_reservado + $2::NUMERIC
		WHERE id = $1`, itemID, quantidade); err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE ordem_servico_item SET quantidade_reservada = quantidade_reservada + $2::NUMERIC
		WHERE id = $1`, osItemID, quantidade); err != nil {
		return "", err
	}
	var reservaID string
	err := tx.QueryRow(ctx, `
		INSERT INTO reserva_estoque (ordem_servico_item_id, item_estoque_id, quantidade, status)
		VALUES ($1, $2, $3::NUMERIC, 'ATIVA')
		RETURNING id`, osItemID, itemID, quantidade,
	).Scan(&reservaID)
	return reservaID, err
}

func registrarMovimentacaoReserva(ctx context.Context, tx pgx.Tx, ordemServicoID, reservaID, itemID, quantidade string) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO movimentacao_estoque (item_estoque_id, ordem_servico_id, reserva_estoque_id, tipo, quantidade)
		VALUES ($1, $2, $3, 'RESERVA', $4::NUMERIC)`, itemID, ordemServicoID, reservaID, quantidade)
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

func solicitarCompra(ctx context.Context, tx pgx.Tx, pedidoID, osItemID string, item itemAprovacao, quantidade string) error {
	var pedidoItemID string
	if err := tx.QueryRow(ctx, `
		INSERT INTO pedido_compra_item (
			pedido_compra_id, item_estoque_id, quantidade_necessaria, quantidade_pedida, quantidade_reservada, custo_unitario
		) VALUES ($1, $2, $3::NUMERIC, $3::NUMERIC, $3::NUMERIC, $4::NUMERIC)
		RETURNING id`, pedidoID, item.id, quantidade, item.valorUnitario,
	).Scan(&pedidoItemID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		INSERT INTO pedido_compra_item_os (pedido_compra_item_id, ordem_servico_item_id, quantidade_atendida)
		VALUES ($1, $2, $3::NUMERIC)`, pedidoItemID, osItemID, quantidade); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
		UPDATE ordem_servico_item SET quantidade_reservada = quantidade_reservada + $2::NUMERIC
		WHERE id = $1`, osItemID, quantidade); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO reserva_estoque (ordem_servico_item_id, item_estoque_id, pedido_compra_item_id, quantidade, status)
		VALUES ($1, $2, $3, $4::NUMERIC, 'ATIVA')`, osItemID, item.id, pedidoItemID, quantidade)
	return err
}

func menorDecimal(a, b string) string {
	if compararDecimal(a, b) <= 0 {
		return normalizarDecimal(decimal(a).FloatString(3))
	}
	return normalizarDecimal(decimal(b).FloatString(3))
}

func subtrairDecimal(a, b string) string {
	resultado := decimal(a)
	resultado.Sub(resultado, decimal(b))
	return normalizarDecimal(resultado.FloatString(3))
}

func compararDecimal(a, b string) int {
	return decimal(a).Cmp(decimal(b))
}

func decimal(valor string) *big.Rat {
	numero, ok := new(big.Rat).SetString(strings.TrimSpace(valor))
	if !ok {
		return new(big.Rat)
	}
	return numero
}

func normalizarDecimal(valor string) string {
	valor = strings.TrimRight(valor, "0")
	valor = strings.TrimRight(valor, ".")
	if valor == "" || valor == "-0" {
		return "0"
	}
	return valor
}

func (repository PostgresRepository) Consultar(ctx context.Context, ordemServicoID, clienteID string) (domain.Consulta, error) {
	consulta := domain.Consulta{OrdemServicoID: ordemServicoID, Orcamentos: []domain.Orcamento{}}
	err := repository.db.QueryRow(ctx, `
		SELECT c.id, c.nome, c.documento, c.tipo_documento, os.status
		FROM ordem_servico os JOIN cliente c ON c.id = os.cliente_id
		WHERE os.id = $1`, ordemServicoID,
	).Scan(&consulta.Cliente.ID, &consulta.Cliente.Nome, &consulta.Cliente.Documento, &consulta.Cliente.TipoDocumento, &consulta.StatusOrdemServico)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Consulta{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return domain.Consulta{}, err
	}
	if clienteID != "" && clienteID != consulta.Cliente.ID {
		return domain.Consulta{}, application.ErrAcessoNegado
	}
	rows, err := repository.db.Query(ctx, `
		SELECT id, COALESCE(orcamento_original_id::text, ''), tipo_orcamento, status,
		       estimativa_entrega_dias, criado_em
		FROM orcamento
		WHERE ordem_servico_id = $1
		ORDER BY CASE tipo_orcamento WHEN 'PRINCIPAL' THEN 0 ELSE 1 END, criado_em`, ordemServicoID)
	if err != nil {
		return domain.Consulta{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var budget domain.Orcamento
		var estimativa pgtype.Int4
		if err := rows.Scan(&budget.ID, &budget.OriginalID, &budget.Tipo, &budget.Status, &estimativa, &budget.DataGeracao); err != nil {
			return domain.Consulta{}, err
		}
		if estimativa.Valid {
			value := int(estimativa.Int32)
			budget.EstimativaDias = &value
		}
		budget.Itens, budget.ValorTotal, err = itensDoOrcamento(ctx, repository.db, budget.ID)
		if err != nil {
			return domain.Consulta{}, err
		}
		budget.Problemas, err = problemasDoOrcamento(ctx, repository.db, budget.ID)
		if err != nil {
			return domain.Consulta{}, err
		}
		consulta.ValorTotalGeral += budget.ValorTotal
		consulta.Orcamentos = append(consulta.Orcamentos, budget)
	}
	if err := rows.Err(); err != nil {
		return domain.Consulta{}, err
	}
	if len(consulta.Orcamentos) == 0 {
		return domain.Consulta{}, application.ErrOrcamentoNaoEncontrado
	}
	if len(consulta.Orcamentos) > 1 {
		estimativa := 0
		for _, budget := range consulta.Orcamentos {
			if budget.EstimativaDias == nil {
				return consulta, nil
			}
			estimativa += *budget.EstimativaDias
		}
		consulta.EstimativaEntregaDias = &estimativa
	}
	return consulta, nil
}

type rowQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

func itensDoOrcamento(ctx context.Context, queryer rowQueryer, orcamentoID string) ([]domain.Item, float64, error) {
	rows, err := queryer.Query(ctx, `
		SELECT tipo_item, descricao, quantidade, valor_unitario, valor_total
		FROM orcamento_item WHERE orcamento_id = $1 ORDER BY id`, orcamentoID)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	itens := []domain.Item{}
	var total float64
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.Tipo, &item.Descricao, &item.Quantidade, &item.ValorUnitario, &item.ValorTotal); err != nil {
			return nil, 0, err
		}
		total += item.ValorTotal
		itens = append(itens, item)
	}
	return itens, total, rows.Err()
}

func problemasDoOrcamento(ctx context.Context, queryer rowQueryer, orcamentoID string) ([]domain.Problema, error) {
	rows, err := queryer.Query(ctx, `
		SELECT id, descricao, COALESCE(observacoes, ''), registrado_em
		FROM problema_ordem_servico WHERE orcamento_id = $1 ORDER BY registrado_em`, orcamentoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	problemas := []domain.Problema{}
	for rows.Next() {
		var problema domain.Problema
		if err := rows.Scan(&problema.ID, &problema.Descricao, &problema.Observacoes, &problema.RegistradoEm); err != nil {
			return nil, err
		}
		problemas = append(problemas, problema)
	}
	return problemas, rows.Err()
}
