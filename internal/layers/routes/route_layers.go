package routes

import (
	"geo-project/internal/layers/handler"
	"geo-project/internal/layers/service"
	"geo-project/internal/layers/repository"

	"github.com/doug-martin/goqu/v9"
	"github.com/labstack/echo/v5"
)

func RegisterLayersRoutes(api *echo.Group, goquDB *goqu.Database) {

	layerRepo := repositories.NewLayerRepository(goquDB)
	layerService := service.NewLayerService(layerRepo)
	layerHandler := handler.NewLayerHandler(layerService)

	api.GET("/layers", layerHandler.HandleGetLayers)

}
