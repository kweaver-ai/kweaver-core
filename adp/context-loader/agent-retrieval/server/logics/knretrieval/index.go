// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knretrieval 基于业务知识网络实现统一检索
// file: index.go
package knretrieval

import (
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knrerank"
)

type knRetrievalServiceImpl struct {
	logger              interfaces.Logger
	agentClient         interfaces.AgentApp
	ontologyQueryAccess interfaces.DrivenOntologyQuery
	bknBackendAccess    interfaces.BknBackendAccess
	knReranker          *knrerank.KnowledgeReranker
}

var (
	krOnce             sync.Once
	knRetrievalService interfaces.IKnRetrievalService
)

func NewKnRetrievalService() interfaces.IKnRetrievalService {
	krOnce.Do(func() {
		conf := config.NewConfigLoader()
		logger := conf.GetLogger()

		// 创建统一的mf-model-api客户端（同时提供LLM和Rerank能力）
		mfModelClient := drivenadapters.NewMFModelAPIClient()

		knRetrievalService = &knRetrievalServiceImpl{
			logger:              logger,
			agentClient:         drivenadapters.NewAgentAppClient(),
			ontologyQueryAccess: drivenadapters.NewOntologyQueryAccess(),
			bknBackendAccess:    drivenadapters.NewBknBackendAccess(),
			knReranker:          knrerank.NewKnowledgeReranker(mfModelClient, logger), // 单例
		}
	})
	return knRetrievalService
}
