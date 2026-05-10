package service

import (
	"context"
	"encoding/json"
	"fmt"
	"geo-project/internal/features/model"
	"geo-project/internal/features/repository"
	apperrors "geo-project/pkg/errors"
	"math"
	"net/http"
	"strings"
)

type FeatureService interface {
	CreateFeature(ctx context.Context, ownerExternalID int32, ownerLogin string, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error)
	UpdateFeature(ctx context.Context, featureID int32, ownerExternalID int32, geometry *string, properties *string) (*model.Feature, error)
	DeleteFeature(ctx context.Context, featureID int32, ownerExternalID int32) error
	GetFeature(ctx context.Context, featureID int32) (*model.Feature, error)
	GetFeaturesByLayer(ctx context.Context, layerID int32, featureType string, sortBy string, page int, pageSize int, bbox string) ([]model.Feature, int, error)
}

type featureService struct {
	repo repository.FeatureRepository
}

func NewFeatureService(repo repository.FeatureRepository) FeatureService {
	return &featureService{repo: repo}
}

func (s *featureService) CreateFeature(ctx context.Context, ownerExternalID int32, ownerLogin string, name string, featureType string, geometry string, properties string, layerID int32) (*model.Feature, error) {
	if err := verifyGeometryTypeWithLayerSvc(layerID, featureType); err != nil {
		return nil, err
	}
	
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

func (s *featureService) GetFeature(ctx context.Context, featureID int32) (*model.Feature, error) {
	return s.repo.GetFeatureByIDWithOwner(featureID)
}

func (s *featureService) GetFeaturesByLayer(ctx context.Context, layerID int32, featureType string, sortBy string, page int, pageSize int, bbox string) ([]model.Feature, int, error) {
	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 50
	} else if pageSize > 100 {
		pageSize = 100
	}

	features, total, err := s.repo.GetFeaturesByLayer(layerID, featureType, sortBy, page, pageSize, bbox)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch features: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return features, totalPages, nil
}

func verifyGeometryTypeWithLayerSvc(layerID int32, requestedFeatureType string) error {

	url := fmt.Sprintf("http://svc-layers:8080/api/v1/layers/%d", layerID)
	
	resp, err := http.Get(url)
	if err != nil {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to connect to layers service")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apperrors.NewAppError("NOT_FOUND", "target layer does not exist")
	}
	if resp.StatusCode != http.StatusOK {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to fetch layer details")
	}

	var layerData struct {
		GeometryType string `json:"geometry_type"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&layerData); err != nil {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to decode layer data")
	}

	if !strings.EqualFold(layerData.GeometryType, requestedFeatureType) {
		return apperrors.NewAppError(
			"BAD_REQUEST", 
			fmt.Sprintf("geometry type mismatch: layer accepts only %s, but you provided %s", layerData.GeometryType, requestedFeatureType),
		)
	}

	return nil
}