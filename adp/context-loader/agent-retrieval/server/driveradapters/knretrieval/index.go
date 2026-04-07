// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knretrieval provides HTTP handler for knowledge network retrieval operations.
package knretrieval

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
	logicskn "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knretrieval"
)

// KnRetrievalHandler 基于业务知识网络实现统一Retrieval
type KnRetrievalHandler interface {
	SemanticSearch(c *gin.Context)
}

type knRetrievalHandle struct {
	Logger             interfaces.Logger
	KnRetrievalService interfaces.IKnRetrievalService
}

var (
	knOnce    sync.Once
	knHandler KnRetrievalHandler
)

// NewKnRetrievalHandler 新建KnRetrievalHandler
func NewKnRetrievalHandler() KnRetrievalHandler {
	knOnce.Do(func() {
		conf := config.NewConfigLoader()
		knHandler = &knRetrievalHandle{
			Logger:             conf.GetLogger(),
			KnRetrievalService: logicskn.NewKnRetrievalService(),
		}
	})
	return knHandler
}

// SemanticSearch 语义检索
func (k *knRetrievalHandle) SemanticSearch(c *gin.Context) {
	var err error
	req := &interfaces.SemanticSearchRequest{
		SearchScope: &interfaces.SearchScopeConfig{},
	}

	if err = c.ShouldBindHeader(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}
	if err = c.ShouldBindQuery(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	if err = c.ShouldBindJSON(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	if err = defaults.Set(req); err != nil {
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, err.Error())
		rest.ReplyError(c, err)
		return
	}

	err = validator.New().Struct(req)
	if err != nil {
		rest.ReplyError(c, err)
		return
	}

	var resp *interfaces.SemanticSearchResponse
	switch req.Mode {
	case interfaces.AgentIntentRetrieval:
		resp, err = k.KnRetrievalService.AgentIntentRetrieval(c.Request.Context(), req)
	case interfaces.AgentIntentPlanning:
		resp, err = k.KnRetrievalService.AgentIntentPlanning(c.Request.Context(), req)
	case interfaces.KeywordVectorRetrieval:
		resp, err = k.KnRetrievalService.KeywordVectorRetrieval(c.Request.Context(), req)
	default:
		err = errors.DefaultHTTPError(c.Request.Context(), http.StatusBadRequest, "mode not support")
	}
	if err != nil {
		k.Logger.Errorf("SemanticSearch mode:%s err, err: %v", req.Mode, err)
		rest.ReplyError(c, err)
		return
	}

	resp.QueryUnderstanding = nil
	resp.HitsTotal = 0
	for i := range resp.KnowledgeConcepts {
		resp.KnowledgeConcepts[i].IntentScore = 0
		resp.KnowledgeConcepts[i].MatchScore = 0
		resp.KnowledgeConcepts[i].RerankScore = 0
		resp.KnowledgeConcepts[i].Samples = nil
	}
	rest.ReplyOK(c, http.StatusOK, resp)
}
