package handler

type CreateAnnotationRequest struct {
	LayerID int32    `json:"layer_id" validate:"required,gt=0"`
	Text    string   `json:"text" validate:"required,min=1"`
	Lng     *float64 `json:"lng" validate:"required,gte=-180,lte=180"`
	Lat     *float64 `json:"lat" validate:"required,gte=-90,lte=90"`
}

type NearbyAnnotationsRequest struct {
	Lat         *float64 `query:"lat" validate:"required,gte=-90,lte=90"`
	Lng         *float64 `query:"lng" validate:"required,gte=-180,lte=180"`
	MaxDistance *float64 `query:"max_distance" validate:"required,gt=0"`
}
