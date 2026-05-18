package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MongoGeoJSON struct {
	Type        string    `bson:"type"`
	Coordinates []float64 `bson:"coordinates"`
}

type Annotation struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	AuthorID  int32         `bson:"author_id"`
	FeatureID int32         `bson:"feature_id"`
	Text      string        `bson:"text"`
	Location  MongoGeoJSON  `bson:"location"`
	CreatedAt time.Time     `bson:"created_at"`
}

type NearbyAnnotation struct {
	ID             bson.ObjectID `bson:"_id,omitempty"`
	AuthorID       int32         `bson:"author_id"`
	FeatureID      int32         `bson:"feature_id"`
	Text           string        `bson:"text"`
	Location       MongoGeoJSON  `bson:"location"`
	CreatedAt      time.Time     `bson:"created_at"`
	DistanceMeters float64       `bson:"distance_meters"`
}

type FeatureStats struct {
	AuthorID         int32     `bson:"author_id"`
	TotalAnnotations int       `bson:"total_annotations"`
	LatestActivity   time.Time `bson:"latest_activity"`
}