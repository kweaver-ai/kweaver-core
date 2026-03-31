package errors

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/i18n"
)

const PublicErrorType = "Public"

var CustomErrorType = "Custom"

type RestError struct {
	HTTPCode     int         `json:"-"`
	ErrorCode    string      `json:"code"`
	MainCode     string      `json:"-"`
	Description  string      `json:"description"`
	Solution     string      `json:"solution"`
	ErrorDetails interface{} `json:"detail,omitempty"`
	ErrorLink    string      `json:"link,omitempty"`
}

// ErrorParams 错误码构建参数
type ErrorParams struct {
	XLanguage string
	MainCode  string
	SubCode   string
	Key       string
	Details   interface{}
}

func (i *RestError) Error() string {
	errMap := map[string]interface{}{
		"description": i.Description,
		"detail":      i.ErrorDetails,
	}
	errMapBytes, _ := json.Marshal(errMap)
	return string(errMapBytes)
}

// InitServiceName 如果服务需要自定义错误码前缀，在服务初始化时设置自定义错误码第一段前缀
func InitServiceName(name string) {
	CustomErrorType = name
}

/*
NewPublicRestError 创建一个公共REST错误（默认HTTP状态码为0）
公共错误使用预定义的错误码，不包含服务特定前缀
参数:
  - ctx: 包含语言信息的上下文
  - mainCode: 主错误码（如"InvalidParameter"）
  - descriptionKey: 错误描述的翻译键
  - detail: 额外的错误详情（可为nil）
*/
func NewPublicRestError(ctx context.Context, mainCode, descriptionKey string, detail interface{}) *RestError {
	return newRestError(ctx, 0, PublicErrorType, mainCode, "", descriptionKey, detail)
}

/*
NewCustomRestError 创建一个自定义REST错误（默认HTTP状态码为0）
自定义错误包含服务特定前缀的错误码
参数:
  - ctx: 包含语言信息的上下文
  - mainCode: 主错误码（如"ValidationError"）
  - extCode: 扩展错误码（如"EmptyField"）
  - descriptionKey: 错误描述的翻译键
  - detail: 额外的错误详情（可为nil）
*/
func NewCustomRestError(ctx context.Context, mainCode, extCode, descriptionKey string, detail interface{}) *RestError {
	return newRestError(ctx, 0, CustomErrorType, mainCode, extCode, descriptionKey, detail)
}

/*
NewPublicRestErrorWithHTTPCode 创建一个带指定HTTP状态码的公共REST错误
公共错误使用预定义的错误码，不包含服务特定前缀
参数:
  - ctx: 包含语言信息的上下文
  - httpCode: HTTP状态码（如400, 404等）
  - mainCode: 主错误码（如"NotFound"）
  - descriptionKey: 错误描述的翻译键
  - detail: 额外的错误详情（可为nil）
*/
func NewPublicRestErrorWithHTTPCode(ctx context.Context, httpCode int, mainCode, descriptionKey string, detail interface{}) *RestError {
	return newRestError(ctx, httpCode, PublicErrorType, mainCode, "", descriptionKey, detail)
}

/*
NewCustomRestErrorWithHTTPCode 创建一个带指定HTTP状态码的自定义REST错误
自定义错误包含服务特定前缀的错误码
参数:
  - ctx: 包含语言信息的上下文
  - httpCode: HTTP状态码（如400, 404等）
  - mainCode: 主错误码（如"AuthError"）
  - extCode: 扩展错误码（如"InvalidToken"）
  - descriptionKey: 错误描述的翻译键
  - detail: 额外的错误详情（可为nil）
*/
func NewCustomRestErrorWithHTTPCode(ctx context.Context, httpCode int, mainCode, extCode, descriptionKey string, detail interface{}) *RestError {
	return newRestError(ctx, httpCode, CustomErrorType, mainCode, extCode, descriptionKey, detail)
}

func newRestError(ctx context.Context, httpCode int, codeType, mainCode, extCode, descriptionKey string, detail interface{}) *RestError {
	lang := i18n.GetLangFromCTX(ctx)
	re := NewRestErrorBuilder().
		WithDescription(descriptionKey, lang.String()).
		WithHTTPCode(httpCode).
		WithDetails(detail)

	switch codeType {
	case PublicErrorType:
		re.WithPublicCode(mainCode)
	default:
		re.WithCustomCode(mainCode, extCode)
	}

	return re.Build()
}

// NewRestError 创建一个Rest Error对象
func NewRestErrorBuilder() *RestError {
	return &RestError{}
}

// Build 必要参数回填
func (i *RestError) Build() *RestError {
	if i.Description == "" {
		i.Description = i.MainCode
	}
	return i
}

// WithHTTPCode 设置HTTP 请求响应状态码
func (i *RestError) WithHTTPCode(code int) *RestError {
	i.HTTPCode = code
	return i
}

// WithPublicCode 设置公共错误码
func (i *RestError) WithPublicCode(code string) *RestError {
	if len(code) == 0 {
		code = "InternalError"
	}

	i.ErrorCode = fmt.Sprintf("%s.%s", PublicErrorType, code)
	i.MainCode = code
	return i
}

// WithCustomCode 设置自定义错误码
func (i *RestError) WithCustomCode(codes ...string) *RestError {
	if len(codes) == 0 {
		codes = append(codes, "InternalError")
	}

	// 错误码不能超过三段,超过三段式的部分丢弃
	if len(codes) > 2 {
		codes = codes[0:2]
	}

	i.ErrorCode = fmt.Sprintf("%s.%s", CustomErrorType, strings.Join(codes, "."))
	i.MainCode = codes[0]
	return i
}

// WithDescription 设置国际化资源描述
func (i *RestError) WithDescription(key, lang string) *RestError {
	l := i18n.Language(lang)
	langPtr := l.GetLangType()
	if langPtr == nil {
		lang = string(i18n.SimplifiedChinese)
	} else {
		lang = langPtr.String()
	}
	translator := i18n.NewI18nTranslator(lang)
	i.Description = translator.Trans("desc", key)
	i.Solution = translator.Trans("sol", key)
	return i
}

func (i *RestError) WithDetails(details interface{}) *RestError {
	i.ErrorDetails = details
	return i
}

func Is(err error, codes ...string) bool {
	rErr, ok := err.(*RestError)
	if !ok {
		return false
	}
	errCode := strings.Join(codes, ".")

	return rErr.ErrorCode == errCode
}
