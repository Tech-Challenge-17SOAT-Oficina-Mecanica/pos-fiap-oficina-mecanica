package mecanico

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/mecanico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/mecanico"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type transaction interface {
	QueryRow(context.Context, string, ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Commit(context.Context) error
	Rollback(context.Context) error
}

type PostgresRepository struct {
	db    queryer
	begin func(context.Context) (transaction, error)
}

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db, begin: func(ctx context.Context) (transaction, error) { return db.Begin(ctx) }}
}

func (repository PostgresRepository) EmailExiste(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM usuario WHERE lower(email) = lower($1))`, email).Scan(&exists)
	return exists, err
}

func (repository PostgresRepository) SalvarMecanico(ctx context.Context, mecanico domain.Mecanico, senhaHash string) (domain.Mecanico, error) {
	tx, err := repository.begin(ctx)
	if err != nil {
		return domain.Mecanico{}, err
	}
	defer tx.Rollback(ctx)

	const insertUsuario = `INSERT INTO usuario (email, senha_hash, ativo)
		VALUES (lower($1), $2, TRUE)
		RETURNING id, email, ativo`
	err = tx.QueryRow(ctx, insertUsuario, mecanico.Email, senhaHash).Scan(&mecanico.UsuarioID, &mecanico.Email, &mecanico.Ativo)
	if violacaoUnica(err) {
		return domain.Mecanico{}, application.ErrEmailDuplicado
	}
	if err != nil {
		return domain.Mecanico{}, err
	}
	const insertMecanico = `INSERT INTO mecanico (usuario_id, nome, version)
		VALUES ($1, $2, 1)
		RETURNING id, nome, version`
	err = tx.QueryRow(ctx, insertMecanico, mecanico.UsuarioID, mecanico.Nome).Scan(&mecanico.ID, &mecanico.Nome, &mecanico.Version)
	if err != nil {
		return domain.Mecanico{}, err
	}
	for _, escopo := range mecanico.Escopos {
		if _, err := tx.Exec(ctx, `INSERT INTO usuario_escopo (usuario_id, escopo) VALUES ($1, $2)`, mecanico.UsuarioID, escopo); err != nil {
			return domain.Mecanico{}, err
		}
	}
	return mecanico, tx.Commit(ctx)
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
