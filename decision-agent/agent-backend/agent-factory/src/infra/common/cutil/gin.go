package cutil

import (
	"context"

	"github.com/gin-gonic/gin"
)

func UpdateGinReqCtx(c *gin.Context, ctx context.Context) {
	req := c.Request.WithContext(ctx)
	c.Request = req
}
