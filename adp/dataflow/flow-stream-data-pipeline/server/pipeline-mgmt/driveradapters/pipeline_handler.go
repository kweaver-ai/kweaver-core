package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"flow-stream-data-pipeline/common"
	"flow-stream-data-pipeline/common/visitor"
	serrors "flow-stream-data-pipeline/errors"
	"flow-stream-data-pipeline/pipeline-mgmt/interfaces"
)

func (r *restHandler) CreatePipelineByEx(c *gin.Context) {
	logger.Debug("Handler CreatePipelineByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreatePipeline(c, visitor)
}

func (r *restHandler) CreatePipelineByIn(c *gin.Context) {
	logger.Debug("Handler CreatePipelineByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.CreatePipeline(c, visitor)
}

// CreatePipeline 创建管道
func (r *restHandler) CreatePipeline(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CreatePipeline Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Create pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置与API相关的Attributes
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接收绑定参数
	pipelineInfo := interfaces.Pipeline{}
	err := c.ShouldBindJSON(&pipelineInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.ValidatePipeline(ctx, &pipelineInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 管道总数是否超过配置数，默认100
	pipelineQuery := &interfaces.ListPipelinesQuery{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Limit:  interfaces.NO_LIMIT,
			Offset: interfaces.MIN_OFFSET,
		},
	}
	total, err := r.pipelineMgmtService.GetPipelineTotals(ctx, pipelineQuery)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if total >= r.appSetting.ServerSetting.MaxPipelineCount {
		logger.Errorf("The number of pipelines has reached the upper limit %d", r.appSetting.ServerSetting.MaxPipelineCount)
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_CountExceeded_Pipelines).
			WithErrorDetails(fmt.Sprintf("The number of pipelines has reached the upper limit %d", r.appSetting.ServerSetting.MaxPipelineCount))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if pipelineInfo.PipelineID != "" {
		// 校验 id 是否已经存在
		_, exist, err := r.pipelineMgmtService.CheckPipelineExistByID(ctx, pipelineInfo.PipelineID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		if exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, serrors.StreamDataPipeline_Duplicated_PipelineID).
				WithErrorDetails("pipeline id exist")

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 校验名称是否已经存在
	_, exist, err := r.pipelineMgmtService.CheckPipelineExistByName(ctx, pipelineInfo.PipelineName)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, serrors.StreamDataPipeline_Duplicated_PipelineName).
			WithErrorDetails("pipeline name exist")
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 创建任务
	pipelineId, err := r.pipelineMgmtService.CreatePipeline(ctx, &pipelineInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]string{"id": pipelineId}
	c.Writer.Header().Set("Location", "/api/flow-stream-data-pipeline/v1/pipelines/"+pipelineId)

	logger.Debug("CreatePipeline Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusCreated)
	rest.ReplyOK(c, http.StatusCreated, result)
}

func (r *restHandler) DeletePipelineByEx(c *gin.Context) {
	logger.Debug("Handler DeletePipelineByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.DeletePipeline(c, visitor)
}

func (r *restHandler) DeletePipelineByIn(c *gin.Context) {
	logger.Debug("Handler DeletePipelineByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.DeletePipeline(c, visitor)
}

// Delete 删除管道
func (r *restHandler) DeletePipeline(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("DeletePipeline")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 接收pipeline_id参数
	pipelineID := c.Param("id")

	// 校验管道是否存在
	_, exist, err := r.pipelineMgmtService.CheckPipelineExistByID(ctx, pipelineID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, serrors.StreamDataPipeline_NotFound_Pipeline)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.pipelineMgmtService.DeletePipeline(ctx, pipelineID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("DeletePipeline Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) UpdatePipelineByEx(c *gin.Context) {
	logger.Debug("Handler UpdatePipelineByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Delete pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdatePipeline(c, visitor)
}

func (r *restHandler) UpdatePipelineByIn(c *gin.Context) {
	logger.Debug("Handler UpdatePipelineByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.UpdatePipeline(c, visitor)
}

// UpdatePipeline 修改管道配置
func (r *restHandler) UpdatePipeline(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler UpdatePipeline Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update pipeline", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	pipelineID := c.Param("id")

	// 接收绑定参数
	pipelineInfo := interfaces.Pipeline{}
	err := c.ShouldBindJSON(&pipelineInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	pipelineInfo.PipelineID = pipelineID

	err = r.ValidatePipeline(ctx, &pipelineInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	err = r.pipelineMgmtService.UpdatePipeline(ctx, &pipelineInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("UpdatePipeline Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

func (r *restHandler) GetPipelineByEx(c *gin.Context) {
	logger.Debug("Handler GetPipelineByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get pipeline by id", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetPipeline(c, visitor)
}

func (r *restHandler) GetPipelineByIn(c *gin.Context) {
	logger.Debug("Handler GetPipelineByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetPipeline(c, visitor)
}

// GetPipeline 根据 pipeline_id 获取任务详情
func (r *restHandler) GetPipeline(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetPipeline Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Get pipeline by id", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 获取 pipeline_id 参数
	pipelineID := c.Param("id")

	// 获取 is_listen 参数
	isListenStr := c.DefaultQuery("is_listen", "false")

	isListen, _ := strconv.ParseBool(isListenStr)

	pipelineInfo, _, err := r.pipelineMgmtService.GetPipeline(ctx, pipelineID, isListen)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	logger.Debug("GetPipeline Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, pipelineInfo)
}

func (r *restHandler) ListPipelinesByEx(c *gin.Context) {
	logger.Debug("Handler ListPipelinesByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List pipelines", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListPipelines(c, visitor)
}

func (r *restHandler) ListPipelinesByIn(c *gin.Context) {
	logger.Debug("Handler ListPipelinesByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.ListPipelines(c, visitor)
}

// ListPipelines 获取管道列表
func (r *restHandler) ListPipelines(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Hnadler ListPipelines Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List pipelines", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	//接收参数
	// name := c.Query("name")
	namePattern := c.Query("name_pattern")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", interfaces.DEFAULT_SORT)
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")

	// httpErr := validateNameandNamePattern(ctx, name, namePattern)
	// if httpErr != nil {
	// 	httpErr := httpErr.(*rest.HTTPError)
	// 	rest.ReplyError(c, httpErr)
	// 	return
	// }

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	pipelineStatus := common.StringToStringArr(c.DefaultQuery("status", ""))
	// 去掉标签前后的所有空格进行搜索, 因为存储的时候去掉了空格
	tag := strings.Trim(c.Query("tag"), " ")

	// 分页参数校验
	pq, err := ValidateListPipelinesQuery(ctx, offset, limit, sort, direction)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// pipelineStatus参数校验
	for _, val := range pipelineStatus {
		if val != interfaces.PipelineStatus_Error && val != interfaces.PipelineStatus_Running &&
			val != interfaces.PipelineStatus_Close {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_PipelineStatus).
				WithErrorDetails("the value of pipeline status is not failed, running or close")

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 去掉标签前后的所有空格进行搜索
	ListPipelinesQuery := interfaces.ListPipelinesQuery{
		// Name:                      name,
		NamePattern:               namePattern,
		PipelineStatus:            pipelineStatus,
		Tag:                       tag,
		PaginationQueryParameters: pq,
		Builtin:                   builtinArr,
	}

	// 分页获取任务列表
	pipelineList, total, err := r.pipelineMgmtService.ListPipelines(ctx, &ListPipelinesQuery)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{"entries": pipelineList, "total_count": total}
	logger.Debug("ListPipelines Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

func (r *restHandler) UpdatePipelineStatusByEx(c *gin.Context) {
	logger.Debug("Handler UpdatePipelineStatusByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update pipeline status", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdatePipelineStatus(c, visitor)
}

func (r *restHandler) UpdatePipelineStatusByIn(c *gin.Context) {
	logger.Debug("Handler UpdatePipelineStatusByIn Start")

	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.UpdatePipelineStatus(c, visitor)
}

// UpdatePipelineStatus 根据pipeline_id修改任务状态
func (r *restHandler) UpdatePipelineStatus(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler UpdatePipelineStatus Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: Update pipeline status", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 动作
	// op := "update"

	// 接收pipeline_id
	pipelineID := c.Param("id")

	// 接收参数
	isInnerRequest, _ := strconv.ParseBool(c.DefaultQuery("is_inner_request", "false"))

	_, exist, err := r.pipelineMgmtService.CheckPipelineExistByID(ctx, pipelineID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, serrors.StreamDataPipeline_NotFound_Pipeline)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//接收绑定 status
	pipelineStatusInfo := interfaces.PipelineStatusParamter{}
	err = c.ShouldBindJSON(&pipelineStatusInfo)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		// if !isInnerRequest {
		// 发审计日志
		// }

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验 status参数
	err = ValidatePipelineStatusInfo(ctx, pipelineStatusInfo)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// if !isInnerRequest {
		// 	// 发审计日志
		// 	// errorCode := httpErr.(*rest.HTTPError).BaseError.ErrorCode
		// 	// audit.NewWarningAuditLog(c, audit.ACTION, op, interfaces.OBJECTTYPE, "", audit.FAILED, errorCode+".Description")
		// }
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// if pipelineStatusInfo.Status == interfaces.PipelineStatus_Paused {
	// 	// op = "pause"
	// } else if pipelineStatusInfo.Status == interfaces.PipelineStatus_Running {
	// 	// op = "start"
	// }

	// 修改任务状态
	err = r.pipelineMgmtService.UpdatePipelineStatus(ctx, pipelineID, &pipelineStatusInfo, isInnerRequest)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// if !isInnerRequest {
		// 	// errorCode := httpErr.(*rest.HTTPError).BaseError.ErrorCode
		// 	// audit.NewWarningAuditLog(c, audit.ACTION, op, interfaces.OBJECTTYPE, jobInfo.JobName, audit.FAILED, errorCode+".Description")

		// }

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// if !isInnerRequest {
	// 	// audit.NewInfoAuditLog(c, audit.ACTION, op, interfaces.OBJECTTYPE, jobInfo.JobName, audit.SUCCESS, "")
	// }

	logger.Debug("UpdatePipelineStatus Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusNoContent)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ListPipelines 获取管道列表
func (r *restHandler) ListPipelineSources(c *gin.Context) {
	logger.Debug("Hnadler ListPipelineSrcs Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c),
		"driver layer: List pipeline sources", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	//接收参数
	// name := c.Query("name")
	namePattern := c.Query("keyword")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.RESOURCES_PAGE_LIMIT)
	sort := c.DefaultQuery("sort", "name")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)
	builtinQueryArr := c.QueryArray("builtin")

	// 校验builtin参数
	builtinArr := make([]bool, 0, len(builtinQueryArr))
	for _, val := range builtinQueryArr {
		if val == "" {
			continue
		}
		builtin, err := strconv.ParseBool(val)
		if err != nil {
			errDetails := fmt.Sprintf(`The value of param 'builtin' should be bool type, but got '%s'`, val)
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_Builtin).
				WithErrorDetails(errDetails)

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		builtinArr = append(builtinArr, builtin)
	}

	pipelineStatus := common.StringToStringArr(c.DefaultQuery("status", ""))
	// 去掉标签前后的所有空格进行搜索, 因为存储的时候去掉了空格
	tag := strings.Trim(c.Query("tag"), " ")

	// 分页参数校验
	pq, err := ValidateListPipelinesQuery(ctx, offset, limit, sort, direction)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// pipelineStatus参数校验
	for _, val := range pipelineStatus {
		if val != interfaces.PipelineStatus_Error && val != interfaces.PipelineStatus_Running &&
			val != interfaces.PipelineStatus_Close {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, serrors.StreamDataPipeline_InvalidParameter_PipelineStatus).
				WithErrorDetails("the value of pipeline status is not failed, running or close")

			o11y.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 去掉标签前后的所有空格进行搜索
	ListPipelinesQuery := interfaces.ListPipelinesQuery{
		// Name:                      name,
		NamePattern:               namePattern,
		PipelineStatus:            pipelineStatus,
		Tag:                       tag,
		PaginationQueryParameters: pq,
		Builtin:                   builtinArr,
	}

	// 分页获取任务列表
	resources, total, err := r.pipelineMgmtService.ListPipelineResources(ctx, &ListPipelinesQuery)

	if err != nil {
		httpErr := err.(*rest.HTTPError)
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]interface{}{"entries": resources, "total_count": total}

	logger.Debug("List pipeline resources Success")
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
