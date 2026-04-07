// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch provides HTTP handler for knowledge network search operations.
package knsearch

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
	logicskn "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knsearch"
)

// KnSearchHandler kn_search 处理器
type KnSearchHandler interface {
	KnSearch(c *gin.Context)
}

type knSearchHandler struct {
	Logger          interfaces.Logger
	KnSearchService interfaces.IKnSearchService
}

var (
	ksOnce    sync.Once
	ksHandler KnSearchHandler
)

// NewKnSearchHandler 新建 KnSearchHandler
func NewKnSearchHandler() KnSearchHandler {
	ksOnce.Do(func() {
		conf := config.NewConfigLoader()
		ksHandler = &knSearchHandler{
			Logger:          conf.GetLogger(),
			KnSearchService: logicskn.NewKnSearchService(),
		}
	})
	return ksHandler
}

// KnSearch 知识网络检索
func (h *knSearchHandler) KnSearch(c *gin.Context) {
	var err error
	req := &interfaces.KnSearchReq{}

	// 绑定 Header
	if err = c.ShouldBindHeader(req); err != nil {
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
	resp, err := h.KnSearchService.KnSearch(c.Request.Context(), req)
	if err != nil {
		h.Logger.Errorf("[KnSearchHandler#KnSearch] KnSearch failed, err: %v", err)
		rest.ReplyError(c, err)
		return
	}

	// 返回成功响应
	rest.ReplyOK(c, http.StatusOK, resp)
}
