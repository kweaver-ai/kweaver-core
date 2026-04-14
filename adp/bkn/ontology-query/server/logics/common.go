// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package logics

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	dtype "ontology-query/interfaces/data_type"
)

// 构建视图的默认的排序
func BuildViewSort(objectType interfaces.ObjectType) []*interfaces.SortParams {
	sorts := []*interfaces.SortParams{
		{
			Field:     interfaces.SORT_FIELD_SCORE,
			Direction: interfaces.DESC_DIRECTION,
		},
	}
	// 属性到视图的映射。主键是属性名，需要找到主键属性再找到其映射的视图字段名
	propFieldMap := map[string]string{}
	for _, prop := range objectType.DataProperties {
		propFieldMap[prop.Name] = prop.MappedField.Name
	}

	for _, pri := range objectType.PrimaryKeys {
		if fieldName, exists := propFieldMap[pri]; exists {
			// 如果主键映射的字段为空，则不拼排序
			if fieldName != "" {
				// 存在映射，则组装到对象属性中
				sorts = append(sorts, &interfaces.SortParams{
					Field:     pri,
					Direction: interfaces.ASC_DIRECTION,
				})
			}
		}
	}
	return sorts
}

// MapSortFieldsForDataView maps sort fields from object-type data property names to view column names
// (MappedField.Name) for virtualized data-view queries. SORT_FIELD_SCORE is passed through unchanged.
func MapSortFieldsForDataView(sort []*interfaces.SortParams, objectType interfaces.ObjectType) ([]*interfaces.SortParams, error) {
	if len(sort) == 0 {
		return sort, nil
	}
	propNameToViewField := make(map[string]string, len(objectType.DataProperties))
	for _, prop := range objectType.DataProperties {
		propNameToViewField[prop.Name] = prop.MappedField.Name
	}
	out := make([]*interfaces.SortParams, len(sort))
	for i, sp := range sort {
		if sp == nil {
			return nil, fmt.Errorf("排序配置不能为空")
		}
		field := sp.Field
		if field == interfaces.SORT_FIELD_SCORE {
			out[i] = &interfaces.SortParams{Field: field, Direction: sp.Direction}
			continue
		}
		viewField, ok := propNameToViewField[field]
		if !ok {
			return nil, fmt.Errorf("排序字段[%s]不是对象类的数据属性", field)
		}
		if viewField == "" {
			return nil, fmt.Errorf("排序字段[%s]未配置视图映射列(mapped_field)，无法在数据视图上排序", field)
		}
		out[i] = &interfaces.SortParams{Field: viewField, Direction: sp.Direction}
	}
	return out, nil
}

// 构建对象索引的默认的排序
func BuildIndexSort(objectType interfaces.ObjectType, propMap map[string]cond.DataProperty) []*interfaces.SortParams {
	sorts := []*interfaces.SortParams{
		{
			Field:     interfaces.SORT_FIELD_SCORE,
			Direction: interfaces.DESC_DIRECTION,
		},
	}

	for _, pri := range objectType.PrimaryKeys {
		// 存在映射，则组装到对象属性中. 字段还需设置有关键字索引才能进行排序，否则不参与排序

		// text类型的字段需要看其下有没有配置keyword索引，配了就用 xxx.keyword 进行排序。否则不纳入排序
		// string类型的字段直接支持排序，若其有全文索引，则在字段的 keyword 下有 text。
		if propMap[pri].Type == dtype.DATATYPE_TEXT {
			if propMap[pri].IndexConfig != nil && propMap[pri].IndexConfig.KeywordConfig.Enabled {
				sorts = append(sorts, &interfaces.SortParams{
					Field:     pri + "." + dtype.KEYWORD_SUFFIX,
					Direction: interfaces.ASC_DIRECTION,
				})
			}
		} else {
			sorts = append(sorts, &interfaces.SortParams{
				Field:     pri,
				Direction: interfaces.ASC_DIRECTION,
			})
		}

	}
	return sorts
}

// 构建路径键用于循环检测
func BuildPathKey(path interfaces.RelationPath, nextNodeID string) string {
	key := fmt.Sprintf("%s:%s", path.Relations[0].RelationTypeId, path.Relations[0].SourceObjectId)
	for i := 0; i < len(path.Relations); i++ {
		key += fmt.Sprintf("->%s", path.Relations[i].TargetObjectId)
	}
	key += fmt.Sprintf("->%s", nextNodeID)
	return key
}

// 过滤有效路径（排除包含循环的路径）
func FilterValidPaths(paths []interfaces.RelationPath, visitedNodes map[string]bool) []interfaces.RelationPath {
	var validPaths []interfaces.RelationPath

	for _, path := range paths {
		if IsPathValid(path, visitedNodes) {
			validPaths = append(validPaths, path)
		}
	}

	return validPaths
}

// 检查路径是否有效（无循环）
func IsPathValid(path interfaces.RelationPath, visitedNodes map[string]bool) bool {
	if len(path.Relations) == 0 {
		return true
	}

	nodeSet := make(map[string]bool)
	var prevTargetId string

	// 检查路径中的每个节点是否唯一，同时检查路径连续性
	for i, relation := range path.Relations {
		// 检查路径连续性：除了第一条边外，前一条边的目标节点应该等于当前边的源节点
		if i > 0 {
			if prevTargetId != relation.SourceObjectId {
				logger.Debugf("检测到路径不连续 - 前一条边的目标节点[%s]不等于当前边的源节点[%s]", prevTargetId, relation.SourceObjectId)
				return false
			}
		}

		// 对于第一条边：检查源节点是否已访问（不应重复）
		if i == 0 {
			if nodeSet[relation.SourceObjectId] {
				logger.Debugf("检测到路径起始节点重复: %s", relation.SourceObjectId)
				return false
			}
			nodeSet[relation.SourceObjectId] = true
		}

		// 对于所有边：检查目标节点是否已访问（如果已访问，说明形成了循环）
		if nodeSet[relation.TargetObjectId] {
			logger.Debugf("检测到循环路径 - 重复节点: %s", relation.TargetObjectId)
			return false
		}
		nodeSet[relation.TargetObjectId] = true

		// 检查是否与已访问的节点冲突（如果提供了visitedNodes参数）
		if visitedNodes != nil {
			if visitedNodes[relation.SourceObjectId] || visitedNodes[relation.TargetObjectId] {
				logger.Debugf("检测到路径与已访问节点冲突")
				return false
			}
		}

		prevTargetId = relation.TargetObjectId
	}

	return true
}

// 检查限制
func CanGenerate(quotaManager *interfaces.PathQuotaManager, pathID int) bool {
	if quotaManager == nil {
		return true
	}

	// 检查全局限制
	currentGlobal := atomic.LoadInt64(&quotaManager.GlobalCount)
	if currentGlobal >= quotaManager.TotalLimit {
		logger.Debugf("达到全局限制: %d/%d", currentGlobal, quotaManager.TotalLimit)
		return false
	}

	used := 0
	if value, exist := quotaManager.UsedQuota.Load(pathID); !exist {
		quotaManager.UsedQuota.Store(pathID, used)
	} else {
		used = value.(int)
	}

	if quotaManager.RequestPathTypeNum > 1 {
		// 动态配额：根据权重和剩余总量计算
		maxQuota := quotaManager.TotalLimit - quotaManager.GlobalCount
		return int64(used) < maxQuota
	} else {
		// 总数小于限制数，可添加
		return quotaManager.GlobalCount < quotaManager.TotalLimit
	}
}

// RecordGenerated 记录已生成的路径
func RecordGenerated(quotaManager *interfaces.PathQuotaManager, typePathID int, cnt int) {
	if quotaManager == nil {
		return
	}

	// 原子操作增加全局计数
	newGlobalCount := atomic.AddInt64(&quotaManager.GlobalCount, int64(cnt))

	// 更新特定路径类型的配额使用情况
	if value, exist := quotaManager.UsedQuota.Load(typePathID); !exist {
		quotaManager.UsedQuota.Store(typePathID, cnt)
	} else {
		quotaManager.UsedQuota.Store(typePathID, value.(int)+cnt)
	}
	logger.Debugf("路径配额更新 - 路径ID: %d, 新增: %d, 全局计数: %d/%d",
		typePathID, cnt, newGlobalCount, quotaManager.TotalLimit)
}

// 从对象数据中提取对象ID
func GetObjectID(objectData map[string]any, objectType *interfaces.ObjectType) (string, map[string]any) {
	if objectType == nil || len(objectType.PrimaryKeys) == 0 {
		return "", nil
	}

	// 使用主键构建对象ID
	var idParts []string
	uk := map[string]any{}
	for _, pk := range objectType.PrimaryKeys {
		if value, exists := objectData[pk]; exists {
			idParts = append(idParts, fmt.Sprintf("%v", value))
			uk[pk] = value
		} else {
			idParts = append(idParts, "__NULL__")
		}
	}

	if len(idParts) == 0 {
		return "", uk
	}

	return objectType.OTID + "-" + strings.Join(idParts, "_"), uk
}

// 构建直接映射的批量条件
func BuildDirectBatchConditions(currentLevelObjects []interfaces.LevelObject,
	edge *interfaces.TypeEdge, isForward bool) ([]*cond.CondCfg, error) {

	var conditions []*cond.CondCfg
	var inValues []any
	var inField string
	mappingRules := edge.RelationType.MappingRules.([]interfaces.Mapping)

	for _, levelObj := range currentLevelObjects {
		// 按关联关系构建了过滤子句。
		// 多个字段关联，需用and连接各个对象的过滤条件，然后再用or拼接各个对象的过滤条件
		// 1个关联字段，则对多个对象的过滤条件采用in操作
		objectConditions, targetField, inValue := BuildCondition(nil, mappingRules, isForward, levelObj.ObjectData)
		if inValue != nil {
			inValues = append(inValues, inValue)
		}
		inField = targetField

		// 一个对象下如果是多个过滤子句，则用and关联
		if len(objectConditions) > 1 {
			conditions = append(conditions, &cond.CondCfg{
				Operation: "and",
				SubConds:  objectConditions,
			})
		} else if len(objectConditions) == 1 {
			conditions = append(conditions, objectConditions[0])
		}
	}

	if len(mappingRules) == 1 && len(inValues) > 0 {
		return []*cond.CondCfg{
			{
				Name:      inField,
				Operation: "in",
				ValueOptCfg: cond.ValueOptCfg{
					ValueFrom: "const",
					Value:     inValues,
				},
			},
		}, nil
	}

	return conditions, nil
}

func BuildCondition(viewQuery *interfaces.ViewQuery, mappingRules []interfaces.Mapping,
	isForward bool, currentObjectData map[string]any) ([]*cond.CondCfg, string, any) {

	conditions := []*cond.CondCfg{}
	var inValue any
	var targetField string
	// 视图作为中间表时，查询视图数据按 _score desc, 关联字段 asc
	sort := []*interfaces.SortParams{
		{
			Field:     interfaces.SORT_FIELD_SCORE,
			Direction: interfaces.DESC_DIRECTION,
		},
	}

	for _, mapping := range mappingRules {
		// 默认先取正向的，若是反向，再修改起终点字段
		targetName := mapping.TargetProp.Name
		sourceName := mapping.SourceProp.Name
		if !isForward {
			targetName = mapping.SourceProp.Name
			sourceName = mapping.TargetProp.Name
		}
		// 一个关联字段，则取字段值作为in的过滤条件
		if len(mappingRules) == 1 {
			inValue = currentObjectData[sourceName]
			targetField = targetName
		}
		// 多个字段关联，则构造多个过滤条件，在上层用and连接
		conditions = append(conditions, &cond.CondCfg{
			Name:      targetName, // 注意正向反向的差别
			Operation: "==",       // 关联时只有等于的关系
			ValueOptCfg: cond.ValueOptCfg{
				ValueFrom: "const",
				Value:     currentObjectData[sourceName], // 从起点对象中获取的起点属性
			},
		})
		sort = append(sort, &interfaces.SortParams{
			Field:     targetName,
			Direction: interfaces.ASC_DIRECTION,
		})
	}

	if viewQuery != nil {
		if len(conditions) > 0 {
			viewQuery.Filters = &cond.CondCfg{
				Operation: "and",
				SubConds:  conditions,
			}
		}
		viewQuery.Sort = sort
	}

	return conditions, targetField, inValue
}

// 检查直接映射条件
func CheckDirectMappingConditions(currentObjectData map[string]any,
	nextObject map[string]any, mappingRules []interfaces.Mapping, isForward bool) bool {

	for _, mapping := range mappingRules {
		var sourceValue, targetValue interface{}
		var ok bool

		if isForward {
			// 正向：currentObject -> nextObject
			sourceValue, ok = currentObjectData[mapping.SourceProp.Name]
			if !ok {
				return false
			}
			targetValue, ok = nextObject[mapping.TargetProp.Name]
			if !ok {
				return false
			}
		} else {
			// 反向：nextObject -> currentObject
			sourceValue, ok = nextObject[mapping.SourceProp.Name]
			if !ok {
				return false
			}
			targetValue, ok = currentObjectData[mapping.TargetProp.Name]
			if !ok {
				return false
			}
		}

		// 比较值是否相等
		if !CompareValues(sourceValue, targetValue) {
			return false
		}
	}

	return true
}

// 比较两个值是否相等（处理不同类型的情况）
func CompareValues(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// 转换为字符串比较，避免类型不匹配的问题
	aStr := fmt.Sprintf("%v", a)
	bStr := fmt.Sprintf("%v", b)

	return aStr == bStr
}

// 检查视图数据是否满足查询条件
func CheckViewDataMatchesCondition(viewData map[string]any,
	condition *cond.CondCfg, mappingRules []interfaces.Mapping, isForward bool) bool {

	// 简化实现：根据映射规则检查视图数据
	for _, mapping := range mappingRules {
		var fieldName string
		if isForward {
			fieldName = mapping.TargetProp.Name
		} else {
			fieldName = mapping.SourceProp.Name
		}

		expectedValue, exists := viewData[fieldName]
		if !exists {
			return false
		}

		// 比较值是否相等。因为关系关联都是等于的关系，直接取值比较
		// 从条件里取值，不能只取一个，还需要考虑多字段关联的情况
		conditionValue := condition.ValueOptCfg.Value
		if !CompareValues(expectedValue, conditionValue) {
			return false
		}
	}

	return true
}

// 使用视图数据检查间接映射条件
func CheckIndirectMappingConditionsWithViewData(currentObjectData map[string]any,
	nextObject map[string]any, mappingRules *interfaces.InDirectMapping, isForward bool,
	viewData []map[string]any) bool {

	// 检查是否存在一个视图记录能够连接当前对象和下一层对象
	for _, viewRecord := range viewData {
		// 检查当前对象 -> 视图记录的映射
		var sourceMapping []interfaces.Mapping
		if isForward {
			sourceMapping = mappingRules.SourceMappingRules
		} else {
			sourceMapping = mappingRules.TargetMappingRules
		}

		sourceMatch := true
		for _, mapping := range sourceMapping {
			sorcePropName := ""
			targetPropName := ""
			if isForward {
				sorcePropName = mapping.SourceProp.Name
				targetPropName = mapping.TargetProp.Name
			} else {
				sorcePropName = mapping.TargetProp.Name
				targetPropName = mapping.SourceProp.Name
			}
			currentValue, ok1 := currentObjectData[sorcePropName]
			viewValue, ok2 := viewRecord[targetPropName]
			if !ok1 || !ok2 || !CompareValues(currentValue, viewValue) {
				sourceMatch = false
				break
			}
		}
		if !sourceMatch {
			continue
		}

		// 检查视图记录 -> 下一层对象的映射
		var targetMapping []interfaces.Mapping
		if isForward {
			targetMapping = mappingRules.TargetMappingRules
		} else {
			targetMapping = mappingRules.SourceMappingRules
		}

		targetMatch := true
		for _, mapping := range targetMapping {
			sorcePropName := ""
			targetPropName := ""
			if isForward {
				sorcePropName = mapping.SourceProp.Name
				targetPropName = mapping.TargetProp.Name
			} else {
				sorcePropName = mapping.TargetProp.Name
				targetPropName = mapping.SourceProp.Name
			}

			viewValue, ok1 := viewRecord[sorcePropName]
			nextValue, ok2 := nextObject[targetPropName]
			if !ok1 || !ok2 || !CompareValues(viewValue, nextValue) {
				targetMatch = false
				break
			}
		}

		if targetMatch {
			return true
		}
	}

	return false
}

// 根据对象唯一标识构建对象查询的过滤条件
func BuildInstanceIdentitiesCondition(uks []map[string]any) *cond.CondCfg {

	if len(uks) == 0 {
		return nil
	}

	ukSubConds := []*cond.CondCfg{}
	for _, uk := range uks { // 多个对象
		conds := []*cond.CondCfg{}
		for k, v := range uk { // 联合主键
			conds = append(conds, &cond.CondCfg{
				Name:      k,
				Operation: "==",
				ValueOptCfg: cond.ValueOptCfg{
					Value: v,
				},
			})
		}
		ukSubConds = append(ukSubConds, &cond.CondCfg{
			Operation: "and",
			SubConds:  conds,
		})
	}
	var ukCondition *cond.CondCfg
	if len(ukSubConds) > 1 {
		ukCondition = &cond.CondCfg{
			Operation: "or",
			SubConds:  ukSubConds,
		}
	} else {
		ukCondition = ukSubConds[0]
	}
	return ukCondition
}

func TransferPropsToPropMap(props []cond.DataProperty) map[string]*cond.DataProperty {
	propMap := map[string]*cond.DataProperty{}
	for _, prop := range props {
		propMap[prop.Name] = &prop
		// 后面计划:若修改了索引配置,则把索引状态设置为不可用,那么就走虚拟化查询,不走持久化.所以查询这里可以认为是准确的
		// if prop.IndexConfig != nil && prop.IndexConfig.FulltextConfig.Enabled {
		// 	// 配置了全文索引的,字段类型定义为text
		// 	propMap[prop.Name] = prop
		// } else if prop.IndexConfig != nil && prop.IndexConfig.VectorConfig.Enabled {
		// 	// 配置了向量化的字段
		// 	propMap[prop.Name] = prop

		// 	propMap[prop.Name] = prop
		// } else {
		// 	// 其他情况
		// 	propMap[prop.Name] = prop
		// }
	}

	return propMap
}

// 构建dsl的query
func BuildDslQuery(ctx context.Context, queryStr string, query *interfaces.ObjectQueryBaseOnObjectType) (map[string]any, error) {

	var dslMap map[string]any
	err := json.Unmarshal([]byte(queryStr), &dslMap)
	if err != nil {
		return map[string]any{}, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_InternalError_UnMarshalDataFailed).
			WithErrorDetails(fmt.Sprintf("failed to unMarshal dslStr to map, %s", err.Error()))
	}

	// 处理 sort
	sort := []map[string]any{}
	for _, sp := range query.Sort {
		// 不做排序字段参数校验了，如果排序字段不存在，opensearch会报错，由opensearch来报错
		// _score 是传递给视图的字段,这里是直接查opensearch,那么这个_score需要改为_score
		// if sp.Field == interfaces.SORT_FIELD_SCORE {
		// 	sort = append(sort, map[string]any{
		// 		"_score": sp.Direction,
		// 	})
		// } else {
		sort = append(sort, map[string]any{
			sp.Field: sp.Direction,
		})
		// }
	}

	if len(query.SearchAfter) > 0 {
		query.NeedTotal = false
	}

	// 如果没传 limit，传了 search_after 参数，设置默认limit 10000
	if query.Limit == 0 && query.SearchAfter != nil && len(query.SearchAfter) > 0 {
		query.Limit = interfaces.SearchAfter_Limit
	}

	dsl := map[string]any{
		"size":         query.Limit,
		"sort":         sort,
		"track_scores": true,
	}
	if len(dslMap) > 0 {
		dsl["query"] = dslMap
	}

	// 存在search after时需加上search_after
	if query.SearchAfter != nil {
		dsl["search_after"] = query.SearchAfter
	}

	return dsl, nil
}

// shouldExcludeSystemProperty 检查是否应该排除指定的系统字段
func ShouldExcludeSystemProperty(fieldName string, excludeList []string) bool {
	if len(excludeList) == 0 {
		return false
	}
	for _, excludeField := range excludeList {
		if excludeField == fieldName {
			return true
		}
	}
	return false
}

// EvaluateInstanceAgainstCondition evaluates whether instance data satisfies the condition
func EvaluateInstanceAgainstCondition(ctx context.Context,
	instanceData map[string]any,
	condition *cond.CondCfg,
	objectType *interfaces.ObjectType) (bool, error) {

	if condition == nil {
		return true, nil
	}

	// Build property map for quick lookup
	propMap := make(map[string]*cond.DataProperty)
	for i := range objectType.DataProperties {
		propMap[objectType.DataProperties[i].Name] = &objectType.DataProperties[i]
	}

	return evaluateConditionRecursive(ctx, instanceData, condition, propMap)
}

// EvaluateDataAgainstCondition 判断 data 是否满足 condition
// data: 待评估数据，key 为字段名
// condition: 条件配置
// paramDefs: 参数定义，用于类型推断（可为 nil，则 fieldType 为空）
func EvaluateDataAgainstCondition(ctx context.Context,
	data map[string]any,
	condition *cond.CondCfg,
	paramDefs []interfaces.Parameter) (bool, error) {

	if condition == nil {
		return true, nil
	}

	propMap := make(map[string]*cond.DataProperty)
	for i := range paramDefs {
		propMap[paramDefs[i].Name] = &cond.DataProperty{
			Name: paramDefs[i].Name,
			Type: paramDefs[i].Type,
		}
	}
	return evaluateConditionRecursive(ctx, data, condition, propMap)
}

// evaluateConditionRecursive recursively evaluates condition
func evaluateConditionRecursive(ctx context.Context,
	instanceData map[string]any,
	condition *cond.CondCfg,
	propMap map[string]*cond.DataProperty) (bool, error) {

	if condition == nil {
		return true, nil
	}

	// Handle logical operators
	switch condition.Operation {
	case cond.OperationAnd:
		for _, subCond := range condition.SubConds {
			result, err := evaluateConditionRecursive(ctx, instanceData, subCond, propMap)
			if err != nil {
				return false, err
			}
			if !result {
				return false, nil
			}
		}
		return true, nil

	case cond.OperationOr:
		for _, subCond := range condition.SubConds {
			result, err := evaluateConditionRecursive(ctx, instanceData, subCond, propMap)
			if err != nil {
				return false, err
			}
			if result {
				return true, nil
			}
		}
		return false, nil
	}

	// Handle field-based operators
	fieldName := condition.Name
	fieldValue, fieldExists := instanceData[fieldName]

	// Handle operators that don't require field existence
	switch condition.Operation {
	case cond.OperationExist:
		return fieldExists, nil
	case cond.OperationNotExist:
		return !fieldExists, nil
	case cond.OperationEmpty:
		if !fieldExists {
			return true, nil
		}
		return isEmpty(fieldValue), nil
	case cond.OperationNotEmpty:
		if !fieldExists {
			return false, nil
		}
		return !isEmpty(fieldValue), nil
	case cond.OperationNull:
		return !fieldExists || fieldValue == nil, nil
	case cond.OperationNotNull:
		return fieldExists && fieldValue != nil, nil
	}

	// For other operators, field must exist (except for != and not_like which can return true if field doesn't exist)
	if !fieldExists {
		if condition.Operation == cond.OperationNotEq || condition.Operation == cond.OperationNotLike {
			return true, nil
		}
		return false, nil
	}

	// Get property info for type checking
	prop, hasProp := propMap[fieldName]
	fieldType := ""
	if hasProp {
		fieldType = prop.Type
	}

	// Evaluate based on operation
	switch condition.Operation {
	case cond.OperationEq:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, "==", fieldType)
	case cond.OperationNotEq:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, "!=", fieldType)
	case cond.OperationGt:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, ">", fieldType)
	case cond.OperationGte:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, ">=", fieldType)
	case cond.OperationLt:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, "<", fieldType)
	case cond.OperationLte:
		return compareValues(fieldValue, condition.ValueOptCfg.Value, "<=", fieldType)
	case cond.OperationIn:
		return evaluateIn(fieldValue, condition.ValueOptCfg.Value)
	case cond.OperationNotIn:
		result, err := evaluateIn(fieldValue, condition.ValueOptCfg.Value)
		if err != nil {
			return false, err
		}
		return !result, nil
	case cond.OperationRange:
		return evaluateRange(fieldValue, condition.ValueOptCfg.Value, fieldType, true)
	case cond.OperationOutRange:
		result, err := evaluateRange(fieldValue, condition.ValueOptCfg.Value, fieldType, true)
		if err != nil {
			return false, err
		}
		return !result, nil
	case cond.OperationTrue:
		return isTrue(fieldValue), nil
	case cond.OperationFalse:
		return isFalse(fieldValue), nil
	case cond.OperationBefore:
		return evaluateBefore(fieldValue, condition.ValueOptCfg.Value)
	case cond.OperationBetween:
		return evaluateRange(fieldValue, condition.ValueOptCfg.Value, fieldType, false)
	case cond.OperationLike:
		pattern, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("like operation requires string value")
		}
		return evaluateLike(fieldValue, pattern)
	case cond.OperationNotLike:
		pattern, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("not_like operation requires string value")
		}
		result, err := evaluateLike(fieldValue, pattern)
		if err != nil {
			return false, err
		}
		return !result, nil
	case cond.OperationPrefix:
		prefix, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("prefix operation requires string value")
		}
		return evaluatePrefix(fieldValue, prefix)
	case cond.OperationNotPrefix:
		prefix, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("not_prefix operation requires string value")
		}
		result, err := evaluatePrefix(fieldValue, prefix)
		if err != nil {
			return false, err
		}
		return !result, nil
	case cond.OperationRegex:
		pattern, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("regex operation requires string value")
		}
		return evaluateRegex(fieldValue, pattern)
	case cond.OperationContain:
		return evaluateContain(fieldValue, condition.ValueOptCfg.Value)
	case cond.OperationNotContain:
		result, err := evaluateContain(fieldValue, condition.ValueOptCfg.Value)
		if err != nil {
			return false, err
		}
		return !result, nil
	case cond.OperationCurrent:
		unit, ok := condition.ValueOptCfg.Value.(string)
		if !ok {
			return false, fmt.Errorf("current operation requires string value (year/month/week/day/hour/minute)")
		}
		return evaluateCurrent(fieldValue, unit)
	default:
		return false, fmt.Errorf("unsupported operation: %s", condition.Operation)
	}
}

// isEmpty checks if a value is empty
func isEmpty(value any) bool {
	if value == nil {
		return true
	}

	switch v := value.(type) {
	case string:
		return v == ""
	case []any:
		return len(v) == 0
	case []string:
		return len(v) == 0
	case map[string]any:
		return len(v) == 0
	case map[any]any:
		return len(v) == 0
	}

	// Use reflection for other types
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array, reflect.Map:
		return rv.Len() == 0
	case reflect.String:
		return rv.Len() == 0
	}

	return false
}

// isTrue checks if value is boolean true
func isTrue(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return false
}

// isFalse checks if value is boolean false
func isFalse(value any) bool {
	if b, ok := value.(bool); ok {
		return !b
	}
	return false
}

// evaluateIn checks if fieldValue is in the value list
func evaluateIn(fieldValue any, valueList any) (bool, error) {
	if valueList == nil {
		return false, nil
	}

	list, ok := valueList.([]any)
	if !ok {
		return false, fmt.Errorf("in operation requires array value")
	}

	for _, v := range list {
		if compareValuesSimple(fieldValue, v) {
			return true, nil
		}
	}
	return false, nil
}

// evaluateRange checks if fieldValue is in range [min, max) or [min, max] based on inclusive
func evaluateRange(fieldValue any, rangeValue any, fieldType string, leftClosedRightOpen bool) (bool, error) {
	if rangeValue == nil {
		return false, nil
	}

	rangeList, ok := rangeValue.([]any)
	if !ok || len(rangeList) != 2 {
		return false, fmt.Errorf("range operation requires array of length 2")
	}

	min := rangeList[0]
	max := rangeList[1]

	// Compare fieldValue with min
	minResult, err := compareValues(fieldValue, min, ">=", fieldType)
	if err != nil {
		return false, err
	}
	if !minResult {
		return false, nil
	}

	// Compare fieldValue with max
	if leftClosedRightOpen {
		maxResult, err := compareValues(fieldValue, max, "<", fieldType)
		if err != nil {
			return false, err
		}
		return maxResult, nil
	} else {
		maxResult, err := compareValues(fieldValue, max, "<=", fieldType)
		if err != nil {
			return false, err
		}
		return maxResult, nil
	}
}

// evaluateBefore checks if fieldValue (date) is before conditionValue
func evaluateBefore(fieldValue any, conditionValue any) (bool, error) {
	fieldTime, err := parseTime(fieldValue)
	if err != nil {
		return false, err
	}

	condTime, err := parseTime(conditionValue)
	if err != nil {
		return false, err
	}

	return fieldTime.Before(condTime), nil
}

// evaluateLike checks if fieldValue matches the like pattern (pattern is wrapped as %pattern% for contains semantics)
func evaluateLike(fieldValue any, pattern string) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}
	fieldStr := fmt.Sprintf("%v", fieldValue)
	// Like semantics: %pattern% means contains; use ReplaceLikeWildcards then wrap with .* for full match
	regexPattern := common.ReplaceLikeWildcards(pattern)
	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return false, fmt.Errorf("like pattern invalid: %w", err)
	}
	return re.MatchString(fieldStr), nil
}

// evaluatePrefix checks if fieldValue has the given prefix
func evaluatePrefix(fieldValue any, prefix string) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}
	fieldStr := fmt.Sprintf("%v", fieldValue)
	return strings.HasPrefix(fieldStr, prefix), nil
}

// evaluateRegex checks if fieldValue matches the regex pattern
func evaluateRegex(fieldValue any, pattern string) (bool, error) {
	if fieldValue == nil {
		return false, nil
	}
	fieldStr := fmt.Sprintf("%v", fieldValue)
	re, err := regexp2.Compile(pattern, regexp2.RE2)
	if err != nil {
		return false, fmt.Errorf("regex pattern invalid: %w", err)
	}
	matched, err := re.MatchString(fieldStr)
	if err != nil {
		return false, err
	}
	return matched, nil
}

// evaluateContain checks if fieldValue (array or string) contains target (single value or array)
func evaluateContain(fieldValue any, target any) (bool, error) {
	if fieldValue == nil || target == nil {
		return false, nil
	}

	// Target as array: field must contain all elements
	if targetList, ok := target.([]any); ok {
		if len(targetList) == 0 {
			return false, fmt.Errorf("contain operation requires non-empty array value")
		}
		for _, t := range targetList {
			contains, err := evaluateContainSingle(fieldValue, t)
			if err != nil {
				return false, err
			}
			if !contains {
				return false, nil
			}
		}
		return true, nil
	}

	return evaluateContainSingle(fieldValue, target)
}

// evaluateContainSingle checks if fieldValue contains a single target value
func evaluateContainSingle(fieldValue any, target any) (bool, error) {
	// Field is string: substring contain
	if fieldStr, ok := fieldValue.(string); ok {
		targetStr := fmt.Sprintf("%v", target)
		return strings.Contains(fieldStr, targetStr), nil
	}

	// Field is slice/array: element contain
	rv := reflect.ValueOf(fieldValue)
	switch rv.Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			elem := rv.Index(i).Interface()
			if compareValuesSimple(elem, target) {
				return true, nil
			}
		}
		return false, nil
	}

	return false, fmt.Errorf("contain operation requires string or array field, got %T", fieldValue)
}

// evaluateCurrent checks if fieldValue (datetime) falls within the current year/month/week/day/hour/minute
func evaluateCurrent(fieldValue any, unit string) (bool, error) {
	fieldTime, err := parseTime(fieldValue)
	if err != nil {
		return false, err
	}

	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}
	location, err := time.LoadLocation(tz)
	if err != nil {
		location = time.UTC
	}

	now := time.Now().In(location)
	var start, end time.Time

	switch unit {
	case "year":
		start = time.Date(now.Year(), 1, 1, 0, 0, 0, 0, location)
		end = start.AddDate(1, 0, 0)
	case "month":
		start = time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, location)
		end = start.AddDate(0, 1, 0)
	case "week":
		weekday := now.Weekday()
		offset := int(time.Monday - weekday)
		if offset > 0 {
			offset -= 7
		}
		start = time.Date(now.Year(), now.Month(), now.Day()+offset, 0, 0, 0, 0, location)
		end = start.AddDate(0, 0, 7)
	case "day":
		start = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, location)
		end = start.AddDate(0, 0, 1)
	case "hour":
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, location)
		end = start.Add(time.Hour)
	case "minute":
		start = time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, location)
		end = start.Add(time.Minute)
	default:
		return false, fmt.Errorf("current operation requires unit year/month/week/day/hour/minute, got %s", unit)
	}

	// fieldTime must be in [start, end)
	return !fieldTime.Before(start) && fieldTime.Before(end), nil
}

// parseTime parses time from various formats
func parseTime(value any) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("time value is nil")
	}

	// Try to parse as timestamp (int64 or float64)
	switch v := value.(type) {
	case int64:
		return time.Unix(v, 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case string:
		// Try parsing as RFC3339
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			return t, nil
		}
		// Try parsing as Unix timestamp string
		if ts, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.Unix(ts, 0), nil
		}
		return time.Time{}, fmt.Errorf("unable to parse time string: %s", v)
	default:
		return time.Time{}, fmt.Errorf("unsupported time type: %T", value)
	}
}

// compareValues compares two values based on operation
func compareValues(left any, right any, operation string, fieldType string) (bool, error) {
	if left == nil || right == nil {
		return false, nil
	}

	// Handle datetime comparison
	if fieldType == dtype.DATATYPE_DATETIME {
		leftTime, err := parseTime(left)
		if err != nil {
			return false, err
		}
		rightTime, err := parseTime(right)
		if err != nil {
			return false, err
		}

		switch operation {
		case "==":
			return leftTime.Equal(rightTime), nil
		case "!=":
			return !leftTime.Equal(rightTime), nil
		case ">":
			return leftTime.After(rightTime), nil
		case ">=":
			return leftTime.After(rightTime) || leftTime.Equal(rightTime), nil
		case "<":
			return leftTime.Before(rightTime), nil
		case "<=":
			return leftTime.Before(rightTime) || leftTime.Equal(rightTime), nil
		}
	}

	// Handle numeric comparison
	if isNumeric(left) && isNumeric(right) {
		leftNum := toFloat64(left)
		rightNum := toFloat64(right)

		switch operation {
		case "==":
			return leftNum == rightNum, nil
		case "!=":
			return leftNum != rightNum, nil
		case ">":
			return leftNum > rightNum, nil
		case ">=":
			return leftNum >= rightNum, nil
		case "<":
			return leftNum < rightNum, nil
		case "<=":
			return leftNum <= rightNum, nil
		}
	}

	// Handle string comparison
	leftStr := fmt.Sprintf("%v", left)
	rightStr := fmt.Sprintf("%v", right)

	switch operation {
	case "==":
		return leftStr == rightStr, nil
	case "!=":
		return leftStr != rightStr, nil
	case ">":
		return leftStr > rightStr, nil
	case ">=":
		return leftStr >= rightStr, nil
	case "<":
		return leftStr < rightStr, nil
	case "<=":
		return leftStr <= rightStr, nil
	}

	return false, fmt.Errorf("unsupported operation: %s", operation)
}

// compareValuesSimple compares two values for equality (used in in/not_in)
func compareValuesSimple(left any, right any) bool {
	if left == nil && right == nil {
		return true
	}
	if left == nil || right == nil {
		return false
	}

	// Try direct comparison first
	if left == right {
		return true
	}

	// Try numeric comparison
	if isNumeric(left) && isNumeric(right) {
		return toFloat64(left) == toFloat64(right)
	}

	// String comparison
	return fmt.Sprintf("%v", left) == fmt.Sprintf("%v", right)
}

// isNumeric checks if value is numeric
func isNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64:
		return true
	case uint, uint8, uint16, uint32, uint64:
		return true
	case float32, float64:
		return true
	}
	return false
}

// toFloat64 converts numeric value to float64
func toFloat64(value any) float64 {
	switch v := value.(type) {
	case int:
		return float64(v)
	case int8:
		return float64(v)
	case int16:
		return float64(v)
	case int32:
		return float64(v)
	case int64:
		return float64(v)
	case uint:
		return float64(v)
	case uint8:
		return float64(v)
	case uint16:
		return float64(v)
	case uint32:
		return float64(v)
	case uint64:
		return float64(v)
	case float32:
		return float64(v)
	case float64:
		return v
	}
	return 0
}

// CondCfgToFilterMap serializes a condition tree for vega resource filter_condition JSON.
func CondCfgToFilterMap(c *cond.CondCfg) map[string]any {
	if c == nil {
		return nil
	}
	raw, err := json.Marshal(c)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}
