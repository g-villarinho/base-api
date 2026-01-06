package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	"github.com/gbvillarinho/base-project/config"
	"github.com/gbvillarinho/base-project/internal/infra/database"
	_ "github.com/mattn/go-sqlite3"
)

const migrationsDir = "internal/infra/database/migrations"

func main() {
	direction := flag.String("direction", "up", "Migration direction: up, down, or status")
	flag.Parse()

	cfg, err := config.NewConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", cfg.SqlLite.DatabaseName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to database: %v\n", err)
		os.Exit(1)
	}

	runner := database.NewMigrationRunner(db, migrationsDir)

	var runErr error
	switch *direction {
	case "up":
		runErr = runner.Up()
	case "down":
		runErr = runner.Down()
	case "status":
		runErr = runner.Status()
	default:
		fmt.Fprintf(os.Stderr, "Invalid direction: %s. Use: up, down, or status\n", *direction)
		os.Exit(1)
	}

	if runErr != nil {
		os.Exit(1)
	}
}
