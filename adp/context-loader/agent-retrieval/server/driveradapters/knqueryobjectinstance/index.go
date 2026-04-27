// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knqueryobjectinstance provides HTTP handler for object instance query operations.
package knqueryobjectinstance

import (
	"net/http"
	"sync"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// KnQueryObjectInstanceHandler 查询对象实例处理器
type KnQueryObjectInstanceHandler interface {
	QueryObjectInstance(c *gin.Context)
}

type knQueryObjectInstanceHandler struct {
	Logger        interfaces.Logger
	OntologyQuery interfaces.DrivenOntologyQuery
}

var (
	koiOnce    sync.Once
	koiHandler KnQueryObjectInstanceHandler
)

// NewKnQueryObjectInstanceHandler 新建 KnQueryObjectInstanceHandler
func NewKnQueryObjectInstanceHandler() KnQueryObjectInstanceHandler {
	koiOnce.Do(func() {
		conf := config.NewConfigLoader()
		koiHandler = &knQueryObjectInstanceHandler{
			Logger:        conf.GetLogger(),
			OntologyQuery: drivenadapters.NewOntologyQueryAccess(),
		}
	})
	return koiHandler
}

// QueryObjectInstance 查询对象实例
func (h *knQueryObjectInstanceHandler) QueryObjectInstance(c *gin.Context) {
	var err error
	req := &interfaces.QueryObjectInstancesReq{}

	// 绑定 Header
	if err = c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 Query Parameters
	if err = c.ShouldBindQuery(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 绑定 JSON Body
	if err = c.ShouldBindJSON(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	// 设置默认值
	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	req.IncludeTypeInfo = false

	// 参数校验
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 调用业务逻辑
	resp, err := h.OntologyQuery.QueryObjectInstances(c.Request.Context(), req)
	if err != nil {
		h.Logger.Errorf("[KnQueryObjectInstanceHandler#QueryObjectInstance] QueryObjectInstances failed, err: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
