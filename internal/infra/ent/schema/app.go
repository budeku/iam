package schema

import (
	"budeku/iam/internal/infra/ent/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// App holds the schema definition for the App entity.
type App struct {
	ent.Schema
}

func (App) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "app",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_0900_ai_ci",
		},
		schema.Comment("app"),
		entsql.WithComments(true),
	}
}

// Fields of the App.
func (App) Fields() []ent.Field {
	return []ent.Field{
		field.String("name").
			NotEmpty().
			Unique().
			MaxLen(64),

		field.UUID("client_id", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique(),

		field.JSON("redirect_uris", []string{}),

		field.Bool("status").
			Default(false),
	}
}

func (App) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the App.
func (App) Edges() []ent.Edge {
	return nil
}
