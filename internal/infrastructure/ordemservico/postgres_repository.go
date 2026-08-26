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

func (repository PostgresRepository) Criar(ctx context.Context, input application.CriarInput) (ordem domain.OrdemDeServico, err error) {
	tx, err := repository.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ordem, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var clienteID string
	err = tx.QueryRow(ctx, `SELECT id FROM cliente WHERE id = $1 AND ativo FOR SHARE`, input.ClienteID).Scan(&clienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}

	var veiculoClienteID, placa string
	err = tx.QueryRow(ctx, `SELECT cliente_id, placa FROM veiculo WHERE id = $1 AND ativo FOR SHARE`, input.VeiculoID).Scan(&veiculoClienteID, &placa)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrVeiculoNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}
	if veiculoClienteID != input.ClienteID {
		return ordem, application.ErrVeiculoNaoVinculadoCliente
	}

	err = tx.QueryRow(ctx, `INSERT INTO ordem_servico (cliente_id, veiculo_id, placa_veiculo, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, cliente_id, veiculo_id, placa_veiculo, status, criada_em`, input.ClienteID, input.VeiculoID, placa, domain.StatusRecebida).
		Scan(&ordem.ID, &ordem.ClienteID, &ordem.VeiculoID, &ordem.PlacaVeiculo, &ordem.Status, &ordem.CriadaEm)
	if err != nil {
		return ordem, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return ordem, nil
}

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

func (repository PostgresRepository) RegistrarServicos(ctx context.Context, ordemServicoID string, cadastros []domain.ServicoCadastro) (application.ResultadoRegistroServicos, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	defer tx.Rollback(ctx)

	var status string
	err = tx.QueryRow(ctx, "SELECT status FROM ordem_servico WHERE id = $1 FOR UPDATE", ordemServicoID).Scan(&status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroServicos{}, application.ErrOrdemServicoNaoEncontrada
	}
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	tipo, err := domain.TipoOrcamentoParaServico(status)
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}

	var orcamento domain.Orcamento
	err = tx.QueryRow(ctx, `
		SELECT id, tipo_orcamento, status
		FROM orcamento
		WHERE ordem_servico_id = $1 AND tipo_orcamento = $2 AND status = 'CRIADO'
		ORDER BY criado_em DESC LIMIT 1 FOR UPDATE`, ordemServicoID, tipo,
	).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	if errors.Is(err, pgx.ErrNoRows) {
		return application.ResultadoRegistroServicos{}, domain.ErrOrcamentoAplicavelNaoEncontrado
	}
	if err != nil {
		return application.ResultadoRegistroServicos{}, err
	}

	resultado := application.ResultadoRegistroServicos{Orcamento: orcamento, Servicos: make([]domain.ServicoRegistrado, 0, len(cadastros))}
	for _, cadastro := range cadastros {
		var descricao string
		var valor float64
		var ativo bool
		err = tx.QueryRow(ctx, "SELECT nome, valor, ativo FROM servico WHERE id = $1", cadastro.ServicoID).Scan(&descricao, &valor, &ativo)
		if errors.Is(err, pgx.ErrNoRows) {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoNaoEncontrado
		}
		if err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		if !ativo {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoInativo
		}
		var existe bool
		if err = tx.QueryRow(ctx, "SELECT EXISTS (SELECT 1 FROM orcamento_item WHERE orcamento_id = $1 AND servico_id = $2)", orcamento.ID, cadastro.ServicoID).Scan(&existe); err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		if existe {
			return application.ResultadoRegistroServicos{}, domain.ErrServicoDuplicado
		}
		if _, err = tx.Exec(ctx, `
			INSERT INTO orcamento_item (orcamento_id, servico_id, tipo_item, descricao, quantidade, valor_unitario, valor_total, observacao)
			VALUES ($1, $2, 'SERVICO', $3, 1, $4, $4, NULLIF($5, ''))`, orcamento.ID, cadastro.ServicoID, descricao, valor, cadastro.Observacao,
		); err != nil {
			return application.ResultadoRegistroServicos{}, err
		}
		resultado.Servicos = append(resultado.Servicos, domain.ServicoRegistrado{ServicoID: cadastro.ServicoID, Descricao: descricao, ValorUnitario: valor, Observacao: cadastro.Observacao})
	}
	if err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(valor_total), 0) FROM orcamento_item WHERE orcamento_id = $1", orcamento.ID).Scan(&resultado.Orcamento.ValorTotal); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	if _, err = tx.Exec(ctx, "UPDATE orcamento SET data_atualizacao = CURRENT_TIMESTAMP WHERE id = $1", orcamento.ID); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	if err = tx.Commit(ctx); err != nil {
		return application.ResultadoRegistroServicos{}, err
	}
	return resultado, nil
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
	if tipo == domain.OrcamentoComplementar {
		var principalID string
		err = tx.QueryRow(ctx, "SELECT id FROM orcamento WHERE ordem_servico_id = $1 AND tipo_orcamento = 'PRINCIPAL'", ordemServicoID).Scan(&principalID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Orcamento{}, domain.ErrOrcamentoPrincipalNaoEncontrado
		}
		if err != nil {
			return domain.Orcamento{}, err
		}
		err = tx.QueryRow(ctx, `
			INSERT INTO orcamento (ordem_servico_id, orcamento_original_id, tipo_orcamento, status)
			VALUES ($1, $2, $3, $4)
			RETURNING id, tipo_orcamento, status`, ordemServicoID, principalID, tipo, domain.OrcamentoCriado,
		).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
		return orcamento, err
	}
	err = tx.QueryRow(ctx, `
		INSERT INTO orcamento (ordem_servico_id, tipo_orcamento, status)
		VALUES ($1, $2, $3)
		RETURNING id, tipo_orcamento, status`, ordemServicoID, tipo, domain.OrcamentoCriado,
	).Scan(&orcamento.ID, &orcamento.Tipo, &orcamento.Status)
	return orcamento, err
}
