package seguranca

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestBuscarPorEmail(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL não configurada")
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	usuario, err := NewPostgresRepository(db).BuscarPorEmail(context.Background(), "mecanico@oficina.local")
	if err != nil || usuario.Email != "mecanico@oficina.local" || len(usuario.Escopos) == 0 {
		t.Fatalf("usuario: %#v, %v", usuario, err)
	}
}
