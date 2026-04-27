// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package interfaces

import (
	"encoding/json"

	cond "bkn-backend/common/condition"
)

// Metric type and scope constants (bkn-metrics.yaml, DESIGN §3.2).
const (
	MetricTypeAtomic    = "atomic"
	MetricTypeDerived   = "derived"
	MetricTypeComposite = "composite"

	ScopeTypeObjectType = "object_type"
	ScopeTypeSubgraph   = "subgraph"

	// Default time range policies for MetricTimeDimension (DESIGN appendix B.2).
	MetricTimeDefaultRangePolicyLast1h      = "last_1h"
	MetricTimeDefaultRangePolicyLast24h     = "last_24h"
	MetricTimeDefaultRangePolicyCalendarDay = "calendar_day"
	MetricTimeDefaultRangePolicyNone        = "none"

	// Aggregation functions for MetricAggregation.aggr (DESIGN appendix B.1).
	MetricAggrCountDistinct = "count_distinct"
	MetricAggrSum           = "sum"
	MetricAggrMax           = "max"
	MetricAggrMin           = "min"
	MetricAggrAvg           = "avg"
	MetricAggrCount         = "count"

	MetricOrderDirectionAsc  = "asc"
	MetricOrderDirectionDesc = "desc"

	// MetricHavingFieldValue is the fixed field name for formula.having (DESIGN appendix B.1).
	MetricHavingFieldValue = "__value"
)

var (
	// ValidMetricUnitTypes 单位类型枚举
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

	// ValidMetricUnits 度量单位枚举
	ValidMetricUnits = map[string]struct{}{
		"none":        {},
		"K":           {},
		"Mil":         {},
		"Bil":         {},
		"Tri":         {},
		"bit":         {},
		"Byte":        {},
		"KB":          {},
		"MB":          {},
		"GB":          {},
		"TB":          {},
		"PB":          {},
		"bps":         {},
		"Kbps":        {},
		"Mbps":        {},
		"μs":          {},
		"ms":          {},
		"s":           {},
		"m":           {},
		"h":           {},
		"day":         {},
		"week":        {},
		"month":       {},
		"year":        {},
		"quarter":     {},
		"Fen":         {},
		"Jiao":        {},
		"CNY":         {},
		"10K_CNY":     {},
		"1M_CNY":      {},
		"100M_CNY":    {},
		"US_Cent":     {},
		"USD":         {},
		"EUR_Cent":    {},
		"%":           {},
		"‰":           {},
		"household":   {},
		"transaction": {},
		"piece":       {},
		"item":        {},
		"times":       {},
		"man_day":     {},
		"family":      {},
		"hand":        {},
		"sheet":       {},
		"packet":      {},
		"ton":         {},
		"kg":          {},
		"rank":        {},
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

	// ValidMetricAggrs 聚合函数枚举（与 ontology-query MetricAggr 对齐）。
	ValidMetricAggrs = map[string]struct{}{
		MetricAggrCountDistinct: {},
		MetricAggrSum:           {},
		MetricAggrMax:           {},
		MetricAggrMin:           {},
		MetricAggrAvg:           {},
		MetricAggrCount:         {},
	}
)

// MetricTimeDimension is the optional time column semantics for stats (DESIGN §3.2.2, appendix B.2).
type MetricTimeDimension struct {
	Property           string `json:"property" mapstructure:"property"`
	DefaultRangePolicy string `json:"default_range_policy,omitempty" mapstructure:"default_range_policy"`
}

// UnmarshalJSON accepts "property" (preferred) or legacy "field" for persisted time_dimension JSON.
func (m *MetricTimeDimension) UnmarshalJSON(data []byte) error {
	var aux struct {
		Property           string `json:"property"`
		Field              string `json:"field"`
		DefaultRangePolicy string `json:"default_range_policy,omitempty"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	if aux.Property != "" {
		m.Property = aux.Property
	} else {
		m.Property = aux.Field
	}
	m.DefaultRangePolicy = aux.DefaultRangePolicy
	return nil
}

// MetricAggregation is the single aggregation block (DESIGN appendix B.1).
type MetricAggregation struct {
	Property string `json:"property" mapstructure:"property"`
	Aggr     string `json:"aggr" mapstructure:"aggr"`
}

// MetricGroupBy is one group_by entry (property-based; DESIGN appendix B.1).
type MetricGroupBy struct {
	Property    string `json:"property" mapstructure:"property"`
	Description string `json:"description,omitempty" mapstructure:"description"`
}

// MetricOrderBy is one order_by entry (DESIGN appendix B.1).
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

// MetricCalculationFormula is dialect-agnostic filter / aggregate / group / sort / having (DESIGN §3.2.3, appendix B.1).
// condition is the same JSON shape as common/condition.CondCfg (ontology-query Condition).
type MetricCalculationFormula struct {
	Condition   *cond.CondCfg     `json:"condition,omitempty" mapstructure:"condition"`
	Aggregation MetricAggregation `json:"aggregation" mapstructure:"aggregation"`
	GroupBy     []MetricGroupBy   `json:"group_by,omitempty" mapstructure:"group_by"`
	OrderBy     []MetricOrderBy   `json:"order_by,omitempty" mapstructure:"order_by"`
	Having      *MetricHaving     `json:"having,omitempty" mapstructure:"having"`
}

// MetricAnalysisDimension is one analysis drill-down dimension (DESIGN appendix B.2).
type MetricAnalysisDimension struct {
	Name        string `json:"name" mapstructure:"name"`
	DisplayName string `json:"display_name,omitempty" mapstructure:"display_name"`
}

// MetricDefinition is the persisted metric entity (DESIGN §3.2.1, bkn-metrics.yaml MetricDefinition).
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

	Vector []float32 `json:"_vector,omitempty"`
	Score  *float64  `json:"_score,omitempty"`

	// BKNRawContent is used only for concept index embedding text (not returned in API JSON).
	BKNRawContent string `json:"-" mapstructure:"-"`
}

// ReqMetrics is the batch-create body for POST .../metrics with x-http-method-override: POST (bkn-metrics.yaml ReqMetrics).
type ReqMetrics struct {
	Entries []*MetricDefinition `json:"entries" mapstructure:"entries"`
}

// MetricsListQueryParams lists metrics under a knowledge network (GET .../metrics query params).
type MetricsListQueryParams struct {
	PaginationQueryParameters
	NamePattern string
	Tag         string
	GroupID     string
	Branch      string
	KNID        string
	ScopeType   string
	ScopeRef    string
}

// MetricsList is the list response for GET .../metrics (bkn-metrics.yaml ListMetrics: entries, total_count).
type MetricsList struct {
	Entries    []*MetricDefinition `json:"entries"`
	TotalCount int64               `json:"total_count"`
}

// MetricSearchResult is the POST override GET response (concept search, MetricSearchResponse in OpenAPI).
type MetricSearchResult struct {
	Entries     []*MetricDefinition `json:"entries"`
	TotalCount  int64               `json:"total_count,omitempty"`
	SearchAfter []any               `json:"search_after,omitempty"`
	Groups      []any               `json:"groups,omitempty"`
	Type        string              `json:"type,omitempty"`
}

var (
	MetricSort = map[string]string{
		"name":        "f_name",
		"update_time": "f_update_time",
	}
)
