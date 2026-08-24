package cliente

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/cliente"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/cliente"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type PostgresRepository struct{ db queryer }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) ExisteAtivoPorDocumento(ctx context.Context, documento string) (bool, error) {
	var exists bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cliente WHERE documento = $1 AND ativo)`, documento).Scan(&exists)
	return exists, err
}

func (repository PostgresRepository) Salvar(ctx context.Context, cliente cliente.Cliente) (cliente.Cliente, error) {
	const query = `INSERT INTO cliente (nome, documento, tipo_documento, telefone, email, ativo, version)
		VALUES ($1, $2, $3, NULLIF($4, ''), NULLIF($5, ''), TRUE, 1)
		RETURNING id, nome, documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), ativo, version`
	err := repository.db.QueryRow(ctx, query, cliente.Nome, cliente.Documento, cliente.TipoDocumento, cliente.Telefone, cliente.Email).
		Scan(&cliente.ID, &cliente.Nome, &cliente.Documento, &cliente.TipoDocumento, &cliente.Telefone, &cliente.Email, &cliente.Ativo, &cliente.Version)
	if violacaoUnica(err) {
		return cliente, application.ErrClienteDuplicado
	}
	return cliente, err
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
