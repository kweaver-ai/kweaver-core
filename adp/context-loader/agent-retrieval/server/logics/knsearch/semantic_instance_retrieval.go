// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch（语义实例召回）
// file: semantic_instance_retrieval.go
package knsearch

import (
	"context"
	"fmt"
	"sort"
	"strings"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// semanticInstanceRetrieval 语义实例召回主逻辑
// 流程：遍历对象类型 -> 向量检索 -> 打分与排序 -> 全局分数过滤 -> 属性过滤
func (s *localSearchImpl) semanticInstanceRetrieval(
	ctx context.Context,
	req *interfaces.KnSearchLocalRequest,
	objectTypes []*interfaces.KnSearchObjectType,
	config *interfaces.KnSearchRetrievalConfig,
) (*interfaces.KnSearchSemanticInstanceResult, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	if len(objectTypes) == 0 {
		return &interfaces.KnSearchSemanticInstanceResult{
			Message: "没有可检索的对象类型",
		}, nil
	}

	instanceConfig := config.SemanticInstanceRetrieval
	propertyConfig := config.PropertyFilter

	var allNodes []*interfaces.KnSearchNode
	var maxScore float64

	// 遍历每个对象类型进行语义检索
	for _, objType := range objectTypes {
		nodes, err := s.retrieveInstancesForObjectType(ctx, req, objType, instanceConfig)
		if err != nil {
			s.logger.WithContext(ctx).Warnf("[SemanticInstanceRetrieval] Failed to retrieve instances for %s: %v",
				objType.ConceptID, err)
			continue
		}

		// 更新最高分
		for _, node := range nodes {
			if node.Score > maxScore {
				maxScore = node.Score
			}
		}

		allNodes = append(allNodes, nodes...)
	}

	s.logger.WithContext(ctx).Infof("[SemanticInstanceRetrieval] Retrieved %d instances from %d object types, max_score=%.4f",
		len(allNodes), len(objectTypes), maxScore)

	// 全局分数过滤
	if boolValue(instanceConfig.EnableGlobalFinalScoreRatioFilter) && maxScore > 0 && len(allNodes) > 0 {
		threshold := maxScore * instanceConfig.GlobalFinalScoreRatio
		var topNode *interfaces.KnSearchNode
		for _, n := range allNodes {
			if topNode == nil || n.Score > topNode.Score {
				topNode = n
			}
		}
		allNodes = s.filterNodesByScore(allNodes, threshold)
		if len(allNodes) == 0 && topNode != nil {
			allNodes = []*interfaces.KnSearchNode{topNode}
			s.logger.WithContext(ctx).Debugf("[SemanticInstanceRetrieval] Global score filter kept at least one (top score=%.4f)", topNode.Score)
		}
		s.logger.WithContext(ctx).Debugf("[SemanticInstanceRetrieval] After global score filter (threshold=%.4f): %d nodes",
			threshold, len(allNodes))
	}

	// 属性过滤
	if boolValue(propertyConfig.EnablePropertyFilter) {
		allNodes = s.filterNodeProperties(allNodes, propertyConfig)
	}

	result := &interfaces.KnSearchSemanticInstanceResult{
		Nodes: allNodes,
	}

	if len(allNodes) == 0 {
		result.Message = "未检索到符合条件的实例数据"
	}

	return result, nil
}

// retrieveInstancesForObjectType 对单个对象类型进行语义检索
func (s *localSearchImpl) retrieveInstancesForObjectType(
	ctx context.Context,
	req *interfaces.KnSearchLocalRequest,
	objType *interfaces.KnSearchObjectType,
	config *interfaces.KnSearchSemanticInstanceRetrievalConfig,
) ([]*interfaces.KnSearchNode, error) {
	if len(findSemanticSearchableFields(objType)) == 0 {
		s.logger.WithContext(ctx).Infof("[SemanticInstanceRetrieval] Object type %s has no semantic-searchable properties, skip", objType.ConceptID)
		return nil, nil
	}

	cond := s.buildSemanticSearchConditionStruct(req.Query, objType, config)
	if cond == nil {
		return nil, nil
	}

	// 调用 ontology-query 进行实例检索
	queryReq := &interfaces.QueryObjectInstancesReq{
		KnID:               req.KnID,
		OtID:               objType.ConceptID,
		IncludeTypeInfo:    true,
		IncludeLogicParams: false,
		Limit:              config.InitialCandidateCount,
		Cond:               cond,
	}

	resp, err := s.ontologyQuery.QueryObjectInstances(ctx, queryReq)
	if err != nil {
		return nil, fmt.Errorf("query instances failed: %w", err)
	}

	// 转换为 KnSearchNode 格式
	nodes := make([]*interfaces.KnSearchNode, 0, len(resp.Data))
	for _, data := range resp.Data {
		if dataMap, ok := data.(map[string]any); ok {
			node := s.convertToKnSearchNode(objType, dataMap)
			nodes = append(nodes, node)
		}
	}

	// 计算相关性分数
	s.scoreNodes(req.Query, nodes, config)

	// 按分数降序排序
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Score > nodes[j].Score
	})

	// 取 Top-K
	if len(nodes) > config.PerTypeInstanceLimit {
		nodes = nodes[:config.PerTypeInstanceLimit]
	}

	// 过滤低相关性节点
	nodes = s.filterLowRelevanceNodes(nodes, config.MinDirectRelevance)

	return nodes, nil
}

// buildSemanticSearchConditionStruct 构建语义检索条件结构体
func (s *localSearchImpl) buildSemanticSearchConditionStruct(
	query string,
	objType *interfaces.KnSearchObjectType,
	config *interfaces.KnSearchSemanticInstanceRetrievalConfig,
) *interfaces.KnCondition {
	searchable := findSemanticSearchableFields(objType)
	if len(searchable) == 0 {
		return nil
	}
	maxSub := config.MaxSemanticSubConditions
	if maxSub <= 0 {
		maxSub = 10
	}
	knnLimit := config.PerTypeInstanceLimit
	if knnLimit <= 0 {
		knnLimit = 5
	}

	var subConditions []*interfaces.KnCondition

	for i := range searchable {
		if len(subConditions) >= maxSub {
			break
		}
		f := &searchable[i]
		if !f.HasKnn {
			continue
		}
		subConditions = append(subConditions, &interfaces.KnCondition{
			Field:      f.Name,
			Operation:  interfaces.KnOperationTypeKnn,
			Value:      query,
			ValueFrom:  interfaces.CondValueFromConst,
			LimitKey:   interfaces.CondLimitKeyK,
			LimitValue: knnLimit,
		})
	}
	for i := range searchable {
		if len(subConditions) >= maxSub {
			break
		}
		f := &searchable[i]
		if f.HasExactMatch {
			subConditions = append(subConditions, &interfaces.KnCondition{
				Field:     f.Name,
				Operation: interfaces.KnOperationTypeEqual,
				Value:     query,
				ValueFrom: interfaces.CondValueFromConst,
			})
			if len(subConditions) >= maxSub {
				break
			}
		}
		if f.HasMatch {
			subConditions = append(subConditions, &interfaces.KnCondition{
				Field:     f.Name,
				Operation: interfaces.KnOperationTypeMatch,
				Value:     query,
				ValueFrom: interfaces.CondValueFromConst,
			})
		}
	}

	if len(subConditions) > maxSub {
		subConditions = subConditions[:maxSub]
	}

	return &interfaces.KnCondition{
		Operation:     interfaces.KnOperationTypeOr,
		SubConditions: subConditions,
	}
}

// convertToKnSearchNode 将原始数据转换为 KnSearchNode 格式
func (s *localSearchImpl) convertToKnSearchNode(objType *interfaces.KnSearchObjectType, data map[string]any) *interfaces.KnSearchNode {
	node := &interfaces.KnSearchNode{
		ObjectTypeID:   objType.ConceptID,
		ObjectTypeName: objType.ConceptName,
		Properties:     make(map[string]any),
	}

	// 提取唯一标识
	if uid, ok := data["unique_identities"]; ok {
		if uidMap, ok := uid.(map[string]any); ok {
			node.UniqueIdentities = uidMap
		}
	}

	// 提取实例名称
	if name, ok := data["instance_name"]; ok {
		if nameStr, ok := name.(string); ok {
			node.InstanceName = nameStr
		}
	}

	// 提取其他属性
	for key, value := range data {
		if key != "unique_identities" && key != "instance_name" && key != "_score" {
			node.Properties[key] = value
		}
	}

	// 提取分数（如果有）
	if score, ok := data["_score"]; ok {
		switch v := score.(type) {
		case float64:
			node.Score = v
		case int:
			node.Score = float64(v)
		}
	}

	return node
}

// scoreNodes 计算节点的相关性分数
func (s *localSearchImpl) scoreNodes(query string, nodes []*interfaces.KnSearchNode, config *interfaces.KnSearchSemanticInstanceRetrievalConfig) {
	for _, node := range nodes {
		// 如果已有分数（来自向量检索），保留
		if node.Score > 0 {
			continue
		}

		if strings.TrimSpace(query) == "" {
			node.Score = 0
			continue
		}

		// 基于名称匹配计算分数
		score := 0.0

		// 完全匹配加高分
		if node.InstanceName == query {
			score = config.ExactNameMatchScore
		} else if containsFold(node.InstanceName, query) {
			score = 0.5
		} else if containsFold(query, node.InstanceName) {
			score = 0.3
		}

		node.Score = score
	}
}

// filterLowRelevanceNodes 过滤低相关性节点
func (s *localSearchImpl) filterLowRelevanceNodes(nodes []*interfaces.KnSearchNode, minRelevance float64) []*interfaces.KnSearchNode {
	var filtered []*interfaces.KnSearchNode
	for _, node := range nodes {
		if node.Score >= minRelevance {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// filterNodesByScore 按分数阈值过滤节点
func (s *localSearchImpl) filterNodesByScore(nodes []*interfaces.KnSearchNode, threshold float64) []*interfaces.KnSearchNode {
	var filtered []*interfaces.KnSearchNode
	for _, node := range nodes {
		if node.Score >= threshold {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

// filterNodeProperties 过滤节点属性
func (s *localSearchImpl) filterNodeProperties(nodes []*interfaces.KnSearchNode, config *interfaces.KnSearchPropertyFilterConfig) []*interfaces.KnSearchNode {
	for _, node := range nodes {
		if len(node.Properties) > config.MaxPropertiesPerInstance {
			keys := make([]string, 0, len(node.Properties))
			for key := range node.Properties {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			newProps := make(map[string]any)
			for i, key := range keys {
				if i >= config.MaxPropertiesPerInstance {
					break
				}
				newProps[key] = node.Properties[key]
			}
			node.Properties = newProps
		}

		// 截断过长的属性值
		for key, value := range node.Properties {
			if strVal, ok := value.(string); ok {
				if config.MaxPropertyValueLength > 0 {
					runes := []rune(strVal)
					if len(runes) > config.MaxPropertyValueLength {
						node.Properties[key] = string(runes[:config.MaxPropertyValueLength]) + "..."
					}
				}
			}
		}
	}
	return nodes
}
