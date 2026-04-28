package repositories

import (
	"context"
	"fmt"
	models "geo-project/internal/layers/model"

	"github.com/doug-martin/goqu/v9"
)

type LayerRepository interface {
	GetLayers(ctx context.Context, status *string, geometryType *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error)
	CreateLayer(ctx context.Context, layer models.Layer) (models.Layer, error)
}

type layerRepository struct {
	db *goqu.Database
}

func NewLayerRepository(db *goqu.Database) LayerRepository {
	return &layerRepository{
		db: db,
	}
}

type LayerWithCount struct {
	models.Layer
	TotalCount int `db:"total_count"`
}

func (r *layerRepository) GetLayers(ctx context.Context, status *string, geometryType *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error) {

	query := r.db.From("layers")

	if status != nil {
		query = query.Where(goqu.Ex{"status": *status})
	}

	if geometryType != nil {
		query = query.Where(goqu.Ex{"geometry_type": *geometryType})
	}

	offset := (page - 1) * pageSize

	query = query.Select(
		"id",
		"name",
		"description",
		"geometry_type",
		"status",
		"owner_id",
		"created_at",
		"updated_at",
		goqu.L("COUNT(*) OVER()").As("total_count"),
	)
	if sortBy != nil {
		query.Order(goqu.I(*sortBy).Desc())
	} else {
		query.Order(goqu.I("updated_at").Desc())
	}

	query.Limit(uint(pageSize)).
		Offset(uint(offset))

	var rows []LayerWithCount
	if err := query.ScanStructsContext(ctx, &rows); err != nil {
		return nil, 0, fmt.Errorf("db error: %w", err)
	}

	total := 0
	if len(rows) > 0 {
		total = rows[0].TotalCount
	}

	layers := make([]models.Layer, len(rows))
	for i, row := range rows {
		layers[i] = models.Layer{
			ID:           row.ID,
			Name:         row.Name,
			Status:       row.Status,
			GeometryType: row.GeometryType,
			Description:  row.Description,
			OwnerID:      row.OwnerID,
			CreatedAt:    row.CreatedAt,
			UpdatedAt:    row.UpdatedAt,
		}
	}

	return layers, total, nil

}

func (r *layerRepository) CreateLayer(ctx context.Context, layer models.Layer) (models.Layer, error) {
	query := r.db.Insert("layers").Rows(goqu.Record{
		"name":          layer.Name,
		"description":   layer.Description,
		"geometry_type": layer.GeometryType,
		"status":        layer.Status,
		"srid":          layer.SRID,
		"owner_id":      layer.OwnerID,
	}).Returning(
		"id",
		"name",
		"description",
		"geometry_type",
		"status",
		"srid",
		"owner_id",
		"created_at",
		"updated_at",
	)

	var created models.Layer
	if _, err := query.Executor().ScanStructContext(ctx, &created); err != nil {
		return models.Layer{}, fmt.Errorf("db error: %w", err)
	}

	return created, nil
}
