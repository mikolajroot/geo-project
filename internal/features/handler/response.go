package handler

type FeatureResponse struct {
	ID         int32  `json:"id"`
	LayerID    int32  `json:"layer_id"`
	OwnerID    int32  `json:"owner_id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	Geometry   string `json:"geometry"`
	Properties string `json:"properties"`
	CreatedAt  string `json:"created_at"`
	UpdatedAt  string `json:"updated_at"`
}
