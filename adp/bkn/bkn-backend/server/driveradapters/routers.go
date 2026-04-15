// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	libCommon "github.com/kweaver-ai/kweaver-go-lib/common"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/middleware"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"bkn-backend/common"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
	"bkn-backend/logics/action_schedule"
	"bkn-backend/logics/action_type"
	"bkn-backend/logics/auth"
	"bkn-backend/logics/bkn"
	"bkn-backend/logics/concept_group"
	"bkn-backend/logics/job"
	"bkn-backend/logics/knowledge_network"
	metriclogics "bkn-backend/logics/metric"
	"bkn-backend/logics/object_type"
	"bkn-backend/logics/relation_type"
	"bkn-backend/logics/risk_type"
	"bkn-backend/version"
)

type RestHandler interface {
	RegisterPublic(engine *gin.Engine)
}

type restHandler struct {
	appSetting *common.AppSetting
	as         interfaces.AuthService
	ass        interfaces.ActionScheduleService
	ats        interfaces.ActionTypeService
	cgs        interfaces.ConceptGroupService
	js         interfaces.JobService
	kns        interfaces.KNService
	ots        interfaces.ObjectTypeService
	rts        interfaces.RelationTypeService
	rtsRisk    interfaces.RiskTypeService
	ms         interfaces.MetricService
	bs         interfaces.BKNService
}

func NewRestHandler(appSetting *common.AppSetting) RestHandler {
	r := &restHandler{
		appSetting: appSetting,
		as:         auth.NewAuthService(appSetting),
		ass:        action_schedule.NewActionScheduleService(appSetting),
		ats:        action_type.NewActionTypeService(appSetting),
		cgs:        concept_group.NewConceptGroupService(appSetting),
		js:         job.NewJobService(appSetting),
		kns:        knowledge_network.NewKNService(appSetting),
		ots:        object_type.NewObjectTypeService(appSetting),
		rts:        relation_type.NewRelationTypeService(appSetting),
		rtsRisk:    risk_type.NewRiskTypeService(appSetting),
		ms:         metriclogics.NewMetricService(appSetting),
		bs:         bkn.NewBKNService(appSetting),
	}
	return r
}

func (r *restHandler) RegisterPublic(c *gin.Engine) {
	c.Use(r.AccessLog())
	c.Use(middleware.TracingMiddleware())

	c.GET("/health", r.HealthCheck)

	bknApiV1 := c.Group("/api/bkn-backend/v1")
	otlApiV1 := c.Group("/api/ontology-manager/v1")

	for _, apiV1 := range []*gin.RouterGroup{bknApiV1, otlApiV1} {
		// 业务知识网络
		apiV1.POST("/knowledge-networks", r.verifyJsonContentTypeMiddleWare(), r.CreateKNByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id", r.DeleteKN)
		apiV1.PUT("/knowledge-networks/:kn_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateKNByEx)
		apiV1.GET("/knowledge-networks", r.ListKNsByEx)
		apiV1.GET("/knowledge-networks/:kn_id", r.GetKNByEx)
		apiV1.POST("/knowledge-networks/:kn_id/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateKNByEx)
		apiV1.POST("/knowledge-networks/:kn_id/relation-type-paths", r.GetRelationTypePathsByEx)

		// 概念分组
		apiV1.POST("/knowledge-networks/:kn_id/concept-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateConceptGroupByEx)
		apiV1.POST("/knowledge-networks/:kn_id/concept-groups/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateConceptGroupsByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.DeleteConceptGroup) // 不支持批量删
		apiV1.PUT("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateConceptGroupByEx)
		apiV1.GET("/knowledge-networks/:kn_id/concept-groups", r.ListConceptGroupsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.GetConceptGroupByEx)
		apiV1.POST("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types", r.AddObjectTypesToConceptGroupByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types/:ot_ids", r.DeleteObjectTypesFromGroupByEx)

		// 对象类
		apiV1.POST("/knowledge-networks/:kn_id/object-types", r.verifyJsonContentTypeMiddleWare(), r.HandleObjectTypeGetOverrideByEx)
		apiV1.POST("/knowledge-networks/:kn_id/object-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateObjectTypesByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/object-types/:ot_ids", r.DeleteObjectTypes) // path上用kn_ids接，实际上只能传一个id
		apiV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectTypeByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id/data_properties/:property_names", r.verifyJsonContentTypeMiddleWare(), r.UpdateDataProperties)
		apiV1.GET("/knowledge-networks/:kn_id/object-types", r.ListObjectTypesByEx)        // path上用kn_ids接，实际上只能传一个id
		apiV1.GET("/knowledge-networks/:kn_id/object-types/:ot_ids", r.GetObjectTypesByEx) // path上用kn_ids接，实际上只能传一个id

		// 关系类
		apiV1.POST("/knowledge-networks/:kn_id/relation-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRelationTypeGetOverrideByEx)
		apiV1.POST("/knowledge-networks/:kn_id/relation-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateRelationTypesByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.DeleteRelationTypes)
		apiV1.PUT("/knowledge-networks/:kn_id/relation-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRelationTypeByEx)
		apiV1.GET("/knowledge-networks/:kn_id/relation-types", r.ListRelationTypesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.GetRelationTypesByEx)

		// 行动类
		apiV1.POST("/knowledge-networks/:kn_id/action-types", r.verifyJsonContentTypeMiddleWare(), r.HandleActionTypeGetOverrideByEx)
		apiV1.POST("/knowledge-networks/:kn_id/action-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateActionTypesByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/action-types/:at_ids", r.DeleteActionTypes)
		apiV1.PUT("/knowledge-networks/:kn_id/action-types/:at_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionTypeByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-types", r.ListActionTypesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-types/:at_ids", r.GetActionTypesByEx)

		// 指标
		apiV1.POST("/knowledge-networks/:kn_id/metrics", r.verifyJsonContentTypeMiddleWare(), r.HandleMetricGetOverrideByEx)
		apiV1.POST("/knowledge-networks/:kn_id/metrics/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateMetricsByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/metrics/:metric_ids", r.DeleteMetricsByIDsByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/metrics/:metric_ids", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricByEx)
		apiV1.GET("/knowledge-networks/:kn_id/metrics", r.ListMetricsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/metrics/:metric_ids", r.GetMetricsByIDsByEx)

		// 风险类
		apiV1.POST("/knowledge-networks/:kn_id/risk-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRiskTypeGetOverrideByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/risk-types/:rt_ids", r.DeleteRiskTypes)
		apiV1.PUT("/knowledge-networks/:kn_id/risk-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRiskTypeByEx)
		apiV1.GET("/knowledge-networks/:kn_id/risk-types", r.ListRiskTypesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/risk-types/:rt_ids", r.GetRiskTypesByEx)

		// 任务管理
		apiV1.POST("/knowledge-networks/:kn_id/jobs", r.verifyJsonContentTypeMiddleWare(), r.CreateJobByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/jobs/:job_ids", r.DeleteJobsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/jobs", r.ListJobsByEx)
		apiV1.GET("/knowledge-networks/:kn_id/jobs/:job_id/tasks", r.ListTasksByEx)

		// 行动计划管理
		apiV1.POST("/knowledge-networks/:kn_id/action-schedules", r.verifyJsonContentTypeMiddleWare(), r.CreateActionScheduleByEx)
		apiV1.DELETE("/knowledge-networks/:kn_id/action-schedules/:schedule_ids", r.DeleteActionSchedulesByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleByEx)
		apiV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id/status", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleStatusByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-schedules", r.ListActionSchedulesByEx)
		apiV1.GET("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.GetActionScheduleByEx)

		// 业务知识网络资源示例列表
		apiV1.GET("/resources", r.ListResources)

		// BKN 导入导出 (RESTful 设计)
		apiV1.POST("/bkns", r.UploadBKN)         // 上传 BKN tar 包导入
		apiV1.GET("/bkns/:kn_id", r.DownloadBKN) // 下载 BKN tar 包导出
	}

	bknApiInV1 := c.Group("/api/bkn-backend/in/v1")
	otlApiInV1 := c.Group("/api/ontology-manager/in/v1")

	for _, apiInV1 := range []*gin.RouterGroup{bknApiInV1, otlApiInV1} {
		// 业务知识网络
		apiInV1.POST("/knowledge-networks", r.verifyJsonContentTypeMiddleWare(), r.CreateKNByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateKNByIn)
		apiInV1.GET("/knowledge-networks", r.ListKNsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id", r.GetKNByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateKNByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/relation-type-paths", r.GetRelationTypePathsByIn)

		// 概念分组
		apiInV1.POST("/knowledge-networks/:kn_id/concept-groups", r.verifyJsonContentTypeMiddleWare(), r.CreateConceptGroupByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/concept-groups/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateConceptGroupsByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateConceptGroupByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/concept-groups", r.ListConceptGroupsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/concept-groups/:cg_id", r.GetConceptGroupByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types", r.AddObjectTypesToConceptGroupByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/concept-groups/:cg_id/object-types/:ot_ids", r.DeleteObjectTypesFromGroupByIn)

		// 对象类
		apiInV1.POST("/knowledge-networks/:kn_id/object-types", r.verifyJsonContentTypeMiddleWare(), r.HandleObjectTypeGetOverrideByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/object-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateObjectTypesByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/object-types/:ot_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateObjectTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/object-types", r.ListObjectTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/object-types/:ot_ids", r.GetObjectTypesByIn) // path上用kn_ids接，实际上只能传一个id

		// 关系类
		apiInV1.POST("/knowledge-networks/:kn_id/relation-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRelationTypeGetOverrideByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/relation-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateRelationTypesByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/relation-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRelationTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/relation-types", r.ListRelationTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/relation-types/:rt_ids", r.GetRelationTypesByIn)

		// 行动类
		apiInV1.POST("/knowledge-networks/:kn_id/action-types", r.verifyJsonContentTypeMiddleWare(), r.HandleActionTypeGetOverrideByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/action-types/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateActionTypesByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-types/:at_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-types", r.ListActionTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-types/:at_ids", r.GetActionTypesByIn)

		// 指标（内部）
		apiInV1.POST("/knowledge-networks/:kn_id/metrics", r.verifyJsonContentTypeMiddleWare(), r.HandleMetricGetOverrideByIn)
		apiInV1.POST("/knowledge-networks/:kn_id/metrics/validation", r.verifyJsonContentTypeMiddleWare(), r.ValidateMetricsByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/metrics/:metric_ids", r.DeleteMetricsByIDsByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/metrics/:metric_ids", r.verifyJsonContentTypeMiddleWare(), r.UpdateMetricByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/metrics", r.ListMetricsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/metrics/:metric_ids", r.GetMetricsByIDsByIn)

		// 风险类（内部 API：GetRiskTypesByIn 支持 risk_type_ids 查询参数）
		apiInV1.POST("/knowledge-networks/:kn_id/risk-types", r.verifyJsonContentTypeMiddleWare(), r.HandleRiskTypeGetOverrideByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/risk-types/:rt_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateRiskTypeByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/risk-types", r.GetRiskTypesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/risk-types/:rt_ids", r.GetRiskTypesByInWithPath)
		apiInV1.DELETE("/knowledge-networks/:kn_id/risk-types/:rt_ids", r.DeleteRiskTypes)

		// 行动计划管理
		apiInV1.POST("/knowledge-networks/:kn_id/action-schedules", r.verifyJsonContentTypeMiddleWare(), r.CreateActionScheduleByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/action-schedules/:schedule_ids", r.DeleteActionSchedulesByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleByIn)
		apiInV1.PUT("/knowledge-networks/:kn_id/action-schedules/:schedule_id/status", r.verifyJsonContentTypeMiddleWare(), r.UpdateActionScheduleStatusByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-schedules", r.ListActionSchedulesByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/action-schedules/:schedule_id", r.GetActionScheduleByIn)

		// 任务管理
		apiInV1.POST("/knowledge-networks/:kn_id/jobs", r.verifyJsonContentTypeMiddleWare(), r.CreateJobByIn)
		apiInV1.DELETE("/knowledge-networks/:kn_id/jobs/:job_ids", r.DeleteJobsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/jobs", r.ListJobsByIn)
		apiInV1.GET("/knowledge-networks/:kn_id/jobs/:job_id/tasks", r.ListTasksByIn)
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

// gin中间件 校验content type
func (r *restHandler) verifyJsonContentTypeMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		//拦截请求，判断ContentType是否为XXX
		if c.ContentType() != interfaces.CONTENT_TYPE_JSON {
			httpErr := rest.NewHTTPError(c, http.StatusNotAcceptable, berrors.BknBackend_InvalidRequestHeader_ContentType).
				WithErrorDetails(fmt.Sprintf("Content-Type header [%s] is not supported, expected is [application/json].", c.ContentType()))
			rest.ReplyError(c, httpErr)

			c.Abort()
		}

		//执行后续操作
		c.Next()
	}
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
