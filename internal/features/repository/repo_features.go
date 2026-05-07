package repository

import (
	"geo-project/internal/features/model"
	apperrors "geo-project/pkg/errors"

	"gorm.io/gorm"
)

type FeatureRepository interface {
	CreateFeatureWithOwner(owner *model.Owner, feature *model.Feature) error
}

type featureRepository struct {
	db *gorm.DB
}

func NewFeatureRepository(db *gorm.DB) FeatureRepository {
	return &featureRepository{db: db}
}

func (r *featureRepository) CreateFeatureWithOwner(owner *model.Owner, feature *model.Feature) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.Assign(model.Owner{Login: owner.Login}).FirstOrCreate(owner, model.Owner{ExternalID: owner.ExternalID})
		if result.Error != nil {
			return apperrors.NewAppError("BAD_REQUEST", "failed to process owner")
		}

		feature.OwnerID = owner.ID

		if err := feature.BeforeCreate(tx); err != nil {
			return err
		}

		err := tx.Raw(`
			INSERT INTO features (layer_id, owner_id, name, type, geometry, properties)
			VALUES (?, ?, ?, ?, ST_GeomFromGeoJSON(?), ?::jsonb)
			RETURNING 
				id, 
				layer_id, 
				owner_id, 
				name, 
				type, 
				ST_AsGeoJSON(geometry) as geometry,
				properties, 
				created_at, 
				updated_at
		`,
			feature.LayerID,
			feature.OwnerID,
			feature.Name,
			feature.Type,
			feature.Geometry,
			feature.Properties,
		).Scan(&feature).Error

		if err != nil {
			return apperrors.NewAppError("BAD_REQUEST", "failed to create feature")
		}

		return nil
	})
}
