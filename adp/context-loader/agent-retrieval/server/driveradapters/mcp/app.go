// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package mcp provides Streamable HTTP MCP Server for Agent Retrieval.
package mcp

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	logicsKar "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knactionrecall"
	logicsKlp "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knlogicpropertyresolver"
	logicsKqs "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knquerysubgraph"
	logicsKr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knretrieval"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/logics/knsearch"
)

const (
	serverName                      = "context-loader"
	serverVersion                   = "1.0.0"
	endpointPath                    = "/api/agent-retrieval/v1/mcp"
	toolKeyKnSearch                 = "kn_search"
	toolKeyKnSchemaSearch           = "kn_schema_search"
	toolKeyQueryObjectInstance      = "query_object_instance"
	toolKeyQueryInstanceSubgraph    = "query_instance_subgraph"
	toolKeyGetLogicPropertiesValues = "get_logic_properties_values"
	toolKeyGetActionInfo            = "get_action_info"
)

// NewMCPHandler creates an http.Handler for the MCP Streamable HTTP Server.
// Tool metadata comes from schemas/tools_meta.json; schemas from schemas/*.json.
func NewMCPHandler() http.Handler {
	mcpServer := server.NewMCPServer(serverName, serverVersion,
		server.WithToolCapabilities(true),
	)

	knSearchService := knsearch.NewKnSearchService()
	knSearchName, knSearchDesc := loadToolMeta(toolKeyKnSearch)
	knSearchInput, knSearchOutput := loadToolSchemas(toolKeyKnSearch)
	mcpServer.AddTool(
		newToolWithSchemas(knSearchName, knSearchDesc, knSearchInput, knSearchOutput),
		handleKnSearch(knSearchService),
	)

	knSchemaSearchService := logicsKr.NewKnRetrievalService()
	knSchemaSearchName, knSchemaSearchDesc := loadToolMeta(toolKeyKnSchemaSearch)
	knSchemaSearchInput, knSchemaSearchOutput := loadToolSchemas(toolKeyKnSchemaSearch)
	mcpServer.AddTool(
		newToolWithSchemas(knSchemaSearchName, knSchemaSearchDesc, knSchemaSearchInput, knSchemaSearchOutput),
		handleKnSchemaSearch(knSchemaSearchService),
	)

	ontologyQuery := drivenadapters.NewOntologyQueryAccess()
	queryObjectInstanceName, queryObjectInstanceDesc := loadToolMeta(toolKeyQueryObjectInstance)
	qoiInput, qoiOutput := loadToolSchemas(toolKeyQueryObjectInstance)
	mcpServer.AddTool(
		newToolWithSchemas(queryObjectInstanceName, queryObjectInstanceDesc, qoiInput, qoiOutput),
		handleQueryObjectInstance(ontologyQuery),
	)

	knQuerySubgraphService := logicsKqs.NewKnQuerySubgraphService()
	queryInstanceSubgraphName, queryInstanceSubgraphDesc := loadToolMeta(toolKeyQueryInstanceSubgraph)
	qisInput, qisOutput := loadToolSchemas(toolKeyQueryInstanceSubgraph)
	mcpServer.AddTool(
		newToolWithSchemas(queryInstanceSubgraphName, queryInstanceSubgraphDesc, qisInput, qisOutput),
		handleQueryInstanceSubgraph(knQuerySubgraphService),
	)

	getLogicPropertiesValuesService := logicsKlp.NewKnLogicPropertyResolverService()
	getLogicPropertiesValuesName, getLogicPropertiesValuesDesc := loadToolMeta(toolKeyGetLogicPropertiesValues)
	glpvInput, glpvOutput := loadToolSchemas(toolKeyGetLogicPropertiesValues)
	mcpServer.AddTool(
		newToolWithSchemas(getLogicPropertiesValuesName, getLogicPropertiesValuesDesc, glpvInput, glpvOutput),
		handleGetLogicPropertiesValues(getLogicPropertiesValuesService),
	)

	getActionInfoService := logicsKar.NewKnActionRecallService()
	getActionInfoName, getActionInfoDesc := loadToolMeta(toolKeyGetActionInfo)
	gaiInput, gaiOutput := loadToolSchemas(toolKeyGetActionInfo)
	mcpServer.AddTool(
		newToolWithSchemas(getActionInfoName, getActionInfoDesc, gaiInput, gaiOutput),
		handleGetActionInfo(getActionInfoService),
	)

	streamableServer := server.NewStreamableHTTPServer(mcpServer,
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			return r.Context()
		}),
		server.WithEndpointPath(endpointPath),
	)

	return streamableServer
}

func newToolWithSchemas(name, description string, input, output json.RawMessage) mcp.Tool {
	tool := mcp.NewToolWithRawSchema(name, description, input)
	tool.RawOutputSchema = output
	return tool
}
