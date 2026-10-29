package database

import (
	"context"
	"net"
	"net/url"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func OpenPool() (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), dsn())
}

func dsn() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	mode := os.Getenv("DB_SSLMODE")
	if mode == "" {
		mode = "disable"
	}
	query := url.Values{"sslmode": {mode}}
	if cert := os.Getenv("DB_SSLROOTCERT"); cert != "" {
		query.Set("sslrootcert", cert)
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD")),
		Host:     net.JoinHostPort(os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     "/" + os.Getenv("DB_NAME"),
		RawQuery: query.Encode(),
	}
	return u.String()
}
