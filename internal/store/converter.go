package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/infra/ent/db"
)

func toUserGetResp(info *db.User) *biz.UserGetResp {
	if info == nil {
		return nil
	}

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
	if info == nil {
		return nil
	}

	return &biz.SessionCreateResp{
		ID: info.ID,
	}
}

func toSessionInfoResp(info *db.Session) *biz.SessionInfoResp {
	if info == nil {
		return nil
	}

	return &biz.SessionInfoResp{
		ID:       info.ID,
		Identity: info.Identity,
	}
}

func toRefreshTokenCreateResp(info *db.RefreshToken) *biz.RefreshTokenCreateResp {
	if info == nil {
		return nil
	}

	return &biz.RefreshTokenCreateResp{
		ID: info.ID,
	}
}

func toKeypairGetResp(info *db.Keypair) *biz.KeypairGetResp {
	if info == nil {
		return nil
	}

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
