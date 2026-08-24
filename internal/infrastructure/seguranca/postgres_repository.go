package seguranca

import (
	"context"
	"database/sql"
	"strings"

	"github.com/lazaro-contato/pos-fiap-oficina-mecanica/internal/domain/seguranca"
)

type PostgresRepository struct{ db *sql.DB }

func NewPostgresRepository(db *sql.DB) PostgresRepository { return PostgresRepository{db: db} }

func (repository PostgresRepository) BuscarPorEmail(ctx context.Context, email string) (seguranca.Usuario, error) {
	const query = `SELECT u.id, u.email, u.senha_hash, u.ativo, COALESCE(string_agg(ue.escopo, ','), '')
		FROM usuario u
		JOIN mecanico m ON m.usuario_id = u.id
		LEFT JOIN usuario_escopo ue ON ue.usuario_id = u.id
		WHERE lower(u.email) = lower($1)
		GROUP BY u.id`
	var usuario seguranca.Usuario
	var escopos string
	err := repository.db.QueryRowContext(ctx, query, email).Scan(&usuario.ID, &usuario.Email, &usuario.SenhaHash, &usuario.Ativo, &escopos)
	if escopos != "" {
		usuario.Escopos = strings.Split(escopos, ",")
	}
	return usuario, err
}
