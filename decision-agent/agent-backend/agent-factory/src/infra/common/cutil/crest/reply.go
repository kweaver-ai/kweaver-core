package crest

import (
	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
)

// ReplyError2 回复错误
// 【注意】： 如果err是rest.HTTPError类型，会直接回复该错误。此时可能会丢失错误信息（比如errors.Wrap的错误信息）。
func ReplyError2(c *gin.Context, err error) {
	httpErr, ok := GetRestHttpErr(err)
	if ok {
		err = httpErr
	}

	rest.ReplyError(c, err)
}
