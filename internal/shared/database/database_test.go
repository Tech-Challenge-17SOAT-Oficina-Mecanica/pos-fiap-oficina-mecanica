package database

import (
	"net/url"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOpenPool(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://oficina:oficina@postgres:5432/oficina?sslmode=disable")
	db, err := OpenPool()
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
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("DB_SSLROOTCERT", "")
	db, err := OpenPool()
	if err != nil {
		t.Fatal(err)
	}
	db.Close()
}

func TestDSNTLSAndCredentials(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "oficina.example.rds.amazonaws.com")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "oficina")
	t.Setenv("DB_USER", "oficina")
	t.Setenv("DB_PASSWORD", "password@:/?#%")
	t.Setenv("DB_SSLMODE", "verify-full")
	t.Setenv("DB_SSLROOTCERT", "/etc/ssl/certs/aws-rds-global-bundle.pem")
	u, err := url.Parse(dsn())
	if err != nil {
		t.Fatal(err)
	}
	password, _ := u.User.Password()
	if password != "password@:/?#%" || u.Query().Get("sslmode") != "verify-full" || u.Query().Get("sslrootcert") != "/etc/ssl/certs/aws-rds-global-bundle.pem" {
		t.Fatal("DSN must preserve credentials and TLS verification settings")
	}
}

func TestDSNLocalDefault(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_PORT", "5432")
	t.Setenv("DB_NAME", "oficina")
	t.Setenv("DB_USER", "oficina")
	t.Setenv("DB_PASSWORD", "oficina")
	t.Setenv("DB_SSLMODE", "")
	t.Setenv("DB_SSLROOTCERT", "")
	config, err := pgxpool.ParseConfig(dsn())
	if err != nil {
		t.Fatal(err)
	}
	if config.ConnConfig.TLSConfig != nil || config.ConnConfig.Database != "oficina" {
		t.Fatal("Local connection must continue working without TLS")
	}
}

func TestDatabaseURLTakesPrecedence(t *testing.T) {
	const explicit = "postgres://user:password@localhost:5432/explicit?sslmode=disable"
	t.Setenv("DATABASE_URL", explicit)
	t.Setenv("DB_SSLMODE", "verify-full")
	if dsn() != explicit {
		t.Fatal("DATABASE_URL must take precedence over individual settings")
	}
}
