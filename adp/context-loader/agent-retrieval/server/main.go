// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package main

import (
	"fmt"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/driveradapters"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/telemetry"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"

	"github.com/gin-gonic/gin"
)

// Server Service
type Server struct {
	// Health check
	httpHealthHandler  interfaces.HTTPRouterInterface
	restPublicHandler  interfaces.HTTPRouterInterface
	restPrivateHandler interfaces.HTTPRouterInterface
	config             *config.Config
}

// Start starts the server
func (s *Server) Start() {
	gin.SetMode(gin.ReleaseMode)

	go func() {
		// Register router - health check
		engine := gin.New()
		engine.Use(gin.Recovery())
		engine.UseRawPath = true
		routerHealth := engine.Group("/health")
		s.httpHealthHandler.RegisterRouter(routerHealth)

		// Register internal interface router - operator related interfaces
		routerInternalGroup := engine.Group("/api/agent-retrieval/in/v1")
		routerInternalGroup.Use(gin.Recovery())
		s.restPrivateHandler.RegisterRouter(routerInternalGroup)

		// Register external router
		routerGroup := engine.Group("/api/agent-retrieval/v1")
		routerGroup.Use(gin.Recovery())
		s.restPublicHandler.RegisterRouter(routerGroup)

		url := fmt.Sprintf("%s:%d", s.config.Project.Host, s.config.Project.Port)
		err := engine.Run(url)
		if err != nil {
			s.config.Logger.Errorf("start server failed, error: %v", err)
		}
	}()
}

func main() {
	// Initialize global configuration
	config := config.NewConfigLoader()
	// Set error code language
	common.SetLang(config.Project.Language)
	s := &Server{
		config:             config,
		httpHealthHandler:  driveradapters.NewHTTPHealthHandler(),
		restPublicHandler:  driveradapters.NewRestPublicHandler(config.Logger),
		restPrivateHandler: driveradapters.NewRestPrivateHandler(config.Logger),
	}
	s.config.Logger.Info("start agent-retrieval server")
	// Check if observability is enabled
	if config.Observability.TraceEnabled {
		defer telemetry.StopTrace()
	}
	defer s.config.Logger.Info("stop agent-retrieval server")
	s.Start()
	select {}
}
