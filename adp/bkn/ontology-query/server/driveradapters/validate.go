// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// 校验 x-http-method-override 重载方法，只在header里传递 method
func ValidateHeaderMethodOverride(ctx context.Context, headerMethod string) error {
	// 校验 method
	if headerMethod == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_NullParameter_OverrideMethod)
	}
	if headerMethod != "GET" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_OverrideMethod).
			WithErrorDetails(fmt.Sprintf("X-HTTP-Method-Override is expected to be GET, but it is actually %s", headerMethod))
	}

	return nil
}

// 校验对象类的查询参数
func validateObjectsQueryParameters(ctx context.Context, includeTypeInfo string, ignoringStoreCache string,
	includeLogicParams string, excludeSystemProperties []string) (interfaces.CommonQueryParameters, error) {

	includeType, err := strconv.ParseBool(includeTypeInfo)
	if err != nil {
		return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter_IncludeTypeInfo).
			WithErrorDetails(fmt.Sprintf("The include_type_info:%s is invalid", includeTypeInfo))
	}

	includeLogicP, err := strconv.ParseBool(includeLogicParams)
	if err != nil {
		return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter_IncludeTypeInfo).
			WithErrorDetails(fmt.Sprintf("The include_logic_params:%s is invalid", includeLogicParams))
	}

	ignoringStore, err := strconv.ParseBool(ignoringStoreCache)
	if err != nil {
		return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter_IgnoringStoreCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_store_cache:%s is invalid", ignoringStoreCache))
	}

	// 校验排除的系统字段
	validFields := map[string]bool{
		interfaces.SYSTEM_PROPERTY_INSTANCE_ID:       true,
		interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: true,
		interfaces.SYSTEM_PROPERTY_DISPLAY:           true,
	}
	for _, field := range excludeSystemProperties {
		if !validFields[field] {
			return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("无效的系统字段: %s，支持的字段有: _instance_id, _instance_identity, _display", field))
		}
	}

	return interfaces.CommonQueryParameters{
		IncludeTypeInfo:         includeType,
		IncludeLogicParams:      includeLogicP,
		IgnoringStore:           ignoringStore,
		ExcludeSystemProperties: excludeSystemProperties,
	}, nil
}

// 校验子图查询的查询参数
func validateSugraphQueryParameters(ctx context.Context,
	includeLogicParams string, ignoringStoreCache string, excludeSystemProperties []string) (interfaces.CommonQueryParameters, error) {

	includeLogicP, err := strconv.ParseBool(includeLogicParams)
	if err != nil {
		return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter_IncludeTypeInfo).
			WithErrorDetails(fmt.Sprintf("The include_logic_params:%s is invalid", includeLogicParams))
	}

	ignoringStore, err := strconv.ParseBool(ignoringStoreCache)
	if err != nil {
		return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter_IgnoringStoreCache).
			WithErrorDetails(fmt.Sprintf("The ignoring_store_cache:%s is invalid", ignoringStoreCache))
	}

	// 校验排除的系统字段
	validFields := map[string]bool{
		interfaces.SYSTEM_PROPERTY_INSTANCE_ID:       true,
		interfaces.SYSTEM_PROPERTY_INSTANCE_IDENTITY: true,
		interfaces.SYSTEM_PROPERTY_DISPLAY:           true,
	}
	for _, field := range excludeSystemProperties {
		if !validFields[field] {
			return interfaces.CommonQueryParameters{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("无效的系统字段: %s，支持的字段有: _instance_id, _instance_identity, _display", field))
		}
	}

	return interfaces.CommonQueryParameters{
		IncludeLogicParams:      includeLogicP,
		IgnoringStore:           ignoringStore,
		ExcludeSystemProperties: excludeSystemProperties,
	}, nil
}

// 基于起点、方向和路径长度获取对象子图的参数校验
func validateSubgraphSearchRequest(ctx context.Context, query *interfaces.SubGraphQueryBaseOnSource) error {

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *cond.CondCfg
	err := mapstructure.Decode(query.Condition, &actualCond)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode condition failed: %s", err.Error()))
	}
	query.ActualCondition = actualCond

	// 起点对象类非空
	if query.SourceObjecTypeId == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_NullParameter_SourceObjectTypeId)
	}

	// 方向非空
	if query.Direction == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_NullParameter_Direction)
	}

	// 方向有效性
	if !interfaces.DIRECTION_MAP[query.Direction] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter_Direction).
			WithErrorDetails(fmt.Sprintf("当前支持的方向有: forward, backward, bidirectional. 请求的方向为: %s", query.Direction))
	}

	// 路径长度不超过3
	if query.PathLength > 3 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter_PathLength).
			WithErrorDetails(fmt.Sprintf("路径长度不超过3, 请求的路径长度为%d", query.PathLength))
	}

	// sort 非空时，排序字段非空，排序方向非空，排序字段可以是对象类的数据属性, _score
	if len(query.Sort) > 0 {
		for _, sp := range query.Sort {
			if sp.Field == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails("排序字段不能为空")
			}
			if sp.Direction == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails("排序方向不能为空")
			}
			if sp.Direction != interfaces.DESC_DIRECTION && sp.Direction != interfaces.ASC_DIRECTION {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("排序方向只能是desc或asc, 当前排序方向为%s", sp.Direction))
			}

			// 排序字段可以是对象类的数据属性, _score，放在service层 get 到对象类信息后校验。
		}
	}

	// limit 可选值 1-10000, 默认值为 1000
	if query.Limit == 0 {
		query.Limit = interfaces.DEFAULT_LIMIT
	}
	if query.Limit < 1 || query.Limit > interfaces.MAX_LIMIT {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("limit可选值 1-10000, 当前limit为 %d", query.Limit))
	}

	return nil

}

// 基于路径获取对象子图的参数校验
func validateSubgraphQueryByPathRequest(ctx context.Context, query *interfaces.SubGraphQueryBaseOnTypePath) error {

	for i := range query.Paths.TypePaths {
		if len(query.Paths.TypePaths[i].Edges) > 10 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails("实例关系查询最多支持10度")
		}
		if query.Paths.TypePaths[i].Limit == 0 {
			query.Paths.TypePaths[i].Limit = interfaces.DEFAULT_PATHS // 不给路径长度时，给最大值2000
		}
	}

	for pathIndex, path := range query.Paths.TypePaths {
		// 1. 各路径的节点不能为空
		if len(path.ObjectTypes) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_NullParameter_TypePathObjectTypes)
		}
		// 2. 路径不能为空
		if len(path.Edges) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_NullParameter_TypePathRelationTypes)
		}

		// 3. 路径的起始点在边中需存在且位置正确
		for i, edge := range path.Edges {
			// 关系类id非空, 关系类存在在service层校验
			if edge.RelationTypeId == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("关系类id不能为空, 当前第%d条边的关系类id为%s", i+1, edge.RelationTypeId))
			}
			// 起点对象类id非空，对象类存在在service层校验
			if edge.SourceObjectTypeId == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("起点对象类id不能为空, 当前第%d条边的起点对象类id为%s", i+1, edge.SourceObjectTypeId))
			}
			// 终点对象类id非空，对象类存在在service层校验
			if edge.TargetObjectTypeId == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("终点对象类id不能为空, 当前第%d条边的终点对象类id为%s", i+1, edge.TargetObjectTypeId))
			}

			// 第i条边的起点等于第i个位置的对象类
			if edge.SourceObjectTypeId != path.ObjectTypes[i].OTID {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter_TypePath).
					WithErrorDetails(fmt.Sprintf("路径的边[%d]的起点对象类指向[%s],在对象类数组中对应的位置[%d]找到的对象类为[%s]",
						i, edge.SourceObjectTypeId, i, path.ObjectTypes[i].OTID))
			}
			// 第i条边的终点等于第i+1个位置的对象类
			if edge.TargetObjectTypeId != path.ObjectTypes[i+1].OTID {
				return rest.NewHTTPError(ctx, http.StatusBadRequest,
					oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter_TypePath).
					WithErrorDetails(fmt.Sprintf("路径的边[%d]的终点对象类指向[%s],在对象类数组中对应的位置[%d]找到的对象类为[%s]",
						i, edge.TargetObjectTypeId, i+1, path.ObjectTypes[i+1].OTID))
			}
			// 路径上当前边的终点是上一条边的起点
			if i > 0 {
				if edge.SourceObjectTypeId != path.Edges[i-1].TargetObjectTypeId {
					return rest.NewHTTPError(ctx, http.StatusBadRequest,
						oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter_TypePath).
						WithErrorDetails(fmt.Sprintf("当前请求的边无法组成一条路径,路径的边的起点是上一条边的终点,当前请求的路径的边[%d]的起点对象类指向[%s]，而前序边的终点对象类是[%s]",
							i, edge.SourceObjectTypeId, path.Edges[i-1].TargetObjectTypeId))
				}
			}
		}

		// 4. 各路径下的节点的过滤条件过滤条件用map接，然后再decode到condCfg中
		for i := range path.ObjectTypes {
			var actualCond *cond.CondCfg
			err := mapstructure.Decode(path.ObjectTypes[i].Condition, &actualCond)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails(fmt.Sprintf("mapstructure decode condition failed: %s", err.Error()))
			}
			query.Paths.TypePaths[pathIndex].ObjectTypes[i].ActualCondition = actualCond

			// 排序字段的校验在获取对象类的对象数据的时候校验，在当前层不用校验
			if len(path.ObjectTypes[i].Sort) > 0 {
				for _, sp := range path.ObjectTypes[i].Sort {
					if sp.Field == "" {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
							WithErrorDetails("排序字段不能为空")
					}
					if sp.Direction == "" {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
							WithErrorDetails("排序方向不能为空")
					}
					if sp.Direction != interfaces.DESC_DIRECTION && sp.Direction != interfaces.ASC_DIRECTION {
						return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
							WithErrorDetails(fmt.Sprintf("排序方向只能是desc或asc, 当前第%d个对象类上指定的排序方向为%s", i+1, sp.Direction))
					}
				}
			}

			// limit 可选值 1-10000, 默认值为 1000
			if path.ObjectTypes[i].Limit == 0 {
				path.ObjectTypes[i].Limit = interfaces.DEFAULT_LIMIT
			}
			if path.ObjectTypes[i].Limit < 1 || path.ObjectTypes[i].Limit > interfaces.MAX_LIMIT {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("limit可选值 1-10000, 当前limit为 %d", path.ObjectTypes[i].Limit))
			}
		}
	}

	return nil
}

// 基于对象类的对象数据查询的参数校验
func validateObjectSearchRequest(ctx context.Context, query *interfaces.ObjectQueryBaseOnObjectType) error {

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *cond.CondCfg
	err := mapstructure.Decode(query.Condition, &actualCond)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode condition failed: %s", err.Error()))
	}
	query.ActualCondition = actualCond

	// sort 非空时，排序字段非空，排序方向非空，排序字段可以是对象类的数据属性, _score
	if len(query.Sort) > 0 {
		for _, sp := range query.Sort {
			if sp.Field == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails("排序字段不能为空")
			}
			if sp.Direction == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails("排序方向不能为空")
			}
			if sp.Direction != interfaces.DESC_DIRECTION && sp.Direction != interfaces.ASC_DIRECTION {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("排序方向只能是desc或asc, 当前排序方向为%s", sp.Direction))
			}

			// 排序字段可以是对象类的数据属性, _score，放在service层 get 到对象类信息后校验。
		}
	}

	// limit 可选值 1-10000, 默认值为 10
	if query.Limit < 1 || query.Limit > interfaces.MAX_LIMIT {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("limit可选值 1-10000, 当前limit为 %d", query.Limit))
	}
	if query.Limit == 0 {
		query.Limit = interfaces.DEFAULT_OBJECT_LIMIT
	}

	return nil
}

// 基于行动类的行动数据查询的参数校验
func validateActionQuery(ctx context.Context, query *interfaces.ActionQuery) error {

	// 唯一标识非空，可接受空，当为空时，按action type 的条件去过滤数据
	// if len(query.InstanceIdentities) == 0 {
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ActionType_InvalidParameter).
	// 		WithErrorDetails("行动查询的唯一标识不能为空")
	// }
	return nil
}

// 属性值查询的参数校验
func validateObjectPropertyValueQuery(ctx context.Context, query *interfaces.ObjectPropertyValueQuery) error {

	// 唯一标识非空
	if len(query.InstanceIdentities) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails("属性查询的唯一标识不能为空")
	}

	// 属性列表非空
	if len(query.Properties) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails("属性查询的属性列表不能为空")
	}

	// 处理 start end instant 和 step
	// 如果没给参数，那么给默认参数， 放在service层获取到对象类信息之后再校验

	return nil
}

// 基于一组对象实例组织关系子图的参数校验
func validateSubgraphQueryByObjectsRequest(ctx context.Context, query *interfaces.SubGraphQueryBaseOnObjects) error {

	// entries 数组非空
	if len(query.Entries) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails("对象实例数组不能为空")
	}

	// entries 数量限制（最大1000）
	if len(query.Entries) > 1000 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails("对象实例数组长度不能超过1000")
	}

	// 验证每个 entry
	for i, entry := range query.Entries {
		// object_type_id 非空
		if entry.ObjectTypeID == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("第%d个对象的对象类型ID不能为空", i+1))
		}

		// _instance_identity 非空
		if len(entry.InstanceIdentity) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("第%d个对象的唯一标识不能为空", i+1))
		}
	}

	return nil
}
