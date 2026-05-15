package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Comment struct {
	AuthorID  int32     `bson:"author_id" validate:"required"`
	Text      string    `bson:"text" validate:"required,min=5,max=500"`
	CreatedAt time.Time `bson:"created_at"`
}

type Revision struct {
	ID        bson.ObjectID `bson:"_id,omitempty"`
	FeatureID int32         `bson:"feature_id" validate:"required,gt=0"`
	AuthorID  int32         `bson:"author_id" validate:"required"`
	ChangeLog string        `bson:"change_log" validate:"required,min=10"`
	Comments  []Comment     `bson:"comments"`
	Status    string        `bson:"status"`
	CreatedAt time.Time     `bson:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at"`
}

func (r *Revision) Prepare() {
	r.CreatedAt = time.Now().UTC()
	r.UpdatedAt = time.Now().UTC()
	r.Status = "PENDING_REVIEW"
	if r.Comments == nil {
		r.Comments = make([]Comment, 0)
	}

	if len(r.ChangeLog) > 0 {
		r.ChangeLog = r.ChangeLog[0:1] + r.ChangeLog[1:]
	}
}

func (r *Revision) CanBeEditedBy(userID int32) bool {
	return r.AuthorID == userID
}

func (c *Comment) Prepare() {
	c.CreatedAt = time.Now().UTC()
}
