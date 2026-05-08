package service

import (
	"context"
	"geo-project/internal/features/model"
	"geo-project/internal/features/repository"
	apperrors "geo-project/pkg/errors"
)

type FeatureService interface {
	CreateFeature(ctx context.Context, ownerExternalID int32, ownerLogin string, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error)
	UpdateFeature(ctx context.Context, featureID int32, ownerExternalID int32, geometry *string, properties *string) (*model.Feature, error)
	DeleteFeature(ctx context.Context, featureID int32, ownerExternalID int32) error
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

func (s *featureService) UpdateFeature(ctx context.Context, featureID int32, ownerExternalID int32, geometry *string, properties *string) (*model.Feature, error) {
	return s.repo.UpdateFeatureByIDAndOwner(featureID, ownerExternalID, geometry, properties)
}

func (s *featureService) DeleteFeature(ctx context.Context, featureID int32, ownerExternalID int32) error {
	feature, err := s.repo.GetFeatureByIDWithOwner(featureID)
	if err != nil {
		return err
	}

	if feature.Owner.ExternalID != ownerExternalID {
		return apperrors.NewAppError("FORBIDDEN", "you are not the owner of this feature")
	}

	return s.repo.DeleteFeatureByID(featureID)
}
