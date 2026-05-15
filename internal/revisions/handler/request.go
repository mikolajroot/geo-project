package handler

type CreateRevisionRequest struct {
	FeatureID int32  `json:"feature_id" validate:"required,gt=0"`
	ChangeLog string `json:"change_log" validate:"required,min=10"`
}
