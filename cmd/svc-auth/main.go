package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/labstack/echo/v5"
	_ "github.com/lib/pq"

	"geo-project/internal/auth/ent"
	"geo-project/internal/auth/routes"
	apperrors "geo-project/pkg/errors"
)

func main() {
	log.Println("Starting Auth Service...")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("POSTGRES_DB")
	user := os.Getenv("DB_USER")
	pass := os.Getenv("DB_PASS")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbName, pass)

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("?Connection error: %v", err)
	}
	defer client.Close()
	log.Println("Connected to database")

	ctx := context.Background()
	if err := client.Schema.Create(ctx); err != nil {
		log.Fatalf("Error when migrating %v", err)
	}
	log.Println("Migration complete")

	e := echo.New()
	

	e.HTTPErrorHandler = apperrors.CustomHTTPErrorHandler


	api := e.Group("/api/v1/auth")

	routes.RegisterAuthRoutes(api,client)


	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8082"
	}

	log.Printf("Auth service started at %s", portEnv)
	if err := e.Start(":" + portEnv); err != nil {
		log.Fatalf("Server disconnected: %v", err)
	}
}