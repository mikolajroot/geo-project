package handler

import "time"

type LayerResponse struct {
	ID           int32   `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	GeometryType string  `json:"geometry_type"`
	Status       string  `json:"status"`
	OwnerID      int32   `json:"owner_id"`
	CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type PaginationResponse struct {
	TotalPages  int  `json:"total_pages"`
	Page 	int  `json:"page"`
}

type okResponse[T any] struct {
	Data  	[]T
	Meta	PaginationResponse
}
