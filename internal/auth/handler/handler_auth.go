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
	ipAddress := c.RealIP()
	device := c.Request().Header.Get("User-Agent")
	jtwToken, err := h.service.RegisterService(ctx, req.Login, req.Password, ipAddress, device)
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
	ipAddress := c.RealIP()
	device := c.Request().Header.Get("User-Agent")
	jwtToken, err := h.service.LoginService(ctx, req.Login, req.Password, ipAddress, device)
	if err != nil {
		return err
	}

	response := LoginAndRegisterUserResponse{
		JwtToken: jwtToken,
	}

	return c.JSON(http.StatusOK, response)
}

func (h *authHandler) HandleGetStats(c *echo.Context) error {
	ctx := c.Request().Context()

	stats, err := h.service.GetSystemStatistics(ctx)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   stats,
	})
}