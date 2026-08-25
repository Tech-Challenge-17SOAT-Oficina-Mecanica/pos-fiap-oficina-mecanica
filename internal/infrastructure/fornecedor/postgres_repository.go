package fornecedor

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/fornecedor"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/fornecedor"
)

type queryer interface {
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

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
