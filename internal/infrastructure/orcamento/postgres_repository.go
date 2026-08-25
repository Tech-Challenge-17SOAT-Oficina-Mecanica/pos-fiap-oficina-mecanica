package orcamento

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/orcamento"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) Calcular(ctx context.Context, orcamentoID, usuarioID string) (json.Number, error) {
	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return "", err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	calculo, ordemServicoID, err := carregarCalculo(ctx, tx, orcamentoID)
	if err != nil {
		return "", err
	}
	estimativa, err := calculo.EstimativaEntregaDias()
	if err != nil {
		return "", err
	}
	if _, err := tx.Exec(ctx, `UPDATE orcamento_item SET valor_total = ROUND(quantidade * valor_unitario, 2) WHERE orcamento_id = $1`, orcamentoID); err != nil {
		return "", fmt.Errorf("%w: atualizar itens: %v", application.ErrFalhaPersistencia, err)
	}
	if _, err := tx.Exec(ctx, `UPDATE orcamento SET estimativa_entrega_dias = $2, data_atualizacao = CURRENT_TIMESTAMP WHERE id = $1`, orcamentoID, estimativa); err != nil {
		return "", fmt.Errorf("%w: atualizar orçamento: %v", application.ErrFalhaPersistencia, err)
	}
	var total string
	if err := tx.QueryRow(ctx, `SELECT COALESCE(SUM(oi.valor_total), 0)::text
		FROM orcamento o JOIN orcamento_item oi ON oi.orcamento_id = o.id
		WHERE o.ordem_servico_id = $1 AND o.status IN ('CRIADO', 'APROVADO')`, ordemServicoID).Scan(&total); err != nil {
		return "", fmt.Errorf("%w: calcular total geral: %v", application.ErrFalhaPersistencia, err)
	}
	if _, err := tx.Exec(ctx, `INSERT INTO auditoria_ordem_servico
		(ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em)
		VALUES ($1, NULLIF($2, '')::uuid, 'ORCAMENTO', $3, 'ORCAMENTO_CALCULADO',
			jsonb_build_object('estimativaEntregaDias', $4::integer, 'valorTotalGeral', $5::numeric),
			jsonb_build_object('origem', 'api'), CURRENT_TIMESTAMP)`, ordemServicoID, usuarioID, orcamentoID, estimativa, total); err != nil {
		return "", fmt.Errorf("%w: registrar auditoria: %v", application.ErrFalhaPersistencia, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("%w: confirmar transação: %v", application.ErrFalhaPersistencia, err)
	}
	return json.Number(total), nil
}

func carregarCalculo(ctx context.Context, tx pgx.Tx, orcamentoID string) (domain.Calculo, string, error) {
	const budgetQuery = `SELECT o.ordem_servico_id, o.tipo_orcamento, o.status,
		COALESCE(o.orcamento_original_id::text, ''), COALESCE(principal.id::text, ''), principal.estimativa_entrega_dias
		FROM orcamento o
		LEFT JOIN orcamento principal ON principal.ordem_servico_id = o.ordem_servico_id AND principal.tipo_orcamento = 'PRINCIPAL'
		WHERE o.id = $1 FOR UPDATE OF o`
	var calculo domain.Calculo
	var ordemServicoID string
	if err := tx.QueryRow(ctx, budgetQuery, orcamentoID).Scan(&ordemServicoID, &calculo.Tipo, &calculo.Status, &calculo.OrcamentoOriginalID, &calculo.OrcamentoPrincipalID, &calculo.EstimativaPrincipalDias); errors.Is(err, pgx.ErrNoRows) {
		return calculo, "", application.ErrOrcamentoNaoEncontrado
	} else if err != nil {
		return calculo, "", err
	}
	if err := tx.QueryRow(ctx, `SELECT capacidade_diaria_os, minutos_produtivos_dia FROM configuracao_oficina WHERE id = 1`).Scan(&calculo.CapacidadeDiariaOS, &calculo.MinutosProdutivosDia); errors.Is(err, pgx.ErrNoRows) {
		return calculo, "", domain.ErrConfiguracaoInvalida
	} else if err != nil {
		return calculo, "", err
	}
	if err := tx.QueryRow(ctx, `SELECT COUNT(*) FROM ordem_servico WHERE status = 'AGUARDANDO_EXECUCAO' AND id <> $1`, ordemServicoID).Scan(&calculo.OrdensNaFila); err != nil {
		return calculo, "", err
	}
	rows, err := tx.Query(ctx, `SELECT oi.tipo_item, oi.quantidade::text, oi.valor_unitario::text,
		COALESCE(s.tempo_estimado_minutos, 0),
		CASE WHEN ie.id IS NULL THEN TRUE ELSE (ie.saldo_fisico - ie.saldo_reservado) >= oi.quantidade END,
		CASE WHEN ie.id IS NULL OR (ie.saldo_fisico - ie.saldo_reservado) >= oi.quantidade THEN NULL ELSE (
			SELECT MAX(f.prazo_entrega_dias) FROM pedido_compra_item pci
			JOIN pedido_compra pc ON pc.id = pci.pedido_compra_id
			JOIN fornecedor f ON f.id = pc.fornecedor_id
			WHERE pci.item_estoque_id = ie.id AND pc.status IN ('ABERTO', 'PARCIAL') AND f.ativo
		) END
		FROM orcamento_item oi
		LEFT JOIN servico s ON s.id = oi.servico_id
		LEFT JOIN item_estoque ie ON ie.id = oi.item_estoque_id
		WHERE oi.orcamento_id = $1 ORDER BY oi.id FOR UPDATE OF oi`, orcamentoID)
	if err != nil {
		return calculo, "", err
	}
	defer rows.Close()
	for rows.Next() {
		var item domain.Item
		if err := rows.Scan(&item.Tipo, &item.Quantidade, &item.ValorUnitario, &item.TempoServicoMinutos, &item.MaterialDisponivel, &item.PrazoEntregaDias); err != nil {
			return calculo, "", err
		}
		calculo.Itens = append(calculo.Itens, item)
	}
	if err := rows.Err(); err != nil {
		return calculo, "", fmt.Errorf("consultar itens: %w", err)
	}
	return calculo, ordemServicoID, nil
}
