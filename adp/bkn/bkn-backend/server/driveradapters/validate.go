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
	"strings"
	"unicode/utf8"

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"github.com/mitchellh/mapstructure"

	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// 对象名称错误码字典, key为对象类型, value为其错误码数组
var objectNameErrorCode = map[string][]string{

	interfaces.MODULE_TYPE_KN: {
		berrors.BknBackend_KnowledgeNetwork_NullParameter_Name,
		berrors.BknBackend_KnowledgeNetwork_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_OBJECT_TYPE: {
		berrors.BknBackend_ObjectType_NullParameter_Name,
		berrors.BknBackend_ObjectType_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_RELATION_TYPE: {
		berrors.BknBackend_RelationType_NullParameter_Name,
		berrors.BknBackend_RelationType_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_ACTION_TYPE: {
		berrors.BknBackend_ActionType_NullParameter_Name,
		berrors.BknBackend_ActionType_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_JOB: {
		berrors.BknBackend_Job_NullParameter_Name,
		berrors.BknBackend_Job_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_CONCEPT_GROUP: {
		berrors.BknBackend_ConceptGroup_NullParameter_Name,
		berrors.BknBackend_ConceptGroup_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_RISK_TYPE: {
		berrors.BknBackend_RiskType_NullParameter_Name,
		berrors.BknBackend_RiskType_LengthExceeded_Name,
	},

	interfaces.MODULE_TYPE_METRIC: {
		berrors.BknBackend_Metric_NullParameter_Name,
		berrors.BknBackend_Metric_LengthExceeded_Name,
	},
}

// 校验的导入模式
func validateImportMode(ctx context.Context, mode string) *rest.HTTPError {
	switch mode {
	case interfaces.ImportMode_Normal,
		interfaces.ImportMode_Ignore,
		interfaces.ImportMode_Overwrite:
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_ImportMode).
			WithErrorDetails("The import_mode value can be 'overwrite', 'normal', 'ignore'")
	}

	return nil
}

// 公共校验函数(1): 对象名称合法性校验
func validateObjectName(ctx context.Context, objectName string, objectType string) error {
	if objectName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, objectNameErrorCode[objectType][0])
	}

	if utf8.RuneCountInString(objectName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, objectNameErrorCode[objectType][1]).
			WithErrorDetails(fmt.Sprintf("The length of the %v named %v exceeds %v", objectType, objectName, interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	return nil
}

// tags 的合法性校验
func ValidateTags(ctx context.Context, Tags []string) error {
	if len(Tags) > interfaces.TAGS_MAX_NUMBER {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_CountExceeded_TagTotal)
	}

	for _, tag := range Tags {
		err := validateDataTagName(ctx, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// 数据标签名称合法性校验
func validateDataTagName(ctx context.Context, dataTagName string) error {
	// 去除dataTagName的左右空格
	dataTagName = strings.Trim(dataTagName, " ")

	if dataTagName == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_DataTagName)
		// .WithErrorDetails("Data tag name is null")
	}

	if utf8.RuneCountInString(dataTagName) > interfaces.OBJECT_NAME_MAX_LENGTH {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_DataTagName).
			WithErrorDetails(fmt.Sprintf("The length of the data tag name exceeds %d", interfaces.OBJECT_NAME_MAX_LENGTH))
	}

	if isInvalid := strings.ContainsAny(interfaces.NAME_INVALID_CHARACTER, dataTagName); isInvalid {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_DataTagName).
			WithErrorDetails(fmt.Sprintf("Data tag name contains special characters, such as %s", interfaces.NAME_INVALID_CHARACTER))
	}

	return nil
}

// 分页参数合法性校验
func validatePaginationQueryParameters(ctx context.Context, offset, limit, sort, direction string,
	supportedSortTypes map[string]string) (interfaces.PaginationQueryParameters, error) {
	pageParams := interfaces.PaginationQueryParameters{}

	off, err := strconv.Atoi(offset)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Offset).
			WithErrorDetails(err.Error())
	}

	if off < interfaces.MIN_OFFSET {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Offset).
			WithErrorDetails(fmt.Sprintf("The offset is not greater than %d", interfaces.MIN_OFFSET))
	}

	lim, err := strconv.Atoi(limit)
	if err != nil {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Limit).
			WithErrorDetails(err.Error())
	}

	if !(limit == interfaces.NO_LIMIT || (lim >= interfaces.MIN_LIMIT && lim <= interfaces.MAX_LIMIT)) {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("The number per page does not equal %s is not in the range of [%d,%d]", interfaces.NO_LIMIT, interfaces.MIN_LIMIT, interfaces.MAX_LIMIT))
	}

	_, ok := supportedSortTypes[sort]
	if !ok {
		types := make([]string, 0)
		for t := range supportedSortTypes {
			types = append(types, t)
		}
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Sort).
			WithErrorDetails(fmt.Sprintf("Wrong sort type, does not belong to any item in set %v ", types))
	}

	if direction != interfaces.DESC_DIRECTION && direction != interfaces.ASC_DIRECTION {
		return pageParams, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Direction).
			WithErrorDetails("The sort direction is not desc or asc")
	}

	return interfaces.PaginationQueryParameters{
		Offset:    off,
		Limit:     lim,
		Sort:      supportedSortTypes[sort],
		Direction: direction,
	}, nil
}

func validateConceptsQuery(ctx context.Context, query *interfaces.ConceptsQuery) error {

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *cond.CondCfg
	err := mapstructure.Decode(query.Condition, &actualCond)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Condition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode condition failed: %s", err.Error()))
	}
	query.ActualCondition = actualCond

	knFilter := &cond.CondCfg{
		Field:     "kn_id",
		Operation: cond.OperationEq,
		ValueOptCfg: cond.ValueOptCfg{
			ValueFrom: cond.ValueFrom_Const,
			Value:     query.KNID,
		},
	}

	// 3. module type的过滤
	typeFilter := &cond.CondCfg{
		Field:     "module_type",
		Operation: cond.OperationEq,
		ValueOptCfg: cond.ValueOptCfg{
			ValueFrom: cond.ValueFrom_Const,
			Value:     query.ModuleType,
		},
	}

	// 4. branch的过滤
	branchFilter := &cond.CondCfg{
		Field:     "branch",
		Operation: cond.OperationEq,
		ValueOptCfg: cond.ValueOptCfg{
			ValueFrom: cond.ValueFrom_Const,
			Value:     query.Branch,
		},
	}

	// 如果包含了knn，则把kn_id、module_type、branch的过滤条件放在knn的sub condition里
	err = validateCond(ctx, query.ActualCondition)
	if err != nil {
		return err
	}

	query.ActualCondition = &cond.CondCfg{
		Operation: cond.OperationAnd,
		SubConds:  []*cond.CondCfg{query.ActualCondition, knFilter, typeFilter, branchFilter},
	}

	return nil
}

func validateCond(ctx context.Context, cfg *cond.CondCfg) error {
	if cfg == nil {
		return nil
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_NullParameter_ConditionOperation)
	}

	_, exists := cond.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_UnsupportConditionOperation)
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr, cond.OperationKNN:
		// 子过滤条件不能超过10个
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_CountExceeded_Conditions).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", cond.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateCond(ctx, subCond)
			if err != nil {
				return err
			}
		}
	default:
		// 过滤字段名称不能为空
		if cfg.Operation != cond.OperationMultiMatch && cfg.Field == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_NullParameter_ConditionName)
		}

		// if _, ok := cond.NotRequiredValueOperationMap[cfg.Operation]; !ok {
		// 	if cfg.ValueFrom != cond.ValueFrom_Const {
		// 		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.Uniquery_InvalidParameter_ValueFrom).
		// 			WithErrorDetails(fmt.Sprintf("condition does not support value_from type('%s')", cfg.ValueFrom))
		// 	}
		// }
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte,
		cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex,
		cond.OperationMatch, cond.OperationMatchPhrase, cond.OperationCurrent, cond.OperationMultiMatch:
		// 右侧值为单个值
		_, ok := cfg.Value.([]any)
		if ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}

		if cfg.Operation == cond.OperationLike || cfg.Operation == cond.OperationNotLike ||
			cfg.Operation == cond.OperationPrefix || cfg.Operation == cond.OperationNotPrefix {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}

		if cfg.Operation == cond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
					WithErrorDetails("[regex] operation's value should be a string")
			}

			_, err := regexp2.Compile(val, regexp2.RE2)
			if err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
					WithErrorDetails(fmt.Sprintf("[regex] operation regular expression error: %s", err.Error()))
			}

		}

	case cond.OperationIn, cond.OperationNotIn:
		// 当 operation 是 in, not_in 时，value 为任意基本类型的数组，且长度大于等于1；
		_, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}

		if len(cfg.Value.([]any)) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
				WithErrorDetails("[in not_in] operation's value should contains at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		// 当 operation 是 range 时，value 是个由范围的下边界和上边界组成的长度为 2 的数值型数组
		// 当 operation 是 out_range 时，value 是个长度为 2 的数值类型的数组，查询的数据范围为 (-inf, value[0]) || [value[1], +inf)
		v, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}

		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ConditionValue).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}
	}

	return nil
}

func validateID(ctx context.Context, id string) error {
	if id != "" {
		//  id，只包含小写英文字母和数字和下划线(_)和连字符(-)，且不能以下划线开头，不能超过40个字符
		re := regexp2.MustCompile(interfaces.RegexPattern_NonBuiltin_ID, regexp2.RE2)
		match, err := re.MatchString(id)
		if err != nil || !match {
			errDetails := fmt.Sprintf(`The id can contain only lowercase letters, digits and underscores(_),
			it cannot start with underscores and cannot exceed 40 characters, but got %s`, id)
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_ID).
				WithErrorDetails(errDetails)
		}
	}

	return nil
}

// 校验 x-http-method-override 重载方法，只在 header里传递
func ValidateHeaderMethodOverride(ctx context.Context, methodOverride string) error {

	if methodOverride == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_OverrideMethod).
			WithErrorDetails("x-http-method-override must be set")
	}

	if methodOverride != "GET" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_OverrideMethod).
			WithErrorDetails(fmt.Sprintf("x-http-method-overide is expected to be GET, but got %s", methodOverride))
	}
	return nil
}
