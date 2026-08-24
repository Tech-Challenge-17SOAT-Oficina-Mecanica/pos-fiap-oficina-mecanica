package cliente

import (
	"context"
	"encoding/json"
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

func (repository PostgresRepository) ExisteAtivoPorDocumentoExcetoID(ctx context.Context, documento, id string) (bool, error) {
	var exists bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cliente WHERE documento = $1 AND id <> $2 AND ativo)`, documento, id).Scan(&exists)
	return exists, err
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (cliente.Cliente, error) {
	const query = `SELECT id, nome, documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), ativo, version
		FROM cliente
		WHERE id = $1 AND ativo`
	var cliente cliente.Cliente
	err := repository.db.QueryRow(ctx, query, id).
		Scan(&cliente.ID, &cliente.Nome, &cliente.Documento, &cliente.TipoDocumento, &cliente.Telefone, &cliente.Email, &cliente.Ativo, &cliente.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return cliente, application.ErrClienteNaoEncontrado
	}
	return cliente, err
}

func (repository PostgresRepository) BuscarPorDocumento(ctx context.Context, documento string) (cliente.Cliente, error) {
	const query = `SELECT c.id, c.nome, c.documento, c.tipo_documento, COALESCE(c.telefone, ''), COALESCE(c.email, ''), c.ativo, c.version,
		COALESCE(jsonb_agg(jsonb_build_object('id', v.id, 'placa', v.placa, 'marca', v.marca, 'modelo', v.modelo, 'ano', v.ano) ORDER BY v.placa)
			FILTER (WHERE v.id IS NOT NULL), '[]'::jsonb)
		FROM cliente c
		LEFT JOIN veiculo v ON v.cliente_id = c.id AND v.ativo
		WHERE c.documento = $1 AND c.ativo
		GROUP BY c.id`
	var cliente cliente.Cliente
	var veiculos []byte
	err := repository.db.QueryRow(ctx, query, documento).
		Scan(&cliente.ID, &cliente.Nome, &cliente.Documento, &cliente.TipoDocumento, &cliente.Telefone, &cliente.Email, &cliente.Ativo, &cliente.Version, &veiculos)
	if errors.Is(err, pgx.ErrNoRows) {
		return cliente, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return cliente, err
	}
	err = json.Unmarshal(veiculos, &cliente.Veiculos)
	return cliente, err
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

func (repository PostgresRepository) Atualizar(ctx context.Context, cliente cliente.Cliente, version int) (cliente.Cliente, error) {
	const query = `UPDATE cliente
		SET nome = $2, documento = $3, tipo_documento = $4, telefone = NULLIF($5, ''), email = NULLIF($6, ''), version = version + 1
		WHERE id = $1 AND ativo AND version = $7
		RETURNING id, nome, documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), ativo, version`
	err := repository.db.QueryRow(ctx, query, cliente.ID, cliente.Nome, cliente.Documento, cliente.TipoDocumento, cliente.Telefone, cliente.Email, version).
		Scan(&cliente.ID, &cliente.Nome, &cliente.Documento, &cliente.TipoDocumento, &cliente.Telefone, &cliente.Email, &cliente.Ativo, &cliente.Version)
	if violacaoUnica(err) {
		return cliente, application.ErrClienteDuplicado
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return cliente, application.ErrVersaoDivergente
	}
	return cliente, err
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
