package routes

import (
	"geo-project/internal/auth/ent"
	"geo-project/internal/auth/handler"
	repositories "geo-project/internal/auth/repository"
	"geo-project/internal/auth/service"

	"github.com/labstack/echo/v5"
)

func RegisterAuthRoutes(api *echo.Group, client *ent.Client,jwtSecret string) {

	userRepo := repositories.NewUserRepository(client)
	authService := service.NewAuthService(userRepo,jwtSecret)
	authHandler := handler.NewAuthHandler(authService)

	api.POST("/register", authHandler.HandleRegister)
	// api.POST("/login", authHandler.HandleLogin)
	// api.POST("/refresh", authHandler.HandleRefreshToken)
	

}
