package model

import "time"

type NearbyFeature struct {
	ID             int32     `json:"id"`
	LayerID        int32     `json:"layer_id"`
	OwnerID        int32     `json:"owner_id"`
	Name           string    `json:"name"`
	Type           string    `json:"type"`
	Geometry       string    `json:"geometry"`
	Properties     string    `json:"properties"`
	DistanceMeters float64   `json:"distance_m"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
