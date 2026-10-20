package data

import (
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/store"
	"context"
)

type Result struct {
	AppCreated  bool
	AppClientID string

	KeypairCreated bool
	KeypairKid     string

	UserCreated bool
	Username    string
	Email       string
	Password    string
}

type Options struct {
	Username string
	Email    string
}

func Init(ctx context.Context, client *db.Client, opts *Options) (*Result, error) {
	result := &Result{}

	kpStore := store.NewKeypairStore(client)
	if err := initKeypair(ctx, kpStore, result); err != nil {
		return result, err
	}

	userStore := store.NewUserStore(client)
	if _, err := initUser(ctx, userStore, result, opts); err != nil {
		return result, err
	}

	return result, nil
}
