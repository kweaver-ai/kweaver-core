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

	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

var (
	validMetricTimeRangePolicies = map[string]struct{}{
		interfaces.MetricTimeDefaultRangePolicyLast1h:      {},
		interfaces.MetricTimeDefaultRangePolicyLast24h:     {},
		interfaces.MetricTimeDefaultRangePolicyCalendarDay: {},
		interfaces.MetricTimeDefaultRangePolicyNone:        {},
	}
	validMetricTimeRangePoliciesArr = []string{
		interfaces.MetricTimeDefaultRangePolicyLast1h,
		interfaces.MetricTimeDefaultRangePolicyLast24h,
		interfaces.MetricTimeDefaultRangePolicyCalendarDay,
		interfaces.MetricTimeDefaultRangePolicyNone,
	}
)

func validateMetricDefinitionExecutionScope(ctx context.Context, def *interfaces.MetricDefinition) error {
	if def == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric definition is required")
	}
	if def.ScopeType != interfaces.ScopeTypeObjectType {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_UnsupportedScope)
	}
	return nil
}

// validateMetricDefinitionLikeBknSave 与 bkn-backend 保存指标时 ValidateMetricRequest 一致，但跳过 id、name、tags、comment。
// strictMode 与 bkn 创建接口默认 strict_mode=true 对齐。
func validateMetricConfig(ctx context.Context, metric *interfaces.MetricDefinition) error {
	if metric == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric config is required when using metric dry run")
	}
	if err := validateMetricType(ctx, strings.TrimSpace(metric.MetricType)); err != nil {
		return err
	}
	if err := validateMetricUnits(ctx, strings.TrimSpace(metric.UnitType), strings.TrimSpace(metric.Unit)); err != nil {
		return err
	}
	if err := validateMetricScopeBody(ctx, strings.TrimSpace(metric.ScopeType), strings.TrimSpace(metric.ScopeRef)); err != nil {
		return err
	}
	if err := validateMetricTimeDimensionBody(ctx, metric.TimeDimension); err != nil {
		return err
	}
	if err := validateMetricCalculationFormulaBody(ctx, metric.CalculationFormula); err != nil {
		return err
	}
	return nil
}

func validateMetricType(ctx context.Context, metricType string) error {
	if metricType == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric_type is required when using metric dry run")
	}
	if metricType != interfaces.MetricTypeAtomic {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric_type must be atomic when using metric dry run")
	}
	return nil
}

func validateMetricUnits(ctx context.Context, unitType string, unit string) error {
	// 单位信息可以为空
	if unitType != "" {
		if _, ok := interfaces.ValidMetricUnitTypes[unitType]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit_type %q, expected one of %v", unitType, interfaces.ValidMetricUnitTypesArr))
		}
	}
	if unit != "" {
		if _, ok := interfaces.ValidMetricUnits[unit]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid unit %q, expected one of %v", unit, interfaces.ValidMetricUnitsArr))
		}
	}
	return nil
}

func validateMetricScopeBody(ctx context.Context, scopeType, scopeRef string) error {
	if scopeType == "" || scopeRef == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("scope_type and scope_ref are required when using metric dry run")
	}
	if scopeType != interfaces.ScopeTypeObjectType {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("only scope_type object_type is supported for metrics when using metric dry run")
	}
	return nil
}

func timeDimensionPresent(td *interfaces.MetricTimeDimension) bool {
	if td == nil {
		return false
	}
	return strings.TrimSpace(td.Property) != "" || strings.TrimSpace(td.DefaultRangePolicy) != ""
}

func validateMetricTimeDimensionBody(ctx context.Context, td *interfaces.MetricTimeDimension) error {
	if td == nil || !timeDimensionPresent(td) {
		return nil
	}
	prop := strings.TrimSpace(td.Property)
	if prop == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("time_dimension.property is required when time_dimension is provided")
	}
	pol := strings.TrimSpace(td.DefaultRangePolicy)
	if pol != "" {
		if _, ok := validMetricTimeRangePolicies[pol]; !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("invalid time_dimension.default_range_policy %q, expected one of %v",
					pol, validMetricTimeRangePoliciesArr))
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

func validateMetricCalculationFormulaBody(ctx context.Context, f *interfaces.MetricCalculationFormula) error {

	if metricCalculationFormulaEmpty(f) {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula is required when using metric dry run")
	}

	if f.Condition != nil {
		if err := validateMetricCond(ctx, f.Condition); err != nil {
			return err
		}
	}
	if err := validateMetricAggregation(ctx, &f.Aggregation); err != nil {
		return err
	}
	for i := range f.GroupBy {
		p := strings.TrimSpace(f.GroupBy[i].Property)
		if p == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("calculation_formula.group_by[%d].property is required when using metric dry run", i))
		}
	}
	for i := range f.OrderBy {
		p := strings.TrimSpace(f.OrderBy[i].Property)
		if p == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("calculation_formula.order_by[%d].property is required when using metric dry run", i))
		}
		d := strings.TrimSpace(f.OrderBy[i].Direction)
		if d != "" && d != interfaces.MetricOrderDirectionAsc && d != interfaces.MetricOrderDirectionDesc {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
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

func validateMetricAggregation(ctx context.Context, a *interfaces.MetricAggregation) error {
	if a == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.aggregation is required in strict mode")

	}
	if strings.TrimSpace(a.Property) == "" || strings.TrimSpace(a.Aggr) == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.aggregation.property and calculation_formula.aggregation.aggr are required")
	}
	ag := strings.TrimSpace(a.Aggr)
	if _, ok := interfaces.ValidMetricAggrs[ag]; !ok {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("invalid calculation_formula.aggregation.aggr %s, expected one of %v", ag, interfaces.ValidMetricAggrsArr))
	}
	return nil
}

func validateMetricHaving(ctx context.Context, h *interfaces.MetricHaving) error {
	if strings.TrimSpace(h.Field) == "" {
		h.Field = interfaces.MetricHavingFieldValue
	}
	if strings.TrimSpace(h.Operation) == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.having operation is required")
	}
	if h.Value == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula.having value is required")
	}
	switch h.Value.(type) {
	case int, int8, int16, int32, int64, float32, float64:
		return nil
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("calculation_formula.having value must be a number, got %T", h.Value))
	}
}

// validateConditionRecursive 与 bkn-backend validateCond 一致，用于指标公式中 and/or 子条件树。
func validateConditionRecursive(ctx context.Context, cfg *cond.CondCfg) error {
	if cfg == nil {
		return nil
	}
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails("condition operation is required")
	}
	if _, exists := cond.OperationMap[cfg.Operation]; !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails("unsupported condition operation")
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr, cond.OperationKNN:
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("the number of sub_conditions exceeds %d", cond.MaxSubCondition))
		}
		for _, subCond := range cfg.SubConds {
			if err := validateConditionRecursive(ctx, subCond); err != nil {
				return err
			}
		}
	default:
		if cfg.Operation != cond.OperationMultiMatch && cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("condition field name is required")
		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte,
		cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex,
		cond.OperationMatch, cond.OperationMatchPhrase, cond.OperationCurrent, cond.OperationMultiMatch:
		if _, ok := cfg.Value.([]any); ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}
		if cfg.Operation == cond.OperationLike || cfg.Operation == cond.OperationNotLike ||
			cfg.Operation == cond.OperationPrefix || cfg.Operation == cond.OperationNotPrefix {
			if _, ok := cfg.Value.(string); !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}
		if cfg.Operation == cond.OperationRegex {
			_, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails("[regex] operation's value should be a string")
			}
		}
	case cond.OperationIn, cond.OperationNotIn:
		arr, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}
		if len(arr) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[in not_in] operation's value should contain at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		v, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}
		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}
	}
	return nil
}

func validateMetricCond(ctx context.Context, cfg *cond.CondCfg) error {
	if cfg == nil {
		return nil
	}
	if cfg.Operation == "" {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails("condition operation is required")
	}
	if _, exists := cond.OperationMap[cfg.Operation]; !exists {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails("unsupported condition operation")
	}

	switch cfg.Operation {
	case cond.OperationAnd, cond.OperationOr:
		if len(cfg.SubConds) > cond.MaxSubCondition {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("the number of sub_conditions exceeds %d", cond.MaxSubCondition))
		}
		for _, subCond := range cfg.SubConds {
			if err := validateConditionRecursive(ctx, subCond); err != nil {
				return err
			}
		}
	default:
		if cfg.Name == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("condition field name is required")
		}
	}

	switch cfg.Operation {
	case cond.OperationEq, cond.OperationNotEq, cond.OperationGt, cond.OperationGte, cond.OperationLt, cond.OperationLte,
		cond.OperationLike, cond.OperationNotLike, cond.OperationPrefix, cond.OperationNotPrefix, cond.OperationRegex,
		cond.OperationCurrent:
		if _, ok := cfg.Value.([]any); ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("[%s] operation's value should be a single value", cfg.Operation))
		}
		if cfg.Operation == cond.OperationLike || cfg.Operation == cond.OperationNotLike ||
			cfg.Operation == cond.OperationPrefix || cfg.Operation == cond.OperationNotPrefix {
			if _, ok := cfg.Value.(string); !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails("[like not_like prefix not_prefix] operation's value should be a string")
			}
		}
		if cfg.Operation == cond.OperationRegex {
			val, ok := cfg.Value.(string)
			if !ok {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails("[regex] operation's value should be a string")
			}
			if _, err := regexp2.Compile(val, regexp2.RE2); err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
					WithErrorDetails(fmt.Sprintf("[regex] regular expression error: %s", err.Error()))
			}
		}
	case cond.OperationIn, cond.OperationNotIn:
		arr, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[in not_in] operation's value must be an array")
		}
		if len(arr) <= 0 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[in not_in] operation's value should contain at least 1 value")
		}
	case cond.OperationRange, cond.OperationOutRange, cond.OperationBefore, cond.OperationBetween:
		v, ok := cfg.Value.([]any)
		if !ok {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[range, out_range] operation's value must be an array")
		}
		if len(v) != 2 {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails("[range, out_range] operation's value must contain 2 values")
		}
	default:
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
			WithErrorDetails(fmt.Sprintf("[%s] operation is not supported in metric condition", cfg.Operation))
	}
	return nil
}

// metricQueryIsTrendTime mirrors logics/metric: instant omitted or false => 趋势查询.
func metricQueryIsTrendTime(tw *interfaces.MetricTimeWindow) bool {
	return tw != nil && (tw.Instant == nil || !*tw.Instant)
}

// validateMetricCalendarStep checks trend calendar step (与 resource 下推约定一致).
func validateMetricCalendarStep(raw string) error {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch s {
	case "day", "week", "month", "quarter", "year":
		return nil
	default:
		return fmt.Errorf("step must be a calendar interval: day, week, month, quarter, year (got %q)", raw)
	}
}

// validateMetricQueryRequest checks shared MetricQueryRequest fields (query data and dry-run runtime block).
// FillNull is set on the request by the handler via SetFillNullFromQueryParam (URL query fill_null, json:"-").
func validateMetricQueryRequest(ctx context.Context, body *interfaces.MetricQueryRequest) error {
	if body == nil {
		return nil
	}

	// limit可以不传，
	if body.Limit != nil {
		if *body.Limit < 0 || *body.Limit > interfaces.MAX_LIMIT {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("limit must be between 0 and %d when set", interfaces.MAX_LIMIT))
		}
	}
	for i, ob := range body.OrderBy {
		if strings.TrimSpace(ob.Property) == "" {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("order_by[%d].property is required when order_by is present", i))
		}
		d := strings.TrimSpace(ob.Direction)
		if d != "" && d != interfaces.ASC_DIRECTION && d != interfaces.DESC_DIRECTION {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(fmt.Sprintf("order_by[%d].direction must be asc or desc when set", i))
		}
	}
	// Time window shape: all validation lives here; logics layer only merges request time with def.time_dimension.default_range_policy (DESIGN §3.2.2).
	if body.Time != nil {
		tw := body.Time
		if (tw.Start != nil) != (tw.End != nil) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("time.start and time.end must both be set when either is set")
		}
		if tw.Start != nil && tw.End != nil && *tw.Start > *tw.End {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("time.start must be <= time.end")
		}
		if metricQueryIsTrendTime(tw) {
			if tw.Step == nil || strings.TrimSpace(*tw.Step) == "" {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails("trend query requires time.step (calendar interval only)")
			}
			if err := validateMetricCalendarStep(*tw.Step); err != nil {
				return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails(err.Error())
			}
		}
	}
	if body.FillNull {
		if body.Time == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("fill_null requires a time range (time is required)")
		}
		if !metricQueryIsTrendTime(body.Time) {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("fill_null is only valid for trend (range) queries: set time.instant to false or omit it")
		}
		if body.Time.Start == nil || body.Time.End == nil {
			return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("fill_null requires time.start and time.end")
		}
	}
	return nil
}

// validateMetricDryRunForExecution validates metric_config and embedded runtime fields; returns parsed definition.
func validateMetricDryRunForExecution(ctx context.Context, body *interfaces.MetricDryRunRequest) error {
	if body == nil || body.MetricConfig == nil || body.MetricConfig.CalculationFormula == nil {
		return rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric_config with calculation_formula is required")
	}

	if err := validateMetricQueryRequest(ctx, &body.MetricQueryRequest); err != nil {
		return err
	}
	if err := validateMetricDefinitionExecutionScope(ctx, body.MetricConfig); err != nil {
		return err
	}

	// 与 bkn-backend 保存指标时一致（strict 默认 true），不含 id / name / tags / comment
	return validateMetricConfig(ctx, body.MetricConfig)
}
