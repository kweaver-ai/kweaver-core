package ihandlerportdriver

import "github.com/gin-gonic/gin"

type IHTTPHealthRouter interface {
	RegHealthRouter(engine *gin.RouterGroup)
}
