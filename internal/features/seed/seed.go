package seeder

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"gorm.io/gorm"

	"geo-project/internal/features/service"
	apperrors "geo-project/pkg/errors"
)

type Seeder struct {
	featureService service.FeatureService
	db             *gorm.DB
}

func NewSeeder(featureService service.FeatureService, db *gorm.DB) *Seeder {
	return &Seeder{
		featureService: featureService,
		db:             db,
	}
}

func (s *Seeder) SeedDomainData(ctx context.Context, n int) error {
	log.Printf("Starting seeding %d random features...", n)

	gofakeit.Seed(0)
	rnd := rand.New(rand.NewSource(0))

	layerIDs, err := s.fetchLayerIDs(ctx)
	if err != nil {
		return err
	}
	if len(layerIDs) == 0 {
		return fmt.Errorf("no layers found, cannot seed features")
	}

	featureTypes := []string{"poi", "building", "road", "parcel", "water"}
	geometryTypes := []string{"POINT", "LINESTRING", "POLYGON"}
	defaultOwner := int32(1)

	for range n {
		geometryType := gofakeit.RandomString(geometryTypes)
		geometry := randomGeoJSON(geometryType)
		featureType := gofakeit.RandomString(featureTypes)
		layerID := layerIDs[rnd.Intn(len(layerIDs))]
		name := fmt.Sprintf("%s %s", gofakeit.Street(), gofakeit.Noun())

		created, err := s.featureService.CreateFeature(
			ctx,
			defaultOwner,
			"system",
			name,
			featureType,
			geometry,
			randomPropertiesJSON(),
			layerID,
		)
		if err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) && appErr.Code() == "BAD_REQUEST" {
				log.Printf("Feature '%s' skipped: %v", name, err)
				continue
			}
			return fmt.Errorf("seeding failed for feature %q: %w", name, err)
		}

		log.Printf("Added feature: %s (id=%d, layer=%d, type=%s, geometry=%s)", created.Name, created.ID, layerID, featureType, geometryType)
	}

	log.Println("Feature seeding ended successfully")
	return nil
}

func (s *Seeder) fetchLayerIDs(ctx context.Context) ([]int32, error) {
	var ids []int32
	if err := s.db.WithContext(ctx).Table("layers").Order("id ASC").Pluck("id", &ids).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch layer ids: %w", err)
	}
	return ids, nil
}

func randomGeoJSON(geometryType string) string {
	switch geometryType {
	case "POINT":
		return fmt.Sprintf(`{"type":"Point","coordinates":[%.6f,%.6f]}`, randomLon(), randomLat())
	case "LINESTRING":
		return fmt.Sprintf(`{"type":"LineString","coordinates":[[%.6f,%.6f],[%.6f,%.6f],[%.6f,%.6f]]}`, randomLon(), randomLat(), randomLon(), randomLat(), randomLon(), randomLat())
	case "POLYGON":
		lon := randomLon()
		lat := randomLat()
		return fmt.Sprintf(`{"type":"Polygon","coordinates":[[[%.6f,%.6f],[%.6f,%.6f],[%.6f,%.6f],[%.6f,%.6f],[%.6f,%.6f]]]}`, lon, lat, lon+0.01, lat, lon+0.01, lat+0.01, lon, lat+0.01, lon, lat)
	default:
		return fmt.Sprintf(`{"type":"Point","coordinates":[%.6f,%.6f]}`, randomLon(), randomLat())
	}
}

func randomLon() float64 {
	return gofakeit.Float64Range(-180, 180)
}

func randomLat() float64 {
	return gofakeit.Float64Range(-90, 90)
}

func randomPropertiesJSON() string {
	return fmt.Sprintf(`{"category":"%s","status":"%s","source":"seed"}`, gofakeit.RandomString([]string{"infrastructure", "landuse", "hydro", "poi"}), gofakeit.RandomString([]string{"new", "mapped", "verified"}))
}

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}
