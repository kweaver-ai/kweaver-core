// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/drivenadapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knactionrecall"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knfindskills"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knlogicpropertyresolver"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knqueryobjectinstance"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knquerysubgraph"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knretrieval"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/knsearch"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters/mcp"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type restPublicHandler struct {
	Hydra                          interfaces.Hydra
	KnRetrievalHandler             knretrieval.KnRetrievalHandler
	MCPHandler                     http.Handler
	KnLogicPropertyResolverHandler knlogicpropertyresolver.KnLogicPropertyResolverHandler
	KnActionRecallHandler          knactionrecall.KnActionRecallHandler
	KnQueryObjectInstanceHandler   knqueryobjectinstance.KnQueryObjectInstanceHandler
	KnQuerySubgraphHandler         knquerysubgraph.KnQuerySubgraphHandler
	KnSearchHandler                knsearch.KnSearchHandler
	KnFindSkillsHandler            knfindskills.KnFindSkillsHandler
	Logger                         interfaces.Logger
}

// NewRestPublicHandler 创建restHandler实例
func NewRestPublicHandler(logger interfaces.Logger) interfaces.HTTPRouterInterface {
	return &restPublicHandler{
		Hydra:                          drivenadapters.NewHydra(),
		KnRetrievalHandler:             knretrieval.NewKnRetrievalHandler(),
		MCPHandler:                     mcp.NewMCPHandler(),
		KnLogicPropertyResolverHandler: knlogicpropertyresolver.NewKnLogicPropertyResolverHandler(),
		KnActionRecallHandler:          knactionrecall.NewKnActionRecallHandler(),
		KnQueryObjectInstanceHandler:   knqueryobjectinstance.NewKnQueryObjectInstanceHandler(),
		KnQuerySubgraphHandler:         knquerysubgraph.NewKnQuerySubgraphHandler(),
		KnSearchHandler:                knsearch.NewKnSearchHandler(),
		KnFindSkillsHandler:            knfindskills.NewKnFindSkillsHandler(),
		Logger:                         logger,
	}
}

// RegisterPublic 注册公共路由
func (r *restPublicHandler) RegisterRouter(engine *gin.RouterGroup) {
	mws := []gin.HandlerFunc{}
	mws = append(mws, middlewareRequestLog(r.Logger), middlewareTrace, middlewareIntrospectVerify(r.Hydra), middlewareResponseFormat())
	engine.Use(mws...)

	engine.POST("/kn/semantic-search", r.KnRetrievalHandler.SemanticSearch)
	engine.POST("/kn/logic-property-resolver", r.KnLogicPropertyResolverHandler.ResolveLogicProperties)
	engine.POST("/kn/get_action_info", r.KnActionRecallHandler.GetActionInfo)
	engine.POST("/kn/query_object_instance", r.KnQueryObjectInstanceHandler.QueryObjectInstance)
	engine.POST("/kn/query_instance_subgraph", r.KnQuerySubgraphHandler.QueryInstanceSubgraph)
	engine.POST("/kn/kn_search", r.KnSearchHandler.KnSearch)
	engine.POST("/kn/find_skills", r.KnFindSkillsHandler.FindSkills)

	// MCP Server (Bearer token auth, supports Cursor/Claude Desktop)
	engine.Any("/mcp/*path", gin.WrapH(r.MCPHandler))
}
