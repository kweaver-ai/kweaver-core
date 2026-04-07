package response

import (
	"fmt"
	"net/http"
	"oss-gateway/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构，参考 Python FastAPI 的响应格式
type Response struct {
	Res         int         `json:"res,omitempty"`         // 兼容旧格式：0表示成功
	Code        string      `json:"code,omitempty"`        // 错误码字符串
	Message     string      `json:"message,omitempty"`     // 错误消息
	Description string      `json:"description,omitempty"` // 错误描述
	Detail      string      `json:"detail,omitempty"`      // 错误详情
	Solution    string      `json:"solution,omitempty"`    // 解决方案
	Cause       string      `json:"cause,omitempty"`       // 错误原因
	Data        interface{} `json:"data,omitempty"`        // 成功时返回的数据
	Count       int         `json:"count,omitempty"`       // 列表总数（分页时使用）
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Res:  0,
		Data: data,
	})
}

// SuccessWithCount 成功响应，带总数（用于分页列表）
func SuccessWithCount(c *gin.Context, data interface{}, count int) {
	c.JSON(http.StatusOK, Response{
		Count: count,
		Data:  data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpStatus int, code string, message string, description string, solution string, cause string) {
	c.JSON(httpStatus, Response{
		Code:        code,
		Message:     message,
		Description: description,
		Detail:      description,
		Solution:    solution,
		Cause:       cause,
	})
}

// BadRequest 400 错误请求
func BadRequest(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusBadRequest, &errors.BadRequest, message)
}

// NotFound 404 未找到
func NotFound(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusNotFound, &errors.NotFound, message)
}

// InternalError 500 内部错误
func InternalError(c *gin.Context, message string) {
	ErrorWithCode(c, http.StatusInternalServerError, &errors.InternalError, message)
}

// InvalidParam 无效参数
func InvalidParam(c *gin.Context, param string) {
	code := &errors.InvalidParam
	Error(c, http.StatusBadRequest, code.Code, code.Message, fmt.Sprintf(code.Description, param), code.Solution, param)
}

// StorageNotFound 存储未找到
func StorageNotFound(c *gin.Context) {
	code := &errors.StorageNotFound
	Error(c, http.StatusNotFound, code.Code, code.Message, code.Description, code.Solution, "")
}

// ConnectionFailed 连接失败
func ConnectionFailed(c *gin.Context, cause string) {
	code := &errors.ConnectionFailed
	Error(c, http.StatusInternalServerError, code.Code, code.Message, code.Description, code.Solution, cause)
}

// StorageNameExist 存储名称已存在
func StorageNameExist(c *gin.Context, storageName string) {
	code := &errors.StorageNameExists
	Error(c, http.StatusBadRequest, code.Code, code.Message,
		fmt.Sprintf(code.Description, storageName),
		code.Solution, "")
}

// StorageExist 存储已存在（bucket+host 或 bucket+siteId）
func StorageExist(c *gin.Context, description string) {
	code := &errors.StorageExists
	Error(c, http.StatusBadRequest, code.Code, code.Message,
		description,
		code.Solution, "")
}

// TooManyKeys 键过多
func TooManyKeys(c *gin.Context, maxKeys int) {
	code := &errors.TooManyKeys
	Error(c, http.StatusBadRequest, code.Code, code.Message,
		fmt.Sprintf(code.Description, maxKeys),
		code.Solution, "")
}

// InvalidVendorType 无效的供应商类型
func InvalidVendorType(c *gin.Context, vendorType string) {
	code := &errors.InvalidVendorType
	Error(c, http.StatusBadRequest, code.Code, code.Message,
		fmt.Sprintf(code.Description, vendorType),
		code.Solution, "")
}

// DefaultStorageExists 默认存储已存在
func DefaultStorageExists(c *gin.Context, existingStorageName string) {
	code := &errors.DefaultStorageExists
	Error(c, http.StatusBadRequest, code.Code, code.Message,
		fmt.Sprintf(code.Description, existingStorageName),
		code.Solution, "")
}

// ErrorWithCode 使用 ErrorCode 结构返回错误
func ErrorWithCode(c *gin.Context, httpStatus int, code *errors.ErrorCode, customMessage string) {
	description := customMessage
	if description == "" {
		description = code.Description
	}
	Error(c, httpStatus, code.Code, code.Message, description, code.Solution, "")
}
