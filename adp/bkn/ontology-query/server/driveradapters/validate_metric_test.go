// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"
	. "github.com/smartystreets/goconvey/convey"

	cond "ontology-query/common/condition"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

func Test_validateMetricQueryRequest_timeAndStep(t *testing.T) {
	ctx := context.Background()

	Convey("time.start/end must appear together\n", t, func() {
		s := int64(1)
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Start: &s},
		}
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldNotBeNil)
		httpErr := err.(*rest.HTTPError)
		So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_Metric_InvalidParameter)
	})

	Convey("time.start must be <= time.end\n", t, func() {
		a, b := int64(10), int64(5)
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Start: &a, End: &b},
		}
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldNotBeNil)
		httpErr := err.(*rest.HTTPError)
		So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
	})

	Convey("trend requires step\n", t, func() {
		instant := false
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Instant: &instant},
		}
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldNotBeNil)
	})

	Convey("trend step must be calendar interval\n", t, func() {
		instant := false
		step := "1h"
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Instant: &instant, Step: &step},
		}
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldNotBeNil)
		httpErr := err.(*rest.HTTPError)
		So(httpErr.BaseError.ErrorCode, ShouldEqual, oerrors.OntologyQuery_Metric_InvalidParameter)
	})

	Convey("instant query does not require step\n", t, func() {
		instant := true
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Instant: &instant},
		}
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldBeNil)
	})

	Convey("fill_null rejects instant query\n", t, func() {
		instant := true
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Instant: &instant},
		}
		body.FillNull = true
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldNotBeNil)
	})

	Convey("fill_null allows trend with time range and step\n", t, func() {
		instant := false
		step := "day"
		a, b := int64(1), int64(2)
		body := &interfaces.MetricQueryRequest{
			Time: &interfaces.MetricTimeWindow{Start: &a, End: &b, Instant: &instant, Step: &step},
		}
		body.FillNull = true
		err := validateMetricQueryRequest(ctx, body)
		So(err, ShouldBeNil)
	})
}

func Test_validateMetricDryRunForExecution_alignedWithBknSave(t *testing.T) {
	ctx := context.Background()

	Convey("dry-run rejects invalid aggregation op\n", t, func() {
		body := &interfaces.MetricDryRunRequest{
			MetricConfig: &interfaces.MetricDefinition{
				MetricType: interfaces.MetricTypeAtomic,
				ScopeType:  interfaces.ScopeTypeObjectType,
				ScopeRef:   "ot1",
				UnitType:   "numUnit",
				Unit:       "none",
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Aggregation: interfaces.MetricAggregation{Property: "p", Aggr: "bogus"},
				},
			},
		}
		err := validateMetricDryRunForExecution(ctx, body)
		So(err, ShouldNotBeNil)
	})

	Convey("dry-run accepts minimal valid config\n", t, func() {
		body := &interfaces.MetricDryRunRequest{
			MetricConfig: &interfaces.MetricDefinition{
				MetricType: interfaces.MetricTypeAtomic,
				ScopeType:  interfaces.ScopeTypeObjectType,
				ScopeRef:   "ot1",
				UnitType:   "numUnit",
				Unit:       "none",
				CalculationFormula: &interfaces.MetricCalculationFormula{
					Condition:   &cond.CondCfg{Operation: cond.OperationEq, Name: "f1", ValueOptCfg: cond.ValueOptCfg{Value: 1}},
					Aggregation: interfaces.MetricAggregation{Property: "amount", Aggr: interfaces.MetricAggrSum},
				},
			},
		}
		err := validateMetricDryRunForExecution(ctx, body)
		So(err, ShouldBeNil)
	})
}
