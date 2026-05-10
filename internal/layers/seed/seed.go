package seeder

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/doug-martin/goqu/v9"

	"geo-project/internal/layers/service"
	apperrors "geo-project/pkg/errors"
)

type Seeder struct {
	layerService service.LayerService
	goquDB       *goqu.Database
}

func NewSeeder(layerService service.LayerService, goquDB *goqu.Database) *Seeder {
	return &Seeder{
		layerService: layerService,
		goquDB:       goquDB,
	}
}

func (s *Seeder) SeedDomainData(ctx context.Context, n int) error {
	log.Printf("Starting seeding %d random layers...", n)

	if s.goquDB != nil {
		if _, err := s.goquDB.Exec("TRUNCATE TABLE layers RESTART IDENTITY CASCADE"); err != nil {
			log.Printf("warning: truncate layers failed: %v", err)
		}
	}

	gofakeit.Seed(0)

	geometryTypes := []string{"POINT", "LINESTRING", "POLYGON", "MULTIPOINT", "MULTIPOLYGON"}
	defaultOwner := int32(1)

	for range n {
		randomName := fmt.Sprintf("%s - %s", gofakeit.City(), gofakeit.JobDescriptor())

		params := service.CreateLayerParams{
			Name:         randomName,
			Description:  gofakeit.Sentence(8),
			GeometryType: gofakeit.RandomString(geometryTypes),
			SRID:         4326,
			OwnerID:      &defaultOwner,
		}

		_, err := s.layerService.CreateLayer(ctx, params)
		if err != nil {
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) && appErr.Code() == "CONFLICT" {
				log.Printf("Layer '%s' already exist, skiping.", params.Name)
				continue
			}
			log.Printf("Critical error when seeding '%s': %v", params.Name, err)
			return fmt.Errorf("seeding failed: %w", err)
		} else {
			log.Printf("Added: %s (%s)", params.Name, params.GeometryType)
		}
	}

	log.Println("Seeding ended succesfully")
	return nil
}
