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

	// 4. 如果只召回概念，直接返回
	if req.OnlySchema {
		s.logger.WithContext(ctx).Infof("[KnSearchLocal] Only schema mode, skip semantic instance retrieval")
		return response, nil
	}

	// 5. 语义实例召回（如果有过滤后的对象类型）
	if len(conceptResult.ObjectTypes) > 0 {
		instanceResult, err := s.semanticInstanceRetrieval(ctx, req, conceptResult.ObjectTypes, mergedConfig)
		if err != nil {
			s.logger.WithContext(ctx).Warnf("[KnSearchLocal] Semantic instance retrieval failed: %v", err)
			response.Message = "语义实例召回失败: " + err.Error()
		} else {
			response.Nodes = instanceResult.Nodes
			if instanceResult.Message != "" {
				response.Message = instanceResult.Message
			}
			s.logger.WithContext(ctx).Infof("[KnSearchLocal] Semantic instance retrieval completed: nodes=%d", len(response.Nodes))
		}
	} else {
		response.Message = "未召回到相关概念，无法进行实例检索"
	}

	return response, nil
}
