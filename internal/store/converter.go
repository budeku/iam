package store

import (
	"budeku/iam/internal/biz"
	"budeku/iam/internal/conf"
	"budeku/iam/internal/infra/ent/db"

	"github.com/google/uuid"
)

func toUserGet(info *db.User) *biz.UserGetResp {
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

func toUserCreate(info *db.User) *biz.UserCreateResp {
	if info == nil {
		return nil
	}

	return &biz.UserCreateResp{
		ID: info.ID,
	}
}

func toUserList(list []*db.User) []biz.UserListItemResp {
	if len(list) == 0 {
		return []biz.UserListItemResp{}
	}

	res := make([]biz.UserListItemResp, 0, len(list))
	for _, v := range list {
		item := biz.UserListItemResp{
			ID:           v.ID,
			Username:     v.Username,
			Email:        v.Email,
			UUID:         v.UUID,
			Status:       v.Status,
			LastLoginAt:  v.LastLoginAt,
			IsSuperAdmin: v.IsSuperAdmin,
		}

		res = append(res, item)
	}

	return res
}

func toUserSessions(list []*db.Session, currentSessionID uuid.UUID, appNameMap map[int]string) []biz.UserSessionResp {
	if len(list) == 0 {
		return nil
	}

	res := make([]biz.UserSessionResp, 0, len(list))
	for _, v := range list {
		var appName string
		if name, ok := appNameMap[v.AppID]; ok {
			appName = name
		} else if v.AppID == conf.IAMAppID {
			appName = conf.IAMAppName
		}

		item := biz.UserSessionResp{
			ID:              v.ID,
			Identity:        v.Identity,
			AppID:           v.AppID,
			AppName:         appName,
			IP:              v.IP,
			UserAgent:       v.UserAgent,
			AuthMethod:      v.AuthMethod,
			AuthenticatedAt: v.AuthenticatedAt,
			LastActiveAt:    v.LastActiveAt,
			ExpiresAt:       v.ExpiresAt,
			IsCurrent:       v.Identity == currentSessionID,
		}

		if tokens, err := v.Edges.RefreshTokensOrErr(); err == nil && len(tokens) > 0 {
			item.RefreshTokens = make([]biz.UserSessionRefreshTokenResp, len(tokens))
			for i, t := range tokens {
				item.RefreshTokens[i] = biz.UserSessionRefreshTokenResp{
					ID:        t.ID,
					Jti:       t.Jti,
					ExpiresAt: t.ExpiresAt,
				}
			}
		} else {
			item.RefreshTokens = []biz.UserSessionRefreshTokenResp{}
		}

		res = append(res, item)
	}

	return res
}

func toSessionCreate(info *db.Session) *biz.SessionCreateResp {
	if info == nil {
		return nil
	}

	return &biz.SessionCreateResp{
		ID: info.ID,
	}
}

func toSessionGet(info *db.Session) *biz.SessionGetResp {
	if info == nil {
		return nil
	}

	resp := &biz.SessionGetResp{
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

	// 查询时没有加 WithUser，这里会是 nil
	if info.Edges.User != nil {
		resp.User = toUserGet(info.Edges.User)
	}

	// 询时没有加 WithRefreshTokens，这里长度为 0
	if len(info.Edges.RefreshTokens) > 0 {
		tokens := make([]biz.RefreshTokenGetResp, len(info.Edges.RefreshTokens))
		for i, t := range info.Edges.RefreshTokens {
			tokens[i] = *toRefreshTokenGet(t)
		}
		resp.RefreshTokens = tokens
	}

	return resp
}

func toRefreshTokenGet(info *db.RefreshToken) *biz.RefreshTokenGetResp {
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

func toRefreshTokenCreate(info *db.RefreshToken) *biz.RefreshTokenCreateResp {
	if info == nil {
		return nil
	}

	return &biz.RefreshTokenCreateResp{
		ID: info.ID,
	}
}

func toKeypairGet(info *db.Keypair) *biz.KeypairGetResp {
	if info == nil {
		return nil
	}

	resp := &biz.KeypairGetResp{
		ID:          info.ID,
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

	// 查询时加了 WithApp()
	if info.Edges.App != nil {
		resp.AppID = info.Edges.App.ID
		resp.AppName = info.Edges.App.Name
	}

	return resp
}

func toKeypairGetAll(list []*db.Keypair) []biz.KeypairGetResp {
	if len(list) == 0 {
		return []biz.KeypairGetResp{}
	}

	res := make([]biz.KeypairGetResp, 0, len(list))
	for _, v := range list {
		item := biz.KeypairGetResp{
			ID:          v.ID,
			Kid:         v.Kid,
			Algorithm:   v.Algorithm,
			PublicKey:   v.PublicKey,
			PrivateKey:  v.PrivateKey,
			ActivatedAt: v.ActivatedAt,
			Status:      v.Status,
			RetireAt:    v.RetireAt,
			CreatedAt:   v.CreatedAt,
			UpdatedAt:   v.UpdatedAt,
		}
		if v.Edges.App != nil {
			item.AppID = v.Edges.App.ID
			item.AppName = v.Edges.App.Name
		}
		res = append(res, item)
	}

	return res
}

func toKeypairList(list []*db.Keypair) []biz.KeypairListItemResp {
	if len(list) == 0 {
		return []biz.KeypairListItemResp{}
	}

	res := make([]biz.KeypairListItemResp, 0, len(list))
	for _, v := range list {
		item := biz.KeypairListItemResp{
			ID:          v.ID,
			Kid:         v.Kid,
			Algorithm:   v.Algorithm,
			Status:      v.Status,
			ActivatedAt: v.ActivatedAt,
		}

		if v.Edges.App != nil {
			item.AppID = v.Edges.App.ID
			item.AppName = v.Edges.App.Name
		}

		res = append(res, item)
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

func toAppGet(info *db.App) *biz.AppGetResp {
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

func toAppOptions(list []*db.App) []biz.AppOptionItemResp {
	if len(list) == 0 {
		return []biz.AppOptionItemResp{}
	}

	res := make([]biz.AppOptionItemResp, 0, len(list))
	for _, v := range list {
		res = append(res, biz.AppOptionItemResp{
			ID:   v.ID,
			Name: v.Name,
		})
	}

	return res
}
