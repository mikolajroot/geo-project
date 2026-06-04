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

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"geo-project/internal/features/handler"
	"geo-project/internal/features/routes"
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

	sqlDB, err := database.NewStandardDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer sqlDB.Close()

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}))
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}
	log.Println("GORM ORM initialized successfully")

	e := echo.New()
	e.HTTPErrorHandler = apperrors.CustomHTTPErrorHandler

	v := validator.New()
	handler.RegisterCustomValidators(v)
	e.Validator = &CustomValidator{validator: v}

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/health", func(c *echo.Context) error {
		rawDB, _ := gormDB.DB()
		if err := rawDB.Ping(); err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})

	api := e.Group("/api/v1")
	jwtSecret := database.ReadSecret("/run/secrets/secret_key")

	routes.RegisterFeaturesRoutes(api, gormDB, jwtSecret)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8082"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	server := &http.Server{Addr: ":" + port, Handler: e}

	go func() {
		log.Printf("Starting Spatial Features Service on port %s\n", port)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Error shutting down features service: %v", err)
	}
}
