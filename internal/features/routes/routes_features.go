package routes

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"geo-project/internal/features/handler"
	repositories "geo-project/internal/features/repository"
	seeder "geo-project/internal/features/seed"
	"geo-project/internal/features/service"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"
)

func RegisterFeaturesRoutes(api *echo.Group, gormDB *gorm.DB, jwtSecret string) {
	featureRepo := repositories.NewFeatureRepository(gormDB)
	featureService := service.NewFeatureService(featureRepo)
	featureHandler := handler.NewFeatureHandler(featureService)

	api.POST("/features", featureHandler.HandleCreateFeature, midleware.JWTAuthMiddleware(jwtSecret))
	api.GET("/features/layer/:layer_id", featureHandler.HandleGetFeaturesByLayer)
	api.GET("/features/:id", featureHandler.HandleGetFeature)
	api.PUT("/features/:id", featureHandler.HandleUpdateFeature, midleware.JWTAuthMiddleware(jwtSecret))
	api.DELETE("/features/:id", featureHandler.HandleDeleteFeature, midleware.JWTAuthMiddleware(jwtSecret))

	if os.Getenv("SEED_DB") == "true" {
		dbSeeder := seeder.NewSeeder(featureService, gormDB)

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
