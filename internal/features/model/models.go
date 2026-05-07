package model

import (
	"time"
	"gorm.io/gorm"
)


type Owner struct {
	ID         int32     `gorm:"primaryKey" json:"id"`
	ExternalID int32     `gorm:"uniqueIndex;not null" json:"external_id"`
	Login      string    `gorm:"size:100;not null" json:"name"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	
	Features   []Feature `gorm:"foreignKey:OwnerID" json:"features,omitempty"`
}


type Feature struct {
	ID          int32     `gorm:"primaryKey" json:"id"`
	LayerID     int32     `gorm:"not null;index" json:"layer_id"`
	OwnerID     int32     `gorm:"not null;index" json:"owner_id"`
	Name        string    `gorm:"size:255;not null" json:"name"`
	Type        string    `gorm:"size:50;index" json:"type"`
	Geometry    string    `gorm:"type:geometry(Geometry,4326);index:idx_geom,type:gist" json:"geometry"`
	Properties  string    `gorm:"type:jsonb" json:"properties"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	Owner       Owner     `gorm:"foreignKey:OwnerID" json:"owner"`
}

func (f *Feature) BeforeCreate(tx *gorm.DB) (err error) {
	if f.Name == "" {
		f.Name = "Unnamed Feature"
	}
	return nil
}