package error

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	jsoniter "github.com/json-iterator/go"
)

// 错误码格式：
//   - 公共错误码：Public.<错误标识>
//   - 具体错误码：<服务名>.<错误标识>.<错误说明>

// 错误标识
// FIXME: RequestTimeout 和 UnsupportedMediaType 在 DocSet 中的实际使用场景
// 与 HTTP 状态码的语义不匹配，为避免意料之外的破坏性变更，此处增加对这两个错误码的支持。
const (
	BadRequest           = "BadRequest"           // 400 [http.StatusBadRequest]
	Unauthorized         = "Unauthorized"         // 401 [http.StatusUnauthorized]
	Forbidden            = "Forbidden"            // 403 [http.StatusForbidden]
	NotFound             = "NotFound"             // 404 [http.StatusNotFound]
	RequestTimeout       = "RequestTimeout"       // 408 [http.StatusRequestTimeout]
	Conflict             = "Conflict"             // 409 [http.StatusConflict]
	ContentTooLarge      = "ContentTooLarge"      // 413 [http.StatusRequestEntityTooLarge]
	UnsupportedMediaType = "UnsupportedMediaType" // 415 [http.StatusUnsupportedMediaType]

	InternalServerError = "InternalServerError" // 500 [http.StatusInternalServerError]
	ServiceUnavailable  = "ServiceUnavailable"  // 503 [http.StatusServiceUnavailable]
)

// strCodeStatusCodeMap 错误标识到HTTP状态码的映射
var strCodeStatusCodeMap = map[string]int{
	BadRequest:           http.StatusBadRequest,
	Unauthorized:         http.StatusUnauthorized,
	Forbidden:            http.StatusForbidden,
	NotFound:             http.StatusNotFound,
	RequestTimeout:       http.StatusRequestTimeout,
	Conflict:             http.StatusConflict,
	ContentTooLarge:      http.StatusRequestEntityTooLarge,
	UnsupportedMediaType: http.StatusUnsupportedMediaType,

	InternalServerError: http.StatusInternalServerError,
	ServiceUnavailable:  http.StatusServiceUnavailable,
}

// StrCodeToStatusCode 将错误标识转换为HTTP状态码
// 不支持的错误标识会引发 panic，并打印所有支持的错误标识
func StrCodeToStatusCode(strCode string) int {
	v, ok := strCodeStatusCodeMap[strCode]
	if !ok {
		validCodes := make([]string, 0)
		for code := range strCodeStatusCodeMap {
			validCodes = append(validCodes, code)
		}
		panic(fmt.Sprintf("unsupported string code '%s', valid string codes: %s", strCode, strings.Join(validCodes, ", ")))
	}
	return v
}

// statusCodeStrCodeMap HTTP状态码到错误标识的映射
var statusCodeStrCodeMap = make(map[int]string)

// 用 [strCodeStatusCodeMap] 初始化 statusCodeStrCodeMap
func init() {
	for k, v := range strCodeStatusCodeMap {
		statusCodeStrCodeMap[v] = k
	}
}

// StatusCodeToStrCode 将HTTP状态码转换为错误标识
// 不支持的状态码会引发 panic，并打印所有支持的状态码
func StatusCodeToStrCode(statusCode int) string {
	v, ok := statusCodeStrCodeMap[statusCode]
	if !ok {
		validCodes := make([]string, 0)
		for code := range statusCodeStrCodeMap {
			validCodes = append(validCodes, strconv.Itoa(code))
		}
		panic(fmt.Sprintf("unsupported status code %d, valid status codes: %s", statusCode, strings.Join(validCodes, ", ")))
	}
	return v
}

// 公共错误码前缀
const publicStrCodePrefix = "Public."

// 公共错误码
const (
	PublicBadRequest           = publicStrCodePrefix + BadRequest
	PublicUnauthorized         = publicStrCodePrefix + Unauthorized
	PublicForbidden            = publicStrCodePrefix + Forbidden
	PublicNotFound             = publicStrCodePrefix + NotFound
	PublicRequestTimeout       = publicStrCodePrefix + RequestTimeout
	PublicConflict             = publicStrCodePrefix + Conflict
	PublicContentTooLarge      = publicStrCodePrefix + ContentTooLarge
	PublicUnsupportedMediaType = publicStrCodePrefix + UnsupportedMediaType

	PublicInternalServerError = publicStrCodePrefix + InternalServerError
	PublicServiceUnavailable  = publicStrCodePrefix + ServiceUnavailable
)

// StatusCodeToPublicErr 将HTTP状态码转换为对应的公共错误码
func StatusCodeToPublicErr(statusCode int) string {
	return publicStrCodePrefix + StatusCodeToStrCode(statusCode)
}

// Error 错误信息
type Error struct {
	Code        string
	Description string
	Solution    string
	Detail      map[string]any
	Link        string
}

// SetErrAttribute 设置参数
type SetErrAttribute func(*Error)

// SetSolution 设置solution参数
func SetSolution(solution string) SetErrAttribute {
	return func(err *Error) {
		err.Solution = solution
	}
}

// SetDetail 设置detail参数
func SetDetail(detail map[string]any) SetErrAttribute {
	return func(e *Error) {
		e.Detail = detail
	}
}

// SetLink 设置Link参数
func SetLink(link string) SetErrAttribute {
	return func(e *Error) {
		e.Link = link
	}
}

// NewError 新建一个Error
func NewError(code string, description string, setters ...SetErrAttribute) *Error {
	CheckCodeValid(code)
	e := &Error{
		Code:        code,
		Description: description,
	}

	// 设置可选属性
	for _, setter := range setters {
		setter(e)
	}

	return e
}

func (e *Error) Error() string {
	data := map[string]any{
		"code":        e.Code,
		"description": e.Description,
	}
	if e.Solution != "" {
		data["solution"] = e.Solution
	}
	if len(e.Detail) != 0 {
		data["detail"] = e.Detail
	}
	if e.Link != "" {
		data["link"] = e.Link
	}
	errstr, err := jsoniter.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(errstr)
}

func CheckCodeValid(code string) {
	parts := strings.Split(code, ".")
	partCount := len(parts)

	// 检查段数
	if partCount != 2 && partCount != 3 {
		panic("parameter 'code' should follow 'Public.xxx' or 'xxx.xxx.xxx' format, but got: " + code)
	}

	// 检查错误标识
	_ = StrCodeToStatusCode(parts[1])

	// Public.<错误标识>
	if partCount == 2 {
		if parts[0]+"." != publicStrCodePrefix {
			panic("parameter 'code' should follow 'Public.xxx' format, but got: " + code)
		}
		return
	}

	// <服务名>.<错误标识>.<错误说明>
	if serviceNameLen := len(parts[0]); serviceNameLen < 1 || serviceNameLen > 16 {
		panic(fmt.Sprintf("service name length should be between 1 and 16 characters, but got: %s, length: %d", parts[0], serviceNameLen))
	}

	if descriptionLen := len(parts[2]); descriptionLen < 1 || descriptionLen > 36 {
		panic(fmt.Sprintf("description length should be between 1 and 36 characters, but got: %s, length: %d", parts[2], descriptionLen))
	}
}
