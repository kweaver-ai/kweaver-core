// Package interfaces 定义接口
// @file driveradapters.go
// @description: 定义驱动适配器接口
package interfaces

//go:generate mockgen -source=driveradapters.go -destination=../mocks/driveradapters.go -package=mocks
import "github.com/gin-gonic/gin"

// HTTPRouterInterface 路由公共接口
type HTTPRouterInterface interface {
	RegisterRouter(engine *gin.RouterGroup)
}

// MQHandler MQ处理接口
type MQHandler interface {
	Subscribe()
}
