package routes

import (
	"geo-project/internal/revisions/handler"
	"geo-project/internal/revisions/repository"
	"geo-project/internal/revisions/service"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func RegisterRevisionsRoutes(api *echo.Group, jwtSecret string, db *mongo.Database) {
	repo := repository.NewRevisionRepository(db)
	svc := service.NewRevisionService(repo)
	h := handler.NewRevisionHandler(svc)

	revisions := api.Group("/revisions")
	revisions.Use(midleware.JWTAuthMiddleware(jwtSecret))

	revisions.POST("", h.CreateRevision)
	revisions.GET("", h.ListByFeatureID)
	revisions.GET("/:id", h.GetRevisionByID)
	revisions.POST("/:id/comments", h.AddComment)
}
