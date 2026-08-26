package servico

import (
	"context"
	"encoding/json"
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

func (repository PostgresRepository) ExisteAtivoPorNomeNormalizadoExcetoID(ctx context.Context, nome, id string) (bool, error) {
	var existe bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(
		SELECT 1 FROM servico WHERE nome_normalizado = $1 AND id <> $2 AND ativo
	)`, nome, id).Scan(&existe)
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

func (repository PostgresRepository) Listar(ctx context.Context, filtros application.Filtros) ([]domain.Servico, int, error) {
	const query = `WITH filtrados AS (
			SELECT id, codigo, nome, COALESCE(descricao, '') descricao, valor::text valor,
				tempo_estimado_minutos, ativo
			FROM servico
			WHERE ($1 OR ativo) AND ($2 = '' OR nome_normalizado LIKE '%' || $2 || '%')
		), pagina AS (
			SELECT * FROM filtrados ORDER BY codigo LIMIT $3 OFFSET $4
		)
		SELECT COALESCE(jsonb_agg(jsonb_build_object(
			'id', id, 'codigo', codigo, 'nome', nome, 'descricao', descricao, 'valor', valor,
			'tempoEstimadoMinutos', tempo_estimado_minutos, 'ativo', ativo
		) ORDER BY codigo), '[]'::jsonb), (SELECT COUNT(*) FROM filtrados)
		FROM pagina`
	var raw []byte
	var total int
	err := repository.db.QueryRow(ctx, query, filtros.IncluirInativos, domain.NormalizarNome(filtros.Nome),
		filtros.Tamanho, filtros.Pagina*filtros.Tamanho).Scan(&raw, &total)
	if err != nil {
		return nil, 0, err
	}
	var rows []struct {
		ID, Codigo, Nome, Descricao, Valor string
		TempoEstimadoMinutos               int  `json:"tempoEstimadoMinutos"`
		Ativo                              bool `json:"ativo"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, 0, err
	}
	servicos := make([]domain.Servico, 0, len(rows))
	for _, row := range rows {
		servicos = append(servicos, domain.Servico{ID: row.ID, Codigo: row.Codigo, Nome: row.Nome,
			Descricao: row.Descricao, Valor: row.Valor, TempoEstimadoMinutos: row.TempoEstimadoMinutos, Ativo: row.Ativo})
	}
	return servicos, total, nil
}

func (repository PostgresRepository) BuscarPorID(ctx context.Context, id string) (domain.Servico, error) {
	const query = `SELECT id, codigo, nome, nome_normalizado, COALESCE(descricao, ''), valor::text,
		tempo_estimado_minutos, ativo, version, data_criacao, COALESCE(usuario_criacao::text, '')
		FROM servico WHERE id = $1`
	var servico domain.Servico
	err := repository.db.QueryRow(ctx, query, id).Scan(&servico.ID, &servico.Codigo, &servico.Nome,
		&servico.NomeNormalizado, &servico.Descricao, &servico.Valor, &servico.TempoEstimadoMinutos,
		&servico.Ativo, &servico.Version, &servico.DataCriacao, &servico.UsuarioCriacao)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Servico{}, application.ErrServicoNaoEncontrado
	}
	return servico, err
}

func (repository PostgresRepository) Atualizar(ctx context.Context, servico domain.Servico, version int, usuarioID string) (domain.Servico, error) {
	const query = `UPDATE servico SET nome = $2, nome_normalizado = $3, descricao = NULLIF($4, ''),
		valor = $5::numeric, tempo_estimado_minutos = $6, data_atualizacao = CURRENT_TIMESTAMP,
		usuario_atualizacao = $7, version = version + 1
		WHERE id = $1 AND version = $8
		RETURNING id, codigo, nome, nome_normalizado, COALESCE(descricao, ''), valor::text,
			tempo_estimado_minutos, ativo, version, data_criacao, data_atualizacao, usuario_atualizacao::text`
	err := repository.db.QueryRow(ctx, query, servico.ID, servico.Nome, servico.NomeNormalizado,
		servico.Descricao, servico.Valor, servico.TempoEstimadoMinutos, usuarioID, version).Scan(
		&servico.ID, &servico.Codigo, &servico.Nome, &servico.NomeNormalizado, &servico.Descricao,
		&servico.Valor, &servico.TempoEstimadoMinutos, &servico.Ativo, &servico.Version,
		&servico.DataCriacao, &servico.DataAtualizacao, &servico.UsuarioAtualizacao)
	if violacaoUnica(err) {
		return domain.Servico{}, application.ErrServicoDuplicado
	}
	if errors.Is(err, pgx.ErrNoRows) {
		var versionAtual int
		err = repository.db.QueryRow(ctx, `SELECT version FROM servico WHERE id = $1`, servico.ID).Scan(&versionAtual)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Servico{}, application.ErrServicoNaoEncontrado
		}
		if err != nil {
			return domain.Servico{}, err
		}
		return domain.Servico{}, application.ErrVersaoDivergente
	}
	return servico, err
}

func (repository PostgresRepository) Desativar(ctx context.Context, id, usuarioID string) (domain.Servico, error) {
	const query = `UPDATE servico SET ativo = FALSE, data_desativacao = CURRENT_TIMESTAMP,
		usuario_desativacao = $2, version = version + 1 WHERE id = $1 AND ativo
		RETURNING id, codigo, nome, nome_normalizado, COALESCE(descricao, ''), valor::text,
			tempo_estimado_minutos, ativo, version, data_criacao, data_desativacao, usuario_desativacao::text`
	return repository.alterarSituacao(ctx, query, domain.ErrServicoJaInativo, id, usuarioID)
}

func (repository PostgresRepository) Reativar(ctx context.Context, id string) (domain.Servico, error) {
	const query = `UPDATE servico SET ativo = TRUE, data_desativacao = NULL,
		usuario_desativacao = NULL, version = version + 1 WHERE id = $1 AND NOT ativo
		RETURNING id, codigo, nome, nome_normalizado, COALESCE(descricao, ''), valor::text,
			tempo_estimado_minutos, ativo, version, data_criacao, data_desativacao, COALESCE(usuario_desativacao::text, '')`
	return repository.alterarSituacao(ctx, query, domain.ErrServicoJaAtivo, id)
}

func (repository PostgresRepository) alterarSituacao(ctx context.Context, query string, conflito error, argumentos ...any) (domain.Servico, error) {
	var servico domain.Servico
	err := repository.db.QueryRow(ctx, query, argumentos...).Scan(&servico.ID, &servico.Codigo, &servico.Nome,
		&servico.NomeNormalizado, &servico.Descricao, &servico.Valor, &servico.TempoEstimadoMinutos,
		&servico.Ativo, &servico.Version, &servico.DataCriacao, &servico.DataDesativacao, &servico.UsuarioDesativacao)
	if violacaoUnica(err) {
		return domain.Servico{}, application.ErrServicoDuplicado
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Servico{}, conflito
	}
	return servico, err
}

func violacaoUnica(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
