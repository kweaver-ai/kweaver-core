// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package driveradapters provides HTTP handlers (primary adapters).
package driveradapters

import (
	"context"
	"net/http"
	"time"
	"vega-backend/worker"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"vega-backend/common"
	"vega-backend/interfaces"
	"vega-backend/logics/auth"
	"vega-backend/logics/build_task"
	"vega-backend/logics/catalog"
	"vega-backend/logics/connector_type"
	"vega-backend/logics/dataset"
	"vega-backend/logics/discover_schedule"
	"vega-backend/logics/discover_task"
	"vega-backend/logics/resource"
	"vega-backend/logics/resource_data"
	"vega-backend/version"
)

// RestHandler interface
type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	as         interfaces.AuthService
	cs         interfaces.CatalogService
	rs         interfaces.ResourceService
	bts        interfaces.BuildTaskService
	ds         interfaces.DatasetService
	cts        interfaces.ConnectorTypeService
	dts        interfaces.DiscoverTaskService
	dss        interfaces.DiscoverScheduleService
	rds        interfaces.ResourceDataService

	scheduler *worker.Scheduler
}

// NewRestHandler creates a new RestHandler.
func NewRestHandler(appSetting *common.AppSetting, scheduler *worker.Scheduler) RestHandler {
	cs := catalog.NewCatalogService(appSetting)
	rs := resource.NewResourceService(appSetting)
	bts := build_task.NewBuildTaskService(appSetting)
	ds := dataset.NewDatasetService(appSetting)
	dts := discover_task.NewDiscoverTaskService(appSetting)
	dss := discover_schedule.NewDiscoverScheduleService(appSetting, dts)

	return &restHandler{
		appSetting: appSetting,
		as:         auth.NewAuthService(appSetting),
		cs:         cs,
		rs:         rs,
		bts:        bts,
		ds:         ds,
		cts:        connector_type.NewConnectorTypeService(appSetting),
		dts:        dts,
		dss:        dss,
		scheduler:  scheduler,
		rds:        resource_data.NewResourceDataService(appSetting),
	}
}

// RegisterPublic registers public API routes.
func (r *restHandler) RegisterPublic(engine *gin.Engine) {
	engine.Use(r.accessLog())
	engine.Use(middleware.TracingMiddleware())

	engine.GET("/health", r.HealthCheck)

	// 外部 API (External)
	apiV1 := engine.Group("/api/vega-backend/v1")
	{
		// Catalog APIs - External
		catalogs := apiV1.Group("/catalogs")
		{
			catalogs.GET("", r.ListCatalogsByEx)
			catalogs.POST("", r.verifyJsonContentType(), r.CreateCatalogByEx)
			catalogs.GET("/:ids", r.GetCatalogsByEx)
			catalogs.PUT("/:id", r.verifyJsonContentType(), r.UpdateCatalogByEx)
			catalogs.DELETE("/:ids", r.DeleteCatalogsByEx)
			catalogs.GET("/:ids/health-status", r.GetCatalogHealthStatusByEx)
			catalogs.POST("/:id/test-connection", r.TestConnectionByEx)

			// 资源发现
			catalogs.POST("/:id/discover", r.DiscoverCatalogResourcesByEx)

		}

		// DiscoverTask APIs - External
		discoverTasks := apiV1.Group("/discover-tasks")
		{
			discoverTasks.GET("", r.ListDiscoverTasksByEx)
			discoverTasks.GET("/:id", r.GetDiscoverTaskByEx)
			discoverTasks.DELETE("/:ids", r.DeleteDiscoverTasksByEx)
		}

		// DiscoverSchedule APIs - External
		discoverSchedules := apiV1.Group("/discover-schedules")
		{
			discoverSchedules.POST("", r.verifyJsonContentType(), r.CreateDiscoverScheduleByEx)
			discoverSchedules.GET("", r.ListDiscoverSchedulesByEx)
			discoverSchedules.GET("/:id", r.GetDiscoverScheduleByEx)
			discoverSchedules.PUT("/:id", r.verifyJsonContentType(), r.UpdateDiscoverScheduleByEx)
			discoverSchedules.DELETE("/:id", r.DeleteDiscoverScheduleByEx)
			discoverSchedules.POST("/:id/enable", r.EnableDiscoverScheduleByEx)
			discoverSchedules.POST("/:id/disable", r.DisableDiscoverScheduleByEx)
		}

		// Resource APIs - External
		resources := apiV1.Group("/resources")
		{
			resources.GET("", r.ListResourcesByEx)
			resources.POST("", r.verifyJsonContentType(), r.CreateResourceByEx)
			resources.GET("/:id", r.GetResourcesByEx) // id为资源ID，多个资源ID逗号分隔
			resources.PUT("/:id", r.verifyJsonContentType(), r.UpdateResourceByEx)
			resources.DELETE("/:id", r.DeleteResourcesByEx) // id为资源ID，多个资源ID逗号分隔

			resources.POST("/:id/data", r.verifyJsonContentType(), r.PostResourceDataByEx)
			resources.PUT("/:id/data", r.verifyJsonContentType(), r.PutResourceDataByEx)
			resources.GET("/:id/data/:doc_id", r.GetResourceDataDocByEx)
			resources.PUT("/:id/data/:doc_id", r.verifyJsonContentType(), r.PutResourceDataDocByEx)
			resources.DELETE("/:id/data/:doc_ids", r.DeleteResourceDataByEx)

			resources.POST("/query", r.verifyJsonContentType(), r.RawQueryByEx)
		}

		// BuildTask APIs - External
		buildTasks := apiV1.Group("/build-tasks")
		{
			buildTasks.POST("", r.verifyJsonContentType(), r.CreateBuildTaskByEx)
			buildTasks.GET("", r.ListBuildTasksByEx)
			buildTasks.GET("/:id", r.GetBuildTaskByEx)
			buildTasks.DELETE("/:ids", r.DeleteBuildTasksByEx)
			buildTasks.POST("/:id/start", r.StartBuildTaskByEx)
			buildTasks.POST("/:id/stop", r.StopBuildTaskByEx)
		}

		// ConnectorType APIs - External
		connectorTypes := apiV1.Group("/connector-types")
		{
			connectorTypes.GET("", r.ListConnectorTypes)
			connectorTypes.POST("", r.verifyJsonContentType(), r.RegisterConnectorType)
			connectorTypes.GET("/:type", r.GetConnectorType)
			connectorTypes.PUT("/:type", r.verifyJsonContentType(), r.UpdateConnectorType)
			connectorTypes.DELETE("/:type", r.DeleteConnectorType)
			connectorTypes.POST("/:type/enable", r.EnableConnectorType)
			connectorTypes.POST("/:type/disable", r.DisableConnectorType)
		}
	}

	// 内部 API (Internal)
	apiInV1 := engine.Group("/api/vega-backend/in/v1")
	{
		// Catalog APIs - Internal
		catalogs := apiInV1.Group("/catalogs")
		{
			catalogs.GET("", r.ListCatalogsByIn)
			catalogs.POST("", r.verifyJsonContentType(), r.CreateCatalogByIn)
			catalogs.GET("/:ids", r.GetCatalogsByIn)
			catalogs.PUT("/:id", r.verifyJsonContentType(), r.UpdateCatalogByIn)
			catalogs.DELETE("/:ids", r.DeleteCatalogsByIn)
			catalogs.GET("/:ids/health-status", r.GetCatalogHealthStatusByIn)
			catalogs.POST("/:id/test-connection", r.TestConnectionByIn)

			//
			catalogs.POST("/:id/discover", r.DiscoverCatalogResourcesByIn)

		}

		// DiscoverTask APIs - Internal
		discoverTasks := apiInV1.Group("/discover-tasks")
		{
			discoverTasks.GET("", r.ListDiscoverTasksByIn)
			discoverTasks.GET("/:id", r.GetDiscoverTaskByIn)
			discoverTasks.DELETE("/:ids", r.DeleteDiscoverTasksByIn)
		}

		// DiscoverSchedule APIs - Internal
		discoverSchedules := apiInV1.Group("/discover-schedules")
		{
			discoverSchedules.POST("", r.verifyJsonContentType(), r.CreateDiscoverScheduleByIn)
			discoverSchedules.GET("", r.ListDiscoverSchedulesByIn)
			discoverSchedules.GET("/:id", r.GetDiscoverScheduleByIn)
			discoverSchedules.PUT("/:id", r.verifyJsonContentType(), r.UpdateDiscoverScheduleByIn)
			discoverSchedules.DELETE("/:id", r.DeleteDiscoverScheduleByIn)
			discoverSchedules.POST("/:id/enable", r.EnableDiscoverScheduleByIn)
			discoverSchedules.POST("/:id/disable", r.DisableDiscoverScheduleByIn)
		}

		// Resource APIs - Internal
		resources := apiInV1.Group("/resources")
		{
			resources.GET("", r.ListResourcesByIn)
			resources.POST("", r.verifyJsonContentType(), r.CreateResourceByIn)
			resources.GET("/:id", r.GetResourcesByIn) // id为资源ID，多个资源ID逗号分隔
			resources.PUT("/:id", r.verifyJsonContentType(), r.UpdateResourceByIn)
			resources.DELETE("/:id", r.DeleteResourcesByIn) // id为资源ID，多个资源ID逗号分隔

			resources.POST("/:id/data", r.verifyJsonContentType(), r.PostResourceDataByIn)
			resources.PUT("/:id/data", r.verifyJsonContentType(), r.PutResourceDataByIn)
			resources.GET("/:id/data/:doc_id", r.GetResourceDataDocByIn)
			resources.PUT("/:id/data/:doc_id", r.verifyJsonContentType(), r.PutResourceDataDocByIn)
			resources.DELETE("/:id/data/:doc_ids", r.DeleteResourceDataByIn)

			resources.POST("/query", r.verifyJsonContentType(), r.RawQueryByIn)
		}

		// BuildTask APIs - Internal
		buildTasks := apiInV1.Group("/build-tasks")
		{
			buildTasks.POST("", r.verifyJsonContentType(), r.CreateBuildTaskByIn)
			buildTasks.GET("", r.ListBuildTasksByIn)
			buildTasks.GET("/:id", r.GetBuildTaskByIn)
			buildTasks.DELETE("/:ids", r.DeleteBuildTasksByIn)
			buildTasks.POST("/:id/start", r.StartBuildTaskByIn)
			buildTasks.POST("/:id/stop", r.StopBuildTaskByIn)
		}
	}

	logger.Info("RestHandler RegisterPublic")
}

// HealthCheck 健康检查
func (r *restHandler) HealthCheck(c *gin.Context) {
	// 返回服务信息
	serverInfo := o11y.ServerInfo{
		ServerName:    version.ServerName,
		ServerVersion: version.ServerVersion,
		Language:      version.LanguageGo,
		GoVersion:     version.GoVersion,
		GoArch:        version.GoArch,
	}
	rest.ReplyOK(c, http.StatusOK, serverInfo)
}

// verifyJsonContentType middleware
func (r *restHandler) verifyJsonContentType() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, "VegaBackend.InvalidRequestHeader.ContentType")
			rest.ReplyError(c, httpErr)
			c.Abort()
			return
		}
		c.Next()
	}
}

// accessLog middleware
func (r *restHandler) accessLog() gin.HandlerFunc {
	return func(c *gin.Context) {
		beginTime := time.Now()
		c.Next()
		endTime := time.Now()
		durTime := endTime.Sub(beginTime).Seconds()

		logger.Debugf("access log: url: %s, method: %s, begin_time: %s, end_time: %s, subTime: %f",
			c.Request.URL.Path,
			c.Request.Method,
			beginTime.Format(libCommon.RFC3339Milli),
			endTime.Format(libCommon.RFC3339Milli),
			durTime,
		)
	}
}

// verifyOAuth verifies OAuth token
func (r *restHandler) verifyOAuth(ctx context.Context, c *gin.Context) (hydra.Visitor, error) {
	visitor, err := r.as.VerifyToken(ctx, c)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusUnauthorized, rest.PublicError_Unauthorized).
			WithErrorDetails(err.Error())
		rest.ReplyError(c, httpErr)
		return visitor, err
	}

	return visitor, nil
}

// GenerateVisitor generates visitor from request headers (for internal APIs)
func GenerateVisitor(c *gin.Context) hydra.Visitor {
	accountInfo := interfaces.AccountInfo{
		ID:   c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_ID),
		Type: c.GetHeader(interfaces.HTTP_HEADER_ACCOUNT_TYPE),
	}

	visitor := hydra.Visitor{
		ID:         accountInfo.ID,
		Type:       hydra.VisitorType(accountInfo.Type),
		TokenID:    "", // 无token
		IP:         c.ClientIP(),
		Mac:        c.GetHeader("X-Request-MAC"),
		UserAgent:  c.GetHeader("User-Agent"),
		ClientType: hydra.ClientType_Linux,
	}

	return visitor
}
