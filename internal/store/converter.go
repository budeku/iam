package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
)

func toUserGetResp(info *db.User) *biz.UserGetResp {
	return &biz.UserGetResp{
		ID:           info.ID,
		Username:     info.Username,
		UUID:         info.UUID,
		Email:        info.Email,
		Password:     info.Password,
		Status:       info.Status,
		LastLoginAt:  info.LastLoginAt,
		IsSuperAdmin: info.IsSuperAdmin,
	}
}

func toSessionCreateResp(info *db.Session) *biz.SessionCreateResp {
	return &biz.SessionCreateResp{
		ID: info.ID,
	}
}

func toRefreshTokenCreateResp(info *db.RefreshToken) *biz.RefreshTokenCreateResp {
	return &biz.RefreshTokenCreateResp{
		ID: info.ID,
	}
}

func toKeypairGetResp(info *db.Keypair) *biz.KeypairGetResp {
	return &biz.KeypairGetResp{
		ID:          info.ID,
		AppID:       info.AppID,
		Kid:         info.Kid,
		Algorithm:   info.Algorithm,
		PublicKey:   info.PublicKey,
		PrivateKey:  info.PrivateKey,
		ActivatedAt: info.ActivatedAt,
		Status:      info.Status,
		RetireAt:    info.RetireAt,
	}
}
