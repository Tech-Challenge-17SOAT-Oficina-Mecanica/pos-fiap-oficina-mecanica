package database

import (
	"context"
	"database/sql"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn())
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func OpenPool() (*pgxpool.Pool, error) {
	return pgxpool.New(context.Background(), dsn())
}

func dsn() string {
	if value := os.Getenv("DATABASE_URL"); value != "" {
		return value
	}
	return "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME") + "?sslmode=disable"
}
