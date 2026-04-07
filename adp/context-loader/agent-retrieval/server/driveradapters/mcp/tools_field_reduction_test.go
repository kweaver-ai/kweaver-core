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

type stubKnRetrievalService struct {
	resp *interfaces.SemanticSearchResponse
	err  error
	req  *interfaces.SemanticSearchRequest
}

func (s *stubKnRetrievalService) KeywordVectorRetrieval(_ context.Context, req *interfaces.SemanticSearchRequest) (*interfaces.SemanticSearchResponse, error) {
	s.req = req
	return s.resp, s.err
}

func (s *stubKnRetrievalService) AgentIntentRetrieval(_ context.Context, req *interfaces.SemanticSearchRequest) (*interfaces.SemanticSearchResponse, error) {
	s.req = req
	return s.resp, s.err
}

func (s *stubKnRetrievalService) AgentIntentPlanning(_ context.Context, req *interfaces.SemanticSearchRequest) (*interfaces.SemanticSearchResponse, error) {
	s.req = req
	return s.resp, s.err
}

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

// ==================== FR-2: kn_schema_search ====================

func TestHandleKnSchemaSearch_StripsOutputFields(t *testing.T) {
	convey.Convey("handleKnSchemaSearch strips score/samples/query_understanding/hits_total from output", t, func() {
		stub := &stubKnRetrievalService{
			resp: &interfaces.SemanticSearchResponse{
				QueryUnderstanding: &interfaces.QueryUnderstanding{
					OriginQuery: "test query",
				},
				HitsTotal: 42,
				KnowledgeConcepts: []*interfaces.ConceptResult{
					{
						ConceptType:   interfaces.KnConceptTypeObject,
						ConceptID:     "ot_1",
						ConceptName:   "TestObject",
						ConceptDetail: map[string]any{"id": "ot_1"},
						IntentScore:   0.85,
						MatchScore:    0.90,
						RerankScore:   0.75,
						Samples:       []any{"sample1", "sample2"},
					},
				},
			},
		}

		handler := handleKnSchemaSearch(stub)
		req := mcpReq(map[string]any{
			"query":           "test query",
			"kn_id":           "kn-001",
			"response_format": "json",
		})

		result, err := handler(context.Background(), req)
		convey.So(err, convey.ShouldBeNil)
		convey.So(result, convey.ShouldNotBeNil)
		convey.So(result.IsError, convey.ShouldBeFalse)

		m := resultToMap(t, result)

		convey.So(m, convey.ShouldNotContainKey, "query_understanding")
		convey.So(m, convey.ShouldNotContainKey, "hits_total")
		convey.So(m, convey.ShouldContainKey, "concepts")

		concepts := m["concepts"].([]interface{})
		convey.So(len(concepts), convey.ShouldEqual, 1)

		concept := concepts[0].(map[string]interface{})
		convey.So(concept, convey.ShouldNotContainKey, "intent_score")
		convey.So(concept, convey.ShouldNotContainKey, "match_score")
		convey.So(concept, convey.ShouldNotContainKey, "rerank_score")
		convey.So(concept, convey.ShouldNotContainKey, "samples")

		convey.So(concept, convey.ShouldContainKey, "concept_type")
		convey.So(concept, convey.ShouldContainKey, "concept_id")
		convey.So(concept, convey.ShouldContainKey, "concept_name")
		convey.So(concept, convey.ShouldContainKey, "concept_detail")
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
			"ot_id":               "ot-001",
			"query":               "last year revenue",
			"_instance_identities": []any{map[string]any{"id": "inst_1"}},
			"properties":          []any{"revenue"},
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
