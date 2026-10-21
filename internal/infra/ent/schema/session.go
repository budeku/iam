package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "session",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_0900_ai_ci",
		},
		schema.Comment("session"),
		entsql.WithComments(true),
	}
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("identity", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique(),

		field.Int("app_id"),

		field.Int("user_id"),

		field.String("ip").
			MaxLen(45).
			Optional(),

		field.Text("user_agent").
			MaxLen(512).
			Optional(),

		field.Time("last_active_at").
			Default(time.Now).
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

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

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return nil
}
