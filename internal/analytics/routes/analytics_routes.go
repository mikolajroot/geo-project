package routes

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v5"

	"geo-project/internal/analytics/handler"
	"geo-project/internal/analytics/repository"
	"geo-project/internal/analytics/service"
)

func RegisterAnalyticsRoutes(api *echo.Group, pgx *pgxpool.Pool, jwtSecret string) {
	repo := repository.NewAnalyticsRepository(pgx)
	svc := service.NewAnalyticsService(repo)
	h := handler.NewAnalyticsHandler(svc)

	api.GET("/analytics/nearby", h.HandleNearby)
	api.POST("/analytics/intersect", h.HandleIntersect)
}
