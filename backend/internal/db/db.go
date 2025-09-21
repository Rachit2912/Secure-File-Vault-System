package db

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
    connStr := os.Getenv("DB_URL")
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        return fmt.Errorf("error opening DB: %w", err)
    }
    return DB.Ping()
}
