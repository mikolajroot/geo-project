package handler

import (
	"encoding/json"
	"net/http"

	"geo-project/internal/features/service"
	apperrors "geo-project/pkg/errors"
	"geo-project/pkg/midleware"

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

	cl, ok := midleware.GetClaims(c)
	if !ok || cl.UserID <= 0 {
		return apperrors.NewAppError("UNAUTHORIZED", "missing or invalid user claims")
	}
	userID := cl.UserID
	userLogin := cl.Login

	ctx := c.Request().Context()
	geoBytes, _ := json.Marshal(req.Geometry)
	propBytes, _ := json.Marshal(req.Properties)

	feature, err := h.service.CreateFeature(
		ctx,
		int32(userID),
		userLogin,
		req.Name,
		req.Type,
		string(geoBytes),
		string(propBytes),
		req.LayerID,
	)
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
