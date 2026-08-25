package ordemservico

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/ordemservico"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/ordemservico"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db} }

func (repository PostgresRepository) RegistrarProblemaRelatado(ctx context.Context, osID string, problema domain.ProblemaRelatado) (result domain.OrdemDeServico, err error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return result, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx, `SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE`, osID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return result, err
	}
	var jaRegistrado bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM problema_ordem_servico WHERE ordem_servico_id = $1 AND tipo = $2)`, osID, domain.TipoProblemaRelatado).Scan(&jaRegistrado); err != nil {
		return result, err
	}
	if jaRegistrado {
		return result, application.ErrProblemaRelatadoJaRegistrado
	}
	if status != domain.StatusRecebida {
		return result, application.ErrOrdemServicoForaDeRecebida
	}

	const insertProblem = `INSERT INTO problema_ordem_servico (ordem_servico_id, tipo, descricao, observacoes)
		VALUES ($1, $2, $3, NULLIF($4, '')) RETURNING descricao, COALESCE(observacoes, ''), registrado_em`
	if err = tx.QueryRow(ctx, insertProblem, osID, domain.TipoProblemaRelatado, problema.Descricao, problema.Observacoes).
		Scan(&result.ProblemaRelatado.Descricao, &result.ProblemaRelatado.Observacoes, &result.ProblemaRelatado.RegistradoEm); err != nil {
		return result, err
	}
	const updateOS = `UPDATE ordem_servico SET status = $2, iniciada_em = CURRENT_TIMESTAMP WHERE id = $1 RETURNING id, status, iniciada_em`
	if err = tx.QueryRow(ctx, updateOS, osID, domain.StatusEmDiagnostico).Scan(&result.ID, &result.Status, &result.DataInicioDiagnostico); err != nil {
		return result, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return result, nil
}
