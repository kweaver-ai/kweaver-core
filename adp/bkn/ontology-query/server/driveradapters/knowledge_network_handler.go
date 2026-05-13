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
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

	"ontology-query/common/visitor"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// 基于起点、方向和路径长度获取对象子图（内部）
func (r *restHandler) GetObjectsSubgraphByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectsSubgraphByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	// 查询类型，默认是以起点扩展子图。
	queryType := c.DefaultQuery("query_type", "")
	switch queryType {
	case "":
		r.GetObjectsSubgraph(c, visitor)
	case interfaces.QUERY_TYPE_RELATION_TYPE_PATH:
		r.GetObjectsSubgraphByTypePath(c, visitor)
	}
}

// 基于起点、方向和路径长度获取对象子图（外部）
func (r *restHandler) GetObjectsSubgraphByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectsSubgraphByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	// 查询类型，默认是以起点扩展子图。
	queryType := c.DefaultQuery("query_type", "")
	switch queryType {
	case "":
		r.GetObjectsSubgraph(c, visitor)
	case interfaces.QUERY_TYPE_RELATION_TYPE_PATH:
		r.GetObjectsSubgraphByTypePath(c, visitor)
	}

}

// 基于对象类的对象数据查询
func (r *restHandler) GetObjectsSubgraph(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetObjectsSubgraph Start")
	startTime := time.Now()

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
	otellog.LogInfo(ctx, fmt.Sprintf("对象子图查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// 是否包含逻辑属性计算参数
	includeLogicParams := c.DefaultQuery("include_logic_params", interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS)
	// 是否忽略持久化数据,走虚拟化查询,默认是false,不忽略
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")
	// 校验查询参数
	queryParams, err := validateSugraphQueryParameters(ctx, includeLogicParams, ignoringStoreCache, excludeSystemProperties)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	//接收绑定参数
	query := interfaces.SubGraphQueryBaseOnSource{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	query.KNID = knID
	query.Branch = branch
	query.CommonQueryParameters = queryParams
	query.PathQuotaManager = &interfaces.PathQuotaManager{
		TotalLimit: interfaces.MAX_PATHS,
	}

	err = validateSubgraphSearchRequest(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 执行查询
	result, err := r.kns.SearchSubgraph(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 设置 trace 的成功信息的 attributes
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)

	result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
	rest.ReplyOK(c, http.StatusOK, result)
}

// 基于对象类的对象数据查询
func (r *restHandler) GetObjectsSubgraphByTypePath(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetObjectsSubgraphByTypePath Start")
	// startTime := time.Now()

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
	otellog.LogInfo(ctx, fmt.Sprintf("对象子图查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// 是否包含逻辑属性计算参数
	includeLogicParams := c.DefaultQuery("include_logic_params", interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS)
	// 是否忽略持久化数据,走虚拟化查询,默认是false,不忽略
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")
	// 校验查询参数
	queryParams, err := validateSugraphQueryParameters(ctx, includeLogicParams, ignoringStoreCache, excludeSystemProperties)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	//接收绑定参数
	paths := interfaces.QueryRelationTypePaths{}
	err = c.ShouldBindJSON(&paths)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}
	query := interfaces.SubGraphQueryBaseOnTypePath{
		Paths:                 paths,
		KNID:                  knID,
		Branch:                branch,
		CommonQueryParameters: queryParams,
	}

	err = validateSubgraphQueryByPathRequest(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 执行查询
	result, err := r.kns.SearchSubgraphByTypePath(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 设置 trace 的成功信息的 attributes
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)

	// result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
	rest.ReplyOK(c, http.StatusOK, result)
}

// 基于一组对象实例组织关系子图（外部）
func (r *restHandler) GetObjectsSubgraphByObjectsByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectsSubgraphByObjectsByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}

	r.GetObjectsSubgraphByObjects(c, visitor)
}

// 基于一组对象实例组织关系子图（内部）
func (r *restHandler) GetObjectsSubgraphByObjectsByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectsSubgraphByObjectsByIn Start")
	visitor := visitor.GenerateVisitor(c)
	r.GetObjectsSubgraphByObjects(c, visitor)
}

// 基于一组对象实例组织关系子图（通用处理函数）
func (r *restHandler) GetObjectsSubgraphByObjects(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetObjectsSubgraphByObjects Start")
	startTime := time.Now()

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
	otellog.LogInfo(ctx, fmt.Sprintf("基于一组对象实例组织关系子图查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// 是否包含对象类信息
	includeTypeInfo := c.DefaultQuery("include_type_info", interfaces.DEFAULT_INCLUDE_TYPE_INFO)
	// 是否包含逻辑属性计算参数
	includeLogicParams := c.DefaultQuery("include_logic_params", interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS)
	// 是否忽略持久化数据,走虚拟化查询,默认是false,不忽略
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")
	// 校验查询参数
	queryParams, err := validateObjectsQueryParameters(ctx, includeTypeInfo, ignoringStoreCache, includeLogicParams, excludeSystemProperties)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	//接收绑定参数
	query := interfaces.SubGraphQueryBaseOnObjects{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_KnowledgeNetwork_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	query.KNID = knID
	query.Branch = branch
	query.CommonQueryParameters = queryParams

	err = validateSubgraphQueryByObjectsRequest(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 执行查询
	result, err := r.kns.SearchSubgraphByObjects(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	// 设置 trace 的成功信息的 attributes
	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)

	result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
	rest.ReplyOK(c, http.StatusOK, result)
}
