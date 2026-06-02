package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"geo-project/internal/annotations/routes"
	seeder "geo-project/internal/annotations/seed"
	revisionRoutes "geo-project/internal/revisions/routes"

	"github.com/go-playground/validator/v10"

	"geo-project/pkg/database"
	apperrors "geo-project/pkg/errors"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.validator.Struct(i)
}

func main() {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://admin:secretpassword@localhost:27017/"
	}
	dbName := os.Getenv("MONGO_DB_NAME")
	if dbName == "" {
		dbName = "geo_annotations"
	}
	jwtSecret := database.ReadSecret("/run/secrets/secret_key")

	client, err := database.GetMongoClient(mongoURI)
	if err != nil {
		log.Fatalf("Critical: Could not connect to MongoDB: %v", err)
	}

	db := client.Database(dbName)

	var sqlDBClose func()
	if os.Getenv("SEED_DB") == "true" {
		sqlDB, err := database.NewStandardDB()
		if err != nil {
			log.Printf("annotations seeding skipped: postgres connection failed: %v", err)
		} else {
			sqlDBClose = func() {
				if err := sqlDB.Close(); err != nil {
					log.Printf("Error closing postgres connection: %v", err)
				}
			}

			seedCount := 200
			if v := os.Getenv("SEED_DB_COUNT"); v != "" {
				if n, convErr := strconv.Atoi(v); convErr == nil && n > 0 {
					seedCount = n
				}
			}

			seedCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			dbSeeder := seeder.NewSeeder(db, sqlDB)
			if err := dbSeeder.SeedDomainData(seedCtx, seedCount); err != nil {
				log.Printf("annotations db seeding failed: %v", err)
			}
		}
	}
	if sqlDBClose != nil {
		defer sqlDBClose()
	}

	e := echo.New()
	e.HTTPErrorHandler = apperrors.CustomHTTPErrorHandler
	e.Validator = &CustomValidator{validator: validator.New()}
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	api := e.Group("/api/v1")

	e.GET("/health", func(c *echo.Context) error {
		if err := client.Ping(c.Request().Context(), nil); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	routes.RegisterAnnotationsRoutes(api, jwtSecret, db)
	revisionRoutes.RegisterRevisionsRoutes(api, jwtSecret, db)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT)
	defer stop()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}
	s := http.Server{Addr: fmt.Sprintf(":%s", port), Handler: e}

	go func() {
		log.Printf("Annotations server listening on %s", s.Addr)
		if err := s.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("Server error: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	if err := client.Disconnect(shutdownCtx); err != nil {
		log.Printf("Error disconnecting MongoDB: %v", err)
	}
}
