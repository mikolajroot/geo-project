package service

import (
	"context"
	"geo-project/internal/features/model"
	"geo-project/internal/features/repository"
)

type FeatureService interface {
	CreateFeature(ctx context.Context, ownerExternalID int32, ownerLogin string, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error)
}

type featureService struct {
	repo repository.FeatureRepository
}

func NewFeatureService(repo repository.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

func (s *featureService) CreateFeature(ctx context.Context, ownerExternalID int32, ownerLogin string, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error) {
	owner := &model.Owner{
		ExternalID: ownerExternalID,
		Login:      ownerLogin,
	}

	feature := &model.Feature{
		LayerID:    layerID,
		Name:       name,
		Type:       featureType,
		Geometry:   geometry,
		Properties: properties,
	}

	if err := s.repo.CreateFeatureWithOwner(owner, feature); err != nil {
		return nil, err
	}

	return feature, nil
}
