package handler

type GetLayersRequest struct {
	GeometryType *string `query:"geometry_type" validate:"omitempty,oneof=POINT LINESTRING POLYGON MULTIPOINT MULTILINESTRING MULTIPOLYGON COLLECTION"`
	Status       *string `query:"status" validate:"omitempty,oneof=active draft archived"`
	SortBy       *string `query:"sort_by" validate:"omitempty,oneof=id name geometry_type status owner_id created_at updated_at"`
	Page         int     `query:"page" validate:"omitempty,min=1"`
	PageSize     int     `query:"page_size" validate:"omitempty,min=1,max=100"`
}

type CreateLayerRequest struct {
	Name         string `json:"name" validate:"required,min=3"`
	Description  string `json:"description" validate:"omitempty"`
	GeometryType string `json:"geometry_type" validate:"required,oneof=POINT LINESTRING POLYGON MULTIPOINT MULTILINESTRING MULTIPOLYGON COLLECTION"`
	SRID         int32  `json:"srid" validate:"required,min=1"`
	OwnerID      *int32 `json:"owner_id" validate:"omitempty,min=1"`
}

type GetLayerByIDRequest struct {
	ID int32 `param:"id" validate:"required,min=1"`
}
