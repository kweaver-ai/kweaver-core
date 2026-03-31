package utils

import (
	"bytes"
	"mime/multipart"
	"net/http"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// GetBindJSONRaw 获取原始请求体
func GetBindJSONRaw(c *gin.Context, req interface{}) (err error) {
	// 读取请求体之前先判断是否为空
	if c.Request.Body == nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "request body is empty")
		return
	}
	// 读取请求体
	err = c.ShouldBindJSON(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
	}
	return
}

// GetBindMultipartFormRaw 获取原始 multipart/form-data 请求体
func GetBindMultipartFormRaw(c *gin.Context, req interface{}, fileKey string, fileSizeLimit int64) (fileBytes []byte, err error) {
	// 读取请求体之前先判断是否为空
	if c.Request.Body == nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "request body is empty")
		return
	}
	// 读取请求体
	err = c.ShouldBindWith(req, binding.Form)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		return
	}
	// 获取文件
	var file *multipart.FileHeader
	file, err = c.FormFile(fileKey)
	if err != nil {
		// 判断是否是文件不存在的错误
		if err == http.ErrMissingFile || err.Error() == "http: no such file" {
			err = nil
		} else {
			err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		}
		return
	}
	// TODO : 检查文件大小
	if fileSizeLimit > 0 && file.Size > fileSizeLimit {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "file size exceeds limit")
		return
	}
	var fileContent multipart.File
	fileContent, err = file.Open()
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		return
	}
	defer func() {
		_ = fileContent.Close()
	}()
	// 读取文件内容
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(fileContent); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		return
	}
	fileBytes = buf.Bytes()
	return
}

// GetBindFormRaw 获取原始 application/x-www-form-urlencoded 请求体
func GetBindFormRaw(c *gin.Context, req interface{}) (err error) {
	// 读取请求体之前先判断是否为空
	if c.Request.Body == nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "request body is empty")
		return
	}
	// 读取请求体
	err = c.ShouldBind(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
	}
	return
}
