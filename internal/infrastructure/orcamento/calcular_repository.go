package orcamento

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	orcamentoApplication "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

// BuscarParaCalculo carrega o orcamento alvo com seus itens e devolve tambem a OS dona,
// que e o escopo do total geral.
func (repository PostgresRepository) BuscarParaCalculo(ctx context.Context, orcamentoID string) (orcamento.Orcamento, string, error) {
	var alvo orcamento.Orcamento
	var ordemServicoID string
	var originalID *string

	err := repository.db.QueryRow(ctx, `
		SELECT id, ordem_servico_id, orcamento_original_id::text, tipo_orcamento, status
		FROM orcamento WHERE id = $1`, orcamentoID).
		Scan(&alvo.ID, &ordemServicoID, &originalID, &alvo.Tipo, &alvo.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return orcamento.Orcamento{}, "", orcamentoApplication.ErrOrcamentoNaoEncontrado
	}
	if err != nil {
		return orcamento.Orcamento{}, "", err
	}
	if originalID != nil {
		alvo.OriginalID = *originalID
	}

	alvo.Itens, err = repository.itensDe(ctx, orcamentoID)
	if err != nil {
		return orcamento.Orcamento{}, "", err
	}
	return alvo, ordemServicoID, nil
}

// OrcamentosDaOrdem devolve os orcamentos da OS com seus itens. Quem decide quais entram
// no total e o caso de uso; aqui so se carrega o que existe.
func (repository PostgresRepository) OrcamentosDaOrdem(ctx context.Context, ordemServicoID string) ([]orcamentoApplication.OrcamentoDaOS, error) {
	linhas, err := repository.db.Query(ctx, `
		SELECT id, tipo_orcamento, status FROM orcamento
		WHERE ordem_servico_id = $1 ORDER BY criado_em`, ordemServicoID)
	if err != nil {
		return nil, err
	}

	var orcamentos []orcamentoApplication.OrcamentoDaOS
	for linhas.Next() {
		var atual orcamentoApplication.OrcamentoDaOS
		if err := linhas.Scan(&atual.ID, &atual.Tipo, &atual.Status); err != nil {
			linhas.Close()
			return nil, err
		}
		orcamentos = append(orcamentos, atual)
	}
	linhas.Close()
	if err := linhas.Err(); err != nil {
		return nil, err
	}

	for indice := range orcamentos {
		if orcamentos[indice].Itens, err = repository.itensDe(ctx, orcamentos[indice].ID); err != nil {
			return nil, err
		}
	}
	return orcamentos, nil
}

// SalvarItens grava os valores recalculados. Roda em transacao para que o orcamento nao
// fique com parte dos itens atualizada, e nao toca no status — recalcular nao e decidir.
func (repository PostgresRepository) SalvarItens(ctx context.Context, orcamentoID string, itens []orcamento.Item) error {
	transacao, err := repository.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = transacao.Rollback(ctx) }()

	for _, item := range itens {
		if _, err = transacao.Exec(ctx, `
			UPDATE orcamento_item
			SET quantidade = $2, valor_unitario = $3, valor_total = $4
			WHERE id = $1`, item.ID, item.Quantidade, item.ValorUnitario, item.ValorTotal); err != nil {
			return err
		}
	}

	if _, err = transacao.Exec(ctx, `
		UPDATE orcamento SET data_atualizacao = CURRENT_TIMESTAMP WHERE id = $1`, orcamentoID); err != nil {
		return err
	}
	return transacao.Commit(ctx)
}

func (repository PostgresRepository) itensDe(ctx context.Context, orcamentoID string) ([]orcamento.Item, error) {
	linhas, err := repository.db.Query(ctx, `
		SELECT id, tipo_item, descricao, quantidade, valor_unitario, valor_total
		FROM orcamento_item WHERE orcamento_id = $1 ORDER BY id`, orcamentoID)
	if err != nil {
		return nil, err
	}
	defer linhas.Close()

	var itens []orcamento.Item
	for linhas.Next() {
		var item orcamento.Item
		if err := linhas.Scan(&item.ID, &item.Tipo, &item.Descricao,
			&item.Quantidade, &item.ValorUnitario, &item.ValorTotal); err != nil {
			return nil, err
		}
		itens = append(itens, item)
	}
	return itens, linhas.Err()
}
