package repositories

import (
	"context"
	"errors"
	"fmt"
	models "geo-project/internal/layers/model"
	"strings"

	"github.com/doug-martin/goqu/v9"
)

type LayerRepository interface {
	GetLayers(ctx context.Context, status *string, geometryType *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error)
	GetLayerByID(ctx context.Context, id int32) (models.Layer, error)
	EnsureOwner(ctx context.Context, externalID int32, login string) (int32, error)
	CreateLayer(ctx context.Context, layer models.Layer) (models.Layer, error)
	UpdateLayer(ctx context.Context, id int32, name *string, description *string, status *string) (models.Layer, error)
	ArchiveLayer(ctx context.Context, id int32) (models.Layer, error)
}

type layerRepository struct {
	db *goqu.Database
}

func NewLayerRepository(db *goqu.Database) LayerRepository {
	return &layerRepository{
		db: db,
	}
}

var ErrDuplicateLayerName = errors.New("layer with this name already exists")
var ErrLayerNotFound = errors.New("layer not found")

type ownerRow struct {
	ID int32 `db:"id"`
}

type LayerWithCount struct {
	models.Layer
	TotalCount int `db:"total_count"`
}

func (r *layerRepository) GetLayers(ctx context.Context, status *string, geometryType *string, sortBy *string, page int, pageSize int) ([]models.Layer, int, error) {

	query := r.db.From("layers")

	if status != nil {
		query = query.Where(goqu.Ex{"status": *status})
	} else {
		query = query.Where(goqu.Ex{"status": goqu.Op{"neq": "archived"}})
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
		query = query.Order(goqu.I(*sortBy).Asc())
	} else {
		query = query.Order(goqu.I("updated_at").Asc())
	}

	query = query.Limit(uint(pageSize)).Offset(uint(offset))

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

func (r *layerRepository) GetLayerByID(ctx context.Context, id int32) (models.Layer, error) {
	query := r.db.From("layers").
		Select(
			"id",
			"name",
			"description",
			"geometry_type",
			"status",
			"srid",
			"owner_id",
			"created_at",
			"updated_at",
		).
		Where(goqu.Ex{"id": id})

	var layer models.Layer
	found, err := query.Executor().ScanStructContext(ctx, &layer)
	if err != nil {
		return models.Layer{}, fmt.Errorf("db error: %w", err)
	}

	if !found {
		return models.Layer{}, ErrLayerNotFound
	}

	return layer, nil
}

func (r *layerRepository) EnsureOwner(ctx context.Context, externalID int32, login string) (int32, error) {
	query := r.db.Insert("owners").Rows(goqu.Record{
		"external_id": externalID,
		"login":       login,
	}).OnConflict(goqu.DoUpdate("external_id", goqu.Record{
		"login": login,
	})).Returning("id")

	var owner ownerRow
	found, err := query.Executor().ScanStructContext(ctx, &owner)
	if err != nil {
		return 0, fmt.Errorf("db error: %w", err)
	}
	if !found {
		return 0, fmt.Errorf("db error: failed to ensure owner")
	}

	return owner.ID, nil
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
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return models.Layer{}, ErrDuplicateLayerName
		}

		return models.Layer{}, fmt.Errorf("db error: %w", err)
	}

	return created, nil
}

func (r *layerRepository) UpdateLayer(ctx context.Context, id int32, name *string, description *string, status *string) (models.Layer, error) {
	changes := goqu.Record{}
	if name != nil {
		changes["name"] = *name
	}
	if description != nil {
		changes["description"] = *description
	}
	if status != nil {
		changes["status"] = *status
	}

	if len(changes) == 0 {
		return models.Layer{}, fmt.Errorf("No fields to update")
	}

	query := r.db.Update("layers").
		Set(changes).
		Where(goqu.Ex{"id": id}).
		Returning(
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

	var updated models.Layer
	found, err := query.Executor().ScanStructContext(ctx, &updated)
	if err != nil {
		if strings.Contains(err.Error(), "SQLSTATE 23505") {
			return models.Layer{}, ErrDuplicateLayerName
		}

		return models.Layer{}, fmt.Errorf("db error: %w", err)
	}

	if !found {
		return models.Layer{}, ErrLayerNotFound
	}

	return updated, nil
}

func (r *layerRepository) ArchiveLayer(ctx context.Context, id int32) (models.Layer, error) {
	query := r.db.Update("layers").
		Set(goqu.Record{"status": "archived"}).
		Where(goqu.Ex{"id": id}).
		Returning(
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

	var archived models.Layer
	found, err := query.Executor().ScanStructContext(ctx, &archived)
	if err != nil {
		return models.Layer{}, fmt.Errorf("db error: %w", err)
	}

	if !found {
		return models.Layer{}, ErrLayerNotFound
	}

	return archived, nil
}
