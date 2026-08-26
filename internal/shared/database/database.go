package database

import (
	"context"
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
	return "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME") + "?sslmode=disable"
}
