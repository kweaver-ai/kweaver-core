// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch (本地检索主服务实现)
// file: service.go
package knsearch

import (
	"context"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// Search 知识网络检索本地主入口
func (s *localSearchImpl) Search(ctx context.Context, req *interfaces.KnSearchLocalRequest) (*interfaces.KnSearchLocalResponse, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	s.logger.WithContext(ctx).Infof("[KnSearchLocal] Start Search, kn_id=%s, query=%s, only_schema=%v",
		req.KnID, req.Query, req.OnlySchema)

	// 1. 合并配置
	mergedConfig := MergeRetrievalConfig(req.RetrievalConfig)
	s.logger.WithContext(ctx).Debugf("[KnSearchLocal] Merged config: concept_top_k=%d, schema_brief=%v, enable_coarse_recall=%v",
		mergedConfig.ConceptRetrieval.TopK,
		boolValue(mergedConfig.ConceptRetrieval.SchemaBrief),
		boolValue(mergedConfig.ConceptRetrieval.EnableCoarseRecall))

	// 2. 概念召回（Schema Recall）
	conceptResult, err := s.conceptRetrieval(ctx, req, mergedConfig.ConceptRetrieval)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[KnSearchLocal] Concept retrieval failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Infof("[KnSearchLocal] Concept retrieval completed: object_types=%d, relation_types=%d, action_types=%d",
		len(conceptResult.ObjectTypes), len(conceptResult.RelationTypes), len(conceptResult.ActionTypes))

	// 3. 构建响应
	response := &interfaces.KnSearchLocalResponse{
		ObjectTypes:   conceptResult.ObjectTypes,
		RelationTypes: conceptResult.RelationTypes,
		ActionTypes:   conceptResult.ActionTypes,
	}

	// shared logic 已收敛为 Schema-only，兼容字段仍可传入，但不再触发实例检索。
	s.logger.WithContext(ctx).Infof("[KnSearchLocal] Shared logic converged to schema-only, skip semantic instance retrieval")
	return response, nil
}
