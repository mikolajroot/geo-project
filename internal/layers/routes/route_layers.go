package routes

import (
	"context"
	"geo-project/internal/layers/handler"
	repositories "geo-project/internal/layers/repository"
	seeder "geo-project/internal/layers/seed"
	"geo-project/internal/layers/service"
	"os"
	"time"
	"log"

	"github.com/doug-martin/goqu/v9"
	"github.com/labstack/echo/v5"
)

func RegisterLayersRoutes(api *echo.Group, goquDB *goqu.Database) {

	layerRepo := repositories.NewLayerRepository(goquDB)
	layerService := service.NewLayerService(layerRepo)
	layerHandler := handler.NewLayerHandler(layerService)

	api.GET("/layers", layerHandler.HandleGetLayers)
	api.GET("/layers/:id", layerHandler.HandleGetLayerByID)
	api.POST("/layers", layerHandler.HandleCreateLayer)
	api.PATCH("/layers/:id", layerHandler.HandleUpdateLayer)
	api.DELETE("/layers/:id", layerHandler.HandleDeleteLayer)

	if os.Getenv("SEED_DB") == "true" {
        dbSeeder := seeder.NewSeeder(layerService)
        
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) 
        defer cancel()
        
        dbSeeder.SeedDomainData(ctx, 200)
		os.Exit(0)
		log.Println("Exiting backend after seeding")
    }


}
