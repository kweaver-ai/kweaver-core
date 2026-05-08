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

	"github.com/dlclark/regexp2"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	cond "bkn-backend/common/condition"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

var validMetricTypesEnum = map[string]struct{}{
	interfaces.MetricTypeAtomic:    {},
	interfaces.MetricTypeDerived:   {},
	interfaces.MetricTypeComposite: {},
}

var validMetricTimeRangePolicies = map[string]struct{}{
	interfaces.MetricTimeDefaultRangePolicyLast1h:      {},
	interfaces.MetricTimeDefaultRangePolicyLast24h:     {},
	interfaces.MetricTimeDefaultRangePolicyCalendarDay: {},
	interfaces.MetricTimeDefaultRangePolicyNone:        {},
}

// ValidateMetricRequests 校验批量创建指标请求体：id/名称/tag、单位、统计主体、公式、时间维度、分析维度等（不写库、不查依赖资源）。
func ValidateMetricRequests(ctx context.Context, entries []*interfaces.MetricDefinition, strictMode bool, importMode string) error {
	_ = importMode
	if len(entries) == 0 {
		return nil
	}
	seenName := make(map[string]struct{}, len(entries))
	seenID := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		name := strings.TrimSpace(e.Name)
		if name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("metric name is required")
		}
		if _, dup := seenName[name]; dup {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("duplicate metric name in request body: %s", name))
		}
		seenName[name] = struct{}{}

		id := strings.TrimSpace(e.ID)
		if id != "" {
			if _, dup := seenID[id]; dup {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
					WithErrorDetails(fmt.Sprintf("duplicate metric id in request body: %s", id))
			}
			seenID[id] = struct{}{}
		}

		if err := ValidateMetricRequest(ctx, e, strictMode); err != nil {
			return err
		}
	}
	return nil
}

// ValidateMetricRequest 校验单条创建指标请求体（与 ValidateObjectType 对单条对象类的作用类似）。
func ValidateMetricRequest(ctx context.Context, metric *interfaces.MetricDefinition, strictMode bool) error {
	if err := validateID(ctx, strings.TrimSpace(metric.ID)); err != nil {
		return err
	}
	if err := validateObjectName(ctx, strings.TrimSpace(metric.Name), interfaces.MODULE_TYPE_METRIC); err != nil {
		return err
	}
	if err := ValidateTags(ctx, metric.Tags); err != nil {
		return err
	}
	if err := validateMetricType(ctx, strings.TrimSpace(metric.MetricType), strictMode); err != nil {
		return err
	}
	if err := validateMetricUnits(ctx, strings.TrimSpace(metric.UnitType), strings.TrimSpace(metric.Unit), strictMode); err != nil {
		return err
	}
	if err := validateMetricScopeBody(ctx, strings.TrimSpace(metric.ScopeType), strings.TrimSpace(metric.ScopeRef), strictMode); err != nil {
		return err
	}
	if err := validateMetricTimeDimensionBody(ctx, metric.TimeDimension, strictMode); err != nil {
		return err
	}
	if err := validateMetricCalculationFormulaBody(ctx, metric.CalculationFormula, strictMode); err != nil {
		return err
	}
	if err := validateMetricAnalysisDimensionsBody(ctx, metric.AnalysisDimensions, strictMode); err != nil {
		return err
	}
	return nil
}

// ValidateUpdateMetricRequest 校验 PUT 更新指标请求体：允许等价于 "{}" 的空 body（名称与各字段均可省略，由路由覆盖 id/kn_id/branch）。
func ValidateUpdateMetricRequest(ctx context.Context, metric *interfaces.MetricDefinition, strictMode bool) error {
	if err := validateID(ctx, strings.TrimSpace(metric.ID)); err != nil {
		return err
	}
	if name := strings.TrimSpace(metric.Name); name != "" {
		if err := validateObjectName(ctx, name, interfaces.MODULE_TYPE_METRIC); err != nil {
			return err
		}
	}
	if err := ValidateTags(ctx, metric.Tags); err != nil {
		return err
	}
	if err := validateMetricTypeUpdate(ctx, strings.TrimSpace(metric.MetricType), strictMode); err != nil {
		return err
	}
	if err := validateMetricUnitsUpdate(ctx, strings.TrimSpace(metric.UnitType), strings.TrimSpace(metric.Unit), strictMode); err != nil {
		return err
	}
	if err := validateMetricScopeBodyUpdate(ctx, strings.TrimSpace(metric.ScopeType), strings.TrimSpace(metric.ScopeRef), strictMode); err != nil {
		return err
	}
	if err := validateMetricTimeDimensionBody(ctx, metric.TimeDimension, strictMode); err != nil {
		return err
	}
	if err := validateMetricCalculationFormulaBodyUpdate(ctx, metric.CalculationFormula, strictMode); err != nil {
		return err
	}
	if err := validateMetricAnalysisDimensionsBody(ctx, metric.AnalysisDimensions, strictMode); err != nil {
		return err
	}
	return nil
}

func validateMetricTypeUpdate(ctx context.Context, metricType string, strictMode bool) error {
	if metricType == "" {
		return nil
	}
	if strictMode {
		if metricType != interfaces.MetricTypeAtomic {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidMetricType)
		}
		return nil
	}
	if _, ok := validMetricTypesEnum[metricType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("metric_type must be one of atomic, derived, composite")
	}
	return nil
}

func validateMetricUnitsUpdate(ctx context.Context, unitType string, unit string, strictMode bool) error {
	if unitType == "" && unit == "" {
		return nil
	}
	if strictMode {
		if unitType == "" || unit == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("unit_type and unit are required in strict mode when either is provided")
		}
	}
	if unitType != "" {
		if _, ok := interfaces.ValidMetricUnitTypes[unitType]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit_type %q, expected one of %v", unitType, interfaces.ValidMetricUnitTypesArr))
		}
	}
	if unit != "" {
		if _, ok := interfaces.ValidMetricUnits[unit]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit %q, expected one of %v", unit, interfaces.ValidMetricUnitsArr))
		}
	}
	return nil
}

func validateMetricScopeBodyUpdate(ctx context.Context, scopeType, scopeRef string, strictMode bool) error {
	if scopeType == "" && scopeRef == "" {
		return nil
	}
	if strictMode {
		if scopeType == "" || scopeRef == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("scope_type and scope_ref are required in strict mode when scope is provided")
		}
		if scopeType != interfaces.ScopeTypeObjectType {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("only scope_type object_type is supported for metrics")
		}
		return nil
	}
	if scopeType != interfaces.ScopeTypeObjectType {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("only scope_type object_type is supported for metrics")
	}
	if scopeRef == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("scope_ref is required when scope is provided")
	}
	return nil
}

func validateMetricCalculationFormulaBodyUpdate(ctx context.Context, f *interfaces.MetricCalculationFormula, strictMode bool) error {
	if f == nil || metricCalculationFormulaEmpty(f) {
		return nil
	}
	return validateMetricCalculationFormulaBody(ctx, f, strictMode)
}

func validateMetricType(ctx context.Context, metricType string, strictMode bool) error {
	if strictMode {
		if metricType == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("metric_type is required in strict mode")
		}
		if metricType != interfaces.MetricTypeAtomic {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidMetricType)
		}
		return nil
	}
	// strictMode 为 false 时，metricType 为空可跳过，便于先导入后补类型。
	if metricType == "" {
		return nil
	}
	// strict mode 为false，非空时，需要是一个有效值
	if _, ok := validMetricTypesEnum[metricType]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("metric_type must be one of atomic, derived, composite")
	}
	return nil
}

func validateMetricUnits(ctx context.Context, unitType string, unit string, strictMode bool) error {
	if strictMode {
		if unitType == "" || unit == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("unit_type and unit are required in strict mode")
		}
	}
	if unitType != "" {
		if _, ok := interfaces.ValidMetricUnitTypes[unitType]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit_type %q, expected one of %v", unitType, interfaces.ValidMetricUnitTypesArr))
		}
	}
	if unit != "" {
		if _, ok := interfaces.ValidMetricUnits[unit]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit %q, expected one of %v", unit, interfaces.ValidMetricUnitsArr))
		}
	}
	return nil
}

// 统计主体：scope_type + scope_ref；非 strict 时允许皆空（导入占位）；非空时须一致且 scope_ref 符合 ID 规则。
func validateMetricScopeBody(ctx context.Context, scopeType, scopeRef string, strictMode bool) error {
	if strictMode {
		if scopeType == "" || scopeRef == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("scope_type and scope_ref are required in strict mode")
		}
		if scopeType != interfaces.ScopeTypeObjectType {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("only scope_type object_type is supported for metrics")
		}
		return nil
	}
	if scopeType == "" && scopeRef == "" {
		return nil
	}
	if scopeType != interfaces.ScopeTypeObjectType {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("only scope_type object_type is supported for metrics")
	}
	if scopeRef == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("scope_ref is required when scope is provided")
	}
	return nil
}

func timeDimensionPresent(td *interfaces.MetricTimeDimension) bool {
	if td == nil {
		return false
	}
	return strings.TrimSpace(td.Property) != "" || strings.TrimSpace(td.DefaultRangePolicy) != ""
}

func validateMetricTimeDimensionBody(ctx context.Context, td *interfaces.MetricTimeDimension, _ bool) error {
	// time dimension 为空或不完整时，可跳过，不是必填
	if td == nil || !timeDimensionPresent(td) {
		return nil
	}
	// time dimension 不为空时，需有效
	prop := strings.TrimSpace(td.Property)
	if prop == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("time_dimension.property is required when time_dimension is provided")
	}
	pol := strings.TrimSpace(td.DefaultRangePolicy)
	if pol != "" {
		if _, ok := validMetricTimeRangePolicies[pol]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("invalid time_dimension.default_range_policy")
		}
	}
	return nil
}

func validateMetricAnalysisDimensionsBody(ctx context.Context, ads []interfaces.MetricAnalysisDimension, strictMode bool) error {
	_ = strictMode
	if len(ads) == 0 {
		return nil
	}
	for i := range ads {
		n := strings.TrimSpace(ads[i].Name)
		if n == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("analysis_dimensions[%d].name is required", i))
		}
		if ads[i].DisplayName != "" {
			if err := validateObjectName(ctx, ads[i].DisplayName, interfaces.MODULE_TYPE_METRIC); err != nil {
				return err
			}
		}
	}
	return nil
}

func metricCalculationFormulaEmpty(f *interfaces.MetricCalculationFormula) bool {
	if f == nil {
		return true
	}
	if f.Condition != nil {
		return false
	}
	if strings.TrimSpace(f.Aggregation.Property) != "" || strings.TrimSpace(f.Aggregation.Aggr) != "" {
		return false
	}
	if len(f.GroupBy) > 0 || len(f.OrderBy) > 0 || f.Having != nil {
		return false
	}
	return true
}

func validateMetricCalculationFormulaBody(ctx context.Context, f *interfaces.MetricCalculationFormula, strictMode bool) error {
	if f == nil {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("calculation_formula is required in strict mode")
		}
		return nil
	}
	if strictMode && metricCalculationFormulaEmpty(f) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula is required in strict mode")
	}
	if !strictMode && metricCalculationFormulaEmpty(f) {
		return nil
	}
	if f.Condition != nil {
		if err := validateMetricCond(ctx, f.Condition); err != nil {
			return err
		}
	}
	if err := validateMetricAggregation(ctx, &f.Aggregation, strictMode); err != nil {
		return err
	}
	for i := range f.GroupBy {
		p := strings.TrimSpace(f.GroupBy[i].Property)
		if p == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("calculation_formula.group_by[%d].property is required", i))
		}
		if err := ValidatePropertyName(ctx, p); err != nil {
			return err
		}
	}
	for i := range f.OrderBy {
		p := strings.TrimSpace(f.OrderBy[i].Property)
		if p == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("calculation_formula.order_by[%d].property is required", i))
		}
		if err := ValidatePropertyName(ctx, p); err != nil {
			return err
		}
		d := strings.TrimSpace(f.OrderBy[i].Direction)
		if d != "" && d != interfaces.MetricOrderDirectionAsc && d != interfaces.MetricOrderDirectionDesc {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("calculation_formula.order_by direction must be asc or desc")
		}
	}
	if f.Having != nil {
		if err := validateMetricHaving(ctx, f.Having); err != nil {
			return err
		}
	}
	return nil
}

func validateMetricAggregation(ctx context.Context, a *interfaces.MetricAggregation, strictMode bool) error {
	if a == nil {
		if strictMode {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
				WithErrorDetails("calculation_formula.aggregation is required in strict mode")
		}
		return nil
	}
	if strings.TrimSpace(a.Property) == "" || strings.TrimSpace(a.Aggr) == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.aggregation.property and calculation_formula.aggregation.aggr are required")
	}
	ag := strings.TrimSpace(a.Aggr)
	if _, ok := interfaces.ValidMetricAggrs[ag]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("invalid calculation_formula.aggregation.aggr %s ", ag))
	}
	return nil
}

func validateMetricHaving(ctx context.Context, h *interfaces.MetricHaving) error {
	// 如果 having 的field为空，默认用 __value，不是必填
	if strings.TrimSpace(h.Field) == "" {
		h.Field = interfaces.MetricHavingFieldValue
	}
	// 如果 having 的operation为空，返回错误
	if strings.TrimSpace(h.Operation) == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.having operation is required")
	}

	// 如果 having 的value为空，返回错误
	if h.Value == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.having value is required")
	}

	// value 需是数值类型
	switch v := h.Value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("calculation_formula.having value must be a number, got %T", v))
	}
}

func validateMetricCond(ctx context.Context, cfg *cond.CondCfg) error {
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

	// 指标的过滤条件不支持模糊查询和语义查询操作符
	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
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
		if cfg.Field == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_NullParameter_ConditionName)
		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte,
		cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex,
		cond.OperationCurrent:
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
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_UnsupportConditionOperation).
			WithErrorDetails(fmt.Sprintf("[%s] operation is not supported in metric condition", cfg.Operation))
	}
	return nil
}
