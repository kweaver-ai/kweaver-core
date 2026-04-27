// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package dsl

import (
	"context"
	"fmt"
	"net/http"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// logicViewDSLGenerator 用于生成DSL
type logicViewDSLGenerator struct {
	nodes         map[string]*interfaces.LogicDefinitionNode
	outputNode    *interfaces.LogicDefinitionNode
	nodeFieldsMap map[string]map[string]*interfaces.ViewProperty
}

// NewlogicViewSQLGenerator 创建SQL生成器
func NewlogicViewDSLGenerator(nodes []*interfaces.LogicDefinitionNode) *logicViewDSLGenerator {
	nodeMap := make(map[string]*interfaces.LogicDefinitionNode)
	var outputNode *interfaces.LogicDefinitionNode
	for i := range nodes {
		nodeMap[nodes[i].ID] = nodes[i]
		if nodes[i].Type == interfaces.LogicDefinitionNodeType_Output {
			outputNode = nodes[i]
		}
	}
	return &logicViewDSLGenerator{
		nodes:         nodeMap,
		outputNode:    outputNode,
		nodeFieldsMap: make(map[string]map[string]*interfaces.ViewProperty),
	}
}

// DSL生成器
func (g *logicViewDSLGenerator) BuildDSL(ctx context.Context, query interfaces.ResourceDataQueryParams, view *interfaces.LogicView,
	viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	sortParams := completeDSLSortParams(query.Sort, query.QueryType)

	var dsl interfaces.DSLCfg
	// 设置分页参数和track_total_hits
	dsl.From = query.Offset
	dsl.Size = query.Limit
	if query.NeedTotal {
		dsl.TrackTotalHits = true
	}

	if len(sortParams) > 0 {
		sort := []map[string]any{}
		for _, sp := range sortParams {
			if sp.Field == "" || sp.Direction == "" {
				return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
					rest.PublicError_BadRequest).
					WithErrorDetails("The sort field and direction cannot be empty")
			}

			sortFieldName := sp.Field
			sortField, ok := view.FieldsMap[sp.Field]

			if ok {
				if sortField.Type == interfaces.DataType_Binary {
					return dsl, rest.NewHTTPError(ctx, http.StatusBadRequest,
						rest.PublicError_BadRequest).
						WithErrorDetails(fmt.Sprintf("The sort field '%s' is binary type, do not support sorting", sp.Field))
				}

				// text类型的字段需要看其下有没有配置keyword索引，配了就用 xxx.keyword 进行排序。否则不纳入排序
				// string类型的字段直接支持排序，若其有全文索引，则在字段的 keyword 下有 text
				if IsTextType(sortField) {
					if HasFeature(sortField, interfaces.PropertyFeatureType_Keyword) {
						sortFieldName = sortFieldName + ".keyword"
					} else {
						continue
					}
				}
			}

			// 需要将视图字段__score转为opensearch内置字段_score, 暂时不修改，兼容处理
			if sortFieldName == "__score" {
				sortFieldName = "_score"
			}

			sort = append(sort, map[string]any{
				sortFieldName: sp.Direction,
			})
		}

		dsl.Sort = sort
	}

	// 获取searchAfter参数
	searchAfterDSL, err := getSearchAfterDSL(nil)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to get search after dsl, %s", err.Error()))
	}

	// 合并searchAfterDSL到主DSL结构体
	dsl.SearchAfter = searchAfterDSL.SearchAfter
	dsl.Pit = searchAfterDSL.Pit

	// 构建查询条件
	queryDSL, err := g.buildDSLQuery(ctx, view, viewIndicesMap)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to build query dsl, %s", err.Error()))
	}

	// 合并查询条件到主DSL结构体
	dsl.Query = queryDSL.Query

	// 添加全局过滤条件，全局过滤条件的字段应该在视图字段列表里
	dsl, err = addGlobalFiltersToDSL(ctx, dsl, query.FilterCondCfg, view.FieldsMap)
	if err != nil {
		return dsl, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("failed to add global filters to dsl, %s", err.Error()))
	}

	logger.Infof("view_indices_map is %v", viewIndicesMap)

	return dsl, nil
}

// 生成Resource节点的查询条件, 返回查询条件DSL
func (g *logicViewDSLGenerator) buildResourceQuery(ctx context.Context, node *interfaces.LogicDefinitionNode,
	refResources map[string]*interfaces.Resource, viewIndicesMap map[string][]string) (map[string]any, error) {
	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to decode view node config, %s", err.Error())
	}

	if cfg.ResourceID == "" {
		return nil, fmt.Errorf("resource id is empty")
	}

	indices, exists := viewIndicesMap[cfg.ResourceID]
	if !exists {
		return nil, fmt.Errorf("no indices found for resource ID: %s", cfg.ResourceID)
	}

	indexConditions := map[string]any{
		"terms": map[string]any{
			"_index": indices,
		},
	}

	fieldMap := map[string]*interfaces.Property{}
	for _, prop := range refResources[cfg.ResourceID].SchemaDefinition {
		fieldMap[prop.Name] = prop
	}

	filterCond, err := g.buildDSLCondition(ctx, cfg.Filters, fieldMap)
	if err != nil {
		return nil, err
	}

	if filterCond == nil {
		return indexConditions, nil
	}

	return map[string]any{
		"bool": map[string]any{
			"must": []any{indexConditions, filterCond},
		},
	}, nil

}

// 添加全局过滤条件到DSL
func addGlobalFiltersToDSL(ctx context.Context, dsl interfaces.DSLCfg, filters *interfaces.FilterCondCfg,
	fieldsMap map[string]*interfaces.ViewProperty) (interfaces.DSLCfg, error) {
	// condStr, needScore, err := buildDSLCondition(ctx, filters, fieldsMap)
	// if err != nil {
	// 	return dsl, err
	// }

	// if condStr != "" {
	// 	var filterCondition map[string]any
	// 	if err := sonic.Unmarshal([]byte(condStr), &filterCondition); err != nil {
	// 		return dsl, fmt.Errorf("failed to unmarshal filter condition, %s", err.Error())
	// 	}

	// 	// 如果需要打分，使用must查询
	// 	if needScore {
	// 		dsl.TrackScores = true
	// 		dsl.Query.Bool.Must = append(dsl.Query.Bool.Must, filterCondition)
	// 	} else {
	// 		dsl.Query.Bool.Filter = append(dsl.Query.Bool.Filter, filterCondition)
	// 	}
	// }

	// return dsl, nil
	return dsl, nil
}

func (g *logicViewDSLGenerator) buildDSLQuery(ctx context.Context, view *interfaces.LogicView,
	viewIndicesMap map[string][]string) (interfaces.DSLCfg, error) {
	// 自定义视图logic definition不能为null
	if view.LogicDefinition == nil {
		return interfaces.DSLCfg{}, fmt.Errorf("logic definition is nil")
	}

	// 提取所有视图节点
	var viewNodes []*interfaces.LogicDefinitionNode
	for _, node := range view.LogicDefinition {
		switch node.Type {
		case interfaces.LogicDefinitionNodeType_Resource:
			viewNodes = append(viewNodes, node)
		case interfaces.LogicDefinitionNodeType_Union:
			var unionCfg *interfaces.UnionNodeCfg
			err := mapstructure.Decode(node.Config, &unionCfg)
			if err != nil {
				return interfaces.DSLCfg{}, fmt.Errorf("failed to decode union node config, %s", err.Error())
			}

			// interfaces.DSLCfg 类视图只允许配置 all
			if unionCfg.UnionType != interfaces.UnionType_All {
				return interfaces.DSLCfg{}, fmt.Errorf("unsupported union type: %s", unionCfg.UnionType)
			}
		case interfaces.LogicDefinitionNodeType_Output:
		default:
			return interfaces.DSLCfg{}, fmt.Errorf("unsupported node type: %s", node.Type)
		}
	}

	var dsl interfaces.DSLCfg
	// 根据视图节点数量决定查询结构
	if len(viewNodes) == 1 {
		// 单视图节点，直接使用filter，不用should
		query, err := g.buildResourceQuery(ctx, viewNodes[0], view.RefResources, viewIndicesMap)
		if err != nil {
			return interfaces.DSLCfg{}, err
		}
		dsl.Query.Bool.Filter = []any{query}

	} else {
		// 多视图节点，使用should
		shouldQueries := make([]any, 0, len(viewNodes))
		for _, node := range viewNodes {
			query, err := g.buildResourceQuery(ctx, node, view.RefResources, viewIndicesMap)
			if err != nil {
				return interfaces.DSLCfg{}, err
			}
			shouldQueries = append(shouldQueries, query)
		}

		dsl.Query.Bool.Should = shouldQueries
		// 设置min_should_match为1，确保至少匹配一个should条件
		dsl.Query.Bool.MinShouldMatch = 1
	}

	return dsl, nil
}

// 构造过滤条件
func (g *logicViewDSLGenerator) buildDSLCondition(ctx context.Context, filters *interfaces.FilterCondCfg,
	fieldMap map[string]*interfaces.Property) (map[string]any, error) {
	// 将过滤条件拼接到 dsl 的 query 中
	filterCond, err := filter_condition.NewFilterCondition(ctx, filters, fieldMap)
	if err != nil {
		return nil, fmt.Errorf("failed to new condition, %s", err.Error())
	}

	if filterCond == nil {
		return nil, nil
	}

	dslCond, err := g.ConvertFilterCondition(ctx, filterCond, fieldMap)
	if err != nil {
		return nil, fmt.Errorf("failed to convert condition to dsl, %s", err.Error())
	}

	return dslCond, nil
}

// 补充 sort 字段
func completeDSLSortParams(sort []*interfaces.SortField, queryType string) []*interfaces.SortField {
	defaultSort := []*interfaces.SortField{}
	if queryType == "stream" {
		defaultSort = []*interfaces.SortField{
			{Field: "_id", Direction: interfaces.DESC_DIRECTION},
		}
	}

	sort = append(sort, defaultSort...)
	newSort := []*interfaces.SortField{}
	// 去重
	sortFieldSet := map[string]struct{}{}
	for _, sortParam := range sort {
		if _, ok := sortFieldSet[sortParam.Field]; !ok {
			newSort = append(newSort, sortParam)
			sortFieldSet[sortParam.Field] = struct{}{}
		}
	}

	return newSort
}

// 检查字段是否为 text 类型
func IsTextType(fieldInfo *interfaces.ViewProperty) bool {
	return fieldInfo != nil && fieldInfo.Type == interfaces.DataType_Text
}

// 检查字段特征是否包含指定特征
func HasFeature(fieldInfo *interfaces.ViewProperty, feature string) bool {
	for _, f := range fieldInfo.Features {
		if f.FeatureType == feature {
			return true
		}
	}
	return false
}

// 三种情况需要拼接 dsl
// 1. 没有pit，有search_after
// 2. 有pit，有search_after
// 3. 有pit，没有search_after
func getSearchAfterDSL(searchAfterParams *interfaces.SearchAfterParams) (interfaces.DSLCfg, error) {
	var dsl interfaces.DSLCfg

	if searchAfterParams == nil {
		return dsl, nil
	}

	if len(searchAfterParams.SearchAfter) > 0 {
		dsl.SearchAfter = searchAfterParams.SearchAfter
	}

	// 设置pit
	if searchAfterParams.PitID != "" {
		dsl.Pit = &struct {
			ID        string `json:"id,omitempty"`
			KeepAlive string `json:"keep_alive,omitempty"`
		}{}
		dsl.Pit.ID = searchAfterParams.PitID
		if searchAfterParams.PitKeepAlive != "" {
			dsl.Pit.KeepAlive = searchAfterParams.PitKeepAlive
		}
	}

	return dsl, nil

}
