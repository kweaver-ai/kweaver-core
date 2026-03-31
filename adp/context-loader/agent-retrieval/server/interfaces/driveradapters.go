// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

//go:generate mockgen -source=driveradapters.go -destination=../mocks/driveradapters.go -package=mocks
import "github.com/gin-gonic/gin"

// HTTPRouterInterface 路由公共接口
type HTTPRouterInterface interface {
	RegisterRouter(engine *gin.RouterGroup)
}
