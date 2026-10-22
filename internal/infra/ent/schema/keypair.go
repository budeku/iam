package schema

import (
	"budeku/iam/internal/enum"
	"budeku/iam/internal/infra/ent/mixin"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Keypair holds the schema definition for the Keypair entity.
type Keypair struct {
	ent.Schema
}

func (Keypair) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table:     "keypair",
			Charset:   "utf8mb4",
			Collation: "utf8mb4_0900_ai_ci",
		},
		schema.Comment("keypair"),
		entsql.WithComments(true),
	}
}

// Fields of the Keypair.
func (Keypair) Fields() []ent.Field {
	return []ent.Field{
		field.Int("app_id"),

		field.String("kid").
			NotEmpty().
			Unique().
			MaxLen(64),

		field.Uint8("algorithm").
			GoType(enum.KeypairAlgorithm(0)).
			Default(uint8(enum.KeypairAlgoRS256)),

		field.Text("public_key").
			Immutable().
			NotEmpty(),

		field.Text("private_key").
			Immutable().
			NotEmpty().
			Sensitive(),

		field.Time("activated_at").
			Default(time.Now).
			SchemaType(map[string]string{
				"mysql": "datetime(3)",
			}),

		field.Uint8("status").
			GoType(enum.KeypairStatus(0)).
			Default(uint8(enum.KeypairStatusActive)),

		field.Time("retire_at").
			Optional().
			Nillable(),
	}
}

func (Keypair) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DatetimeMixin{},
	}
}

// Edges of the Keypair.
func (Keypair) Edges() []ent.Edge {
	return nil
}
