// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"context"
	"fmt"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func newTestCoordinator(bkn *testBknBackend, oq *testOntologyQuery) *recallCoordinator {
	return &recallCoordinator{
		logger: &testLogger{},
		config: &config.FindSkillsConfig{
			DefaultTopK:        10,
			MaxTopK:            20,
			RecallTimeoutMs:    5000,
			TotalTimeoutMs:     10000,
			SkillsObjectTypeID: "skills",
		},
		ontologyQuery: oq,
		bknBackend:    bkn,
	}
}

func TestRecallNetwork_NoRelation_ReturnsSkills(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{
		queryObjectInstancesFunc: func(_ context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
			if req.OtID != "skills" {
				t.Errorf("expected OtID=skills, got %s", req.OtID)
			}
			return &interfaces.QueryObjectInstancesResp{
				Data: makeSkillInstances(3),
			}, nil
		},
	}

	rc := newTestCoordinator(bkn, oq)
	matches, err := rc.recallNetwork(context.Background(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("expected 3 matches, got %d", len(matches))
	}
	for _, m := range matches {
		if m.MatchedScope != "network" {
			t.Errorf("expected scope=network, got %s", m.MatchedScope)
		}
		if m.Priority != 10 {
			t.Errorf("expected priority=10, got %d", m.Priority)
		}
	}
}

func TestRecallNetwork_HasRelation_ReturnsEmpty(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_1", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}
	oq := &testOntologyQuery{}

	rc := newTestCoordinator(bkn, oq)
	matches, err := rc.recallNetwork(context.Background(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches when skills has relations, got %d", len(matches))
	}
}

func TestRecallObjectType_RelationExists(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_contract_skills", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			return &interfaces.QueryInstanceSubgraphResp{
				Entries: makeSubgraphEntries("skills", map[string]interface{}{
					"skill_id":    "skill_review",
					"name":        "合同审查",
					"description": "审查合同",
					"_score":      0.95,
				}),
			}, nil
		},
	}

	rc := newTestCoordinator(bkn, oq)
	matches, err := rc.recallObjectType(context.Background(), &interfaces.FindSkillsReq{KnID: "kn1", ObjectTypeID: "contract", TopK: 10}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].SkillID != "skill_review" {
		t.Errorf("expected skill_review, got %s", matches[0].SkillID)
	}
	if matches[0].MatchedScope != "object_type" {
		t.Errorf("expected scope=object_type, got %s", matches[0].MatchedScope)
	}
	if matches[0].Priority != 50 {
		t.Errorf("expected priority=50, got %d", matches[0].Priority)
	}
}

func TestRecallObjectType_NoRelation_ReturnsEmpty(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{}

	rc := newTestCoordinator(bkn, oq)
	matches, err := rc.recallObjectType(context.Background(), &interfaces.FindSkillsReq{KnID: "kn1", ObjectTypeID: "contract", TopK: 10}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestRecallInstance_RelationExists_DualPath(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_contract_skills", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}

	callCount := 0
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, req *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			callCount++
			return &interfaces.QueryInstanceSubgraphResp{
				Entries: makeSubgraphEntries("skills", map[string]interface{}{
					"skill_id":    fmt.Sprintf("skill_%d", callCount),
					"name":        fmt.Sprintf("技能_%d", callCount),
					"description": "描述",
					"_score":      0.8,
				}),
			}, nil
		},
	}

	rc := newTestCoordinator(bkn, oq)
	req := &interfaces.FindSkillsReq{
		KnID:               "kn1",
		ObjectTypeID:       "contract",
		InstanceIdentities: []map[string]interface{}{{"contract_id": "C-001"}},
		TopK:               10,
	}
	matches, err := rc.recallInstance(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Both paths should be called concurrently
	if oq.subgraphCallCount != 2 {
		t.Errorf("expected 2 subgraph calls (ot-level + instance-level), got %d", oq.subgraphCallCount)
	}
	if len(matches) < 1 {
		t.Errorf("expected at least 1 match, got %d", len(matches))
	}
}

func TestRecallInstance_NoRelation_ReturnsEmpty(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{}

	rc := newTestCoordinator(bkn, oq)
	req := &interfaces.FindSkillsReq{
		KnID:               "kn1",
		ObjectTypeID:       "contract",
		InstanceIdentities: []map[string]interface{}{{"contract_id": "C-001"}},
		TopK:               10,
	}
	matches, err := rc.recallInstance(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 0 {
		t.Errorf("expected 0 matches, got %d", len(matches))
	}
}

func TestRecallInstance_SinglePathFailure_PartialReturn(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_1", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}

	callIdx := 0
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			callIdx++
			if callIdx == 1 {
				return &interfaces.QueryInstanceSubgraphResp{
					Entries: makeSubgraphEntries("skills", map[string]interface{}{
						"skill_id": "skill_ok",
						"name":     "OK",
						"_score":   0.9,
					}),
				}, nil
			}
			return nil, fmt.Errorf("simulated failure")
		},
	}

	rc := newTestCoordinator(bkn, oq)
	req := &interfaces.FindSkillsReq{
		KnID:               "kn1",
		ObjectTypeID:       "contract",
		InstanceIdentities: []map[string]interface{}{{"id": "1"}},
		TopK:               10,
	}
	matches, err := rc.recallInstance(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("unexpected error (should partial return): %v", err)
	}
	if len(matches) == 0 {
		t.Error("expected partial results from the successful path")
	}
}

func TestBuildInstanceFilterCondition_Single(t *testing.T) {
	cond := buildInstanceFilterCondition([]map[string]interface{}{
		{"contract_id": "C-001"},
	})
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.Field != "contract_id" {
		t.Errorf("expected field=contract_id, got %s", cond.Field)
	}
}

func TestBuildInstanceFilterCondition_Multiple(t *testing.T) {
	cond := buildInstanceFilterCondition([]map[string]interface{}{
		{"contract_id": "C-001"},
		{"contract_id": "C-002"},
	})
	if cond == nil {
		t.Fatal("expected non-nil condition")
	}
	if cond.Operation != interfaces.KnOperationTypeOr {
		t.Errorf("expected OR for multiple identities, got %s", cond.Operation)
	}
	if len(cond.SubConditions) != 2 {
		t.Errorf("expected 2 sub-conditions, got %d", len(cond.SubConditions))
	}
}

func TestBuildInstanceFilterCondition_Empty(t *testing.T) {
	cond := buildInstanceFilterCondition(nil)
	if cond != nil {
		t.Error("expected nil for empty identities")
	}
}

func TestExtractSkillMatchesFromInstances(t *testing.T) {
	data := makeSkillInstances(2)
	matches := extractSkillMatchesFromInstances(data, "network", 10)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].SkillID != "skill_0" {
		t.Errorf("expected skill_0, got %s", matches[0].SkillID)
	}
	if matches[0].MatchedScope != "network" {
		t.Errorf("expected network, got %s", matches[0].MatchedScope)
	}
}

func TestExtractSubgraph_PropertiesNil_FallbackToParentMap(t *testing.T) {
	resp := &interfaces.QueryInstanceSubgraphResp{
		Entries: []interface{}{
			map[string]interface{}{
				"objects": map[string]interface{}{
					"skills-s1": map[string]interface{}{
						"id":             "skills-s1",
						"object_type_id": "skills",
						"display":        "直接属性技能",
						"skill_id":       "s1",
						"name":           "技能A",
					},
				},
			},
		},
	}
	matches := extractSkillMatchesFromSubgraph(resp, "skills", "test", 10)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].SkillID != "s1" {
		t.Errorf("expected s1, got %s", matches[0].SkillID)
	}
	if matches[0].Name != "技能A" {
		t.Errorf("expected 技能A, got %s", matches[0].Name)
	}
}

func TestExtractSubgraph_MixedObjectTypes_OnlySkillsExtracted(t *testing.T) {
	resp := &interfaces.QueryInstanceSubgraphResp{
		Entries: []interface{}{
			map[string]interface{}{
				"objects": map[string]interface{}{
					"contract-c1": map[string]interface{}{
						"id":             "contract-c1",
						"object_type_id": "contract",
						"properties": map[string]interface{}{
							"name": "合同A",
						},
					},
					"skills-s1": map[string]interface{}{
						"id":             "skills-s1",
						"object_type_id": "skills",
						"properties": map[string]interface{}{
							"skill_id":    "s1",
							"name":        "技能X",
							"description": "描述X",
						},
					},
					"skills-s2": map[string]interface{}{
						"id":             "skills-s2",
						"object_type_id": "skills",
						"properties": map[string]interface{}{
							"skill_id": "s2",
							"name":     "技能Y",
						},
					},
				},
			},
		},
	}
	matches := extractSkillMatchesFromSubgraph(resp, "skills", "test", 50)
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches (skills only), got %d", len(matches))
	}
	ids := map[string]bool{}
	for _, m := range matches {
		ids[m.SkillID] = true
	}
	if !ids["s1"] || !ids["s2"] {
		t.Errorf("expected s1 and s2, got %v", ids)
	}
}

func TestExtractSubgraph_SkillIDEmpty_Skipped(t *testing.T) {
	resp := &interfaces.QueryInstanceSubgraphResp{
		Entries: []interface{}{
			map[string]interface{}{
				"objects": map[string]interface{}{
					"skills-auto-id": map[string]interface{}{
						"id":             "skills-auto-id",
						"object_type_id": "skills",
						"display":        "自动ID技能",
						"properties": map[string]interface{}{
							"name":        "技能无ID",
							"description": "没有 skill_id 字段",
						},
					},
				},
			},
		},
	}
	matches := extractSkillMatchesFromSubgraph(resp, "skills", "test", 10)
	if len(matches) != 0 {
		t.Fatalf("expected 0 matches (no skill_id should be skipped), got %d", len(matches))
	}
}

func TestExtractSubgraph_ScoreMissing_DefaultsToZero(t *testing.T) {
	resp := &interfaces.QueryInstanceSubgraphResp{
		Entries: []interface{}{
			map[string]interface{}{
				"objects": map[string]interface{}{
					"skills-s1": map[string]interface{}{
						"id":             "skills-s1",
						"object_type_id": "skills",
						"properties": map[string]interface{}{
							"skill_id": "s1",
							"name":     "无分数技能",
						},
					},
				},
			},
		},
	}
	matches := extractSkillMatchesFromSubgraph(resp, "skills", "test", 10)
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].Score != 0 {
		t.Errorf("expected score=0 when _score missing, got %f", matches[0].Score)
	}
}

func TestExtractSubgraph_DuplicateSkillsAcrossEntries_AllCollected(t *testing.T) {
	resp := &interfaces.QueryInstanceSubgraphResp{
		Entries: []interface{}{
			map[string]interface{}{
				"objects": map[string]interface{}{
					"skills-s1": map[string]interface{}{
						"id":             "skills-s1",
						"object_type_id": "skills",
						"properties": map[string]interface{}{
							"skill_id": "s1",
							"name":     "技能1",
						},
					},
				},
			},
			map[string]interface{}{
				"objects": map[string]interface{}{
					"skills-s1": map[string]interface{}{
						"id":             "skills-s1",
						"object_type_id": "skills",
						"properties": map[string]interface{}{
							"skill_id": "s1",
							"name":     "技能1",
						},
					},
				},
			},
		},
	}
	matches := extractSkillMatchesFromSubgraph(resp, "skills", "test", 10)
	if len(matches) != 2 {
		t.Fatalf("expected 2 raw matches (dedup is Assemble's job), got %d", len(matches))
	}
}
