package handler

import "encoding/json"

type NearbyRequest struct {
	LayerID      int32   `query:"layer_id" validate:"required,gt=0"`
	Lat          float64 `query:"lat" validate:"required,gte=-90,lte=90"`
	Lng          float64 `query:"lng" validate:"required,gte=-180,lte=180"`
	RadiusMeters float64 `query:"radius_meters" validate:"required,gt=0"`
}

type IntersectRequest struct {
	LayerID  int32           `json:"layer_id" validate:"required,gt=0"`
	Geometry json.RawMessage `json:"geometry" validate:"required"`
}
