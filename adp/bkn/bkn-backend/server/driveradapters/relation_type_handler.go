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

func (r *restHandler) HandleRelationTypeGetOverrideByIn(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateRelationTypesByIn(c)
	case http.MethodGet:
		r.SearchRelationTypesByIn(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

func (r *restHandler) HandleRelationTypeGetOverrideByEx(c *gin.Context) {
	switch c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE) {
	case "", http.MethodPost:
		r.CreateRelationTypesByEx(c)
	case http.MethodGet:
		r.SearchRelationTypesByEx(c)
	default:
		httpErr := rest.NewHTTPError(rest.GetLanguageCtx(c), http.StatusBadRequest,
			berrors.BknBackend_InvalidParameter_OverrideMethod)
		rest.ReplyError(c, httpErr)
	}
}

// 创建关系类(内部)
func (r *restHandler) CreateRelationTypesByIn(c *gin.Context) {
	logger.Debug("Handler CreateRelationTypesByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.CreateRelationTypes(c, visitor)
}

// 创建关系类（外部）
func (r *restHandler) CreateRelationTypesByEx(c *gin.Context) {
	logger.Debug("Handler CreateRelationTypesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.CreateRelationTypes(c, visitor)
}

// 创建关系类
func (r *restHandler) CreateRelationTypes(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler CreateRelationTypes Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// account_type 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	// 查询参数
	mode := c.DefaultQuery(interfaces.QueryParam_ImportMode, interfaces.ImportMode_Normal)
	httpErr := validateImportMode(ctx, mode)
	if httpErr != nil {
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

	// Whether to validate dependencies, default true. Parse priority: strict_mode > validate_dependency (legacy) > true
	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

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
	var requestData struct {
		Entries []*interfaces.RelationType `json:"entries"`
	}
	err = c.ShouldBindJSON(&requestData)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	relationTypes := requestData.Entries

	// 如果传入的模型对象为[], 应报错
	if len(relationTypes) == 0 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_InvalidParameter_RequestBody).
			WithErrorDetails("No relation type was passed in")

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("创建关系类请求参数: [%s,%v]", c.Request.RequestURI, relationTypes))

	// request来的relationTypes的branch都用url里的branch
	for i := range relationTypes {
		relationTypes[i].KNID = knID
		relationTypes[i].Branch = branch
	}

	// 校验 请求体中目标模型名称合法性
	err = ValidateRelationTypes(ctx, knID, relationTypes, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//调用创建
	// 直接创建关系类接口默认进行依赖校验
	rtIDs, err := r.rts.CreateRelationTypes(ctx, nil, relationTypes, mode, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 成功，发送多条
	for _, relationType := range relationTypes {
		//每次成功创建 记录审计日志
		audit.NewInfoLog(audit.OPERATION, audit.CREATE, audit.TransforOperator(visitor),
			interfaces.GenerateRelationTypeAuditObject(relationType.RTID, relationType.RTName), "")
	}

	result := []any{}
	for _, rtID := range rtIDs {
		result = append(result, map[string]any{"id": rtID})
	}

	logger.Debug("Handler CreateRelationTypes Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusCreated, result)
}

// ValidateRelationTypesByIn 仅校验关系类依赖存在性，不写库（内部）
func (r *restHandler) ValidateRelationTypesByIn(c *gin.Context) {
	logger.Debug("Handler ValidateRelationTypesByIn Start")
	v := visitor.GenerateVisitor(c)
	r.ValidateRelationTypesForKN(c, v)
}

// ValidateRelationTypesByEx 仅校验关系类依赖存在性，不写库（外部）
func (r *restHandler) ValidateRelationTypesByEx(c *gin.Context) {
	logger.Debug("Handler ValidateRelationTypesByEx Start")
	ctx, _ := oteltrace.StartServerSpan(c)

	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ValidateRelationTypesForKN(c, visitor)
}

// ValidateRelationTypesForKN 仅校验关系类依赖存在性，不写库
func (r *restHandler) ValidateRelationTypesForKN(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler ValidateRelationTypesForKN Start")
	ctx, _ := oteltrace.StartServerSpan(c)

	accountInfo := interfaces.AccountInfo{ID: visitor.ID, Type: string(visitor.Type)}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	strictModeStr := c.DefaultQuery(interfaces.QueryParam_StrictMode, "true")
	strictMode, err := strconv.ParseBool(strictModeStr)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
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
		Entries []*interfaces.RelationType `json:"entries"`
	}
	if err = c.ShouldBindJSON(&requestData); err != nil {
		rest.ReplyError(c, rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("Binding Parameter Failed: "+err.Error()))
		return
	}
	relationTypes := requestData.Entries
	if len(relationTypes) == 0 {
		rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": true})
		return
	}

	// request来的relationTypes的branch都用url里的branch
	for i := range relationTypes {
		relationTypes[i].KNID = knID
		relationTypes[i].Branch = branch
	}

	if err = ValidateRelationTypes(ctx, knID, relationTypes, strictMode); err != nil {
		rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": false, "detail": err.Error()})
		return
	}
	if err = r.rts.ValidateRelationTypes(ctx, knID, branch, relationTypes, strictMode, nil, mode); err != nil {
		rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": false, "detail": err.Error()})
		return
	}
	rest.ReplyOK(c, http.StatusOK, map[string]any{"valid": true})
}

// 更新关系类(内部)
func (r *restHandler) UpdateRelationTypeByIn(c *gin.Context) {
	logger.Debug("Handler UpdateRelationTypeByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.UpdateRelationType(c, visitor)
}

// 更新关系类（外部）
func (r *restHandler) UpdateRelationTypeByEx(c *gin.Context) {
	logger.Debug("Handler UpdateRelationTypeByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.UpdateRelationType(c, visitor)
}

// 更新关系类
func (r *restHandler) UpdateRelationType(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler UpdateRelationType Start")
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
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Invalid strict_mode parameter: %s", strictModeStr))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验业务知识网络存在性
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

	// 1. 接受 rt_id 参数
	rtID := c.Param("rt_id")
	span.SetAttributes(attr.Key("rt_id").String(rtID))

	//接收绑定参数
	relationType := interfaces.RelationType{}
	err = c.ShouldBindJSON(&relationType)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails("Binding Paramter Failed:" + err.Error())

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	relationType.RTID = rtID
	relationType.KNID = knID
	relationType.Branch = branch

	// 记录接口调用参数： c.Request.RequestURI, body
	otellog.LogInfo(ctx, fmt.Sprintf("修改关系类请求参数: [%s, %v]", c.Request.RequestURI, relationType))

	// 先按id获取原对象
	_, exist, err = r.rts.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !exist {
		httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_RelationType_RelationTypeNotFound)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验 关系类基本参数的合法性, 非空、长度、是枚举值
	err = ValidateRelationType(ctx, &relationType, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("Validate relation type[%s] failed: %s. %v", relationType.RTName,
			httpErr.BaseError.Description, httpErr.BaseError.ErrorDetails), nil)

		// 设置 trace 的错误信息的 attributes
		span.SetAttributes(attr.Key("rt_name").String(relationType.RTName))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//根据id修改信息
	err = r.rts.UpdateRelationType(ctx, nil, &relationType, strictMode)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	audit.NewInfoLog(audit.OPERATION, audit.UPDATE, audit.TransforOperator(visitor),
		interfaces.GenerateRelationTypeAuditObject(rtID, relationType.RTName), "")

	logger.Debug("Handler UpdateRelationType Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 批量删除关系类
func (r *restHandler) DeleteRelationTypes(c *gin.Context) {
	logger.Debug("Handler DeleteRelationTypes Start")
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
	otellog.LogInfo(ctx, fmt.Sprintf("删除关系类请求参数: [%s]", c.Request.RequestURI))

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
	otIDsStr := c.Param("rt_ids")
	span.SetAttributes(attr.Key("rt_ids").String(otIDsStr))

	//解析字符串 转换为 []string
	rtIDs := common.StringToStringSlice(otIDsStr)

	//检查 rtIDs 是否都存在
	var relationTypes []*interfaces.RelationTypeWithKeyField
	for _, rtID := range rtIDs {
		rtName, exist, err := r.rts.CheckRelationTypeExistByID(ctx, knID, branch, rtID)
		if err != nil {
			httpErr := err.(*rest.HTTPError)

			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)

			rest.ReplyError(c, httpErr)
			return
		}
		if !exist {
			httpErr := rest.NewHTTPError(ctx, http.StatusNotFound, berrors.BknBackend_RelationType_RelationTypeNotFound)

			// 设置 trace 的错误信息的 attributes
			oteltrace.AddHttpAttrs4HttpError(span, httpErr)
			rest.ReplyError(c, httpErr)
			return
		}

		relationTypes = append(relationTypes, &interfaces.RelationTypeWithKeyField{RTID: rtID, RTName: rtName})
	}

	// 批量删除关系类
	err = r.rts.DeleteRelationTypesByIDs(ctx, nil, knID, branch, rtIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	//循环记录审计日志
	for _, relationType := range relationTypes {
		audit.NewWarnLog(audit.OPERATION, audit.DELETE, audit.TransforOperator(visitor),
			interfaces.GenerateRelationTypeAuditObject(relationType.RTID, relationType.RTName), audit.SUCCESS, "")
	}

	logger.Debug("Handler DeleteRelationTypes Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusNoContent, nil)
}

// 分页获取关系类列表(内部)
func (r *restHandler) ListRelationTypesByIn(c *gin.Context) {
	logger.Debug("Handler ListRelationTypesByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.ListRelationTypes(c, visitor)
}

// 分页获取关系类列表（外部）
func (r *restHandler) ListRelationTypesByEx(c *gin.Context) {
	logger.Debug("Handler ListRelationTypesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.ListRelationTypes(c, visitor)
}

// 分页获取关系类列表
func (r *restHandler) ListRelationTypes(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("ListRelationTypes Start")
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
	otellog.LogInfo(ctx, fmt.Sprintf("分页获取关系类列表请求参数: [%s]", c.Request.RequestURI))

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
	tag := c.Query("tag")
	sourceObjectTypeIDs := c.QueryArray("source_object_type_id")
	targetObjectTypeIDs := c.QueryArray("target_object_type_id")
	boundObjectTypeIDs := c.QueryArray("bound_object_type_id")

	offset := c.DefaultQuery("offset", interfaces.DEFAULT_OFFEST)
	limit := c.DefaultQuery("limit", interfaces.DEFAULT_LIMIT)
	sort := c.DefaultQuery("sort", "update_time")
	direction := c.DefaultQuery("direction", interfaces.DESC_DIRECTION)

	//去掉标签前后的所有空格进行搜索
	tag = strings.Trim(tag, " ")

	// 校验分页查询参数
	pageParam, err := validatePaginationQueryParameters(ctx,
		offset, limit, sort, direction, interfaces.RELATION_TYPE_SORT)
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
	parameter := interfaces.RelationTypesQueryParams{
		NamePattern: namePattern,
		Tag:         tag,
		Branch:      branch,
		KNID:        knID,
	}

	// 不为空时，赋值
	if len(sourceObjectTypeIDs) > 0 {
		parameter.SourceObjectTypeIDs = sourceObjectTypeIDs
	}
	if len(targetObjectTypeIDs) > 0 {
		parameter.TargetObjectTypeIDs = targetObjectTypeIDs
	}
	if len(boundObjectTypeIDs) > 0 {
		parameter.BoundObjectTypeIDs = boundObjectTypeIDs
	}

	parameter.Sort = pageParam.Sort
	parameter.Direction = pageParam.Direction
	parameter.Limit = pageParam.Limit
	parameter.Offset = pageParam.Offset

	// var result map[string]any
	// if simpleInfo {
	// 获取关系类简单信息
	otList, total, err := r.rts.ListRelationTypes(ctx, parameter)
	result := map[string]any{"entries": otList, "total_count": total}
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

	logger.Debug("Handler ListRelationTypes Success")
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, result)
}

// 按 id 获取关系类对象信息(内部)
func (r *restHandler) GetRelationTypesByIn(c *gin.Context) {
	logger.Debug("Handler GetRelationTypesByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetRelationTypes(c, visitor)
}

// 按 id 获取关系类对象信息（外部）
func (r *restHandler) GetRelationTypesByEx(c *gin.Context) {
	logger.Debug("Handler ListRelationTypesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetRelationTypes(c, visitor)
}

// 按 id 获取关系类对象信息
func (r *restHandler) GetRelationTypes(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetRelationTypes Start")
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

	//获取参数字符串
	rtIDsStr := c.Param("rt_ids")
	span.SetAttributes(attr.Key("rt_ids").String(rtIDsStr))

	//解析字符串 转换为 []string
	rtIDs := common.StringToStringSlice(rtIDsStr)

	// 获取关系类的详细信息，根据 include_view 参数来判断是否包含数据视图的过滤条件
	result, err := r.rts.GetRelationTypesByIDs(ctx, knID, branch, rtIDs)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	httpResult := map[string]any{"entries": result}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler GetRelationTypes Success")
	rest.ReplyOK(c, http.StatusOK, httpResult)
}

// 检索关系类（外部）
func (r *restHandler) SearchRelationTypesByIn(c *gin.Context) {
	logger.Debug("Handler SearchRelationTypesByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.SearchRelationTypes(c, visitor)
}

// 检索关系类（外部）
func (r *restHandler) SearchRelationTypesByEx(c *gin.Context) {
	logger.Debug("Handler SearchRelationTypesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.SearchRelationTypes(c, visitor)
}

// 检索对象类
func (r *restHandler) SearchRelationTypes(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("SearchRelationTypes Start")
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
	otellog.LogInfo(ctx, fmt.Sprintf("检索对象类请求参数: [%s]", c.Request.RequestURI))

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

	//接收绑定参数
	query := interfaces.ConceptsQuery{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, berrors.BknBackend_RelationType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Concept Query Paramter Failed:%s", err.Error()))

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	query.KNID = knID
	query.Branch = branch
	query.ModuleType = interfaces.MODULE_TYPE_RELATION_TYPE

	// 校验：概念类型非空时，需要是指定的枚举类型；过滤条件的字段只能是type_id, type_name, property_name, property_dispaly_name, comment, *
	if query.Limit == 0 {
		query.Limit = interfaces.DEFAULT_CONCEPT_SEARCH_LIMIT
	}

	if query.Sort == nil {
		query.Sort = []*interfaces.SortParams{
			{
				Field:     interfaces.OPENSEARCH_SCORE_FIELD,
				Direction: interfaces.DESC_DIRECTION,
			},
			{
				Field:     interfaces.CONCEPT_ID_FIELD,
				Direction: interfaces.ASC_DIRECTION,
			},
		}
	}

	err = validateConceptsQuery(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), nil)

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 搜索概念
	result, err := r.rts.SearchRelationTypes(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)

		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	logger.Debug("Handler SearchRelationTypes Success")
	rest.ReplyOK(c, http.StatusOK, result)
}
