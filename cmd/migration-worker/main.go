package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	cmd := "up"
	if len(os.Args) > 1 {
		cmd = strings.ToLower(os.Args[1])
	}

	switch cmd {
	case "up":
		if err := runMigrations(ctx, db, migrationsDir); err != nil {
			log.Fatalf("Migration UP failed: %v", err)
		}
	case "down":
		if err := runDown(ctx, db, migrationsDir); err != nil {
			log.Fatalf("Migration DOWN failed: %v", err)
		}
	case "drop":
		if err := runDrop(ctx, db); err != nil {
			log.Fatalf("Migration DROP failed: %v", err)
		}
	default:
		log.Fatalf("Unknown command: %s. Use 'up', 'down', or 'drop'.", cmd)
	}

	log.Printf("Migration worker completed command '%s' successfully", cmd)
}

func runMigrations(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	appliedList, err := loadAppliedVersions(ctx, db)
	if err != nil {
		return err
	}

	applied := make(map[int64]bool)
	for _, v := range appliedList {
		applied[v] = true
	}

	files, err := loadMigrationFiles(migrationsDir, ".up.sql")
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

func runDown(ctx context.Context, db *sql.DB, migrationsDir string) error {
	if err := ensureMigrationsTable(ctx, db); err != nil {
		return err
	}

	appliedList, err := loadAppliedVersions(ctx, db)
	if err != nil {
		return err
	}

	if len(appliedList) == 0 {
		log.Println("No applied migrations to revert")
		return nil
	}

	lastApplied := appliedList[len(appliedList)-1]

	files, err := loadMigrationFiles(migrationsDir, ".down.sql")
	if err != nil {
		return err
	}

	var targetFile *migrationFile
	for _, f := range files {
		if f.version == lastApplied {
			targetFile = &f
			break
		}
	}

	if targetFile == nil {
		return fmt.Errorf("down migration file not found for version %d", lastApplied)
	}

	return applyDownMigration(ctx, db, *targetFile)
}

func runDrop(ctx context.Context, db *sql.DB) error {
	log.Println("Dropping public schema...")
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `DROP SCHEMA public CASCADE; CREATE SCHEMA public;`); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("drop schema public: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	log.Println("Schema public dropped and recreated successfully")
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

func loadAppliedVersions(ctx context.Context, db *sql.DB) ([]int64, error) {
	rows, err := db.QueryContext(ctx, `SELECT version FROM schema_migrations ORDER BY version ASC`)
	if err != nil {
		return nil, fmt.Errorf("load applied migrations: %w", err)
	}
	defer rows.Close()

	var applied []int64
	for rows.Next() {
		var version int64
		if err := rows.Scan(&version); err != nil {
			return nil, fmt.Errorf("scan applied migration version: %w", err)
		}
		applied = append(applied, version)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate applied migrations: %w", err)
	}

	return applied, nil
}

func loadMigrationFiles(migrationsDir string, suffix string) ([]migrationFile, error) {
	matches, err := filepath.Glob(filepath.Join(migrationsDir, "*"+suffix))
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

func applyDownMigration(ctx context.Context, db *sql.DB, file migrationFile) error {
	log.Printf("Reverting migration %s", file.name)

	contents, err := os.ReadFile(file.path)
	if err != nil {
		return fmt.Errorf("read migration %s: %w", file.name, err)
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin down migration transaction for %s: %w", file.name, err)
	}

	if _, err := tx.ExecContext(ctx, string(contents)); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("execute down migration %s: %w", file.name, err)
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM schema_migrations WHERE version = $1`, file.version); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("remove migration record %s: %w", file.name, err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit down migration %s: %w", file.name, err)
	}

	return nil
}