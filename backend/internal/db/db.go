package db

import (
	"backend/internal/config"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func Connect() error {
    connStr := config.AppConfig.DBUrl
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        return fmt.Errorf("error opening DB: %w", err)
    }

	// Connection pool configs : 
	DB.SetMaxOpenConns(20)            
	DB.SetMaxIdleConns(5)             
	DB.SetConnMaxLifetime(30 * time.Minute) 
	DB.SetConnMaxIdleTime(10 * time.Minute)  
    
    // Verify connection works (pings database)
	if err := DB.Ping(); err != nil {
		return fmt.Errorf("error pinging DB: %w", err)
	}

	return nil
}
