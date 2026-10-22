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
		ID:              info.ID,
		Identity:        info.Identity,
		AppID:           info.AppID,
		UserID:          info.UserID,
		IP:              info.IP,
		UserAgent:       info.UserAgent,
		AuthMethod:      info.AuthMethod,
		AuthenticatedAt: info.AuthenticatedAt,
		LastActiveAt:    info.LastActiveAt,
		RevokedAt:       info.RevokedAt,
		ExpiresAt:       info.ExpiresAt,
		CreatedAt:       info.CreatedAt,
	}
}

func toRefreshTokenInfoResp(info *db.RefreshToken) *biz.RefreshTokenInfoResp {
	if info == nil {
		return nil
	}
	return &biz.RefreshTokenInfoResp{
		ID:        info.ID,
		AppID:     info.AppID,
		SessionID: info.SessionID,
		UserID:    info.UserID,
		Identity:  info.Identity,
		Jti:       info.Jti,
		RevokedAt: info.RevokedAt,
		ExpiresAt: info.ExpiresAt,
		CreatedAt: info.CreatedAt,
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

func toAppList(list []*db.App) []biz.AppListItemResp {
	if len(list) == 0 {
		return []biz.AppListItemResp{}
	}

	res := make([]biz.AppListItemResp, 0, len(list))
	for _, v := range list {
		res = append(res, biz.AppListItemResp{
			ID:        v.ID,
			Name:      v.Name,
			ClientID:  v.ClientID,
			Status:    v.Status,
			IsSystem:  v.IsSystem,
			CreatedAt: v.CreatedAt,
		})
	}

	return res
}

func toAppInfo(info *db.App) *biz.AppInfoResp {
	return &biz.AppInfoResp{
		ID:           info.ID,
		Name:         info.Name,
		ClientID:     info.ClientID,
		RedirectUris: info.RedirectUris,
		Status:       info.Status,
		IsSystem:     info.IsSystem,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}
}
