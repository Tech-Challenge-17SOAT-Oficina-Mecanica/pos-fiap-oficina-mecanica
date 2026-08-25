package servico

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/servico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/servico"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository {
	return PostgresRepository{db: db}
}

func (r PostgresRepository) Cadastrar(ctx context.Context, s domain.Servico) (domain.Servico, error) {
	var seq int
	if err := r.db.QueryRow(ctx, `SELECT nextval('seq_servico_codigo')`).Scan(&seq); err != nil {
		return domain.Servico{}, err
	}
	s.Codigo = fmt.Sprintf("SER-%06d", seq)

	err := r.db.QueryRow(ctx, `
		INSERT INTO servico (codigo, nome, nome_normalizado, descricao, valor, tempo_estimado_minutos)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, ativo, version, data_criacao`,
		s.Codigo,
		s.Nome,
		s.NomeNormalizado,
		nullIfEmpty(s.Descricao),
		s.Valor,
		s.TempoEstimadoMinutos,
	).Scan(&s.ID, &s.Ativo, &s.Version, &s.DataCriacao)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Servico{}, application.ErrNomeDuplicado
		}
		return domain.Servico{}, err
	}
	return s, nil
}

func (r PostgresRepository) Listar(ctx context.Context, filtros application.FiltrosConsulta) ([]domain.Servico, int, error) {
	const countQuery = `
		SELECT COUNT(*)
		FROM servico
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%')
		AND ($2 OR ativo)`
	const listQuery = `
		SELECT id, codigo, nome, COALESCE(descricao,''), valor, tempo_estimado_minutos, ativo, version, data_criacao
		FROM servico
		WHERE ($1 = '' OR nome ILIKE '%' || $1 || '%')
		AND ($2 OR ativo)
		ORDER BY nome ASC, id ASC
		LIMIT $3 OFFSET $4`

	var total int
	if err := r.db.QueryRow(ctx, countQuery, filtros.Nome, filtros.IncluirInativos).Scan(&total); err != nil {
		return nil, 0, err
	}
	rows, err := r.db.Query(ctx, listQuery, filtros.Nome, filtros.IncluirInativos, filtros.Tamanho, filtros.Pagina*filtros.Tamanho)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	servicos := make([]domain.Servico, 0)
	for rows.Next() {
		s, err := scanServico(rows.Scan)
		if err != nil {
			return nil, 0, err
		}
		servicos = append(servicos, s)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return servicos, total, nil
}

func (r PostgresRepository) BuscarPorID(ctx context.Context, servicoID string) (domain.Servico, error) {
	const query = `
		SELECT id, codigo, nome, COALESCE(descricao,''), valor, tempo_estimado_minutos, ativo, version, data_criacao
		FROM servico WHERE id = $1`
	s, err := scanServico(r.db.QueryRow(ctx, query, servicoID).Scan)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Servico{}, application.ErrServicoNaoEncontrado
		}
		return domain.Servico{}, err
	}
	return s, nil
}

func (r PostgresRepository) Atualizar(ctx context.Context, servicoID string, input domain.AtualizacaoInput, version int, usuarioID string) (domain.Servico, error) {
	// Build dynamic SET clause for PATCH semantics
	const query = `
		UPDATE servico
		SET nome                   = COALESCE($2, nome),
		    nome_normalizado        = COALESCE($3, nome_normalizado),
		    descricao               = CASE WHEN $4::boolean THEN $5 ELSE descricao END,
		    valor                   = COALESCE($6, valor),
		    tempo_estimado_minutos  = COALESCE($7, tempo_estimado_minutos),
		    data_atualizacao        = CURRENT_TIMESTAMP,
		    usuario_atualizacao     = NULLIF($8,'')::uuid,
		    version                 = version + 1
		WHERE id = $1 AND ativo AND version = $9
		RETURNING id, codigo, nome, COALESCE(descricao,''), valor, tempo_estimado_minutos, ativo, version, data_criacao, data_atualizacao`

	var nomeNorm *string
	if input.Nome != nil {
		n := domain.NormalizarNome(*input.Nome)
		nomeNorm = &n
	}

	descricaoSet := input.Descricao != nil
	var descricaoVal *string
	if descricaoSet {
		descricaoVal = input.Descricao
	}

	s, err := scanServicoAtualizado(r.db.QueryRow(ctx, query,
		servicoID,
		input.Nome,
		nomeNorm,
		descricaoSet,
		descricaoVal,
		input.Valor,
		input.TempoEstimadoMinutos,
		usuarioID,
		version,
	).Scan)
	if err == nil {
		return s, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Servico{}, application.ErrNomeDuplicado
		}
		return domain.Servico{}, err
	}

	// Distinguish 404 / 409-inativo / 412-versão
	var ativo bool
	var versionAtual int
	diagErr := r.db.QueryRow(ctx, `SELECT ativo, version FROM servico WHERE id = $1`, servicoID).Scan(&ativo, &versionAtual)
	if errors.Is(diagErr, pgx.ErrNoRows) {
		return domain.Servico{}, application.ErrServicoNaoEncontrado
	}
	if diagErr != nil {
		return domain.Servico{}, diagErr
	}
	if !ativo {
		return domain.Servico{}, application.ErrServicoInativo
	}
	if versionAtual != version {
		return domain.Servico{}, application.ErrVersaoDivergente
	}
	return domain.Servico{}, application.ErrAtualizacaoInvalida
}

func (r PostgresRepository) Desativar(ctx context.Context, servicoID, usuarioID string) (domain.Servico, error) {
	var ativo bool
	err := r.db.QueryRow(ctx, `SELECT ativo FROM servico WHERE id = $1`, servicoID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Servico{}, application.ErrServicoNaoEncontrado
	}
	if err != nil {
		return domain.Servico{}, err
	}
	if !ativo {
		return domain.Servico{}, application.ErrServicoJaInativo
	}

	const query = `
		UPDATE servico
		SET ativo = false,
		    data_desativacao    = CURRENT_TIMESTAMP,
		    usuario_desativacao = NULLIF($2,'')::uuid
		WHERE id = $1 AND ativo
		RETURNING id, codigo, nome, COALESCE(descricao,''), valor, tempo_estimado_minutos, ativo, version, data_criacao, data_desativacao, COALESCE(usuario_desativacao::text,'')`
	s, err := scanServicoSituacao(r.db.QueryRow(ctx, query, servicoID, usuarioID).Scan)
	if err != nil {
		return domain.Servico{}, err
	}
	return s, nil
}

func (r PostgresRepository) Reativar(ctx context.Context, servicoID, _ string) (domain.Servico, error) {
	var ativo bool
	err := r.db.QueryRow(ctx, `SELECT ativo FROM servico WHERE id = $1`, servicoID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Servico{}, application.ErrServicoNaoEncontrado
	}
	if err != nil {
		return domain.Servico{}, err
	}
	if ativo {
		return domain.Servico{}, application.ErrServicoJaAtivo
	}

	const query = `
		UPDATE servico
		SET ativo = true,
		    data_desativacao    = NULL,
		    usuario_desativacao = NULL
		WHERE id = $1 AND NOT ativo
		RETURNING id, codigo, nome, COALESCE(descricao,''), valor, tempo_estimado_minutos, ativo, version, data_criacao, data_desativacao, COALESCE(usuario_desativacao::text,'')`
	s, err := scanServicoSituacao(r.db.QueryRow(ctx, query, servicoID).Scan)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Servico{}, application.ErrNomeAtivoDuplicado
		}
		return domain.Servico{}, err
	}
	return s, nil
}

// ── scan helpers ──────────────────────────────────────────────────────────────

func scanServico(scan func(dest ...any) error) (domain.Servico, error) {
	var s domain.Servico
	err := scan(&s.ID, &s.Codigo, &s.Nome, &s.Descricao, &s.Valor, &s.TempoEstimadoMinutos, &s.Ativo, &s.Version, &s.DataCriacao)
	return s, err
}

func scanServicoAtualizado(scan func(dest ...any) error) (domain.Servico, error) {
	var s domain.Servico
	var dataAtualizacao pgtype.Timestamptz
	err := scan(&s.ID, &s.Codigo, &s.Nome, &s.Descricao, &s.Valor, &s.TempoEstimadoMinutos, &s.Ativo, &s.Version, &s.DataCriacao, &dataAtualizacao)
	if err != nil {
		return domain.Servico{}, err
	}
	if dataAtualizacao.Valid {
		t := dataAtualizacao.Time
		s.DataAtualizacao = &t
	}
	return s, nil
}

func scanServicoSituacao(scan func(dest ...any) error) (domain.Servico, error) {
	var s domain.Servico
	var dataDesativacao pgtype.Timestamptz
	err := scan(&s.ID, &s.Codigo, &s.Nome, &s.Descricao, &s.Valor, &s.TempoEstimadoMinutos, &s.Ativo, &s.Version, &s.DataCriacao, &dataDesativacao, &s.UsuarioDesativacao)
	if err != nil {
		return domain.Servico{}, err
	}
	if dataDesativacao.Valid {
		t := dataDesativacao.Time
		s.DataDesativacao = &t
	}
	return s, nil
}

func nullIfEmpty(v string) any {
	if v == "" {
		return nil
	}
	return v
}
