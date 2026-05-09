// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mcp

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/smartystreets/goconvey/convey"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// ==================== Stub Implementations ====================

type stubLogicPropertyResolverService struct {
	resp *interfaces.ResolveLogicPropertiesResponse
	err  error
	req  *interfaces.ResolveLogicPropertiesRequest
}

func (s *stubLogicPropertyResolverService) ResolveLogicProperties(_ context.Context, req *interfaces.ResolveLogicPropertiesRequest) (*interfaces.ResolveLogicPropertiesResponse, error) {
	s.req = req
	return s.resp, s.err
}

type stubOntologyQuery struct {
	resp *interfaces.QueryObjectInstancesResp
	err  error
	req  *interfaces.QueryObjectInstancesReq
}

func (s *stubOntologyQuery) QueryObjectInstances(_ context.Context, req *interfaces.QueryObjectInstancesReq) (*interfaces.QueryObjectInstancesResp, error) {
	s.req = req
	return s.resp, s.err
}

func (s *stubOntologyQuery) QueryLogicProperties(_ context.Context, _ *interfaces.QueryLogicPropertiesReq) (*interfaces.QueryLogicPropertiesResp, error) {
	return nil, nil
}

func (s *stubOntologyQuery) QueryActions(_ context.Context, _ *interfaces.QueryActionsRequest) (*interfaces.QueryActionsResponse, error) {
	return nil, nil
}

func (s *stubOntologyQuery) QueryInstanceSubgraph(_ context.Context, _ *interfaces.QueryInstanceSubgraphReq) (*interfaces.QueryInstanceSubgraphResp, error) {
	return nil, nil
}

type stubMCPKnSearchService struct {
	knSearchResp *interfaces.KnSearchResp
	knSearchErr  error
	knSearchReq  *interfaces.KnSearchReq

	searchSchemaResp *interfaces.SearchSchemaResp
	searchSchemaErr  error
	searchSchemaReq  *interfaces.SearchSchemaReq
}

func (s *stubMCPKnSearchService) KnSearch(_ context.Context, req *interfaces.KnSearchReq) (*interfaces.KnSearchResp, error) {
	s.knSearchReq = req
	if s.knSearchResp == nil {
		s.knSearchResp = &interfaces.KnSearchResp{}
	}
	return s.knSearchResp, s.knSearchErr
}

func (s *stubMCPKnSearchService) SearchSchema(_ context.Context, req *interfaces.SearchSchemaReq) (*interfaces.SearchSchemaResp, error) {
	s.searchSchemaReq = req
	if s.searchSchemaErr != nil {
		return nil, s.searchSchemaErr
	}
	if s.searchSchemaResp != nil {
		return s.searchSchemaResp, nil
	}
	return &interfaces.SearchSchemaResp{
		ObjectTypes:   []any{},
		RelationTypes: []any{},
		ActionTypes:   []any{},
		MetricTypes:   []any{},
	}, nil
}

// ==================== Helper ====================

func withAuthCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, interfaces.KeyAccountAuthContext, &interfaces.AccountAuthContext{
		AccountID:   "test-account",
		AccountType: "user",
	})
}

func mcpReq(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func resultToMap(t *testing.T, result *mcp.CallToolResult) map[string]interface{} {
	t.Helper()
	data, err := json.Marshal(result.StructuredContent)
	if err != nil {
		t.Fatalf("failed to marshal StructuredContent: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal StructuredContent: %v", err)
	}
	return m
}

// ==================== FR-2: search_schema ====================

func TestHandleSearchSchema_AllowsAnonymousAccess(t *testing.T) {
	convey.Convey("handleSearchSchema keeps schema exploration available without auth context", t, func() {
		stub := &stubMCPKnSearchService{
			searchSchemaResp: &interfaces.SearchSchemaResp{
				ObjectTypes:   []any{map[string]any{"id": "ot_1"}},
				RelationTypes: []any{},
				ActionTypes:   []any{},
			},
		}

		handler := handleSearchSchema(stub)
		req := mcpReq(map[string]any{
			"query":           "test query",
			"kn_id":           "kn-001",
			"response_format": "json",
		})

		result, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)
		convey.So(stub.searchSchemaReq, convey.ShouldNotBeNil)
		convey.So(stub.searchSchemaReq.XAccountID, convey.ShouldEqual, "")
		convey.So(stub.searchSchemaReq.XAccountType, convey.ShouldEqual, "")
	})
}

func TestHandleSearchSchema_MaxConceptsPassedThrough(t *testing.T) {
	convey.Convey("handleSearchSchema passes max_concepts through to SearchSchema", t, func() {
		stub := &stubMCPKnSearchService{
			searchSchemaResp: &interfaces.SearchSchemaResp{
				ObjectTypes:   []any{map[string]any{"id": "ot_1"}},
				RelationTypes: []any{},
				ActionTypes:   []any{},
			},
		}

		handler := handleSearchSchema(stub)
		req := mcpReq(map[string]any{
			"query":           "test query",
			"kn_id":           "kn-001",
			"max_concepts":    5,
			"response_format": "json",
		})

		result, err := handler(withAuthCtx(context.Background()), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)
		convey.So(stub.searchSchemaReq, convey.ShouldNotBeNil)
		convey.So(stub.searchSchemaReq.MaxConcepts, convey.ShouldNotBeNil)
		convey.So(*stub.searchSchemaReq.MaxConcepts, convey.ShouldEqual, 5)
	})
}

func TestHandleSearchSchema_IncludeMetricTypesPassedThrough(t *testing.T) {
	convey.Convey("handleSearchSchema passes include_metric_types through to SearchSchema", t, func() {
		stub := &stubMCPKnSearchService{
			searchSchemaResp: &interfaces.SearchSchemaResp{
				ObjectTypes:   []any{},
				RelationTypes: []any{},
				ActionTypes:   []any{},
				MetricTypes:   []any{map[string]any{"id": "m_1"}},
			},
		}

		handler := handleSearchSchema(stub)
		req := mcpReq(map[string]any{
			"query":           "test query",
			"kn_id":           "kn-001",
			"response_format": "json",
			"search_scope": map[string]any{
				"concept_groups":       []any{"supply_chain"},
				"include_metric_types": true,
			},
		})

		result, err := handler(withAuthCtx(context.Background()), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)
		convey.So(stub.searchSchemaReq, convey.ShouldNotBeNil)
		convey.So(stub.searchSchemaReq.SearchScope, convey.ShouldNotBeNil)
		convey.So(stub.searchSchemaReq.SearchScope.ConceptGroups, convey.ShouldResemble, []string{"supply_chain"})
		convey.So(stub.searchSchemaReq.SearchScope.IncludeMetricTypes, convey.ShouldNotBeNil)
		convey.So(*stub.searchSchemaReq.SearchScope.IncludeMetricTypes, convey.ShouldBeTrue)
	})
}

func TestHandleSearchSchema_ReturnsMetricTypes(t *testing.T) {
	convey.Convey("handleSearchSchema includes metric_types in structured output", t, func() {
		stub := &stubMCPKnSearchService{
			searchSchemaResp: &interfaces.SearchSchemaResp{
				ObjectTypes:   []any{},
				RelationTypes: []any{},
				ActionTypes:   []any{},
				MetricTypes: []any{
					map[string]any{
						"id":                  "m_1",
						"name":                "cpu_usage",
						"metric_type":         "atomic",
						"scope_type":          "object_type",
						"scope_ref":           "pod",
						"calculation_formula": map[string]any{"op": "avg"},
					},
				},
			},
		}

		handler := handleSearchSchema(stub)
		req := mcpReq(map[string]any{
			"query":           "test query",
			"kn_id":           "kn-001",
			"response_format": "json",
		})

		result, err := handler(withAuthCtx(context.Background()), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)

		resultMap := resultToMap(t, result)
		metricTypes, ok := resultMap["metric_types"].([]interface{})
		convey.So(ok, convey.ShouldBeTrue)
		convey.So(len(metricTypes), convey.ShouldEqual, 1)
	})
}

// ==================== FR-3: get_logic_properties_values ====================

func TestHandleGetLogicPropertiesValues_FixesDefaultParams(t *testing.T) {
	convey.Convey("handleGetLogicPropertiesValues overrides options with fixed defaults", t, func() {
		stub := &stubLogicPropertyResolverService{
			resp: &interfaces.ResolveLogicPropertiesResponse{
				Datas: []map[string]any{{"metric_1": 100}},
			},
		}

		handler := handleGetLogicPropertiesValues(stub)
		req := mcpReq(map[string]any{
			"kn_id":                "kn-001",
			"ot_id":                "ot-001",
			"query":                "last year revenue",
			"_instance_identities": []any{map[string]any{"id": "inst_1"}},
			"properties":           []any{"revenue"},
			"options": map[string]any{
				"return_debug":      true,
				"max_repair_rounds": 5,
				"max_concurrency":   10,
			},
		})

		ctx := withAuthCtx(context.Background())
		result, err := handler(ctx, req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)

		convey.So(stub.req, convey.ShouldNotBeNil)
		convey.So(stub.req.Options, convey.ShouldNotBeNil)
		convey.So(stub.req.Options.ReturnDebug, convey.ShouldBeFalse)
		convey.So(stub.req.Options.MaxRepairRounds, convey.ShouldEqual, 1)
		convey.So(stub.req.Options.MaxConcurrency, convey.ShouldEqual, 4)
	})
}

func TestHandleGetLogicPropertiesValues_RequiresAuth(t *testing.T) {
	convey.Convey("handleGetLogicPropertiesValues returns error without auth context", t, func() {
		stub := &stubLogicPropertyResolverService{}
		handler := handleGetLogicPropertiesValues(stub)
		req := mcpReq(map[string]any{})

		result, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeTrue)
	})
}

// ==================== FR-4: query_object_instance ====================

func TestHandleQueryObjectInstance_StripsObjectType(t *testing.T) {
	convey.Convey("handleQueryObjectInstance strips object_type from output", t, func() {
		stub := &stubOntologyQuery{
			resp: &interfaces.QueryObjectInstancesResp{
				Data: []any{
					map[string]any{"id": "inst_1", "name": "Instance1"},
				},
				ObjectConcept: map[string]any{
					"id":   "ot_1",
					"name": "ObjectType1",
				},
			},
		}

		handler := handleQueryObjectInstance(stub)
		req := mcpReq(map[string]any{
			"kn_id":           "kn-001",
			"ot_id":           "ot-001",
			"limit":           5,
			"response_format": "json",
		})

		result, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)

		m := resultToMap(t, result)
		convey.So(m, convey.ShouldNotContainKey, "object_type")
		convey.So(m, convey.ShouldContainKey, "datas")
	})
}

func TestHandleQueryObjectInstance_FixesIncludeTypeInfoFalse(t *testing.T) {
	convey.Convey("handleQueryObjectInstance forces include_type_info=false", t, func() {
		stub := &stubOntologyQuery{
			resp: &interfaces.QueryObjectInstancesResp{
				Data: []any{map[string]any{"id": "inst_1"}},
			},
		}

		handler := handleQueryObjectInstance(stub)
		req := mcpReq(map[string]any{
			"kn_id":             "kn-001",
			"ot_id":             "ot-001",
			"include_type_info": true,
			"limit":             10,
			"response_format":   "json",
		})

		_, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)

		convey.So(stub.req, convey.ShouldNotBeNil)
		convey.So(stub.req.IncludeTypeInfo, convey.ShouldBeFalse)
	})
}

func TestHandleQueryObjectInstance_DefaultsLimitTo10(t *testing.T) {
	convey.Convey("handleQueryObjectInstance defaults limit to 10 when not provided", t, func() {
		stub := &stubOntologyQuery{
			resp: &interfaces.QueryObjectInstancesResp{
				Data: []any{map[string]any{"id": "inst_1"}},
			},
		}

		handler := handleQueryObjectInstance(stub)
		req := mcpReq(map[string]any{
			"kn_id":           "kn-001",
			"ot_id":           "ot-001",
			"response_format": "json",
		})

		_, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)

		convey.So(stub.req, convey.ShouldNotBeNil)
		convey.So(stub.req.Limit, convey.ShouldEqual, 10)
	})
}

func TestHandleQueryObjectInstance_RespectsExplicitLimit(t *testing.T) {
	convey.Convey("handleQueryObjectInstance respects explicit limit value", t, func() {
		stub := &stubOntologyQuery{
			resp: &interfaces.QueryObjectInstancesResp{
				Data: []any{map[string]any{"id": "inst_1"}},
			},
		}

		handler := handleQueryObjectInstance(stub)
		req := mcpReq(map[string]any{
			"kn_id":           "kn-001",
			"ot_id":           "ot-001",
			"limit":           25,
			"response_format": "json",
		})

		_, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)

		convey.So(stub.req, convey.ShouldNotBeNil)
		convey.So(stub.req.Limit, convey.ShouldEqual, 25)
	})
}
