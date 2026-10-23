package schema

import (
	"budeku/iam/internal/enum"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
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

		field.Int("app_id").
			Immutable(),

		field.Int("user_id").
			Immutable(),

		field.String("ip").
			MaxLen(45).
			Optional(),

		field.Text("user_agent").
			MaxLen(512).
			Optional(),

		field.Uint8("auth_method").
			GoType(enum.SessionAuthMethod(0)).
			Default(uint8(enum.SessionAuthMethodPassword)),

		field.Time("authenticated_at").
			Default(time.Now).
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

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
	return []ent.Edge{
		// 多会话属于一个用户
		edge.From("user", User.Type).
			Ref("sessions").
			Unique().
			Required().
			Field("user_id").
			Immutable(),

		// 一个会话可以有多个刷新令牌
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
