package driveradapters

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
	"flow-stream-data-pipeline/pipeline-mgmt/logics/auth"
	"flow-stream-data-pipeline/pipeline-mgmt/logics/pipeline"
	"flow-stream-data-pipeline/version"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting          *common.AppSetting
	as                  interfaces.AuthService
	pipelineMgmtService interfaces.PipelineMgmtService
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	return &restHandler{
		appSetting:          appSetting,
		as:                  auth.NewAuthService(appSetting),
		pipelineMgmtService: pipeline.NewPipelineMgmtService(appSetting),
	}
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(r.AccessLog())
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	apiV1 := c.Group("/api/flow-stream-data-pipeline/v1")
	{
		apiV1.GET("/pipelines/:id", r.GetPipelineByEx)
		apiV1.GET("/pipelines", r.ListPipelinesByEx)
		apiV1.PUT("/pipelines/:id", r.UpdatePipelineByEx)
		apiV1.POST("/pipelines", r.CreatePipelineByEx)
		apiV1.DELETE("/pipelines/:id", r.DeletePipelineByEx)
		apiV1.PUT("/pipelines/:id/attrs/:fields", r.UpdatePipelineStatusByEx)

		apiV1.GET("/resources", r.ListResources)
	}

	apiInV1 := c.Group("/api/flow-stream-data-pipeline/in/v1")
	{
		apiInV1.GET("/pipelines/:id", r.GetPipelineByIn)
		apiInV1.GET("/pipelines", r.ListPipelinesByIn)
		apiInV1.PUT("/pipelines/:id", r.UpdatePipelineByIn)
		apiInV1.POST("/pipelines", r.CreatePipelineByIn)
		apiInV1.DELETE("/pipelines/:id", r.DeletePipelineByIn)
		apiInV1.PUT("/pipelines/:id/attrs/:fields", r.UpdatePipelineStatusByIn)
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

// gin中间件 访问日志
func (r *restHandler) AccessLog() gin.HandlerFunc {
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

// 校验oauth
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
