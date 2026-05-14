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
	Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error)
	UpdateAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error)
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

func (r *annotationsRepository) Nearby(ctx context.Context, lat float64, lng float64, maxDistance float64) ([]model.NearbyAnnotation, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$geoNear", Value: bson.D{
			{Key: "near", Value: bson.D{
				{Key: "type", Value: "Point"},
				{Key: "coordinates", Value: bson.A{lng, lat}},
			}},
			{Key: "distanceField", Value: "distance_meters"},
			{Key: "maxDistance", Value: maxDistance},
			{Key: "spherical", Value: true},
			{Key: "key", Value: "location"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to search nearby annotations")
	}
	defer cursor.Close(ctx)

	results := make([]model.NearbyAnnotation, 0)
	for cursor.Next(ctx) {
		var item model.NearbyAnnotation
		if err := cursor.Decode(&item); err != nil {
			return nil, apperrors.NewAppError("BAD_REQUEST", "failed to decode nearby annotation")
		}
		results = append(results, item)
	}

	if err := cursor.Err(); err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to read nearby annotations")
	}

	return results, nil
}

func (r *annotationsRepository) UpdateAnnotationText(ctx context.Context, id string, authorID int32, text string) (model.Annotation, error) {
	oid, err := bson.ObjectIDFromHex(id)
	if err != nil {
		return model.Annotation{}, apperrors.NewAppError("BAD_REQUEST", "invalid annotation id" + err.Error())
	}

	filter := bson.M{"_id": oid, "author_id": authorID}
	update := bson.M{
		"$set": bson.M{
			"text": text,
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return model.Annotation{}, apperrors.NewAppError("BAD_REQUEST", "failed to update annotation")
	}
	if result.MatchedCount == 0 {
		return model.Annotation{}, apperrors.NewAppError("FORBIDDEN", "you are not the author of this annotation")
	}

	var annotation model.Annotation
	if err := r.collection.FindOne(ctx, filter).Decode(&annotation); err != nil {
		return model.Annotation{}, apperrors.NewAppError("BAD_REQUEST", "failed to load updated annotation")
	}

	return annotation, nil
}
