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

func (repository PostgresRepository) BuscarPorIDIncluindoInativo(ctx context.Context, id string) (cliente.Cliente, error) {
	const query = `SELECT id, nome, documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), ativo,
			inativado_em, COALESCE(inativado_por::text, ''), COALESCE(motivo_inativacao, ''), version
		FROM cliente
		WHERE id = $1`
	var cliente cliente.Cliente
	err := repository.db.QueryRow(ctx, query, id).
		Scan(&cliente.ID, &cliente.Nome, &cliente.Documento, &cliente.TipoDocumento, &cliente.Telefone, &cliente.Email, &cliente.Ativo, &cliente.InativadoEm, &cliente.InativadoPor, &cliente.Motivo, &cliente.Version)
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

func (repository PostgresRepository) BuscarOSAbertas(ctx context.Context, clienteID string) ([]application.OrdemServicoAberta, error) {
	const query = `SELECT COALESCE(jsonb_agg(jsonb_build_object('ordemServicoId', id, 'status', status) ORDER BY criada_em), '[]'::jsonb)
		FROM ordem_servico
		WHERE cliente_id = $1 AND status NOT IN ('ENTREGUE', 'CANCELADA')`
	var raw []byte
	if err := repository.db.QueryRow(ctx, query, clienteID).Scan(&raw); err != nil {
		return nil, err
	}
	var ordens []application.OrdemServicoAberta
	return ordens, json.Unmarshal(raw, &ordens)
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

func (repository PostgresRepository) Inativar(ctx context.Context, input application.InativarRepositoryInput) (application.Inativacao, error) {
	const query = `WITH cliente_inativado AS (
			UPDATE cliente
			SET ativo = FALSE, inativado_em = CURRENT_TIMESTAMP, inativado_por = $2, motivo_inativacao = NULLIF($3, ''), version = version + 1
			WHERE id = $1 AND ativo
			RETURNING id, nome, documento, tipo_documento, COALESCE(telefone, '') telefone, COALESCE(email, '') email,
				ativo, inativado_em, COALESCE(inativado_por::text, '') inativado_por, COALESCE(motivo_inativacao, '') motivo, version
		), veiculos_inativados AS (
			UPDATE veiculo v
			SET ativo = FALSE, inativado_em = (SELECT inativado_em FROM cliente_inativado), inativado_por = $2,
				motivo_inativacao = 'Cliente inativado', version = v.version + 1
			FROM cliente_inativado c
			WHERE v.cliente_id = c.id AND v.ativo
			RETURNING v.id, v.placa
		)
		SELECT c.id, c.nome, c.documento, c.tipo_documento, c.telefone, c.email, c.ativo, c.inativado_em,
			c.inativado_por, c.motivo, c.version,
			COALESCE(jsonb_agg(jsonb_build_object('id', v.id, 'placa', v.placa) ORDER BY v.placa)
				FILTER (WHERE v.id IS NOT NULL), '[]'::jsonb)
		FROM cliente_inativado c
		LEFT JOIN veiculos_inativados v ON TRUE
		GROUP BY c.id, c.nome, c.documento, c.tipo_documento, c.telefone, c.email, c.ativo, c.inativado_em,
			c.inativado_por, c.motivo, c.version`
	result := application.Inativacao{DocumentoLiberado: true}
	var veiculos []byte
	err := repository.db.QueryRow(ctx, query, input.ClienteID, input.InativadoPor, input.Motivo).
		Scan(&result.Cliente.ID, &result.Cliente.Nome, &result.Cliente.Documento, &result.Cliente.TipoDocumento, &result.Cliente.Telefone, &result.Cliente.Email, &result.Cliente.Ativo, &result.Cliente.InativadoEm, &result.Cliente.InativadoPor, &result.Cliente.Motivo, &result.Cliente.Version, &veiculos)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, application.ErrClienteJaInativo
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(veiculos, &result.VeiculosInativados)
	return result, err
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

func (repository PostgresRepository) Reativar(ctx context.Context, input application.ReativarRepositoryInput) (application.Reativacao, error) {
	const query = `UPDATE cliente
		SET ativo = TRUE, inativado_em = NULL, inativado_por = NULL, motivo_inativacao = NULL, version = version + 1
		WHERE id = $1 AND NOT ativo
		RETURNING id, nome, documento, tipo_documento, COALESCE(telefone, ''), COALESCE(email, ''), ativo, version, CURRENT_TIMESTAMP`
	var result application.Reativacao
	err := repository.db.QueryRow(ctx, query, input.ClienteID).
		Scan(&result.Cliente.ID, &result.Cliente.Nome, &result.Cliente.Documento, &result.Cliente.TipoDocumento, &result.Cliente.Telefone, &result.Cliente.Email, &result.Cliente.Ativo, &result.Cliente.Version, &result.ReativadoEm)
	if violacaoUnica(err) {
		return result, application.ErrClienteDuplicado
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return result, application.ErrClienteJaAtivo
	}
	return result, err
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
