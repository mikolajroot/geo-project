package seeder

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"time"

	"geo-project/internal/annotations/model"

	"github.com/brianvoe/gofakeit/v6"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type Seeder struct {
	collection *mongo.Collection
	sqlDB      *sql.DB
}

func NewSeeder(db *mongo.Database, sqlDB *sql.DB) *Seeder {
	return &Seeder{
		collection: db.Collection("annotations"),
		sqlDB:      sqlDB,
	}
}

func (s *Seeder) SeedDomainData(ctx context.Context, n int) error {
	log.Printf("Starting seeding %d random annotations...", n)

	if _, err := s.collection.DeleteMany(ctx, bson.M{}); err != nil {
		return fmt.Errorf("failed to clear annotations collection: %w", err)
	}

	featureIDs, err := s.fetchFeatureIDs(ctx)
	if err != nil {
		return err
	}
	if len(featureIDs) == 0 {
		return fmt.Errorf("no features found, cannot seed annotations")
	}

	gofakeit.Seed(0)
	rnd := rand.New(rand.NewSource(0))

	for range n {
		featureID := featureIDs[rnd.Intn(len(featureIDs))]
		annotation := model.Annotation{
			AuthorID:  int32(gofakeit.Number(1, 50)),
			FeatureID: featureID,
			Text:      gofakeit.Sentence(10),
			Location: model.MongoGeoJSON{
				Type:        "Point",
				Coordinates: []float64{gofakeit.Float64Range(-180, 180), gofakeit.Float64Range(-90, 90)},
			},
			CreatedAt: time.Now().UTC(),
		}

		if _, err := s.collection.InsertOne(ctx, annotation); err != nil {
			return fmt.Errorf("failed to insert annotation seed: %w", err)
		}
	}

	log.Println("Annotation seeding ended successfully")
	return nil
}

func (s *Seeder) fetchFeatureIDs(ctx context.Context) ([]int32, error) {
	rows, err := s.sqlDB.QueryContext(ctx, "SELECT id FROM features ORDER BY id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch features for annotations seed: %w", err)
	}
	defer rows.Close()

	ids := make([]int32, 0)
	for rows.Next() {
		var id int32
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan feature id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed while iterating feature ids: %w", err)
	}

	return ids, nil
}
