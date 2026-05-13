package handler

import (
	"net/http"

	"geo-project/internal/annotations/service"
	apperrors "geo-project/pkg/errors"
	"geo-project/pkg/midleware"

	"github.com/labstack/echo/v5"
)

type annotationsHandler struct {
	service service.AnnotationsService
}

func NewAnnotationsHandler(service service.AnnotationsService) *annotationsHandler {
	return &annotationsHandler{service: service}
}

func (h *annotationsHandler) CreateAnnotation(c *echo.Context) error {
	var req CreateAnnotationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	cl, ok := midleware.GetClaims(c)
	if !ok || cl.UserID <= 0 {
		return apperrors.NewAppError("UNAUTHORIZED", "missing or invalid user claims")
	}

	annotation, err := h.service.CreateAnnotation(c.Request().Context(), service.CreateAnnotationParams{
		AuthorID: cl.UserID,
		LayerID:  req.LayerID,
		Text:     req.Text,
		Lng:      *req.Lng,
		Lat:      *req.Lat,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, CreateAnnotationResponse{
		ID:        annotation.ID.Hex(),
		AuthorID:  annotation.AuthorID,
		LayerID:   annotation.LayerID,
		Text:      annotation.Text,
		Location:  annotation.Location,
		CreatedAt: annotation.CreatedAt,
	})
}
