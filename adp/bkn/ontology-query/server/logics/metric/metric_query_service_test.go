// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package metric

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"
	"go.uber.org/mock/gomock"

	"ontology-query/common"
	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
	dtype "ontology-query/interfaces/data_type"
	omock "ontology-query/interfaces/mock"
)

func Test_metricQueryService_QueryMetricData(t *testing.T) {
	Convey("QueryMetricData\n", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		oma := omock.NewMockOntologyManagerAccess(ctrl)
		vba := omock.NewMockVegaBackendAccess(ctrl)

		svc := &metricQueryService{
			appSetting: &common.AppSetting{},
			oma:        oma,
			vba:        vba,
		}

		def := &interfaces.MetricDefinition{
			ID:        "m1",
			KnID:      "kn1",
			ScopeType: interfaces.ScopeTypeObjectType,
			ScopeRef:  "ot1",
			UnitType:  "numUnit",
			Unit:      "none",
			CalculationFormula: &interfaces.MetricCalculationFormula{
				Condition:   &cond.CondCfg{Operation: cond.OperationEq, Name: "f1", ValueOptCfg: cond.ValueOptCfg{Value: 1}},
				Aggregation: interfaces.MetricAggregation{Property: "amount", Aggr: "sum"},
			},
		}

		Convey("Success\n", func() {
			oma.EXPECT().GetMetricDefinition(gomock.Any(), "kn1", "main", "m1").Return(def, true, nil)
			oma.EXPECT().GetObjectType(gomock.Any(), "kn1", "main", "ot1").Return(interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res1",
					},
					DataProperties: []cond.DataProperty{
						{Name: "f1", Type: dtype.DATATYPE_STRING, MappedField: cond.Field{Name: "f1_res"}},
						{Name: "amount", Type: dtype.DATATYPE_DOUBLE, MappedField: cond.Field{Name: "amount_res"}},
					},
				},
			}, true, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), "res1", gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries: []map[string]any{{"__value": 42.0}},
			}, nil)

			out, err := svc.QueryMetricData(ctx, "kn1", "main", "m1", &interfaces.MetricQueryRequest{})
			So(err, ShouldBeNil)
			So(len(out.Datas), ShouldEqual, 1)
			So(out.Datas[0].Values[0], ShouldEqual, 42.0)
			So(out.Model.UnitType, ShouldEqual, "numUnit")
		})

		Convey("Trend query uses calendar step, time group_by, time sort, and uniquery-style time block\n", func() {
			defTrend := &interfaces.MetricDefinition{
				ID:        "m1",
				KnID:      "kn1",
				ScopeType: interfaces.ScopeTypeObjectType,
				ScopeRef:  "ot1",
				UnitType:  "numUnit",
				Unit:      "none",
				TimeDimension: &interfaces.MetricTimeDimension{
					Property: "evt_time",
				},
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Condition:   &cond.CondCfg{Operation: cond.OperationEq, Name: "f1", ValueOptCfg: cond.ValueOptCfg{Value: 1}},
					Aggregation: interfaces.MetricAggregation{Property: "amount", Aggr: "sum"},
				},
			}
			instant := false
			step := "month"
			start := int64(1_000)
			end := int64(2_000)
			var captured *interfaces.ResourceDataQueryParams
			oma.EXPECT().GetMetricDefinition(gomock.Any(), "kn1", "main", "m1").Return(defTrend, true, nil)
			oma.EXPECT().GetObjectType(gomock.Any(), "kn1", "main", "ot1").Return(interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res1",
					},
					DataProperties: []cond.DataProperty{
						{Name: "f1", Type: dtype.DATATYPE_STRING, MappedField: cond.Field{Name: "f1_res"}},
						{Name: "amount", Type: dtype.DATATYPE_DOUBLE, MappedField: cond.Field{Name: "amount_res"}},
						{Name: "evt_time", Type: dtype.DATATYPE_DATETIME, MappedField: cond.Field{Name: "evt_time_res"}},
					},
				},
			}, true, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), "res1", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, p *interfaces.ResourceDataQueryParams) (*interfaces.DatasetQueryResponse, error) {
					captured = p
					return &interfaces.DatasetQueryResponse{
						Entries: []map[string]any{{"__value": 3.0, "evt_time_res": 1700000000000.0}},
					}, nil
				},
			)

			out, err := svc.QueryMetricData(ctx, "kn1", "main", "m1", &interfaces.MetricQueryRequest{
				Time: &interfaces.MetricTimeWindow{
					Start: &start, End: &end, Instant: &instant, Step: &step,
				},
			})
			So(err, ShouldBeNil)
			So(captured, ShouldNotBeNil)
			So(captured.GroupBy[0]["property"], ShouldEqual, "evt_time_res")
			So(captured.GroupBy[0]["calendar_interval"], ShouldEqual, "month")
			So(out.Step, ShouldEqual, "month")
			So(out.IsCalendar, ShouldBeTrue)
			So(out.Datas[0].Times[0], ShouldEqual, 1700000000000.0)
		})

		Convey("Trend query resolves time range from default_range_policy when start/end omitted\n", func() {
			defTrend := &interfaces.MetricDefinition{
				ID:        "m1",
				KnID:      "kn1",
				ScopeType: interfaces.ScopeTypeObjectType,
				ScopeRef:  "ot1",
				UnitType:  "numUnit",
				Unit:      "none",
				TimeDimension: &interfaces.MetricTimeDimension{
					Property:           "evt_time",
					DefaultRangePolicy: interfaces.MetricTimeDefaultRangePolicyLast1h,
				},
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Aggregation: interfaces.MetricAggregation{Property: "amount", Aggr: "sum"},
				},
			}
			instant := false
			step := "day"
			var captured *interfaces.ResourceDataQueryParams
			oma.EXPECT().GetMetricDefinition(gomock.Any(), "kn1", "main", "m1").Return(defTrend, true, nil)
			oma.EXPECT().GetObjectType(gomock.Any(), "kn1", "main", "ot1").Return(interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res1",
					},
					DataProperties: []cond.DataProperty{
						{Name: "amount", Type: dtype.DATATYPE_DOUBLE, MappedField: cond.Field{Name: "amount_res"}},
						{Name: "evt_time", Type: dtype.DATATYPE_DATETIME, MappedField: cond.Field{Name: "evt_time_res"}},
					},
				},
			}, true, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), "res1", gomock.Any()).DoAndReturn(
				func(_ context.Context, _ string, p *interfaces.ResourceDataQueryParams) (*interfaces.DatasetQueryResponse, error) {
					captured = p
					return &interfaces.DatasetQueryResponse{
						Entries: []map[string]any{{"__value": 1.0, "evt_time_res": 1.0}},
					}, nil
				},
			)

			_, err := svc.QueryMetricData(ctx, "kn1", "main", "m1", &interfaces.MetricQueryRequest{
				Time: &interfaces.MetricTimeWindow{
					Instant: &instant,
					Step:    &step,
				},
			})
			So(err, ShouldBeNil)
			So(captured, ShouldNotBeNil)
			So(captured.GroupBy[0]["property"], ShouldEqual, "evt_time_res")
			So(captured.GroupBy[0]["calendar_interval"], ShouldEqual, "day")
		})

		Convey("Not found when bkn returns nil definition\n", func() {
			oma.EXPECT().GetMetricDefinition(gomock.Any(), "kn1", "main", "m1").Return(nil, false, nil)
			_, err := svc.QueryMetricData(ctx, "kn1", "main", "m1", &interfaces.MetricQueryRequest{})
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_Metric_NotFound)
		})
	})
}

// 趋势 + 日历 day 曾误走 ParseDuration("day")→step=0，产生 -28800000ms 等错误时间轴；fill_null 须走日历对齐。
func Test_correctingTime_trendCalendarDay(t *testing.T) {
	Convey("correctingTime uses calendar path for trend day (not ParseDuration)\n", t, func() {
		instant := false
		step := "day"
		start := int64(1_776_729_600_000)
		end := int64(1_776_816_000_000)
		q := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Start: &start, End: &end, Instant: &instant, Step: &step},
		}
		fs, fe := correctingTime(q, time.UTC)
		So(fs, ShouldNotEqual, -28800000)
		So(fe >= fs, ShouldBeTrue)
	})
}

// 同环比「同期」须与 convert2TimeSeries 的日历分桶键一致；仅靠毫秒相等会错配桶，导致增长值=本期-错误的同期。
func Test_lookupSamePeriodBaseValue_timeStrAlignment(t *testing.T) {
	Convey("lookupSamePeriodBaseValue matches calendar TimeStr when millis differ\n", t, func() {
		step := "day"
		msTarget := int64(1_776_787_200_000)
		key := common.FormatTimeMiliis(msTarget, step)
		prev := interfaces.BknMetricData{
			Times:    []any{int64(0), int64(999)}, // 与分桶毫秒不一致，仅靠 ptm==compareDate 会对不上
			TimeStrs: []string{"1970-01-01", key},
			Values:   []any{0.0, 2.0},
		}
		v, ok := lookupSamePeriodBaseValue(prev, msTarget, step)
		So(ok, ShouldBeTrue)
		So(v, ShouldEqual, 2.0)
	})
}

func Test_metricQueryService_DryRunMetricData(t *testing.T) {
	Convey("DryRunMetricData\n", t, func() {
		ctx := context.Background()
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		oma := omock.NewMockOntologyManagerAccess(ctrl)
		vba := omock.NewMockVegaBackendAccess(ctrl)

		svc := &metricQueryService{
			appSetting: &common.AppSetting{},
			oma:        oma,
			vba:        vba,
		}

		Convey("Fails when kn_id mismatches metric_config.kn_id\n", func() {
			body := &interfaces.MetricDryRunRequest{
				MetricConfig: &interfaces.MetricDefinition{
					ID:         "x",
					KnID:       "other",
					MetricType: "atomic",
					ScopeType:  interfaces.ScopeTypeObjectType,
					ScopeRef:   "ot1",
					CalculationFormula: &interfaces.MetricCalculationFormula{
						Aggregation: interfaces.MetricAggregation{Property: "p", Aggr: "sum"},
					},
				},
			}
			_, err := svc.DryRunMetricData(ctx, "kn1", "main", body)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Success without persisting metric_id\n", func() {
			body := &interfaces.MetricDryRunRequest{
				MetricConfig: &interfaces.MetricDefinition{
					ID:         "tmp",
					KnID:       "kn1",
					MetricType: "atomic",
					ScopeType:  interfaces.ScopeTypeObjectType,
					ScopeRef:   "ot1",
					UnitType:   "numUnit",
					Unit:       "none",
					CalculationFormula: &interfaces.MetricCalculationFormula{
						Aggregation: interfaces.MetricAggregation{Property: "amount", Aggr: "sum"},
					},
				},
			}
			oma.EXPECT().GetObjectType(gomock.Any(), "kn1", "main", "ot1").Return(interfaces.ObjectType{
				ObjectTypeWithKeyField: interfaces.ObjectTypeWithKeyField{
					OTID: "ot1",
					DataSource: &interfaces.ResourceInfo{
						Type: interfaces.DATA_SOURCE_TYPE_RESOURCE,
						ID:   "res1",
					},
					DataProperties: []cond.DataProperty{
						{Name: "amount", Type: dtype.DATATYPE_DOUBLE, MappedField: cond.Field{Name: "amount_res"}},
					},
				},
			}, true, nil)
			vba.EXPECT().QueryResourceData(gomock.Any(), "res1", gomock.Any()).Return(&interfaces.DatasetQueryResponse{
				Entries: []map[string]any{{"__value": 1.0}},
			}, nil)

			out, err := svc.DryRunMetricData(ctx, "kn1", "main", body)
			So(err, ShouldBeNil)
			So(out.Datas[0].Values[0], ShouldEqual, 1.0)
		})
	})
}
