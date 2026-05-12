package service

import (
	"context"

	"geo-project/internal/analytics/model"
	"geo-project/internal/analytics/repository"
)

type AnalyticsService interface {
	Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error)
	Intersect(ctx context.Context, layerID int32, geometry string) ([]model.NearbyFeature, error)
	LayerStats(ctx context.Context, layerID int32) (model.LayerStats, error)
}

type analyticsService struct {
	repo repository.AnalyticsRepository
}

func NewAnalyticsService(repo repository.AnalyticsRepository) AnalyticsService {
	return &analyticsService{repo: repo}
}

func (s *analyticsService) Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error) {
	return s.repo.Nearby(ctx, layerID, lat, lng, radius)
}

func (s *analyticsService) Intersect(ctx context.Context, layerID int32, geometry string) ([]model.NearbyFeature, error) {
	return s.repo.Intersect(ctx, layerID, geometry)
}

func (s *analyticsService) LayerStats(ctx context.Context, layerID int32) (model.LayerStats, error) {
	totals, err := s.repo.LayerStatsTotals(ctx, layerID)
	if err != nil {
		return model.LayerStats{}, err
	}

	featureTypes, err := s.repo.LayerStatsFeatureTypes(ctx, layerID)
	if err != nil {
		return model.LayerStats{}, err
	}

	spatial, err := s.repo.LayerStatsSpatial(ctx, layerID)
	if err != nil {
		return model.LayerStats{}, err
	}

	totals.FeatureTypesCount = featureTypes
	totals.LayerExtent = spatial.LayerExtent
	totals.LastUpdatedFeature = spatial.LastUpdatedFeature
	totals.LastUpdatedFeatureType = spatial.LastUpdatedFeatureType
	return totals, nil
}
