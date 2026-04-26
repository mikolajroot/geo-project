package handler



type GetLayersRequest struct {
    GeometryType *string `query:"geometry_type" validate:"omitempty,oneof=POINT LINESTRING POLYGON MULTIPOINT MULTILINESTRING MULTIPOLYGON COLLECTION"`
    Status       *string `query:"status" validate:"omitempty,oneof=active draft archived"`
    Page         int     `query:"page" validate:"omitempty,min=1"`
    PageSize     int     `query:"page_size" validate:"omitempty,min=1,max=100"`
}

type LayerResponse struct {
    ID           int32  `json:"id"`
    Name         string `json:"name"`
	Description  *string `json:"description"`
    GeometryType string `json:"geometry_type"`
    Status       string `json:"status"`
    OwnerID      int32  `json:"owner_id"`
}