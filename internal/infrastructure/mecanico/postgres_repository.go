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

func (repository PostgresRepository) EmailExisteExcetoMecanico(ctx context.Context, email, mecanicoID string) (bool, error) {
	var exists bool
	const query = `SELECT EXISTS(
		SELECT 1
		FROM usuario u
		LEFT JOIN mecanico m ON m.usuario_id = u.id
		WHERE lower(u.email) = lower($1) AND COALESCE(m.id::text, '') <> $2
	)`
	err := repository.db.QueryRow(ctx, query, email, mecanicoID).Scan(&exists)
	return exists, err
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (domain.Mecanico, error) {
	const query = `SELECT m.id, m.usuario_id, m.nome, u.email, u.ativo,
			COALESCE(array_agg(ue.escopo ORDER BY ue.escopo) FILTER (WHERE ue.escopo IS NOT NULL), '{}'), m.version
		FROM mecanico m
		JOIN usuario u ON u.id = m.usuario_id
		LEFT JOIN usuario_escopo ue ON ue.usuario_id = u.id
		WHERE m.id = $1
		GROUP BY m.id, u.id`
	var mecanico domain.Mecanico
	err := repository.db.QueryRow(ctx, query, id).
		Scan(&mecanico.ID, &mecanico.UsuarioID, &mecanico.Nome, &mecanico.Email, &mecanico.Ativo, &mecanico.Escopos, &mecanico.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mecanico{}, application.ErrMecanicoNaoEncontrado
	}
	return mecanico, err
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

func (repository PostgresRepository) AtualizarMecanico(ctx context.Context, mecanico domain.Mecanico, version int) (domain.Mecanico, error) {
	tx, err := repository.begin(ctx)
	if err != nil {
		return domain.Mecanico{}, err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `UPDATE usuario SET email = lower($2) WHERE id = $1`, mecanico.UsuarioID, mecanico.Email); violacaoUnica(err) {
		return domain.Mecanico{}, application.ErrEmailDuplicado
	} else if err != nil {
		return domain.Mecanico{}, err
	}
	const updateMecanico = `UPDATE mecanico
		SET nome = $2, version = version + 1
		WHERE id = $1 AND version = $3
		RETURNING id, nome, version`
	err = tx.QueryRow(ctx, updateMecanico, mecanico.ID, mecanico.Nome, version).Scan(&mecanico.ID, &mecanico.Nome, &mecanico.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Mecanico{}, application.ErrVersaoDivergente
	}
	if err != nil {
		return domain.Mecanico{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM usuario_escopo WHERE usuario_id = $1`, mecanico.UsuarioID); err != nil {
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
