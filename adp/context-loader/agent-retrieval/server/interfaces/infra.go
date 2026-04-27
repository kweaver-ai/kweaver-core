// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"context"
	"net/url"
)

// ContextKey Context Value Key
//
//go:generate mockgen -source=infra.go -destination=../mocks/infra.go -package=mocks
type ContextKey string

const (
	// KeyToken token key in context
	KeyToken ContextKey = "token"
	// KeyIP ip key in context
	KeyIP ContextKey = "ip"
	// OperationID API operation unique identifier
	OperationID ContextKey = "operationID"
	// XLanguageKey Language type
	XLanguageKey ContextKey = "X-Language"
	// FileNameKey File name
	FileNameKey ContextKey = "FileName"
	// Headers header request parameters
	Headers ContextKey = "headers"
	// UserAgent user agent
	UserAgent ContextKey = "User-Agent"
	// KeyTokenInfo token info
	// KeyTokenInfo ContextKey = "Token-Info"
	// KeyUserID userID
	// KeyUserID ContextKey = "user_id"
	// IsPublic Whether public
	IsPublic ContextKey = "is_public"
	// // XUserID internal interface parameter XUserID
	// XUserID ContextKey = "x-user"
	// // XVisitorType internal interface parameter XVisitorType, indicates account type
	// XVisitorType ContextKey = "x-visitor-type"
	// // KeyVisitorType Account Type
	// KeyVisitorType ContextKey = "visitor_type"
	// KeyRequestID Request ID
	KeyRequestID ContextKey = "request_id"
	// KeyAccountAuthContext Account authentication context
	KeyAccountAuthContext ContextKey = "account_auth_context"
	// KeyResponseFormat 响应格式（json / toon），用于 HTTP 与 MCP 的序列化出口
	KeyResponseFormat ContextKey = "response_format"
)

// HeaderKey Context Key
type HeaderKey string

const (
	// HeaderXAccountID Account ID header parameter
	HeaderXAccountID HeaderKey = "x-account-id"
	// HeaderXAccountType Account Type header parameter
	HeaderXAccountType HeaderKey = "x-account-type"
	// HeaderUserID User ID
	HeaderUserID HeaderKey = "user_id"
)

const (
	HTTP  = "http"  // http protocol
	HTTPS = "https" // https protocol
)

// Logger Logger Interface
type Logger interface {
	Debug(v ...interface{})
	Info(v ...interface{})
	Warn(v ...interface{})
	Error(v ...interface{})
	Debugf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
	WithContext(ctx context.Context) Logger
}

// HTTPClient HTTP Client Service Interface
type HTTPClient interface {
	Get(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respData interface{}, err error)
	GetNoUnmarshal(ctx context.Context, url string, queryValues url.Values, headers map[string]string) (respCode int, respBody []byte, err error)
	Delete(ctx context.Context, url string, headers map[string]string) (respCode int, respData interface{}, err error)
	DeleteNoUnmarshal(ctx context.Context, url string, headers map[string]string) (respCode int, respBody []byte, err error)
	Post(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PostNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Put(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PutNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
	Patch(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respData interface{}, err error)
	PatchNoUnmarshal(ctx context.Context, url string, headers map[string]string, reqParam interface{}) (respCode int, respBody []byte, err error)
}
