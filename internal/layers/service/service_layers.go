package service

import (
	"context"
	"fmt"
	models "geo-project/internal/layers/model"
	repositories "geo-project/internal/layers/repository"
	"math"
)

type LayerService interface {
	GetLayers(ctx context.Context, geometryType *string, status *string, sortBy *string,page int, pageSize int) (result []models.Layer,totalPages int,pageResult int, err error)
}

type layerService struct {
	repo repositories.LayerRepository
}


func NewLayerService(repo repositories.LayerRepository) LayerService {
	return &layerService{
		repo: repo,
	}
}

func (s *layerService) GetLayers(ctx context.Context, geometryType *string, status *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, int, error) {
	
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
		return nil, 0, 0, fmt.Errorf("failed to fetch layers: %w", err)
	}

	totalPages := int(math.Ceil(float64(total) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}
	
	return layers, totalPages, page, nil
}