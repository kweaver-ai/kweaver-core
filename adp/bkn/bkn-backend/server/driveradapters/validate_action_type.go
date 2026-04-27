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

	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

func ValidateActionTypes(ctx context.Context, knID string, actionTypes []*interfaces.ActionType, strictMode bool) error {
	tmpNameMap := make(map[string]any)
	idMap := make(map[string]any)
	for i := 0; i < len(actionTypes); i++ {
		// 校验导入模型时模块是否是行动类
		if actionTypes[i].ModuleType != "" && actionTypes[i].ModuleType != interfaces.MODULE_TYPE_ACTION_TYPE {
			return rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
				WithErrorDetails("Action type name is not 'action_type'")
		}

		// 0.校验请求体中多个模型 ID 是否重复
		atID := actionTypes[i].ATID
		if _, ok := idMap[atID]; !ok || atID == "" {
			idMap[atID] = nil
		} else {
			errDetails := fmt.Sprintf("ActionType ID '%s' already exists in the request body", atID)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_Duplicated_IDInFile).
				WithDescription(map[string]any{"actionTypeID": atID}).
				WithErrorDetails(errDetails)
		}

		// 1. 校验 行动类必要创建参数的合法性, 非空、长度、是枚举值
		err := ValidateActionType(ctx, actionTypes[i], strictMode)
		if err != nil {
			return err
		}

		// 2. 校验 请求体中行动类名称重复性
		if _, ok := tmpNameMap[actionTypes[i].ATName]; !ok {
			tmpNameMap[actionTypes[i].ATName] = nil
		} else {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_Duplicated_Name)
		}

		actionTypes[i].KNID = knID
	}
	return nil
}

// 对象类必要创建参数的非空校验。
func ValidateActionType(ctx context.Context, actionType *interfaces.ActionType, strictMode bool) error {
	// 校验id的合法性
	err := validateID(ctx, actionType.ATID)
	if err != nil {
		return err
	}

	// 校验名称合法性
	// 去掉名称的前后空格
	actionType.ATName = strings.TrimSpace(actionType.ATName)
	err = validateObjectName(ctx, actionType.ATName, interfaces.MODULE_TYPE_ACTION_TYPE)
	if err != nil {
		return err
	}

	// 若输入了 tags，校验 tags 的合法性
	err = ValidateTags(ctx, actionType.Tags)
	if err != nil {
		return err
	}

	// 去掉tag前后空格以及数组去重
	actionType.Tags = libCommon.TagSliceTransform(actionType.Tags)

	// 校验行动类型为有效类型
	if !interfaces.ActionTypeMap[actionType.ActionType] {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("The action type is expected one of [add, modify, delete], actual is [%s]",
				actionType.ActionType))
	}

	// 根据是否绑定对象类，校验行动条件和参数
	if actionType.ObjectTypeID == "" && strictMode {
		// 未绑定对象类时，行动条件必须为空
		if actionType.Condition != nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("未绑定对象类时，行动条件必须为空")
		}

		// 未绑定对象类时，参数 ValueFrom 不能是 property，只能是 const 或 input
		for _, param := range actionType.Parameters {
			if param.ValueFrom == interfaces.VALUE_FROM_PROPERTY {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
					WithErrorDetails("未绑定对象类时，行动资源参数不支持从数据属性获取值")
			}
		}
	}

	// 校验类型
	if actionType.ActionSource.Type != "" {
		// type 不为空，则代表在配置映射了，则需要校验映射
		if !interfaces.IsValidActionSourceType(actionType.ActionSource.Type) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("The type of action source is expected one of [tool, mcp], actual is [%s]",
					actionType.ActionSource.Type))
		}
		// strict_mode off: allow empty or draft combinations for McpID, ToolName, BoxID, ToolID (no cross-kind checks).
		if strictMode {
			switch actionType.ActionSource.Type {
			case interfaces.ACTION_SOURCE_TYPE_TOOL:
				// tool 时，mcp_id或者tool_name不为空，则报错
				if actionType.ActionSource.McpID != "" || actionType.ActionSource.ToolName != "" {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("tool type should not have mcp data, current mcp_id is[%s], tool_name is [%s]",
							actionType.ActionSource.McpID, actionType.ActionSource.ToolName))
				}
			case interfaces.ACTION_SOURCE_TYPE_MCP:
				// map 时，box_id或者tool_id不为空，则报错
				if actionType.ActionSource.BoxID != "" || actionType.ActionSource.ToolID != "" {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
						WithErrorDetails(fmt.Sprintf("mcp type should not have tool data, current box_id is[%s], tool_id is [%s]",
							actionType.ActionSource.BoxID, actionType.ActionSource.ToolID))
				}
			}
		}
	}

	// parameters 非空时：参数名称非空
	if len(actionType.Parameters) > 0 {
		for _, param := range actionType.Parameters {
			if param.Name == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("行动类[%s]行动资源参数名称不能为空", actionType.ATName))
			}
		}
	}

	// 行动条件非空时，校验行动条件（strict_mode 关闭时不校验）
	if actionType.Condition != nil && strictMode {
		err = validateActionCondition(ctx, actionType.Condition, actionType.ObjectTypeID)
		if err != nil {
			return err
		}
	}

	return nil
}

// 校验行动条件的合法性
func validateActionCondition(ctx context.Context, cfg *interfaces.ActionCondCfg, objectTypeID string) error {
	if cfg == nil {
		return nil
	}

	// 如果行动条件不给对象类id，那么就默认使用行动类的对象类id
	if cfg.ObjectTypeID == "" {
		cfg.ObjectTypeID = objectTypeID
	}
	// if cfg.ObjectTypeID == "" {
	// 	return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
	// 		WithErrorDetails("行动条件的对象类不能为空")
	// }

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
			WithErrorDetails("行动条件的过滤条件不能为空")
	}

	_, exists := interfaces.ActionCondOperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("行动条件的操作符[%s]不支持", cfg.Operation))
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		// 子过滤条件不能超过100个
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_CountExceeded_Conditions).
				WithErrorDetails(fmt.Sprintf("行动条件的子条件不能超过 %d 个", cond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateActionCondition(ctx, subCond, objectTypeID)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Field == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("行动条件的过滤字段不能为空")

		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte:
		// 右侧值为单个值
		_, ok := cfg.Value.([]any)
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]any)) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[range, out_range, before, between] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[range, out_range, before, between] operation's value must contain 2 values")
		}
	case cond.OperationExist, cond.OperationNotExist, cond.OperationNull, cond.OperationNotNull:
		// exist, not_exist, null, not_null 不需要值
		// 这些操作符已在 NotRequiredValueOperationMap 中定义，不需要额外验证

	case cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex:
		// like, not_like, prefix, not_prefix, regex 的值应该是单个字符串值
		_, ok := cfg.Value.([]any)
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single string value", cfg.Operation))
		}
		_, ok = cfg.Value.(string)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a string", cfg.Operation))
		}

	case cond.OperationContain, cond.OperationNotContain:
		// contain, not_contain 的值可以是单个值或数组
		// 如果是数组，长度应该大于等于1
		if cfg.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value cannot be nil", cfg.Operation))
		}
		if arr, ok := cfg.Value.([]any); ok {
			if len(arr) <= 0 {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("[%s] operation's value array should contains at least 1 value", cfg.Operation))
			}
		}

	case cond.OperationCurrent:
		// current 的值应该是字符串（unit），不能是数组
		_, ok := cfg.Value.([]any)
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[current] operation's value should be a string, not an array")
		}
		unit, ok := cfg.Value.(string)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails("[current] operation's value should be a string")
		}
		// 验证 unit 值
		validUnits := map[string]bool{
			"year":   true,
			"month":  true,
			"week":   true,
			"day":    true,
			"hour":   true,
			"minute": true,
		}
		if !validUnits[unit] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ActionType_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("[current] operation's unit value should be one of [year, month, week, day, hour, minute], actual is [%s]", unit))
		}
	}

	return nil
}
