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

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

var newBknBackendAccess = drivenadapters.NewBknBackendAccess

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

	metricTypes := []any{}
	if scope.IncludeMetricTypes {
		metricTypes, err = s.resolveMetricTypes(ctx, req, resp)
		if err != nil {
			return nil, err
		}
	}

	return FilterSearchSchemaResp(resp, metricTypes, scope, *req.MaxConcepts), nil
}

// SearchSchemaScope holds the resolved boolean flags for output filtering.
type SearchSchemaScope struct {
	IncludeObjectTypes   bool
	IncludeRelationTypes bool
	IncludeActionTypes   bool
	IncludeMetricTypes   bool
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
		IncludeMetricTypes:   true,
	}
	if req.SearchScope != nil {
		scope.IncludeObjectTypes = *req.SearchScope.IncludeObjectTypes
		scope.IncludeRelationTypes = *req.SearchScope.IncludeRelationTypes
		scope.IncludeActionTypes = *req.SearchScope.IncludeActionTypes
		scope.IncludeMetricTypes = *req.SearchScope.IncludeMetricTypes
	}
	if !scope.IncludeObjectTypes && !scope.IncludeRelationTypes && !scope.IncludeActionTypes && !scope.IncludeMetricTypes {
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
func FilterSearchSchemaResp(resp *interfaces.KnSearchResp, metricTypes []any, scope SearchSchemaScope, maxConcepts int) *interfaces.SearchSchemaResp {
	objectTypes := []any{}
	relationTypes := []any{}
	actionTypes := []any{}
	if resp != nil {
		objectTypes = toAnySlice(resp.ObjectTypes)
		relationTypes = toAnySlice(resp.RelationTypes)
		actionTypes = toAnySlice(resp.ActionTypes)
	}

	if scope.IncludeRelationTypes {
		relationTypes = limitAnySlice(relationTypes, maxConcepts)
	}
	if scope.IncludeObjectTypes {
		if scope.IncludeRelationTypes && len(relationTypes) > 0 {
			objectTypes = filterObjectTypesByRelations(objectTypes, relationTypes)
		} else {
			objectTypes = limitAnySlice(objectTypes, maxConcepts)
		}
	}

	result := &interfaces.SearchSchemaResp{
		ObjectTypes:   []any{},
		RelationTypes: []any{},
		ActionTypes:   []any{},
		MetricTypes:   []any{},
	}
	if scope.IncludeObjectTypes {
		result.ObjectTypes = objectTypes
	}
	if scope.IncludeRelationTypes {
		result.RelationTypes = relationTypes
	}
	if scope.IncludeActionTypes {
		result.ActionTypes = actionTypes
	}
	if scope.IncludeMetricTypes {
		result.MetricTypes = limitAnySlice(metricTypes, maxConcepts)
	}
	return result
}

func (s *knSearchService) resolveMetricTypes(ctx context.Context, req *interfaces.SearchSchemaReq, resp *interfaces.KnSearchResp) ([]any, error) {
	backend := newBknBackendAccess()
	if backend == nil {
		return []any{}, nil
	}

	directReq := buildMetricRecallQuery(strings.TrimSpace(req.KnID), strings.TrimSpace(req.Query), *req.MaxConcepts)
	if strings.TrimSpace(req.XKnID) != "" {
		directReq.KnID = strings.TrimSpace(req.XKnID)
	}

	directResp, err := backend.SearchMetricTypes(ctx, directReq)
	if err != nil {
		return nil, err
	}

	objectIDs := extractObjectCandidateIDs(resp)
	expansionMetrics := []*interfaces.MetricType{}
	if len(objectIDs) > 0 {
		expansionReq := buildMetricExpansionQuery(directReq.KnID, strings.TrimSpace(req.Query), objectIDs, *req.MaxConcepts)
		expansionResp, expansionErr := backend.SearchMetricTypes(ctx, expansionReq)
		if expansionErr != nil {
			s.Logger.WithContext(ctx).Warnf("[SearchSchema] metric expansion failed, fallback to direct recall: %v", expansionErr)
		} else if expansionResp != nil {
			expansionMetrics = expansionResp.Entries
		}
	}

	directMetrics := []*interfaces.MetricType{}
	if directResp != nil {
		directMetrics = directResp.Entries
	}

	return toAnySlice(mergeMetricTypesByID(directMetrics, expansionMetrics, *req.MaxConcepts)), nil
}

func buildMetricRecallQuery(knID, query string, limit int) *interfaces.QueryConceptsReq {
	return &interfaces.QueryConceptsReq{
		KnID: knID,
		Cond: &interfaces.KnCondition{
			Operation: interfaces.KnOperationTypeOr,
			SubConditions: []*interfaces.KnCondition{
				{
					Field:      "*",
					Operation:  interfaces.KnOperationTypeKnn,
					Value:      query,
					ValueFrom:  interfaces.CondValueFromConst,
					LimitKey:   interfaces.CondLimitKeyK,
					LimitValue: limit,
				},
				{
					Field:     "*",
					Operation: interfaces.KnOperationTypeMatch,
					Value:     query,
					ValueFrom: interfaces.CondValueFromConst,
				},
			},
		},
		Sort: []*interfaces.KnSortParams{
			{Field: "_score", Direction: "desc"},
		},
		Limit:     limit,
		NeedTotal: false,
	}
}

func buildMetricExpansionQuery(knID, query string, objectIDs []string, limit int) *interfaces.QueryConceptsReq {
	return &interfaces.QueryConceptsReq{
		KnID: knID,
		Cond: &interfaces.KnCondition{
			Operation: interfaces.KnOperationTypeAnd,
			SubConditions: []*interfaces.KnCondition{
				{
					Field:     "scope_type",
					Operation: interfaces.KnOperationTypeEqual,
					Value:     "object_type",
					ValueFrom: interfaces.CondValueFromConst,
				},
				{
					Field:     "scope_ref",
					Operation: interfaces.KnOperationTypeIn,
					Value:     objectIDs,
					ValueFrom: interfaces.CondValueFromConst,
				},
				buildMetricRecallQuery(knID, query, limit).Cond,
			},
		},
		Sort: []*interfaces.KnSortParams{
			{Field: "_score", Direction: "desc"},
		},
		Limit:     limit,
		NeedTotal: false,
	}
}

func extractObjectCandidateIDs(resp *interfaces.KnSearchResp) []string {
	if resp == nil {
		return nil
	}
	objectTypes := toAnySlice(resp.ObjectTypes)
	if len(objectTypes) == 0 {
		return nil
	}

	objectIDs := make([]string, 0, len(objectTypes))
	seen := make(map[string]struct{}, len(objectTypes))
	for _, obj := range objectTypes {
		objMap, ok := obj.(map[string]any)
		if !ok {
			continue
		}
		conceptID, ok := objMap["concept_id"].(string)
		if !ok || conceptID == "" {
			continue
		}
		if _, exists := seen[conceptID]; exists {
			continue
		}
		seen[conceptID] = struct{}{}
		objectIDs = append(objectIDs, conceptID)
	}
	return objectIDs
}

func mergeMetricTypesByID(directMetrics, expansionMetrics []*interfaces.MetricType, limit int) []*interfaces.MetricType {
	merged := make([]*interfaces.MetricType, 0, len(directMetrics)+len(expansionMetrics))
	seen := make(map[string]struct{}, len(directMetrics)+len(expansionMetrics))

	appendMetric := func(metric *interfaces.MetricType) {
		if metric == nil || metric.ID == "" {
			return
		}
		if _, exists := seen[metric.ID]; exists {
			return
		}
		seen[metric.ID] = struct{}{}
		merged = append(merged, metric)
	}

	for _, metric := range directMetrics {
		appendMetric(metric)
	}
	for _, metric := range expansionMetrics {
		appendMetric(metric)
	}

	if limit > 0 && len(merged) > limit {
		return merged[:limit]
	}
	return merged
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

func limitAnySlice(items []any, limit int) []any {
	if limit <= 0 || len(items) <= limit {
		return items
	}
	return items[:limit]
}

const relationReferencedObjectTypes = 2

func filterObjectTypesByRelations(objectTypes, relationTypes []any) []any {
	if len(objectTypes) == 0 || len(relationTypes) == 0 {
		return objectTypes
	}

	referenced := make(map[string]struct{}, len(relationTypes)*relationReferencedObjectTypes)
	for _, rel := range relationTypes {
		relMap, ok := rel.(map[string]any)
		if !ok {
			continue
		}
		if sourceID, ok := relMap["source_object_type_id"].(string); ok && sourceID != "" {
			referenced[sourceID] = struct{}{}
		}
		if targetID, ok := relMap["target_object_type_id"].(string); ok && targetID != "" {
			referenced[targetID] = struct{}{}
		}
	}
	if len(referenced) == 0 {
		return objectTypes
	}

	filtered := make([]any, 0, len(objectTypes))
	for _, obj := range objectTypes {
		objMap, ok := obj.(map[string]any)
		if !ok {
			continue
		}
		conceptID, ok := objMap["concept_id"].(string)
		if !ok {
			continue
		}
		if _, keep := referenced[conceptID]; keep {
			filtered = append(filtered, obj)
		}
	}
	return filtered
}
