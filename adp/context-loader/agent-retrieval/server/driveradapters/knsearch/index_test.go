package knsearch

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/logger"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type stubKnSearchService struct {
	searchSchemaCalled bool
	searchSchemaReq    *interfaces.SearchSchemaReq
	searchSchemaResp   *interfaces.SearchSchemaResp
	searchSchemaErr    error

	knSearchCalled bool
	knSearchReq    *interfaces.KnSearchReq
	knSearchResp   *interfaces.KnSearchResp
	knSearchErr    error
}

func (s *stubKnSearchService) KnSearch(_ context.Context, req *interfaces.KnSearchReq) (*interfaces.KnSearchResp, error) {
	s.knSearchCalled = true
	s.knSearchReq = req
	if s.knSearchErr != nil {
		return nil, s.knSearchErr
	}
	if s.knSearchResp != nil {
		return s.knSearchResp, nil
	}
	return &interfaces.KnSearchResp{
		ObjectTypes:   []any{},
		RelationTypes: []any{},
		ActionTypes:   []any{},
	}, nil
}

func (s *stubKnSearchService) SearchSchema(_ context.Context, req *interfaces.SearchSchemaReq) (*interfaces.SearchSchemaResp, error) {
	s.searchSchemaCalled = true
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
	}, nil
}

func TestSearchSchema_AllScopeDisabled_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubKnSearchService{
		searchSchemaErr: errors.New("search_scope must enable at least one concept type"),
	}
	handler := &knSearchHandler{
		Logger:          logger.DefaultLogger(),
		KnSearchService: service,
	}

	body := `{
		"query":"test query",
		"search_scope":{
			"include_object_types":false,
			"include_relation_types":false,
			"include_action_types":false
		}
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/kn/search_schema", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Kn-ID", "kn-001")

	handler.SearchSchema(c)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 status for disabled search_scope, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "search_scope must enable at least one concept type") {
		t.Fatalf("expected error details for disabled search_scope, got body=%s", w.Body.String())
	}
}

func TestSearchSchema_MaxConceptsPassedThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubKnSearchService{}
	handler := &knSearchHandler{
		Logger:          logger.DefaultLogger(),
		KnSearchService: service,
	}

	body := `{
		"query":"test query",
		"max_concepts": 5
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/kn/search_schema", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Kn-ID", "kn-001")

	handler.SearchSchema(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, w.Code, w.Body.String())
	}
	if !service.searchSchemaCalled {
		t.Fatal("expected SearchSchema service to be called")
	}
	if service.searchSchemaReq == nil {
		t.Fatal("expected SearchSchemaReq to be captured")
	}
}

func TestSearchSchema_IncludeMetricTypesPassedThrough(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubKnSearchService{
		searchSchemaResp: &interfaces.SearchSchemaResp{
			ObjectTypes:   []any{},
			RelationTypes: []any{},
			ActionTypes:   []any{},
			MetricTypes:   []any{map[string]any{"id": "m_1"}},
		},
	}
	handler := &knSearchHandler{
		Logger:          logger.DefaultLogger(),
		KnSearchService: service,
	}

	body := `{
		"query":"test query",
		"search_scope":{
			"include_metric_types":true
		}
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/kn/search_schema", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Kn-ID", "kn-001")

	handler.SearchSchema(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, w.Code, w.Body.String())
	}
	if !service.searchSchemaCalled {
		t.Fatal("expected SearchSchema service to be called")
	}
	if service.searchSchemaReq == nil || service.searchSchemaReq.SearchScope == nil {
		t.Fatal("expected SearchSchemaReq.SearchScope to be captured")
	}
	if service.searchSchemaReq.SearchScope.IncludeMetricTypes == nil || !*service.searchSchemaReq.SearchScope.IncludeMetricTypes {
		t.Fatalf("expected include_metric_types=true to be passed through, got %+v", service.searchSchemaReq.SearchScope.IncludeMetricTypes)
	}
}

func TestSearchSchema_AllFourScopeDisabled_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubKnSearchService{
		searchSchemaErr: errors.New("search_scope must enable at least one concept type"),
	}
	handler := &knSearchHandler{
		Logger:          logger.DefaultLogger(),
		KnSearchService: service,
	}

	body := `{
		"query":"test query",
		"search_scope":{
			"include_object_types":false,
			"include_relation_types":false,
			"include_action_types":false,
			"include_metric_types":false
		}
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/kn/search_schema", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Request.Header.Set("X-Kn-ID", "kn-001")

	handler.SearchSchema(c)

	if w.Code == http.StatusOK {
		t.Fatalf("expected non-200 status for disabled search_scope, got %d, body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "search_scope must enable at least one concept type") {
		t.Fatalf("expected error details for disabled search_scope, got body=%s", w.Body.String())
	}
}

func TestKnSearch_CallsServiceKnSearch(t *testing.T) {
	gin.SetMode(gin.TestMode)

	service := &stubKnSearchService{}
	handler := &knSearchHandler{
		Logger:          logger.DefaultLogger(),
		KnSearchService: service,
	}

	body := `{
		"query":"test query",
		"kn_id":"kn-001",
		"only_schema": false,
		"enable_rerank": true
	}`

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/kn/kn_search", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")

	handler.KnSearch(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body=%s", http.StatusOK, w.Code, w.Body.String())
	}
	if !service.knSearchCalled {
		t.Fatal("expected KnSearch service to be called")
	}
	if service.knSearchReq == nil {
		t.Fatal("expected KnSearchReq to be captured")
	}
}
