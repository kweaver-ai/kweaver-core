// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"sync"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/panjf2000/ants"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// 概念召回策略处理并发数
	conceptRecallConcurrency = 10
)

// AgentIntentRetrieval 语义检索: 基于意图分析智能体 + 传统召回策略
/*
1. 请求"概念意图分析智能体（Agent）"获取意图相关概念概览信息候选集 -- 意图粗识别
2. 上一步会返回一个候选集列表，根据列表中每个候选集requires_reasoning字段判断时候需要“获取候选集详细信息”
	a. requires_reasoning 为 false, 不需要进一步推理，可以直接根据意图中的相关概念请求业务知识网络获取详情，并加入候选集
	b. requires_reasoning 为 true, 需要进一步推理，请求“概念召回策略智能体(Agent)”获取召回的策略
	c. 执行召回策略
		I. 概念召回策略智能体(Agent)返回的策略信息
		II. 根据原始Query构建的策略信息
	d. 调用工具进行排序
	e. 请求业务知识网络接口获取示例采样
	f. 输出结果
*/
func (k *knRetrievalServiceImpl) AgentIntentRetrieval(ctx context.Context, req *interfaces.SemanticSearchRequest) (resp *interfaces.SemanticSearchResponse, err error) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 请求“概念意图分析智能体（Agent）”获取意图相关概念概览信息候选集 -- 意图粗识别
	queryUnderstanding, err := k.agentClient.ConceptIntentionAnalysisAgent(ctx, &interfaces.ConceptIntentionAnalysisAgentReq{
		PreviousQueries: req.PreviousQueries,
		Query:           req.Query,
		KnID:            req.KnID,
	})
	if err != nil {
		k.logger.WithContext(ctx).Warnf("SemanticSearchV2 ConceptIntentionAnalysisAgent err, err: %v", err)
	}
	if queryUnderstanding == nil {
		queryUnderstanding = &interfaces.QueryUnderstanding{}
	}
	k.logger.WithContext(ctx).Infof("SemanticSearchV2 ConceptIntentionAnalysisAgent resp, resp: %v", queryUnderstanding)
	// 查询策略
	var queryStrategys []*interfaces.SemanticQueryStrategy
	// 概念结果候选集
	conceptResults := []*interfaces.ConceptResult{}
	if len(queryUnderstanding.Intent) > 0 {
		// 意图识别不为空，根据意图识别结果获取概念结果候选集
		var intentRecallConceptResults []*interfaces.ConceptResult
		intentRecallConceptResults, queryStrategys, err = k.intentRecall(ctx, req, queryUnderstanding.Intent)
		if err != nil {
			k.logger.WithContext(ctx).Warnf("SemanticSearchV2 intentRecall err, err: %v", err)
		}
		if len(intentRecallConceptResults) > 0 {
			conceptResults = append(conceptResults, intentRecallConceptResults...)
		}
	}
	if len(queryStrategys) == 0 {
		// 当查询策略为空时，自定义构建查询策略，请求业务知识网络接口做关键词匹配
		queryStrategys = k.longtailRecallByKnowledgeNetwork(req.Query)
	}
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
		// 返回执行的策略
		queryUnderstanding.QueryStrategys = queryStrategys
	}
	// TODO：实例数据采样（本版本跳过）
	// 排序：按概念类型排序, 去重
	rerankConceptResults, err := k.rerankConcepts(ctx, queryUnderstanding, conceptResults, req.RerankAction, req.MaxConcepts)
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

// 根据意图中匹配的相关概念，获取概念详情,并组装概念结果候选集
func (k *knRetrievalServiceImpl) intentRecallGetConceptResultsByIntent(ctx context.Context, knID string,
	intent *interfaces.SemanticQueryIntent,
) (conceptResults []*interfaces.ConceptResult, err error) {
	if intent == nil || len(intent.RelatedConcepts) == 0 {
		return
	}
	// 对象类概念ID列表
	var objectConceptIDs []string
	// 行动类概念ID列表
	var actionConceptIDs []string
	// 关系类概念ID列表
	var relationConceptIDs []string

	// 分类概念ID列表，key为概念类型，value为概念ID列表
	conceptDetailsMap := map[interfaces.KnConceptType][]any{}
	for _, concept := range intent.RelatedConcepts {
		switch concept.ConceptType {
		case interfaces.KnConceptTypeObject:
			// 加入对象类概念ID列表
			objectConceptIDs = append(objectConceptIDs, concept.ConceptID)
			conceptDetailsMap[interfaces.KnConceptTypeObject] = []any{}
		case interfaces.KnConceptTypeAction:
			actionConceptIDs = append(actionConceptIDs, concept.ConceptID)
			conceptDetailsMap[interfaces.KnConceptTypeAction] = []any{}
		case interfaces.KnConceptTypeRelation:
			relationConceptIDs = append(relationConceptIDs, concept.ConceptID)
			conceptDetailsMap[interfaces.KnConceptTypeRelation] = []any{}
		}
	}

	// 查询对象类概念详情
	if len(objectConceptIDs) > 0 {
		var objectDetails []*interfaces.ObjectType
		objectDetails, err = k.bknBackendAccess.GetObjectTypeDetail(ctx, knID, objectConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[getConceptDetails] getObjectTypeDetail failed. knId:%s, objectConceptIDs:%v\n",
				knID, objectConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeObject] = append(conceptDetailsMap[interfaces.KnConceptTypeObject], objectDetails)
	}
	// 查询行动类概念详情
	if len(actionConceptIDs) > 0 {
		var actionDetails []*interfaces.ActionType
		actionDetails, err = k.bknBackendAccess.GetActionTypeDetail(ctx, knID, actionConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[getConceptDetails] getActionTypeDetail failed. knId:%s, actionConceptIDs:%v\n",
				knID, actionConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeAction] = append(conceptDetailsMap[interfaces.KnConceptTypeAction], actionDetails)
	}
	// 查询关系类概念详情
	if len(relationConceptIDs) > 0 {
		var relationDetails []*interfaces.RelationType
		relationDetails, err = k.bknBackendAccess.GetRelationTypeDetail(ctx, knID, relationConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[getConceptDetails] getRelationTypeDetail failed. knId:%s, relationConceptIDs:%v\n",
				knID, relationConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeRelation] = append(conceptDetailsMap[interfaces.KnConceptTypeRelation], relationDetails)
	}
	// 概念集合处理成候选集
	conceptResults = []*interfaces.ConceptResult{}
	for conceptType, conceptDetails := range conceptDetailsMap {
		if len(conceptDetails) == 0 {
			continue
		}
		for _, conceptDetail := range conceptDetails {
			conceptResult := &interfaces.ConceptResult{
				ConceptType:   conceptType,
				ConceptDetail: conceptDetail,
			}
			switch conceptType {
			case interfaces.KnConceptTypeObject:
				concept := conceptDetail.(*interfaces.ObjectType)
				conceptResult.ConceptID = concept.ID
				conceptResult.ConceptName = concept.Name
				conceptResult.MatchScore = interfaces.MaxMatchScore
			case interfaces.KnConceptTypeAction:
				concept := conceptDetail.(*interfaces.ActionType)
				conceptResult.ConceptID = concept.ID
				conceptResult.ConceptName = concept.Name
				conceptResult.MatchScore = interfaces.MaxMatchScore
			case interfaces.KnConceptTypeRelation:
				concept := conceptDetail.(*interfaces.RelationType)
				conceptResult.ConceptID = concept.ID
				conceptResult.ConceptName = concept.Name
				conceptResult.MatchScore = interfaces.MaxMatchScore
			}
			conceptResults = append(conceptResults, conceptResult)
		}
	}
	return
}

// 基于意图召回
func (k *knRetrievalServiceImpl) intentRecall(ctx context.Context, req *interfaces.SemanticSearchRequest, intents []*interfaces.SemanticQueryIntent) (
	conceptResults []*interfaces.ConceptResult, queryStrategys []*interfaces.SemanticQueryStrategy, err error,
) {
	// 记录可观测
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)
	// 并发处理意图
	var (
		resultsLock    sync.Mutex
		strategiesLock sync.Mutex
	)
	wg := sync.WaitGroup{}
	pool, err := ants.NewPoolWithFunc(conceptRecallConcurrency, func(i interface{}) {
		intent, ok := i.(*interfaces.SemanticQueryIntent)
		if !ok {
			k.logger.Errorf("[SemanticSearchV2] NewPoolWithFunc failed. intent type error. intent:%v", intent)
			wg.Done()
			return
		}
		// 筛选意图识别结果，收集需要获取补充信息的意图requires_reasoning为 true
		// 直接加入“概念结果候选集”的意图，requires_reasoning为 false
		var currentConceptResults []*interfaces.ConceptResult // 当前意图的概念结果候选集
		currentConceptResults, err = k.intentRecallGetConceptResultsByIntent(ctx, req.KnID, intent)
		if err != nil {
			k.logger.Errorf("[SemanticSearchV2] intentRecallGetConceptResultsByIntent failed. knId:%s, intent:%v, err:%v", req.KnID, intent, err)
			wg.Done()
			return
		}
		// 判断是否需要进一步推理
		if !intent.RequiresReasoning {
			// 直接加入“概念结果候选集”
			resultsLock.Lock()
			conceptResults = append(conceptResults, currentConceptResults...)
			resultsLock.Unlock()
			wg.Done()
			return
		}
		// 请求”概念召回策略智能体（Agent）“，根据意图识别及相关概念详情生成查询规划（策略）
		var retrievalStrategys []*interfaces.SemanticQueryStrategy
		retrievalStrategys, err = k.agentClient.ConceptRetrievalStrategistAgent(ctx, &interfaces.ConceptRetrievalStrategistReq{
			QueryParam: &interfaces.ConceptRetrievalStrategistQueryParam{
				OriginalQuery:        req.Query,
				CurrentIntentSegment: intent,
				ConceptCandidates:    currentConceptResults,
			},
			KnID:            req.KnID,
			PreviousQueries: req.PreviousQueries,
		})
		if err != nil {
			k.logger.WithContext(ctx).Warnf("[SemanticSearchV2] ConceptRetrievalStrategistAgent failed. knId:%s, intent:%v, err:%v", req.KnID, intent, err)
			wg.Done()
			return
		}
		if len(retrievalStrategys) > 0 {
			// 加入“查询策略候选集”
			strategiesLock.Lock()
			queryStrategys = append(queryStrategys, retrievalStrategys...)
			strategiesLock.Unlock()
		}
		wg.Done()
	})
	if err != nil {
		k.logger.Errorf("[SemanticSearchV2] NewPoolWithFunc failed. err:%v", err)
		return
	}
	defer pool.Release()
	for _, intent := range intents {
		wg.Add(1)
		err = pool.Invoke(intent)
		if err != nil {
			k.logger.Errorf("[SemanticSearchV2] pool.Invoke failed. intent:%v, err:%v", intent, err)
			continue
		}
	}
	wg.Wait()
	return
}

// 长尾召回策略:基于业务知识网络做关键词匹配 -- 构建查询策略
func (k *knRetrievalServiceImpl) longtailRecallByKnowledgeNetwork(query string) (queryStrategys []*interfaces.SemanticQueryStrategy) {
	// 根据用户数据的原始Query生成查询策略
	var empty []*interfaces.QueryStrategyCondition
	objectTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeObject, query, empty)
	if objectTypeDiscoveryStrategy != nil {
		queryStrategys = append(queryStrategys, objectTypeDiscoveryStrategy)
	}
	releationTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeRelation, query, empty)
	if releationTypeDiscoveryStrategy != nil {
		queryStrategys = append(queryStrategys, releationTypeDiscoveryStrategy)
	}
	actionTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeAction, query, empty)
	if actionTypeDiscoveryStrategy != nil {
		queryStrategys = append(queryStrategys, actionTypeDiscoveryStrategy)
	}
	return
}
