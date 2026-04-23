// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	cond "ontology-query/common/condition"
)

// Metric scope (bkn-metrics.yaml / DESIGN §3.2).
const (
	ScopeTypeObjectType = "object_type"

	MetricTypeAtomic    = "atomic"
	MetricTypeDerived   = "derived"
	MetricTypeComposite = "composite"

	// 同环比\占比类型
	METRICS_SAMEPERIOD string = "sameperiod"
	METRICS_PROPORTION string = "proportion"

	MetricAggrCountDistinct = "count_distinct"
	MetricAggrSum           = "sum"
	MetricAggrMax           = "max"
	MetricAggrMin           = "min"
	MetricAggrAvg           = "avg"
	MetricAggrCount         = "count"

	MetricOrderDirectionAsc  = "asc"
	MetricOrderDirectionDesc = "desc"

	MetricHavingFieldValue = "__value"

	// 聚合/时序别名字段，与 mdl-uniquery / Vega 约定一致；resource 聚合 alias 为 __value。
	VALUE_FIELD = "__value"
	TIME_FIELD  = "__time"

	// 同环比 metrics.sameperiod_config.method
	METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE = "growth_value"
	METRICS_SAMEPERIOD_METHOD_GROWTH_RATE  = "growth_rate"

	// 同环比计算的时间粒度
	METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY     string = "day"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_MONTH   string = "month"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_QUARTER string = "quarter"
	METRICS_SAMEPERIOD_TIME_GRANULARITY_YEAR    string = "year"

	// DefaultFillNullQuery is the default for URL query "fill_null" on metric data endpoints (mdl-uniquery: range query null-padding).
	DefaultFillNullQuery = "false"

	// Default time range policies for MetricTimeDimension (DESIGN appendix B.2; aligned with bkn-backend).
	MetricTimeDefaultRangePolicyLast1h      = "last_1h"
	MetricTimeDefaultRangePolicyLast24h     = "last_24h"
	MetricTimeDefaultRangePolicyCalendarDay = "calendar_day"
	MetricTimeDefaultRangePolicyNone        = "none"
)

var (
	// ValidMetricUnitTypes / ValidMetricModelUnits 与 bkn-backend、OpenAPI MetricModel 对齐。
	ValidMetricUnitTypes = map[string]struct{}{
		"numUnit":          {},
		"storeUnit":        {},
		"percent":          {},
		"transmissionRate": {},
		"timeUnit":         {},
		"currencyUnit":     {},
		"percentageUnit":   {},
		"countUnit":        {},
		"weightUnit":       {},
		"ordinalRankUnit":  {},
	}
	ValidMetricUnitTypesArr = []string{
		"numUnit",
		"storeUnit",
		"percent",
		"transmissionRate",
		"timeUnit",
		"currencyUnit",
		"percentageUnit",
		"countUnit",
		"weightUnit",
		"ordinalRankUnit",
	}

	ValidMetricUnits = map[string]struct{}{
		"none": {}, "K": {}, "Mil": {}, "Bil": {}, "Tri": {},
		"bit": {}, "Byte": {}, "KB": {}, "MB": {}, "GB": {}, "TB": {}, "PB": {},
		"bps": {}, "Kbps": {}, "Mbps": {},
		"μs": {}, "ms": {}, "s": {}, "m": {}, "h": {},
		"day": {}, "week": {}, "month": {}, "year": {}, "quarter": {},
		"Fen": {}, "Jiao": {}, "CNY": {}, "10K_CNY": {}, "1M_CNY": {}, "100M_CNY": {},
		"US_Cent": {}, "USD": {}, "EUR_Cent": {},
		"%": {}, "‰": {},
		"household": {}, "transaction": {}, "piece": {}, "item": {}, "times": {},
		"man_day": {}, "family": {}, "hand": {}, "sheet": {}, "packet": {},
		"ton": {}, "kg": {}, "rank": {},
	}

	ValidMetricUnitsArr = []string{
		"none",
		"K",
		"Mil",
		"Bil",
		"Tri",
		"bit",
		"Byte",
		"KB",
		"MB",
		"GB",
		"TB",
		"PB",
		"bps",
		"Kbps",
		"Mbps",
		"μs",
		"ms",
		"s",
		"m",
		"h",
		"day",
		"week",
		"month",
		"year",
		"quarter",
		"Fen",
		"Jiao",
		"CNY",
		"10K_CNY",
		"1M_CNY",
		"100M_CNY",
		"US_Cent",
		"USD",
		"EUR_Cent",
		"%",
		"‰",
		"household",
		"transaction",
		"piece",
		"item",
		"times",
		"man_day",
		"family",
		"hand",
		"sheet",
		"packet",
		"ton",
		"kg",
		"rank",
	}

	ValidMetricAggrs = map[string]struct{}{
		MetricAggrCountDistinct: {},
		MetricAggrSum:           {},
		MetricAggrMax:           {},
		MetricAggrMin:           {},
		MetricAggrAvg:           {},
		MetricAggrCount:         {},
	}
	ValidMetricAggrsArr = []string{
		MetricAggrCountDistinct,
		MetricAggrSum,
		MetricAggrMax,
		MetricAggrMin,
		MetricAggrAvg,
		MetricAggrCount,
	}
)

// MetricTimeWindow is the shared `time` block in metric query / dry-run requests.
type MetricTimeWindow struct {
	Start   *int64  `json:"start"`
	End     *int64  `json:"end"`
	Instant *bool   `json:"instant"`
	Step    *string `json:"step"`
}

// MetricQueryRequest is POST .../metrics/{metric_id}/data body (DESIGN §3.3.1, appendix B.3; ontology-query.yaml MetricQueryRequestBody).
// order_by / having 与 calculation_formula 内同名块同构（附录 B.1）；metrics 与 uniquery Metrics / §3.3.1.1 一致。
type MetricQueryRequest struct {
	Time               *MetricTimeWindow `json:"time,omitempty"`
	Condition          *cond.CondCfg     `json:"condition,omitempty"`
	AnalysisDimensions []string          `json:"analysis_dimensions,omitempty"`
	OrderBy            []MetricOrderBy   `json:"order_by,omitempty"`
	Having             *MetricHaving     `json:"having,omitempty"`
	Metrics            *Metrics          `json:"metrics,omitempty"`
	Limit              *int              `json:"limit,omitempty"`
	// FillNull is set from URL query fill_null by the handler (SetFillNullFromQueryParam); not in JSON body (mdl-uniquery).
	FillNull bool `json:"-"`
}

// MetricDryRunRequest is POST .../metrics/dry-run body (ontology-query.yaml MetricDryRun).
type MetricDryRunRequest struct {
	MetricConfig *MetricDefinition `json:"metric_config"`
	MetricQueryRequest
}

// MetricTimeDimension is the optional time column semantics for stats (DESIGN §3.2.2).
type MetricTimeDimension struct {
	Property           string `json:"property" mapstructure:"property"`
	DefaultRangePolicy string `json:"default_range_policy,omitempty" mapstructure:"default_range_policy"`
}

// MetricAggregation is the single aggregation block (DESIGN appendix B.1).
type MetricAggregation struct {
	Property string `json:"property" mapstructure:"property"`
	Aggr     string `json:"aggr" mapstructure:"aggr"`
}

// MetricGroupBy is one group_by entry.
type MetricGroupBy struct {
	Property    string `json:"property" mapstructure:"property"`
	Description string `json:"description,omitempty" mapstructure:"description"`
}

// MetricOrderBy is one order_by entry.
type MetricOrderBy struct {
	Property  string `json:"property" mapstructure:"property"`
	Direction string `json:"direction" mapstructure:"direction"`
}

// MetricHaving filters aggregated results (DESIGN appendix B.1).
type MetricHaving struct {
	Field     string `json:"field" mapstructure:"field"`
	Operation string `json:"operation" mapstructure:"operation"`
	Value     any    `json:"value,omitempty" mapstructure:"value"`
}

// MetricCalculationFormula is persisted on MetricDefinition (DESIGN §3.2.3).
type MetricCalculationFormula struct {
	Condition   *cond.CondCfg     `json:"condition,omitempty" mapstructure:"condition"`
	Aggregation MetricAggregation `json:"aggregation" mapstructure:"aggregation"`
	GroupBy     []MetricGroupBy   `json:"group_by,omitempty" mapstructure:"group_by"`
	OrderBy     []MetricOrderBy   `json:"order_by,omitempty" mapstructure:"order_by"`
	Having      *MetricHaving     `json:"having,omitempty" mapstructure:"having"`
}

// MetricAnalysisDimension is one analysis drill-down dimension.
type MetricAnalysisDimension struct {
	Name        string `json:"name" mapstructure:"name"`
	DisplayName string `json:"display_name,omitempty" mapstructure:"display_name"`
}

// MetricDefinition matches bkn-backend GET /metrics/{metric_ids} entries (bkn-metrics.yaml).
type MetricDefinition struct {
	ID      string   `json:"id" mapstructure:"id"`
	KnID    string   `json:"kn_id" mapstructure:"kn_id"`
	Branch  string   `json:"branch,omitempty" mapstructure:"branch"`
	Name    string   `json:"name" mapstructure:"name"`
	Comment string   `json:"comment,omitempty" mapstructure:"comment"`
	Tags    []string `json:"tags,omitempty" mapstructure:"tags"`

	UnitType           string                    `json:"unit_type,omitempty" mapstructure:"unit_type"`
	Unit               string                    `json:"unit,omitempty" mapstructure:"unit"`
	MetricType         string                    `json:"metric_type" mapstructure:"metric_type"`
	ScopeType          string                    `json:"scope_type" mapstructure:"scope_type"`
	ScopeRef           string                    `json:"scope_ref" mapstructure:"scope_ref"`
	TimeDimension      *MetricTimeDimension      `json:"time_dimension,omitempty" mapstructure:"time_dimension"`
	CalculationFormula *MetricCalculationFormula `json:"calculation_formula" mapstructure:"calculation_formula"`
	AnalysisDimensions []MetricAnalysisDimension `json:"analysis_dimensions,omitempty" mapstructure:"analysis_dimensions"`

	Creator    AccountInfo `json:"creator,omitempty" mapstructure:"creator"`
	CreateTime int64       `json:"create_time,omitempty" mapstructure:"create_time"`
	Updater    AccountInfo `json:"updater,omitempty" mapstructure:"updater"`
	UpdateTime int64       `json:"update_time,omitempty" mapstructure:"update_time"`
	ModuleType string      `json:"module_type,omitempty" mapstructure:"module_type"`
}

type MetricResponse struct {
	Model          interface{}     `json:"model,omitempty"`
	Datas          []BknMetricData `json:"datas"`
	VegaDurationMs int64           `json:"vega_duration_ms,omitempty"`
	OverallMs      int64           `json:"overall_ms,omitempty"`
}

type BknMetricData struct {
	Labels       map[string]string `json:"labels"`
	Times        []any             `json:"times"`
	TimeStrs     []string          `json:"time_strs,omitempty"`
	Values       []any             `json:"values"`
	GrowthValues []any             `json:"growth_values,omitempty"`
	GrowthRates  []any             `json:"growth_rates,omitempty"`
	Proportions  []any             `json:"proportions,omitempty"`
}
