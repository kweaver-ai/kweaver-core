// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"sort"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// rerankByConceptType 收集不同概念类集合，并进行排序，每个概念集取前limit个
//
//nolint:unused // 预留函数，后续可能使用
func (k *knRetrievalServiceImpl) rerankByConceptType(conceptResults []*interfaces.ConceptResult, limit int) []*interfaces.ConceptResult {
	// 去重
	conceptResults = k.deduplicateConcepts(conceptResults)
	conceptTypeMap := make(map[interfaces.KnConceptType][]*interfaces.ConceptResult)
	// 按概念类型分类
	for _, concept := range conceptResults {
		conceptTypeMap[concept.ConceptType] = append(conceptTypeMap[concept.ConceptType], concept)
	}
	// 按概念类型排序
	for _, concepts := range conceptTypeMap {
		sort.Slice(concepts, func(i, j int) bool {
			return concepts[i].MatchScore > concepts[j].MatchScore
		})
		if len(concepts) > limit {
			concepts = concepts[:limit]
		}
	}
	result := []*interfaces.ConceptResult{}
	// 顺序要求：对象类、关系类、行动类
	if len(conceptTypeMap[interfaces.KnConceptTypeObject]) > 0 {
		result = append(result, conceptTypeMap[interfaces.KnConceptTypeObject]...)
	}
	if len(conceptTypeMap[interfaces.KnConceptTypeRelation]) > 0 {
		result = append(result, conceptTypeMap[interfaces.KnConceptTypeRelation]...)
	}
	if len(conceptTypeMap[interfaces.KnConceptTypeAction]) > 0 {
		result = append(result, conceptTypeMap[interfaces.KnConceptTypeAction]...)
	}
	if len(result) > limit {
		result = result[:limit]
	}
	return result
}

func (k *knRetrievalServiceImpl) rerankByDataRetrieval(ctx context.Context, queryUnderstandResult *interfaces.QueryUnderstanding, conceptResults []*interfaces.ConceptResult,
	action interfaces.KnowledgeRerankActionType, limit int) (rerankResults []*interfaces.ConceptResult, err error) {
	// 去重
	conceptResults = k.deduplicateConcepts(conceptResults)

	// 优化1：如果没有概念，直接返回空列表，无需调用 rerank
	if len(conceptResults) == 0 {
		k.logger.WithContext(ctx).Debug("[knretrieval#rerank] No concepts to rerank, returning empty list")
		return []*interfaces.ConceptResult{}, nil
	}

	if action == interfaces.KnowledgeRerankActionDefault {
		rerankResults = conceptResults
	} else if k.useLocalRerank {
		// 使用本地Rerank模块
		k.logger.WithContext(ctx).Info("[knretrieval#rerank] Using local KnowledgeReranker")
		rerankResults, err = k.knReranker.Rerank(ctx, &interfaces.KnowledgeRerankReq{
			QueryUnderstanding: queryUnderstandResult,
			KnowledgeConcepts:  conceptResults,
			Action:             action,
		})
		if err != nil {
			// 优化2：本地 rerank 失败时，直接返回原始概念列表（降级），不再调用远程服务
			k.logger.WithContext(ctx).Warnf("[knretrieval#rerank] Local rerank failed: %v, using original concepts as fallback", err)
			rerankResults = conceptResults
			err = nil // 清除错误，确保不影响核心功能
		}
	} else {
		// 使用原有远程调用
		rerankResults, err = k.dataRetrieval.KnowledgeRerank(ctx, &interfaces.KnowledgeRerankReq{
			QueryUnderstanding: queryUnderstandResult,
			KnowledgeConcepts:  conceptResults,
			Action:             action,
		})
		if err != nil {
			// 远程 rerank 失败时，也降级到返回原始概念列表
			k.logger.WithContext(ctx).Warnf("[knretrieval#rerank] Remote rerank failed: %v, using original concepts as fallback", err)
			rerankResults = conceptResults
			err = nil // 清除错误，确保不影响核心功能
		}
	}
	// 按 RerankScore 降序、相同时按 MatchScore 降序排序（不再过滤 RerankScore=0，避免 concepts 为 null）
	rerankResults = k.sortByRerankAndMatchScore(rerankResults)
	// 分页
	if len(rerankResults) > limit {
		rerankResults = rerankResults[:limit]
	}
	return
}

// sortByRerankAndMatchScore 按 RerankScore 降序排序，相同时按 MatchScore 降序
func (k *knRetrievalServiceImpl) sortByRerankAndMatchScore(conceptResults []*interfaces.ConceptResult) []*interfaces.ConceptResult {
	if conceptResults == nil {
		return nil
	}
	sort.Slice(conceptResults, func(i, j int) bool {
		if conceptResults[i].RerankScore != conceptResults[j].RerankScore {
			return conceptResults[i].RerankScore > conceptResults[j].RerankScore
		}
		return conceptResults[i].MatchScore > conceptResults[j].MatchScore
	})
	return conceptResults
}
