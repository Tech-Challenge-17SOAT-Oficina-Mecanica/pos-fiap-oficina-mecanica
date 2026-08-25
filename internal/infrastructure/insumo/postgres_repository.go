package insumo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	i.unidade_medida, i.custo_unitario::text,
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

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (insumo.Insumo, error) {
	consulta := fmt.Sprintf(`SELECT %s FROM item_estoque i
		JOIN categoria c ON c.id = i.categoria_id
		WHERE i.id = $1 AND i.tipo = 'INSUMO'`, colunas)

	item, err := ler(repository.db.QueryRow(ctx, consulta, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return insumo.Insumo{}, pgx.ErrNoRows
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

	var id string
	var criadoEm time.Time
	err = repository.db.QueryRow(ctx, `
		INSERT INTO item_estoque (
			categoria_id, tipo, codigo, nome, descricao, descricao_normalizada,
			unidade_medida, custo_unitario, estoque_minimo
		) VALUES (
			$1, 'INSUMO', 'INS-' || LPAD(nextval('seq_insumo_codigo')::TEXT, 6, '0'),
			$2, $3, $4, $5, $6::NUMERIC, $7::NUMERIC
		)
		RETURNING id, criado_em`,
		cadastro.CategoriaID, cadastro.Nome, cadastro.Descricao, cadastro.DescricaoNormalizada,
		cadastro.UnidadeMedida, cadastro.CustoUnitario, cadastro.EstoqueMinimo,
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

type scanner interface {
	Scan(destinos ...any) error
}

func ler(linha scanner) (insumo.Insumo, error) {
	var item insumo.Insumo
	err := linha.Scan(
		&item.ID, &item.Codigo, &item.Nome, &item.Descricao, &item.CategoriaID, &item.Categoria,
		&item.UnidadeMedida, &item.CustoUnitario,
		&item.SaldoFisico, &item.SaldoReservado, &item.EstoqueMinimo,
		&item.Ativo, &item.Version, &item.PossuiPedidoEmAberto,
	)
	return item, err
}
