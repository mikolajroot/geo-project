package routes

import (
	"geo-project/internal/annotations/handler"
	"geo-project/internal/annotations/repository"
	"geo-project/internal/annotations/service"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterAnnotationsRoutes(api *echo.Group, jwtSecret string, db *mongo.Database) {

	repo := repository.NewAnnotationsRepository(db)
	svc := service.NewAnnotationsService(repo)
	h := handler.NewAnnotationsHandler(svc)

	annotations := api.Group("/annotations")
	annotations.Use(midleware.JWTAuthMiddleware(jwtSecret))

	annotations.POST("", h.CreateAnnotation)
	annotations.PATCH("/:id", h.PatchAnnotation)
	annotations.GET("/nearby", h.HandleNearby)
}
