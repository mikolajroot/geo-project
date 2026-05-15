package handler

type CreateRevisionRequest struct {
	FeatureID int32  `json:"feature_id" validate:"required,gt=0"`
	ChangeLog string `json:"change_log" validate:"required,min=10"`
}

type ListRevisionsRequest struct {
	FeatureID int32 `query:"feature_id" validate:"required,gt=0"`
}

type AddCommentRequest struct {
	ID   string `param:"id" validate:"required,len=24,hexadecimal"`
	Text string `json:"text" validate:"required,min=5,max=500"`
}
