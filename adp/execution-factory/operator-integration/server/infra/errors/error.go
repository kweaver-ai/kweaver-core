// Package errors 定义错误码
// @file errors.go
// @description: 错误码统一处理
package errors

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/common"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/localize"
	jsoniter "github.com/json-iterator/go"
)

// HTTPError HTTP错误
type HTTPError struct {
	HTTPCode     int         `json:"-"`
	Language     string      `json:"-"`
	Code         string      `json:"code,omitempty"`
	Description  string      `json:"description,omitempty"` // 错误描述
	Solution     string      `json:"solution,omitempty"`    // 解决方法
	ErrorLink    string      `json:"link,omitempty"`        // 错误链接
	ErrorDetails interface{} `json:"details,omitempty"`     // 详细内容
	// DescriptionTemplateData map[string]any `json:"-"`                     // 错误描述参数
	// SolutionTemplateData    map[string]any `json:"-"`                     // 解决方法参数
}

func (e *HTTPError) WithDescription(extCode ErrorCode, params ...interface{}) *HTTPError {
	if e.HTTPCode == 0 {
		e.HTTPCode = http.StatusInternalServerError
	}
	if e.Language == "" {
		e.Language = "zh_CN"
	}
	if e.Code == "" {
		e.Code = fmt.Sprintf("Public.%d.%s", e.HTTPCode, extCode)
	}
	tr := localize.NewI18nTranslator(e.Language)
	e.Description = fmt.Sprintf(tr.Trans("desc."+extCode.String()), params...)
	return e
}

// Error 返回错误信息
func (e *HTTPError) Error() string {
	errBys, _ := jsoniter.Marshal(e)
	return string(errBys)
}

var (
	errServerName = "AgentOperatorIntegration"

	errCodeMap = map[int]string{
		http.StatusBadRequest:          "BadRequest",
		http.StatusUnauthorized:        "Unauthorized",
		http.StatusForbidden:           "Forbidden",
		http.StatusNotFound:            "NotFound",
		http.StatusMethodNotAllowed:    "MethodNotAllowed",
		http.StatusConflict:            "Conflict",
		http.StatusInternalServerError: "InternalServerError",
		http.StatusNotImplemented:      "NotImplemented",
		http.StatusServiceUnavailable:  "ServiceUnavailable",
		http.StatusRequestTimeout:      "RequestTimeout",
	}
)

// DefaultHTTPError 公共错误码
func DefaultHTTPError(ctx context.Context, httpCode int, details interface{}) *HTTPError {
	language := common.GetLanguageFromCtx(ctx)
	tr := localize.NewI18nTranslator(language)
	errCode := errCodeMap[httpCode]
	if errCode == "" {
		errCode = errCodeMap[http.StatusInternalServerError]
	}
	// 获取带默认值的解决方案和错误链接
	solutionKey := "sol." + errCode
	solution := tr.Trans(solutionKey)
	if solution == solutionKey { // 没有找到对应翻译时回退通用方案
		solution = tr.Trans("sol.Common")
	}

	errorLinkKey := "link." + errCode
	errorLink := tr.Trans(errorLinkKey)
	if errorLink == errorLinkKey { // 没有找到对应翻译时返回"无"
		errorLink = tr.Trans("link.None")
	}

	return &HTTPError{
		HTTPCode:     httpCode,
		Language:     language,
		Code:         "Public." + errCode,
		Description:  tr.Trans("desc." + errCode),
		Solution:     solution,
		ErrorLink:    errorLink,
		ErrorDetails: details,
	}
}

// NewHTTPError 创建 HTTPError @extCode: 拓展错误码
func NewHTTPError(ctx context.Context, httpCode int, extCode ErrorCode, details interface{}, descParams ...interface{}) *HTTPError {
	language := common.GetLanguageFromCtx(ctx)
	tr := localize.NewI18nTranslator(language)
	errCode := errCodeMap[httpCode]
	if errCode == "" {
		errCode = errCodeMap[http.StatusInternalServerError]
	}
	// 获取带默认值的解决方案和错误链接
	solutionKey := "sol." + extCode.String()
	solution := tr.Trans(solutionKey)
	if solution == solutionKey { // 没有找到对应翻译时回退通用方案
		solution = tr.Trans("sol.Common")
	}

	errorLinkKey := "link." + extCode.String()
	errorLink := tr.Trans(errorLinkKey)
	if errorLink == errorLinkKey { // 没有找到对应翻译时返回"无"
		errorLink = tr.Trans("link.None")
	}
	return &HTTPError{
		HTTPCode:     httpCode,
		Language:     language,
		Code:         fmt.Sprintf("%s.%s.%s", errServerName, errCode, extCode),
		Description:  fmt.Sprintf(tr.Trans("desc."+extCode.String()), descParams...), // 可以处理%d、%s、%v等格式化占位符。
		Solution:     solution,
		ErrorLink:    errorLink,
		ErrorDetails: details,
	}
}
