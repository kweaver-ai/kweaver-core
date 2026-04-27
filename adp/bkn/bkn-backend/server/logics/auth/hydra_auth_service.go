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
	"bkn-backend/interfaces"
	"bkn-backend/logics"
)

type hydraAuthService struct {
	appSetting *common.AppSetting
	aa         interfaces.AuthAccess
}

func NewHydraAuthService(appSetting *common.AppSetting) interfaces.AuthService {
	return &hydraAuthService{
		appSetting: appSetting,
		aa:         logics.AA,
	}
}

func (s *hydraAuthService) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return s.aa.VerifyToken(ctx, c)
}
