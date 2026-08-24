package database

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

func Open(ctx context.Context) (*pgxpool.Pool, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", env("DB_USER", "oficina"), env("DB_PASSWORD", "oficina"), env("DB_HOST", "localhost"), env("DB_PORT", "5432"), env("DB_NAME", "oficina"))
	}
	return pgxpool.New(ctx, dsn)
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
