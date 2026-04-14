// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knfindskills

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
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

func makeSkillsObjectTypeWithProps(propNames ...string) *interfaces.ObjectType {
	props := make([]*interfaces.DataProperty, 0, len(propNames))
	for _, name := range propNames {
		props = append(props, &interfaces.DataProperty{Name: name})
	}
	return &interfaces.ObjectType{
		ID:             "skills",
		Name:           "skills",
		DataProperties: props,
	}
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

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{KnID: "kn1", ObjectTypeID: "skills", TopK: 10})
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

func TestFindSkills_ObjectTypeRequired(t *testing.T) {
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), &testOntologyQuery{}, &testBknBackend{})

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{KnID: "kn1", TopK: 10})
	if err == nil {
		t.Fatal("expected error when object_type_id is missing")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %#v", resp)
	}

	httpErr := &infraErr.HTTPError{}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.HTTPCode != 400 {
		t.Fatalf("expected HTTP 400, got %d", httpErr.HTTPCode)
	}
	if !strings.Contains(httpErr.Error(), "object_type_id") {
		t.Fatalf("expected error to mention object_type_id, got %s", httpErr.Error())
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

func TestFindSkills_ObjectTypeNotFound(t *testing.T) {
	bkn := &testBknBackend{
		getObjectTypeDetailFunc: func(_ context.Context, knID string, otIDs []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
			if knID != "kn1" {
				t.Fatalf("expected knID=kn1, got %s", knID)
			}
			if len(otIDs) != 1 {
				t.Fatalf("expected a single object type lookup, got %v", otIDs)
			}
			switch otIDs[0] {
			case "skills":
				if !includeDetail {
					t.Fatal("expected includeDetail=true for skills contract check")
				}
				return []*interfaces.ObjectType{makeSkillsObjectTypeWithProps("skill_id", "name", "description")}, nil
			case "contract":
				if includeDetail {
					t.Fatal("expected includeDetail=false for business object type existence check")
				}
				return []*interfaces.ObjectType{}, nil
			default:
				t.Fatalf("unexpected object type lookup: %v", otIDs)
				return nil, nil
			}
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), &testOntologyQuery{}, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err == nil {
		t.Fatal("expected error when object_type_id does not exist in current knowledge network")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %#v", resp)
	}

	httpErr := &infraErr.HTTPError{}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.HTTPCode != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.HTTPCode)
	}

	details, ok := httpErr.ErrorDetails.(map[string]interface{})
	if !ok {
		t.Fatalf("expected structured error details, got %T", httpErr.ErrorDetails)
	}
	if details["kn_id"] != "kn1" {
		t.Fatalf("expected details.kn_id=kn1, got %#v", details["kn_id"])
	}
	if details["object_type_id"] != "contract" {
		t.Fatalf("expected details.object_type_id=contract, got %#v", details["object_type_id"])
	}
}

func TestFindSkills_SkillsObjectTypeNotFound(t *testing.T) {
	bkn := &testBknBackend{
		getObjectTypeDetailFunc: func(_ context.Context, _ string, otIDs []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
			if len(otIDs) != 1 {
				t.Fatalf("expected a single object type lookup, got %v", otIDs)
			}
			switch otIDs[0] {
			case "contract":
				if includeDetail {
					t.Fatal("expected includeDetail=false for business object type existence check")
				}
				return []*interfaces.ObjectType{{ID: "contract", Name: "contract"}}, nil
			case "skills":
				if !includeDetail {
					t.Fatal("expected includeDetail=true for skills contract check")
				}
				return []*interfaces.ObjectType{}, nil
			default:
				t.Fatalf("unexpected object type lookup: %v", otIDs)
				return nil, nil
			}
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), &testOntologyQuery{}, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err == nil {
		t.Fatal("expected error when skills object type is missing")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %#v", resp)
	}

	httpErr := &infraErr.HTTPError{}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.HTTPCode != 404 {
		t.Fatalf("expected HTTP 404, got %d", httpErr.HTTPCode)
	}
}

func TestFindSkills_SkillsContractMissingRequiredProperties(t *testing.T) {
	bkn := &testBknBackend{
		getObjectTypeDetailFunc: func(_ context.Context, _ string, otIDs []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
			if len(otIDs) != 1 {
				t.Fatalf("expected a single object type lookup, got %v", otIDs)
			}
			switch otIDs[0] {
			case "contract":
				if includeDetail {
					t.Fatal("expected includeDetail=false for business object type existence check")
				}
				return []*interfaces.ObjectType{{ID: "contract", Name: "contract"}}, nil
			case "skills":
				if !includeDetail {
					t.Fatal("expected includeDetail=true for skills contract check")
				}
				return []*interfaces.ObjectType{makeSkillsObjectTypeWithProps("description")}, nil
			default:
				t.Fatalf("unexpected object type lookup: %v", otIDs)
				return nil, nil
			}
		},
	}
	svc := NewFindSkillsServiceWith(&testLogger{}, newTestConfig(), &testOntologyQuery{}, bkn)

	resp, err := svc.FindSkills(zhCtx(), &interfaces.FindSkillsReq{
		KnID: "kn1", ObjectTypeID: "contract", TopK: 10,
	})
	if err == nil {
		t.Fatal("expected error when skills contract is missing required properties")
	}
	if resp != nil {
		t.Fatalf("expected nil response on error, got %#v", resp)
	}

	httpErr := &infraErr.HTTPError{}
	if !errors.As(err, &httpErr) {
		t.Fatalf("expected HTTPError, got %T", err)
	}
	if httpErr.HTTPCode != 400 {
		t.Fatalf("expected HTTP 400, got %d", httpErr.HTTPCode)
	}

	details, ok := httpErr.ErrorDetails.(map[string]interface{})
	if !ok {
		t.Fatalf("expected structured error details, got %T", httpErr.ErrorDetails)
	}
	missing, ok := details["missing_data_properties"].([]string)
	if !ok {
		t.Fatalf("expected missing_data_properties to be []string, got %T", details["missing_data_properties"])
	}
	if len(missing) != 2 {
		t.Fatalf("expected 2 missing properties, got %v", missing)
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
					{Name: "skill_id"},
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
