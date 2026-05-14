package handler

import "time"

type CreateAnnotationResponse struct {
	ID        string      `json:"id"`
	AuthorID  int32       `json:"author_id"`
	FeatureID int32       `json:"feature_id"`
	Text      string      `json:"text"`
	Location  interface{} `json:"location"`
	CreatedAt time.Time   `json:"created_at"`
}

type NearbyAnnotationResponse struct {
	ID             string      `json:"id"`
	AuthorID       int32       `json:"author_id"`
	FeatureID      int32       `json:"feature_id"`
	Text           string      `json:"text"`
	Location       interface{} `json:"location"`
	CreatedAt      time.Time   `json:"created_at"`
	DistanceMeters float64     `json:"distance_meters"`
}

type NearbyAnnotationsResponse struct {
	Data []NearbyAnnotationResponse `json:"data"`
}

type AnnotationsResponse struct {
	Data []CreateAnnotationResponse `json:"data"`
}
