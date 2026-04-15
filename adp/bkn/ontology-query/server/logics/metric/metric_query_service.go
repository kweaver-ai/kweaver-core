// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	"ontology-query/logics"
)

var (
	metricQueryServiceOnce sync.Once
	metricQueryServiceInst interfaces.MetricQueryService
)

type metricQueryService struct {
	appSetting *common.AppSetting
	oma        interfaces.OntologyManagerAccess
	mfa        interfaces.ModelFactoryAccess
	vba        interfaces.VegaBackendAccess
}

// NewMetricQueryService constructs the metric query service (same pattern as NewObjectTypeService / bkn-backend NewMetricService).
func NewMetricQueryService(appSetting *common.AppSetting) interfaces.MetricQueryService {
	metricQueryServiceOnce.Do(func() {
		metricQueryServiceInst = &metricQueryService{
			appSetting: appSetting,
			oma:        logics.OMA,
			mfa:        logics.MFA,
			vba:        logics.VBA,
		}
	})
	return metricQueryServiceInst
}

func metricHavingToResourceMap(h *interfaces.MetricHaving, _ interfaces.ObjectType) (map[string]any, error) {
	if h == nil {
		return nil, nil
	}
	field := h.Field
	if field == "" {
		field = "__value"
	}
	if field != "__value" {
		return nil, fmt.Errorf("having field should be empty or __value, got %s", field)
	}
	return map[string]any{
		"field":     field,
		"operation": h.Operation,
		"value":     h.Value,
	}, nil
}

func mergeConditions(a, b *cond.CondCfg) *cond.CondCfg {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return &cond.CondCfg{
		Operation: cond.OperationAnd,
		SubConds:  []*cond.CondCfg{a, b},
	}
}

// mapObjectDataPropertyToResourceField maps an object-type data property name to resource payload field (mapped_field.Name).
func mapDataPropertyToResourceField(propName string, propMap map[string]*cond.DataProperty) (string, error) {
	prop, ok := propMap[propName]
	if !ok {
		return "", fmt.Errorf("属性[%s]不是对象类的数据属性", propName)
	}
	if prop.MappedField.Name == "" {
		return "", fmt.Errorf("属性[%s]未配置映射列(mapped_field)，无法下推到 resource", propName)
	}
	return prop.MappedField.Name, nil
}

// mapMetricOrderByPropertyToResourceSortField maps calculation_formula.order_by / request order_by property
// (object-type data property name, or __value for the aggregation alias) to resource payload field names for Sort.
func mapMetricOrderByPropertyToResourceSortField(propName string, propMap map[string]*cond.DataProperty) (string, error) {
	p := strings.TrimSpace(propName)
	if p == "" {
		return "", fmt.Errorf("order_by property is empty")
	}
	if p == "__value" {
		return "__value", nil
	}
	return mapDataPropertyToResourceField(p, propMap)
}

// buildResourceSortFromMergedOrder maps merged order_by entries to resource []*SortParams (field = mapped_field / __value).
func buildResourceSortFromMergedOrder(merged []interfaces.MetricOrderBy, propMap map[string]*cond.DataProperty) ([]*interfaces.SortParams, error) {
	if len(merged) == 0 {
		return nil, nil
	}
	out := make([]*interfaces.SortParams, 0, len(merged))
	for _, o := range merged {
		p := strings.TrimSpace(o.Property)
		if p == "" {
			continue
		}
		resField, err := mapMetricOrderByPropertyToResourceSortField(p, propMap)
		if err != nil {
			return nil, err
		}
		dir := strings.TrimSpace(o.Direction)
		if dir == "" {
			dir = interfaces.ASC_DIRECTION
		}
		if dir != interfaces.ASC_DIRECTION && dir != interfaces.DESC_DIRECTION {
			return nil, fmt.Errorf("order_by direction for property %s must be asc or desc", p)
		}
		out = append(out, &interfaces.SortParams{Field: resField, Direction: dir})
	}
	return out, nil
}

// trendMeta carries calendar trend context for mapping vega rows to MetricData (times / step / is_calendar).
type trendMeta struct {
	step         string
	timeProperty string
	timeResField string
}

// isTrendMetricTime must match driveradapters.metricQueryIsTrendTime: instant omitted or false => range (trend) query.
func isTrendMetricTime(tw *interfaces.MetricTimeWindow) bool {
	return tw != nil && (tw.Instant == nil || !*tw.Instant)
}

func startOfLocalCalendarDayMs() int64 {
	tz := os.Getenv("TZ")
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	n := time.Now().In(loc)
	t := time.Date(n.Year(), n.Month(), n.Day(), 0, 0, 0, 0, loc)
	return t.UnixMilli()
}

// defaultRangePolicyToMS resolves time_dimension.default_range_policy to [start, end] in unix milliseconds (end inclusive semantics for callers who add +1 for lt).
func defaultRangePolicyToMS(policy string) (start, end int64, ok bool, err error) {
	p := strings.TrimSpace(strings.ToLower(policy))
	now := time.Now().UnixMilli()
	switch p {
	case interfaces.MetricTimeDefaultRangePolicyLast1h:
		return now - 3600_000, now, true, nil
	case interfaces.MetricTimeDefaultRangePolicyLast24h:
		return now - 24*3600_000, now, true, nil
	case interfaces.MetricTimeDefaultRangePolicyCalendarDay:
		return startOfLocalCalendarDayMs(), now, true, nil
	case interfaces.MetricTimeDefaultRangePolicyNone, "":
		return 0, 0, false, nil
	default:
		return 0, 0, false, fmt.Errorf("invalid default_range_policy %q", policy)
	}
}

// mergeMetricTimeRangeMS resolves the effective [start, end] by combining the request time window (when both are set)
// with time_dimension.default_range_policy from the definition when the request does not provide a full window.
// DESIGN §3.2.2: default policies (last_1h, last_24h, calendar_day, none) apply only when the client does not pass dynamic time.
// Request pairing (start/end together) and start<=end are validated only in driveradapters.validateMetricQueryRequest.
func mergeMetricTimeRangeMS(def *interfaces.MetricDefinition, tw *interfaces.MetricTimeWindow) (start, end int64, ok bool, err error) {
	if tw != nil && tw.Start != nil && tw.End != nil {
		return *tw.Start, *tw.End, true, nil
	}
	if def == nil || def.TimeDimension == nil {
		return 0, 0, false, nil
	}
	pol := strings.TrimSpace(def.TimeDimension.DefaultRangePolicy)
	return defaultRangePolicyToMS(pol)
}

func resolveTrendTimeRangeMS(def *interfaces.MetricDefinition, tw *interfaces.MetricTimeWindow) (start, end int64, err error) {
	s, e, ok, err := mergeMetricTimeRangeMS(def, tw)
	if err != nil {
		return 0, 0, err
	}
	if !ok {
		return 0, 0, fmt.Errorf("trend query requires time.start and time.end, or time_dimension.default_range_policy other than none")
	}
	return s, e, nil
}

func (s *metricQueryService) buildResourceDataQueryParams(ctx context.Context, def *interfaces.MetricDefinition,
	metricQuery *interfaces.MetricQueryRequest, ot interfaces.ObjectType) (*interfaces.ResourceDataQueryParams, *trendMeta, error) {

	if def.CalculationFormula == nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("calculation_formula is required on metric definition")
	}
	metricFormula := def.CalculationFormula

	// ot 的 property 变成 map
	propMap := logics.TransferPropsToPropMap(ot.DataProperties)

	var trend *trendMeta
	var timeCond *cond.CondCfg

	if metricQuery != nil && isTrendMetricTime(metricQuery.Time) {
		tw := metricQuery.Time
		if def.TimeDimension == nil || strings.TrimSpace(def.TimeDimension.Property) == "" {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("trend query requires metric time_dimension.property on the definition")
		}
		// step 与 time 窗口形态由 handler validateMetricQueryRequest 校验；此处仅归一化日历步长
		calStep := strings.TrimSpace(strings.ToLower(*tw.Step))
		startMs, endMs, err := resolveTrendTimeRangeMS(def, tw)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(err.Error())
		}
		timeProp := strings.TrimSpace(def.TimeDimension.Property)
		timeResField, err := mapDataPropertyToResourceField(timeProp, propMap)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(err.Error())
		}
		trend = &trendMeta{step: calStep, timeProperty: timeProp, timeResField: timeResField}
		// range 为左闭右开 [gte, lt)；between 为双闭区间 [gte, lt]
		timeCond = &cond.CondCfg{
			Operation: cond.OperationBetween,
			Name:      timeProp,
			ValueOptCfg: cond.ValueOptCfg{
				Value: []any{float64(startMs), float64(endMs)},
			},
		}
	} else {
		var tw *interfaces.MetricTimeWindow
		if metricQuery != nil {
			tw = metricQuery.Time
		}
		s, e, ok, err := mergeMetricTimeRangeMS(def, tw)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(err.Error())
		}
		if ok {
			if def.TimeDimension == nil || strings.TrimSpace(def.TimeDimension.Property) == "" {
				return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails("time range filter requires metric time_dimension.property")
			}
			timeProp := strings.TrimSpace(def.TimeDimension.Property)
			if _, err := mapDataPropertyToResourceField(timeProp, propMap); err != nil {
				return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails(err.Error())
			}
			timeCond = &cond.CondCfg{
				Operation: cond.OperationRange,
				Name:      timeProp,
				ValueOptCfg: cond.ValueOptCfg{
					Value: []any{float64(s), float64(e) + 1},
				},
			}
		}
	}

	// 处理过滤条件
	var reqCond *cond.CondCfg
	if metricQuery != nil {
		reqCond = metricQuery.Condition
	}
	merged := mergeConditions(metricFormula.Condition, reqCond)
	merged = mergeConditions(merged, timeCond)
	var fc map[string]any
	if merged != nil {
		rewriteCondition, err := cond.RewriteCondition(ctx, merged, propMap,
			func(ctx context.Context, property *cond.DataProperty, word string) ([]cond.VectorResp, error) {
				return s.handlerVector(ctx, property, word)
			})
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_InvalidParameter_Condition).
				WithErrorDetails(fmt.Sprintf("failed to rewrite ontology condition for resource, %s", err.Error()))
		}
		fc = logics.CondCfgToFilterMap(rewriteCondition)
	}

	params := &interfaces.ResourceDataQueryParams{
		FilterCondition: fc,
	}

	// 处理聚合
	aggrProp, err := mapDataPropertyToResourceField(metricFormula.Aggregation.Property, propMap)
	if err != nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(err.Error())
	}
	params.Aggregation = map[string]any{
		"property": aggrProp,
		"aggr":     metricFormula.Aggregation.Aggr,
		"alias":    "__value",
	}

	// 处理分组字段
	var gb []map[string]any
	if len(metricFormula.GroupBy) > 0 {
		gb = make([]map[string]any, 0, len(metricFormula.GroupBy))
		for _, g := range metricFormula.GroupBy {
			resProp, err := mapDataPropertyToResourceField(g.Property, propMap)
			if err != nil {
				return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails(err.Error())
			}
			gb = append(gb, map[string]any{
				"property": resProp,
			})
		}
	}
	if trend != nil {
		// 拼上时间趋势的分组聚合
		gb = append(gb, map[string]any{
			"property":          trend.timeResField,
			"calendar_interval": trend.step, // 日历步长
		})
	}
	if len(gb) > 0 {
		params.GroupBy = gb
	}

	// 排序：先合并 definition.order_by 与 request.order_by（definition 在前，同 property 以后者为准），
	// 再逐项将对象类属性名/__value 转为 resource 映射列，组装为 params.Sort 交给执行层 resource 排序。
	var mergedOrder []interfaces.MetricOrderBy
	if len(metricFormula.OrderBy) > 0 {
		mergedOrder = append(mergedOrder, metricFormula.OrderBy...)
	}
	if len(metricQuery.OrderBy) > 0 {
		mergedOrder = append(mergedOrder, metricQuery.OrderBy...)
	}

	// 时间趋势排序bkn拼不了，因为源不同，排序语法不同，由vega适配
	// if trend != nil {
	// 	// 拼接上按时间分组字段排序
	// 	mergedOrder = append(mergedOrder, interfaces.MetricOrderBy{
	// 		Property:  trend.timeProperty,
	// 		Direction: interfaces.ASC_DIRECTION,
	// 	})
	// }
	resourceSort, err := buildResourceSortFromMergedOrder(mergedOrder, propMap)
	if err != nil {
		return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails(err.Error())
	}
	if len(resourceSort) > 0 {
		params.Sort = resourceSort
	}

	// 优先用请求的having，请求没有，则用definition的having，都没有，则不加having
	if metricQuery != nil && metricQuery.Having != nil {
		hm, err := metricHavingToResourceMap(metricQuery.Having, ot)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(err.Error())
		}
		params.Having = hm
	} else if metricFormula.Having != nil {
		hm, err := metricHavingToResourceMap(metricFormula.Having, ot)
		if err != nil {
			return nil, nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails(err.Error())
		}
		params.Having = hm
	}
	// 没有limit，则表示查全部
	if metricQuery != nil && metricQuery.Limit != nil && *metricQuery.Limit > 0 {
		params.Limit = *metricQuery.Limit
	}
	return params, trend, nil
}

// handlerVector resolves text to vectors for condition rewrite (same role as object_type_service).
func (s *metricQueryService) handlerVector(ctx context.Context, property *cond.DataProperty, word string) ([]cond.VectorResp, error) {
	if property == nil || property.IndexConfig == nil {
		return nil, fmt.Errorf("vector 条件需要属性索引配置")
	}
	model, err := s.mfa.GetModelByID(ctx, property.IndexConfig.VectorConfig.ModelID)
	if err != nil {
		return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError,
			oerrors.OntologyQuery_ObjectType_InternalError_GetSmallModelByIDFailed).
			WithErrorDetails(err.Error())
	}
	if model == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusNotFound,
			oerrors.OntologyQuery_ObjectType_SmallModelNotFound).
			WithErrorDetails(fmt.Sprintf("小模型[%s]不存在", property.IndexConfig.VectorConfig.ModelID))
	}
	if model.EmbeddingDim == 0 || model.BatchSize == 0 || model.MaxTokens == 0 {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest,
			oerrors.OntologyQuery_ObjectType_InvalidParameter_SmallModel).
			WithErrorDetails(fmt.Sprintf("model %s has invalid embedding dim, batch size or max tokens", model.ModelID))
	}
	return s.mfa.GetVector(ctx, model, []string{word})
}

// metricGroupByDimension 描述 group by 的一个维度：PropertyName 为对象类 data property 名（接口返回 labels 的 key）；
// ResourceFieldName 为下推到 Vega/resource 的列名（与 buildResourceDataQueryParams 中 group_by 的 "property" 一致）。
type metricGroupByDimension struct {
	PropertyName      string
	ResourceFieldName string
}

// metricGroupByDimensions 从 def 与 query 收集维度顺序（与下推时 group 维度一致，顺序：先 group_by 后 analysis_dimensions 去重），
// 并对每一项解析出 resource 列名。当 query 带 analysis_dimensions 时，只保留在定义与请求的交集中，顺序与 query 一致；交集为空则返回空 slice。
func metricGroupByDimensions(def *interfaces.MetricDefinition, query *interfaces.MetricQueryRequest, propMap map[string]*cond.DataProperty) ([]metricGroupByDimension, error) {
	if def == nil {
		return nil, nil
	}
	if propMap == nil {
		return nil, fmt.Errorf("propMap is required for group by dimension mapping")
	}
	defined := make(map[string]struct{})
	var ordered []string
	add := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if _, ok := defined[s]; ok {
			return
		}
		defined[s] = struct{}{}
		ordered = append(ordered, s)
	}
	if def.CalculationFormula != nil {
		for _, g := range def.CalculationFormula.GroupBy {
			add(g.Property)
		}
	}

	var propertyNames []string
	if query == nil || len(query.AnalysisDimensions) == 0 {
		propertyNames = ordered
	} else {
		seen := make(map[string]struct{})
		for _, a := range query.AnalysisDimensions {
			p := strings.TrimSpace(a)
			if p == "" {
				continue
			}
			if _, ok := defined[p]; !ok {
				continue
			}
			if _, dup := seen[p]; dup {
				continue
			}
			seen[p] = struct{}{}
			propertyNames = append(propertyNames, p)
		}
	}

	out := make([]metricGroupByDimension, 0, len(propertyNames))
	for _, p := range propertyNames {
		res, err := mapDataPropertyToResourceField(p, propMap)
		if err != nil {
			return nil, err
		}
		out = append(out, metricGroupByDimension{PropertyName: p, ResourceFieldName: res})
	}
	return out, nil
}

func bknMetricDataSliceToDataSlice(in []interfaces.BknMetricData) []interfaces.Data {
	if len(in) == 0 {
		return nil
	}
	out := make([]interfaces.Data, len(in))
	for i := range in {
		out[i] = interfaces.Data{
			Labels:       in[i].Labels,
			Times:        in[i].Times,
			Values:       in[i].Values,
			GrowthValues: in[i].GrowthValues,
			GrowthRates:  in[i].GrowthRates,
			Proportions:  in[i].Proportions,
		}
	}
	return out
}

// buildEntryDimKey matches parseVegaResult2Uniresponse: concat dimension column values in fixed order, excluding the aggregate __value.
func buildEntryDimKey(entry map[string]any, groupDims []metricGroupByDimension) string {
	var b strings.Builder
	for _, d := range groupDims {
		b.WriteString(fmt.Sprintf("%v|", entry[d.ResourceFieldName]))
	}
	return b.String()
}

func getProportionTotalEntries(ctx context.Context, entries []map[string]any) (float64, error) {
	var total float64
	for _, e := range entries {
		v, ok := e[interfaces.VALUE_FIELD]
		if !ok {
			continue
		}
		f, err := common.AnyToFloat64(v)
		if err != nil {
			return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
				WithErrorDetails(err.Error())
		}
		total += f
	}
	return total, nil
}

// appendMetricValue appends a scalar to Values (raw float64, same as bkn/HTTP expectations).
func appendMetricValue(ctx context.Context, v any, values *[]any) (float64, error) {
	if v == nil {
		*values = append(*values, nil)
		return 0, nil
	}
	f, err := common.AnyToFloat64(v)
	if err != nil {
		return 0, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error())
	}
	*values = append(*values, f)
	return f, nil
}

// entryTimeToMillis supports numeric buckets (ms) and RFC3339 strings (resource / Vega may use either).
func entryTimeToMillis(v any) (int64, error) {
	if v == nil {
		return 0, fmt.Errorf("time field is nil")
	}
	if s, ok := v.(string); ok {
		t, err := time.ParseInLocation(time.RFC3339, s, common.APP_LOCATION)
		if err != nil {
			return 0, err
		}
		return t.UnixMilli(), nil
	}
	f, err := common.AnyToFloat64(v)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

// vegaEntriesToMetricData maps resource /resources/:id/data "entries" ([]map) to the same BknMetricData shape
// as mdl-uniquery parseVegaResult2Uniresponse (Vega: Columns + Data rows). Labels use object data property names, not resource column names.
func vegaEntriesToMetricData(ctx context.Context, def interfaces.MetricDefinition,
	datas *interfaces.DatasetQueryResponse,
	samePeriodDatas *interfaces.DatasetQueryResponse, query *interfaces.MetricQueryRequest,
	trend *trendMeta, vegaFetchDur int64, propMap map[string]*cond.DataProperty) (interfaces.MetricResponse, error) {

	groupDims, err := metricGroupByDimensions(&def, query, propMap)
	if err != nil {
		return interfaces.MetricResponse{}, err
	}

	if trend != nil {
		return convertVegaDatas2TimeSeries(ctx, def, datas, samePeriodDatas, query, trend, vegaFetchDur, propMap)
	}

	entries := []map[string]any(nil)
	if datas != nil {
		entries = datas.Entries
	}

	resp := interfaces.MetricResponse{
		VegaDurationMs: vegaFetchDur,
	}

	var total float64
	hasGrowthValue := false
	hasGrowthRate := false
	samePeriodMap := make(map[string]float64)
	if query != nil && query.Metrics != nil {
		switch query.Metrics.Type {
		case interfaces.METRICS_PROPORTION:
			total, err = getProportionTotalEntries(ctx, entries)
			if err != nil {
				return resp, err
			}
		case interfaces.METRICS_SAMEPERIOD:
			if query.Metrics.SameperiodConfig == nil {
				return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
					WithErrorDetails("sameperiod_config is required for metrics type sameperiod")
			}
			methods := query.Metrics.SameperiodConfig.Method
			for _, method := range methods {
				if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE {
					hasGrowthValue = true
				}
				if method == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE {
					hasGrowthRate = true
				}
			}
			peerEntries := []map[string]any(nil)
			if samePeriodDatas != nil {
				peerEntries = samePeriodDatas.Entries
			}
			for _, entry := range peerEntries {
				key := buildEntryDimKey(entry, groupDims)
				v, ok := entry[interfaces.VALUE_FIELD]
				if !ok {
					continue
				}
				value, perr := common.AnyToFloat64(v)
				if perr != nil {
					return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
						WithErrorDetails(perr.Error())
				}
				samePeriodMap[key] = value
			}
		}
	}

	bknRows := make([]interfaces.BknMetricData, 0, len(entries))
	for _, entry := range entries {
		labels := make(map[string]string, len(groupDims))
		values := make([]any, 0, 1)
		growthValues := make([]any, 0)
		growthRates := make([]any, 0)
		proportions := make([]any, 0)

		key := buildEntryDimKey(entry, groupDims)
		for _, d := range groupDims {
			labels[d.PropertyName] = fmt.Sprintf("%v", entry[d.ResourceFieldName])
		}

		var currentValue float64
		agg, ok := entry[interfaces.VALUE_FIELD]
		if !ok {
			_, aerr := appendMetricValue(ctx, nil, &values)
			if aerr != nil {
				return resp, aerr
			}
		} else {
			var aerr error
			currentValue, aerr = appendMetricValue(ctx, agg, &values)
			if aerr != nil {
				return resp, aerr
			}
		}

		if query != nil && query.Metrics != nil {
			switch query.Metrics.Type {
			case interfaces.METRICS_SAMEPERIOD:
				if samePeriod, exists := samePeriodMap[key]; exists {
					if hasGrowthValue {
						growthValues = append(growthValues, currentValue-samePeriod)
					}
					if hasGrowthRate {
						if samePeriod != 0 {
							growthRates = append(growthRates, (currentValue-samePeriod)/samePeriod*100)
						} else {
							growthRates = append(growthRates, nil)
						}
					}
				} else {
					if hasGrowthValue {
						growthValues = append(growthValues, nil)
					}
					if hasGrowthRate {
						growthRates = append(growthRates, nil)
					}
				}
			case interfaces.METRICS_PROPORTION:
				if total != 0 {
					proportions = append(proportions, currentValue/total*100)
				} else {
					proportions = append(proportions, nil)
				}
			}
		}

		var instantMs int64
		if query != nil && query.Time != nil && query.Time.End != nil {
			instantMs = *query.Time.End
		}
		timeStr := common.FormatRFC3339Milli(instantMs)
		mData := interfaces.BknMetricData{
			Labels:       labels,
			Times:        []any{instantMs},
			TimeStrs:     []string{timeStr},
			Values:       values,
			GrowthRates:  growthRates,
			GrowthValues: growthValues,
			Proportions:  proportions,
		}
		bknRows = append(bknRows, mData)
	}

	resp.Datas = bknRows
	return resp, nil
}

// metricQuery.FillNull 由 handler 从 URL 查询 fill_null 解析写入（json:"-"）。
func (s *metricQueryService) executeMetric(ctx context.Context, knID string, branch string,
	def *interfaces.MetricDefinition, metricQuery *interfaces.MetricQueryRequest) (interfaces.MetricData, error) {

	// 获取统计主体-对象类信息
	ot, ok, err := s.oma.GetObjectType(ctx, knID, branch, def.ScopeRef)
	if err != nil {
		logger.Errorf("GetObjectType for metric scope: %v", err)
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error())
	}
	if !ok {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_Metric_ObjectTypeNotFound)
	}
	if ot.DataSource == nil || ot.DataSource.Type != interfaces.DATA_SOURCE_TYPE_RESOURCE || ot.DataSource.ID == "" {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidDataSource)
	}

	params, trend, err := s.buildResourceDataQueryParams(ctx, def, metricQuery, ot)
	if err != nil {
		return interfaces.MetricData{}, err
	}

	// vega查数, 记录请求的开始结束,返回到接口上
	start := time.Now().UnixMilli()
	datas, err := s.vba.QueryResourceData(ctx, ot.DataSource.ID, params)
	if err != nil {
		logger.Errorf("QueryResourceData: %v", err)
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error())
	}
	if datas == nil {
		return interfaces.MetricData{}, nil
	}

	// 如果有请求同环比,计算同期对应的时间范围
	samePeriodDatas := &interfaces.DatasetQueryResponse{}
	if metricQuery.Metrics != nil && metricQuery.Metrics.Type == interfaces.METRICS_SAMEPERIOD {
		if metricQuery.Time == nil || metricQuery.Time.Start == nil || metricQuery.Time.End == nil {
			return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("sameperiod requires time.start and time.end")
		}
		if metricQuery.Metrics.SameperiodConfig == nil {
			return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("sameperiod_config is required for metrics type sameperiod")
		}
		startTime := time.UnixMilli(*metricQuery.Time.Start).In(common.APP_LOCATION)
		endTime := time.UnixMilli(*metricQuery.Time.End).In(common.APP_LOCATION)
		chainStart := calcComparisonTime(startTime, *metricQuery.Metrics.SameperiodConfig).UnixMilli()
		chainEnd := calcComparisonTime(endTime, *metricQuery.Metrics.SameperiodConfig).UnixMilli()
		tw := *metricQuery.Time
		tw.Start = &chainStart
		tw.End = &chainEnd
		nq := *metricQuery
		nq.Time = &tw

		params, _, err := s.buildResourceDataQueryParams(ctx, def, &nq, ot)
		if err != nil {
			return interfaces.MetricData{}, err
		}

		samePeriodDatas, err = s.vba.QueryResourceData(ctx, ot.DataSource.ID, params)
		if err != nil {
			logger.Errorf("QueryResourceData: %v", err)
			return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
				WithErrorDetails(err.Error())
		}
		if samePeriodDatas == nil {
			return interfaces.MetricData{}, nil
		}
	}
	vegaFetchDur := time.Now().UnixMilli() - start

	propMap := logics.TransferPropsToPropMap(ot.DataProperties)
	vegaResp, err := vegaEntriesToMetricData(ctx, *def, datas, samePeriodDatas, metricQuery, trend, vegaFetchDur, propMap)
	if err != nil {
		return interfaces.MetricData{}, err
	}

	out := interfaces.MetricData{
		Model:      interfaces.MetricModel{UnitType: def.UnitType, Unit: def.Unit},
		Datas:      bknMetricDataSliceToDataSlice(vegaResp.Datas),
		IsVariable: false,
		IsCalendar: false,
	}
	if trend != nil {
		out.Step = trend.step
		out.IsCalendar = true
	}
	return out, nil
}

func (s *metricQueryService) GetMetricDefinition(ctx context.Context, knID, branch, metricID string) (*interfaces.MetricDefinition, bool, error) {
	return s.oma.GetMetricDefinition(ctx, knID, branch, metricID)
}

func (s *metricQueryService) QueryMetricData(ctx context.Context, knID string, branch string, metricID string,
	metricQuery *interfaces.MetricQueryRequest) (interfaces.MetricData, error) {

	if metricQuery == nil {
		metricQuery = &interfaces.MetricQueryRequest{}
	}
	def, exist, err := s.oma.GetMetricDefinition(ctx, knID, branch, metricID)
	if err != nil {
		if httpErr, ok := err.(*rest.HTTPError); ok {
			return interfaces.MetricData{}, httpErr
		}
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails(err.Error())
	}
	if !exist || def == nil {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusNotFound, oerrors.OntologyQuery_Metric_NotFound)
	}
	// 指标的统计主体为 对象类
	if def.ScopeType != interfaces.ScopeTypeObjectType {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_UnsupportedScope)
	}
	return s.executeMetric(ctx, knID, branch, def, metricQuery)
}

func (s *metricQueryService) DryRunMetricData(ctx context.Context, knID, branch string,
	metricDryRun *interfaces.MetricDryRunRequest) (interfaces.MetricData, error) {

	if metricDryRun == nil || metricDryRun.MetricConfig == nil || metricDryRun.MetricConfig.CalculationFormula == nil {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric_config with calculation_formula is required")
	}
	def := metricDryRun.MetricConfig
	if def.KnID != "" && def.KnID != knID {
		return interfaces.MetricData{}, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("metric_config.kn_id must match path kn_id")
	}

	return s.executeMetric(ctx, knID, branch, def, &metricDryRun.MetricQueryRequest)
}

func calcComparisonTime(t time.Time, granlarCfg interfaces.SameperiodConfig) time.Time {
	switch granlarCfg.TimeGranularity {
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_DAY:
		return t.AddDate(0, 0, -granlarCfg.Offset)
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_MONTH:
		// 月环比, k个月同期
		newTime := t.AddDate(0, -granlarCfg.Offset, 0)
		// 处理月末日期不存在的情况
		if t.Day() != newTime.Day() {
			newTime = common.LastDayOfMonth(newTime)
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(),
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_QUARTER:
		// 上k个季度
		newTime := t.AddDate(0, -3*granlarCfg.Offset, 0)

		// 处理季度末日期不存在的情况
		if t.Day() != newTime.Day() {
			newTime = common.LastDayOfMonth(newTime)
			newTime = time.Date(newTime.Year(), newTime.Month(), newTime.Day(),
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	case interfaces.METRICS_SAMEPERIOD_TIME_GRANULARITY_YEAR:
		// 上k年
		newTime := t.AddDate(-granlarCfg.Offset, 0, 0)

		// 处理闰年2月29日的情况
		if t.Month() == time.February && t.Day() == 29 && !common.IsLeap(newTime.Year()) {
			newTime = time.Date(newTime.Year(), time.February, 28,
				t.Hour(), t.Minute(), t.Second(), t.Nanosecond()/1e6, newTime.Location())
		}
		return newTime
	}

	return t
}

// convertVegaDatas2TimeSeries maps resource trend entries to time series (Vega: __time, __value, dimensions;
// resource: timeResField, __value, and group-by resource columns).
func convertVegaDatas2TimeSeries(ctx context.Context, def interfaces.MetricDefinition,
	datas, samePeriodDatas *interfaces.DatasetQueryResponse, query *interfaces.MetricQueryRequest,
	trend *trendMeta, vegaDuration int64, propMap map[string]*cond.DataProperty) (interfaces.MetricResponse, error) {

	resp := interfaces.MetricResponse{
		VegaDurationMs: vegaDuration,
	}
	if query == nil {
		return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("query is required for trend")
	}
	if trend == nil {
		return resp, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
			WithErrorDetails("trend metadata missing")
	}

	currentSeriesMap, err := convert2TimeSeries(ctx, def, datas, query, trend, propMap, false)
	if err != nil {
		return resp, err
	}

	if query.Metrics == nil {
		out := make([]interfaces.BknMetricData, 0, len(currentSeriesMap))
		for _, ts := range currentSeriesMap {
			out = append(out, ts)
		}
		resp.Datas = out
		return resp, nil
	}

	switch query.Metrics.Type {
	case interfaces.METRICS_SAMEPERIOD:
		if samePeriodDatas == nil {
			return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("sameperiod trend requires same-period dataset")
		}
		if query.Metrics.SameperiodConfig == nil {
			return resp, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("sameperiod_config is required")
		}
		previousMap, err := convert2TimeSeries(ctx, def, samePeriodDatas, query, trend, propMap, true)
		if err != nil {
			return resp, err
		}
		slice, err := calcSamePeriodValueBkn(ctx, currentSeriesMap, previousMap, query.Metrics, query)
		if err != nil {
			return resp, err
		}
		resp.Datas = slice
		return resp, nil
	case interfaces.METRICS_PROPORTION:
		slice, err := calcProportionValueBkn(ctx, currentSeriesMap)
		if err != nil {
			return resp, err
		}
		resp.Datas = slice
		return resp, nil
	}

	return resp, nil
}

func findTimeStrIndex(timePoints []string, timeStr string) int {
	for i, t := range timePoints {
		if t == timeStr {
			return i
		}
	}
	return -1
}

// isSamePeriod affects only the fill_null time grid: shifts [start,end] to the comparison window (mdl convert2TimeSeries).
func convert2TimeSeries(ctx context.Context, def interfaces.MetricDefinition, datas *interfaces.DatasetQueryResponse,
	query *interfaces.MetricQueryRequest, trend *trendMeta, propMap map[string]*cond.DataProperty, isSamePeriod bool) (map[string]interfaces.BknMetricData, error) {

	seriesMap := make(map[string]interfaces.BknMetricData)
	if datas == nil {
		return seriesMap, nil
	}
	fillNull := query != nil && query.FillNull
	groupDims, err := metricGroupByDimensions(&def, query, propMap)
	if err != nil {
		return nil, err
	}
	valueField := interfaces.VALUE_FIELD
	timeResField := trend.timeResField

	var allTimes []any
	var allTimeStrs []string
	if fillNull {
		if query.Time == nil || query.Time.Start == nil || query.Time.End == nil || query.Time.Step == nil {
			return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
				WithErrorDetails("fill_null requires time.start, time.end, and time.step")
		}
		loc := common.AppLocationOrUTC()
		fixedStart, fixedEnd := correctingTime(query, loc)
		if isSamePeriod {
			if query.Metrics == nil || query.Metrics.SameperiodConfig == nil {
				return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
					WithErrorDetails("same-period fill_null query missing metrics.sameperiod_config")
			}
			cfg := *query.Metrics.SameperiodConfig
			fixedStart = calcComparisonTime(time.UnixMilli(fixedStart).In(loc), cfg).UnixMilli()
			fixedEnd = calcComparisonTime(time.UnixMilli(fixedEnd).In(loc), cfg).UnixMilli()
		}
		step := *query.Time.Step
		for currentTime := fixedStart; currentTime <= fixedEnd; {
			allTimes = append(allTimes, currentTime)
			allTimeStrs = append(allTimeStrs, common.FormatTimeMiliis(currentTime, step))
			currentTime = getNextPointTime(query, currentTime)
		}
	}

	for _, entry := range datas.Entries {
		if entry == nil {
			continue
		}
		key := buildEntryDimKey(entry, groupDims)
		labels := make(map[string]string, len(groupDims))
		for _, d := range groupDims {
			labels[d.PropertyName] = fmt.Sprintf("%v", entry[d.ResourceFieldName])
		}

		timeRaw, hasTime := entry[timeResField]
		if !hasTime {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
				WithErrorDetails("missing time bucket field in resource entry")
		}
		timei, err := entryTimeToMillis(timeRaw)
		if err != nil {
			return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
				WithErrorDetails(fmt.Sprintf("time field: %v", err))
		}
		var timeStr string
		if query.Time != nil && query.Time.Step != nil {
			timeStr = common.FormatTimeMiliis(timei, *query.Time.Step)
		} else {
			timeStr = common.FormatRFC3339Milli(timei)
		}

		ts, exists := seriesMap[key]
		if !exists {
			if fillNull {
				ts = interfaces.BknMetricData{
					Labels:   labels,
					Times:    allTimes,
					TimeStrs: allTimeStrs,
					Values:   make([]any, len(allTimeStrs)),
				}
				// 趋势对齐时间轴上无数据的桶：用 nil（JSON null），不补 0
				for i := range ts.Values {
					ts.Values[i] = nil
				}
			} else {
				ts = interfaces.BknMetricData{
					Labels:   labels,
					Times:    make([]any, 0),
					TimeStrs: make([]string, 0),
					Values:   make([]any, 0),
				}
			}
		}

		if fillNull {
			idx := findTimeStrIndex(allTimeStrs, timeStr)
			if idx == -1 {
				seriesMap[key] = ts
				continue
			}
			if v, vok := entry[valueField]; vok && v != nil {
				f, ferr := common.AnyToFloat64(v)
				if ferr != nil {
					return nil, rest.NewHTTPError(ctx, http.StatusInternalServerError, oerrors.OntologyQuery_Metric_InternalError_QueryFailed).
						WithErrorDetails(ferr.Error())
				}
				ts.Values[idx] = f
			}
		} else {
			ts.Times = append(ts.Times, timei)
			ts.TimeStrs = append(ts.TimeStrs, timeStr)
			if v, vok := entry[valueField]; vok {
				_, err = appendMetricValue(ctx, v, &ts.Values)
			} else {
				_, err = appendMetricValue(ctx, nil, &ts.Values)
			}
			if err != nil {
				return nil, err
			}
		}
		seriesMap[key] = ts
	}

	return seriesMap, nil
}

func toFloat64ForMetricValue(ctx context.Context, v any) (float64, error) {
	if v == nil {
		return 0, nil
	}
	return common.AnyToFloat64(v)
}

func toMillisAny(v any) (int64, error) {
	f, err := common.AnyToFloat64(v)
	if err != nil {
		return 0, err
	}
	return int64(f), nil
}

// lookupSamePeriodBaseValue 在对比期序列中找「同期」桶：与 convert2TimeSeries 一致，优先用日历 timeStr 对齐，避免仅用毫秒相等错配到其它分桶。
func lookupSamePeriodBaseValue(prev interfaces.BknMetricData, compareDateMs int64, step string) (any, bool) {
	step = strings.TrimSpace(step)
	if len(prev.Times) == 0 {
		return nil, false
	}
	if step != "" && len(prev.TimeStrs) > 0 {
		want := common.FormatTimeMiliis(compareDateMs, step)
		for j := range prev.TimeStrs {
			if j >= len(prev.Values) {
				continue
			}
			if prev.TimeStrs[j] == want {
				return prev.Values[j], true
			}
		}
	}
	for j := range prev.Times {
		if j >= len(prev.Values) {
			continue
		}
		ptm, perr := toMillisAny(prev.Times[j])
		if perr != nil {
			continue
		}
		if ptm == compareDateMs {
			return prev.Values[j], true
		}
	}
	return nil, false
}

// calcSamePeriodValueBkn aligns mdl calcSamePeriodValue for BknMetricData and resource time buckets.
// 增长值=本期数-同期数：同期取对比查询结果中与本期按 sameperiod_config 映射到同一日历桶的值（非本序列上一索引）。
func calcSamePeriodValueBkn(ctx context.Context, currentSeriesMap, previousMap map[string]interfaces.BknMetricData,
	metrics *interfaces.Metrics, query *interfaces.MetricQueryRequest) ([]interfaces.BknMetricData, error) {

	if metrics == nil || metrics.SameperiodConfig == nil {
		return nil, rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_Metric_InvalidParameter).
			WithErrorDetails("sameperiod_config is required")
	}
	cfg := metrics.SameperiodConfig
	var hasGrowthValue, hasGrowthRate bool
	for _, m := range cfg.Method {
		if m == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_VALUE {
			hasGrowthValue = true
		}
		if m == interfaces.METRICS_SAMEPERIOD_METHOD_GROWTH_RATE {
			hasGrowthRate = true
		}
	}
	step := ""
	if query != nil && query.Time != nil && query.Time.Step != nil {
		step = *query.Time.Step
	}
	out := make([]interfaces.BknMetricData, 0, len(currentSeriesMap))
	for key, currentPoints := range currentSeriesMap {
		previousPoints := previousMap[key]
		ts := interfaces.BknMetricData{
			Labels:       currentPoints.Labels,
			Times:        make([]any, 0, len(currentPoints.Times)),
			TimeStrs:     make([]string, 0, len(currentPoints.TimeStrs)),
			Values:       make([]any, 0, len(currentPoints.Values)),
			GrowthValues: make([]any, 0),
			GrowthRates:  make([]any, 0),
		}
		for i := range currentPoints.Times {
			ts.Times = append(ts.Times, currentPoints.Times[i])
			if i < len(currentPoints.TimeStrs) {
				ts.TimeStrs = append(ts.TimeStrs, currentPoints.TimeStrs[i])
			}
			if i < len(currentPoints.Values) {
				ts.Values = append(ts.Values, currentPoints.Values[i])
			}
			curT, err := toMillisAny(currentPoints.Times[i])
			if err != nil {
				return nil, err
			}
			compareDate := calcComparisonTime(time.UnixMilli(curT).In(common.AppLocationOrUTC()), *cfg).UnixMilli()
			previousV, found := lookupSamePeriodBaseValue(previousPoints, compareDate, step)
			if !found {
				previousV = nil
			}
			if previousV != nil && i < len(currentPoints.Values) && currentPoints.Values[i] != nil {
				cur, err1 := toFloat64ForMetricValue(ctx, currentPoints.Values[i])
				if err1 != nil {
					return nil, err1
				}
				prev, err2 := toFloat64ForMetricValue(ctx, previousV)
				if err2 != nil {
					return nil, err2
				}
				if hasGrowthValue {
					ts.GrowthValues = append(ts.GrowthValues, cur-prev)
				}
				if hasGrowthRate {
					if prev != 0 {
						ts.GrowthRates = append(ts.GrowthRates, (cur-prev)/prev*100)
					} else {
						ts.GrowthRates = append(ts.GrowthRates, nil)
					}
				}
			} else {
				if hasGrowthValue {
					ts.GrowthValues = append(ts.GrowthValues, nil)
				}
				if hasGrowthRate {
					ts.GrowthRates = append(ts.GrowthRates, nil)
				}
			}
		}
		out = append(out, ts)
	}
	return out, nil
}

func getSeriesProportionTotalBkn(ctx context.Context, seriesMap map[string]interfaces.BknMetricData) (map[string]float64, error) {
	totals := make(map[string]float64)
	for _, series := range seriesMap {
		for i, timeStr := range series.TimeStrs {
			if i >= len(series.Values) {
				continue
			}
			v, err := toFloat64ForMetricValue(ctx, series.Values[i])
			if err != nil {
				return nil, err
			}
			totals[timeStr] += v
		}
	}
	return totals, nil
}

func calcProportionValueBkn(ctx context.Context, currentSeriesMap map[string]interfaces.BknMetricData) ([]interfaces.BknMetricData, error) {
	timeTotals, err := getSeriesProportionTotalBkn(ctx, currentSeriesMap)
	if err != nil {
		return nil, err
	}
	datas := make([]interfaces.BknMetricData, 0, len(currentSeriesMap))
	for _, ts := range currentSeriesMap {
		nts := ts
		nts.Proportions = make([]any, 0, len(ts.TimeStrs))
		for i, timeStr := range ts.TimeStrs {
			if i >= len(ts.Values) {
				nts.Proportions = append(nts.Proportions, nil)
				continue
			}
			if total, ok := timeTotals[timeStr]; ok && total != 0 {
				val, err := toFloat64ForMetricValue(ctx, ts.Values[i])
				if err != nil {
					return nil, err
				}
				nts.Proportions = append(nts.Proportions, val/total*100)
			} else {
				nts.Proportions = append(nts.Proportions, nil)
			}
		}
		datas = append(datas, nts)
	}
	return datas, nil
}

// normalizeMetricCalendarStep 将 day/week/... 及常见别名归一为日历粒度名；非日历步长返回 "", false。
// 注意：趋势查询合法 step 为 day|week|month|quarter|year（与 validate 一致），不可走 ParseDuration。
func normalizeMetricCalendarStep(raw string) (name string, ok bool) {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch s {
	case "day", "1d":
		return "day", true
	case "week", "1w":
		return "week", true
	case "month", "1M":
		return "month", true
	case "quarter", "1q":
		return "quarter", true
	case "year", "1y":
		return "year", true
	default:
		return "", false
	}
}

// alignCalendarRangeMillis 与 instant 趋势分桶对齐一致，供 fill_null 时间轴与下推 group_by 共用。
func alignCalendarRangeMillis(startTime, endTime time.Time, cal string, zone *time.Location) (int64, int64) {
	switch cal {
	case "day":
		year, month, day := startTime.Date()
		fixStart := time.Date(year, month, day, 0, 0, 0, 0, zone)
		year, month, day = endTime.Date()
		fixEnd := time.Date(year, month, day, 0, 0, 0, 0, zone)
		return fixStart.UnixMilli(), fixEnd.UnixMilli()
	case "week":
		year, month, day := startTime.Date()
		fixStart := time.Date(year, month, day, 0, 0, 0, 0, zone)
		year, month, day = endTime.Date()
		fixEnd := time.Date(year, month, day, 0, 0, 0, 0, zone)
		startDay := int(fixStart.Weekday())
		endDay := int(fixEnd.Weekday())
		fixStart = fixStart.AddDate(0, 0, -(7+startDay-1)%7)
		fixEnd = fixEnd.AddDate(0, 0, -(7+endDay-1)%7)
		return fixStart.UnixMilli(), fixEnd.UnixMilli()
	case "month":
		fixStart := time.Date(startTime.Year(), startTime.Month(), 1, 0, 0, 0, 0, zone)
		fixEnd := time.Date(endTime.Year(), endTime.Month(), 1, 0, 0, 0, 0, zone)
		return fixStart.UnixMilli(), fixEnd.UnixMilli()
	case "quarter":
		startQuarter := (int(startTime.Month()) - 1) / 3
		endQuarter := (int(endTime.Month()) - 1) / 3
		stMonth := time.Month(startQuarter*3 + 1)
		enMonth := time.Month(endQuarter*3 + 1)
		st := time.Date(startTime.Year(), stMonth, 1, 0, 0, 0, 0, zone)
		en := time.Date(endTime.Year(), enMonth, 1, 0, 0, 0, 0, zone)
		fixStart := time.Date(st.Year(), st.Month(), 1, 0, 0, 0, 0, zone)
		fixEnd := time.Date(en.Year(), en.Month(), 1, 0, 0, 0, 0, zone)
		return fixStart.UnixMilli(), fixEnd.UnixMilli()
	case "year":
		fixStart := time.Date(startTime.Year(), time.January, 1, 0, 0, 0, 0, zone)
		fixEnd := time.Date(endTime.Year(), time.January, 1, 0, 0, 0, 0, zone)
		return fixStart.UnixMilli(), fixEnd.UnixMilli()
	default:
		return 0, 0
	}
}

// correctingTime 修正开始时间和结束时间，符合opensearch的分桶区间
func correctingTime(query *interfaces.MetricQueryRequest, zoneLocation *time.Location) (int64, int64) {
	if query == nil || query.Time == nil || query.Time.Step == nil {
		return 0, 0
	}
	startTime := time.UnixMilli(*query.Time.Start)
	endTime := time.UnixMilli(*query.Time.End)

	// 日历步长：趋势（instant=false）也必须走此分支。误用 ParseDuration("day") 得 step=0 会产生错误时间轴。
	if cal, ok := normalizeMetricCalendarStep(*query.Time.Step); ok {
		return alignCalendarRangeMillis(startTime, endTime, cal, zoneLocation)
	}

	if query.Time.Instant != nil && *query.Time.Instant {
		switch *query.Time.Step {
		case "minute", "1m":
			fixStart := startTime.Truncate(time.Minute)
			fixEnd := endTime.Truncate(time.Minute)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()
		case "hour", "1h":
			fixStart := startTime.Truncate(time.Hour)
			fixEnd := endTime.Truncate(time.Hour)
			return fixStart.UnixMilli(), fixEnd.UnixMilli()
		}
	}
	stepStr := strings.TrimSpace(*query.Time.Step)
	stepT, err := common.ParseDuration(stepStr)
	if err != nil {
		return 0, 0
	}
	step := stepT.Milliseconds()
	if step <= 0 {
		return 0, 0
	}
	_, offset := startTime.In(zoneLocation).Zone()
	fixedStart := int64(math.Floor(float64(*query.Time.Start+int64(offset*1000))/float64(step)))*step - int64(offset*1000)
	fixedEnd := int64(math.Floor(float64(*query.Time.End+int64(offset*1000))/float64(step)))*step - int64(offset*1000)
	return fixedStart, fixedEnd
}

// getNextPointTime 获取下一个时间点
func getNextPointTime(query *interfaces.MetricQueryRequest, currentTime int64) int64 {

	// 将时间戳转换为时间对象
	switch *query.Time.Step {
	case "minute", "1m":
		return currentTime + time.Minute.Milliseconds()
	case "hour", "1h":
		return currentTime + time.Hour.Milliseconds()
	case "day", "1d":
		return currentTime + (time.Hour * 24).Milliseconds()
	case "week", "1w":
		return currentTime + (time.Hour * 24 * 7).Milliseconds()
	case "month", "1M":
		t := time.UnixMilli(currentTime)
		return t.AddDate(0, 1, 0).UnixMilli()
	case "quarter", "1q":
		t := time.UnixMilli(currentTime)
		return t.AddDate(0, 3, 0).UnixMilli()
	case "year", "1y":
		t := time.UnixMilli(currentTime)
		return t.AddDate(1, 0, 0).UnixMilli()
	default:
		stepT, _ := common.ParseDuration(*query.Time.Step)
		return currentTime + stepT.Milliseconds()
	}
}
