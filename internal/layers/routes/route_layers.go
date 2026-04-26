package routes

import (
	"geo-project/controllers"
	"geo-project/service"

	"github.com/doug-martin/goqu/v9"
	"github.com/labstack/echo/v5"
)

func RegisterLayersRoutes(api *echo.Group, queries *goqu.Database) {

	heroService := services.NewHeroesService(queries)
	heroController := controllers.NewHeroesController(heroService)

	api.GET("/heroes", heroController.GetHeroesByStatusAndPower)

}
