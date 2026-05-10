package service

import (
	"context"

	"geo-project/internal/analytics/model"
	"geo-project/internal/analytics/repository"
)

type AnalyticsService interface {
	Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error)
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
