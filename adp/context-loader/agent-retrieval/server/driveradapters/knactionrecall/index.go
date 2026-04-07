// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knactionrecall provides HTTP handler for knowledge network action recall operations.
package knactionrecall

import (
	"net/http"
	"sync"

	"github.com/creasty/defaults"
	"github.com/gin-gonic/gin"
	validator "github.com/go-playground/validator/v10"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	logicsKAR "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knactionrecall"
)

// KnActionRecallHandler 业务知识网络行动召回处理器
type KnActionRecallHandler interface {
	GetActionInfo(c *gin.Context)
}

type knActionRecallHandler struct {
	Logger                interfaces.Logger
	KnActionRecallService interfaces.IKnActionRecallService
}

var (
	karOnce    sync.Once
	karHandler KnActionRecallHandler
)

// NewKnActionRecallHandler 新建 KnActionRecallHandler
func NewKnActionRecallHandler() KnActionRecallHandler {
	karOnce.Do(func() {
		conf := config.NewConfigLoader()
		karHandler = &knActionRecallHandler{
			Logger:                conf.GetLogger(),
			KnActionRecallService: logicsKAR.NewKnActionRecallService(),
		}
	})
	return karHandler
}

// GetActionInfo 获取行动信息（行动召回）
func (h *knActionRecallHandler) GetActionInfo(c *gin.Context) {
	var err error
	req := &interfaces.KnActionRecallRequest{}

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

	// 参数校验
	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	// 调用业务逻辑
	resp, err := h.KnActionRecallService.GetActionInfo(c.Request.Context(), req)
	if err != nil {
		h.Logger.Errorf("[KnActionRecallHandler#GetActionInfo] GetActionInfo failed, err: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
