// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/creasty/defaults"
	validator "github.com/go-playground/validator/v10"
	"github.com/mark3labs/mcp-go/mcp"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	logicsKqs "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knquerysubgraph"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knsearch"
)

// handleKnSearch returns a tool handler for kn_search.
func handleKnSearch(knSearchService knsearch.KnSearchService) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// 1. Get auth context
		authCtx, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			return mcp.NewToolResultError("authentication required"), nil
		}

		// 2. Get kn_id: Header X-Kn-ID first, then arguments
		knID := ""
		if req.Header != nil {
			knID = req.Header.Get("X-Kn-ID")
		}
		if knID == "" {
			knID = req.GetString("kn_id", "")
		}
		if knID == "" {
			return mcp.NewToolResultError(
				"kn_id is required (configure X-Kn-ID header or pass kn_id in arguments)",
			), nil
		}

		// 3. Get query
		query := req.GetString("query", "")
		if query == "" {
			return mcp.NewToolResultError("query is required"), nil
		}

		// 4. Parse response_format early to avoid wasting a service call on bad input
		format, err := GetResponseFormatFromRequest(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		// 5. Build KnSearchReq
		onlySchema := req.GetBool("only_schema", false)
		enableRerank := req.GetBool("enable_rerank", true)
		searchReq := &interfaces.KnSearchReq{
			XAccountID:   authCtx.AccountID,
			XAccountType: string(authCtx.AccountType),
			KnID:         knID,
			Query:        query,
			OnlySchema:   &onlySchema,
			EnableRerank: &enableRerank,
		}
		if raw, _ := req.GetRawArguments().(map[string]any); raw != nil {
			if rc, ok := raw["retrieval_config"]; ok && rc != nil {
				searchReq.RetrievalConfig = rc
			}
		}

		// 6. Call KnSearchService
		resp, err := knSearchService.KnSearch(ctx, searchReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

// handleKnSchemaSearch handles kn_schema_search tool calls.
func handleKnSchemaSearch(service interfaces.IKnRetrievalService) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, err := GetResponseFormatFromRequest(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		searchReq := &interfaces.SemanticSearchRequest{
			SearchScope: &interfaces.SearchScopeConfig{},
		}

		if err := bindArguments(req, searchReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		searchReq.Mode = interfaces.KeywordVectorRetrieval
		searchReq.PreviousQueries = nil
		returnQueryUnderstanding := false
		searchReq.ReturnQueryUnderstanding = &returnQueryUnderstanding
		if searchReq.KnID == "" {
			searchReq.KnID = getKnIDFromHeader(req)
		}

		if err := defaults.Set(searchReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if err := validator.New().Struct(searchReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var resp *interfaces.SemanticSearchResponse
		switch searchReq.Mode {
		case interfaces.AgentIntentRetrieval:
			resp, err = service.AgentIntentRetrieval(ctx, searchReq)
		case interfaces.AgentIntentPlanning:
			resp, err = service.AgentIntentPlanning(ctx, searchReq)
		case interfaces.KeywordVectorRetrieval:
			resp, err = service.KeywordVectorRetrieval(ctx, searchReq)
		default:
			err = errors.DefaultHTTPError(ctx, http.StatusBadRequest, "mode not support")
		}
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp.QueryUnderstanding = nil
		resp.HitsTotal = 0
		for i := range resp.KnowledgeConcepts {
			resp.KnowledgeConcepts[i].IntentScore = 0
			resp.KnowledgeConcepts[i].MatchScore = 0
			resp.KnowledgeConcepts[i].RerankScore = 0
			resp.KnowledgeConcepts[i].Samples = nil
		}
		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

// handleQueryObjectInstance handles query_object_instance tool calls.
func handleQueryObjectInstance(ontologyQuery interfaces.DrivenOntologyQuery) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, err := GetResponseFormatFromRequest(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		queryReq := &interfaces.QueryObjectInstancesReq{}
		if err := bindArguments(req, queryReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		queryReq.KnID = getStringArg(req, "kn_id", queryReq.KnID)
		if queryReq.KnID == "" {
			queryReq.KnID = getKnIDFromHeader(req)
		}
		queryReq.OtID = getStringArg(req, "ot_id", queryReq.OtID)
		queryReq.IncludeTypeInfo = false
		queryReq.IncludeLogicParams = req.GetBool("include_logic_params", queryReq.IncludeLogicParams)
		if queryReq.Limit == 0 {
			queryReq.Limit = 10
		}
		if queryReq.KnID == "" || queryReq.OtID == "" {
			return mcp.NewToolResultError("kn_id and ot_id are required"), nil
		}
		if err := validator.New().Struct(queryReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := ontologyQuery.QueryObjectInstances(ctx, queryReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		resp.ObjectConcept = nil
		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

// handleQueryInstanceSubgraph handles query_instance_subgraph tool calls.
func handleQueryInstanceSubgraph(service logicsKqs.KnQuerySubgraphService) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		format, err := GetResponseFormatFromRequest(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		subgraphReq := &interfaces.QueryInstanceSubgraphReq{}
		if err := bindArguments(req, subgraphReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		subgraphReq.KnID = getStringArg(req, "kn_id", subgraphReq.KnID)
		if subgraphReq.KnID == "" {
			subgraphReq.KnID = getKnIDFromHeader(req)
		}
		subgraphReq.IncludeLogicParams = req.GetBool("include_logic_params", subgraphReq.IncludeLogicParams)
		if subgraphReq.RelationTypePaths == nil {
			return mcp.NewToolResultError("relation_type_paths is required"), nil
		}
		if subgraphReq.KnID == "" {
			return mcp.NewToolResultError("kn_id is required"), nil
		}

		resp, err := service.QueryInstanceSubgraph(ctx, subgraphReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

// handleGetLogicPropertiesValues handles get_logic_properties_values tool calls.
func handleGetLogicPropertiesValues(service interfaces.IKnLogicPropertyResolverService) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authCtx, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			return mcp.NewToolResultError("authentication required"), nil
		}

		format, err := GetResponseFormatFromRequest(req)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resolveReq := &interfaces.ResolveLogicPropertiesRequest{}
		if err := bindArguments(req, resolveReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if resolveReq.KnID == "" {
			resolveReq.KnID = getKnIDFromHeader(req)
		}
		resolveReq.AccountID = authCtx.AccountID
		resolveReq.AccountType = string(authCtx.AccountType)

		resolveReq.Options = &interfaces.ResolveOptions{
			ReturnDebug:     false,
			MaxRepairRounds: 1,
			MaxConcurrency:  4,
		}
		if err := validator.New().Struct(resolveReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := service.ResolveLogicProperties(ctx, resolveReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

// handleGetActionInfo handles get_action_info tool calls.
func handleGetActionInfo(service interfaces.IKnActionRecallService) func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		authCtx, ok := common.GetAccountAuthContextFromCtx(ctx)
		if !ok {
			return mcp.NewToolResultError("authentication required"), nil
		}
		format := rest.FormatJSON

		actionReq := &interfaces.KnActionRecallRequest{}
		if err := bindArguments(req, actionReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		if actionReq.KnID == "" {
			actionReq.KnID = getKnIDFromHeader(req)
		}
		actionReq.AccountID = authCtx.AccountID
		actionReq.AccountType = string(authCtx.AccountType)

		if err := validator.New().Struct(actionReq); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		resp, err := service.GetActionInfo(ctx, actionReq)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		result, err := BuildMCPToolResult(resp, format)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return result, nil
	}
}

func getKnIDFromHeader(req mcp.CallToolRequest) string {
	if req.Header == nil {
		return ""
	}
	return req.Header.Get("X-Kn-ID")
}

func getStringArg(req mcp.CallToolRequest, key, fallback string) string {
	if val := req.GetString(key, ""); val != "" {
		return val
	}
	return fallback
}

func bindArguments(req mcp.CallToolRequest, target any) error {
	raw := req.GetRawArguments()
	if raw == nil {
		return nil
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
