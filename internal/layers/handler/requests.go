package handler


type GetLayersRequest struct {
    GeometryType *string `query:"geometry_type" validate:"omitempty,oneof=POINT LINESTRING POLYGON MULTIPOINT MULTILINESTRING MULTIPOLYGON COLLECTION"`
    Status       *string `query:"status" validate:"omitempty,oneof=active draft archived"`
    SortBy       *string `query:"sortBy" validate:"omitempty"`
    Page         int     `query:"page" validate:"omitempty,min=1"`
    PageSize     int     `query:"page_size" validate:"omitempty,min=1,max=100"`
}
