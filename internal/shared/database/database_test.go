package database

import (
	"context"
	"testing"
)

func TestOpen(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://oficina:oficina@localhost:5432/oficina?sslmode=disable")
	db, err := Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestOpenComVariaveisSeparadas(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "oficina")
	t.Setenv("DB_USER", "oficina")
	t.Setenv("DB_PASSWORD", "oficina")
	db, err := Open(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestEnv(t *testing.T) {
	t.Setenv("DATABASE_TESTE", "valor")
	if env("DATABASE_TESTE", "padrao") != "valor" || env("DATABASE_AUSENTE", "padrao") != "padrao" {
		t.Fatal("variável inválida")
	}
}
