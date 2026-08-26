package ordemservico

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domainEstoque "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/estoque"
	domainOrcamento "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/orcamento"
	domainOS "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) RegistrarItens(ctx context.Context, input application.RegistrarInput) (domainOrcamento.Resultado, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", input.OSID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainOrcamento.Resultado{}, application.ErrOSNaoEncontrada
	}
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}
	if !domainOS.PermiteRegistroDeItens(status) {
		return domainOrcamento.Resultado{}, domainOS.ErrStatusNaoPermiteItens
	}

	budgetID, originalID, budgetType, budgetStatus, err := repository.resolveBudget(ctx, tx, input.OSID, status)
	if err != nil {
		return domainOrcamento.Resultado{}, err
	}

	result := domainOrcamento.Resultado{OrdemServicoID: input.OSID, OrcamentoID: budgetID, OrcamentoOriginal: originalID, TipoOrcamento: budgetType, StatusOrcamento: budgetStatus, RegistradoPor: input.UsuarioID}
	for _, requested := range input.Itens {
		item, err := repository.loadItem(ctx, tx, requested.ItemID)
		if err != nil {
			return domainOrcamento.Resultado{}, err
		}
		if !item.ativo {
			return domainOrcamento.Resultado{}, application.ErrItemInativo
		}
		if item.tipo != input.Tipo {
			return domainOrcamento.Resultado{}, domainEstoque.ErrTipoItemInvalido
		}
		if item.valor == nil {
			return domainOrcamento.Resultado{}, application.ErrItemSemValor
		}

		itemValue := *item.valor
		if _, err = tx.Exec(ctx, `
			INSERT INTO ordem_servico_item (ordem_servico_id, item_estoque_id, quantidade_necessaria, valor_unitario)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (ordem_servico_id, item_estoque_id)
			DO UPDATE SET quantidade_necessaria = ordem_servico_item.quantidade_necessaria + EXCLUDED.quantidade_necessaria,
			              valor_unitario = EXCLUDED.valor_unitario`, input.OSID, requested.ItemID, requested.Quantidade, itemValue); err != nil {
			return domainOrcamento.Resultado{}, fmt.Errorf("inserir item da os: %w", err)
		}
		valueTotal := requested.Quantidade * itemValue
		if _, err = tx.Exec(ctx, `
			INSERT INTO orcamento_item (orcamento_id, item_estoque_id, tipo_item, descricao, quantidade, valor_unitario, valor_total)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`, budgetID, requested.ItemID, input.Tipo, item.descricao, requested.Quantidade, itemValue, valueTotal); err != nil {
			return domainOrcamento.Resultado{}, fmt.Errorf("inserir item do orcamento: %w", err)
		}
		result.ItensRegistrados = append(result.ItensRegistrados, domainOrcamento.Item{ItemID: requested.ItemID, Codigo: item.codigo, Descricao: item.descricao, Tipo: item.tipo, Quantidade: requested.Quantidade, ValorUnitario: itemValue, ValorItem: valueTotal})
	}

	if err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(valor_total), 0) FROM orcamento_item WHERE orcamento_id = $1", budgetID).Scan(&result.ValorOrcamento); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("calcular valor do orcamento: %w", err)
	}
	if err = tx.QueryRow(ctx, `SELECT COALESCE(SUM(oi.valor_total), 0) FROM orcamento_item oi JOIN orcamento o ON o.id = oi.orcamento_id WHERE o.ordem_servico_id = $1`, input.OSID).Scan(&result.ValorTotalGeral); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("calcular total da os: %w", err)
	}
	if _, err = tx.Exec(ctx, `INSERT INTO auditoria_ordem_servico (ordem_servico_id, usuario_id, agregado, agregado_id, tipo_evento, dados, metadados, ocorrido_em) VALUES ($1, NULLIF($2, '')::uuid, 'ORCAMENTO', $3, 'ITENS_REGISTRADOS', jsonb_build_object('tipoItem', $4::text, 'quantidadeItens', $5::integer), '{}'::jsonb, CURRENT_TIMESTAMP)`, input.OSID, input.UsuarioID, budgetID, input.Tipo, len(input.Itens)); err != nil {
		return domainOrcamento.Resultado{}, fmt.Errorf("registrar auditoria: %w", err)
	}
	if err = tx.Commit(ctx); err != nil {
		return domainOrcamento.Resultado{}, err
	}
	return result, nil
}

func (repository PostgresRepository) resolveBudget(ctx context.Context, tx pgx.Tx, osID, status string) (string, string, string, string, error) {
	var id, original, budgetType, budgetStatus string
	err := tx.QueryRow(ctx, "SELECT id, COALESCE(orcamento_original_id::text, ''), tipo_orcamento, status FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = 'PRINCIPAL'", osID).Scan(&id, &original, &budgetType, &budgetStatus)
	if errors.Is(err, pgx.ErrNoRows) {
		if domainOS.EhComplementar(status) {
			return "", "", "", "", application.ErrOrcamentoNaoEncontrado
		}
		err = tx.QueryRow(ctx, "INSERT INTO orcamento (ordem_servico_id, tipo_orcamento, status) VALUES ($1, 'PRINCIPAL', 'CRIADO') RETURNING id", osID).Scan(&id)
		return id, "", domainOrcamento.TipoPrincipal, domainOrcamento.StatusCriado, err
	}
	if err != nil {
		return "", "", "", "", err
	}
	if domainOS.EhComplementar(status) {
		if budgetStatus != "APROVADO" {
			return "", "", "", "", application.ErrOrcamentoAprovado
		}
		principalID := id
		err = tx.QueryRow(ctx, "INSERT INTO orcamento (ordem_servico_id, orcamento_original_id, tipo_orcamento, status) VALUES ($1, $2, 'COMPLEMENTAR', 'CRIADO') RETURNING id", osID, principalID).Scan(&id)
		return id, principalID, domainOrcamento.TipoComplementar, domainOrcamento.StatusCriado, err
	}
	if budgetStatus == "APROVADO" {
		return "", "", "", "", application.ErrOrcamentoAprovado
	}
	return id, original, budgetType, budgetStatus, nil
}

type itemRow struct {
	tipo, codigo, descricao string
	ativo                   bool
	valor                   *float64
}

func (repository PostgresRepository) loadItem(ctx context.Context, tx pgx.Tx, itemID string) (itemRow, error) {
	var item itemRow
	var preco, custo sql.NullFloat64
	err := tx.QueryRow(ctx, "SELECT tipo, codigo, descricao, ativo, preco_venda, custo_unitario FROM item_estoque WHERE id = $1", itemID).Scan(&item.tipo, &item.codigo, &item.descricao, &item.ativo, &preco, &custo)
	if errors.Is(err, pgx.ErrNoRows) {
		return itemRow{}, application.ErrItemNaoEncontrado
	}
	if err != nil {
		return itemRow{}, err
	}
	if item.tipo == domainEstoque.TipoPeca && preco.Valid {
		item.valor = &preco.Float64
	}
	if item.tipo == domainEstoque.TipoInsumo && custo.Valid {
		item.valor = &custo.Float64
	}
	return item, nil
}

func (repository PostgresRepository) ConsultarOrcamentos(ctx context.Context, osID string) ([]domainOrcamento.Resultado, error) {
	var exists bool
	if err := repository.db.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM ordem_servico WHERE id = $1)", osID).Scan(&exists); err != nil {
		return nil, err
	}
	if !exists {
		return nil, application.ErrOSNaoEncontrada
	}
	rows, err := repository.db.Query(ctx, `SELECT o.id, COALESCE(o.orcamento_original_id::text, ''), o.tipo_orcamento, o.status, oi.item_estoque_id, ie.codigo, oi.descricao, oi.tipo_item, oi.quantidade, oi.valor_unitario, oi.valor_total FROM orcamento o LEFT JOIN orcamento_item oi ON oi.orcamento_id = o.id LEFT JOIN item_estoque ie ON ie.id = oi.item_estoque_id WHERE o.ordem_servico_id = $1 ORDER BY o.criado_em, oi.id`, osID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	results := make([]domainOrcamento.Resultado, 0)
	byID := make(map[string]int)
	for rows.Next() {
		var budget domainOrcamento.Resultado
		var itemID, codigo, descricao, tipo sql.NullString
		var quantity, unit, total sql.NullFloat64
		var originalID, budgetType, budgetStatus string
		if err := rows.Scan(&budget.OrcamentoID, &originalID, &budgetType, &budgetStatus, &itemID, &codigo, &descricao, &tipo, &quantity, &unit, &total); err != nil {
			return nil, err
		}
		budget.OrcamentoOriginal = originalID
		budget.TipoOrcamento = budgetType
		budget.StatusOrcamento = budgetStatus
		budget.OrdemServicoID = osID
		index, exists := byID[budget.OrcamentoID]
		if !exists {
			index = len(results)
			byID[budget.OrcamentoID] = index
			results = append(results, budget)
		}
		if itemID.Valid {
			results[index].ItensRegistrados = append(results[index].ItensRegistrados, domainOrcamento.Item{ItemID: itemID.String, Codigo: codigo.String, Descricao: descricao.String, Tipo: tipo.String, Quantidade: quantity.Float64, ValorUnitario: unit.Float64, ValorItem: total.Float64})
			results[index].ValorOrcamento += total.Float64
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var grandTotal float64
	for _, result := range results {
		grandTotal += result.ValorOrcamento
	}
	for index := range results {
		results[index].ValorTotalGeral = grandTotal
	}
	return results, nil
}
