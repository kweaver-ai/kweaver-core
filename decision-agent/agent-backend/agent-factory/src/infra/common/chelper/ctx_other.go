package chelper

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	//"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/types"
)

func GetTraceIDFromCtx(ctx context.Context) (traceID string) {
	vInter := ctx.Value(cenum.TraceIDCtxKey.String())
	if vInter == nil {
		return
	}

	if v, ok := vInter.(string); ok {
		traceID = v
	} else {
		panic("GetTraceIDFromCtx:ctx.Value(enums.TraceIDCtxKey) is not string")
	}

	return
}

// 获取访问语言
func GetVisitLanguageCtx(c context.Context) (language rest.Language) {
	ctxKey := cenum.VisitLangCtxKey.String()

	ctxVal := c.Value(ctxKey)
	if ctxVal == nil {
		language = cglobal.GConfig.GetDefaultLanguage()
		return
	}

	if v, ok := ctxVal.(rest.Language); !ok {
		panic("GetVisitLanguageFromCtx:ctx.Value(enums.VisitLangCtxKey) is not rest.Language")
	} else {
		language = v
	}

	return
}
