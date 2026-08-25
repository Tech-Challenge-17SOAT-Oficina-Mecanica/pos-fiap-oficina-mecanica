package seguranca

import (
	"context"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type PostgresRepository struct{ db *pgxpool.Pool }

func NewPostgresRepository(db *pgxpool.Pool) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) BuscarPorEmail(ctx context.Context, email string) (seguranca.Usuario, error) {
	const query = `SELECT u.id, u.email, u.senha_hash, u.ativo, COALESCE(string_agg(ue.escopo, ','), '')
		FROM usuario u
		JOIN mecanico m ON m.usuario_id = u.id
		LEFT JOIN usuario_escopo ue ON ue.usuario_id = u.id
		WHERE lower(u.email) = lower($1)
		GROUP BY u.id`
	var usuario seguranca.Usuario
	var escopos string
	err := repository.db.QueryRow(ctx, query, email).Scan(&usuario.ID, &usuario.Email, &usuario.SenhaHash, &usuario.Ativo, &escopos)
	if escopos != "" {
		usuario.Escopos = strings.Split(escopos, ",")
	}
	return usuario, err
}
