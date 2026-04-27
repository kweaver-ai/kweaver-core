// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package resource

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"
	"go.opentelemetry.io/otel/codes"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	fcond "vega-backend/logics/filter_condition"
)

// 创建和更新视图的一些通用操作
func (rs *resourceService) validateLogicDefinition(ctx context.Context, view *interfaces.ResourceRequest) (string, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "logic layer: Common operation for creating and updating views")
	defer span.End()

	// 自定义视图
	if view.LogicDefinition == nil {
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
			WithErrorDetails("Logic definition is empty")
	}

	// 校验节点ID的唯一性
	nodeMap := make(map[string]struct{})
	for _, node := range view.LogicDefinition {
		if node.ID == "" {
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("Node ID cannot be empty")
		}
		if _, exists := nodeMap[node.ID]; exists {
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_Duplicated_NodeID).
				WithErrorDetails(fmt.Sprintf("Duplicate node ID found: %s", node.ID))
		}
		nodeMap[node.ID] = struct{}{}
	}

	resourceNodeCount := 0
	outputNodeCount := 0
	refResourceMap := make(map[string]*interfaces.Resource)

	for _, node := range view.LogicDefinition {
		switch node.Type {
		case interfaces.LogicDefinitionNodeType_Resource:
			// 校验资源节点
			err := validateResourceNode(ctx, rs, node, refResourceMap)
			if err != nil {
				return "", err
			}

			resourceNodeCount++
		case interfaces.LogicDefinitionNodeType_Join:
			err := validateJoinNode(ctx, node, nodeMap)
			if err != nil {
				return "", err
			}
		case interfaces.LogicDefinitionNodeType_Union:
			err := validateUnionNode(ctx, view.Category, node, nodeMap)
			if err != nil {
				return "", err
			}
		case interfaces.LogicDefinitionNodeType_Sql:
			err := validateSqlNode(ctx, node, nodeMap)
			if err != nil {
				return "", err
			}
		case interfaces.LogicDefinitionNodeType_Output:
			err := validateOutputNode(ctx, node, nodeMap)
			if err != nil {
				return "", err
			}

			outputNodeCount++

		default:
			return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition node type is invalid")
		}
	}

	// 如果只有一个资源节点和一个输出节点，则为衍生视图
	logicType := interfaces.LogicType_Composite
	if len(view.LogicDefinition) == 2 && resourceNodeCount == 1 && outputNodeCount == 1 {
		logicType = interfaces.LogicType_Derived
	}

	var refResourceCategory string
	refResourceCategoryMap := make(map[string]struct{})
	refResourceCatalogMap := make(map[string]struct{})
	for _, dsView := range refResourceMap {
		refResourceCatalogMap[dsView.CatalogID] = struct{}{}
		refResourceCategoryMap[dsView.Category] = struct{}{}
		refResourceCategory = dsView.Category
	}

	if len(refResourceCategoryMap) != 1 {
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The source view of the custom view must have the same category")
	}

	// 如果数据源类型是opensearch，则不能跨opensearch数据源选择
	if refResourceCategory == interfaces.ResourceCategoryIndex && len(refResourceCatalogMap) > 1 {
		return "", rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The source view of query type DSL must have the same data source when create custom view")
	}

	span.SetStatus(codes.Ok, "")
	return logicType, nil
}

func validateResourceNode(ctx context.Context, dvs *resourceService, node *interfaces.LogicDefinitionNode,
	refResourceMap map[string]*interfaces.Resource) error {
	// 资源节点输入节点必须为空
	if len(node.Inputs) != 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The resource node must have no input node")
	}

	var cfg interfaces.ResourceNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusInternalServerError, rest.PublicError_InternalServerError).
			WithErrorDetails(fmt.Sprintf("decode resource node config failed, %v", err))
	}

	// 判断自定义视图的来源表是否存在，从这个函数能够拿到字段列表
	atomicView, err := dvs.GetByID(ctx, cfg.ResourceID)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("get resource %s failed, %v", cfg.ResourceID, err))
	}

	// 校验来源视图的类型
	switch atomicView.Category {
	case interfaces.ResourceCategoryTable:
	case interfaces.ResourceCategoryFile:
	case interfaces.ResourceCategoryFileset:
	case interfaces.ResourceCategoryAPI:
	case interfaces.ResourceCategoryTopic:
	case interfaces.ResourceCategoryIndex:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("The source resource of the custom view '%s' is not supported", cfg.ResourceID))
	}

	refResourceMap[atomicView.ID] = atomicView

	// fieldsMap 是字段name和字段的映射
	fieldsMap := make(map[string]*interfaces.Property)
	for _, viewField := range atomicView.SchemaDefinition {
		fieldsMap[viewField.Name] = viewField
	}

	// 校验过滤条件
	httpErr := validateCond(ctx, cfg.Filters, fieldsMap)
	if httpErr != nil {
		return httpErr
	}

	// 校验去重配置, 只有 table 去重配置
	if cfg.Distinct {
		if atomicView.Category != interfaces.ResourceCategoryTable {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition view category is not table, distinct config is not supported")
		}
	}

	// 校验输出字段是否在视图字段列表里
	for _, field := range node.OutputFields {
		if field.Name == "*" {
			continue
		}
		if _, ok := fieldsMap[field.Name]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, rest.PublicError_BadRequest).
				WithErrorDetails(fmt.Sprintf("The field '%s' is not in the view '%s' field list", field.Name, atomicView.Name))
		}
	}

	return nil
}

func validateJoinNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 仅支持两个视图join
	if len(node.Inputs) != 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid, only support two views join")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, inputs must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition join config is invalid, input '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 join_on
	var cfg interfaces.JoinNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid")
	}

	// join_type 只能为 inner, left, right
	if _, ok := interfaces.JoinTypeMap[cfg.JoinType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_JoinType).
			WithErrorDetails("The logic definition join config is invalid, join_type must be inner, left, right")
	}

	// join_on 校验
	if len(cfg.JoinOn) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition join config is invalid, join_on must be set")
	}

	// join_on 校验
	for _, joinOn := range cfg.JoinOn {
		if joinOn.LeftField == "" || joinOn.RightField == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, join_on left_field and right_field must be set")
		}

		// 操作符必须只为=
		if joinOn.Operator != "=" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition join config is invalid, join_on operator must be =")
		}
	}

	return nil
}

func validateUnionNode(ctx context.Context, category string, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 当前仅支持两个视图union
	if len(node.Inputs) < 2 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid, need at least two views union")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition union config is invalid, inputs must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition union config is invalid, input '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 union config
	var cfg interfaces.UnionNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid")
	}

	if _, ok := interfaces.UnionTypeMap[cfg.UnionType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition union config is invalid, union_type must be all, distinct")
	}

	// TODO 如果查询类型是DSL或索引基类，只允许union all
	if category == interfaces.ResourceCategoryIndex {
		if cfg.UnionType != interfaces.UnionType_All {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition union config is invalid, DSL or IndexBase view only support union all")
		}
	}

	// TODO 校验 output_fields 中每个字段的 FromList 长度是否与 inputs 长度一致
	if category == interfaces.ResourceCategoryTable {
		for _, field := range node.OutputFields {
			if len(field.FromList) != len(node.Inputs) {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
					WithErrorDetails(fmt.Sprintf("The union output field '%s' from list count (%d) not equal inputs count (%d)",
						field.Name, len(field.FromList), len(node.Inputs)))
			}
		}
	}

	return nil
}

func validateSqlNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 输入节点不能为空
	if len(node.Inputs) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid, inputs must be set")
	}

	// 校验输入节点是否重复
	inputNodesMap := make(map[string]struct{})
	for _, inputNode := range node.Inputs {
		if _, ok := inputNodesMap[inputNode]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The logic definition sql config is invalid, inputs must be unique")
		}
		inputNodesMap[inputNode] = struct{}{}
	}

	// 校验输入节点是否存在
	for _, inputNode := range node.Inputs {
		if _, ok := nodeMap[inputNode]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails(fmt.Sprintf("The logic definition sql config is invalid, input '%s' is not exist", inputNode))
		}
	}

	// mapstructure 解析 sql config
	var cfg interfaces.SQLNodeCfg
	err := mapstructure.Decode(node.Config, &cfg)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid")
	}

	// 校验 sql 是否为空
	if cfg.SQL == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The logic definition sql config is invalid, sql must be set")
	}

	return nil
}

func validateOutputNode(ctx context.Context, node *interfaces.LogicDefinitionNode, nodeMap map[string]struct{}) error {
	// 输入节点只能有一个
	if len(node.Inputs) != 1 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The output node must have one input node")
	}

	// 校验输入节点是否存在
	inputNode := node.Inputs[0]
	if _, ok := nodeMap[inputNode]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails(fmt.Sprintf("The output node input '%s' is not exist", inputNode))
	}

	// 如果没传fields字段列表，默认使用output节点的输出字段
	if len(node.OutputFields) == 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
			WithErrorDetails("The output node must have output fields")
	}

	// 校验name不能重复，display_name 不能重复
	nameMap := make(map[string]struct{})
	// originalNameMap := make(map[string]struct{})
	displayNameMap := make(map[string]struct{})
	for _, field := range node.OutputFields {
		if _, ok := nameMap[field.Name]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The output node field name is repeated")
		}
		nameMap[field.Name] = struct{}{}

		// if _, ok := originalNameMap[field.OriginalName]; ok {
		// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
		// 		WithErrorDetails("The output node field original_name is repeated")
		// }
		// originalNameMap[field.OriginalName] = struct{}{}

		if _, ok := displayNameMap[field.DisplayName]; ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_LogicView_InvalidParameter_LogicDefinition).
				WithErrorDetails("The output node field display_name is repeated")
		}
		displayNameMap[field.DisplayName] = struct{}{}
	}

	return nil
}

// 相比handler层的校验，补充对过滤条件字段类型的校验
func validateCond(ctx context.Context, cfg *interfaces.FilterCondCfg, fieldsMap map[string]*interfaces.Property) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤条件字段不允许 __id 和 __routing
	if cfg.Name == "__id" || cfg.Name == "__routing" {
		return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
			WithErrorDetails("The filter field '__id' and '__routing' is not allowed")
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionOperation)
	}

	_, exists := fcond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_UnsupportFilterConditionOperation).
			WithErrorDetails(fmt.Sprintf("unsupport condition operation %s", cfg.Operation))
	}

	switch cfg.Operation {
	case fcond.OperationAnd, fcond.OperationOr:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > interfaces.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_CountExceeded_FilterConditionSubConds).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", interfaces.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionName)
		}
	}

	switch cfg.Operation {
	case fcond.OperationEqual, fcond.OperationNotEqual, fcond.OperationGt, fcond.OperationGte,
		fcond.OperationLt, fcond.OperationLte, fcond.OperationLike, fcond.OperationNotLike,
		fcond.OperationRegex, fcond.OperationMatch, fcond.OperationMatchPhrase, fcond.OperationCurrent:
		// 右侧值为单个值
		_, ok := cfg.Value.([]interface{})
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == fcond.OperationLike || cfg.Operation == fcond.OperationNotLike ||
			cfg.Operation == fcond.OperationPrefix || cfg.Operation == fcond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == fcond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case fcond.OperationIn, fcond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]interface{})) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case fcond.OperationRange, fcond.OperationOutRange, fcond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]interface{})
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[range, out_range, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[range, out_range, between] operation's value must contain 2 values")
		}
	case fcond.OperationBefore:
		// before时, 长度为2的数组，第一个值为时间长度，数值型；第二个值为时间单位，字符串
		_, ok := cfg.Value.(float64)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[before] operation's value must be an array")
		}

		_, ok = cfg.RemainCfg["unit"]
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
				WithErrorDetails("[before] operation's remain cfg must contain unit")
		}
	}

	switch cfg.Operation {
	case fcond.OperationAnd, fcond.OperationOr:
		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond, fieldsMap)
			if err != nil {
				return err
			}
		}
	default:
		// 除 * 之外的过滤字段在视图字段列表里
		if cfg.Name != interfaces.AllField {
			cField, ok := fieldsMap[cfg.Name]
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails(fmt.Sprintf("Filter field '%s' is not in view fields list", cfg.Name))
			}

			fieldType := cField.Type
			// binary 类型的字段不支持过滤
			if fieldType == interfaces.DataType_Binary {
				return rest.NewHTTPError(ctx, http.StatusForbidden, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails("Binary fields do not support filtering")
			}

			// empty, not_empty 的字段类型必须为 string
			if cfg.Operation == fcond.OperationEmpty || cfg.Operation == fcond.OperationNotEmpty {
				if !interfaces.DataType_IsString(fieldType) {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
						WithErrorDetails("Filter field must be of string type when using 'empty' or 'not_empty' operation")
				}
			}
		} else {
			// 如果字段为 *，则只允许使用 match 和 match_phrase 操作符
			if cfg.Operation != fcond.OperationMatch && cfg.Operation != fcond.OperationMatchPhrase &&
				cfg.Operation != fcond.OperationMultiMatch {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
					WithErrorDetails("Filter field '*' only supports 'match', 'match_phrase' and 'multi_match' operations")
			}
		}
	}

	return nil
}

// 解析 logicDefinition，生成 schemaDefinition
func (rs *resourceService) parseLogicDefinition(ctx context.Context,
	logicDefinition []*interfaces.LogicDefinitionNode) ([]*interfaces.Property, error) {

	// 1. 构建节点映射表
	nodes := make(map[string]*interfaces.LogicDefinitionNode)
	for _, node := range logicDefinition {
		nodes[node.ID] = node
	}

	// 2. 找到终端输出节点 (output 节点)
	var outputNode *interfaces.LogicDefinitionNode
	for _, node := range logicDefinition {
		if node.Type == interfaces.LogicDefinitionNodeType_Output {
			outputNode = node
			break
		}
	}

	if outputNode == nil {
		// 如果没显式定义 output 节点，兜底取最后一个节点
		if len(logicDefinition) > 0 {
			outputNode = logicDefinition[len(logicDefinition)-1]
		} else {
			return nil, fmt.Errorf("logic definition is empty")
		}
	}

	// 3. 递归解析字段元数据 (带缓存避免重复计算)
	memo := make(map[string][]*interfaces.Property)
	var resolve func(nodeID string) ([]*interfaces.Property, error)
	resolve = func(nodeID string) ([]*interfaces.Property, error) {
		if cached, ok := memo[nodeID]; ok {
			return cached, nil
		}

		node, ok := nodes[nodeID]
		if !ok {
			return nil, fmt.Errorf("node %s not found in logic definition", nodeID)
		}

		var result []*interfaces.Property
		var inputFieldsMap = make(map[string][]*interfaces.Property)
		var sourceResourceFields []*interfaces.Property

		// 处理叶子节点：Resource 节点
		if node.Type == interfaces.LogicDefinitionNodeType_Resource {
			var cfg interfaces.ResourceNodeCfg
			if err := mapstructure.Decode(node.Config, &cfg); err != nil {
				return nil, fmt.Errorf("decode resource node config failed: %w", err)
			}
			res, err := rs.GetByID(ctx, cfg.ResourceID)
			if err != nil {
				return nil, fmt.Errorf("get resource %s failed: %w", cfg.ResourceID, err)
			}
			sourceResourceFields = res.SchemaDefinition
		} else {
			// 解析所有输入节点的输出字段
			for _, inputID := range node.Inputs {
				fields, err := resolve(inputID)
				if err != nil {
					return nil, err
				}
				inputFieldsMap[inputID] = fields
			}
		}

		// 处理当前节点的 output_fields
		for _, vProp := range node.OutputFields {
			if vProp.Name == "*" {
				// 通配符模式：全量透传上游字段
				if node.Type == interfaces.LogicDefinitionNodeType_Resource {
					for _, f := range sourceResourceFields {
						result = append(result, copyProperty(f))
					}
				} else {
					for _, inputID := range node.Inputs {
						for _, f := range inputFieldsMap[inputID] {
							result = append(result, copyProperty(f))
						}
					}
				}
				continue
			}

			// 投影/映射/对齐/定义模式：构造 Property
			prop := &interfaces.Property{
				Name:         vProp.Name,
				Type:         vProp.Type,
				DisplayName:  vProp.DisplayName,
				OriginalName: vProp.OriginalName,
				Description:  vProp.Description,
				Features:     vProp.Features,
			}

			// 递归溯源补全元数据 (Type, DisplayName, Description, OriginalName, Features)
			var sourceProp *interfaces.Property
			if node.Type == interfaces.LogicDefinitionNodeType_Resource {
				// Resource 节点从物理 Schema 中找
				for _, f := range sourceResourceFields {
					if f.Name == vProp.Name {
						sourceProp = f
						break
					}
				}
			} else if vProp.From != "" && vProp.FromNode != "" {
				// 映射模式 (Join)：明确指定了来源节点和字段
				if sFields, ok := inputFieldsMap[vProp.FromNode]; ok {
					for _, f := range sFields {
						if f.Name == vProp.From {
							sourceProp = f
							break
						}
					}
				}
			} else if len(vProp.FromList) > 0 {
				// 对齐模式 (Union)：从匹配的第一个来源节点取元数据
				for _, ref := range vProp.FromList {
					if sFields, ok := inputFieldsMap[ref.FromNode]; ok {
						for _, f := range sFields {
							if f.Name == ref.From {
								sourceProp = f
								break
							}
						}
					}
					if sourceProp != nil {
						break
					}
				}
			} else {
				// 投影模式/SQL定义：按名称在上游输入中查找
				for _, inputID := range node.Inputs {
					if sFields, ok := inputFieldsMap[inputID]; ok {
						for _, f := range sFields {
							if f.Name == vProp.Name {
								sourceProp = f
								break
							}
						}
					}
					if sourceProp != nil {
						break
					}
				}
			}

			// 如果找到了源字段，则补全缺失的信息
			if sourceProp != nil {
				fillMissingMetadata(prop, sourceProp)
			}
			result = append(result, prop)
		}

		memo[nodeID] = result
		return result, nil
	}

	return resolve(outputNode.ID)
}

func copyProperty(p *interfaces.Property) *interfaces.Property {
	if p == nil {
		return nil
	}
	cp := *p
	if len(p.Features) > 0 {
		cp.Features = make([]interfaces.PropertyFeature, len(p.Features))
		copy(cp.Features, p.Features)
	}
	return &cp
}

func fillMissingMetadata(target, source *interfaces.Property) {
	if target.Type == "" {
		target.Type = source.Type
	}
	if target.DisplayName == "" {
		target.DisplayName = source.DisplayName
	}
	if target.Description == "" {
		target.Description = source.Description
	}
	if target.OriginalName == "" {
		target.OriginalName = source.OriginalName
	}
	if len(target.Features) == 0 {
		target.Features = source.Features
	}
}
