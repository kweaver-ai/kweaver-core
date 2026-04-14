package iusermanagementacc

import "context"

type UserMgnt interface {
	GetUserInfoByUserID(ctx context.Context, usersID []string, fields []string) (userInfos map[string]*UserInfo, err error)
}

type UserInfo struct {
	Name  string   `json:"name"`
	ID    string   `json:"id"`
	Roles []string `json:"roles"`
}
