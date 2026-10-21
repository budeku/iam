package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// RefreshToken holds the schema definition for the RefreshToken entity.
type RefreshToken struct {
	ent.Schema
}

func (RefreshToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "refresh_token",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_0900_ai_ci",
		},
		schema.Comment("refresh_token"),
		entsql.WithComments(true),
	}
}

// Fields of the RefreshToken.
func (RefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.Int("app_id"),

		field.Int("session_id"),

		field.Int("user_id"),

		field.Bytes("identity").
			Unique().
			MaxLen(32).
			SchemaType(map[string]string{
				"mysql": "binary(32)",
			}),

		field.UUID("jti", uuid.UUID{}).
			Optional().
			Unique().
			Immutable(),

		field.Time("revoked_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Time("expires_at").
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Time("created_at").
			Default(time.Now).
			Immutable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now).
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),
	}
}

// Edges of the RefreshToken.
func (RefreshToken) Edges() []ent.Edge {
	return nil
}
