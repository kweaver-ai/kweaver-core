// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters 定义驱动适配器
// @file rest_private_handler.go
// @description: 定义rest私有接口适配器
package driveradapters

import (
	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knactionrecall"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knlogicpropertyresolver"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knontologyjob"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knqueryobjectinstance"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knquerysubgraph"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knretrieval"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knsearch"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/mcpproxy"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type restPrivateHandler struct {
	KnRetrievalHandler             knretrieval.KnRetrievalHandler
	KnLogicPropertyResolverHandler knlogicpropertyresolver.KnLogicPropertyResolverHandler
	KnActionRecallHandler          knactionrecall.KnActionRecallHandler
	KnQueryObjectInstanceHandler   knqueryobjectinstance.KnQueryObjectInstanceHandler
	KnQuerySubgraphHandler         knquerysubgraph.KnQuerySubgraphHandler
	KnSearchHandler                knsearch.KnSearchHandler
	MCPProxyHandler                mcpproxy.MCPProxyHandler
	KnOntologyJobHandler           knontologyjob.KnOntologyJobHandler
	Logger                         interfaces.Logger
}

// NewRestPrivateHandler 创建restHandler实例
func NewRestPrivateHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
	return &restPrivateHandler{
		KnRetrievalHandler:             knretrieval.NewKnRetrievalHandler(),
		KnLogicPropertyResolverHandler: knlogicpropertyresolver.NewKnLogicPropertyResolverHandler(),
		KnActionRecallHandler:          knactionrecall.NewKnActionRecallHandler(),
		KnQueryObjectInstanceHandler:   knqueryobjectinstance.NewKnQueryObjectInstanceHandler(),
		KnQuerySubgraphHandler:         knquerysubgraph.NewKnQuerySubgraphHandler(),
		KnSearchHandler:                knsearch.NewKnSearchHandler(),
		MCPProxyHandler:                mcpproxy.NewMCPProxyHandler(),
		KnOntologyJobHandler:           knontologyjob.NewKnOntologyJobHandler(),
		Logger:                         logger,
	}
}

// RegisterRouter 注册路由
func (r *restPrivateHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareHeaderAuthContext(), middlewareResponseFormat())
	engine.Use(mws...)

	engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)
	engine.POST("/kn/logic-property-resolver", r.KnLogicPropertyResolverHandler.ResolveLogicProperties)
	engine.POST("/kn/get_action_info", r.KnActionRecallHandler.GetActionInfo)
	engine.POST("/kn/query_object_instance", r.KnQueryObjectInstanceHandler.QueryObjectInstance)
	engine.POST("/kn/query_instance_subgraph", r.KnQuerySubgraphHandler.QueryInstanceSubgraph)
	engine.POST("/kn/kn_search", r.KnSearchHandler.KnSearch)
	engine.POST("/kn/full_build_ontology", r.KnOntologyJobHandler.FullBuildOntology)
	engine.GET("/kn/full_ontology_building_status", r.KnOntologyJobHandler.GetFullOntologyBuildingStatus)

	// MCP Proxy
	engine.POST("/mcp/proxy/:mcp_id/tools/:tool_name/call", r.MCPProxyHandler.CallMCPTool)
}
