package handler

import (
	"net/http"

	"geo-project/internal/auth/service"

	"github.com/labstack/echo/v5"
)

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *authHandler {
	return &authHandler{
		service: service,
	}
}


func (h *authHandler) HandleRegister(c *echo.Context) error {
	var req LoginAndRegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}


	ctx := c.Request().Context()
	jtwToken, err := h.service.RegisterService(ctx, req.Login,req.Password)
	if err != nil {
		return err
	}

	response := LoginAndRegisterUserResponse{
		JwtToken: jtwToken,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *authHandler) HandleLogin(c *echo.Context) error {
	var req LoginAndRegisterUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	ctx := c.Request().Context()
	jwtToken, err := h.service.LoginService(ctx, req.Login,req.Password)
	if err != nil {
		return err
	}

	response := LoginAndRegisterUserResponse{
		JwtToken: jwtToken,
	}

	return c.JSON(http.StatusOK, response)
}