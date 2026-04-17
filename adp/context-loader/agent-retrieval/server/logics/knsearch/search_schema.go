// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package knsearch

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"net/http"
	"strings"

	"github.com/creasty/defaults"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// SearchSchema normalizes the request, delegates to KnSearch, and filters the response.
func (s *knSearchService) SearchSchema(ctx context.Context, req *interfaces.SearchSchemaReq) (*interfaces.SearchSchemaResp, error) {
	knReq, scope, err := NormalizeSearchSchemaReq(req)
	if err != nil {
		return nil, errors.DefaultHTTPError(ctx, http.StatusBadRequest, err.Error())
	}

	resp, err := s.KnSearch(ctx, knReq)
	if err != nil {
		return nil, err
	}

	return FilterSearchSchemaResp(resp, scope), nil
}

// SearchSchemaScope holds the resolved boolean flags for output filtering.
type SearchSchemaScope struct {
	IncludeObjectTypes   bool
	IncludeRelationTypes bool
	IncludeActionTypes   bool
}

// NormalizeSearchSchemaReq converts a SearchSchemaReq into KnSearchReq + scope.
// It applies struct default tags, validates inputs, and forces schema-only semantics.
func NormalizeSearchSchemaReq(req *interfaces.SearchSchemaReq) (*interfaces.KnSearchReq, SearchSchemaScope, error) {
	if err := defaults.Set(req); err != nil {
		return nil, SearchSchemaScope{}, stderrors.New("failed to apply defaults: " + err.Error())
	}

	// SearchScope 为 nil 时（用户未传），默认三类全开；
	// 非 nil 时 defaults.Set 已填充子字段。
	scope := SearchSchemaScope{
		IncludeObjectTypes:   true,
		IncludeRelationTypes: true,
		IncludeActionTypes:   true,
	}
	if req.SearchScope != nil {
		scope.IncludeObjectTypes = *req.SearchScope.IncludeObjectTypes
		scope.IncludeRelationTypes = *req.SearchScope.IncludeRelationTypes
		scope.IncludeActionTypes = *req.SearchScope.IncludeActionTypes
	}
	if !scope.IncludeObjectTypes && !scope.IncludeRelationTypes && !scope.IncludeActionTypes {
		return nil, scope, stderrors.New("search_scope must enable at least one concept type")
	}

	knID := strings.TrimSpace(req.XKnID)
	if knID == "" {
		knID = strings.TrimSpace(req.KnID)
	}
	if knID == "" {
		return nil, scope, stderrors.New("kn_id is required (configure X-Kn-ID header or pass kn_id in body)")
	}

	if strings.TrimSpace(req.Query) == "" {
		return nil, scope, stderrors.New("query is required")
	}

	if *req.MaxConcepts <= 0 {
		return nil, scope, stderrors.New("max_concepts must be greater than 0")
	}

	onlySchema := true
	return &interfaces.KnSearchReq{
		XAccountID:   req.XAccountID,
		XAccountType: req.XAccountType,
		Query:        req.Query,
		KnID:         knID,
		OnlySchema:   &onlySchema,
		EnableRerank: req.EnableRerank,
		RetrievalConfig: &interfaces.RetrievalConfig{
			ConceptRetrieval: &interfaces.ConceptRetrievalConfig{
				TopK:        *req.MaxConcepts,
				SchemaBrief: *req.SchemaBrief,
			},
		},
	}, scope, nil
}

// FilterSearchSchemaResp builds a SearchSchemaResp from KnSearchResp, applying scope filtering.
func FilterSearchSchemaResp(resp *interfaces.KnSearchResp, scope SearchSchemaScope) *interfaces.SearchSchemaResp {
	result := &interfaces.SearchSchemaResp{
		ObjectTypes:   []any{},
		RelationTypes: []any{},
		ActionTypes:   []any{},
	}
	if resp == nil {
		return result
	}
	if scope.IncludeObjectTypes {
		result.ObjectTypes = toAnySlice(resp.ObjectTypes)
	}
	if scope.IncludeRelationTypes {
		result.RelationTypes = toAnySlice(resp.RelationTypes)
	}
	if scope.IncludeActionTypes {
		result.ActionTypes = toAnySlice(resp.ActionTypes)
	}
	return result
}

func toAnySlice(v any) []any {
	if v == nil {
		return []any{}
	}
	data, err := json.Marshal(v)
	if err != nil {
		return []any{}
	}
	var slice []any
	if err := json.Unmarshal(data, &slice); err != nil {
		return []any{}
	}
	return slice
}

