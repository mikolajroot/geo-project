package handler

type CreateAnnotationRequest struct {
	LayerID int32    `json:"layer_id" validate:"required,gt=0"`
	Text    string   `json:"text" validate:"required,min=1"`
	Lng     *float64 `json:"lng" validate:"required"`
	Lat     *float64 `json:"lat" validate:"required"`
}
