package schema

import (
	"budeku/iam/internal/infra/ent/mixin"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "user",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_0900_ai_ci",
		},
		schema.Comment("user"),
		entsql.WithComments(true),
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").
			NotEmpty().
			Unique().
			MaxLen(32),

		field.UUID("uuid", uuid.UUID{}).
			Default(uuid.New).
			Immutable().
			Unique(),

		field.String("email").
			NotEmpty().
			Unique().
			MinLen(6).
			MaxLen(128),

		field.Bytes("password").
			NotEmpty().
			Sensitive().
			MaxLen(255).
			SchemaType(map[string]string{
				"mysql": "varbinary(255)",
			}),

		field.Bool("status").
			Default(false),

		field.Time("last_login_at").
			Optional().
			Nillable().
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Bool("is_super_admin").
			Default(false).
			Immutable(),
	}
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
		mixin.SoftDeleteMixin{},
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		// 一个用户可以有多个会话
		edge.To("sessions", Session.Type),

		// 一个用户可以有多个刷新令牌
		edge.To("refresh_tokens", RefreshToken.Type),
	}
}
