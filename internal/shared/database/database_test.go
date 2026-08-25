package database

import "testing"

func TestOpen(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://oficina:oficina@postgres:5432/oficina?sslmode=disable")
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestOpenComVariaveisSeparadas(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "postgres")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "oficina")
	t.Setenv("DB_USER", "oficina")
	t.Setenv("DB_PASSWORD", "oficina")
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}
