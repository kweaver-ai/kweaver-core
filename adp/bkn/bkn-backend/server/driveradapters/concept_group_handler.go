// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package driveradapters

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kweaver-ai/kweaver-go-lib/audit"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"bkn-backend/common"
	"bkn-backend/common/visitor"
	berrors "bkn-backend/errors"
	"bkn-backend/interfaces"
)

// 创建概念分组(内部)
func (r *restHandler) CreateConceptGroupByIn(c *gin.Context) {
	logger.Debug("Handler CreateConceptGroupByIn Start")
	// 内部接口 user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.CreateConceptGroup(c, visitor)
}

// 创建概念分组（外部）
func (r *restHandler) CreateConceptGroupByEx(c *gin.Context) {
	logger.Debug("Handler CreateConceptGroupByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateConceptGroup(c, visitor)
}

// 创建概念分组
func (r *restHandler) CreateConceptGroup(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CreateConceptGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 导入模式
	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	httpErr := validateImportMode(ctx, mode)
	if httpErr != nil {
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Whether to validate dependencies, default true. Parse priority: strict_mode > validate_dependency (legacy) > true
	strictModeStr := c.Query(interfaces.QueryParam_StrictMode)
	if strictModeStr == "" {
		strictModeStr = c.Query("validate_dependency")
	}
	if strictModeStr == "" {
		strictModeStr = "true"
	}
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

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

	// 接受绑定参数 - 单个概念分组对象
	cg := interfaces.ConceptGroup{}
	err = c.ShouldBindJSON(&cg)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	cg.KNID = knID
	cg.Branch = branch

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("创建概念分组请求参数: [%s,%v]", c.Request.RequestURI, cg))

	// 校验导入模型时模块是否是概念分组
	if cg.ModuleType != "" && cg.ModuleType != interfaces.MODULE_TYPE_CONCEPT_GROUP {
		httpErr := rest.NewHTTPError(ctx, http.StatusForbidden, berrors.BknBackend_InvalidParameter_ModuleType).
			WithErrorDetails("Concept Group's Module Type is not 'concept_group'")

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 1. 校验 概念分组必要创建参数的合法性, 非空、长度、是枚举值
	err = ValidateConceptGroup(ctx, &cg)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("Validate concept group[%s] failed: %s. %v", cg.CGName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("cg_name").String(cg.CGName))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 若kn的对象类，关系类，行动类, 概念分组不为空，则应循环调用对象类、关系类、行动类, 概念分组的校验函数
	if len(cg.ObjectTypes) > 0 {
		err = ValidateObjectTypes(ctx, cg.KNID, cg.ObjectTypes, strictMode)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	if len(cg.RelationTypes) > 0 {
		err = ValidateRelationTypes(ctx, cg.KNID, cg.RelationTypes, strictMode)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	if len(cg.ActionTypes) > 0 {
		err = ValidateActionTypes(ctx, cg.KNID, cg.ActionTypes, strictMode)
		if err != nil {
			httpErr := err.(*rest.HTTPError)
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}

	// 调用创建概念分组
	cgID, err := r.cgs.CreateConceptGroup(ctx, nil, &cg, mode, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功创建记录审计日志
	audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
		interfaces.GenerateConceptGroupAuditObject(knID, cg.CGName), "")

	logger.Debug("Handler CreateConceptGroup Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, map[string]any{"id": cgID})
}

// ValidateConceptGroupsByIn 仅校验概念分组依赖存在性，不写库（内部）
func (r *restHandler) ValidateConceptGroupsByIn(c *gin.Context) {
	logger.Debug("Handler ValidateConceptGroupsByIn Start")
	v := visitor.GenerateVisitor(c)
	r.ValidateConceptGroups(c, v)
}

// ValidateConceptGroupsByEx 仅校验概念分组依赖存在性，不写库（外部）
func (r *restHandler) ValidateConceptGroupsByEx(c *gin.Context) {
	logger.Debug("Handler ValidateConceptGroupsByEx Start")
	ctx, _ := oteltrace.StartServerSpan(c)

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ValidateConceptGroups(c, visitor)
}

// ValidateConceptGroups 仅校验概念分组依赖存在性，不写库
func (r *restHandler) ValidateConceptGroups(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ValidateConceptGroups Start")
	ctx, _ := oteltrace.StartServerSpan(c)

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		rest.ReplyError(c, httpErr)
		return
	}

	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	if httpErr := validateImportMode(ctx, mode); httpErr != nil {
		rest.ReplyError(c, httpErr)
		return
	}

	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)

	_, exist, err := r.kns.CheckKNExistByID(ctx, knID, branch)
	if err != nil {
		rest.ReplyError(c, err.(*rest.HTTPError))
		return
	}
	if !exist {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_KnowledgeNetwork_NotFound))
		return
	}

	var requestData struct {
		Entries []*interfaces.ConceptGroup `json:"entries"`
	}
	if err = c.ShouldBindJSON(&requestData); err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: "+err.Error()))
		return
	}
	conceptGroups := requestData.Entries
	if len(conceptGroups) == 0 {
		rest.ReplyOK(c, http.StatusOK, map[string]bool{"valid": true})
		return
	}

	for _, cg := range conceptGroups {
		cg.KNID = knID
		cg.Branch = branch
	}

	for _, cg := range conceptGroups {
		if err = ValidateConceptGroup(ctx, cg); err != nil {
			rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": false, "detail": err.Error()})
			return
		}
	}
	if err = r.cgs.ValidateConceptGroups(ctx, knID, branch, conceptGroups, strictMode, nil, mode); err != nil {
		rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": false, "detail": err.Error()})
		return
	}
	rest.ReplyOK(c, http.StatusOK, map[string]bool{"valid": true})
}

// 更新概念分组(内部)
func (r *restHandler) UpdateConceptGroupByIn(c *gin.Context) {
	logger.Debug("Handler UpdateConceptGroupByIn Start")
	// 内部接口 user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.UpdateConceptGroup(c, visitor)
}

// 更新概念分组（外部）
func (r *restHandler) UpdateConceptGroupByEx(c *gin.Context) {
	logger.Debug("Handler UpdateConceptGroupByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateConceptGroup(c, visitor)
}

// 更新概念分组
func (r *restHandler) UpdateConceptGroup(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler UpdateConceptGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
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

	// Whether to validate dependencies, default true. Parse priority: strict_mode > validate_dependency (legacy) > true
	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	var exist bool
	_, exist, err = r.kns.CheckKNExistByID(ctx, knID, branch)
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

	// 1. 接受 at_id 参数
	cgID := c.Param("cg_id")
	span.SetAttributes(attr.Key("cg_id").String(cgID))

	//接收绑定参数
	cg := interfaces.ConceptGroup{}
	err = c.ShouldBindJSON(&cg)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	cg.CGID = cgID
	cg.KNID = knID
	cg.Branch = branch // 分组的 branch 从query参数中取

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("修改概念分组请求参数: [%s, %v]", c.Request.RequestURI, cg))

	// 先按id获取原对象.
	oldKNName, exist, err := r.cgs.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验 概念分组基本参数的合法性, 非空、长度、是枚举值
	err = ValidateConceptGroup(ctx, &cg)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("Validate concept group[%s] failed: %s. %v", cg.CGName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("kn_name").String(cg.CGName))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 名称或分组不同，校验新名称是否已存在
	ifNameModify := false
	if oldKNName != cg.CGName {
		ifNameModify = true
		_, exist, err = r.cgs.CheckConceptGroupExistByName(ctx, knID, branch, cg.CGName)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		if exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusForbidden,
				berrors.BknBackend_ConceptGroup_ConceptGroupNameExisted)

			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
	}
	cg.IfNameModify = ifNameModify

	//根据id修改信息
	err = r.cgs.UpdateConceptGroup(ctx, nil, &cg, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateConceptGroupAuditObject(knID, cg.CGName), "")

	logger.Debug("Handler UpdateConceptGroup Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量删除概念分组
func (r *restHandler) DeleteConceptGroup(c *gin.Context) {
	logger.Debug("Handler DeleteConceptGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
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

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("删除概念分组请求参数: [%s]", c.Request.RequestURI))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

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
	cgID := c.Param("cg_id")
	span.SetAttributes(attr.Key("cg_id").String(cgID))

	//检查 atIDs 是否都存在
	cgName, exist, err := r.cgs.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)

		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 批量删除概念分组
	err = r.cgs.DeleteConceptGroupByID(ctx, nil, knID, branch, cgID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
		interfaces.GenerateConceptGroupAuditObject(knID, cgName), audit.SUCCESS, "")

	logger.Debug("Handler DeleteConceptGroup Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取概念分组列表(内部)
func (r *restHandler) ListConceptGroupsByIn(c *gin.Context) {
	logger.Debug("Handler ListConceptGroupsByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.ListConceptGroups(c, visitor)
}

// 分页获取概念分组列表（外部）
func (r *restHandler) ListConceptGroupsByEx(c *gin.Context) {
	logger.Debug("Handler ListConceptGroupsByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListConceptGroups(c, visitor)
}

// 分页获取概念分组列表
func (r *restHandler) ListConceptGroups(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("ListConceptGroups Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("分页获取概念分组列表请求参数: [%s]", c.Request.RequestURI))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(
		attr.Key("kn_id").String(knID),
		attr.Key("branch").String(branch),
	)

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
	tag := c.Query("tag")
	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.KN_SORT)
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

	// 构造标签列表查询参数的结构体
	parameter := interfaces.ConceptGroupsQueryParams{
		NamePattern: namePattern,
		Tag:         tag,
		KNID:        knID,
		Branch:      branch,
	}
	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	// 获取概念分组简单信息
	knList, total, err := r.cgs.ListConceptGroups(ctx, parameter)
	result := map[string]any{"entries": knList, "total_count": total}
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

	logger.Debug("Handler ListConceptGroups Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取概念分组对象信息(内部)
func (r *restHandler) GetConceptGroupByIn(c *gin.Context) {
	logger.Debug("Handler GetKNByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetConceptGroup(c, visitor)
}

// 按 id 获取概念分组对象信息（外部）
func (r *restHandler) GetConceptGroupByEx(c *gin.Context) {
	logger.Debug("Handler GetKNByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetConceptGroup(c, visitor)
}

// 按 id 获取概念分组对象信息
func (r *restHandler) GetConceptGroup(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetConceptGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
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

	mode := c.DefaultQuery(interfaces.QueryParam_Mode, "")
	if mode != "" && mode != interfaces.Mode_Export {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_Mode).
			WithErrorDetails(fmt.Sprintf("The mode:%s is invalid", mode))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	span.SetAttributes(attr.Key(interfaces.QueryParam_Mode).String(mode))

	// 需要统计信息，默认不需要
	includeStatistics := c.DefaultQuery("include_statistics", interfaces.DEFAULT_INCLUDE_STATISTICS)
	includeStat, err := strconv.ParseBool(includeStatistics)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest,
			berrors.BknBackend_ConceptGroup_InvalidParameter_IncludeStatistics).
			WithErrorDetails(fmt.Sprintf("The include_statistics:%s is invalid", includeStatistics))

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//获取参数字符串，单个概念分组
	cgID := c.Param("cg_id")
	span.SetAttributes(attr.Key("cg_id").String(cgID))

	// 获取概念分组的详细信息
	cg, err := r.cgs.GetConceptGroupByID(ctx, knID, branch, cgID, mode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取概念统计信息
	if includeStat {
		statistics, err := r.cgs.GetStatByConceptGroup(ctx, cg)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}
		cg.Statistics = statistics
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler GetConceptGroup Success")
	rest.ReplyOK(c, http.StatusOK, cg)
}

// 创建概念分组(内部)
func (r *restHandler) AddObjectTypesToConceptGroupByIn(c *gin.Context) {
	logger.Debug("Handler AddObjectTypesToConceptGroupByIn Start")
	// 内部接口 user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.AddObjectTypesToConceptGroup(c, visitor)
}

// 创建概念分组（外部）
func (r *restHandler) AddObjectTypesToConceptGroupByEx(c *gin.Context) {
	logger.Debug("Handler AddObjectTypesToConceptGroupByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.AddObjectTypesToConceptGroup(c, visitor)
}

// 创建概念分组
func (r *restHandler) AddObjectTypesToConceptGroup(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler AddObjectTypesToConceptGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
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

	// 1. 接受 at_id 参数
	cgID := c.Param("cg_id")
	span.SetAttributes(attr.Key("cg_id").String(cgID))

	// 先按id获取原对象.
	_, exist, err = r.cgs.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 接受绑定参数 - 对象类
	var requestData struct {
		Entries []interfaces.ID `json:"entries"`
	}
	err = c.ShouldBindJSON(&requestData)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// Whether to validate dependencies, default true
	strictModeStr := c.Query(interfaces.QueryParam_StrictMode)
	if strictModeStr == "" {
		strictModeStr = "true"
	}
	strictMode, parseErr := strconv.ParseBool(strictModeStr)
	if parseErr != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_ConceptGroup_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 调用创建单个知识网络
	otCGIDs, err := r.cgs.AddObjectTypesToConceptGroup(ctx, nil, knID, branch, cgID, requestData.Entries, interfaces.ImportMode_Normal, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	result := []any{}
	for i, id := range otCGIDs {
		result = append(result, map[string]any{"id": id})
		// 成功创建记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			interfaces.GenerateConceptGroupRelationAuditObject(id, fmt.Sprintf("%s-%s-%s-%s", knID, branch, cgID, requestData.Entries[i].ID)), "")
	}

	logger.Debug("Handler AddObjectTypeToGroup Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// 创建概念分组(内部)
func (r *restHandler) DeleteObjectTypesFromGroupByIn(c *gin.Context) {
	logger.Debug("Handler DeleteObjectTypesFromGroupByIn Start")
	// 内部接口 user_id从header中取
	visitor := visitor.GenerateVisitor(c)
	r.DeleteObjectTypesFromGroup(c, visitor)
}

// 创建概念分组（外部）
func (r *restHandler) DeleteObjectTypesFromGroupByEx(c *gin.Context) {
	logger.Debug("Handler DeleteObjectTypesFromGroupByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.DeleteObjectTypesFromGroup(c, visitor)
}

// 创建概念分组
func (r *restHandler) DeleteObjectTypesFromGroup(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler DeleteObjectTypesFromGroup Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
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

	// 1. 接受 at_id 参数
	cgID := c.Param("cg_id")
	span.SetAttributes(attr.Key("cg_id").String(cgID))

	// 先按id获取原对象.
	_, exist, err = r.cgs.CheckConceptGroupExistByID(ctx, knID, branch, cgID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_ConceptGroup_ConceptGroupNotFound)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 获取参数字符串 <id1,id2,id3>
	otIDsStr := c.Param("ot_ids")
	span.SetAttributes(attr.Key("ot_ids").String(otIDsStr))

	// 解析字符串 转换为 []string
	otIDs := common.StringToStringSlice(otIDsStr)
	// id去重后再查
	otIDArr := common.DuplicateSlice(otIDs)
	// 检查组下的 otIDs 是否都存在绑定关系
	cgRelations, err := r.cgs.ListConceptGroupRelations(ctx, interfaces.ConceptGroupRelationsQueryParams{
		PaginationQueryParameters: interfaces.PaginationQueryParameters{
			Limit: -1,
		},
		KNID:        knID,
		Branch:      branch,
		CGIDs:       []string{cgID},
		ConceptType: interfaces.MODULE_TYPE_OBJECT_TYPE,
		OTIDs:       otIDArr,
	})
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}
	if len(cgRelations) != len(otIDArr) {
		errStr := fmt.Sprintf("Exists any object types not in the concept group [%s] knowledge network [%s] branch [%s], expect relations num is [%d], actual relations num is [%d]",
			cgID, knID, branch, len(otIDs), len(otIDArr))

		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound,
			berrors.BknBackend_ConceptGroup_ConceptGroupRelationNotExisted).WithErrorDetails(errStr)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 批量删除对象类
	err = r.cgs.DeleteObjectTypesFromGroup(ctx, nil, knID, branch, cgID, otIDArr)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 循环记录审计日志
	for _, cgr := range cgRelations {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateObjectTypeAuditObject(cgr.ID,
				fmt.Sprintf("%s-%s-%s-%s-%s", cgr.KNID, cgr.Branch, cgr.CGID, cgr.ConceptType, cgr.ConceptID)), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteObjectTypes Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}
