package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	liberrors "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/errors"
	"github.com/kweaver-ai/adp/autoflow/flow-automation/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

var (
	// Languages 支持的语言
	Languages = [3]string{"zh_CN", "zh_TW", "en_US"}
)

// IError ierror struct
type IError struct {
	ErrorCode    string      `json:"code"`
	MainCode     string      `json:"-"`
	Description  string      `json:"description"`
	Solution     string      `json:"solution"`
	ErrorDetails interface{} `json:"detail,omitempty"`
	ErrorLink    string      `json:"link,omitempty"`
}

// IErrorCodeStr ierror code str
type IErrorCodeStr struct {
	ModuleName         string
	MainErrorCode      string
	ExtendedErrorCodes []string
}

// HTTPError http err struct
type HTTPError struct {
	Status int                    `json:"status"`
	Body   map[string]interface{} `json:"body"`
	Err    error                  `json:"err"`
}

// NewIError new error instance
func NewIError(mainCode, exCode string, detail interface{}) *IError {
	var code string
	var moduleName = ModuleName
	if mainCode == UnAuthorization {
		moduleName = "Common"
	}
	if exCode == "" {
		code = fmt.Sprintf("%s.%s", moduleName, mainCode)
	} else {
		code = fmt.Sprintf("%s.%s.%s", moduleName, mainCode, exCode)
	}

	lang := utils.GetLanguage()

	codeType := mainCode
	if _, ok := ErrorsMsg[exCode]; ok {
		codeType = exCode
	}

	errMsg, ok := ErrorsMsg[codeType][lang]
	if !ok {
		errMsg = ErrorsMsg[codeType][Languages[0]]
	}

	return &IError{
		ErrorCode:    code,
		MainCode:     mainCode,
		Description:  errMsg[0],
		Solution:     errMsg[1],
		ErrorDetails: detail,
	}
}

func (e *IError) Error() string {
	return e.Description
}

// NewHTTPError  new http error instance
func NewHTTPError(info string, status int, body map[string]interface{}) *HTTPError {
	return &HTTPError{
		Status: status,
		Body:   body,
		Err:    errors.New(info),
	}
}

func (h *HTTPError) Error() string {
	return ""
}

// ExHTTPError http错误
type ExHTTPError struct {
	Body   string
	Status int
	Err    error
}

func (e ExHTTPError) Error() string {
	errorinfo := fmt.Sprintf("body : %s , status : %v", e.Body, e.Status)
	return errorinfo
}

// ExHTTPErrorParser parse http error
func ExHTTPErrorParser(err error) (map[string]interface{}, error) {
	httpError, ok := err.(ExHTTPError)
	var httpErrorBody map[string]interface{}

	if !ok {
		if httpError, hok := err.(liberrors.ExHTTPError); hok {
			parseErr, rawErr := liberrors.ExHTTPErrorParser(httpError)
			return parseErr.Body, rawErr
		}
		return httpErrorBody, err
	}

	parseErr := json.Unmarshal([]byte(httpError.Body), &httpErrorBody)

	if parseErr != nil {
		return httpErrorBody, err
	}

	return httpErrorBody, nil
}

// Is 检查错误的 error类型
func Is(err error, errCode ...string) bool {
	errCodeStr := strings.Join(errCode, ".")
	errCodeStr = fmt.Sprintf("%s.%s", ModuleName, errCodeStr)
	if apiError, ok := err.(*IError); ok {
		return apiError.ErrorCode == errCodeStr
	}
	return false
}

const dataNotFoundText = "data not found"

// IsNotFoundErr returns whether err represents a missing record across store backends.
func IsNotFoundErr(err error) bool {
	if err == nil {
		return false
	}

	return errors.Is(err, mongo.ErrNoDocuments) ||
		errors.Is(err, gorm.ErrRecordNotFound) ||
		strings.Contains(err.Error(), dataNotFoundText)
}
