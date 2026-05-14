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
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"go.opentelemetry.io/otel/codes"

	"bkn-backend/common"
	"bkn-backend/interfaces"
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
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "VerifyToken")
	defer span.End()

	visitor, err := haa.hydra.VerifyToken(ctx, c)
	if err != nil {
		otellog.LogError(ctx, "Verify token failed", err)
		return visitor, err
	}

	span.SetStatus(codes.Ok, "")
	return visitor, nil
}
