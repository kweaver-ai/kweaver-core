package common

import (
	"context"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
)

// GetLanguageFromCtx 从context中获取语言设置
func GetLanguageFromCtx(ctx context.Context) Language {
	return GetLanguageByCtx(ctx)
}

// IsPublicAPIFromCtx 判断是否为公开API
func IsPublicAPIFromCtx(ctx context.Context) bool {
	if isPublic, ok := ctx.Value(interfaces.IsPublic).(bool); ok {
		return isPublic
	}
	return false
}

func SetPublicAPIToCtx(ctx context.Context, isPublic bool) context.Context {
	return context.WithValue(ctx, interfaces.IsPublic, isPublic)
}

// SetLanguageToCtx 设置语言到context
func SetLanguageToCtx(ctx context.Context, languageInfo Language) context.Context {
	return SetLanguageByCtx(ctx, languageInfo)
}

// SetAccountAuthContextToCtx 设置账户认证上下文到context
func SetAccountAuthContextToCtx(ctx context.Context, authContext *interfaces.AccountAuthContext) context.Context {
	return context.WithValue(ctx, interfaces.KeyAccountAuthContext, authContext)
}

func GetAccountAuthContextFromCtx(ctx context.Context) (*interfaces.AccountAuthContext, bool) {
	authContext, ok := ctx.Value(interfaces.KeyAccountAuthContext).(*interfaces.AccountAuthContext)
	return authContext, ok
}

// GetTokenInfoFromCtx 从context中获取token信息
// func GetTokenInfoFromCtx(ctx context.Context) (*interfaces.TokenInfo, bool) {
// 	authContext, ok := GetAccountAuthContextFromCtx(ctx)
// 	if !ok {
// 		return nil, false
// 	}
// 	if authContext.TokenInfo == nil {
// 		return nil, false
// 	}
// 	return authContext.TokenInfo, true
// }

// SetExecutionModeToCtx 设置执行模式到context
func SetExecutionModeToCtx(ctx context.Context, executionMode interfaces.ExecutionMode) context.Context {
	return context.WithValue(ctx, interfaces.KeyExecutionMode, executionMode)
}

// GetExecutionModeFromCtx 从context中获取执行模式
func GetExecutionModeFromCtx(ctx context.Context) interfaces.ExecutionMode {
	executionMode, ok := ctx.Value(interfaces.KeyExecutionMode).(interfaces.ExecutionMode)
	if !ok || executionMode == "" {
		executionMode = interfaces.ExecutionModeSync
	}
	return executionMode
}

// GetStreamingModeFromCtx 从context中获取流式模式
func GetStreamingModeFromCtx(ctx context.Context) (interfaces.StreamingMode, bool) {
	streamingMode, ok := ctx.Value(interfaces.KeyStreamingMode).(interfaces.StreamingMode)
	return streamingMode, ok
}

// SetStreamingModeToCtx 设置流式模式到context
func SetStreamingModeToCtx(ctx context.Context, streamingMode interfaces.StreamingMode) context.Context {
	return context.WithValue(ctx, interfaces.KeyStreamingMode, streamingMode)
}

// SetResponseWriterToCtx 设置response writer到context
func SetResponseWriterToCtx(ctx context.Context, responseWriter http.ResponseWriter) context.Context {
	return context.WithValue(ctx, interfaces.KeyResponseWriter, responseWriter)
}

// GetResponseWriterFromCtx 从context中获取response writer
func GetResponseWriterFromCtx(ctx context.Context) (http.ResponseWriter, bool) {
	responseWriter, ok := ctx.Value(interfaces.KeyResponseWriter).(http.ResponseWriter)
	return responseWriter, ok
}

// GetHeaderFromCtx 请求外部接口时，从context中获取Header参数传递
func GetHeaderFromCtx(ctx context.Context) (header map[string]string) {
	header = map[string]string{}
	authContext, ok := GetAccountAuthContextFromCtx(ctx)
	if !ok {
		return
	}
	header[string(interfaces.HeaderXAccountID)] = authContext.AccountID
	header[string(interfaces.HeaderXAccountType)] = string(authContext.AccountType)
	return
}

// SetBusinessDomainToCtx 设置业务域id到context
func SetBusinessDomainToCtx(ctx context.Context, businessDomain string) context.Context {
	return context.WithValue(ctx, interfaces.XBusinessDomain, businessDomain)
}

// GetBusinessDomainFromCtx 从context中获取业务域id
func GetBusinessDomainFromCtx(ctx context.Context) (string, bool) {
	businessDomain, ok := ctx.Value(interfaces.XBusinessDomain).(string)
	return businessDomain, ok
}
