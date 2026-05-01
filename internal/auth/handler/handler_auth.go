package handler

import (
	"net/http"

	"geo-project/internal/auth/service"
	apperrors "geo-project/pkg/errors"

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

