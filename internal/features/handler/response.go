package handler

import "encoding/json"

type OwnerResponse struct {
	ID    int32  `json:"id"`
	Login string `json:"login"`
}

type FeatureResponse struct {
	ID         int32           `json:"id"`
	LayerID    int32           `json:"layer_id"`
	OwnerID    int32           `json:"owner_id"`
	Name       string          `json:"name"`
	Type       string          `json:"type"`
	Geometry   json.RawMessage `json:"geometry"`
	Properties json.RawMessage `json:"properties"`
	Owner      *OwnerResponse  `json:"owner,omitempty"`
	CreatedAt  string          `json:"created_at"`
	UpdatedAt  string          `json:"updated_at"`
}

type PaginationResponse struct {
	TotalPages int `json:"total_pages"`
	Page       int `json:"page"`
}

type okResponse[T any] struct {
	Data []T
	Meta PaginationResponse
}
