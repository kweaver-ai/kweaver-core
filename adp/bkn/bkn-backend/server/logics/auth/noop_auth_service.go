// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	"bkn-backend/interfaces"
)

// NoopAuthService 空认证服务（认证禁用时使用）
type NoopAuthService struct {
	appSetting *common.AppSetting
}

func NewNoopAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &NoopAuthService{
		appSetting: appSetting,
	}
}

func (n *NoopAuthService) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	// 返回空 Visitor，不做任何认证校验
	return visitor.GenerateVisitor(c), nil
}
