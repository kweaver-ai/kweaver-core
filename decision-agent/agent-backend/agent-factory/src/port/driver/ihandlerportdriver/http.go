package ihandlerportdriver

import "github.com/gin-gonic/gin"

type IHTTPRouter interface {
	RegPubRouter(engine *gin.RouterGroup)
	RegPriRouter(engine *gin.RouterGroup)
}
