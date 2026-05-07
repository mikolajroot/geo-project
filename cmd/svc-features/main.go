package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

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
	e.Validator = &CustomValidator{validator: validator.New()}

	
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

	log.Printf("Starting Spatial Features Service on port %s\n", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}