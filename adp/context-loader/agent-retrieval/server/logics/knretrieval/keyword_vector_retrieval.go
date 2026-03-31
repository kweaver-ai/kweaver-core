// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// KeywordVectorRetrieval 基于关键词+向量召回
func (k *knRetrievalServiceImpl) KeywordVectorRetrieval(ctx context.Context, req *interfaces.SemanticSearchRequest) (resp *interfaces.SemanticSearchResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 查询策略
	var queryStrategys []*interfaces.SemanticQueryStrategy
	// 概念结果候选集
	conceptResults := []*interfaces.ConceptResult{}
	// 自定义构建查询策略，请求业务知识网络接口做关键词匹配
	queryStrategys = k.longtailRecallByKnowledgeNetwork(req.Query)
	// 筛选查询策略
	queryStrategys = k.filterQueryStrategysBySearchScope(queryStrategys, req.SearchScope)
	if len(queryStrategys) > 0 {
		// 并发执行查询策略
		var queryConceptResults []*interfaces.ConceptResult
		queryConceptResults, err = k.parallelExecSemanticQueryStrategy(ctx, req.KnID, queryStrategys)
		if err != nil {
			k.logger.WithContext(ctx).Warnf("[SemanticSearchV2] parallelExecSemanticQueryStrategy failed. knId:%s, queryStrategys:%v, err:%v", req.KnID, queryStrategys, err)
			return
		}
		if len(queryConceptResults) > 0 {
			conceptResults = append(conceptResults, queryConceptResults...)
		}
	}
	queryUnderstanding := &interfaces.QueryUnderstanding{
		OriginQuery:    req.Query,
		QueryStrategys: queryStrategys,
	}
	// TODO：实例数据采样（本版本跳过）
	// 排序：根据匹配分数排序，去重
	rerankConceptResults, err := k.rerankByDataRetrieval(ctx, queryUnderstanding, conceptResults, req.RerankAction, req.MaxConcepts)
	if err != nil {
		return
	}
	// 组装结果
	resp = &interfaces.SemanticSearchResponse{
		QueryUnderstanding: queryUnderstanding,
		KnowledgeConcepts:  rerankConceptResults,
		HitsTotal:          len(conceptResults),
	}
	return
}
