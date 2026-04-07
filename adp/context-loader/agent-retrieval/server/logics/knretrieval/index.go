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

// useLocalRerank Feature Flag: 是否使用本地Rerank
// 迁移验证通过后改为true，最终删除此开关和远程调用代码
const useLocalRerank = true

type knRetrievalServiceImpl struct {
	logger                interfaces.Logger
	agentClient           interfaces.AgentApp
	ontologyQueryAccess   interfaces.DrivenOntologyQuery
	bknBackendAccess      interfaces.BknBackendAccess
	dataRetrieval         interfaces.DataRetrieval
	knReranker            *knrerank.KnowledgeReranker
	useLocalRerank        bool
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
			logger:                logger,
			agentClient:           drivenadapters.NewAgentAppClient(),
			ontologyQueryAccess:   drivenadapters.NewOntologyQueryAccess(),
			bknBackendAccess:      drivenadapters.NewBknBackendAccess(),
			dataRetrieval:         drivenadapters.NewDataRetrievalClient(),
			knReranker:            knrerank.NewKnowledgeReranker(mfModelClient, logger), // 单例
			useLocalRerank:        useLocalRerank,
		}
	})
	return knRetrievalService
}
