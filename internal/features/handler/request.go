package handler

type CreateFeatureRequest struct {
	Name       string `json:"name" validate:"required"`
	Type       string `json:"type" validate:"required"`
	Geometry   map[string]interface{} `json:"geometry" validate:"required"`
	Properties map[string]interface{} `json:"properties"`
	LayerID    int32  `json:"layer_id" validate:"required,gt=0"`
}
