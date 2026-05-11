// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func ValidateRelationTypes(ctx context.Context, knID string, relationTypes []*interfaces.RelationType, strictMode bool) error {
	idMap := make(map[string]any)
	for i := 0; i < len(relationTypes); i++ {
		// 校验导入模型时模块是否是关系类
		if relationTypes[i].ModuleType != "" && relationTypes[i].ModuleType != interfaces.MODULE_TYPE_RELATION_TYPE {
			return rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
				WithErrorDetails("Relation type name is not 'relation_type'")
		}

		// 0.校验请求体中多个模型 ID 是否重复
		rtID := relationTypes[i].RTID
		if _, ok := idMap[rtID]; !ok || rtID == "" {
			idMap[rtID] = nil
		} else {
			errDetails := fmt.Sprintf("RelationType ID '%s' already exists in the request body", rtID)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_Duplicated_IDInFile).
				WithDescription(map[string]any{"relationTypeID": rtID}).
				WithErrorDetails(errDetails)
		}

		// 1. 校验 关系类必要创建参数的合法性, 非空、长度、是枚举值
		err := ValidateRelationType(ctx, relationTypes[i], strictMode)
		if err != nil {
			return err
		}

		// 2. 如果关系类的branch为空，则赋值 main
		if relationTypes[i].Branch == "" {
			relationTypes[i].Branch = interfaces.MAIN_BRANCH
		}

		relationTypes[i].KNID = knID
	}
	return nil
}

// 对象类必要创建参数的非空校验。
func ValidateRelationType(ctx context.Context, relationType *interfaces.RelationType, strictMode bool) error {
	// 校验id的合法性
	err := validateID(ctx, relationType.RTID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉名称的前后空格
	relationType.RTName = strings.TrimSpace(relationType.RTName)
	err = validateObjectName(ctx, relationType.RTName, interfaces.MODULE_TYPE_RELATION_TYPE)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, relationType.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	relationType.Tags = libCommon.TagSliceTransform(relationType.Tags)

	// 校验type字段
	if relationType.Type != "" {
		if relationType.Type != interfaces.RELATION_TYPE_DIRECT && relationType.Type != interfaces.RELATION_TYPE_DATA_VIEW &&
			relationType.Type != interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("关系类类型只支持 %s、%s 和 %s，当前类型为: %s",
					interfaces.RELATION_TYPE_DIRECT, interfaces.RELATION_TYPE_DATA_VIEW,
					interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN, relationType.Type))
		}
	}

	if relationType.SourceObjectTypeID == "" {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("关系类的 source_object_type_id 不能为空")
		}
	}
	if relationType.TargetObjectTypeID == "" {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("关系类的 target_object_type_id 不能为空")
		}
	}

	// 校验mapping_rules字段
	if relationType.MappingRules == nil {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("关系类的映射规则 mapping_rules 不能为空")
		}
		return nil
	}

	// 如果mapping_rules不为空，type必须非空
	if relationType.Type == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("关系类的关系类型 type 字段不能为空")
	}

	rules, err := validateMappingRules(ctx, relationType.Type, relationType.MappingRules, strictMode)
	if err != nil {
		return err
	}
	relationType.MappingRules = rules

	return nil
}

// 校验mapping_rules的有效性
func validateMappingRules(ctx context.Context, relationType string, mappingRules any, strictMode bool) (any, error) {
	switch relationType {
	case interfaces.RELATION_TYPE_DIRECT:
		return validateDirectMappingRules(ctx, mappingRules, strictMode)
	case interfaces.RELATION_TYPE_DATA_VIEW:
		return validateInDirectMappingRules(ctx, mappingRules, strictMode)
	case interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN:
		return validateFilteredCrossJoinMappingRules(ctx, mappingRules, strictMode)
	default:
		// 如果type不是direct或data_view，返回错误
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("关系类的关系类型 %s 不支持", relationType))
	}
}

// 校验直接关联的mapping_rules
func validateDirectMappingRules(ctx context.Context, mappingRules any, strictMode bool) ([]interfaces.Mapping, error) {
	// mappingRules 先转成 []any 再解码成 []interfaces.Mapping
	var mappings []interfaces.Mapping
	if err := mapstructure.Decode(mappingRules, &mappings); err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("直接关联的 mapping_rules 解码失败: " + err.Error())
	}

	// 数组非空
	if len(mappings) == 0 {
		if strictMode {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("直接关联的 mapping_rules 不能为空")
		}
	}

	// mappings 里的每对映射规则的起点属性名和终点属性名都不能重复出现
	mappingsRuleMap := map[string]bool{}
	for idx, item := range mappings {
		// 校验起点属性非空
		if item.SourceProp.Name == "" {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("直接关联的 mapping_rules[%d] 的起点属性名不能为空", idx))
		}

		// 校验终点属性非空
		if item.TargetProp.Name == "" {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("直接关联的 mapping_rules[%d] 的终点属性名不能为空", idx))
		}

		// 映射规则重复出现，则报错
		if mappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("直接关联的 mapping_rules[%d] 的起点和终点的映射规则不能重复出现", idx))
		}
		mappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] = true
	}

	return mappings, nil
}

// 校验间接关联的mapping_rules
func validateInDirectMappingRules(ctx context.Context, mappingRules any, strictMode bool) (*interfaces.InDirectMapping, error) {
	// 尝试类型断言
	var mapping interfaces.InDirectMapping
	if err := mapstructure.Decode(mappingRules, &mapping); err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("间接关联的 mapping_rules 格式不正确，应为 InDirectMapping 对象")
	}

	// 校验关联的数据来源类型非空，且为 data_view 或 resource
	if mapping.BackingDataSource == nil {
		if strictMode {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("间接关联的 backing_data_source 不能为空")
		}
	} else {
		if mapping.BackingDataSource.Type == "" {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("间接关联的 backing_data_source.type 不能为空")
		}
		if mapping.BackingDataSource.Type != interfaces.DATA_SOURCE_TYPE_DATA_VIEW &&
			mapping.BackingDataSource.Type != interfaces.DATA_SOURCE_TYPE_RESOURCE {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("间接关联的 backing_data_source.type 必须为 %s 或 %s，当前为: %s",
					interfaces.DATA_SOURCE_TYPE_DATA_VIEW, interfaces.DATA_SOURCE_TYPE_RESOURCE, mapping.BackingDataSource.Type))
		}
		// 校验关联的数据视图id非空（数据视图存在性校验在logics层）
		if mapping.BackingDataSource.ID == "" {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("间接关联的 backing_data_source.id 不能为空")
		}
	}

	// 校验起点对象类与数据集的关联规则非空
	if len(mapping.SourceMappingRules) == 0 {
		if strictMode {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("间接关联的 source_mapping_rules 不能为空")
		}
	} else {
		// 起点到中间表的每对映射规则不能重复
		sourceMappingsRuleMap := map[string]bool{}
		for idx, item := range mapping.SourceMappingRules {
			// 校验起点对象类的属性非空（属性存在于起点对象类中的校验在logics层）
			if item.SourceProp.Name == "" {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 source_mapping_rules[%d] 的起点对象类属性名不能为空", idx))
			}
			// 校验中间的桥梁字段非空（桥梁字段存在于数据视图中的校验在logics层）
			if item.TargetProp.Name == "" {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 source_mapping_rules[%d] 的桥梁字段名不能为空", idx))
			}

			// 映射规则重复出现，则报错
			if sourceMappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 source_mapping_rules[%d] 的起点和终点的映射规则不能重复出现", idx))
			}
			sourceMappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] = true
		}
	}

	// 校验数据集与终点对象类的关联规则非空
	if len(mapping.TargetMappingRules) == 0 {
		if strictMode {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails("间接关联的 target_mapping_rules 不能为空")
		}
	} else {
		// 中间表到终点的每对映射规则不能重复
		targetMappingsRuleMap := map[string]bool{}
		for idx, item := range mapping.TargetMappingRules {
			// 校验中间的桥梁字段非空（桥梁字段存在于数据视图中的校验在logics层）
			if item.SourceProp.Name == "" {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 target_mapping_rules[%d] 的桥梁字段名不能为空", idx))
			}
			// 校验终点对象类的属性非空（属性存在于终点对象类中的校验在logics层）
			if item.TargetProp.Name == "" {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 target_mapping_rules[%d] 的终点对象类属性名不能为空", idx))
			}

			// 映射规则重复出现，则报错
			if targetMappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] {
				return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("间接关联的 target_mapping_rules[%d] 的起点和终点的映射规则不能重复出现", idx))
			}
			targetMappingsRuleMap[item.SourceProp.Name+":"+item.TargetProp.Name] = true
		}
	}

	return &mapping, nil
}

func validateFilteredCrossJoinMappingRules(ctx context.Context, mappingRules any, _ bool) (*interfaces.FilteredCrossJoinMapping, error) {
	var mapping interfaces.FilteredCrossJoinMapping
	if err := mapstructure.Decode(mappingRules, &mapping); err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("分侧过滤全连接 mapping_rules 解码失败: " + err.Error())
	}

	// 校验 source_condition
	if mapping.SourceCondition != nil {
		if err := validateRelationTypeCondition(ctx, mapping.SourceCondition, "source_condition"); err != nil {
			return nil, err
		}
	}

	// 校验 target_condition
	if mapping.TargetCondition != nil {
		if err := validateRelationTypeCondition(ctx, mapping.TargetCondition, "target_condition"); err != nil {
			return nil, err
		}
	}

	return &mapping, nil
}

// validateRelationTypeCondition 校验关系类过滤条件的合法性
func validateRelationTypeCondition(ctx context.Context, cfg *cond.CondCfg, prefix string) error {
	if cfg == nil {
		return nil
	}

	// 校验 operation 不能为空
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的 operation 不能为空", prefix))
	}

	// 校验 operation 是否有效
	if _, ok := cond.OperationMap[cfg.Operation]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的 operation[%s] 无效", prefix, cfg.Operation))
	}

	// and/or 操作需要校验子条件
	if cfg.Operation == cond.OperationAnd || cfg.Operation == cond.OperationOr {
		if len(cfg.SubConds) == 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的子条件(sub_conditions)不能为空", prefix))
		}

		// 校验子条件数量限制
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的子条件数量[%d]超过最大限制[%d]", prefix, len(cfg.SubConds), cond.MaxSubCondition))
		}

		// 递归校验每个子条件
		for i, subCond := range cfg.SubConds {
			subPrefix := fmt.Sprintf("%s.sub_conditions[%d]", prefix, i)
			if err := validateRelationTypeCondition(ctx, subCond, subPrefix); err != nil {
				return err
			}
		}
		return nil
	}

	// 非 and/or 操作需要校验 field
	if cfg.Field == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的 field 不能为空", prefix))
	}

	// 校验 value：非特殊操作需要 value
	if _, ok := cond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
		if cfg.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("分侧过滤全连接 %s 的值(value)不能为空", prefix))
		}
	}

	return nil
}
