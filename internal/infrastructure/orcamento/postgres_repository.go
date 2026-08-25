package orcamento

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (domain.Consulta, error) {
	const osQuery = `SELECT os.id FROM orcamento o JOIN ordem_servico os ON os.id = o.ordem_servico_id WHERE o.id = $1`
	var ordemServicoID string
	if err := repository.db.QueryRow(ctx, osQuery, id).Scan(&ordemServicoID); errors.Is(err, pgx.ErrNoRows) {
		return domain.Consulta{}, application.ErrOrcamentoNaoEncontrado
	} else if err != nil {
		return domain.Consulta{}, err
	}
	return repository.buscarOrdem(ctx, ordemServicoID)
}

func (repository PostgresRepository) BuscarPorDocumento(ctx context.Context, documento string, offset, limit int) ([]domain.Consulta, int, error) {
	var clienteExiste bool
	if err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cliente WHERE documento = $1)`, documento).Scan(&clienteExiste); err != nil {
		return nil, 0, err
	}
	if !clienteExiste {
		return nil, 0, application.ErrClienteNaoEncontrado
	}
	var total int
	if err := repository.db.QueryRow(ctx, `SELECT COUNT(DISTINCT os.id) FROM ordem_servico os JOIN cliente c ON c.id = os.cliente_id JOIN orcamento o ON o.ordem_servico_id = os.id WHERE c.documento = $1`, documento).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.db.Query(ctx, `SELECT DISTINCT os.id, os.criada_em FROM ordem_servico os JOIN cliente c ON c.id = os.cliente_id JOIN orcamento o ON o.ordem_servico_id = os.id WHERE c.documento = $1 ORDER BY os.criada_em DESC, os.id LIMIT $2 OFFSET $3`, documento, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	ids := make([]string, 0, limit)
	for rows.Next() {
		var id string
		var ignored any
		if err := rows.Scan(&id, &ignored); err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	data := make([]domain.Consulta, 0, len(ids))
	for _, id := range ids {
		consulta, err := repository.buscarOrdem(ctx, id)
		if err != nil {
			return nil, 0, err
		}
		data = append(data, consulta)
	}
	return data, total, nil
}

func (repository PostgresRepository) buscarOrdem(ctx context.Context, id string) (domain.Consulta, error) {
	const query = `SELECT c.id, c.documento, os.id, os.status,
		o.id, o.tipo_orcamento, COALESCE(o.orcamento_original_id::text, ''), o.status,
		o.estimativa_entrega_dias, o.criado_em,
		COALESCE(oi.tipo_item, ''), COALESCE(oi.descricao, ''), COALESCE(oi.quantidade, 0),
		COALESCE(oi.valor_unitario, 0), COALESCE(oi.valor_total, 0)
		FROM ordem_servico os
		JOIN cliente c ON c.id = os.cliente_id
		JOIN orcamento o ON o.ordem_servico_id = os.id
		LEFT JOIN orcamento_item oi ON oi.orcamento_id = o.id
		WHERE os.id = $1
		ORDER BY CASE WHEN o.tipo_orcamento = 'PRINCIPAL' THEN 0 ELSE 1 END, o.criado_em, oi.id`
	rows, err := repository.db.Query(ctx, query, id)
	if err != nil {
		return domain.Consulta{}, err
	}
	defer rows.Close()
	var result domain.Consulta
	index := map[string]int{}
	for rows.Next() {
		var budget domain.Orcamento
		var item domain.Item
		if err := rows.Scan(&result.Cliente.ID, &result.Cliente.Documento, &result.OrdemServicoID, &result.StatusOrdemServico,
			&budget.ID, &budget.Tipo, &budget.OrcamentoOriginalID, &budget.Status, &budget.EstimativaEntregaDias, &budget.DataGeracao,
			&item.Tipo, &item.Descricao, &item.Quantidade, &item.ValorUnitario, &item.ValorTotal); err != nil {
			return domain.Consulta{}, err
		}
		position, exists := index[budget.ID]
		if !exists {
			budget.Itens = []domain.Item{}
			position = len(result.Orcamentos)
			index[budget.ID] = position
			result.Orcamentos = append(result.Orcamentos, budget)
		}
		if item.Tipo != "" {
			result.Orcamentos[position].Itens = append(result.Orcamentos[position].Itens, item)
			result.Orcamentos[position].ValorTotal += item.ValorTotal
			result.ValorTotalGeral += item.ValorTotal
		}
	}
	if err := rows.Err(); err != nil {
		return domain.Consulta{}, err
	}
	if result.OrdemServicoID == "" {
		return domain.Consulta{}, application.ErrOrcamentoNaoEncontrado
	}
	return result, nil
}
