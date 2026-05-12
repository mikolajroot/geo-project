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
	DistanceMeters float64         `json:"distance_m,omitempty"`
	CreatedAt      string          `json:"created_at"`
	UpdatedAt      string          `json:"updated_at"`
}

type NearbyResponse struct {
	Data []NearbyFeatureResponse `json:"data"`
}

type LayerStatsResponse struct {
	LayerID            int32            `json:"layer_id"`
	TotalFeatures      int64            `json:"total_features"`
	TotalAreaSqMeters  float64          `json:"total_area_sq_meters"`
	TotalLengthMeters  float64          `json:"total_length_meters"`
	LayerExtent        []float64        `json:"layer_extent,omitempty"`
	LastUpdatedFeature string           `json:"last_updated_feature,omitempty"`
	Type               string           `json:"type,omitempty"`
	FeatureTypesCount  map[string]int64 `json:"feature_types_count"`
}
