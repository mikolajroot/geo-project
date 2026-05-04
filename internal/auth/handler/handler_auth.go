package handler

import (
	"net/http"
	"time"

	"geo-project/internal/auth/service"

	"github.com/labstack/echo/v5"
)

const refreshTokenCookieName = "refresh_token"
const refreshTokenCookieMaxAge = 30 * 24 * 60 * 60

type authHandler struct {
	service service.AuthService
}

func NewAuthHandler(service service.AuthService) *authHandler {
	return &authHandler{
		service: service,
	}
}

func setRefreshTokenCookie(c *echo.Context, refreshToken string) {
	cookie := &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    refreshToken,
		Path:     "/api/v1/auth",
		MaxAge:   refreshTokenCookieMaxAge,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(30 * 24 * time.Hour),
	}

	c.SetCookie(cookie)
}

func getRefreshTokenFromRequest(c *echo.Context) string {
	if cookie, err := c.Cookie(refreshTokenCookieName); err == nil && cookie != nil && cookie.Value != "" {
		return cookie.Value
	}

	return ""
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
	tokens, err := h.service.RegisterService(ctx, req.Login, req.Password, ipAddress, device)
	if err != nil {
		return err
	}

	response := LoginAndRegisterUserResponse{
		AccessToken: tokens.AccessToken,
	}
	setRefreshTokenCookie(c, tokens.RefreshToken)

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
	tokens, err := h.service.LoginService(ctx, req.Login, req.Password, ipAddress, device)
	if err != nil {
		return err
	}

	response := LoginAndRegisterUserResponse{
		AccessToken: tokens.AccessToken,
	}
	setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(http.StatusOK, response)
}

func (h *authHandler) HandleRefreshToken(c *echo.Context) error {
	refreshToken := getRefreshTokenFromRequest(c)
	if refreshToken == "" {
		return echo.NewHTTPError(http.StatusUnauthorized, "missing refresh token cookie")
	}

	ctx := c.Request().Context()
	tokens, err := h.service.RefreshService(ctx, refreshToken)
	if err != nil {
		return err
	}
	setRefreshTokenCookie(c, tokens.RefreshToken)

	return c.JSON(http.StatusOK, LoginAndRegisterUserResponse{
		AccessToken: tokens.AccessToken,
	})
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
