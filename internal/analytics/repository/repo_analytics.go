package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"geo-project/internal/analytics/model"
	apperrors "geo-project/pkg/errors"
)

type AnalyticsRepository interface {
	Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error)
}

type analyticsRepository struct {
	pgx *pgxpool.Pool
}

func NewAnalyticsRepository(pgx *pgxpool.Pool) AnalyticsRepository {
	return &analyticsRepository{pgx: pgx}
}

func (r *analyticsRepository) Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error) {
	sql := `SELECT id, layer_id, owner_id, name, type, ST_AsGeoJSON(geometry) as geometry, properties::text, created_at, updated_at,
        ST_Distance(geometry::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_m
        FROM features
        WHERE layer_id = $3
          AND ST_DWithin(geometry::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $4)
        ORDER BY distance_m ASC
        LIMIT 1000;`

	rows, err := r.pgx.Query(ctx, sql, lng, lat, layerID, radius)
	if err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", fmt.Sprintf("nearby query failed: %v", err))
	}
	defer rows.Close()

	var out []model.NearbyFeature
	for rows.Next() {
		var f model.NearbyFeature
		if err := rows.Scan(&f.ID, &f.LayerID, &f.OwnerID, &f.Name, &f.Type, &f.Geometry, &f.Properties, &f.CreatedAt, &f.UpdatedAt, &f.DistanceMeters); err != nil {
			return nil, apperrors.NewAppError("BAD_REQUEST", fmt.Sprintf("scan failed: %v", err))
		}
		out = append(out, f)
	}

	return out, nil
}
