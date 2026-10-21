package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
)

var _ biz.AppStore = (*AppStore)(nil)

type AppStore struct {
	*baseStore
}

func NewAppStore(client *db.Client) *AppStore {
	return &AppStore{newBaseStore(client)}
}
