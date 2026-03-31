package chelper

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
)

// GetBizDomainIDFromGinHeader 从Gin请求头中获取业务域ID
func GetBizDomainIDFromGinHeader(c *gin.Context) (bizDomainID string, isExist bool, err error) {
	if c == nil {
		err = errors.New("c is nil")
		return
	}

	// 1. 从header中获取bizDomainID
	bizDomainID = c.GetHeader(cenum.HeaderXBizDomainID.String())

	// 2. 如果bizDomainID为空，返回
	if bizDomainID == "" {
		return
	}

	// 3. 如果bizDomainID不为空，设置isExist为true
	isExist = true

	return
}

func GetBizDomainIDFromCtx(c context.Context) (bizDomainID string) {
	ctxKey := cenum.BizDomainIDCtxKey.String()

	vInter := c.Value(ctxKey)
	if vInter == nil {
		return
	}

	var ok bool
	if bizDomainID, ok = vInter.(string); !ok {
		panic("GetBizDomainIDFromCtx:ctx.Value(enums.BizDomainIDCtxKey) is not string")
	}

	return
}
