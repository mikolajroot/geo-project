package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"geo-project/internal/analytics/routes"
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
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found. Relying on system environment variables.")
	}

	dsn := database.BuildDSN()
	ctx := context.Background()

	pgxPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("pgx pool connection failed: %v", err)
	}
	defer pgxPool.Close()

	e := echo.New()

	e.HTTPErrorHandler = apperrors.CustomHTTPErrorHandler

	e.Validator = &CustomValidator{validator: validator.New()}

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		if err := pgxPool.Ping(ctx); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")

	jwtSecret := database.ReadSecret("/run/secrets/secret_key")
	routes.RegisterAnalyticsRoutes(api, pgxPool, jwtSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{Addr: ":" + port, Handler: e}

	go func() {
		log.Printf("Starting Catalog Service on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down analytics service: %v", err)
	}
}
