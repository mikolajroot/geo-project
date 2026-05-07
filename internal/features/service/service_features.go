package service

import (
	"context"
	"geo-project/internal/features/model"
	"geo-project/internal/features/repository"
)

type FeatureService interface {
	CreateFeature(ctx context.Context, ownerExternalID int32, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error)
}

type featureService struct {
	repo repository.FeatureRepository
}

func NewFeatureService(repo repository.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

func (s *featureService) CreateFeature(ctx context.Context, ownerExternalID int32, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error) {
	owner, err := s.repo.GetOwnerByExternalID(ownerExternalID)
	if err != nil {
		return nil, err
	}

	feature := &model.Feature{
		LayerID:    layerID,
		OwnerID:    owner.ID,
		Name:       name,
		Type:       featureType,
		Geometry:   geometry,
		Properties: properties,
	}

	err = s.repo.CreateFeatureWithOwner(owner, feature)
	if err != nil {
		return nil, err
	}

	return feature, nil
}
