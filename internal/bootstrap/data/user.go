package data

import (
	"budeku/iam/internal/infra/ent/db"
	"budeku/iam/internal/store"
	"budeku/iam/pkg/password"
	"budeku/iam/pkg/random"
	"context"
)

func initUser(ctx context.Context, userStore *store.UserStore, result *Result, opts *Options) (*db.User, error) {
	if u, err := userStore.GetSuperAdmin(ctx); err == nil {
		result.Username = u.Username
		result.Email = u.Email
		return u, nil
	} else if !db.IsNotFound(err) {
		return nil, err
	}

	plaintext, err := random.Unambiguous(12)
	if err != nil {
		return nil, err
	}
	passwordHash, err := password.Hash(plaintext)
	if err != nil {
		return nil, err
	}

	u, err := userStore.InitSuperAdmin(ctx, opts.Username, opts.Email, passwordHash)
	if err != nil {
		return nil, err
	}

	result.UserCreated = true
	result.Username = u.Username
	result.Email = u.Email
	result.Password = plaintext

	return u, nil
}
