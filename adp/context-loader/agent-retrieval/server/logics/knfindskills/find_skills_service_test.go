// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"context"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

func newTestConfig() *config.Config {
	return &config.Config{
		FindSkills: config.FindSkillsConfig{
			DefaultTopK:        10,
			MaxTopK:            20,
			RecallTimeoutMs:    5000,
			TotalTimeoutMs:     10000,
			SkillsObjectTypeID: "skills",
		},
	}
}

func zhCtx() context.Context {
	return common.SetLanguageByCtx(context.Background(), common.SimplifiedChinese)
}

func enCtx() context.Context {
	return common.SetLanguageByCtx(context.Background(), common.AmericanEnglish)
}

func TestFindSkills_NonEmpty_NoMessage(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{
		queryObjectInstancesFunc: func(_ context.Context, _ *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
			return &interfaces.QueryObjectInstancesResp{Data: makeSkillInstances(2)}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.Message != "" {
		t.Errorf("expected no message when entries non-empty, got %q", resp.Message)
	}
}

func TestFindSkills_NetworkScopeTooWide(t *testing.T) {
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
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if resp.Message == "" {
		t.Fatal("expected message for empty result, got empty")
	}
	if !strings.Contains(resp.Message, "object_type_id") {
		t.Errorf("network_scope_too_wide message should suggest object_type_id, got %q", resp.Message)
	}
}

func TestFindSkills_NetworkNoSkills(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{
		queryObjectInstancesFunc: func(_ context.Context, _ *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
			return &interfaces.QueryObjectInstancesResp{Data: []any{}}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if !strings.Contains(resp.Message, "Skill") {
		t.Errorf("network_no_skills message should mention Skill, got %q", resp.Message)
	}
}

func TestFindSkills_ObjectTypeNoBinding(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if !strings.Contains(resp.Message, "绑定") {
		t.Errorf("object_type_no_binding message should mention binding (绑定), got %q", resp.Message)
	}
}

func TestFindSkills_ObjectTypeNoMatch(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_1", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			return &interfaces.QueryInstanceSubgraphResp{Entries: []interface{}{}}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if !strings.Contains(resp.Message, "对象类范围") {
		t.Errorf("object_type_no_match message should mention scope, got %q", resp.Message)
	}
}

func TestFindSkills_InstanceNoMatch(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_1", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
	}
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			return &interfaces.QueryInstanceSubgraphResp{Entries: []interface{}{}}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID:               "kn1",
		ObjectTypeID:       "contract",
		InstanceIdentities: []map[string]interface{}{{"contract_id": "C-001"}},
		TopK:               10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if !strings.Contains(resp.Message, "实例范围") {
		t.Errorf("instance_no_match message should mention instance scope, got %q", resp.Message)
	}
}

func TestFindSkills_SkillQueryNoMatch(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{
				Entries: []*interfaces.RelationType{
					{ID: "rt_1", SourceObjectTypeID: "contract", TargetObjectTypeID: "skills"},
				},
			}, nil
		},
		getObjectTypeDetailFunc: func(_ context.Context, _ string, _ []string, _ bool) ([]*interfaces.ObjectType, error) {
			return []*interfaces.ObjectType{{
				ID:   "skills",
				Name: "skills",
				DataProperties: []*interfaces.DataProperty{
					{Name: "name", ConditionOperations: []interfaces.KnOperationType{interfaces.KnOperationTypeLike}},
				},
			}}, nil
		},
	}
	oq := &testOntologyQuery{
		queryInstanceSubgraphFunc: func(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
			return &interfaces.QueryInstanceSubgraphResp{Entries: []interface{}{}}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID:         "kn1",
		ObjectTypeID: "contract",
		SkillQuery:   "不存在的技能",
		TopK:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if !strings.Contains(resp.Message, "skill_query") {
		t.Errorf("skill_query_no_match message should mention skill_query, got %q", resp.Message)
	}
}

func TestFindSkills_SkillsOTID_ReturnsResults(t *testing.T) {
	bkn := &testBknBackend{}
	oq := &testOntologyQuery{
		queryObjectInstancesFunc: func(_ context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
			return &interfaces.QueryObjectInstancesResp{Data: makeSkillInstances(3)}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "skills", TopK: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 3 {
		t.Fatalf("expected 3 entries with object_type_id=skills, got %d", len(resp.Entries))
	}
	if resp.Message != "" {
		t.Errorf("expected no message when entries non-empty, got %q", resp.Message)
	}
}

func TestFindSkills_SkillsOTID_EmptyResult_MessageIsNoMatch(t *testing.T) {
	bkn := &testBknBackend{}
	oq := &testOntologyQuery{
		queryObjectInstancesFunc: func(_ context.Context, _ *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
			return &interfaces.QueryObjectInstancesResp{Data: []any{}}, nil
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "skills", TopK: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Entries) != 0 {
		t.Fatalf("expected 0 entries, got %d", len(resp.Entries))
	}
	if strings.Contains(resp.Message, "未配置") {
		t.Errorf("object_type_id=skills empty result should NOT be no_binding message (未配置), got %q", resp.Message)
	}
	if !strings.Contains(resp.Message, "对象类范围") {
		t.Errorf("expected object_type_no_match message (对象类范围), got %q", resp.Message)
	}
}

func TestFindSkills_EmptyResult_EnglishMessage(t *testing.T) {
	bkn := &testBknBackend{
		searchRelationTypesFunc: func(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
			return &interfaces.RelationTypeConcepts{Entries: []*interfaces.RelationType{}}, nil
		},
	}
	oq := &testOntologyQuery{}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), oq, bkn)

	resp, err := svc.FindSkills(enCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(resp.Message, "binding") {
		t.Errorf("expected English message containing 'binding', got %q", resp.Message)
	}
}
