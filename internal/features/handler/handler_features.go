package handler

import (
	"net/http"
	"strconv"

	"geo-project/internal/features/service"
	apperrors "geo-project/pkg/errors"

	"github.com/labstack/echo/v5"
)

type featureHandler struct {
	service service.FeatureService
}

func NewFeatureHandler(service service.FeatureService) *featureHandler {
	return &featureHandler{service: service}
}

func (h *featureHandler) HandleCreateFeature(c *echo.Context) error {
	var req CreateFeatureRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	userIDStr := c.Request().Header.Get("X-User-ID")
	if userIDStr == "" {
		return apperrors.NewAppError("UNAUTHORIZED", "user id not found in token")
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 32)
	if err != nil || userID <= 0 {
		return apperrors.NewAppError("BAD_REQUEST", "invalid user id from token")
	}

	ctx := c.Request().Context()
	feature, err := h.service.CreateFeature(ctx, int32(userID), req.Name, req.Type, req.Geometry, req.Properties, req.LayerID)
	if err != nil {
		return err
	}

	response := FeatureResponse{
		ID:         feature.ID,
		LayerID:    feature.LayerID,
		OwnerID:    feature.OwnerID,
		Name:       feature.Name,
		Type:       feature.Type,
		Geometry:   feature.Geometry,
		Properties: feature.Properties,
		CreatedAt:  feature.CreatedAt.String(),
		UpdatedAt:  feature.UpdatedAt.String(),
	}

	return c.JSON(http.StatusCreated, response)
}
