// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/kweaver-ai/kweaver-go-lib/rest"

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

func TestValidateMetricRequest(t *testing.T) {
	ctx := context.Background()
	t.Run("subgraph scope rejected", func(t *testing.T) {
		r := validStrictCreateMetric()
		r.ScopeType = interfaces.ScopeTypeSubgraph
		err := ValidateMetricRequest(ctx, r, true)
		if err == nil {
			t.Fatal("expected error")
		}
		var he *rest.HTTPError
		if !errors.As(err, &he) || he.HTTPCode != http.StatusBadRequest {
			t.Fatalf("got %v", err)
		}
	})
	t.Run("empty scope_ref strict", func(t *testing.T) {
		r := validStrictCreateMetric()
		r.ScopeRef = "   "
		err := ValidateMetricRequest(ctx, r, true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("strict non atomic metric_type", func(t *testing.T) {
		r := validStrictCreateMetric()
		r.MetricType = interfaces.MetricTypeDerived
		err := ValidateMetricRequest(ctx, r, true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("missing aggregation", func(t *testing.T) {
		r := validStrictCreateMetric()
		r.CalculationFormula.Aggregation.Property = ""
		err := ValidateMetricRequest(ctx, r, true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ok strict", func(t *testing.T) {
		r := validStrictCreateMetric()
		if err := ValidateMetricRequest(ctx, r, true); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("non strict empty scope allowed", func(t *testing.T) {
		r := validStrictCreateMetric()
		r.ScopeType = ""
		r.ScopeRef = ""
		r.UnitType = ""
		r.Unit = ""
		r.MetricType = ""
		r.CalculationFormula = &interfaces.MetricCalculationFormula{}
		if err := ValidateMetricRequest(ctx, r, false); err != nil {
			t.Fatal(err)
		}
	})
}

func TestValidateMetricRequests_duplicateName(t *testing.T) {
	ctx := context.Background()
	e := validStrictCreateMetric()
	e.Name = "dup"
	err := ValidateMetricRequests(ctx, []*interfaces.MetricDefinition{e, e}, true, interfaces.ImportMode_Normal)
	if err == nil {
		t.Fatal("expected duplicate error")
	}
}

func TestValidateUpdateMetricRequest(t *testing.T) {
	ctx := context.Background()
	t.Run("calculation_formula partial invalid", func(t *testing.T) {
		req := &interfaces.MetricDefinition{
			CalculationFormula: &interfaces.MetricCalculationFormula{
				Aggregation: interfaces.MetricAggregation{Property: "", Aggr: "sum"},
			},
		}
		err := ValidateUpdateMetricRequest(ctx, req, true)
		if err == nil {
			t.Fatal("expected error")
		}
	})
	t.Run("ok empty body", func(t *testing.T) {
		if err := ValidateUpdateMetricRequest(ctx, &interfaces.MetricDefinition{}, true); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("ok only route metadata", func(t *testing.T) {
		if err := ValidateUpdateMetricRequest(ctx, &interfaces.MetricDefinition{
			ID: "mid", KnID: "kn1", Branch: "main",
		}, true); err != nil {
			t.Fatal(err)
		}
	})
}
