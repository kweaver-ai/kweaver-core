// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knsearch（概念召回）
// file: concept_retrieval.go
package knsearch

import (
	"context"
	"sort"
	"strings"

	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// objectTypeRelationMultiplier 无关系/按关系过滤时对象类型数量相对 topK 的倍数
const objectTypeRelationMultiplier = 2

// conceptRetrieval 概念召回主逻辑
//
// 路由策略：
//   - 当 config.ConceptGroups 非空时，走 conceptRetrievalByGroups：直接调用 BKN
//     的 SearchObjectTypes/SearchRelationTypes/SearchActionTypes，由 BKN 在分组
//     范围内完成召回，跳过本地全量 GetKnowledgeNetworkDetail 与本地过滤。
//   - 当 config.ConceptGroups 为空时，沿用历史路径：拉全量网络详情 -> 可选粗召
//     回 -> 关系排序 -> 对象选择 -> 属性裁剪。
//
// 这样切分的原因：BKN concept_group 以 object_type 集合作为边界，
// relation_type/action_type 的分组范围由 BKN 按对象边界推导。ContextLoader
// 不复制这套分组推导逻辑，而是把 BKN typed search 作为事实来源。
func (s *localSearchImpl) conceptRetrieval(
	ctx context.Context,
	req *interfaces.KnSearchLocalRequest,
	config *interfaces.KnSearchConceptRetrievalConfig,
) (*interfaces.KnSearchConceptResult, error) {
	var err error
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, err)

	if len(config.ConceptGroups) > 0 {
		return s.conceptRetrievalByGroups(ctx, req, config)
	}

	networkDetail, err := s.bknBackend.GetKnowledgeNetworkDetail(ctx, req.KnID)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[ConceptRetrieval] GetKnowledgeNetworkDetail failed: %v", err)
		return nil, err
	}

	s.logger.WithContext(ctx).Debugf("[ConceptRetrieval] Network detail: object_types=%d, relation_types=%d, action_types=%d",
		len(networkDetail.ObjectTypes), len(networkDetail.RelationTypes), len(networkDetail.ActionTypes))

	// 2. 粗召回（可选，针对大规模知识网络）
	if boolValue(config.EnableCoarseRecall) && len(networkDetail.RelationTypes) >= config.CoarseMinRelationCount {
		s.logger.WithContext(ctx).Infof("[ConceptRetrieval] Enable coarse recall, relation_count=%d >= threshold=%d",
			len(networkDetail.RelationTypes), config.CoarseMinRelationCount)
		networkDetail, err = s.coarseRecall(ctx, req.KnID, req.Query, networkDetail, config)
		if err != nil {
			s.logger.WithContext(ctx).Warnf("[ConceptRetrieval] Coarse recall failed, continue with full schema: %v", err)
			// 粗召回失败不影响后续流程，继续使用完整 Schema
		}
	}

	// 3. 关系类型排序（基于语义相关性）并取 Top-K
	rankedRelations := s.rankRelationTypes(ctx, req.Query, networkDetail.ObjectTypes, networkDetail.RelationTypes, config.TopK, req.EnableRerank)
	s.logger.WithContext(ctx).Debugf("[ConceptRetrieval] Ranked relations: %d -> top_k=%d", len(networkDetail.RelationTypes), len(rankedRelations))

	// 4. 对象类型选择：按关系过滤 + 粗召回兜底补齐（以及无关系类型场景的排序截断）
	selectedObjects := s.selectObjectTypesForConceptRetrieval(networkDetail.ObjectTypes, rankedRelations, config.TopK)
	s.logger.WithContext(ctx).Debugf("[ConceptRetrieval] Selected objects: %d", len(selectedObjects))

	// 5. 转换为本地响应结构（与 Python schema_brief 语义一致）
	brief := boolValue(config.SchemaBrief)
	objectTypesLocal := s.convertObjectTypesToLocal(selectedObjects, brief)
	relationTypesLocal := s.convertRelationTypesToLocal(rankedRelations, brief)
	actionTypesLocal := s.convertActionTypesToLocal(networkDetail.ActionTypes, networkDetail.ID, networkDetail.ObjectTypes)

	// 7. 获取样例数据（可选）
	if boolValue(config.IncludeSampleData) && len(objectTypesLocal) > 0 {
		s.fetchSampleData(ctx, req.KnID, objectTypesLocal, boolValue(config.SchemaBrief))
	}

	return &interfaces.KnSearchConceptResult{
		ObjectTypes:   objectTypesLocal,
		RelationTypes: relationTypesLocal,
		ActionTypes:   actionTypesLocal,
	}, nil
}

// conceptRetrievalByGroups 是 concept_groups 非空场景下的概念召回路径。
//
// 设计要点：
//   - 直接调用 BKN 的 typed search API（SearchObjectTypes / SearchRelationTypes /
//     SearchActionTypes），由 BKN 完成 concept_groups 范围过滤；不再依赖
//     GetKnowledgeNetworkDetail 与本地过滤。
//   - 三个调用任一失败立即向上透传错误（包含分组不存在场景）。这里有意不把
//     BKN 错误（例如未知分组返回的 5xx + "all concept group not found ..."）
//     吞成空结果，调用方需要据此区分"分组合法但无概念"与"分组本身不存在"。
//     BKN 侧在分组未知时返回 5xx 是已知语义偏差，应通过独立 issue 推动改为 4xx，
//     ContextLoader 在切换前不做错误码翻译。
//   - relation/action 可能引用未被独立 object search 命中的对象；转换前批量
//     补齐这些对象详情，避免返回缺端点对象的 schema。
//   - 排序、对象选择、属性裁剪复用现有函数（rankRelationTypes /
//     selectObjectTypesForConceptRetrieval / convertXxxToLocal / fetchSampleData），
//     保证两条路径下的下游行为对齐。
func (s *localSearchImpl) conceptRetrievalByGroups(
	ctx context.Context,
	req *interfaces.KnSearchLocalRequest,
	config *interfaces.KnSearchConceptRetrievalConfig,
) (*interfaces.KnSearchConceptResult, error) {
	objectLimit := config.CoarseObjectLimit
	relationLimit := config.CoarseRelationLimit
	// Action 类型缺少专用 limit；BKN 端 action 数量级与 object 类似，复用
	// CoarseObjectLimit 即可，避免给配置面引入新字段。
	actionLimit := config.CoarseObjectLimit

	objectReq := s.buildCoarseRecallQuery(req.KnID, req.Query, objectLimit, config.ConceptGroups)
	objectResp, err := s.bknBackend.SearchObjectTypes(ctx, objectReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[ConceptRetrieval][Groups] SearchObjectTypes failed: %v", err)
		return nil, err
	}

	relationReq := s.buildCoarseRecallQuery(req.KnID, req.Query, relationLimit, config.ConceptGroups)
	relationResp, err := s.bknBackend.SearchRelationTypes(ctx, relationReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[ConceptRetrieval][Groups] SearchRelationTypes failed: %v", err)
		return nil, err
	}

	actionReq := s.buildCoarseRecallQuery(req.KnID, req.Query, actionLimit, config.ConceptGroups)
	actionResp, err := s.bknBackend.SearchActionTypes(ctx, actionReq)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[ConceptRetrieval][Groups] SearchActionTypes failed: %v", err)
		return nil, err
	}

	var (
		objects   []*interfaces.ObjectType
		relations []*interfaces.RelationType
		actions   []*interfaces.ActionType
	)
	if objectResp != nil {
		objects = objectResp.Entries
	}
	if relationResp != nil {
		relations = relationResp.Entries
	}
	if actionResp != nil {
		actions = actionResp.Entries
	}

	s.logger.WithContext(ctx).Debugf(
		"[ConceptRetrieval][Groups] BKN returned: groups=%v object_types=%d relation_types=%d action_types=%d",
		config.ConceptGroups, len(objects), len(relations), len(actions),
	)

	objects, err = s.completeReferencedObjectTypes(ctx, req.KnID, objects, relations, actions)
	if err != nil {
		s.logger.WithContext(ctx).Errorf("[ConceptRetrieval][Groups] complete referenced object types failed: %v", err)
		return nil, err
	}

	rankedRelations := s.rankRelationTypes(ctx, req.Query, objects, relations, config.TopK, req.EnableRerank)
	selectedObjects := s.selectObjectTypesForConceptRetrieval(objects, rankedRelations, config.TopK)

	brief := boolValue(config.SchemaBrief)
	objectTypesLocal := s.convertObjectTypesToLocal(selectedObjects, brief)
	relationTypesLocal := s.convertRelationTypesToLocal(rankedRelations, brief)
	actionTypesLocal := s.convertActionTypesToLocal(actions, req.KnID, objects)

	if boolValue(config.IncludeSampleData) && len(objectTypesLocal) > 0 {
		s.fetchSampleData(ctx, req.KnID, objectTypesLocal, brief)
	}

	return &interfaces.KnSearchConceptResult{
		ObjectTypes:   objectTypesLocal,
		RelationTypes: relationTypesLocal,
		ActionTypes:   actionTypesLocal,
	}, nil
}

func (s *localSearchImpl) completeReferencedObjectTypes(
	ctx context.Context,
	knID string,
	objects []*interfaces.ObjectType,
	relations []*interfaces.RelationType,
	actions []*interfaces.ActionType,
) ([]*interfaces.ObjectType, error) {
	known := make(map[string]struct{}, len(objects))
	for _, obj := range objects {
		if obj == nil || obj.ID == "" {
			continue
		}
		known[obj.ID] = struct{}{}
	}

	missingSeen := map[string]struct{}{}
	missingIDs := []string{}
	addMissing := func(id string) {
		id = strings.TrimSpace(id)
		if id == "" {
			return
		}
		if _, ok := known[id]; ok {
			return
		}
		if _, ok := missingSeen[id]; ok {
			return
		}
		missingSeen[id] = struct{}{}
		missingIDs = append(missingIDs, id)
	}

	for _, rel := range relations {
		if rel == nil {
			continue
		}
		addMissing(rel.SourceObjectTypeID)
		addMissing(rel.TargetObjectTypeID)
	}
	for _, action := range actions {
		if action == nil {
			continue
		}
		addMissing(action.ObjectTypeID)
	}

	if len(missingIDs) == 0 {
		return objects, nil
	}

	enriched, err := s.bknBackend.GetObjectTypeDetail(ctx, knID, missingIDs, true)
	if err != nil {
		return nil, err
	}

	out := make([]*interfaces.ObjectType, 0, len(objects)+len(enriched))
	out = append(out, objects...)
	for _, obj := range enriched {
		if obj == nil || obj.ID == "" {
			continue
		}
		if _, ok := known[obj.ID]; ok {
			continue
		}
		known[obj.ID] = struct{}{}
		out = append(out, obj)
	}

	return out, nil
}

func (s *localSearchImpl) selectObjectTypesForConceptRetrieval(
	objectTypes []*interfaces.ObjectType,
	relations []*interfaces.RelationType,
	topK int,
) []*interfaces.ObjectType {
	if len(objectTypes) == 0 {
		return objectTypes
	}

	maxObjectCountNoRelation := topK * objectTypeRelationMultiplier
	if len(relations) == 0 {
		return sortAndTruncateObjectTypesByScore(objectTypes, maxObjectCountNoRelation)
	}

	filtered := s.filterObjectTypesByRelations(objectTypes, relations)
	maxObjectCount := maxInt(len(relations)*objectTypeRelationMultiplier, topK)
	if len(filtered) >= maxObjectCount {
		return filtered
	}

	included := make(map[string]bool, len(filtered))
	for _, obj := range filtered {
		included[obj.ID] = true
	}

	type scoredObject struct {
		obj   *interfaces.ObjectType
		score float64
	}
	var candidatesWithScore []scoredObject
	var candidatesWithoutScore []*interfaces.ObjectType
	for _, obj := range objectTypes {
		if included[obj.ID] {
			continue
		}
		if obj.Score > 0 {
			candidatesWithScore = append(candidatesWithScore, scoredObject{obj: obj, score: obj.Score})
		} else {
			candidatesWithoutScore = append(candidatesWithoutScore, obj)
		}
	}

	sort.SliceStable(candidatesWithScore, func(i, j int) bool {
		return candidatesWithScore[i].score > candidatesWithScore[j].score
	})

	out := make([]*interfaces.ObjectType, 0, maxObjectCount)
	out = append(out, filtered...)

	remaining := maxObjectCount - len(out)
	for i := 0; i < len(candidatesWithScore) && remaining > 0; i++ {
		out = append(out, candidatesWithScore[i].obj)
		remaining--
	}
	for i := 0; i < len(candidatesWithoutScore) && remaining > 0; i++ {
		out = append(out, candidatesWithoutScore[i])
		remaining--
	}

	return out
}

func sortAndTruncateObjectTypesByScore(objectTypes []*interfaces.ObjectType, limit int) []*interfaces.ObjectType {
	if limit <= 0 || len(objectTypes) <= limit {
		return objectTypes
	}

	type scoredObject struct {
		obj   *interfaces.ObjectType
		score float64
	}

	var withScore []scoredObject
	var withoutScore []*interfaces.ObjectType
	for _, obj := range objectTypes {
		if obj.Score > 0 {
			withScore = append(withScore, scoredObject{obj: obj, score: obj.Score})
		} else {
			withoutScore = append(withoutScore, obj)
		}
	}

	sort.SliceStable(withScore, func(i, j int) bool {
		return withScore[i].score > withScore[j].score
	})

	out := make([]*interfaces.ObjectType, 0, limit)
	for i := 0; i < len(withScore) && len(out) < limit; i++ {
		out = append(out, withScore[i].obj)
	}
	for i := 0; i < len(withoutScore) && len(out) < limit; i++ {
		out = append(out, withoutScore[i])
	}
	return out
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// coarseRecall 粗召回：在大规模知识网络中先裁剪候选集
// 业务逻辑：构造 knn+match 查询条件，调用基础搜索接口
func (s *localSearchImpl) coarseRecall(
	ctx context.Context,
	knID string,
	query string,
	detail *interfaces.KnowledgeNetworkDetail,
	config *interfaces.KnSearchConceptRetrievalConfig,
) (*interfaces.KnowledgeNetworkDetail, error) {
	ctx, _ = o11y.StartInternalSpan(ctx)
	defer o11y.EndSpan(ctx, nil)

	// 构建粗召回后的对象类型 ID 集合
	coarseObjectIDs := make(map[string]bool)
	coarseRelationIDs := make(map[string]bool)
	coarseObjectScores := make(map[string]float64)
	coarseRelationScores := make(map[string]float64)

	// 粗召回对象类型
	objectReq := s.buildCoarseRecallQuery(knID, query, config.CoarseObjectLimit, config.ConceptGroups)
	coarseObjects, objErr := s.bknBackend.SearchObjectTypes(ctx, objectReq)
	if objErr != nil {
		s.logger.WithContext(ctx).Warnf("[CoarseRecall] SearchObjectTypes failed: %v", objErr)
	} else if coarseObjects != nil {
		for _, obj := range coarseObjects.Entries {
			coarseObjectIDs[obj.ID] = true
			if obj.Score > 0 {
				coarseObjectScores[obj.ID] = obj.Score
			}
		}
	}

	// 粗召回关系类型
	relationReq := s.buildCoarseRecallQuery(knID, query, config.CoarseRelationLimit, config.ConceptGroups)
	coarseRelations, relErr := s.bknBackend.SearchRelationTypes(ctx, relationReq)
	if relErr != nil {
		s.logger.WithContext(ctx).Warnf("[CoarseRecall] SearchRelationTypes failed: %v", relErr)
	} else if coarseRelations != nil {
		for _, rel := range coarseRelations.Entries {
			coarseRelationIDs[rel.ID] = true
			if rel.Score > 0 {
				coarseRelationScores[rel.ID] = rel.Score
			}
		}
	}

	// 过滤原始数据
	filteredDetail := &interfaces.KnowledgeNetworkDetail{
		ID:          detail.ID,
		ActionTypes: detail.ActionTypes, // ActionTypes 不做粗召回过滤
	}

	// 过滤对象类型
	if len(coarseObjectIDs) > 0 {
		relationEndpointIDs := make(map[string]bool)
		if len(coarseRelationIDs) > 0 {
			for _, rel := range detail.RelationTypes {
				if !coarseRelationIDs[rel.ID] {
					continue
				}
				if rel.SourceObjectTypeID != "" {
					relationEndpointIDs[rel.SourceObjectTypeID] = true
				}
				if rel.TargetObjectTypeID != "" {
					relationEndpointIDs[rel.TargetObjectTypeID] = true
				}
			}
		}

		candidateObjectIDs := make(map[string]bool, len(coarseObjectIDs)+len(relationEndpointIDs))
		for id := range coarseObjectIDs {
			candidateObjectIDs[id] = true
		}
		for id := range relationEndpointIDs {
			candidateObjectIDs[id] = true
		}

		var pruned []*interfaces.ObjectType
		for _, obj := range detail.ObjectTypes {
			if candidateObjectIDs[obj.ID] {
				if score, ok := coarseObjectScores[obj.ID]; ok {
					obj.Score = score
				}
				pruned = append(pruned, obj)
			}
		}
		if len(pruned) > 0 {
			filteredDetail.ObjectTypes = pruned
		} else {
			filteredDetail.ObjectTypes = detail.ObjectTypes
		}
	} else {
		filteredDetail.ObjectTypes = detail.ObjectTypes
	}

	// 过滤关系类型
	if len(coarseRelationIDs) > 0 {
		var pruned []*interfaces.RelationType
		for _, rel := range detail.RelationTypes {
			if coarseRelationIDs[rel.ID] {
				if score, ok := coarseRelationScores[rel.ID]; ok {
					rel.Score = score
				}
				pruned = append(pruned, rel)
			}
		}
		if len(pruned) > 0 {
			filteredDetail.RelationTypes = pruned
		} else {
			filteredDetail.RelationTypes = detail.RelationTypes
		}
	} else {
		filteredDetail.RelationTypes = detail.RelationTypes
	}

	s.logger.WithContext(ctx).Infof("[CoarseRecall] After coarse recall: objects=%d->%d, relations=%d->%d",
		len(detail.ObjectTypes), len(filteredDetail.ObjectTypes),
		len(detail.RelationTypes), len(filteredDetail.RelationTypes))

	return filteredDetail, nil
}

// buildCoarseRecallQuery 构建粗召回查询条件
// 业务逻辑：使用 knn + match 组合查询，按分数降序排序
func (s *localSearchImpl) buildCoarseRecallQuery(knID, query string, limit int, conceptGroups []string) *interfaces.QueryConceptsReq {
	return &interfaces.QueryConceptsReq{
		KnID:          knID,
		ConceptGroups: normalizeConceptGroups(conceptGroups),
		Cond: &interfaces.KnCondition{
			Operation: interfaces.KnOperationTypeOr,
			SubConditions: []*interfaces.KnCondition{
				{
					Field:      "*",
					Operation:  interfaces.KnOperationTypeKnn,
					Value:      query,
					ValueFrom:  interfaces.CondValueFromConst,
					LimitKey:   interfaces.CondLimitKeyK,
					LimitValue: limit,
				},
				{
					Field:     "*",
					Operation: interfaces.KnOperationTypeMatch,
					Value:     query,
					ValueFrom: interfaces.CondValueFromConst,
				},
			},
		},
		Sort: []*interfaces.KnSortParams{
			{Field: "_score", Direction: "desc"},
		},
		Limit:     limit,
		NeedTotal: false,
	}
}

// rankRelationTypes 对关系类型进行语义排序并取 Top-K
// 使用 Rerank 服务进行语义排序
func (s *localSearchImpl) rankRelationTypes(
	ctx context.Context,
	query string,
	objectTypes []*interfaces.ObjectType,
	relations []*interfaces.RelationType,
	topK int,
	enableRerank bool,
) []*interfaces.RelationType {
	if len(relations) == 0 {
		return relations
	}

	// 不启用 Rerank 时：保持原始顺序，仅截断 Top-K（用于与 Python 当前概念召回行为对齐）
	if !enableRerank {
		if topK <= 0 || topK >= len(relations) {
			return relations
		}
		return relations[:topK]
	}

	objectNameByID := make(map[string]string, len(objectTypes))
	for _, obj := range objectTypes {
		if obj == nil || obj.ID == "" {
			continue
		}
		name := strings.TrimSpace(obj.Name)
		if name == "" {
			name = obj.ID
		}
		objectNameByID[obj.ID] = name
	}

	documents := make([]string, len(relations))
	for i, rel := range relations {
		if rel == nil {
			continue
		}
		sourceName := objectNameByID[rel.SourceObjectTypeID]
		targetName := objectNameByID[rel.TargetObjectTypeID]
		relationName := strings.TrimSpace(rel.Name)
		if relationName == "" {
			relationName = rel.ID
		}
		documents[i] = buildRelationText(sourceName, relationName, targetName, rel.Comment)
	}

	// 调用 Rerank 服务
	rerankResp, err := s.rerankClient.Rerank(ctx, query, documents)
	if err != nil {
		s.logger.WithContext(ctx).Warnf("[RankRelationTypes] Rerank failed, fallback to simple match: %v", err)
		return s.rankRelationTypesBySimpleMatch(query, relations, topK)
	}

	// 按 Rerank 分数排序
	type scoredRelation struct {
		relation *interfaces.RelationType
		score    float64
	}

	scored := make([]scoredRelation, len(relations))
	for i, rel := range relations {
		scored[i] = scoredRelation{
			relation: rel,
			score:    0,
		}
	}
	for _, result := range rerankResp.Results {
		if result.Index >= 0 && result.Index < len(relations) {
			scored[result.Index].score = result.RelevanceScore
		}
	}

	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 取 Top-K
	if topK > len(scored) {
		topK = len(scored)
	}

	result := make([]*interfaces.RelationType, topK)
	for i := 0; i < topK; i++ {
		result[i] = scored[i].relation
	}

	s.logger.WithContext(ctx).Debugf("[RankRelationTypes] Rerank completed, top_k=%d", len(result))

	return result
}

func buildRelationText(sourceName, relationName, targetName, relationComment string) string {
	var parts []string
	if strings.TrimSpace(sourceName) != "" {
		parts = append(parts, strings.TrimSpace(sourceName))
	}
	if strings.TrimSpace(relationName) != "" {
		if strings.TrimSpace(relationComment) != "" {
			parts = append(parts, strings.TrimSpace(relationName)+"，"+strings.TrimSpace(relationComment))
		} else {
			parts = append(parts, strings.TrimSpace(relationName))
		}
	}
	if strings.TrimSpace(targetName) != "" {
		parts = append(parts, strings.TrimSpace(targetName))
	}
	return strings.Join(parts, " ")
}

// rankRelationTypesBySimpleMatch 使用简单匹配进行排序（Rerank 失败时的回退）
func (s *localSearchImpl) rankRelationTypesBySimpleMatch(
	query string,
	relations []*interfaces.RelationType,
	topK int,
) []*interfaces.RelationType {
	// 简单的相关性评分（基于名称匹配）
	type scoredRelation struct {
		relation *interfaces.RelationType
		score    float64
	}

	scored := make([]scoredRelation, len(relations))
	for i, rel := range relations {
		score := s.calculateRelevanceScore(query, rel.Name, rel.Comment)
		scored[i] = scoredRelation{relation: rel, score: score}
	}

	// 按分数降序排序
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	// 取 Top-K
	if topK > len(scored) {
		topK = len(scored)
	}

	result := make([]*interfaces.RelationType, topK)
	for i := 0; i < topK; i++ {
		result[i] = scored[i].relation
	}

	return result
}

// filterObjectTypesByRelations 根据关系类型过滤对象类型
func (s *localSearchImpl) filterObjectTypesByRelations(
	objectTypes []*interfaces.ObjectType,
	relations []*interfaces.RelationType,
) []*interfaces.ObjectType {
	if len(relations) == 0 {
		return objectTypes
	}

	// 收集关系涉及的对象类型 ID
	relatedObjectIDs := make(map[string]bool)
	for _, rel := range relations {
		relatedObjectIDs[rel.SourceObjectTypeID] = true
		relatedObjectIDs[rel.TargetObjectTypeID] = true
	}

	// 过滤对象类型
	var filtered []*interfaces.ObjectType
	for _, obj := range objectTypes {
		if relatedObjectIDs[obj.ID] {
			filtered = append(filtered, obj)
		}
	}

	return filtered
}

// calculateRelevanceScore 计算 Query 与概念的相关性分数
func (s *localSearchImpl) calculateRelevanceScore(query, name, comment string) float64 {
	if strings.TrimSpace(query) == "" {
		return 0
	}

	score := 0.0

	// 名称完全匹配
	if name == query {
		score += 1.0
	}

	// 名称包含 Query
	if name != "" {
		if containsFold(name, query) {
			score += 0.5
		}
		if containsFold(query, name) {
			score += 0.3
		}
	}

	// 描述包含 Query
	if comment != "" && containsFold(comment, query) {
		score += 0.2
	}

	return score
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// pruneProperties 属性裁剪：只保留与 Query 最相关的属性
func (s *localSearchImpl) pruneProperties(
	_ context.Context,
	query string,
	objectTypes []*interfaces.KnSearchObjectType,
	config *interfaces.KnSearchConceptRetrievalConfig,
) []*interfaces.KnSearchObjectType {
	// 收集所有属性及其分数
	type propertyWithScore struct {
		objIndex  int
		propIndex int
		isLogic   bool
		score     float64
	}

	var allProperties []propertyWithScore

	for objIdx, obj := range objectTypes {
		for propIdx, prop := range obj.DataProperties {
			score := s.calculateRelevanceScore(query, prop.Name, prop.Comment)
			allProperties = append(allProperties, propertyWithScore{
				objIndex:  objIdx,
				propIndex: propIdx,
				isLogic:   false,
				score:     score,
			})
		}
		for propIdx, prop := range obj.LogicProperties {
			score := s.calculateRelevanceScore(query, prop.Name, prop.Comment)
			allProperties = append(allProperties, propertyWithScore{
				objIndex:  objIdx,
				propIndex: propIdx,
				isLogic:   true,
				score:     score,
			})
		}
	}

	// 按分数降序排序
	sort.Slice(allProperties, func(i, j int) bool {
		return allProperties[i].score > allProperties[j].score
	})

	// 标记要保留的属性
	perObjectDataCount := make(map[int]int)
	perObjectLogicCount := make(map[int]int)
	keepDataProps := make(map[int]map[int]bool)
	keepLogicProps := make(map[int]map[int]bool)

	for i := range objectTypes {
		keepDataProps[i] = make(map[int]bool)
		keepLogicProps[i] = make(map[int]bool)
	}

	globalCount := 0
	for _, prop := range allProperties {
		if globalCount >= config.GlobalPropertyTopK {
			break
		}

		if prop.isLogic {
			if perObjectLogicCount[prop.objIndex] < config.PerObjectPropertyTopK {
				keepLogicProps[prop.objIndex][prop.propIndex] = true
				perObjectLogicCount[prop.objIndex]++
				globalCount++
			}
		} else {
			if perObjectDataCount[prop.objIndex] < config.PerObjectPropertyTopK {
				keepDataProps[prop.objIndex][prop.propIndex] = true
				perObjectDataCount[prop.objIndex]++
				globalCount++
			}
		}
	}

	// 应用裁剪
	for objIdx, obj := range objectTypes {
		var filteredDataProps []*interfaces.KnSearchDataProperty
		for propIdx, prop := range obj.DataProperties {
			if keepDataProps[objIdx][propIdx] {
				filteredDataProps = append(filteredDataProps, prop)
			}
		}
		obj.DataProperties = filteredDataProps

		var filteredLogicProps []*interfaces.KnSearchLogicProperty
		for propIdx, prop := range obj.LogicProperties {
			if keepLogicProps[objIdx][propIdx] {
				filteredLogicProps = append(filteredLogicProps, prop)
			}
		}
		obj.LogicProperties = filteredLogicProps
	}

	return objectTypes
}

// fetchSampleData 获取样例数据
func (s *localSearchImpl) fetchSampleData(ctx context.Context, knID string, objectTypes []*interfaces.KnSearchObjectType, schemaBrief bool) {
	for _, obj := range objectTypes {
		// 调用实例检索获取一条样例数据
		req := &interfaces.QueryObjectInstancesReq{
			KnID:               knID,
			OtID:               obj.ConceptID,
			IncludeTypeInfo:    false,
			IncludeLogicParams: false,
			Limit:              1,
		}

		resp, err := s.ontologyQuery.QueryObjectInstances(ctx, req)
		if err != nil {
			s.logger.WithContext(ctx).Warnf("[FetchSampleData] Failed to fetch sample for %s: %v", obj.ConceptID, err)
			continue
		}

		if len(resp.Data) > 0 {
			if dataMap, ok := resp.Data[0].(map[string]any); ok {
				if !schemaBrief {
					obj.SampleData = dataMap
					continue
				}
				briefSample := make(map[string]any, len(dataMap))
				for k, v := range dataMap {
					if k == "_score" {
						continue
					}
					briefSample[k] = v
				}
				obj.SampleData = briefSample
			}
		}
	}
}

// ==================== 类型转换函数 ====================

// convertObjectTypesToLocal 将对象类型映射为本地响应结构；brief 控制包含的字段范围。
func (s *localSearchImpl) convertObjectTypesToLocal(objects []*interfaces.ObjectType, brief bool) []*interfaces.KnSearchObjectType {
	result := make([]*interfaces.KnSearchObjectType, len(objects))
	for i, obj := range objects {
		var conceptType string
		var primaryKeys []string
		var tags []string
		var dataSource *interfaces.ResourceInfo
		if !brief {
			conceptType = "object_type"
			primaryKeys = obj.PrimaryKeys
			tags = obj.Tags
			dataSource = obj.DataSource
		}
		localObj := &interfaces.KnSearchObjectType{
			ConceptType: conceptType,
			ConceptID:   obj.ID,
			ConceptName: obj.Name,
			Comment:     obj.Comment,
			Tags:        tags,
			DataSource:  dataSource,
			PrimaryKeys: primaryKeys,
		}

		if len(obj.DataProperties) > 0 {
			localObj.DataProperties = make([]*interfaces.KnSearchDataProperty, len(obj.DataProperties))
			for j, prop := range obj.DataProperties {
				p := &interfaces.KnSearchDataProperty{
					Name:                prop.Name,
					Type:                prop.Type,
					ConditionOperations: prop.ConditionOperations,
				}
				if !brief {
					p.Comment = prop.Comment
				}
				localObj.DataProperties[j] = p
			}
		}

		// Convert logic properties (full fields: DataSource, Parameters)
		if len(obj.LogicProperties) > 0 {
			localObj.LogicProperties = make([]*interfaces.KnSearchLogicProperty, len(obj.LogicProperties))
			for j, prop := range obj.LogicProperties {
				lp := &interfaces.KnSearchLogicProperty{
					Name:       prop.Name,
					Type:       string(prop.Type),
					DataSource: prop.DataSource,
					Parameters: prop.Parameters,
				}
				if !brief {
					lp.Comment = prop.Comment
				}
				localObj.LogicProperties[j] = lp
			}
		}

		result[i] = localObj
	}
	return result
}

// convertRelationTypesToLocal 转换关系类型为本地响应格式，与 Python schema_brief 对齐：
// - brief=true：仅返回 concept_id, concept_name, source_object_type_id, target_object_type_id（不含 concept_type, comment）
// - brief=false：返回完整字段（含 concept_type, comment）
func (s *localSearchImpl) convertRelationTypesToLocal(relations []*interfaces.RelationType, brief bool) []*interfaces.KnSearchRelationType {
	result := make([]*interfaces.KnSearchRelationType, len(relations))
	for i, rel := range relations {
		var conceptType, comment string
		if !brief {
			conceptType = "relation_type"
			comment = rel.Comment
		}
		result[i] = &interfaces.KnSearchRelationType{
			ConceptType:        conceptType,
			ConceptID:          rel.ID,
			ConceptName:        rel.Name,
			Comment:            comment,
			SourceObjectTypeID: rel.SourceObjectTypeID,
			TargetObjectTypeID: rel.TargetObjectTypeID,
		}
	}
	return result
}

func (s *localSearchImpl) convertActionTypesToLocal(
	actions []*interfaces.ActionType,
	knID string,
	objectTypes []*interfaces.ObjectType,
) []*interfaces.KnSearchActionType {
	objNameMap := make(map[string]string, len(objectTypes))
	for _, obj := range objectTypes {
		if obj == nil || obj.ID == "" {
			continue
		}
		objNameMap[obj.ID] = obj.Name
	}

	result := make([]*interfaces.KnSearchActionType, len(actions))
	for i, act := range actions {
		result[i] = &interfaces.KnSearchActionType{
			ID:             act.ID,
			Name:           act.Name,
			ObjectTypeID:   act.ObjectTypeID,
			ObjectTypeName: objNameMap[act.ObjectTypeID],
			Comment:        act.Comment,
			Tags:           act.Tags,
			KnID:           knID,
		}
	}
	return result
}
