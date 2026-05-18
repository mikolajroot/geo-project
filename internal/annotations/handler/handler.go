package handler

import (
	"net/http"
	"strconv"
	"strings"

	"geo-project/internal/annotations/model"
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
		AuthorID:  cl.UserID,
		FeatureID: req.FeatureID,
		Text:      req.Text,
		Lng:       *req.Lng,
		Lat:       *req.Lat,
	})
	if err != nil {
		return err
	}

	return c.JSON(http.StatusCreated, CreateAnnotationResponse{
		ID:        annotation.ID.Hex(),
		AuthorID:  annotation.AuthorID,
		FeatureID: annotation.FeatureID,
		Text:      annotation.Text,
		Location:  annotation.Location,
		CreatedAt: annotation.CreatedAt,
	})
}

func (h *annotationsHandler) DeleteAnnotation(c *echo.Context) error {
	var req DeleteAnnotationRequest
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

	if err := h.service.DeleteAnnotation(c.Request().Context(), req.ID, cl.UserID); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *annotationsHandler) ListAnnotations(c *echo.Context) error {
	var req ListAnnotationsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	featureIDs, err := parseFeatureIDs(req.Features)
	if err != nil {
		return apperrors.NewAppError("BAD_REQUEST", "invalid features query param")
	}

	annotations, err := h.service.ListByFeatureIDs(c.Request().Context(), featureIDs)
	if err != nil {
		return err
	}

	response := AnnotationsResponse{Data: make([]CreateAnnotationResponse, 0, len(annotations))}
	for _, annotation := range annotations {
		response.Data = append(response.Data, createAnnotationToResponse(annotation))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *annotationsHandler) HandleNearby(c *echo.Context) error {
	var req NearbyAnnotationsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	annotations, err := h.service.Nearby(c.Request().Context(), *req.Lat, *req.Lng, *req.MaxDistance)
	if err != nil {
		return err
	}

	response := NearbyAnnotationsResponse{Data: make([]NearbyAnnotationResponse, 0, len(annotations))}
	for _, annotation := range annotations {
		response.Data = append(response.Data, nearbyAnnotationToResponse(annotation))
	}

	return c.JSON(http.StatusOK, response)
}

func (h *annotationsHandler) HandleGetFeatureStats(c *echo.Context) error {
	var req FeatureStatsRequest;
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}
	stats, err := h.service.GetFeatureStats(c.Request().Context(), req.FeatureID)
	if err != nil {
		return err
	}
	response := make([]FeatureStatItem, 0, len(stats))
	for _, stat := range stats {
		response = append(response, FeatureStatItem{
			AuthorID:         stat.AuthorID,
			TotalAnnotations: stat.TotalAnnotations,
			LatestActivity:   stat.LatestActivity,
		})
	}
	return c.JSON(http.StatusOK, FeatureStatsResponse{
		FeatureID: req.FeatureID,
		Stats:     response,
	})
}

func (h *annotationsHandler) PatchAnnotation(c *echo.Context) error {
	var req PatchAnnotationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	idStr := strings.TrimSpace(req.ID)
	idStr = strings.Trim(idStr, "\"")

	cl, ok := midleware.GetClaims(c)
	if !ok || cl.UserID <= 0 {
		return apperrors.NewAppError("UNAUTHORIZED", "missing or invalid user claims")
	}

	annotation, err := h.service.PatchAnnotationText(c.Request().Context(), idStr, cl.UserID, req.Text)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, CreateAnnotationResponse{
		ID:        annotation.ID.Hex(),
		AuthorID:  annotation.AuthorID,
		FeatureID: annotation.FeatureID,
		Text:      annotation.Text,
		Location:  annotation.Location,
		CreatedAt: annotation.CreatedAt,
	})
}

func parseFeatureIDs(raw string) ([]int32, error) {
	parts := strings.Split(raw, ",")
	ids := make([]int32, 0, len(parts))

	for _, part := range parts {
		item := strings.TrimSpace(part)
		if item == "" {
			return nil, strconv.ErrSyntax
		}

		id, err := strconv.ParseInt(item, 10, 32)
		if err != nil || id <= 0 {
			return nil, strconv.ErrSyntax
		}

		ids = append(ids, int32(id))
	}

	if len(ids) == 0 {
		return nil, strconv.ErrSyntax
	}

	return ids, nil
}

func createAnnotationToResponse(annotation model.Annotation) CreateAnnotationResponse {
	return CreateAnnotationResponse{
		ID:        annotation.ID.Hex(),
		AuthorID:  annotation.AuthorID,
		FeatureID: annotation.FeatureID,
		Text:      annotation.Text,
		Location:  annotation.Location,
		CreatedAt: annotation.CreatedAt,
	}
}

func nearbyAnnotationToResponse(annotation model.NearbyAnnotation) NearbyAnnotationResponse {
	return NearbyAnnotationResponse{
		ID:             annotation.ID.Hex(),
		AuthorID:       annotation.AuthorID,
		FeatureID:      annotation.FeatureID,
		Text:           annotation.Text,
		Location:       annotation.Location,
		CreatedAt:      annotation.CreatedAt,
		DistanceMeters: annotation.DistanceMeters,
	}
}
