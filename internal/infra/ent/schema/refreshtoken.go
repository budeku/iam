package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
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
		field.Int("app_id").
			Immutable(),

		field.Int("session_id").
			Immutable(),

		field.Int("user_id").
			Immutable(),

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
	return []ent.Edge{
		// 多个刷新令牌属于一个会话
		edge.From("session", Session.Type).
			Ref("refresh_tokens").
			Unique().
			Required().
			Field("session_id").
			Immutable(),

		// 多个刷新令牌属于一个用户
		edge.From("user", User.Type).
			Ref("refresh_tokens").
			Unique().
			Required().
			Field("user_id").
			Immutable(),
	}
}
