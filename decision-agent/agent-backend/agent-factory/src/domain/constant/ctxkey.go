package constant

import "context"

type ContextKey string

const (
	IncStreamKey ContextKey = "inc_stream"
	AppKey       ContextKey = "app_key"
	// CtxKeyIsShowOriginResponse 是否展示来自agent-executor的原始response
	CtxKeyIsShowOriginResponse ContextKey = "is_show_origin_response"
)

func IsShowOriginResponseFromCtx(ctx context.Context) bool {
	return ctx.Value(CtxKeyIsShowOriginResponse) == "true"
}
