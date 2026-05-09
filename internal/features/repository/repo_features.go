package repository

import (
	"geo-project/internal/features/model"
	apperrors "geo-project/pkg/errors"

	"gorm.io/gorm"
)

type FeatureRepository interface {
	CreateFeatureWithOwner(owner *model.Owner, feature *model.Feature) error
	UpdateFeatureByIDAndOwner(featureID int32, ownerExternalID int32, geometry *string, properties *string) (*model.Feature, error)
	GetFeatureByIDWithOwner(featureID int32) (*model.Feature, error)
	GetFeaturesByLayer(layerID int32, featureType string, sortBy string, page int, pageSize int) ([]model.Feature, int, error)
	DeleteFeatureByID(featureID int32) error
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

func (r *featureRepository) UpdateFeatureByIDAndOwner(featureID int32, ownerExternalID int32, geometry *string, properties *string) (*model.Feature, error) {
	var feature model.Feature
	if err := r.db.
		Preload("Owner").
		First(&feature, featureID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewAppError("NOT_FOUND", "feature not found")
		}
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to load feature")
	}

	if feature.Owner.ExternalID != ownerExternalID {
		return nil, apperrors.NewAppError("FORBIDDEN", "you are not the owner of this feature")
	}

	updates := map[string]any{}
	if geometry != nil {
		updates["geometry"] = gorm.Expr("ST_GeomFromGeoJSON(?)", *geometry)
	}
	if properties != nil {
		updates["properties"] = gorm.Expr("?::jsonb", *properties)
	}

	if len(updates) == 0 {
		return &feature, nil
	}

	if err := r.db.Model(&feature).Updates(updates).Error; err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to update feature")
	}

	if err := r.db.
		Table("features").
		Select("features.id, features.layer_id, features.owner_id, features.name, features.type, ST_AsGeoJSON(features.geometry) as geometry, features.properties, features.created_at, features.updated_at").
		Where("features.id = ?", featureID).
		First(&feature).Error; err != nil {
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to reload updated feature")
	}

	return &feature, nil
}

func (r *featureRepository) GetFeatureByIDWithOwner(featureID int32) (*model.Feature, error) {
	var feature model.Feature
	if err := r.db.
		Preload("Owner").
		Table("features").
		Select("features.id, features.layer_id, features.owner_id, features.name, features.type, ST_AsGeoJSON(features.geometry) as geometry, features.properties, features.created_at, features.updated_at").
		Where("features.id = ?", featureID).
		First(&feature).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.NewAppError("NOT_FOUND", "feature not found")
		}
		return nil, apperrors.NewAppError("BAD_REQUEST", "failed to load feature")
	}
	return &feature, nil
}

func (r *featureRepository) GetFeaturesByLayer(layerID int32, featureType string, sortBy string, page int, pageSize int) ([]model.Feature, int, error) {
	baseQuery := r.db.Model(&model.Feature{}).Where("layer_id = ?", layerID)
	if featureType != "" {
		baseQuery = baseQuery.Where("type = ?", featureType)
	}

	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, apperrors.NewAppError("BAD_REQUEST", "failed to count features")
	}

	sortColumns := map[string]string{
		"name":       "features.name",
		"type":       "features.type",
		"created_at": "features.created_at",
	}
	sortColumn, ok := sortColumns[sortBy]
	if !ok {
		sortColumn = "features.created_at"
		sortBy = "created_at"
	}

	listQuery := r.db.
		Table("features").
		Select("features.id, features.layer_id, features.owner_id, features.name, features.type, ST_AsGeoJSON(features.geometry) as geometry, features.properties, features.created_at, features.updated_at").
		Preload("Owner").
		Where("features.layer_id = ?", layerID)
	if featureType != "" {
		listQuery = listQuery.Where("features.type = ?", featureType)
	}

	switch sortBy {
	case "name", "type":
		listQuery = listQuery.Order(sortColumn + " ASC")
	default:
		listQuery = listQuery.Order(sortColumn + " DESC")
	}

	offset := (page - 1) * pageSize
	var features []model.Feature
	if err := listQuery.Offset(offset).Limit(pageSize).Find(&features).Error; err != nil {
		return nil, 0, apperrors.NewAppError("BAD_REQUEST", "failed to load features")
	}

	return features, int(total), nil
}

func (r *featureRepository) DeleteFeatureByID(featureID int32) error {
	if err := r.db.Delete(&model.Feature{}, featureID).Error; err != nil {
		return apperrors.NewAppError("BAD_REQUEST", "failed to delete feature")
	}
	return nil
}
