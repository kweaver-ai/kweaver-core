// Package common 公共模块操作接口
package common

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"
	"time"

	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/config"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/errors"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/rest"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/infra/validator"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/interfaces"
	"github.com/kweaver-ai/adp/execution-factory/operator-integration/server/logics/impex"
)

type ImpexHandler interface {
	Export(c *gin.Context)
	Import(c *gin.Context)
}

var (
	impexHandlerOnce sync.Once
	impexH           ImpexHandler
)

type impexHandler struct {
	Logger               interfaces.Logger
	ComponentImpexConfig interfaces.IComponentImpexConfig
	Validator            interfaces.Validator
}

// NewImpexHandler 导入导出操作接口
func NewImpexHandler() ImpexHandler {
	impexHandlerOnce.Do(func() {
		confLoader := config.NewConfigLoader()
		impexH = &impexHandler{
			Logger:               confLoader.GetLogger(),
			ComponentImpexConfig: impex.NewComponentImpexManager(),
			Validator:            validator.NewValidator(),
		}
	})
	return impexH
}

// Export 导出
func (impexH *impexHandler) Export(c *gin.Context) {
	var err error
	req := &interfaces.ExportConfigReq{}
	if err = c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	if err = c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = impexH.Validator.ValidatorStruct(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	config, err := impexH.ComponentImpexConfig.ExportConfig(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	filename := fmt.Sprintf("%s_export_%s.adp", req.Type, time.Now().Format("20060102_150405"))
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	rest.ReplyOK(c, http.StatusOK, config)
}

// Import 导入
func (impexH *impexHandler) Import(c *gin.Context) {
	var err error
	req := &interfaces.ImportConfigReq{}
	if err = c.ShouldBindHeader(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	if err = c.ShouldBindUri(req); err != nil {
		rest.ReplyError(c, err)
		return
	}
	// 检查Content-Type
	if c.ContentType() != "multipart/form-data" {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusUnsupportedMediaType, "Content-Type must be multipart/form-data")
		rest.ReplyError(c, err)
		return
	}

	err = c.ShouldBindWith(req, binding.Form)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	var file *multipart.FileHeader
	file, err = c.FormFile("data")
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	var fileContent multipart.File
	// TODO: 检查文件大小
	fileContent, err = file.Open()
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	defer func() {
		_ = fileContent.Close()
	}()
	buf := new(bytes.Buffer)
	if _, err = buf.ReadFrom(fileContent); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	req.Data = buf.Bytes()
	err = defaults.Set(req)
	if err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	err = impexH.Validator.ValidatorStruct(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	err = impexH.ComponentImpexConfig.ImportConfig(c.Request.Context(), req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}
	rest.ReplyOK(c, http.StatusCreated, nil)
}
