package service

import (
	"bytes"
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
	if err := verifyGeometryTypeWithLayerSvc(layerID, geometry); err != nil {
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
	feature, err := s.repo.UpdateFeatureByIDAndOwner(featureID, ownerExternalID, geometry, properties)
	if err != nil {
		return nil, err
	}

	changeLog := buildChangeLog(geometry, properties)
	if err := createRevisionInMongo(featureID, ownerExternalID, changeLog); err != nil {
		fmt.Printf("warning: failed to create revision for feature %d: %v\n", featureID, err)
	}

	return feature, nil
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

func verifyGeometryTypeWithLayerSvc(layerID int32, geometry string) error {
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

	var geom struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal([]byte(geometry), &geom); err != nil {
		return apperrors.NewAppError("BAD_REQUEST", "invalid geometry JSON")
	}

	provided := strings.ToUpper(geom.Type)
	expected := strings.ToUpper(layerData.GeometryType)

	if provided == "" || expected == "" {
		return apperrors.NewAppError("BAD_REQUEST", "geometry type or layer geometry type is empty")
	}

	if provided != expected {
		return apperrors.NewAppError(
			"BAD_REQUEST",
			fmt.Sprintf("geometry type mismatch: layer accepts only %s, but you provided %s", layerData.GeometryType, geom.Type),
		)
	}

	return nil
}

func buildChangeLog(geometry *string, properties *string) string {
	var changes []string
	if geometry != nil {
		changes = append(changes, "updated geometry")
	}
	if properties != nil {
		changes = append(changes, "updated properties")
	}
	return "Feature updated: " + strings.Join(changes, " and ")
}

func createRevisionInMongo(featureID int32, authorID int32, changeLog string) error {
	payload := map[string]any{
		"feature_id": featureID,
		"author_id":  authorID,
		"change_log": changeLog,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal revision payload: %w", err)
	}

	url := "http://svc-revisions:8085/api/v1/revisions"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create revision request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send revision request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("revision service returned status %d", resp.StatusCode)
	}

	return nil
}
