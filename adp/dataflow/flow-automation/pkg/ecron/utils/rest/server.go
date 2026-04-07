package rest

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"

	errorv2 "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils/rest/error"
)

const strErrorCode = "string"

// TimeStampToString 纳秒时间戳转RFC3339格式的字符串
func TimeStampToString(t int64) string {
	const num int64 = 1e9
	return time.Unix(t/num, 0).Format(time.RFC3339)
}

// StringToTimeStamp RFC3339格式的字符串转纳秒时间戳
func StringToTimeStamp(t string) (int64, error) {
	tt, err := time.Parse(time.RFC3339, t)
	if err != nil {
		return 0, err
	}
	return tt.UnixNano(), nil
}

// GetJSONValue 读取请求body
func GetJSONValue(c *gin.Context, v interface{}) error {
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	err = jsoniter.Unmarshal(body, v)
	if err != nil {
		return NewHTTPError(err.Error(), BadRequest, nil)
	}
	return nil
}

// ReplyOK 响应成功
func ReplyOK(c *gin.Context, statusCode int, body interface{}) {
	b := make([]byte, 0)
	if body != nil {
		b, _ = jsoniter.Marshal(body)
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, string(b))
}

// ReplyError 响应错误
func ReplyError(c *gin.Context, err error) {
	var statusCode int
	var body string
	errCodeType := c.GetHeader("x-error-code")
	errCodeType = strings.ToLower(errCodeType)
	switch e := err.(type) {
	case *HTTPError:
		if errCodeType == strErrorCode {
			e.useCodeStr = true
		}
		statusCode = e.Code / 1e6
		body = e.Error()
	case ExHTTPError:
		statusCode = e.Status
		body = e.Error()
	case *errorv2.Error:
		strCode := strings.Split(e.Code, ".")[1]
		statusCode = errorv2.StrCodeToStatusCode(strCode)
		body = e.Error()
	default:
		tmpErr := NewHTTPError(e.Error(), InternalServerError, nil)
		if errCodeType == strErrorCode {
			tmpErr.useCodeStr = true
		}
		statusCode = http.StatusInternalServerError
		body = tmpErr.Error()
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, body)
}

// ReplyErrorV2 返回错误信息，code 是 string 类型
//
// 使用新错误码（string 类型的错误码）之后，服务间的接口调用存在以下四种情况：
//
// 1. 新接口 调用 新接口：
// 新接口中一律使用 errorv2.NewError() 创建新的错误信息实例，且使用 ReplyErrorV2 返回错误信息，不存在错误码类型的转换。
//
// 2. 新接口 调用 旧接口：
// 新接口在调用旧接口时，如果在请求头中显式地设置 x-error-code='string'，那么旧接口使用 ReplyError 返回错误信息时，会根据 x-error-code 将 int code 转换成 string code。
// 否则，旧接口返回的是 int 类型的错误码，即 HTTPError。那么新接口在调用 ReplyErrorV2 返回错误信息时，会执行 case *HTTPError，进行转换。
// 无论是旧接口在调用 ReplyError 时发生的int code --> string code 转换，还是新接口在调用 ReplyErrorV2 时发生的转换，转换后的 string 类型错误码都是以 Public 开头的通用错误码。
//
// 3. 旧接口 调用 旧接口：
// 旧接口默认情况下返回的是 int 类型的错误码。
// 如果请求头中 x-error-code='string'，那么旧接口在使用 ReplyError 返回错误信息时，会将 int code 转换成 string code。注意，这里的 string code 是以 Public 开头的通用错误码！！！。具体转换逻辑见 HTTPError.Error()。
// 如果需要返回特定的错误码，首先应考虑 description 信息是否能够满足需求。
// 如果 description 信息无法满足需求，确实需要 string 类型的具体错误码，那么被调用方可以在逻辑层创建 HTTPError 实例时，使用 SetCodeStr 方法设置具体错误码。
// 具体错误码的定义规范见：https://confluence.aishu.cn/pages/viewpage.action?pageId=62698513 （4xx、5xx系列状态码）
//
// 4. 旧接口 调用 新接口：
// 基本不会出现这种情况。如果真的出现了这种情况，那么实现新的逻辑，将该情况转换成情况 1 或 2。
func ReplyErrorV2(c *gin.Context, err error) {
	var statusCode int
	var body string
	switch e := err.(type) {
	case *HTTPError:
		statusCode = e.Code / 1e6
		e.useCodeStr = true
		body = e.Error()
	case *errorv2.Error:
		strCode := strings.Split(e.Code, ".")[1]
		statusCode = errorv2.StrCodeToStatusCode(strCode)
		body = e.Error()
	default:
		statusCode = http.StatusInternalServerError
		tmpErr := errorv2.NewError(errorv2.PublicInternalServerError, e.Error())
		body = tmpErr.Error()
	}

	c.Writer.Header().Set("Content-Type", "application/json")
	c.String(statusCode, body)
}
