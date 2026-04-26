package main

import (
	"log"
	"net/http"
	"os"


	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"

	"geo-project/pkg/database"
	"geo-project/internal/layers/routes" 
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found. Relying on system environment variables.")
	}

	sqlDB, err := database.NewStandardDB()
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer sqlDB.Close()


	goquDB := goqu.New("postgres", sqlDB)
	log.Println("Goqu Query Builder initialized successfully")

	e := echo.New()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS("http://localhost:3000"))

	e.GET("/health", func(c *echo.Context) error {
		var result int
		_, err := goquDB.Select(1).ScanVal(&result)
		if err != nil {
			return c.JSON(http.StatusServiceUnavailable, map[string]string{"status": "db_error"})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
	})


	api := e.Group("/api/v1/layers")

	routes.RegisterLayersRoutes(api, goquDB) 

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Catalog Service on port %s\n", port)
	if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server failed to start: %v", err)
	}
}