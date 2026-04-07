// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
)

//go:generate mockgen -source ../interfaces/auth_service.go -destination ../interfaces/mock/mock_auth_service.go
type AuthService interface {
	VerifyToken(ctx context.Context, c *gin.Context) (hydra.Visitor, error)
}
