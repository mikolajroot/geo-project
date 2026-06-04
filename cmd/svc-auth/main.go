package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	_ "github.com/lib/pq"

	"geo-project/internal/auth/ent"
	"geo-project/internal/auth/routes"
	"geo-project/pkg/database"
	apperrors "geo-project/pkg/errors"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func main() {
	log.Println("Starting Auth Service...")

	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	dbName := os.Getenv("POSTGRES_DB")
	user := database.ReadSecret("/run/secrets/db_user")
	pass := database.ReadSecret("/run/secrets/db_password")

	dsn := fmt.Sprintf("host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbName, pass)

	sqlDB, err := database.NewStandardDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer sqlDB.Close()

	client, err := ent.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Connection error: %v", err)
	}
	defer client.Close()
	log.Println("Connected to database")

	e := echo.New()

	e.HTTPErrorHandler = apperrors.CustomHTTPErrorHandler

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		if err := sqlDB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1/auth")

	jwtSecret := database.ReadSecret("/run/secrets/secret_key")

	routes.RegisterAuthRoutes(api, client, sqlDB, jwtSecret)

	portEnv := os.Getenv("PORT")
	if portEnv == "" {
		portEnv = "8081"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{Addr: ":" + portEnv, Handler: e}

	go func() {
		log.Printf("Auth service started at %s", portEnv)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server disconnected: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down auth service: %v", err)
	}
}
