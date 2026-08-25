package servico

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

type PostgresRepository struct{ db queryer }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) ExisteAtivoPorNomeNormalizado(ctx context.Context, nome string) (bool, error) {
	var existe bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM servico WHERE nome_normalizado = $1 AND ativo)`, nome).Scan(&existe)
	return existe, err
}

func (repository PostgresRepository) Salvar(ctx context.Context, servico domain.Servico) (domain.Servico, error) {
	const query = `INSERT INTO servico
		(codigo, nome, nome_normalizado, descricao, valor, tempo_estimado_minutos, ativo, version, usuario_criacao)
		VALUES ('SER-' || LPAD(nextval('servico_codigo_seq')::text, 6, '0'), $1, $2, NULLIF($3, ''), $4::numeric, $5, TRUE, 1, $6)
		RETURNING id, codigo, nome, nome_normalizado, COALESCE(descricao, ''), valor::text, tempo_estimado_minutos,
			ativo, version, data_criacao, usuario_criacao::text`
	err := repository.db.QueryRow(ctx, query, servico.Nome, servico.NomeNormalizado, servico.Descricao,
		servico.Valor, servico.TempoEstimadoMinutos, servico.UsuarioCriacao).Scan(
		&servico.ID, &servico.Codigo, &servico.Nome, &servico.NomeNormalizado, &servico.Descricao,
		&servico.Valor, &servico.TempoEstimadoMinutos, &servico.Ativo, &servico.Version,
		&servico.DataCriacao, &servico.UsuarioCriacao)
	if violacaoUnica(err) {
		return domain.Servico{}, application.ErrServicoDuplicado
	}
	return servico, err
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
