// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package business_system

import (
	"context"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

// NoopBusinessSystemService 空业务域服务（业务域禁用时使用）
type NoopBusinessSystemService struct {
	appSetting *common.AppSetting
}

func NewNoopBusinessSystemService(appSetting *common.AppSetting) interfaces.BusinessSystemService {
	return &NoopBusinessSystemService{appSetting: appSetting}
}

func (n *NoopBusinessSystemService) BindResource(ctx context.Context, bd_id string, rid string, rtype string) error {
	return nil // 静默跳过
}

func (n *NoopBusinessSystemService) UnbindResource(ctx context.Context, bd_id string, rid string, rtype string) error {
	return nil // 静默跳过
}
