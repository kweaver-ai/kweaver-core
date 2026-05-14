// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	haAccessOnce sync.Once
	haAccess     interfaces.AuthAccess
)

type hydraAuthAccess struct {
	appSetting *common.AppSetting
	hydra      hydra.Hydra
}

func NewHydraAuthAccess(appSetting *common.AppSetting) interfaces.AuthAccess {
	haAccessOnce.Do(func() {
		haAccess = &hydraAuthAccess{
			appSetting: appSetting,
			hydra:      hydra.NewHydra(appSetting.HydraAdminSetting),
		}
	})

	return haAccess
}

func (haa *hydraAuthAccess) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return haa.hydra.VerifyToken(ctx, c)
}
