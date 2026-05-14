package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"geo-project/internal/annotations/model"
	"geo-project/internal/annotations/repository"

	apperrors "geo-project/pkg/errors"
)

type CreateAnnotationParams struct {
	AuthorID  int32
	FeatureID int32
	Text      string
	Lng       float64
	Lat       float64
}

type AnnotationsService interface {
	CreateAnnotation(ctx context.Context, params CreateAnnotationParams) (model.Annotation, error)
	ListByFeatureIDs(ctx context.Context, featureIDs []int32) ([]model.Annotation, error)
	Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error)
	PatchAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error)
	DeleteAnnotation(ctx context.Context, id string, authorID int32) error
}

type annotationsService struct {
	repo repository.AnnotationsRepository
}

func NewAnnotationsService(repo repository.AnnotationsRepository) AnnotationsService {
	return &annotationsService{repo: repo}
}

func (s *annotationsService) CreateAnnotation(ctx context.Context, params CreateAnnotationParams) (model.Annotation, error) {

	if err := verifyFeatureExists(params.FeatureID); err != nil {
		return model.Annotation{}, err
	}

	annotation := model.Annotation{
		AuthorID:  params.AuthorID,
		FeatureID: params.FeatureID,
		Text:      params.Text,
		Location:  model.MongoGeoJSON{Type: "Point", Coordinates: []float64{params.Lng, params.Lat}},
		CreatedAt: time.Now().UTC(),
	}

	return s.repo.CreateAnnotation(ctx, annotation)
}

func (s *annotationsService) ListByFeatureIDs(ctx context.Context, featureIDs []int32) ([]model.Annotation, error) {
	return s.repo.ListByFeatureIDs(ctx, featureIDs)
}

func (s *annotationsService) Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error) {
	return s.repo.Nearby(ctx, lat, lng, maxDistance)
}

func (s *annotationsService) PatchAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error) {
	return s.repo.UpdateAnnotationText(ctx, id, authorID, text)
}

func (s *annotationsService) DeleteAnnotation(ctx context.Context, id string, authorID int32) error {
	return s.repo.DeleteAnnotation(ctx, id, authorID)
}

func verifyFeatureExists(featureID int32) error {

	url := fmt.Sprintf("http://svc-features:8082/api/v1/features/%d", featureID)

	resp, err := http.Get(url)
	if err != nil {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to communicate with features service")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apperrors.NewAppError("NOT_FOUND", fmt.Sprintf("feature with ID %d does not exist", featureID))
	}

	if resp.StatusCode != http.StatusOK {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to verify feature existence")
	}

	return nil
}
