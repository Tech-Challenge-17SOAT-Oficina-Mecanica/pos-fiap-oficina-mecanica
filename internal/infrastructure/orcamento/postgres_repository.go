package orcamento

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

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
