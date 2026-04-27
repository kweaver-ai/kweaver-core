// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knactionrecall 业务知识网络行动召回业务逻辑
// file: index.go
package knactionrecall

import (
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type knActionRecallServiceImpl struct {
	logger              interfaces.Logger
	config              *config.Config
	ontologyQuery       interfaces.DrivenOntologyQuery
	operatorIntegration interfaces.DrivenOperatorIntegration
}

var (
	karOnce               sync.Once
	knActionRecallService interfaces.IKnActionRecallService
)

// NewKnActionRecallService 创建业务知识网络行动召回服务实例
func NewKnActionRecallService() interfaces.IKnActionRecallService {
	karOnce.Do(func() {
		configLoader := config.NewConfigLoader()
		knActionRecallService = &knActionRecallServiceImpl{
			logger:              configLoader.GetLogger(),
			config:              configLoader,
			ontologyQuery:       drivenadapters.NewOntologyQueryAccess(),
			operatorIntegration: drivenadapters.NewOperatorIntegrationClient(),
		}
	})
	return knActionRecallService
}
