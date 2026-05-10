package handler

import "encoding/json"

type NearbyFeatureResponse struct {
	ID             int32           `json:"id"`
	LayerID        int32           `json:"layer_id"`
	OwnerID        int32           `json:"owner_id"`
	Name           string          `json:"name"`
	Type           string          `json:"type"`
	Geometry       json.RawMessage `json:"geometry"`
	Properties     json.RawMessage `json:"properties"`
	DistanceMeters float64         `json:"distance_m"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

type NearbyResponse struct {
	Data []NearbyFeatureResponse `json:"data"`
}
