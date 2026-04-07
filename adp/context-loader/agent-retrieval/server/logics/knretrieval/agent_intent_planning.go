// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"fmt"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// AgentIntentPlanning 语义搜索: 基于意图分析智能体+规划策略
func (k *knRetrievalServiceImpl) AgentIntentPlanning(ctx context.Context, req *interfaces.SemanticSearchRequest) (resp *interfaces.SemanticSearchResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 调用意图分析智能体 -- 简略意图
	queryUnderstandResult, err := k.agentClient.ConceptIntentionAnalysisAgent(ctx, &interfaces.ConceptIntentionAnalysisAgentReq{
		PreviousQueries: req.PreviousQueries,
		Query:           req.Query,
		KnID:            req.KnID,
	})
	if err != nil {
		k.logger.WithContext(ctx).Warnf("call concept intention analysis agent err, err: %v", err)
	}
	if queryUnderstandResult == nil {
		queryUnderstandResult = &interfaces.QueryUnderstanding{}
	}
	var queryStrategys []*interfaces.SemanticQueryStrategy
	if len(queryUnderstandResult.Intent) > 0 {
		// 基于意图生成查询策略 -- 手动拼接意图查询策略，不需要辅助信息
		queryStrategys = k.generateQueryStrategysForPlanB(queryUnderstandResult)
	} else {
		// 当意图识别策略返回为空时，直接根据用户Query拼接策略
		queryStrategys = k.longtailRecallByKnowledgeNetwork(req.Query)
	}
	// 筛选查询策略
	queryStrategys = k.filterQueryStrategysBySearchScope(queryStrategys, req.SearchScope)
	// TODO: 根据搜索与配置对查询策略进行过滤
	// 策略执行：并发解析并执行query_strategy，获取结果
	conceptResults, err := k.parallelExecSemanticQueryStrategy(ctx, req.KnID, queryStrategys)
	if err != nil {
		return
	}
	// 返回执行的策略
	queryUnderstandResult.QueryStrategys = queryStrategys
	// TODO：实例数据采样（本版本跳过）
	// 排序：精排, 去重
	rerankConceptResults, err := k.rerankByDataRetrieval(ctx, queryUnderstandResult, conceptResults, req.RerankAction, req.MaxConcepts)
	if err != nil {
		return
	}
	// 组装结果
	resp = &interfaces.SemanticSearchResponse{
		QueryUnderstanding: queryUnderstandResult,
		KnowledgeConcepts:  rerankConceptResults,
		HitsTotal:          len(conceptResults),
	}
	return
}

// deduplicateConcepts 概念结果去重: 根据ID、Type去重
func (k *knRetrievalServiceImpl) deduplicateConcepts(concepts []*interfaces.ConceptResult) []*interfaces.ConceptResult {
	seen := make(map[string]bool)
	unique := make([]*interfaces.ConceptResult, 0)
	for _, c := range concepts {
		uniqueKey := fmt.Sprintf("%s:%s", c.ConceptType, c.ConceptID)
		if !seen[uniqueKey] {
			seen[uniqueKey] = true
			unique = append(unique, c)
		}
	}
	return unique
}

// 根据搜索与配置对查询策略进行过滤
func (k *knRetrievalServiceImpl) filterQueryStrategysBySearchScope(queryStrategys []*interfaces.SemanticQueryStrategy, searchScope *interfaces.SearchScopeConfig) []*interfaces.SemanticQueryStrategy {
	// 过滤后的查询策略
	filteredQueryStrategys := make([]*interfaces.SemanticQueryStrategy, 0)
	for _, queryStrategy := range queryStrategys {
		if queryStrategy.Filter != nil {
			switch queryStrategy.Filter.ConceptType {
			case interfaces.KnConceptTypeObject:
				if !*searchScope.IncludeObjectTypes {
					continue
				}
			case interfaces.KnConceptTypeRelation:
				if !*searchScope.IncludeRelationTypes {
					continue
				}
			case interfaces.KnConceptTypeAction:
				if !*searchScope.IncludeActionTypes {
					continue
				}
			}
		}
		filteredQueryStrategys = append(filteredQueryStrategys, queryStrategy)
	}
	return filteredQueryStrategys
}
