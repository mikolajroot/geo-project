package handler

import "time"

type CreateRevisionResponse struct {
	ID        string    `json:"id"`
	FeatureID int32     `json:"feature_id"`
	AuthorID  int32     `json:"author_id"`
	ChangeLog string    `json:"change_log"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListRevisionResponse struct {
	ID        string    `json:"id"`
	FeatureID int32     `json:"feature_id"`
	AuthorID  int32     `json:"author_id"`
	ChangeLog string    `json:"change_log"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AddCommentResponse struct {
	Message string `json:"message"`
}

type CommentResponse struct {
	AuthorID  int32     `json:"author_id"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type GetRevisionResponse struct {
	ID          string            `json:"id"`
	FeatureID   int32             `json:"feature_id"`
	FeatureName string            `json:"feature_name,omitempty"`
	AuthorID    int32             `json:"author_id"`
	ChangeLog   string            `json:"change_log"`
	Comments    []CommentResponse `json:"comments"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}
