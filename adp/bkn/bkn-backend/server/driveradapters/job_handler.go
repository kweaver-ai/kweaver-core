// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	berrors "bkn-backend/errors"
	"bkn-backend/infra/otel/otellog"
	"bkn-backend/infra/otel/oteltrace"
	"bkn-backend/interfaces"
)

// CreateJobByEx 创建 job
func (r *restHandler) CreateJobByEx(c *gin.Context) {
	logger.Debug("Handler CreateJobByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateJob(c, visitor)
}

// CreateJobByIn 创建 job
func (r *restHandler) CreateJobByIn(c *gin.Context) {
	logger.Debug("Handler CreateJobByIn Start")
	_, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 内部接口 account_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.CreateJob(c, visitor)
}

// CreateJobByEx 创建 job
func (r *restHandler) CreateJob(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CreateJob Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// userId 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// 校验业务知识网络存在性
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数
	reqBody := interfaces.JobInfo{}
	err = c.ShouldBindJSON(&reqBody)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_Job_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	reqBody.KNID = knID
	reqBody.Branch = branch

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("创建行动类请求参数: [%s,%v]", c.Request.RequestURI, reqBody))

	// 1. 校验 创建任务必要参数的合法性, 非空、长度、是枚举值
	err = ValidateJob(ctx, &reqBody)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("Validate job[%s] failed: %s. %v", reqBody.Name,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("job_name").String(reqBody.Name))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//调用创建 job 接口
	jobID, err := r.js.CreateJob(ctx, &reqBody)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//每次成功创建 记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateJobAuditObject(jobID, reqBody.Name), "")

	result := map[string]any{"id": jobID}

	logger.Debug("Handler CreateJob Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// DeleteJobsByEx 批量删除 job
func (r *restHandler) DeleteJobsByEx(c *gin.Context) {
	logger.Debug("Handler DeleteJobsByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteJobs(c, visitor)
}

// DeleteJobsByIn 批量删除 job
func (r *restHandler) DeleteJobsByIn(c *gin.Context) {
	logger.Debug("Handler DeleteJobsByIn Start")
	_, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 内部接口 account_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.DeleteJobs(c, visitor)
}

// DeleteJobs 批量删除 job
func (r *restHandler) DeleteJobs(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler DeleteJobs Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// userId 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("批量删除job请求参数: [%s]", c.Request.RequestURI))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// 校验业务知识网络存在性
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//获取参数字符串 <id1,id2,id3>
	jobIDsStr := c.Param("job_ids")
	span.SetAttributes(attr.Key("job_ids").String(jobIDsStr))

	//解析字符串 转换为 []string
	jobIDs := common.StringToStringSlice(jobIDsStr)

	//检查 jobIDs 是否都存在
	jobs, err := r.js.GetJobsByIDs(ctx, jobIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if len(jobIDs) != len(jobs) {
		for _, jobID := range jobIDs {
			if _, exist := jobs[jobID]; !exist {
				httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_Job_JobNotFound)
				// 设置 trace 的错误信息的 attributes
				oteltrace.AddHttpAttrs4HttpError(span, httpErr)
				rest.ReplyError(c, httpErr)
				return
			}
		}
	}
	for _, job := range jobs {
		// 校验任务是否在当前业务知识网络下
		if job.KNID != knID || job.Branch != branch {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_Job_JobNotFound)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if job.State == interfaces.JobStateRunning {
			httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
				berrors.BknBackend_Job_JobRunning)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 批量删除 job
	err = r.js.DeleteJobsByIDs(ctx, knID, branch, jobIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	for _, job := range jobs {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateJobAuditObject(job.ID, job.Name), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteJobs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// ListJobsByEx 列出所有 job
func (r *restHandler) ListJobsByEx(c *gin.Context) {
	logger.Debug("Handler ListJobsByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListJobs(c, visitor)
}

// ListJobsByIn 列出所有 job
func (r *restHandler) ListJobsByIn(c *gin.Context) {
	logger.Debug("Handler ListJobsByIn Start")
	_, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 内部接口 account_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.ListJobs(c, visitor)
}

// ListJobs 列出所有 job
func (r *restHandler) ListJobs(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ListJobs Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// userId 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("列出所有job请求参数: [%s]", c.Request.RequestURI))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// 校验业务知识网络存在性
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取分页参数
	namePattern := c.Query("name_pattern")
	jobType := c.Query("job_type")
	stateArr := c.QueryArray("state")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "create_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.JOB_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验 job_type 参数
	if jobType != "" {
		err = ValidateJobType(ctx, interfaces.JobType(jobType))
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 校验 state 参数
	state := []interfaces.JobState{}
	for _, stateStr := range stateArr {
		err = ValidateJobState(ctx, interfaces.JobState(stateStr))
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		state = append(state, interfaces.JobState(stateStr))
	}

	// 构造标签列表查询参数的结构体
	queryParams := interfaces.JobsQueryParams{
		NamePattern: namePattern,
		KNID:        knID,
		Branch:      branch,
		JobType:     interfaces.JobType(jobType),
		State:       state,
	}
	queryParams.Sort = pageParam.Sort
	queryParams.Direction = pageParam.Direction
	queryParams.Limit = pageParam.Limit
	queryParams.Offset = pageParam.Offset

	jobs, total, err := r.js.ListJobs(ctx, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     jobs,
		"total_count": total,
	}
	logger.Debug("Handler ListJobs Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// ListTasksByEx 列出指定 job 的子任务
func (r *restHandler) ListTasksByEx(c *gin.Context) {
	logger.Debug("Handler ListTasksByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListTasks(c, visitor)
}

// ListTasksByIn 列出指定 job 的子任务
func (r *restHandler) ListTasksByIn(c *gin.Context) {
	logger.Debug("Handler ListTasksByIn Start")
	_, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 内部接口 account_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.ListTasks(c, visitor)
}

// ListTasks 列出指定 job 的子任务
func (r *restHandler) ListTasks(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ListTasks Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// userId 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

	// 校验业务知识网络存在性
	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	jobID := c.Param("job_id")
	span.SetAttributes(attr.Key("job_id").String(jobID))

	job, err := r.js.GetJobByID(ctx, jobID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if job == nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_Job_JobNotFound)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取分页参数
	conceptType := c.Query("concept_type")
	NamePattern := c.Query("name_pattern")
	stateArr := c.QueryArray("state")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "start_time")
	direction := c.DefaultQuery("direction", interfaces.ASC_DIRECTION)

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.TASK_SORT)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if conceptType != "" {
		err = ValidateConceptType(ctx, conceptType)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	state := []interfaces.TaskState{}
	for _, stateStr := range stateArr {
		err = ValidateTaskState(ctx, interfaces.TaskState(stateStr))
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		state = append(state, interfaces.TaskState(stateStr))
	}

	// 构造标签列表查询参数的结构体
	queryParams := interfaces.TasksQueryParams{
		KNID:        knID,
		Branch:      job.Branch,
		JobID:       jobID,
		ConceptType: conceptType,
		NamePattern: NamePattern,
		State:       state,
	}
	queryParams.Sort = pageParam.Sort
	queryParams.Direction = pageParam.Direction
	queryParams.Limit = pageParam.Limit
	queryParams.Offset = pageParam.Offset

	tasks, total, err := r.js.ListTasks(ctx, queryParams)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	result := map[string]any{
		"entries":     tasks,
		"total_count": total,
	}
	logger.Debug("Handler ListTasks Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}
