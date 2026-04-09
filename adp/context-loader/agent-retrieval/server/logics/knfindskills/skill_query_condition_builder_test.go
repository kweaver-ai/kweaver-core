// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func makeSkillsObjType(nameOps, descOps []interfaces.KnOperationType) *interfaces.ObjectType {
	return &interfaces.ObjectType{
		ID:   "skills",
		Name: "skills",
		DataProperties: []*interfaces.DataProperty{
			{
				Name:                "name",
				Type:                "text",
				ConditionOperations: nameOps,
			},
			{
				Name:                "description",
				Type:                "text",
				ConditionOperations: descOps,
			},
		},
	}
}

func TestBuildSkillQueryCondition_EmptyQuery(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeMatch},
	)
	cond := BuildSkillQueryCondition("", ot, 10)
	if cond != nil {
		t.Error("expected nil condition for empty query")
	}
}

func TestBuildSkillQueryCondition_NilObjectType(t *testing.T) {
	cond := BuildSkillQueryCondition("test", nil, 10)
	if cond != nil {
		t.Error("expected nil condition for nil object type")
	}
}

func TestBuildSkillQueryCondition_KnnAndMatch(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.Operation != interfaces.KnOperationTypeOr {
		t.Fatalf("expected OR root, got %s", cond.Operation)
	}
	// 2 fields x 2 ops = 4 flat sub-conditions: name_knn, name_match, desc_knn, desc_match
	if len(cond.SubConditions) != 4 {
		t.Fatalf("expected 4 flat sub-conditions, got %d", len(cond.SubConditions))
	}
	expect := []struct {
		field string
		op    interfaces.KnOperationType
	}{
		{"name", interfaces.KnOperationTypeKnn},
		{"name", interfaces.KnOperationTypeMatch},
		{"description", interfaces.KnOperationTypeKnn},
		{"description", interfaces.KnOperationTypeMatch},
	}
	for i, e := range expect {
		sub := cond.SubConditions[i]
		if sub.Field != e.field || sub.Operation != e.op {
			t.Errorf("sub[%d]: expected %s/%s, got %s/%s", i, e.field, e.op, sub.Field, sub.Operation)
		}
	}
}

func TestBuildSkillQueryCondition_KnnOnly(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.Operation != interfaces.KnOperationTypeOr {
		t.Fatalf("expected OR root, got %s", cond.Operation)
	}
	for _, sub := range cond.SubConditions {
		if sub.Operation != interfaces.KnOperationTypeKnn {
			t.Errorf("expected knn for field %s, got %s", sub.Field, sub.Operation)
		}
	}
}

func TestBuildSkillQueryCondition_MixedKnnMatchPerField(t *testing.T) {
	// name supports knn+match, description supports only match
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeKnn, interfaces.KnOperationTypeMatch},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeMatch},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.Operation != interfaces.KnOperationTypeOr {
		t.Fatalf("expected OR root, got %s", cond.Operation)
	}
	// name contributes knn + match, description contributes match → 3 flat sub-conditions
	if len(cond.SubConditions) != 3 {
		t.Fatalf("expected 3 flat sub-conditions, got %d", len(cond.SubConditions))
	}
	expect := []struct {
		field string
		op    interfaces.KnOperationType
	}{
		{"name", interfaces.KnOperationTypeKnn},
		{"name", interfaces.KnOperationTypeMatch},
		{"description", interfaces.KnOperationTypeMatch},
	}
	for i, e := range expect {
		sub := cond.SubConditions[i]
		if sub.Field != e.field || sub.Operation != e.op {
			t.Errorf("sub[%d]: expected %s/%s, got %s/%s", i, e.field, e.op, sub.Field, sub.Operation)
		}
	}
}

func TestBuildSkillQueryCondition_MatchFallback(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeMatch},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeMatch},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	for _, sub := range cond.SubConditions {
		if sub.Operation != interfaces.KnOperationTypeMatch {
			t.Errorf("expected match for field %s, got %s", sub.Field, sub.Operation)
		}
	}
}

func TestBuildSkillQueryCondition_LikeFallback(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeLike},
		[]interfaces.KnOperationType{},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	// Only name field has like, description has no ops -> single condition returned directly
	if cond.Field != "name" {
		t.Errorf("expected field=name, got %s", cond.Field)
	}
	if cond.Operation != interfaces.KnOperationTypeLike {
		t.Errorf("expected like, got %s", cond.Operation)
	}
}

func TestBuildSkillQueryCondition_NoUsableOps(t *testing.T) {
	ot := makeSkillsObjType(
		[]interfaces.KnOperationType{interfaces.KnOperationTypeEqual},
		[]interfaces.KnOperationType{interfaces.KnOperationTypeEqual},
	)
	cond := BuildSkillQueryCondition("审查", ot, 10)
	if cond != nil {
		t.Error("expected nil when no knn/match/like available")
	}
}
