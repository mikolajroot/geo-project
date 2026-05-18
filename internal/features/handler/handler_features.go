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
		ID:       feature.ID,
		LayerID:  feature.LayerID,
		OwnerID:  feature.OwnerID,
		Name:     feature.Name,
		Type:     feature.Type,
		Geometry: json.RawMessage(feature.Geometry),
		Properties: func() json.RawMessage {
			if feature.Properties == "" {
				return json.RawMessage("{}")
			}
			return json.RawMessage(feature.Properties)
		}(),
		Owner:     &OwnerResponse{ID: feature.Owner.ID, Login: feature.Owner.Login},
		CreatedAt: feature.CreatedAt.String(),
		UpdatedAt: feature.UpdatedAt.String(),
	}

	return c.JSON(http.StatusCreated, response)
}

func (h *featureHandler) HandleUpdateFeature(c *echo.Context) error {
	var req UpdateFeatureRequest
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

	var geometryJSON *string
	if req.Geometry != nil {
		geoBytes, _ := json.Marshal(req.Geometry)
		geoStr := string(geoBytes)
		geometryJSON = &geoStr
	}

	var propertiesJSON *string
	if req.Properties != nil {
		propBytes, _ := json.Marshal(req.Properties)
		propStr := string(propBytes)
		propertiesJSON = &propStr
	}

	updated, err := h.service.UpdateFeature(
		c.Request().Context(),
		req.ID,
		cl.UserID,
		geometryJSON,
		propertiesJSON,
		c.Request().Header.Get("Authorization"),
	)
	if err != nil {
		return err
	}

	response := FeatureResponse{
		ID:       updated.ID,
		LayerID:  updated.LayerID,
		OwnerID:  updated.OwnerID,
		Name:     updated.Name,
		Type:     updated.Type,
		Geometry: json.RawMessage(updated.Geometry),
		Properties: func() json.RawMessage {
			if updated.Properties == "" {
				return json.RawMessage("{}")
			}
			return json.RawMessage(updated.Properties)
		}(),
		Owner:     &OwnerResponse{ID: updated.Owner.ID, Login: updated.Owner.Login},
		CreatedAt: updated.CreatedAt.String(),
		UpdatedAt: updated.UpdatedAt.String(),
	}

	return c.JSON(http.StatusOK, response)
}

func (h *featureHandler) HandleDeleteFeature(c *echo.Context) error {
	var req DeleteFeatureRequest
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

	authToken := c.Request().Header.Get("Authorization")

	if err := h.service.DeleteFeature(c.Request().Context(), req.ID, cl.UserID, authToken); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *featureHandler) HandleGetFeature(c *echo.Context) error {
	var req GetFeatureRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	feature, err := h.service.GetFeature(c.Request().Context(), req.ID)
	if err != nil {
		return err
	}

	response := FeatureResponse{
		ID:       feature.ID,
		LayerID:  feature.LayerID,
		OwnerID:  feature.OwnerID,
		Name:     feature.Name,
		Type:     feature.Type,
		Geometry: json.RawMessage(feature.Geometry),
		Properties: func() json.RawMessage {
			if feature.Properties == "" {
				return json.RawMessage("{}")
			}
			return json.RawMessage(feature.Properties)
		}(),
		Owner:     &OwnerResponse{ID: feature.Owner.ID, Login: feature.Owner.Login},
		CreatedAt: feature.CreatedAt.String(),
		UpdatedAt: feature.UpdatedAt.String(),
	}

	return c.JSON(http.StatusOK, response)
}

func (h *featureHandler) HandleGetFeaturesByLayer(c *echo.Context) error {
	var req GetFeaturesByLayerRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	features, totalPages, err := h.service.GetFeaturesByLayer(
		c.Request().Context(),
		req.LayerID,
		req.Type,
		req.SortBy,
		req.Page,
		req.PageSize,
		req.BBox,
	)
	if err != nil {
		return err
	}

	if req.Page <= 0 {
		req.Page = 1
	}

	response := okResponse[FeatureResponse]{
		Data: make([]FeatureResponse, 0, len(features)),
		Meta: PaginationResponse{TotalPages: totalPages, Page: req.Page},
	}

	for _, feature := range features {
		response.Data = append(response.Data, FeatureResponse{
			ID:       feature.ID,
			LayerID:  feature.LayerID,
			OwnerID:  feature.OwnerID,
			Name:     feature.Name,
			Type:     feature.Type,
			Geometry: json.RawMessage(feature.Geometry),
			Properties: func() json.RawMessage {
				if feature.Properties == "" {
					return json.RawMessage("{}")
				}
				return json.RawMessage(feature.Properties)
			}(),
			Owner:     &OwnerResponse{ID: feature.Owner.ID, Login: feature.Owner.Login},
			CreatedAt: feature.CreatedAt.String(),
			UpdatedAt: feature.UpdatedAt.String(),
		})
	}

	return c.JSON(http.StatusOK, response)
}
