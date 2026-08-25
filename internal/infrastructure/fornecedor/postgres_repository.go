package fornecedor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type queryer interface {
	Begin(context.Context) (pgx.Tx, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type PostgresRepository struct{ db queryer }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (repository PostgresRepository) Cadastrar(ctx context.Context, cadastro domain.Cadastro) (domain.Fornecedor, error) {
	fornecedor := domain.Fornecedor{Cadastro: cadastro}
	err := repository.db.QueryRow(ctx, `
		INSERT INTO fornecedor (
			razao_social, nome_fantasia, documento, tipo_documento, telefone, email, prazo_entrega_dias
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, ativo, version, criado_em`,
		cadastro.RazaoSocial,
		nullIfEmpty(cadastro.NomeFantasia),
		cadastro.Documento,
		cadastro.TipoDocumento,
		nullIfEmpty(cadastro.Telefone),
		nullIfEmpty(cadastro.Email),
		cadastro.PrazoEntregaDias,
	).Scan(&fornecedor.ID, &fornecedor.Ativo, &fornecedor.Version, &fornecedor.CriadoEm)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.Fornecedor{}, application.ErrDocumentoDuplicado
		}
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func (repository PostgresRepository) Listar(ctx context.Context, filtros application.FiltrosConsulta) ([]domain.Fornecedor, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM fornecedor
		WHERE ($1 = '' OR razao_social ILIKE '%' || $1 || '%' OR COALESCE(nome_fantasia, '') ILIKE '%' || $1 || '%')
		AND ($2 = '' OR documento = $2)
		AND ($3 OR ativo)`
	const listQuery = `
		SELECT id, razao_social, COALESCE(nome_fantasia, ''), documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), prazo_entrega_dias, ativo, version, criado_em
		FROM fornecedor
		WHERE ($1 = '' OR razao_social ILIKE '%' || $1 || '%' OR COALESCE(nome_fantasia, '') ILIKE '%' || $1 || '%')
		AND ($2 = '' OR documento = $2)
		AND ($3 OR ativo)
		ORDER BY razao_social ASC, id ASC
		LIMIT $4 OFFSET $5`

	var total int
	if err := repository.db.QueryRow(ctx, countQuery, filtros.Nome, filtros.Documento, filtros.IncluirInativos).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := repository.db.Query(ctx, listQuery, filtros.Nome, filtros.Documento, filtros.IncluirInativos, filtros.Tamanho, filtros.Pagina*filtros.Tamanho)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	fornecedores := make([]domain.Fornecedor, 0)
	for rows.Next() {
		fornecedor, err := scanFornecedor(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		fornecedores = append(fornecedores, fornecedor)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return fornecedores, total, nil
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, fornecedorID string) (domain.Fornecedor, error) {
	const query = `
		SELECT id, razao_social, COALESCE(nome_fantasia, ''), documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), prazo_entrega_dias, ativo, version, criado_em
		FROM fornecedor
		WHERE id = $1`
	fornecedor, err := scanFornecedor(repository.db.QueryRow(ctx, query, fornecedorID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Fornecedor{}, application.ErrFornecedorNaoEncontrado
		}
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func (repository PostgresRepository) Atualizar(ctx context.Context, fornecedorID string, atualizacao domain.Atualizacao, version int, usuarioID string) (domain.Fornecedor, error) {
	const updateQuery = `
		UPDATE fornecedor
		SET razao_social = $2,
			nome_fantasia = $3,
			telefone = $4,
			email = $5,
			prazo_entrega_dias = COALESCE($6, prazo_entrega_dias),
			data_atualizacao = CURRENT_TIMESTAMP,
			usuario_atualizacao = NULLIF($7, '')::uuid,
			version = version + 1
		WHERE id = $1 AND ativo AND version = $8
		RETURNING id, razao_social, COALESCE(nome_fantasia, ''), documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), prazo_entrega_dias, ativo, version, criado_em, data_atualizacao`

	fornecedor, err := scanFornecedorAtualizado(repository.db.QueryRow(ctx, updateQuery,
		fornecedorID,
		atualizacao.RazaoSocial,
		nullIfEmpty(atualizacao.NomeFantasia),
		nullIfEmpty(atualizacao.Telefone),
		nullIfEmpty(atualizacao.Email),
		atualizacao.PrazoEntregaDias,
		usuarioID,
		version,
	).Scan)
	if err == nil {
		return fornecedor, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Fornecedor{}, err
	}

	var ativo bool
	var versionAtual int
	err = repository.db.QueryRow(ctx, `SELECT ativo, version FROM fornecedor WHERE id = $1`, fornecedorID).Scan(&ativo, &versionAtual)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Fornecedor{}, application.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if !ativo {
		return domain.Fornecedor{}, application.ErrFornecedorInativo
	}
	if versionAtual != version {
		return domain.Fornecedor{}, application.ErrVersaoDivergente
	}
	return domain.Fornecedor{}, application.ErrAtualizacaoInvalida
}

func (repository PostgresRepository) Desativar(ctx context.Context, fornecedorID, usuarioID string) (domain.Fornecedor, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	defer tx.Rollback(ctx)

	var ativo bool
	err = tx.QueryRow(ctx, `SELECT ativo FROM fornecedor WHERE id = $1`, fornecedorID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Fornecedor{}, application.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if !ativo {
		return domain.Fornecedor{}, application.ErrFornecedorJaInativo
	}

	var pedidosAbertos int
	err = tx.QueryRow(ctx, `SELECT COUNT(*) FROM pedido_compra WHERE fornecedor_id = $1 AND status IN ('ABERTO', 'PARCIAL')`, fornecedorID).Scan(&pedidosAbertos)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if pedidosAbertos > 0 {
		return domain.Fornecedor{}, application.ErrFornecedorComPedidoAberto
	}

	const query = `
		UPDATE fornecedor
		SET ativo = false,
			inativado_em = CURRENT_TIMESTAMP,
			inativado_por = NULLIF($2, '')::uuid
		WHERE id = $1 AND ativo
		RETURNING id, razao_social, COALESCE(nome_fantasia, ''), documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), prazo_entrega_dias, ativo, version, criado_em, inativado_em, COALESCE(inativado_por::text, '')`
	fornecedor, err := scanFornecedorSituacao(tx.QueryRow(ctx, query, fornecedorID, usuarioID).Scan)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func (repository PostgresRepository) Reativar(ctx context.Context, fornecedorID, _ string) (domain.Fornecedor, error) {
	var ativo bool
	err := repository.db.QueryRow(ctx, `SELECT ativo FROM fornecedor WHERE id = $1`, fornecedorID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Fornecedor{}, application.ErrFornecedorNaoEncontrado
	}
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if ativo {
		return domain.Fornecedor{}, application.ErrFornecedorJaAtivo
	}

	const query = `
		UPDATE fornecedor
		SET ativo = true,
			inativado_em = NULL,
			inativado_por = NULL
		WHERE id = $1 AND NOT ativo
		RETURNING id, razao_social, COALESCE(nome_fantasia, ''), documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), prazo_entrega_dias, ativo, version, criado_em, inativado_em, COALESCE(inativado_por::text, '')`
	fornecedor, err := scanFornecedorSituacao(repository.db.QueryRow(ctx, query, fornecedorID).Scan)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return domain.Fornecedor{}, application.ErrDocumentoAtivoDuplicado
		}
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func scanFornecedor(scan func(dest ...any) error) (domain.Fornecedor, error) {
	var fornecedor domain.Fornecedor
	err := scan(
		&fornecedor.ID,
		&fornecedor.RazaoSocial,
		&fornecedor.NomeFantasia,
		&fornecedor.Documento,
		&fornecedor.TipoDocumento,
		&fornecedor.Telefone,
		&fornecedor.Email,
		&fornecedor.PrazoEntregaDias,
		&fornecedor.Ativo,
		&fornecedor.Version,
		&fornecedor.CriadoEm,
	)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func scanFornecedorAtualizado(scan func(dest ...any) error) (domain.Fornecedor, error) {
	var fornecedor domain.Fornecedor
	err := scan(
		&fornecedor.ID,
		&fornecedor.RazaoSocial,
		&fornecedor.NomeFantasia,
		&fornecedor.Documento,
		&fornecedor.TipoDocumento,
		&fornecedor.Telefone,
		&fornecedor.Email,
		&fornecedor.PrazoEntregaDias,
		&fornecedor.Ativo,
		&fornecedor.Version,
		&fornecedor.CriadoEm,
		&fornecedor.AtualizadoEm,
	)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	return fornecedor, nil
}

func scanFornecedorSituacao(scan func(dest ...any) error) (domain.Fornecedor, error) {
	var fornecedor domain.Fornecedor
	var inativadoEm pgtype.Timestamptz
	err := scan(
		&fornecedor.ID,
		&fornecedor.RazaoSocial,
		&fornecedor.NomeFantasia,
		&fornecedor.Documento,
		&fornecedor.TipoDocumento,
		&fornecedor.Telefone,
		&fornecedor.Email,
		&fornecedor.PrazoEntregaDias,
		&fornecedor.Ativo,
		&fornecedor.Version,
		&fornecedor.CriadoEm,
		&inativadoEm,
		&fornecedor.InativadoPor,
	)
	if err != nil {
		return domain.Fornecedor{}, err
	}
	if inativadoEm.Valid {
		fornecedor.InativadoEm = &inativadoEm.Time
	}
	return fornecedor, nil
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
