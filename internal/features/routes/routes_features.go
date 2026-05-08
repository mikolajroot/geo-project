package routes

import (
	"geo-project/internal/features/handler"
	repositories "geo-project/internal/features/repository"
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
	api.PUT("/features/:id", featureHandler.HandleUpdateFeature, midleware.JWTAuthMiddleware(jwtSecret))
	api.DELETE("/features/:id", featureHandler.HandleDeleteFeature, midleware.JWTAuthMiddleware(jwtSecret))
}
