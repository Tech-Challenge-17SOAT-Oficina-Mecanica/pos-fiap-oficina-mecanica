package veiculo

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db} }
func (repository PostgresRepository) CadastrarParaCliente(ctx context.Context, clienteID string, cadastro domain.Cadastro) (domain.Veiculo, error) {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.Veiculo{}, err
	}
	defer tx.Rollback()
	var ativo bool
	err = tx.QueryRowContext(ctx, "SELECT ativo FROM cliente WHERE id = $1", clienteID).Scan(&ativo)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Veiculo{}, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return domain.Veiculo{}, err
	}
	if !ativo {
		return domain.Veiculo{}, application.ErrClienteInativo
	}
	veiculo := domain.Veiculo{Cadastro: cadastro}
	err = tx.QueryRowContext(ctx, `INSERT INTO veiculo (cliente_id, placa, marca, modelo, ano) VALUES ($1, $2, $3, $4, $5) RETURNING id, cliente_id, ativo`, clienteID, cadastro.Placa, cadastro.Marca, cadastro.Modelo, cadastro.Ano).Scan(&veiculo.ID, &veiculo.ClienteID, &veiculo.Ativo)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Veiculo{}, application.ErrPlacaDuplicada
		}
		return domain.Veiculo{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Veiculo{}, err
	}
	return veiculo, nil
}
