package rest

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	errorv2 "github.com/kweaver-ai/adp/autoflow/flow-automation/pkg/ecron/utils/rest/error"
)

var (
	// Languages 支持的语言
	Languages = [3]string{"zh_CN", "zh_TW", "en_US"}
)

var (
	i18n         = make(map[int]map[string]string)
	code2Message = make(map[int]string)
)

// SetLang 设置语言
func SetLang(lang string) {
	valid := false
	for _, l := range Languages {
		if l == lang {
			valid = true
		}
	}
	if !valid {
		panic("invalid lang")
	}
	for code := range i18n {
		code2Message[code] = i18n[code][lang]
	}
}

// Register 注册code对应message
func Register(langRes map[int]map[string]string) {
	for code, message := range langRes {
		if _, ok := i18n[code]; ok {
			panic(fmt.Sprintf("duplicate code: %v", code))
		}
		i18n[code] = make(map[string]string)
		for _, lang := range Languages {
			if m, ok := message[lang]; ok {
				i18n[code][lang] = m
			} else {
				panic(fmt.Sprintf("language %v not exists", lang))
			}
		}
	}
}

// HTTPError 服务错误结构体
type HTTPError struct {
	Cause       string                 `json:"cause"`
	Code        int                    `json:"code"`
	Message     string                 `json:"message"`
	Detail      map[string]interface{} `json:"detail,omitempty"`
	Description string                 `json:"description,omitempty"`
	Solution    string                 `json:"solution,omitempty"`
	CodeStr     string
	useCodeStr  bool
}

// SetErrAttribute 设置参数
type SetErrAttribute func(*HTTPError)

// SetDescription 设置description参数
// Deprecated: 如需设置 Description, 请使用 NewHTTPErrorV2
func SetDescription(description string) SetErrAttribute {
	return func(err *HTTPError) {
		err.Description = description
	}
}

// SetSolution 设置solution参数
func SetSolution(solution string) SetErrAttribute {
	return func(err *HTTPError) {
		err.Solution = solution
	}
}

// SetDetail 设置detail参数
func SetDetail(detail map[string]interface{}) SetErrAttribute {
	return func(e *HTTPError) {
		e.Detail = detail
	}
}

// SetCodeStr 设置codeStr
func SetCodeStr(codeStr string) SetErrAttribute {
	errorv2.CheckCodeValid(codeStr)
	return func(e *HTTPError) {
		e.CodeStr = codeStr
	}
}

// NewHTTPErrorV2 新建一个HTTPError
func NewHTTPErrorV2(code int, description string, setters ...SetErrAttribute) *HTTPError {
	checkCodeValid(code)
	e := &HTTPError{
		Code:        code,
		Message:     code2Message[code],
		Cause:       description,
		Description: description,
	}

	// 设置可选属性
	for _, setter := range setters {
		setter(e)
	}

	return e
}

// NewHTTPError 新建一个HTTPError
func NewHTTPError(cause string, code int, detail map[string]interface{}, params ...SetErrAttribute) *HTTPError {
	checkCodeValid(code)
	httpError := &HTTPError{
		Cause:   cause,
		Code:    code,
		Message: code2Message[code],
		Detail:  detail,
	}
	for _, p := range params {
		p(httpError)
	}

	return httpError
}

func (e *HTTPError) Error() string {
	data := make(map[string]interface{})
	if len(e.Detail) != 0 {
		data["detail"] = e.Detail
	}
	if e.Solution != "" {
		data["solution"] = e.Solution
	}
	if e.useCodeStr {
		// 使用NewXXX方法创建HTTPError实例时已经检查了code合法性
		if len(e.CodeStr) == 0 {
			data["code"] = errorv2.StatusCodeToPublicErr(e.Code / 1e6)
			// 保留原始code、cause，用于追溯原始错误
			if data["detail"] == nil {
				data["detail"] = make(map[string]interface{})
			}
			detail := data["detail"].(map[string]interface{})
			detail["original_code_cause"] = fmt.Sprintf("code: %d, cause: %s", e.Code, e.Cause)
		} else {
			data["code"] = e.CodeStr
		}
		if e.Description != "" {
			data["description"] = e.Description
		} else {
			// description 会用于 UI 交互，所以这里在给 description 赋值时，考虑了国际化的问题。
			// 最初的想法是，ReplyError 时从报头中取出 x-language 字段，并在 HTTPError 中增加一个字段存储 x-language，
			// 根据 code、language 从 i18n 中取出对应的 message 赋给 description。
			// 但这种做法存在一个问题，i18n 是服务启动时调用 Register 函数进行注册的，即 code 对应的 message 仅在本服务中有定义。
			// 当服务 A 调用 c++ 的服务 B，由于目前 c++ 服务未适配新错误码规范，服务 A 原样抛转服务 B 的错误时，服务 A 并没有服务 B 关于特定 code 对应的 message信息，无法完成转换，
			// 所以这里直接使用 e.Message 作为 description。
			data["description"] = e.Message
		}
	} else {
		data["code"] = e.Code
		data["message"] = e.Message
		data["cause"] = e.Cause
		if e.Description != "" {
			data["description"] = e.Description
		}
	}
	errStr, err := jsoniter.Marshal(data)
	if err != nil {
		panic(err)
	}
	return string(errStr)
}

// ExHTTPError 其他服务响应的错误结构体
type ExHTTPError struct {
	Status int
	Body   []byte
}

func (err ExHTTPError) Error() string {
	return string(err.Body)
}

// checkCodeValid 检查错误码是否合法
func checkCodeValid(code int) {
	if code < BadRequest || code >= 600000000 {
		panic("the parameter 'code' length should be 9")
	}
	digit := code / 1e6
	_ = errorv2.StatusCodeToStrCode(digit)
}
