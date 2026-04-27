// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func TestBuildSubgraphRequest_ForwardDirection(t *testing.T) {
	rt := &interfaces.RelationType{
		ID:                 "rt_contract_skills",
		SourceObjectTypeID: "contract",
		TargetObjectTypeID: "skills",
	}
	skillCond := &interfaces.KnCondition{
		Field:     "name",
		Operation: interfaces.KnOperationTypeMatch,
		Value:     "审查",
		ValueFrom: interfaces.CondValueFromConst,
	}

	req := BuildSubgraphRequest("kn1", "contract", rt, nil, skillCond, 10, "skills")
	if req.KnID != "kn1" {
		t.Errorf("expected kn_id=kn1, got %s", req.KnID)
	}

	paths, ok := req.RelationTypePaths.([]map[string]interface{})
	if !ok || len(paths) != 1 {
		t.Fatal("expected 1 relation type path")
	}

	ots, ok := paths[0]["object_types"].([]map[string]interface{})
	if !ok || len(ots) != 2 {
		t.Fatal("expected 2 object_types in path")
	}

	// Forward: first is source (contract), second is target (skills)
	if ots[0]["id"] != "contract" {
		t.Errorf("expected first object_type=contract, got %v", ots[0]["id"])
	}
	if ots[1]["id"] != "skills" {
		t.Errorf("expected second object_type=skills, got %v", ots[1]["id"])
	}
	if ots[1]["condition"] == nil {
		t.Error("expected skills object_type to have condition")
	}
	if ots[1]["limit"] != 10 {
		t.Errorf("expected limit=10, got %v", ots[1]["limit"])
	}

	rts, ok := paths[0]["relation_types"].([]map[string]interface{})
	if !ok || len(rts) != 1 {
		t.Fatal("expected 1 relation_type edge")
	}
	if rts[0]["source_object_type_id"] != "contract" {
		t.Errorf("expected source=contract, got %v", rts[0]["source_object_type_id"])
	}
}

func TestBuildSubgraphRequest_ReverseDirection(t *testing.T) {
	rt := &interfaces.RelationType{
		ID:                 "rt_skills_contract",
		SourceObjectTypeID: "skills",
		TargetObjectTypeID: "contract",
	}

	req := BuildSubgraphRequest("kn1", "contract", rt, nil, nil, 5, "skills")
	paths := req.RelationTypePaths.([]map[string]interface{})
	ots := paths[0]["object_types"].([]map[string]interface{})

	// Reverse: skills is source, so object_types[0]=skills, object_types[1]=contract
	if ots[0]["id"] != "skills" {
		t.Errorf("expected first=skills (reverse), got %v", ots[0]["id"])
	}
	if ots[1]["id"] != "contract" {
		t.Errorf("expected second=contract (reverse), got %v", ots[1]["id"])
	}

	rts := paths[0]["relation_types"].([]map[string]interface{})
	if rts[0]["source_object_type_id"] != "skills" {
		t.Errorf("expected source=skills (reverse), got %v", rts[0]["source_object_type_id"])
	}
}

func TestBuildSubgraphRequest_WithInstanceCondition(t *testing.T) {
	rt := &interfaces.RelationType{
		ID:                 "rt_contract_skills",
		SourceObjectTypeID: "contract",
		TargetObjectTypeID: "skills",
	}
	instCond := &interfaces.KnCondition{
		Field:     "contract_id",
		Operation: interfaces.KnOperationTypeEqual,
		Value:     "C-001",
		ValueFrom: interfaces.CondValueFromConst,
	}

	req := BuildSubgraphRequest("kn1", "contract", rt, instCond, nil, 10, "skills")
	paths := req.RelationTypePaths.([]map[string]interface{})
	ots := paths[0]["object_types"].([]map[string]interface{})

	if ots[0]["condition"] == nil {
		t.Error("expected contract object_type to have instance condition")
	}
}
