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

// 基于对象类的对象数据查询(内部)
func (r *restHandler) GetObjectsInObjectTypeByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectsInObjectTypeByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetObjectsInObjectType(c, visitor)
}

// 基于对象类的对象数据查询（外部）
func (r *restHandler) GetObjectsInObjectTypeByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectsInObjectTypeByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetObjectsInObjectType(c, visitor)
}

// 基于对象类的对象数据查询
func (r *restHandler) GetObjectsInObjectType(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetObjectsInObjectType Start")
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
	otellog.LogInfo(ctx, fmt.Sprintf("对象数据查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	//获取参数字符串
	otID := c.Param("ot_id")
	span.SetAttributes(attr.Key("ot_id").String(otID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// todo: 分页查询

	// 是否包含对象类信息
	includeTypeInfo := c.DefaultQuery("include_type_info", interfaces.DEFAULT_INCLUDE_TYPE_INFO)
	// 是否包含逻辑属性计算参数
	IncludeLogicParams := c.DefaultQuery("include_logic_params", interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS)
	// 是否忽略持久化数据,走虚拟化查询,默认是false,不忽略
	ignoringStoreCache := c.DefaultQuery("ignoring_store_cache", interfaces.DEFAULT_IGNORING_STORE_CACHE)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")

	// 校验查询参数
	objectsQueryParas, err := validateObjectsQueryParameters(ctx, includeTypeInfo, ignoringStoreCache, IncludeLogicParams, excludeSystemProperties)
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
	// instant query 参数， time（即start 和 end），isInstantQuery, interval = 1
	//接收绑定参数
	query := interfaces.ObjectQueryBaseOnObjectType{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	query.KNID = knID
	query.Branch = branch
	query.ObjectTypeID = otID
	query.CommonQueryParameters = objectsQueryParas

	err = validateObjectSearchRequest(ctx, &query)
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
	result, err := r.ots.GetObjectsByObjectTypeID(ctx, &query)
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

// 基于对象类的对象数据查询(内部)
func (r *restHandler) GetObjectsPropertiesByIn(c *gin.Context) {
	logger.Debug("Handler GetObjectsPropertiesByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetObjectsProperties(c, visitor)
}

// 基于对象类的对象数据查询（外部）
func (r *restHandler) GetObjectsPropertiesByEx(c *gin.Context) {
	logger.Debug("Handler GetObjectsPropertiesByEx Start")
	ctx, span := oteltrace.StartServerSpan(c)

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetObjectsProperties(c, visitor)
}

// 基于对象类的对象数据查询
func (r *restHandler) GetObjectsProperties(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetObjectsProperties Start")
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
	otellog.LogInfo(ctx, fmt.Sprintf("对象属性值查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	//获取参数字符串
	otID := c.Param("ot_id")
	span.SetAttributes(attr.Key("ot_id").String(otID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// 是否包含对象类信息
	includeTypeInfo := c.DefaultQuery("include_type_info", interfaces.DEFAULT_INCLUDE_TYPE_INFO)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")

	// 校验查询参数
	objectsQueryParas, err := validateObjectsQueryParameters(ctx, includeTypeInfo,
		interfaces.DEFAULT_IGNORING_STORE_CACHE,
		interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS,
		excludeSystemProperties)
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
	query := interfaces.ObjectPropertyValueQuery{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails), httpErr)

		rest.ReplyError(c, httpErr)

		return
	}

	query.KNID = knID
	query.Branch = branch
	query.ObjectTypeID = otID
	query.CommonQueryParameters = objectsQueryParas

	err = validateObjectPropertyValueQuery(ctx, &query)
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
	result, err := r.ots.GetObjectPropertyValue(ctx, &query)
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
