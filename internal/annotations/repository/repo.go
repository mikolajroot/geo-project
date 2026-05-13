package repository

import (
	"context"
	"log"
	"time"

	"geo-project/internal/annotations/model"
	apperrors "geo-project/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AnnotationsRepository interface {
	CreateAnnotation(ctx context.Context, annotation model.Annotation) (model.Annotation, error)
}

type annotationsRepository struct {
	collection *mongo.Collection
}

func NewAnnotationsRepository(db *mongo.Database) AnnotationsRepository {
	collection := db.Collection("annotations")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	indexModel := mongo.IndexModel{
		Keys: bson.D{{Key: "location", Value: "2dsphere"}},
	}

	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Warning: Failed to create 2dsphere index: %v", err)
	}

	return &annotationsRepository{collection: collection}
}

func (r *annotationsRepository) CreateAnnotation(ctx context.Context, annotation model.Annotation) (model.Annotation, error) {
	result, err := r.collection.InsertOne(ctx, annotation)
	if err != nil {
		return model.Annotation{}, apperrors.NewAppError("BAD_REQUEST", "failed to create annotation")
	}

	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		annotation.ID = oid
	}

	return annotation, nil
}
