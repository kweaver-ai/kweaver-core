// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package user_mgmt

import (
	"context"

	"vega-backend/common"
	"vega-backend/interfaces"
)

// NoopUserMgmtService 空用户管理服务（认证禁用时使用）
type NoopUserMgmtService struct {
	appSetting *common.AppSetting
}

func NewNoopUserMgmtService(appSetting *common.AppSetting) interfaces.UserMgmtService {
	return &NoopUserMgmtService{appSetting: appSetting}
}

func (n *NoopUserMgmtService) GetAccountNames(ctx context.Context, accountInfos []*interfaces.AccountInfo) error {
	// 认证禁用时，使用 ID 作为名称
	for _, info := range accountInfos {
		if info.Name == "" {
			info.Name = info.ID
		}
	}
	return nil
}
