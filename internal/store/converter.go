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

func toSessionGetResp(info *db.Session) *biz.SessionGetResp {
	if info == nil {
		return nil
	}

	return &biz.SessionGetResp{
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

func toRefreshTokenGetResp(info *db.RefreshToken) *biz.RefreshTokenGetResp {
	if info == nil {
		return nil
	}
	return &biz.RefreshTokenGetResp{
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
		CreatedAt:   info.CreatedAt,
		UpdatedAt:   info.UpdatedAt,
	}
}

func toKeypairGetAllResp(list []*db.Keypair) []biz.KeypairGetResp {
	if len(list) == 0 {
		return []biz.KeypairGetResp{}
	}

	res := make([]biz.KeypairGetResp, 0, len(list))
	for _, v := range list {
		res = append(res, biz.KeypairGetResp{
			ID:          v.ID,
			AppID:       v.AppID,
			Kid:         v.Kid,
			Algorithm:   v.Algorithm,
			PublicKey:   v.PublicKey,
			PrivateKey:  v.PrivateKey,
			ActivatedAt: v.ActivatedAt,
			Status:      v.Status,
			RetireAt:    v.RetireAt,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		})
	}

	return res
}

func toKeypairList(list []*db.Keypair) []biz.KeypairListItemResp {
	if len(list) == 0 {
		return []biz.KeypairListItemResp{}
	}

	res := make([]biz.KeypairListItemResp, 0, len(list))
	for _, v := range list {
		res = append(res, biz.KeypairListItemResp{
			ID:          v.ID,
			AppID:       v.AppID,
			Kid:         v.Kid,
			Algorithm:   v.Algorithm,
			Status:      v.Status,
			ActivatedAt: v.ActivatedAt,
			CreatedAt:   v.CreatedAt,
		})
	}

	return res
}

func toKeypairCreate(info *db.Keypair) *biz.KeypairCreateResp {
	if info == nil {
		return nil
	}

	return &biz.KeypairCreateResp{
		ID: info.ID,
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
			CreatedAt: v.CreatedAt,
		})
	}

	return res
}

func toAppInfo(info *db.App) *biz.AppGetResp {
	if info == nil {
		return nil
	}

	return &biz.AppGetResp{
		ID:           info.ID,
		Name:         info.Name,
		ClientID:     info.ClientID,
		RedirectUris: info.RedirectUris,
		Status:       info.Status,
		CreatedAt:    info.CreatedAt,
		UpdatedAt:    info.UpdatedAt,
	}
}
