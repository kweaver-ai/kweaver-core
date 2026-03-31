package chelper

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	//"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/types"
)

var CtxVistorInfoNotFound = errors.New("ctx_vistor_info not found")

func GetUserIDFromGinContext(c *gin.Context) (string, error) {
	ctxKey := cenum.VisitUserInfoCtxKey.String()

	if v, exists := c.Get(ctxKey); exists {
		if visitor, ok := v.(*rest.Visitor); ok {
			return visitor.ID, nil
		}

		return "", errors.New("invalid 'ctx_vistor_info' context value type: expected rest.Visitor")
	}

	return "", CtxVistorInfoNotFound
}

func GetVisitorFromCtx(c context.Context) (visitor *rest.Visitor) {
	ctxKey := cenum.VisitUserInfoCtxKey.String()

	vInter := c.Value(ctxKey)
	if vInter == nil {
		return
	}

	var ok bool
	if visitor, ok = vInter.(*rest.Visitor); !ok {
		panic("GetVisitorFromCtx:ctx.Value(enums.VisitUserInfoCtxKey) is not *rest.Visitor")
	}

	return
}

func GetUserIDFromCtx(ctx context.Context) (userID string) {
	visitor := GetVisitorFromCtx(ctx)
	if visitor == nil {
		return
	}

	return visitor.ID
}

func GetUserTokenFromCtx(ctx context.Context) (userToken string) {
	visitor := GetVisitorFromCtx(ctx)
	if visitor == nil {
		return
	}

	return visitor.TokenID
}
