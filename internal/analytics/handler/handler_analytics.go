package handler

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"geo-project/internal/analytics/service"
)

type analyticsHandler struct {
	svc service.AnalyticsService
}

func NewAnalyticsHandler(svc service.AnalyticsService) *analyticsHandler {
	return &analyticsHandler{svc: svc}
}

func (h *analyticsHandler) HandleNearby(c *echo.Context) error {
	var req NearbyRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	results, err := h.svc.Nearby(c.Request().Context(), req.LayerID, req.Lat, req.Lng, req.RadiusMeters)
	if err != nil {
		return err
	}

	out := make([]NearbyFeatureResponse, 0, len(results))
	for _, r := range results {
		out = append(out, NearbyFeatureResponse{
			ID: r.ID, LayerID: r.LayerID, OwnerID: r.OwnerID, Name: r.Name, Type: r.Type,
			Geometry: []byte(r.Geometry), Properties: []byte(r.Properties), DistanceMeters: r.DistanceMeters,
			CreatedAt: r.CreatedAt.String(), UpdatedAt: r.UpdatedAt.String(),
		})
	}

	return c.JSON(http.StatusOK, NearbyResponse{Data: out})
}

func (h *analyticsHandler) HandleIntersect(c *echo.Context) error {
	var req IntersectRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	results, err := h.svc.Intersect(c.Request().Context(), req.LayerID, string(req.Geometry))
	if err != nil {
		return err
	}

	out := make([]NearbyFeatureResponse, 0, len(results))
	for _, r := range results {
		out = append(out, NearbyFeatureResponse{
			ID: r.ID, LayerID: r.LayerID, OwnerID: r.OwnerID, Name: r.Name, Type: r.Type,
			Geometry: []byte(r.Geometry), Properties: []byte(r.Properties), DistanceMeters: r.DistanceMeters,
			CreatedAt: r.CreatedAt.String(), UpdatedAt: r.UpdatedAt.String(),
		})
	}

	return c.JSON(http.StatusOK, NearbyResponse{Data: out})
}

func (h *analyticsHandler) HandleLayerStats(c *echo.Context) error {
	var req LayerStatsRequest
	if err := c.Bind(&req); err != nil {
		return err
	}
	if err := c.Validate(&req); err != nil {
		return err
	}

	stats, err := h.svc.LayerStats(c.Request().Context(), req.LayerID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, LayerStatsResponse{
		LayerID:            stats.LayerID,
		TotalFeatures:      stats.TotalFeatures,
		TotalAreaSqMeters:  stats.TotalAreaSqMeters,
		TotalLengthMeters:  stats.TotalLengthMeters,
		LayerExtent:        stats.LayerExtent,
		LastUpdatedFeature: stats.LastUpdatedFeature,
		Type:               stats.LastUpdatedFeatureType,
		FeatureTypesCount:  stats.FeatureTypesCount,
	})
}
