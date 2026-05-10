package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"geo-project/internal/analytics/routes"
	"geo-project/pkg/database"
	"geo-project/pkg/errors"
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

	e.HTTPErrorHandler = errors.CustomHTTPErrorHandler

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

	log.Printf("Starting Catalog Service on port %s\n", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}
