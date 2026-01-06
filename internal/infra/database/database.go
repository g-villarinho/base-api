package database

import (
	"database/sql"
	"fmt"

	"github.com/gbvillarinho/base-project/config"
	_ "github.com/mattn/go-sqlite3"
)

func NewConnection(config *config.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", config.SqlLite.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxIdleConns(config.SqlLite.MaxIdle)
	db.SetMaxOpenConns(config.SqlLite.MaxConn)
	db.SetConnMaxLifetime(config.SqlLite.MaxLifeTime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return db, nil
}
