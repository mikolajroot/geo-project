package models

import "time"

type Layer struct {
	ID           int32     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name" validate:"required,min=3"`
	Description  string    `db:"description" json:"description"`
	GeometryType string    `db:"geometry_type" json:"geometry_type" validate:"required,oneof=POINT LINESTRING POLYGON MULTIPOINT MULTILINESTRING MULTIPOLYGON COLLECTION"`
	Status       string    `db:"status" json:"status" validate:"oneof=active draft archived"`
	SRID         int32     `db:"srid" json:"srid"`
	OwnerID      int32     `db:"owner_id" json:"owner_id"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
