package routes

import (
	"geo-project/internal/layers/handler"
	repositories "geo-project/internal/layers/repository"
	"geo-project/internal/layers/service"

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

}
