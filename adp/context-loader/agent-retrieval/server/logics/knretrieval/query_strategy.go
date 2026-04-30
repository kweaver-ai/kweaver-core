// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knretrieval

import (
	"context"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

const (
	// dataSampleLimit 数据采样数量
	dataSampleLimit = 2
)

// parallelExecSemanticQueryStrategy 执行召回策略（并发）
func (k *knRetrievalServiceImpl) parallelExecSemanticQueryStrategy(ctx context.Context,
	knID string, strategys []*interfaces.SemanticQueryStrategy,
) ([]*interfaces.ConceptResult, error) {
	var wg sync.WaitGroup
	resultChan := make(chan []*interfaces.ConceptResult, len(strategys))
	errChan := make(chan error, len(strategys))

	for _, strategy := range strategys {
		wg.Add(1)
		go func(s *interfaces.SemanticQueryStrategy) {
			defer wg.Done()
			res, err := k.execSemanticQueryStrategy(ctx, knID, s)
			if err != nil {
				errChan <- err
				return
			}
			resultChan <- res
		}(strategy)
	}

	wg.Wait()
	close(resultChan)
	close(errChan)

	var allResults []*interfaces.ConceptResult
	for res := range resultChan {
		allResults = append(allResults, res...)
	}

	var errors []error
	for err := range errChan {
		errors = append(errors, err)
	}

	if len(allResults) == 0 && len(errors) > 0 {
		return nil, errors[0]
	}

	if len(errors) > 0 {
		k.logger.Error("semantic query strategies execution errors",
			"error_count", len(errors),
			"first_error", errors[0].Error())
	}
	return allResults, nil
}

// execSemanticQueryStrategy 执行召回策略： 不同策略模版执行 --- 单策略执行
func (k *knRetrievalServiceImpl) execSemanticQueryStrategy(ctx context.Context,
	knID string, strategy *interfaces.SemanticQueryStrategy,
) (result []*interfaces.ConceptResult, err error) {
	switch strategy.StrategyType {
	case interfaces.ConceptDiscoveryStrategy: // 概念发现
		return k.execConceptDiscoveryStrategy(ctx, knID, strategy)
	case interfaces.ObjectInstanceDiscoveryStrategy: // 对象实例查找
		return k.execObjectInstanceDiscoveryStrategy(ctx, knID, strategy)
	case interfaces.ConceptGetStrategy: // 概念获取
		return k.execConceptGetStrategy(ctx, knID, strategy)
	}
	return
}

func (k *knRetrievalServiceImpl) execObjectInstanceDiscoveryStrategy(ctx context.Context,
	knID string, strategy *interfaces.SemanticQueryStrategy,
) (conceptResults []*interfaces.ConceptResult, err error) {
	if strategy.Filter == nil || strategy.Filter.ConceptID == "" {
		return
	}
	req := &interfaces.QueryObjectInstancesReq{
		KnID:               knID,
		OtID:               strategy.Filter.ConceptID,
		IncludeTypeInfo:    true,
		IncludeLogicParams: true,
		Limit:              dataSampleLimit,
	}
	// todo: condition转换待实现

	resp, err := k.ontologyQueryAccess.QueryObjectInstances(ctx, req)
	if err != nil {
		return nil, err
	}

	conceptResults = []*interfaces.ConceptResult{}
	if resp != nil && resp.ObjectConcept != nil {
		conceptResult := interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeObject,
			ConceptDetail: resp.ObjectConcept,
			Samples:       resp.Data,
		}
		if id, ok := resp.ObjectConcept[string(interfaces.ConceptFieldID)].(string); ok {
			conceptResult.ConceptID = id
		}
		if name, ok := resp.ObjectConcept[string(interfaces.ConceptFieldName)].(string); ok {
			conceptResult.ConceptName = name
		}
		conceptResults = append(conceptResults, &conceptResult)
	}
	return
}

// execConceptGetStrategy 概念获取策略
func (k *knRetrievalServiceImpl) execConceptGetStrategy(ctx context.Context,
	knID string, strategy *interfaces.SemanticQueryStrategy,
) (conceptResults []*interfaces.ConceptResult, err error) {
	if strategy.Filter == nil {
		return
	}
	filter := strategy.Filter
	var ConceptIDs []string
	if filter.ConceptID != "" {
		ConceptIDs = append(ConceptIDs, filter.ConceptID)
	}

	if len(filter.ConceptIDs) > 0 {
		ConceptIDs = append(ConceptIDs, filter.ConceptIDs...)
	}

	if len(ConceptIDs) == 0 {
		return
	}

	conceptDetailsMap := map[interfaces.KnConceptType][]any{}
	switch filter.ConceptType {
	case interfaces.KnConceptTypeObject:
		var objectDetails []*interfaces.ObjectType
		objectDetails, err = k.bknBackendAccess.GetObjectTypeDetail(ctx, knID, ConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[execConceptGetStrategy] execConceptGetStrategy failed. knId:%s, objectConceptIDs:%v\n",
				knID, ConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeObject] = append(conceptDetailsMap[interfaces.KnConceptTypeObject], objectDetails)
	case interfaces.KnConceptTypeRelation:
		var relationDetails []*interfaces.RelationType
		relationDetails, err = k.bknBackendAccess.GetRelationTypeDetail(ctx, knID, ConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[execConceptGetStrategy] execConceptGetStrategy failed. knId:%s, relationConceptIDs:%v\n",
				knID, ConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeObject] = append(conceptDetailsMap[interfaces.KnConceptTypeObject], relationDetails)
	case interfaces.KnConceptTypeAction:
		var actionDetails []*interfaces.ActionType
		actionDetails, err = k.bknBackendAccess.GetActionTypeDetail(ctx, knID, ConceptIDs, true)
		if err != nil {
			k.logger.WithContext(ctx).Errorf("[execConceptGetStrategy] execConceptGetStrategy failed. knId:%s, actionConceptIDs:%v\n",
				knID, ConceptIDs)
			return
		}
		conceptDetailsMap[interfaces.KnConceptTypeObject] = append(conceptDetailsMap[interfaces.KnConceptTypeObject], actionDetails)
	}

	if err != nil {
		k.logger.Errorf("[execConceptGetStrategy] getDetail failed. knId:%s, ConceptIDs:%v\n",
			knID, ConceptIDs)
		return
	}

	if len(conceptDetailsMap) == 0 {
		return
	}

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

// execConceptDiscoveryStrategy 执行概念发现策略
func (k *knRetrievalServiceImpl) execConceptDiscoveryStrategy(ctx context.Context,
	knID string, strategy *interfaces.SemanticQueryStrategy,
) (conceptResults []*interfaces.ConceptResult, err error) {
	if strategy.Filter == nil {
		return
	}
	filter := strategy.Filter
	if len(filter.Conditions) == 0 {
		return
	}
	conceptSearchConfig := config.NewConfigLoader().ConceptSearchConfig
	var subCond []*interfaces.KnCondition
	for _, fCond := range filter.Conditions {
		var operationType interfaces.KnOperationType
		operationType, err = ParseKnOperationType(fCond.Operation)
		if err != nil {
			k.logger.Warnf("[execConceptDiscoveryStrategy],ParseKnOperationType faild, strategy operation: %v", fCond.Operation)
			continue
		}
		knCond := &interfaces.KnCondition{
			Field:     fCond.Field,
			Operation: operationType,
			Value:     fCond.Value,
			ValueFrom: interfaces.CondValueFromConst,
		}

		if operationType == interfaces.KnOperationTypeKnn {
			knCond.Value = fCond.Value
			knCond.LimitKey = interfaces.CondLimitKeyK
			knCond.LimitValue = conceptSearchConfig.KnnKValue
		}

		subCond = append(subCond, knCond)
	}

	if len(subCond) == 0 {
		k.logger.Warnf("[execConceptDiscoveryStrategy], parse condition is empty, strategy: %v", strategy)
		return
	}

	cond := &interfaces.KnCondition{
		Operation:     interfaces.KnOperationTypeOr,
		SubConditions: subCond,
	}

	queryConceptsReq := &interfaces.QueryConceptsReq{
		KnID:  knID,
		Cond:  cond,
		Limit: conceptSearchConfig.ConceptRecallSize,
	}

	switch filter.ConceptType {
	case interfaces.KnConceptTypeObject:
		conceptResults, err = k.discoveryObjectConcepts(ctx, queryConceptsReq)
	case interfaces.KnConceptTypeRelation:
		conceptResults, err = k.discoveryRelationTypeConcepts(ctx, queryConceptsReq)
	case interfaces.KnConceptTypeAction:
		conceptResults, err = k.discoveryActionTypeConcepts(ctx, queryConceptsReq)
	}

	return
}

// discoveryObjectConcepts 发现对象类概念
func (k *knRetrievalServiceImpl) discoveryObjectConcepts(ctx context.Context,
	queryConceptsReq *interfaces.QueryConceptsReq,
) (conceptResults []*interfaces.ConceptResult, err error) {
	var objectTypes *interfaces.ObjectTypeConcepts
	objectTypes, err = k.bknBackendAccess.SearchObjectTypes(ctx, queryConceptsReq)
	if err != nil {
		k.logger.Errorf("[discoveryObjectConcepts] SearchObjectTypes failed, userId: %s, visitorType: %s, req: %v", queryConceptsReq)
		return
	}

	if objectTypes == nil {
		return
	}

	if len(objectTypes.Entries) == 0 {
		return
	}

	conceptResults = []*interfaces.ConceptResult{}
	for _, entry := range objectTypes.Entries {
		conceptResult := interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeObject,
			ConceptDetail: entry,
		}
		conceptResult.ConceptID = entry.ID
		conceptResult.ConceptName = entry.Name
		conceptResult.MatchScore = entry.Score
		conceptResults = append(conceptResults, &conceptResult)
	}
	return
}

// discoveryRelationTypeConcepts 发现关系类概念
func (k *knRetrievalServiceImpl) discoveryRelationTypeConcepts(ctx context.Context,
	queryConceptsReq *interfaces.QueryConceptsReq,
) (conceptResults []*interfaces.ConceptResult, err error) {
	var relationTypes *interfaces.RelationTypeConcepts
	relationTypes, err = k.bknBackendAccess.SearchRelationTypes(ctx, queryConceptsReq)
	if err != nil {
		k.logger.Errorf("[discoveryObjectConcepts] SearchRelationTypes failed, userId: %s, visitorType: %s, req: %v", queryConceptsReq)
		return
	}

	if relationTypes == nil {
		return
	}

	if len(relationTypes.Entries) == 0 {
		return
	}

	conceptResults = []*interfaces.ConceptResult{}
	for _, entry := range relationTypes.Entries {
		conceptResult := interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeRelation,
			ConceptDetail: entry,
		}
		conceptResult.ConceptID = entry.ID
		conceptResult.ConceptName = entry.Name
		conceptResult.MatchScore = entry.Score
		conceptResults = append(conceptResults, &conceptResult)
	}
	return
}

// discoveryActionTypeConcepts 发现行动类概念
func (k *knRetrievalServiceImpl) discoveryActionTypeConcepts(ctx context.Context,
	queryConceptsReq *interfaces.QueryConceptsReq,
) (conceptResults []*interfaces.ConceptResult, err error) {
	var actionTypes *interfaces.ActionTypeConcepts
	actionTypes, err = k.bknBackendAccess.SearchActionTypes(ctx, queryConceptsReq)
	if err != nil {
		k.logger.Errorf("[discoveryActionTypeConcepts] SearchActionTypes failed, userId: %s, visitorType: %s, req: %v", queryConceptsReq)
		return
	}

	if actionTypes == nil {
		return
	}

	if len(actionTypes.Entries) == 0 {
		return
	}

	conceptResults = []*interfaces.ConceptResult{}
	for _, entry := range actionTypes.Entries {
		conceptResult := interfaces.ConceptResult{
			ConceptType:   interfaces.KnConceptTypeAction,
			ConceptDetail: entry,
		}
		conceptResult.ConceptID = entry.ID
		conceptResult.ConceptName = entry.Name
		conceptResult.MatchScore = entry.Score
		conceptResults = append(conceptResults, &conceptResult)
	}
	return
}

// generateQueryStrategysForPlanB 直接拼接查询策略
/*
查询策略生成：
1. 详细策略
2. 精准策略
*/
func (k *knRetrievalServiceImpl) generateQueryStrategysForPlanB(queryUnderstandResult *interfaces.QueryUnderstanding) (queryStrategys []*interfaces.SemanticQueryStrategy) {
	queryStrategys = []*interfaces.SemanticQueryStrategy{}
	intents := queryUnderstandResult.Intent
	if len(intents) == 0 {
		return queryStrategys
	}

	objectTypeIds := []string{}
	relationTypeIds := []string{}
	actionTypeIds := []string{}

	// 简略意图遍历
	for _, intent := range intents {
		if len(intent.RelatedConcepts) == 0 { // 查找相关概念
			break
		}
		var (
			hasObjectType   bool
			hasRelationType bool
			hasActionType   bool
		)
		// 相关概念提取: 收集相关类ID信息
		for _, relatedConcept := range intent.RelatedConcepts {
			if relatedConcept.ConceptType == interfaces.KnConceptTypeObject { // 对象类id提取
				objectTypeIds = append(objectTypeIds, relatedConcept.ConceptID)
				hasObjectType = true
			}

			if relatedConcept.ConceptType == interfaces.KnConceptTypeRelation { // 关系类id提取
				relationTypeIds = append(relationTypeIds, relatedConcept.ConceptID)
				hasRelationType = true
			}

			if relatedConcept.ConceptType == interfaces.KnConceptTypeAction { // 行动类id提取
				actionTypeIds = append(actionTypeIds, relatedConcept.ConceptID)
				hasActionType = true
			}
		}
		/*
			是否需要进一步推理：进一步推理: 收集详细查询策略 （仅收集策略查询请求块）
		*/
		if intent.RequiresReasoning {
			// 详细信息提取: 构建概念发现查询策略
			var empty []*interfaces.QueryStrategyCondition
			// 增加模糊检索
			if hasObjectType { // 对象类模糊检索
				objectTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeObject, intent.QuerySegment, empty)
				if objectTypeDiscoveryStrategy != nil {
					queryStrategys = append(queryStrategys, objectTypeDiscoveryStrategy)
				}
			}

			if hasRelationType { // 关系类模糊检索
				releationTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeRelation, intent.QuerySegment, empty)
				if releationTypeDiscoveryStrategy != nil {
					queryStrategys = append(queryStrategys, releationTypeDiscoveryStrategy)
				}
			}

			if hasActionType { // 行动类模糊检索
				actionTypeDiscoveryStrategy := k.buildConceptDiscoveryStrategy(interfaces.KnConceptTypeAction, intent.QuerySegment, empty)
				if actionTypeDiscoveryStrategy != nil {
					queryStrategys = append(queryStrategys, actionTypeDiscoveryStrategy)
				}
			}
		}
	}

	/*精准策略查询块收集*/

	// 对象类概念精确获取策略
	if len(objectTypeIds) > 0 {
		objectTypeGetStrategy := k.buildConceptGetStrategy(interfaces.KnConceptTypeObject, objectTypeIds)
		if objectTypeGetStrategy != nil {
			queryStrategys = append(queryStrategys, objectTypeGetStrategy)
		}
	}

	// 关系类概念精确获取策略
	if len(relationTypeIds) > 0 {
		relationTypeGetStrategy := k.buildConceptGetStrategy(interfaces.KnConceptTypeRelation, relationTypeIds)
		if relationTypeGetStrategy != nil {
			queryStrategys = append(queryStrategys, relationTypeGetStrategy)
		}
	}

	// 行动类概念精确获取策略
	if len(actionTypeIds) > 0 {
		actionTypeGetStrategy := k.buildConceptGetStrategy(interfaces.KnConceptTypeAction, actionTypeIds)
		if actionTypeGetStrategy != nil {
			queryStrategys = append(queryStrategys, actionTypeGetStrategy)
		}
	}

	return queryStrategys
}

// buildConceptGetStrategy 构建概念精确获取查询策略
func (k *knRetrievalServiceImpl) buildConceptGetStrategy(conceptType interfaces.KnConceptType, conceptIDs []string) (queryStrategy *interfaces.SemanticQueryStrategy) {
	if len(conceptIDs) == 0 {
		return nil
	}
	conceptGetStrategy := &interfaces.SemanticQueryStrategy{
		StrategyType: interfaces.ConceptGetStrategy,
		Filter: &interfaces.QueryStrategyFilter{
			ConceptType: conceptType,
		},
	}
	conceptGetStrategy.Filter.ConceptIDs = append(conceptGetStrategy.Filter.ConceptIDs, conceptIDs...)

	return conceptGetStrategy
}

// buildConceptDiscoveryStrategy 构建概念发现查询策略
func (k *knRetrievalServiceImpl) buildConceptDiscoveryStrategy(conceptType interfaces.KnConceptType,
	query string, otherConds []*interfaces.QueryStrategyCondition,
) (queryStrategy *interfaces.SemanticQueryStrategy) {
	conds := []*interfaces.QueryStrategyCondition{}
	// 根据原始Query切分的片段构建查询策略
	if query != "" {
		// matchCondition 关键词匹配条件
		matchCondition := &interfaces.QueryStrategyCondition{
			Field:     string(interfaces.ConceptFieldAny),
			Operation: string(interfaces.KnOperationTypeMatch),
			Value:     query,
		}
		conds = append(conds, matchCondition)

		// knnCondition 向量检索条件
		knnCondition := &interfaces.QueryStrategyCondition{
			Field:     string(interfaces.ConceptFieldAny),
			Operation: string(interfaces.KnOperationTypeKnn),
			Value:     query,
		}
		conds = append(conds, knnCondition)
	}
	// otherConds 其他条件
	if len(otherConds) > 0 {
		conds = append(conds, otherConds...)
	}

	if len(conds) == 0 {
		return nil
	}

	// 构建概念发现查询策略
	discoveryStrategy := &interfaces.SemanticQueryStrategy{
		StrategyType: interfaces.ConceptDiscoveryStrategy,
		Filter: &interfaces.QueryStrategyFilter{
			ConceptType: conceptType,
		},
	}
	discoveryStrategy.Filter.Conditions = conds

	return discoveryStrategy
}
