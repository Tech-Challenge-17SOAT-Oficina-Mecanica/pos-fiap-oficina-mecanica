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

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) RegistrarProblema(ctx context.Context, ordemServicoID string, cadastro domain.ProblemaCadastro) (application.ResultadoRegistroProblema, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", ordemServicoID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroProblema{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	tipo, err := domain.TipoOrcamentoParaStatus(status)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}

	orcamento, err := obterOuCriarOrcamento(ctx, tx, ordemServicoID, tipo)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	problema := domain.Problema{Descricao: cadastro.Descricao, Observacoes: cadastro.Observacoes}
	err = tx.QueryRow(ctx, `
		INSERT INTO problema_ordem_servico (ordem_servico_id, orcamento_id, descricao, observacoes)
		VALUES ($1, $2, $3, NULLIF($4, ''))
		RETURNING id, registrado_em`, ordemServicoID, orcamento.ID, cadastro.Descricao, cadastro.Observacoes,
	).Scan(&problema.ID, &problema.RegistradoEm)
	if err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return application.ResultadoRegistroProblema{}, err
	}
	return application.ResultadoRegistroProblema{Problema: problema, Orcamento: orcamento}, nil
}

func obterOuCriarOrcamento(ctx context.Context, tx pgx.Tx, ordemServicoID, tipo string) (domain.Orcamento, error) {
	query := "SELECT id, tipo_orcamento, status FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = $2"
	args := []any{ordemServicoID, tipo}
	if tipo == domain.OrcamentoComplementar {
		query += " AND status = 'CRIADO' ORDER BY criado_em DESC LIMIT 1"
	} else {
		query += " LIMIT 1"
	}
	var orcamento domain.Orcamento
	err := tx.QueryRow(ctx, query, args...).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	if err == nil {
		if orcamento.Status != domain.OrcamentoCriado {
			return domain.Orcamento{}, domain.ErrOrcamentoFechado
		}
		return orcamento, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Orcamento{}, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO orcamento (ordem_servico_id, tipo_orcamento, status)
		VALUES ($1, $2, $3)
		RETURNING id, tipo_orcamento, status`, ordemServicoID, tipo, domain.OrcamentoCriado,
	).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	return orcamento, err
}
