package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"geo-project/internal/revisions/model"
	"geo-project/internal/revisions/repository"

	apperrors "geo-project/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RevisionService interface {
	CreateRevision(ctx context.Context, revision *model.Revision) (*model.Revision, error)
	ListByFeatureID(ctx context.Context, featureID int32) ([]*model.Revision, error)
	AddComment(ctx context.Context, revisionID string, authorID int32, text string) error
	GetByID(ctx context.Context, revisionID string) (*model.Revision, string, error)
}

type revisionService struct {
	repo repository.RevisionRepository
}

func NewRevisionService(repo repository.RevisionRepository) RevisionService {
	return &revisionService{repo: repo}
}

func (s *revisionService) CreateRevision(ctx context.Context, revision *model.Revision) (*model.Revision, error) {
	if err := verifyFeatureExists(revision.FeatureID); err != nil {
		return &model.Revision{}, err
	}
	revision.Prepare()
	return s.repo.CreateRevision(ctx, revision)
}

func (s *revisionService) ListByFeatureID(ctx context.Context, featureID int32) ([]*model.Revision, error) {
	if featureID <= 0 {
		return nil, apperrors.NewAppError("BAD_REQUEST", "feature_id must be greater than 0")
	}
	return s.repo.ListByFeatureID(ctx, featureID)
}

func (s *revisionService) AddComment(ctx context.Context, revisionID string, authorID int32, text string) error {
	if len(text) < 5 || len(text) > 500 {
		return apperrors.NewAppError("BAD_REQUEST", "comment text must be between 5 and 500 characters")
	}

	oid, err := bson.ObjectIDFromHex(revisionID)
	if err != nil {
		return apperrors.NewAppError("BAD_REQUEST", "invalid revision id")
	}

	comment := &model.Comment{
		AuthorID: authorID,
		Text:     text,
	}
	comment.Prepare()

	return s.repo.AddComment(ctx, oid, comment)
}

func (s *revisionService) GetByID(ctx context.Context, revisionID string) (*model.Revision,string, error) {
	oid, err := bson.ObjectIDFromHex(revisionID)
	if err != nil {
		return nil, "", apperrors.NewAppError("BAD_REQUEST", "invalid revision id")
	}

	revision, err := s.repo.GetByID(ctx, oid)
	if err != nil {
		return nil, "", err
	}

	featureName, _ := fetchFeatureDetails(revision.FeatureID)


	return revision, featureName, nil
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

type ExternalFeature struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
}

func fetchFeatureDetails(featureID int32) (string, error) {
	url := fmt.Sprintf("http://svc-features:8082/api/v1/features/%d", featureID)

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return "", apperrors.NewAppError("INTERNAL_ERROR", "failed to communicate with features service")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", apperrors.NewAppError("NOT_FOUND", fmt.Sprintf("feature with ID %d not found", featureID))
	}

	var feature ExternalFeature
	if err := json.NewDecoder(resp.Body).Decode(&feature); err != nil {
		return "", apperrors.NewAppError("BAD_REQUEST", "failed to decode feature data")
	}

	return feature.Name, nil
}
