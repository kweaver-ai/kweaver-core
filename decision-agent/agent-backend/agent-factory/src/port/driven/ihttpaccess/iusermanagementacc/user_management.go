package iusermanagementacc

import "context"

type UserMgnt interface {
	GetUserInfoByUserID(ctx context.Context, usersID []string, fields []string) (userInfos map[string]*UserInfo, err error)
	GetUserIDByAccount(ctx context.Context, account string) (invalid bool, userID string, err error)
	IsSuperAdmin(ctx context.Context, userID string) (isSuperAdmin bool, err error)
}

type UserInfo struct {
	Name  string   `json:"name"`
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
}
