package database

import (
	"database/sql"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	migrate "github.com/rubenv/sql-migrate"
)

type MigrationRunner struct {
	db     *sql.DB
	source *migrate.FileMigrationSource
}

func NewMigrationRunner(db *sql.DB, migrationsDir string) *MigrationRunner {
	return &MigrationRunner{
		db: db,
		source: &migrate.FileMigrationSource{
			Dir: migrationsDir,
		},
	}
}

func (r *MigrationRunner) Up() error {
	r.printHeader()

	migrations, err := r.source.FindMigrations()
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}

	records, err := migrate.GetMigrationRecords(r.db, "postgres")
	if err != nil {
		return fmt.Errorf("get migration records: %w", err)
	}

	appliedMap := make(map[string]bool)
	for _, record := range records {
		appliedMap[record.Id] = true
	}

	pending := 0
	for _, m := range migrations {
		if !appliedMap[m.Id] {
			pending++
		}
	}

	if pending == 0 {
		fmt.Println()
		fmt.Println("  Database is up to date. No pending migrations.")
		fmt.Println()
		return nil
	}

	start := time.Now()
	n, err := migrate.Exec(r.db, "postgres", r.source, migrate.Up)
	duration := time.Since(start)

	if err != nil {
		r.printError(err)
		return fmt.Errorf("apply migrations: %w", err)
	}

	for _, m := range migrations {
		if !appliedMap[m.Id] {
			r.printMigrationDetails(m, "up")
		}
	}

	fmt.Printf("\n  Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Println()
	fmt.Println("  ----------------------------------------")
	fmt.Printf("  [OK] %d migration(s) applied successfully\n", n)
	fmt.Println("  ----------------------------------------")
	fmt.Println()

	return nil
}

func (r *MigrationRunner) Down() error {
	r.printHeader()

	records, err := migrate.GetMigrationRecords(r.db, "postgres")
	if err != nil {
		return fmt.Errorf("get migration records: %w", err)
	}

	if len(records) == 0 {
		fmt.Println()
		fmt.Println("  No migrations to revert.")
		fmt.Println()
		return nil
	}

	migrations, err := r.source.FindMigrations()
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}

	lastApplied := records[len(records)-1].Id
	var migrationToRevert *migrate.Migration
	for _, m := range migrations {
		if m.Id == lastApplied {
			migrationToRevert = m
			break
		}
	}

	start := time.Now()
	n, err := migrate.ExecMax(r.db, "postgres", r.source, migrate.Down, 1)
	duration := time.Since(start)

	if err != nil {
		r.printError(err)
		return fmt.Errorf("revert migration: %w", err)
	}

	if migrationToRevert != nil {
		r.printMigrationDetails(migrationToRevert, "down")
	}

	fmt.Printf("\n  Duration: %v\n", duration.Round(time.Millisecond))
	fmt.Println()
	fmt.Println("  ----------------------------------------")
	fmt.Printf("  [OK] %d migration(s) reverted\n", n)
	fmt.Println("  ----------------------------------------")
	fmt.Println()

	return nil
}

func (r *MigrationRunner) Status() error {
	r.printHeader()

	migrations, err := r.source.FindMigrations()
	if err != nil {
		return fmt.Errorf("find migrations: %w", err)
	}

	records, err := migrate.GetMigrationRecords(r.db, "postgres")
	if err != nil {
		return fmt.Errorf("get migration records: %w", err)
	}

	appliedMap := make(map[string]time.Time)
	for _, record := range records {
		appliedMap[record.Id] = record.AppliedAt
	}

	fmt.Println()
	fmt.Printf("  %-40s %-12s %s\n", "MIGRATION", "STATUS", "APPLIED AT")
	fmt.Printf("  %s\n", strings.Repeat("-", 70))

	applied := 0
	pending := 0

	for _, m := range migrations {
		if appliedAt, ok := appliedMap[m.Id]; ok {
			fmt.Printf("  %-40s %-12s %s\n", m.Id, "[applied]", appliedAt.Format("2006-01-02 15:04:05"))
			applied++
		} else {
			fmt.Printf("  %-40s %-12s\n", m.Id, "[pending]")
			pending++
		}
	}

	fmt.Println()
	fmt.Println("  ----------------------------------------")
	fmt.Printf("  Applied: %d | Pending: %d | Total: %d\n", applied, pending, applied+pending)
	fmt.Println("  ----------------------------------------")
	fmt.Println()

	return nil
}

func (r *MigrationRunner) printHeader() {
	fmt.Println()
	fmt.Println("  ========================================")
	fmt.Println("  DATABASE MIGRATIONS")
	fmt.Println("  ========================================")
}

func (r *MigrationRunner) printMigrationDetails(m *migrate.Migration, direction string) {
	fmt.Println()
	if direction == "up" {
		fmt.Printf("  >> Applying: %s\n", m.Id)
	} else {
		fmt.Printf("  << Reverting: %s\n", m.Id)
	}

	var sqlContent string
	if direction == "up" {
		sqlContent = strings.Join(m.Up, "\n")
	} else {
		sqlContent = strings.Join(m.Down, "\n")
	}

	tables := extractTables(sqlContent, direction)
	indexes := extractIndexes(sqlContent, direction)

	if len(tables) > 0 {
		fmt.Println()
		fmt.Println("  Tables:")
		for _, table := range tables {
			if direction == "up" {
				fmt.Printf("    + %s (created)\n", table)
			} else {
				fmt.Printf("    - %s (dropped)\n", table)
			}
		}
	}

	if len(indexes) > 0 {
		fmt.Println()
		fmt.Println("  Indexes:")
		for _, index := range indexes {
			if direction == "up" {
				fmt.Printf("    + %s\n", index)
			} else {
				fmt.Printf("    - %s\n", index)
			}
		}
	}
}

func (r *MigrationRunner) printError(err error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  ----------------------------------------")
	fmt.Fprintln(os.Stderr, "  [ERROR] Migration failed")
	fmt.Fprintln(os.Stderr, "  ----------------------------------------")
	fmt.Fprintf(os.Stderr, "\n  %v\n\n", err)
}

func extractTables(sql string, direction string) []string {
	var tables []string
	var re *regexp.Regexp

	if direction == "up" {
		re = regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	} else {
		re = regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(\w+)`)
	}

	matches := re.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			tables = append(tables, match[1])
		}
	}

	return tables
}

func extractIndexes(sql string, direction string) []string {
	var indexes []string
	var re *regexp.Regexp

	if direction == "up" {
		re = regexp.MustCompile(`(?i)CREATE\s+(?:UNIQUE\s+)?INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	} else {
		re = regexp.MustCompile(`(?i)DROP\s+INDEX\s+(?:IF\s+EXISTS\s+)?(\w+)`)
	}

	matches := re.FindAllStringSubmatch(sql, -1)
	for _, match := range matches {
		if len(match) > 1 {
			indexes = append(indexes, match[1])
		}
	}

	return indexes
}
