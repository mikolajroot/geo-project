package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)


type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("login").
			Unique().
			NotEmpty(),

		field.String("password_hash").
			NotEmpty().
			Sensitive(),

		field.Enum("role").
			Values("user", "admin", "editor").
			Default("user"),

		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

func (User) Edges() []ent.Edge {
	return nil
}

func (User) Indexes() []ent.Index {
	return nil
}