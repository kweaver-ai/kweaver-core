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

	"bkn-backend/interfaces"
)

func validStrictCreateMetric() *interfaces.MetricDefinition {
	return &interfaces.MetricDefinition{
		Name:       "m1",
		MetricType: interfaces.MetricTypeAtomic,
		UnitType:   "numUnit",
		Unit:       "none",
		ScopeType:  interfaces.ScopeTypeObjectType,
		ScopeRef:   "ot1",
		CalculationFormula: &interfaces.MetricCalculationFormula{
			Aggregation: interfaces.MetricAggregation{Property: "p", Aggr: interfaces.MetricAggrSum},
		},
	}
}

func Test_ValidateMetricRequest(t *testing.T) {
	Convey("Test ValidateMetricRequest\n", t, func() {
		ctx := context.Background()

		Convey("Failed with subgraph scope in strict mode\n", func() {
			r := validStrictCreateMetric()
			r.ScopeType = interfaces.ScopeTypeSubgraph
			err := ValidateMetricRequest(ctx, r, true)
			So(err, ShouldNotBeNil)
			httpErr := err.(*rest.HTTPError)
			So(httpErr.HTTPCode, ShouldEqual, http.StatusBadRequest)
		})

		Convey("Failed with empty scope_ref in strict mode\n", func() {
			r := validStrictCreateMetric()
			r.ScopeRef = "   "
			err := ValidateMetricRequest(ctx, r, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with non-atomic metric_type in strict mode\n", func() {
			r := validStrictCreateMetric()
			r.MetricType = interfaces.MetricTypeDerived
			err := ValidateMetricRequest(ctx, r, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Failed with missing aggregation property in strict mode\n", func() {
			r := validStrictCreateMetric()
			r.CalculationFormula.Aggregation.Property = ""
			err := ValidateMetricRequest(ctx, r, true)
			So(err, ShouldNotBeNil)
		})

		Convey("Success with valid payload in strict mode\n", func() {
			r := validStrictCreateMetric()
			err := ValidateMetricRequest(ctx, r, true)
			So(err, ShouldBeNil)
		})

		Convey("Success with minimal fields in non-strict mode\n", func() {
			r := validStrictCreateMetric()
			r.ScopeType = ""
			r.ScopeRef = ""
			r.UnitType = ""
			r.Unit = ""
			r.MetricType = ""
			// 非 strict：formula 可省略；若提供 calculation_formula，aggregation.property 与 aggr 仍须齐全。
			r.CalculationFormula = nil
			err := ValidateMetricRequest(ctx, r, false)
			So(err, ShouldBeNil)
		})
	})
}

func Test_ValidateMetricRequests(t *testing.T) {
	Convey("Test ValidateMetricRequests\n", t, func() {
		ctx := context.Background()

		Convey("Failed with duplicate metric name in batch\n", func() {
			e := validStrictCreateMetric()
			e.Name = "dup"
			err := ValidateMetricRequests(ctx, []*interfaces.MetricDefinition{e, e}, true, interfaces.ImportMode_Normal)
			So(err, ShouldNotBeNil)
		})
	})
}
