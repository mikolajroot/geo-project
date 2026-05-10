package routes

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"geo-project/internal/layers/handler"
	repositories "geo-project/internal/layers/repository"
	seeder "geo-project/internal/layers/seed"
	"geo-project/internal/layers/service"
	"geo-project/pkg/midleware"

	"github.com/doug-martin/goqu/v9"
	"github.com/labstack/echo/v5"
)

func RegisterLayersRoutes(api *echo.Group, goquDB *goqu.Database, jwtSecret string) {

	layerRepo := repositories.NewLayerRepository(goquDB)
	layerService := service.NewLayerService(layerRepo)
	layerHandler := handler.NewLayerHandler(layerService)

	api.GET("/layers", layerHandler.HandleGetLayers)
	api.GET("/layers/:id", layerHandler.HandleGetLayerByID)

	api.POST("/layers", layerHandler.HandleCreateLayer, midleware.JWTAuthMiddleware(jwtSecret))
	api.PATCH("/layers/:id", layerHandler.HandleUpdateLayer, midleware.JWTAuthMiddleware(jwtSecret))
	api.DELETE("/layers/:id", layerHandler.HandleDeleteLayer, midleware.JWTAuthMiddleware(jwtSecret))

	if os.Getenv("SEED_DB") == "true" {
		dbSeeder := seeder.NewSeeder(layerService, goquDB)

		seedCount := 200
		if v := os.Getenv("SEED_DB_COUNT"); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				seedCount = n
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := dbSeeder.SeedDomainData(ctx, seedCount); err != nil {
			log.Printf("db seeding failed: %v", err)
		}
	}
}
