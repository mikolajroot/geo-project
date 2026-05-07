package handler

type CreateFeatureRequest struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Geometry   string `json:"geometry" validate:"required"`
	Properties string `json:"properties"`
	LayerID    int32  `json:"layer_id" validate:"required,gt=0"`
}
