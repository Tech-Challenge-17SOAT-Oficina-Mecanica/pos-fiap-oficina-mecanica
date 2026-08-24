package database

import (
	"database/sql"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open() (*sql.DB, error) {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://" + os.Getenv("DB_USER") + ":" + os.Getenv("DB_PASSWORD") + "@" + os.Getenv("DB_HOST") + ":" + os.Getenv("DB_PORT") + "/" + os.Getenv("DB_NAME") + "?sslmode=disable"
	}
	db, err := sql.Open("pgx", dsn)
	if err != nil { return nil, err }
	return db, db.Ping()
}
