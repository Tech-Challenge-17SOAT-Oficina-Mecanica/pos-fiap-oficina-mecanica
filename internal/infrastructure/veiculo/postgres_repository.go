package veiculo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	application "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/application/veiculo"
	domain "github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/veiculo"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db} }
func (repository PostgresRepository) CadastrarParaCliente(ctx context.Context, clienteID string, cadastro domain.Cadastro) (domain.Veiculo, error) {
	tx, err := repository.db.Begin(ctx)
	if err != nil {
		return domain.Veiculo{}, err
	}
	defer tx.Rollback(ctx)
	var ativo bool
	err = tx.QueryRow(ctx, "SELECT ativo FROM cliente WHERE id = $1", clienteID).Scan(&ativo)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Veiculo{}, application.ErrClienteNaoEncontrado
	}
	if err != nil {
		return domain.Veiculo{}, err
	}
	if !ativo {
		return domain.Veiculo{}, application.ErrClienteInativo
	}
	veiculo := domain.Veiculo{Cadastro: cadastro}
	err = tx.QueryRow(ctx, `INSERT INTO veiculo (cliente_id, placa, marca, modelo, ano) VALUES ($1, $2, $3, $4, $5) RETURNING id, cliente_id, ativo`, clienteID, cadastro.Placa, cadastro.Marca, cadastro.Modelo, cadastro.Ano).Scan(&veiculo.ID, &veiculo.ClienteID, &veiculo.Ativo)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.Veiculo{}, application.ErrPlacaDuplicada
		}
		return domain.Veiculo{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.Veiculo{}, err
	}
	return veiculo, nil
}

func (repository PostgresRepository) BuscarPorIDIncluindoInativo(ctx context.Context, id string) (domain.Veiculo, error) {
	var v domain.Veiculo
	err := repository.db.QueryRow(ctx, `SELECT id,cliente_id,placa,marca,modelo,ano,ativo,COALESCE(inativado_em,'epoch'),COALESCE(inativado_por::text,''),COALESCE(motivo_inativacao,''),version FROM veiculo WHERE id=$1`, id).Scan(&v.ID, &v.ClienteID, &v.Placa, &v.Marca, &v.Modelo, &v.Ano, &v.Ativo, &v.InativadoEm, &v.InativadoPor, &v.Motivo, &v.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return v, application.ErrVeiculoNaoEncontrado
	}
	return v, err
}

func (repository PostgresRepository) BuscarOSAbertas(ctx context.Context, veiculoID string) ([]application.OrdemServicoAberta, error) {
	rows, err := repository.db.Query(ctx, `SELECT id,status FROM ordem_servico WHERE veiculo_id=$1 AND status NOT IN ('ENTREGUE','CANCELADA')`, veiculoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ordens []application.OrdemServicoAberta
	for rows.Next() {
		var ordem application.OrdemServicoAberta
		if err := rows.Scan(&ordem.OrdemServicoID, &ordem.Status); err != nil {
			return nil, err
		}
		ordens = append(ordens, ordem)
	}
	return ordens, rows.Err()
}

func (repository PostgresRepository) Inativar(ctx context.Context, input application.InativarRepositoryInput) (application.Inativacao, error) {
	var result application.Inativacao
	err := repository.db.QueryRow(ctx, `UPDATE veiculo SET ativo=FALSE,inativado_em=CURRENT_TIMESTAMP,inativado_por=$2,motivo_inativacao=NULLIF($3,''),version=version+1 WHERE id=$1 AND ativo RETURNING id,cliente_id,placa,marca,modelo,ano,ativo,inativado_em,inativado_por::text,COALESCE(motivo_inativacao,''),version`, input.VeiculoID, input.InativadoPor, input.Motivo).Scan(&result.Veiculo.ID, &result.Veiculo.ClienteID, &result.Veiculo.Placa, &result.Veiculo.Marca, &result.Veiculo.Modelo, &result.Veiculo.Ano, &result.Veiculo.Ativo, &result.Veiculo.InativadoEm, &result.Veiculo.InativadoPor, &result.Veiculo.Motivo, &result.Veiculo.Version)
	return result, err
}

func (repository PostgresRepository) ExisteAtivoPorPlacaExcetoID(ctx context.Context, placa, id string) (bool, error) {
	var existe bool
	err := repository.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM veiculo WHERE placa=$1 AND ativo AND id<>$2)`, placa, id).Scan(&existe)
	return existe, err
}
func (repository PostgresRepository) ClienteAtivo(ctx context.Context, id string) (bool, error) {
	var ativo bool
	err := repository.db.QueryRow(ctx, `SELECT ativo FROM cliente WHERE id=$1`, id).Scan(&ativo)
	return ativo, err
}
func (repository PostgresRepository) Reativar(ctx context.Context, id, usuarioID string) (application.Reativacao, error) {
	var result application.Reativacao
	err := repository.db.QueryRow(ctx, `UPDATE veiculo SET ativo=TRUE,inativado_em=NULL,inativado_por=NULL,motivo_inativacao=NULL,version=version+1 WHERE id=$1 AND NOT ativo RETURNING id,cliente_id,placa,marca,modelo,ano,ativo,version,CURRENT_TIMESTAMP`, id).Scan(&result.Veiculo.ID, &result.Veiculo.ClienteID, &result.Veiculo.Placa, &result.Veiculo.Marca, &result.Veiculo.Modelo, &result.Veiculo.Ano, &result.Veiculo.Ativo, &result.Veiculo.Version, &result.ReativadoEm)
	result.ReativadoPor = usuarioID
	return result, err
}

func (repository PostgresRepository) ConsultarPorPlaca(ctx context.Context, placa string, incluirInativos bool) (domain.Veiculo, error) {
	query := `SELECT v.id, v.cliente_id, v.placa, v.marca, v.modelo, v.ano, v.ativo, v.version, c.id, c.nome, c.documento FROM veiculo v JOIN cliente c ON c.id = v.cliente_id WHERE v.placa = $1`
	if !incluirInativos {
		query += ` AND v.ativo = TRUE`
	}
	query += ` ORDER BY v.ativo DESC, v.criado_em DESC LIMIT 1`
	var v domain.Veiculo
	err := repository.db.QueryRow(ctx, query, placa).Scan(&v.ID, &v.ClienteID, &v.Placa, &v.Marca, &v.Modelo, &v.Ano, &v.Ativo, &v.Version, &v.Cliente.ID, &v.Cliente.Nome, &v.Cliente.Documento)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Veiculo{}, application.ErrVeiculoNaoEncontrado
	}
	return v, err
}

func (repository PostgresRepository) Atualizar(ctx context.Context, id string, version int, cadastro domain.Cadastro) (domain.Veiculo, error) {
	var v domain.Veiculo
	err := repository.db.QueryRow(ctx, `WITH atualizado AS (
		UPDATE veiculo SET placa=$1,marca=$2,modelo=$3,ano=$4,version=version+1
		WHERE id=$5 AND ativo=TRUE AND version=$6
		RETURNING id,cliente_id,placa,marca,modelo,ano,ativo,version
	) SELECT a.id,a.cliente_id,a.placa,a.marca,a.modelo,a.ano,a.ativo,a.version,c.id,c.nome,c.documento
	FROM atualizado a JOIN cliente c ON c.id=a.cliente_id`, cadastro.Placa, cadastro.Marca, cadastro.Modelo, cadastro.Ano, id, version).Scan(&v.ID, &v.ClienteID, &v.Placa, &v.Marca, &v.Modelo, &v.Ano, &v.Ativo, &v.Version, &v.Cliente.ID, &v.Cliente.Nome, &v.Cliente.Documento)
	if errors.Is(err, pgx.ErrNoRows) {
		var exists bool
		if err = repository.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM veiculo WHERE id=$1 AND ativo)", id).Scan(&exists); err != nil {
			return v, err
		}
		if !exists {
			return v, application.ErrVeiculoNaoEncontrado
		}
		return v, application.ErrVersaoDivergente
	}
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return v, application.ErrPlacaDuplicada
		}
	}
	return v, err
}
