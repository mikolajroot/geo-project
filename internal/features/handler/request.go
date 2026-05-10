package handler

import (
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

func RegisterCustomValidators(v *validator.Validate) {
	v.RegisterValidation("bbox", validateBBox)
}

func validateBBox(fl validator.FieldLevel) bool {
	bbox := fl.Field().String()
	if bbox == "" {
		return true
	}

	parts := strings.Split(bbox, ",")
	if len(parts) != 4 {
		return false
	}

	floatPattern := regexp.MustCompile(`^-?\d+(\.\d+)?$`)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if !floatPattern.MatchString(part) {
			return false
		}
	}

	return true
}

type CreateFeatureRequest struct {
	Name       string         `json:"name" validate:"required"`
	Type       string         `json:"type" validate:"required"`
	Geometry   map[string]any `json:"geometry" validate:"required"`
	Properties map[string]any `json:"properties"`
	LayerID    int32          `json:"layer_id" validate:"required,gt=0"`
}

type UpdateFeatureRequest struct {
	ID         int32          `param:"id" validate:"required,gt=0"`
	Geometry   map[string]any `json:"geometry" validate:"required_without=Properties"`
	Properties map[string]any `json:"properties" validate:"required_without=Geometry"`
}

type DeleteFeatureRequest struct {
	ID int32 `param:"id" validate:"required,gt=0"`
}

type GetFeatureRequest struct {
	ID int32 `param:"id" validate:"required,gt=0"`
}

type GetFeaturesByLayerRequest struct {
	LayerID  int32  `param:"layer_id" validate:"required,gt=0"`
	Page     int    `query:"page" validate:"omitempty,gt=0"`
	PageSize int    `query:"page_size" validate:"omitempty,gt=0,lte=100"`
	SortBy   string `query:"sort_by" validate:"omitempty,oneof=name type created_at"`
	Type     string `query:"type" validate:"omitempty"`
	BBox     string `query:"bbox" validate:"omitempty,bbox"`
}
