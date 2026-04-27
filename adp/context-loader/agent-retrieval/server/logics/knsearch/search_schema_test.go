package knsearch

import (
	"context"
	stderrors "errors"
	"testing"

	infraLogger "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type stubSearchSchemaLocalService struct {
	resp *interfaces.KnSearchLocalResponse
	err  error
	req  *interfaces.KnSearchLocalRequest
}

func (s *stubSearchSchemaLocalService) Search(_ context.Context, req *interfaces.KnSearchLocalRequest) (*interfaces.KnSearchLocalResponse, error) {
	s.req = req
	return s.resp, s.err
}

type stubSearchSchemaBknBackend struct {
	searchMetricTypesFunc func(ctx context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error)
	searchMetricCalls     int
}

func (s *stubSearchSchemaBknBackend) GetKnowledgeNetworkDetail(_ context.Context, _ string) (*interfaces.KnowledgeNetworkDetail, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) SearchObjectTypes(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.ObjectTypeConcepts, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) GetObjectTypeDetail(_ context.Context, _ string, _ []string, _ bool) ([]*interfaces.ObjectType, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) SearchRelationTypes(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.RelationTypeConcepts, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) GetRelationTypeDetail(_ context.Context, _ string, _ []string, _ bool) ([]*interfaces.RelationType, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) SearchActionTypes(_ context.Context, _ *interfaces.QueryConceptsReq) (*interfaces.ActionTypeConcepts, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) GetActionTypeDetail(_ context.Context, _ string, _ []string, _ bool) ([]*interfaces.ActionType, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) SearchMetricTypes(ctx context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
	s.searchMetricCalls++
	if s.searchMetricTypesFunc != nil {
		return s.searchMetricTypesFunc(ctx, req)
	}
	return &interfaces.MetricTypeConcepts{}, nil
}

func (s *stubSearchSchemaBknBackend) CreateFullBuildOntologyJob(_ context.Context, _ string, _ *interfaces.CreateFullBuildOntologyJobReq) (*interfaces.CreateJobResp, error) {
	return nil, nil
}

func (s *stubSearchSchemaBknBackend) ListOntologyJobs(_ context.Context, _ string, _ *interfaces.ListOntologyJobsReq) (*interfaces.ListOntologyJobsResp, error) {
	return nil, nil
}

func withStubSearchSchemaBknBackend(stub interfaces.BknBackendAccess, fn func()) {
	old := newBknBackendAccess
	newBknBackendAccess = func() interfaces.BknBackendAccess {
		return stub
	}
	defer func() {
		newBknBackendAccess = old
	}()
	fn()
}

func isMetricExpansionQuery(req *interfaces.QueryConceptsReq) bool {
	if req == nil || req.Cond == nil || req.Cond.Operation != interfaces.KnOperationTypeAnd {
		return false
	}
	for _, sub := range req.Cond.SubConditions {
		if sub == nil {
			continue
		}
		if sub.Field == "scope_ref" && sub.Operation == interfaces.KnOperationTypeIn {
			return true
		}
	}
	return false
}

func hasMetricExpansionScopeTypeConstraint(req *interfaces.QueryConceptsReq) bool {
	if req == nil || req.Cond == nil || req.Cond.Operation != interfaces.KnOperationTypeAnd {
		return false
	}
	for _, sub := range req.Cond.SubConditions {
		if sub == nil {
			continue
		}
		if sub.Field == "scope_type" && sub.Operation == interfaces.KnOperationTypeEqual {
			value, ok := sub.Value.(string)
			return ok && value == "object_type"
		}
	}
	return false
}

func TestSearchSchema_AppliesMaxConceptsPerResourceType(t *testing.T) {
	maxConcepts := 1
	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Object 1"},
					{ConceptID: "ot_2", ConceptName: "Object 2"},
					{ConceptID: "ot_3", ConceptName: "Object 3"},
				},
				RelationTypes: []*interfaces.KnSearchRelationType{
					{ConceptID: "rt_1", ConceptName: "Relation 1", SourceObjectTypeID: "ot_1", TargetObjectTypeID: "ot_2"},
					{ConceptID: "rt_2", ConceptName: "Relation 2", SourceObjectTypeID: "ot_2", TargetObjectTypeID: "ot_3"},
					{ConceptID: "rt_3", ConceptName: "Relation 3", SourceObjectTypeID: "ot_3", TargetObjectTypeID: "ot_1"},
				},
				ActionTypes: []*interfaces.KnSearchActionType{
					{ID: "at_1", Name: "Action 1"},
					{ID: "at_2", Name: "Action 2"},
					{ID: "at_3", Name: "Action 3"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(&stubSearchSchemaBknBackend{}, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "find schema",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}

		if got := len(resp.RelationTypes); got != 1 {
			t.Fatalf("RelationTypes len=%d, want 1", got)
		}
		if got := len(resp.ObjectTypes); got != 2 {
			t.Fatalf("ObjectTypes len=%d, want 2 relation endpoint objects", got)
		}
		if got := len(resp.ActionTypes); got != 3 {
			t.Fatalf("ActionTypes len=%d, want all actions", got)
		}

		if got := resp.ObjectTypes[0].(map[string]any)["concept_id"]; got != "ot_1" {
			t.Fatalf("ObjectTypes[0] concept_id=%v, want ot_1", got)
		}
		if got := resp.ObjectTypes[1].(map[string]any)["concept_id"]; got != "ot_2" {
			t.Fatalf("ObjectTypes[1] concept_id=%v, want ot_2", got)
		}
		if got := resp.RelationTypes[0].(map[string]any)["concept_id"]; got != "rt_1" {
			t.Fatalf("RelationTypes[0] concept_id=%v, want rt_1", got)
		}
		if got := resp.ActionTypes[0].(map[string]any)["id"]; got != "at_1" {
			t.Fatalf("ActionTypes[0] id=%v, want at_1", got)
		}
		if got := resp.ActionTypes[2].(map[string]any)["id"]; got != "at_3" {
			t.Fatalf("ActionTypes[2] id=%v, want at_3", got)
		}
	})
}

func TestSearchSchema_LimitsObjectTypesWhenRelationTypesExcluded(t *testing.T) {
	maxConcepts := 1
	includeRelationTypes := false
	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Object 1"},
					{ConceptID: "ot_2", ConceptName: "Object 2"},
				},
				RelationTypes: []*interfaces.KnSearchRelationType{
					{ConceptID: "rt_1", ConceptName: "Relation 1", SourceObjectTypeID: "ot_1", TargetObjectTypeID: "ot_2"},
				},
				ActionTypes: []*interfaces.KnSearchActionType{
					{ID: "at_1", Name: "Action 1"},
					{ID: "at_2", Name: "Action 2"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(&stubSearchSchemaBknBackend{}, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "find schema",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
			SearchScope: &interfaces.SearchSchemaScope{
				IncludeRelationTypes: &includeRelationTypes,
			},
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}

		if got := len(resp.RelationTypes); got != 0 {
			t.Fatalf("RelationTypes len=%d, want 0", got)
		}
		if got := len(resp.ObjectTypes); got != 1 {
			t.Fatalf("ObjectTypes len=%d, want 1", got)
		}
		if got := resp.ObjectTypes[0].(map[string]any)["concept_id"]; got != "ot_1" {
			t.Fatalf("ObjectTypes[0] concept_id=%v, want ot_1", got)
		}
		if got := len(resp.ActionTypes); got != 2 {
			t.Fatalf("ActionTypes len=%d, want all actions", got)
		}
	})
}

func TestSearchSchema_DefaultsIncludeMetricTypes(t *testing.T) {
	maxConcepts := 10
	backend := &stubSearchSchemaBknBackend{
		searchMetricTypesFunc: func(_ context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
			if isMetricExpansionQuery(req) {
				return &interfaces.MetricTypeConcepts{}, nil
			}
			return &interfaces.MetricTypeConcepts{
				Entries: []*interfaces.MetricType{
					{ID: "m_1", Name: "cpu_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "pod", CalculationFormula: map[string]any{"op": "avg"}},
				},
				TotalCount: 1,
			}, nil
		},
	}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "cpu usage",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}

		if got := len(resp.MetricTypes); got != 1 {
			t.Fatalf("MetricTypes len=%d, want 1", got)
		}
		if got := resp.MetricTypes[0].(map[string]any)["id"]; got != "m_1" {
			t.Fatalf("MetricTypes[0].id=%v, want m_1", got)
		}
	})
}

func TestSearchSchema_ExcludeMetricTypesFromResponse(t *testing.T) {
	maxConcepts := 10
	includeMetricTypes := false
	backend := &stubSearchSchemaBknBackend{}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "cpu usage",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
			SearchScope: &interfaces.SearchSchemaScope{
				IncludeMetricTypes: &includeMetricTypes,
			},
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}

		if got := len(resp.MetricTypes); got != 0 {
			t.Fatalf("MetricTypes len=%d, want 0", got)
		}
		if backend.searchMetricCalls != 0 {
			t.Fatalf("SearchMetricTypes called %d times, want 0", backend.searchMetricCalls)
		}
	})
}

func TestSearchSchema_MergesDirectAndExpansionMetrics(t *testing.T) {
	maxConcepts := 10
	includeObjectTypes := false
	backend := &stubSearchSchemaBknBackend{
		searchMetricTypesFunc: func(_ context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
			if isMetricExpansionQuery(req) {
				return &interfaces.MetricTypeConcepts{
					Entries: []*interfaces.MetricType{
						{ID: "m_1", Name: "cpu_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "ot_1", CalculationFormula: map[string]any{"op": "avg"}},
						{ID: "m_2", Name: "memory_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "ot_1", CalculationFormula: map[string]any{"op": "max"}},
					},
				}, nil
			}
			return &interfaces.MetricTypeConcepts{
				Entries: []*interfaces.MetricType{
					{ID: "m_1", Name: "cpu_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "ot_1", CalculationFormula: map[string]any{"op": "avg"}},
				},
			}, nil
		},
	}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "pod metrics",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
			SearchScope: &interfaces.SearchSchemaScope{
				IncludeObjectTypes: &includeObjectTypes,
			},
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}

		if got := len(resp.ObjectTypes); got != 0 {
			t.Fatalf("ObjectTypes len=%d, want 0 when object_types excluded", got)
		}
		if got := len(resp.MetricTypes); got != 2 {
			t.Fatalf("MetricTypes len=%d, want 2 after merge+dedup", got)
		}
		if got := resp.MetricTypes[0].(map[string]any)["id"]; got != "m_1" {
			t.Fatalf("MetricTypes[0].id=%v, want m_1", got)
		}
		if got := resp.MetricTypes[1].(map[string]any)["id"]; got != "m_2" {
			t.Fatalf("MetricTypes[1].id=%v, want m_2", got)
		}
	})
}

func TestSearchSchema_ExpansionQueryConstrainsScopeTypeToObjectType(t *testing.T) {
	maxConcepts := 10
	var expansionReq *interfaces.QueryConceptsReq
	backend := &stubSearchSchemaBknBackend{
		searchMetricTypesFunc: func(_ context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
			if isMetricExpansionQuery(req) {
				expansionReq = req
				return &interfaces.MetricTypeConcepts{}, nil
			}
			return &interfaces.MetricTypeConcepts{
				Entries: []*interfaces.MetricType{
					{ID: "m_1", Name: "cpu_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "ot_1", CalculationFormula: map[string]any{"op": "avg"}},
				},
			}, nil
		},
	}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		_, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "pod metrics",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}
	})

	if expansionReq == nil {
		t.Fatal("expected metric expansion query to be issued")
	}
	if !hasMetricExpansionScopeTypeConstraint(expansionReq) {
		t.Fatal("expected metric expansion query to constrain scope_type == object_type")
	}
}

func TestSearchSchema_DirectMetricRecallErrorReturnsError(t *testing.T) {
	maxConcepts := 10
	backend := &stubSearchSchemaBknBackend{
		searchMetricTypesFunc: func(_ context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
			if isMetricExpansionQuery(req) {
				t.Fatal("did not expect expansion query after direct recall failure")
			}
			return nil, stderrors.New("direct recall failed")
		},
	}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		_, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "cpu usage",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
		})
		if err == nil {
			t.Fatal("expected error when direct metric recall fails")
		}
	})
}

func TestSearchSchema_ExpansionMetricRecallErrorFallsBackToDirectOnly(t *testing.T) {
	maxConcepts := 10
	backend := &stubSearchSchemaBknBackend{
		searchMetricTypesFunc: func(_ context.Context, req *interfaces.QueryConceptsReq) (*interfaces.MetricTypeConcepts, error) {
			if isMetricExpansionQuery(req) {
				return nil, stderrors.New("expansion recall failed")
			}
			return &interfaces.MetricTypeConcepts{
				Entries: []*interfaces.MetricType{
					{ID: "m_1", Name: "cpu_usage", MetricType: "atomic", ScopeType: "object_type", ScopeRef: "ot_1", CalculationFormula: map[string]any{"op": "avg"}},
				},
			}, nil
		},
	}

	service := &knSearchService{
		Logger: infraLogger.DefaultLogger(),
		LocalSearch: &stubSearchSchemaLocalService{
			resp: &interfaces.KnSearchLocalResponse{
				ObjectTypes: []*interfaces.KnSearchObjectType{
					{ConceptID: "ot_1", ConceptName: "Pod"},
				},
			},
		},
	}

	withStubSearchSchemaBknBackend(backend, func() {
		resp, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "cpu usage",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
		})
		if err != nil {
			t.Fatalf("SearchSchema returned error: %v", err)
		}
		if got := len(resp.MetricTypes); got != 1 {
			t.Fatalf("MetricTypes len=%d, want 1 direct-only metric", got)
		}
		if got := resp.MetricTypes[0].(map[string]any)["id"]; got != "m_1" {
			t.Fatalf("MetricTypes[0].id=%v, want m_1", got)
		}
	})
}

func TestSearchSchema_AllScopeDisabled_ReturnsBadRequest(t *testing.T) {
	maxConcepts := 10
	includeObjectTypes := false
	includeRelationTypes := false
	includeActionTypes := false
	includeMetricTypes := false
	service := &knSearchService{
		Logger:         infraLogger.DefaultLogger(),
		LocalSearch:    &stubSearchSchemaLocalService{},
	}

	withStubSearchSchemaBknBackend(&stubSearchSchemaBknBackend{}, func() {
		_, err := service.SearchSchema(context.Background(), &interfaces.SearchSchemaReq{
			Query:       "anything",
			KnID:        "kn-001",
			MaxConcepts: &maxConcepts,
			SearchScope: &interfaces.SearchSchemaScope{
				IncludeObjectTypes:   &includeObjectTypes,
				IncludeRelationTypes: &includeRelationTypes,
				IncludeActionTypes:   &includeActionTypes,
				IncludeMetricTypes:   &includeMetricTypes,
			},
		})
		if err == nil {
			t.Fatal("expected error when all concept types are disabled")
		}
	})
}
