package repository

import (
	"geo-project/internal/features/model"
	apperrors "geo-project/pkg/errors"

	"gorm.io/gorm"
)

type FeatureRepository interface {
	CreateFeatureWithOwner(owner *model.Owner, feature *model.Feature) error
	GetOwnerByExternalID(externalID int32) (*model.Owner, error)
}

type featureRepository struct {
	db *gorm.DB
}

func NewFeatureRepository(db *gorm.DB) FeatureRepository {
	return &featureRepository{db: db}
}

func (r *featureRepository) CreateFeatureWithOwner(owner *model.Owner, feature *model.Feature) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		result := tx.FirstOrCreate(owner, model.Owner{ExternalID: owner.ExternalID})
		if result.Error != nil {
			return apperrors.NewAppError("BAD_REQUEST", "failed to process owner")
		}

		feature.OwnerID = owner.ID

		if err := tx.Create(feature).Error; err != nil {
			return apperrors.NewAppError("BAD_REQUEST", "failed to create feature")
		}

		return nil
	})
}

func (r *featureRepository) GetOwnerByExternalID(externalID int32) (*model.Owner, error) {
	var owner model.Owner
	if err := r.db.Where("external_id = ?", externalID).First(&owner).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewAppError("NOT_FOUND", "owner not found")
		}
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to fetch owner")
	}
	return &owner, nil
}
