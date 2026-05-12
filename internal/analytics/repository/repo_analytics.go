package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"geo-project/internal/analytics/model"
	apperrors "geo-project/pkg/errors"
)

type AnalyticsRepository interface {
	Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error)
	Intersect(ctx context.Context, layerID int32, geometry string) ([]model.NearbyFeature, error)
	LayerStatsTotals(ctx context.Context, layerID int32) (model.LayerStats, error)
	LayerStatsSpatial(ctx context.Context, layerID int32) (model.LayerStats, error)
}

type analyticsRepository struct {
	pgx *pgxpool.Pool
}

func NewAnalyticsRepository(pgx *pgxpool.Pool) AnalyticsRepository {
	return &analyticsRepository{pgx: pgx}
}

func (r *analyticsRepository) ensureLayerExists(ctx context.Context, layerID int32) error {
	layerExistsSQL := `SELECT 1 FROM layers WHERE id = $1 LIMIT 1;`
	var exists int
	if err := r.pgx.QueryRow(ctx, layerExistsSQL, layerID).Scan(&exists); err != nil {
		return mapPgxError(err)
	}
	return nil
}

func (r *analyticsRepository) Nearby(ctx context.Context, layerID int32, lat float64, lng float64, radius float64) ([]model.NearbyFeature, error) {
	if err := r.ensureLayerExists(ctx, layerID); err != nil {
		return nil, err
	}

	sql := `SELECT id, layer_id, owner_id, name, type, ST_AsGeoJSON(geometry) as geometry, properties::text, created_at, updated_at,
        ST_Distance(geometry::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography) as distance_m
        FROM features
        WHERE layer_id = $3
          AND ST_DWithin(geometry::geography, ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography, $4)
        ORDER BY distance_m ASC
        LIMIT 1000;`

	rows, err := r.pgx.Query(ctx, sql, lng, lat, layerID, radius)
	if err != nil {
		return nil, mapPgxError(err)
	}
	defer rows.Close()

	var out []model.NearbyFeature
	for rows.Next() {
		var f model.NearbyFeature
		if err := rows.Scan(&f.ID, &f.LayerID, &f.OwnerID, &f.Name, &f.Type, &f.Geometry, &f.Properties, &f.CreatedAt, &f.UpdatedAt, &f.DistanceMeters); err != nil {
			return nil, mapPgxError(err)
		}
		out = append(out, f)
	}

	return out, nil
}

func (r *analyticsRepository) Intersect(ctx context.Context, layerID int32, geometry string) ([]model.NearbyFeature, error) {
	if err := r.ensureLayerExists(ctx, layerID); err != nil {
		return nil, err
	}

	sql := `SELECT id, layer_id, owner_id, name, type, ST_AsGeoJSON(geometry) as geometry, properties::text, created_at, updated_at
		FROM features
		WHERE layer_id = $1
		  AND ST_Intersects(features.geometry, ST_GeomFromGeoJSON($2))
		LIMIT 1000;`

	rows, err := r.pgx.Query(ctx, sql, layerID, geometry)
	if err != nil {
		return nil, mapPgxError(err)
	}
	defer rows.Close()

	var out []model.NearbyFeature
	for rows.Next() {
		var f model.NearbyFeature
		if err := rows.Scan(&f.ID, &f.LayerID, &f.OwnerID, &f.Name, &f.Type, &f.Geometry, &f.Properties, &f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, mapPgxError(err)
		}
		out = append(out, f)
	}

	return out, nil
}

func (r *analyticsRepository) LayerStatsTotals(ctx context.Context, layerID int32) (model.LayerStats, error) {
	if err := r.ensureLayerExists(ctx, layerID); err != nil {
		return model.LayerStats{}, err
	}

	stats := model.LayerStats{
		LayerID: layerID,
	}

	statsSQL := `SELECT
		COUNT(*) AS total_features,
		COALESCE(SUM(CASE WHEN GeometryType(geometry) IN ('POLYGON', 'MULTIPOLYGON') THEN ST_Area(geometry::geography) ELSE 0 END), 0) AS total_area_sq_meters,
		COALESCE(SUM(CASE WHEN GeometryType(geometry) IN ('LINESTRING', 'MULTILINESTRING') THEN ST_Length(geometry::geography) ELSE 0 END), 0) AS total_length_meters
		FROM features
		WHERE layer_id = $1;`

	if err := r.pgx.QueryRow(ctx, statsSQL, layerID).Scan(&stats.TotalFeatures, &stats.TotalAreaSqMeters, &stats.TotalLengthMeters); err != nil {
		return model.LayerStats{}, mapPgxError(err)
	}

	return stats, nil
}

func (r *analyticsRepository) LayerStatsSpatial(ctx context.Context, layerID int32) (model.LayerStats, error) {
	stats := model.LayerStats{
		LayerID:     layerID,
		LayerExtent: []float64{0, 0, 0, 0},
	}

	extentSQL := `SELECT
		COALESCE(ST_XMin(extent), 0) AS min_x,
		COALESCE(ST_YMin(extent), 0) AS min_y,
		COALESCE(ST_XMax(extent), 0) AS max_x,
		COALESCE(ST_YMax(extent), 0) AS max_y
		FROM (
			SELECT ST_Envelope(ST_Collect(geometry)) AS extent
			FROM features
			WHERE layer_id = $1
		) AS layer_extent;`

	var minX, minY, maxX, maxY float64
	if err := r.pgx.QueryRow(ctx, extentSQL, layerID).Scan(&minX, &minY, &maxX, &maxY); err != nil {
		return model.LayerStats{}, mapPgxError(err)
	}
	stats.LayerExtent = []float64{minX, minY, maxX, maxY}

	latestSQL := `SELECT updated_at, lower(GeometryType(geometry)) AS geometry_type
		FROM features
		WHERE layer_id = $1
		ORDER BY updated_at DESC, id DESC
		LIMIT 1;`

	rows, err := r.pgx.Query(ctx, latestSQL, layerID)
	if err != nil {
		return model.LayerStats{}, mapPgxError(err)
	}
	defer rows.Close()

	if rows.Next() {
		var updatedAt time.Time
		var geometryType string
		if err := rows.Scan(&updatedAt, &geometryType); err != nil {
			return model.LayerStats{}, mapPgxError(err)
		}
		stats.LastUpdatedFeature = updatedAt.UTC().Format(time.RFC3339)
		stats.LastUpdatedFeatureType = geometryType
	}

	return stats, nil
}

func mapPgxError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.NewAppError("NOT_FOUND", "resource not found")
	}

	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperrors.NewAppError("CONFLICT", "A record with this information already exists. Please provide unique values.")

		case "23503": // foreign_key_violation
			return apperrors.NewAppError("BAD_REQUEST", "The operation failed because a related item does not exist.")

		case "23502": // not_null_violation
			return apperrors.NewAppError("BAD_REQUEST", "A required field is missing. Please ensure all mandatory information is provided.")

		case "22012": // division_by_zero
			return apperrors.NewAppError("BAD_REQUEST", "The calculation failed due to invalid spatial input data.")

		case "42883": // undefined_function
			return apperrors.NewAppError("INTERNAL_ERROR", "We encountered an issue while processing your map data. Please contact support.")

		default:
			return apperrors.NewAppError("INTERNAL_ERROR", "An unexpected database error occurred. Please try again later.")
		}
	}
	return apperrors.NewAppError("INTERNAL_ERROR", "An unexpected server error occurred. Please try again later.")
}
