// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

	verrors "vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/filter_condition"
)

// 资源数据查询参数校验
func ValidateResourceDataQueryParams(ctx context.Context, params *interfaces.ResourceDataQueryParams) error {
	// 校验format是否为 original 或者 flat
	if params.Format == "" {
		params.Format = interfaces.Format_Original
	} else {
		err := validateFormat(ctx, params.Format)
		if err != nil {
			return err
		}
	}

	// limit 默认值为 10
	if params.Limit == 0 {
		params.Limit = interfaces.DEFAULT_DATA_LIMIT
	}

	// 校验分页参数
	err := validatePaginationParams(ctx, params.Offset, params.Limit)
	if err != nil {
		return err
	}

	// 校验排序参数
	err = validateSortFields(ctx, params.Sort)
	if err != nil {
		return err
	}

	// 聚合模式下的参数校验：当Aggregation、GroupBy或Having任一参数存在时执行
	if isAggregateQuery(params) {
		err = validateAggregateParams(ctx, params)
		if err != nil {
			return err
		}
	}

	// 过滤条件用map接，然后再decode到condCfg中
	var actualCond *interfaces.FilterCondCfg
	err = mapstructure.Decode(params.FilterCondition, &actualCond)
	if err != nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterCondition).
			WithErrorDetails(fmt.Sprintf("mapstructure decode filters failed: %s", err.Error()))
	}
	params.FilterCondCfg = actualCond

	// 校验全局过滤条件：操作符、字段类型和操作符是否匹配
	err = validateFilterCondCfg(ctx, params.FilterCondCfg)
	if err != nil {
		return err
	}

	return nil
}

func validateFormat(ctx context.Context, format string) error {
	if format != interfaces.Format_Original && format != interfaces.Format_Flat {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Format).
			WithErrorDetails(fmt.Sprintf("The output format should be %s or %s", interfaces.Format_Original, interfaces.Format_Flat))
	}

	return nil
}

// 分页排序参数校验
func validatePaginationParams(ctx context.Context, offset, limit int) error {
	// from + size 查询校验
	if offset < 0 {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Offset).
			WithErrorDetails("When execute From + size query, 'offset' should be >= 0")
	}

	if limit < interfaces.MIN_LIMIT || limit > interfaces.MAX_SEARCH_SIZE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("Limit should be in the range of [%d,%d]", interfaces.MIN_LIMIT, interfaces.MAX_SEARCH_SIZE))
	}

	if offset+limit > interfaces.MAX_SEARCH_SIZE {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Limit).
			WithErrorDetails(fmt.Sprintf("Offset + limit should be <= %d", interfaces.MAX_SEARCH_SIZE))
	}

	return nil
}

func validateSortFields(ctx context.Context, sortFields []*interfaces.SortField) error {
	for _, sortField := range sortFields {
		if sortField.Direction != interfaces.ASC_DIRECTION &&
			sortField.Direction != interfaces.DESC_DIRECTION {

			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Direction).
				WithErrorDetails("The sort direction should be desc or asc")
		}
	}

	return nil
}

func validateFilterCondCfg(ctx context.Context, cfg *interfaces.FilterCondCfg) error {
	if cfg == nil {
		return nil
	}

	// 判断过滤器是否为空对象 {}
	if cfg.Name == "" && cfg.Operation == "" && len(cfg.SubConds) == 0 && cfg.ValueFrom == "" && cfg.Value == nil {
		return nil
	}

	// 过滤操作符
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionOperation)
	}

	condFactory, exists := filter_condition.OperationMap[cfg.Operation]
	if !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_UnsupportFilterConditionOperation).
			WithErrorDetails(fmt.Sprintf("Unsupported filter condition operation: %s", cfg.Operation))
	}

	if !condFactory.SupportSubCond() {
		if len(cfg.SubConds) > 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_UnsupportFilterConditionOperation).
				WithErrorDetails(fmt.Sprintf("operation '%s' does not support sub conditions", cfg.Operation))
		}
	} else {
		if len(cfg.SubConds) > interfaces.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_CountExceeded_FilterConditionSubConds).
				WithErrorDetails(fmt.Sprintf("The number of subConditions exceeds %d", interfaces.MaxSubCondition))
		}

		for _, subCond := range cfg.SubConds {
			err := validateFilterCondCfg(ctx, subCond)
			if err != nil {
				return err
			}
		}
	}

	if condFactory.NeedName() {
		// 过滤字段名称不能为空
		if cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionName)
		}
	}

	if condFactory.NeedValue() {
		if cfg.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_NullParameter_FilterConditionValue)
		}

		if cfg.ValueFrom == "" {
			cfg.ValueFrom = interfaces.ValueFrom_Const
		}
		if condFactory.NeedConstValue() {
			// 过滤字段值不能为空
			if cfg.ValueFrom != interfaces.ValueFrom_Const {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValueFrom)
			}
		}

		if condFactory.IsSingleValue() {
			// 右侧值为单个值
			if _, ok := cfg.Value.([]any); ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
			}
		} else if condFactory.IsFixedLenArrayValue() {
			// 右侧值为数组值
			if vals, ok := cfg.Value.([]any); !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
					WithErrorDetails(fmt.Sprintf("[%s] operation's value must be an array", cfg.Operation))
			} else {
				if condFactory.IsFixedLenArrayValue() && len(vals) != condFactory.RequiredValueLen() {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
						WithErrorDetails(fmt.Sprintf("[%s] operation's value must contain %d values", cfg.Operation, condFactory.RequiredValueLen()))
				} else if len(vals) == 0 {
					return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_FilterConditionValue).
						WithErrorDetails(fmt.Sprintf("[%s] operation's value should contains at least 1 value", cfg.Operation))
				}
			}
		}
	}

	return nil
}

// isAggregateQuery 判断是否为聚合查询
func isAggregateQuery(params *interfaces.ResourceDataQueryParams) bool {
	// 根据聚合相关字段推断
	return params.Aggregation != nil || len(params.GroupBy) > 0 || params.Having != nil
}

// validateCalendarInterval 校验 calendar_interval 是否为有效的枚举值
// 允许的值包括：minute, hour, day, week, month, quarter, year
func validateCalendarInterval(ctx context.Context, calendarInterval string) error {
	switch calendarInterval {
	case interfaces.CALENDAR_UNIT_MINUTE,
		interfaces.CALENDAR_UNIT_HOUR,
		interfaces.CALENDAR_UNIT_DAY,
		interfaces.CALENDAR_UNIT_WEEK,
		interfaces.CALENDAR_UNIT_MONTH,
		interfaces.CALENDAR_UNIT_QUARTER,
		interfaces.CALENDAR_UNIT_YEAR:
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_CalendarInterval).
			WithErrorDetails(fmt.Sprintf("Invalid calendar_interval value: %s, must be one of: minute, hour, day, week, month, quarter, year", calendarInterval))
	}
}

// validateAggregateParams 校验聚合查询参数
func validateAggregateParams(ctx context.Context, params *interfaces.ResourceDataQueryParams) error {
	// 校验aggregation
	if params.Aggregation != nil {
		if params.Aggregation.Property == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Aggregation).
				WithErrorDetails("Aggregation property cannot be empty")
		}
		// 校验聚合函数类型
		validAggr := map[string]bool{
			"count": true, "count_distinct": true, "sum": true,
			"max": true, "min": true, "avg": true,
		}
		if !validAggr[params.Aggregation.Aggr] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Aggregation).
				WithErrorDetails(fmt.Sprintf("Unsupported aggregation function: %s", params.Aggregation.Aggr))
		}
	}

	// 校验group_by
	for _, groupByItem := range params.GroupBy {
		if groupByItem.Property == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_GroupBy).
				WithErrorDetails("GroupBy property cannot be empty")
		}
		// 校验calendar_interval
		if groupByItem.CalendarInterval != "" {
			err := validateCalendarInterval(ctx, groupByItem.CalendarInterval)
			if err != nil {
				return err
			}
		}
	}

	// 校验having
	if params.Having != nil {
		// having依赖aggregation或count(*)
		if params.Aggregation == nil && params.Having.Field != "count(*)" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Having).
				WithErrorDetails("Having clause requires aggregation or count(*)")
		}
		// 校验field字段
		if params.Having.Field != "__value" && params.Having.Field != "count(*)" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Having).
				WithErrorDetails("Having field must be '__value' or 'count(*)'")
		}
		// 校验operation
		validOps := map[string]bool{
			"==": true, "!=": true, ">": true, ">=": true,
			"<": true, "<=": true, "in": true, "not_in": true,
			"range": true, "out_range": true,
		}
		if !validOps[params.Having.Operation] {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Having).
				WithErrorDetails(fmt.Sprintf("Unsupported having operation: %s", params.Having.Operation))
		}
		// 校验value
		if params.Having.Value == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, verrors.VegaBackend_InvalidParameter_Having).
				WithErrorDetails("Having value cannot be empty")
		}
	}

	return nil
}
