package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"geo-project/pkg/database"
)

type migrationFile struct {
	version int64
	name    string
	path    string
}

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found. Relying on system environment variables.")
	}

	db, err := database.NewStandardDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "db/migrations"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := runMigrations(ctx, db, migrationsDir); err != nil {
		log.Fatalf("Migration worker failed: %v", err)
	}

	log.Println("Migration worker completed successfully")
}

func runMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	applied, err := loadAppliedVersions(ctx, db)
	if err != nil {
		return err
	}

	files, err := loadMigrationFiles(migrationsDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if applied[file.version] {
			log.Printf("Skipping migration %s", file.name)
			continue
		}

		if err := applyMigration(ctx, db, file); err != nil {
			return err
		}
	}

	return nil
}

func ensureMigrationsTable(ctx context.Context, db *sql.DB) error {
	const query = `
CREATE TABLE IF NOT EXISTS schema_migrations (
    version BIGINT PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);`

	if _, err := db.ExecContext(ctx, query); err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	return nil
}

func loadAppliedVersions(ctx context.Context, db *sql.DB) (map[int64]bool, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[int64]bool)
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration version: %w", err)
		}
		applied[version] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return applied, nil
}

func loadMigrationFiles(migrationsDir string) ([]migrationFile, error) {
	matches, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return nil, fmt.Errorf("list migration files: %w", err)
	}

	files := make([]migrationFile, 0, len(matches))
	for _, path := range matches {
		name := filepath.Base(path)
		version, err := parseMigrationVersion(name)
		if err != nil {
			return nil, err
		}

		files = append(files, migrationFile{
			version: version,
			name:    name,
			path:    path,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].version == files[j].version {
			return files[i].name < files[j].name
		}
		return files[i].version < files[j].version
	})

	return files, nil
}

func parseMigrationVersion(filename string) (int64, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) == 0 || parts[0] == "" {
		return 0, fmt.Errorf("invalid migration filename: %s", filename)
	}

	version, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parse migration version from %s: %w", filename, err)
	}

	return version, nil
}

func applyMigration(ctx context.Context, db *sql.DB, file migrationFile) error {
	log.Printf("Applying migration %s", file.name)

	contents, err := os.ReadFile(file.path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file.name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin migration transaction for %s: %w", file.name, err)
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute migration %s: %w", file.name, err)
	}

	if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations (version, name, applied_at) VALUES ($1, $2, NOW())`, file.version, file.name); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("record migration %s: %w", file.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit migration %s: %w", file.name, err)
	}

	return nil
}
