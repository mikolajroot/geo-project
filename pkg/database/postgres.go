package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func readSecret(filepath string) string {
	data, err := os.ReadFile(filepath)
	if err != nil {

		log.Fatalf("Cannot read secret %s: %v", filepath, err)
	}
	return strings.TrimSpace(string(data))
}

func BuildDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("POSTGRES_DB")

	userFile := os.Getenv("DB_USER_FILE")
	passFile := os.Getenv("DB_PASSWORD_FILE")

	if userFile == "" {
		userFile = "/run/secrets/db_user"
	}
	if passFile == "" {
		passFile = "/run/secrets/db_password"
	}

	user := readSecret(userFile)
	pass := readSecret(passFile)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		user, pass, host, port, dbName)
}

func NewStandardDB() (*sql.DB, error) {
	dsn := BuildDSN()

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("error when opening pgsql database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Error ping: %w", err)
	}

	log.Println("Connected to pgsql database")
	return db, nil
}
