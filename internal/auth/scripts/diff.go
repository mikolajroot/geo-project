package main

import (
	"context"
	"log"
	"os"
	"strings"

	atlasmigrate "ariga.io/atlas/sql/migrate" // <-- NOWY IMPORT
	"ariga.io/atlas/sql/sqltool"
	"entgo.io/ent/dialect/sql/schema"
	_ "github.com/lib/pq"

	"geo-project/internal/auth/ent/migrate"
)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Write migration name, for example go run diff.go create_users")
	}
	migrationName := os.Args[1]

	beforeCount, err := countUpMigrations("./db/migrations")
	if err != nil {
		log.Fatalf("Error reading migrations before diff: %v", err)
	}

	dir, err := sqltool.NewGolangMigrateDir("./db/migrations")
	if err != nil {
		log.Fatalf("Erro loading migrations: %v", err)
	}

	sum, err := dir.Checksum()
	if err != nil {
		log.Fatalf("Checksum error: %v", err)
	}
	if err := atlasmigrate.WriteSumFile(dir, sum); err != nil {
		log.Fatalf("Error saving atlas.sum: %v", err)
	}
	// --------------------------------------------------------------

	opts := []schema.MigrateOption{
		schema.WithDir(dir),
		schema.WithMigrationMode(schema.ModeInspect),
		schema.WithDialect("postgres"),
		schema.WithDropIndex(true),
		schema.WithFormatter(sqltool.GolangMigrateFormatter),
	}

	dbUrl := "postgres://postgres:password@localhost:5432/geo-project?sslmode=disable"
	if envUrl := os.Getenv("DB_URL"); envUrl != "" {
		dbUrl = envUrl
	}

	err = migrate.NamedDiff(context.Background(), dbUrl, migrationName, opts...)
	if err != nil {
		log.Fatalf("Error generating diff: %v", err)
	}

	afterCount, err := countUpMigrations("./db/migrations")
	if err != nil {
		log.Fatalf("Error reading migrations after diff: %v", err)
	}

	if afterCount == beforeCount {
		log.Printf("No schema changes detected. Migration was not created: %s", migrationName)
		return
	}

	log.Printf("Generated migrations : %s\n", migrationName)
}

func countUpMigrations(path string) (int, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if strings.HasSuffix(entry.Name(), ".up.sql") {
			count++
		}
	}

	return count, nil
}
