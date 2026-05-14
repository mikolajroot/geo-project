package handler

type CreateAnnotationRequest struct {
	FeatureID int32    `json:"feature_id" validate:"required,gt=0"`
	Text      string   `json:"text" validate:"required,min=1"`
	Lng       *float64 `json:"lng" validate:"required,gte=-180,lte=180"`
	Lat       *float64 `json:"lat" validate:"required,gte=-90,lte=90"`
}

type NearbyAnnotationsRequest struct {
	Lat         *float64 `query:"lat" validate:"required,gte=-90,lte=90"`
	Lng         *float64 `query:"lng" validate:"required,gte=-180,lte=180"`
	MaxDistance *float64 `query:"max_distance" validate:"required,gt=0"`
}

type PatchAnnotationRequest struct {
	ID   string `param:"id" validate:"required,len=24,hexadecimal"`
	Text string `json:"text" validate:"required,min=1"`
}

type ListAnnotationsRequest struct {
	Features string `query:"features" validate:"required"`
}

type DeleteAnnotationRequest struct {
	ID string `param:"id" validate:"required,len=24,hexadecimal"`
}
