package repository

import (
	"context"

	"geo-project/internal/revisions/model"
	apperrors "geo-project/pkg/errors"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type RevisionRepository interface {
	CreateRevision(ctx context.Context, revision *model.Revision) (*model.Revision, error)
	ListByFeatureID(ctx context.Context, featureID int32) ([]*model.Revision, error)
}

func (r *revisionsRepository) CreateRevision(ctx context.Context, revision *model.Revision) (*model.Revision, error) {
	result, err := r.collection.InsertOne(ctx, revision)
	if err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to create revision")
	}
	if oid, ok := result.InsertedID.(bson.ObjectID); ok {
		revision.ID = oid
	}
	return revision, nil
}

func (r *revisionsRepository) ListByFeatureID(ctx context.Context, featureID int32) ([]*model.Revision, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"feature_id": featureID})
	if err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to fetch revisions")
	}
	defer cursor.Close(ctx)

	var revisions []*model.Revision
	if err := cursor.All(ctx, &revisions); err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to decode revisions")
	}
	return revisions, nil
}

type revisionsRepository struct {
	collection *mongo.Collection
}

func NewRevisionRepository(db *mongo.Database) RevisionRepository {
	collection := db.Collection("revisions")
	return &revisionsRepository{collection: collection}
}
