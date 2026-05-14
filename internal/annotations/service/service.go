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
	AuthorID int32
	LayerID  int32
	Text     string
	Lng      float64
	Lat      float64
}

type AnnotationsService interface {
	CreateAnnotation(ctx context.Context, params CreateAnnotationParams) (model.Annotation, error)
	ListByLayerIDs(ctx context.Context, layerIDs []int32) ([]model.Annotation, error)
	Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error)
	PatchAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error)
}

type annotationsService struct {
	repo repository.AnnotationsRepository
}

func NewAnnotationsService(repo repository.AnnotationsRepository) AnnotationsService {
	return &annotationsService{repo: repo}
}

func (s *annotationsService) CreateAnnotation(ctx context.Context, params CreateAnnotationParams) (model.Annotation, error) {

	if err := verifyLayerExists(params.LayerID); err != nil {
		return model.Annotation{}, err
	}

	annotation := model.Annotation{
		AuthorID:  params.AuthorID,
		LayerID:   params.LayerID,
		Text:      params.Text,
		Location:  model.MongoGeoJSON{Type: "Point", Coordinates: []float64{params.Lng, params.Lat}},
		CreatedAt: time.Now().UTC(),
	}

	return s.repo.CreateAnnotation(ctx, annotation)
}

func (s *annotationsService) ListByLayerIDs(ctx context.Context, layerIDs []int32) ([]model.Annotation, error) {
	return s.repo.ListByLayerIDs(ctx, layerIDs)
}

func (s *annotationsService) Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error) {
	return s.repo.Nearby(ctx, lat, lng, maxDistance)
}

func (s *annotationsService) PatchAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error) {
	return s.repo.UpdateAnnotationText(ctx, id, authorID, text)
}

func verifyLayerExists(layerID int32) error {

	url := fmt.Sprintf("http://svc-layers:8080/api/v1/layers/%d", layerID)

	resp, err := http.Get(url)
	if err != nil {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to communicate with layers service")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return apperrors.NewAppError("NOT_FOUND", fmt.Sprintf("layer with ID %d does not exist", layerID))
	}

	if resp.StatusCode != http.StatusOK {
		return apperrors.NewAppError("INTERNAL_ERROR", "failed to verify layer existence")
	}

	return nil
}
