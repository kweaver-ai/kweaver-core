// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package auth

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"

	"ontology-query/common"
	"ontology-query/interfaces"
)

type hydraAuthAccess struct {
	hydra hydra.Hydra
}

func NewHydraAuthAccess(appSetting *common.AppSetting) interfaces.AuthAccess {
	return &hydraAuthAccess{
		hydra: hydra.NewHydra(appSetting.HydraAdminSetting),
	}
}

func (h *hydraAuthAccess) VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	return h.hydra.VerifyToken(ctx, c)
}
