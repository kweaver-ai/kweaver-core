package chelper

import (
	"context"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
)

func IsInternalAPIFromCtx(c context.Context) (isInternalAPI bool) {
	ctxKey := cenum.InternalAPIFlagCtxKey.String()

	vInter := c.Value(ctxKey)
	if vInter == nil {
		return
	}

	var ok bool
	if isInternalAPI, ok = vInter.(bool); !ok {
		panic("IsInternalAPIFromCtx:ctx.Value(enums.InternalAPIFlagCtxKey) is not bool")
	}

	return
}
