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
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var clienteID string
	err = tx.QueryRow(ctx, `SELECT id FROM cliente WHERE id = $1 AND ativo FOR SHARE`, input.ClienteID).Scan(&clienteID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}

	var veiculoClienteID, placa string
	err = tx.QueryRow(ctx, `SELECT cliente_id, placa FROM veiculo WHERE id = $1 AND ativo FOR SHARE`, input.VeiculoID).
		Scan(&veiculoClienteID, &placa)
	if errors.Is(err, pgx.ErrNoRows) {
		return ordem, application.ErrVeiculoNaoEncontrado
	}
	if err != nil {
		return ordem, err
	}
	if veiculoClienteID != input.ClienteID {
		return ordem, application.ErrVeiculoNaoVinculadoCliente
	}

	const query = `INSERT INTO ordem_servico (cliente_id, veiculo_id, placa_veiculo, status)
		VALUES ($1, $2, $3, $4)
		RETURNING id, cliente_id, veiculo_id, placa_veiculo, status, criada_em`
	err = tx.QueryRow(ctx, query, input.ClienteID, input.VeiculoID, placa, domain.StatusRecebida).
		Scan(&ordem.ID, &ordem.ClienteID, &ordem.VeiculoID, &ordem.PlacaVeiculo, &ordem.Status, &ordem.CriadaEm)
	if err != nil {
		return ordem, err
	}
	if err = tx.Commit(ctx); err != nil {
		return domain.OrdemDeServico{}, err
	}
	return ordem, nil
}
