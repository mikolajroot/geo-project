package service

import (
	"context"
	"errors"
	"fmt"
	models "geo-project/internal/layers/model"
	repositories "geo-project/internal/layers/repository"
	apperrors "geo-project/pkg/errors"
	"math"
)

type CreateLayerParams struct {
	Name         string
	Description  string
	GeometryType string
	SRID         int32
	OwnerID      *int32
}

type LayerService interface {
	GetLayers(ctx context.Context, geometryType *string, status *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error)
	GetLayerByID(ctx context.Context, id int32) (models.Layer, error)
	CreateLayer(ctx context.Context, params CreateLayerParams) (models.Layer, error)
}

type layerService struct {
	repo repositories.LayerRepository
}

func NewLayerService(repo repositories.LayerRepository) LayerService {
	return &layerService{
		repo: repo,
	}
}

func (s *layerService) GetLayers(ctx context.Context, geometryType *string, status *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error) {

	if page <= 0 {
		page = 1
	}

	if pageSize <= 0 {
		pageSize = 50
	} else if pageSize > 100 {
		pageSize = 100
	}

	layers, total, err := s.repo.GetLayers(ctx, status, geometryType, sortBy, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to fetch layers: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return layers, totalPages, nil
}

func (s *layerService) GetLayerByID(ctx context.Context, id int32) (models.Layer, error) {
	layer, err := s.repo.GetLayerByID(ctx, id)
	if err != nil {
		if errors.Is(err, repositories.ErrLayerNotFound) {
			return models.Layer{}, apperrors.NewAppError("NOT_FOUND", "Layer with this id doesnt exist")
		}

		return models.Layer{}, fmt.Errorf("failed to fetch layer: %w", err)
	}

	return layer, nil
}

func (s *layerService) CreateLayer(ctx context.Context, params CreateLayerParams) (models.Layer, error) {
	resolvedOwnerID := int32(1)
	if params.OwnerID != nil && *params.OwnerID > 0 {
		resolvedOwnerID = *params.OwnerID
	}

	layer := models.Layer{
		Name:         params.Name,
		Description:  params.Description,
		GeometryType: params.GeometryType,
		Status:       "active",
		SRID:         params.SRID,
		OwnerID:      resolvedOwnerID,
	}

	created, err := s.repo.CreateLayer(ctx, layer)
	if err != nil {
		if errors.Is(err, repositories.ErrDuplicateLayerName) {
			return models.Layer{}, apperrors.NewAppError("CONFLICT", "layer with this name already exists")
		}

		return models.Layer{}, fmt.Errorf("failed to create layer: %w", err)
	}

	return created, nil
}
