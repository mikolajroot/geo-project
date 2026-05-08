package handler

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
