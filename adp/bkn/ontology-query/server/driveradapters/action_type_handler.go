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
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common/visitor"
	oerrors "ontology-query/errors"
	"ontology-query/interfaces"
)

// 基于对象类的对象数据查询(内部)
func (r *restHandler) GetActionsInActionTypeByIn(c *gin.Context) {
	logger.Debug("Handler GetActionsInActionTypeByIn Start")
	// 内部接口 user_id从header中取，跳过用户有效认证，后面在权限校验时就会校验这个用户是否有权限，无效用户无权限
	// 自行构建一个visitor
	visitor := visitor.GenerateVisitor(c)
	r.GetActionsInActionType(c, visitor)
}

// 基于对象类的对象数据查询（外部）
func (r *restHandler) GetActionsInActionTypeByEx(c *gin.Context) {
	logger.Debug("Handler GetActionsInActionTypeByEx Start")
	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取对象类[%s]的对象数据API",
		trace.WithSpanKind(trace.SpanKindServer))

	defer span.End()

	// 校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.GetActionsInActionType(c, visitor)
}

// 基于对象类的对象数据查询
func (r *restHandler) GetActionsInActionType(c *gin.Context, visitor hydra.Visitor) {
	logger.Debug("Handler GetActionsInActionType Start")
	startTime := time.Now()

	ctx, span := ar_trace.Tracer.Start(rest.GetLanguageCtx(c), "获取行动类[%s]的行动数据API", trace.WithSpanKind(trace.SpanKindServer))
	defer span.End()

	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	// accountID 存入 context 中
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	// 设置 trace 的相关 api 的属性
	o11y.AddHttpAttrs4API(span, o11y.GetAttrsByGinCtx(c))

	// 记录接口调用参数： c.Request.RequestURI, body
	o11y.Info(ctx, fmt.Sprintf("行动数据查询请求参数: [%s,%v]", c.Request.RequestURI, c.Request.Body))

	// 1. 接受 kn_id 参数
	knID := c.Param("kn_id")
	span.SetAttributes(attr.Key("kn_id").String(knID))

	//获取参数字符串
	otID := c.Param("at_id")
	span.SetAttributes(attr.Key("at_id").String(otID))

	// 接受 branch 参数
	branch := c.DefaultQuery("branch", interfaces.MAIN_BRANCH)
	span.SetAttributes(attr.Key("branch").String(branch))

	// 是否包含行动类信息
	includeTypeInfo := c.DefaultQuery("include_type_info", interfaces.DEFAULT_INCLUDE_TYPE_INFO)
	// 排除系统字段列表
	excludeSystemProperties := c.QueryArray("exclude_system_properties")
	// 校验查询参数
	objectsQueryParas, err := validateObjectsQueryParameters(ctx, includeTypeInfo, interfaces.DEFAULT_IGNORING_STORE_CACHE,
		interfaces.DEFAULT_INCLUDE_LOGIC_PARAMS, excludeSystemProperties)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		return
	}

	err = ValidateHeaderMethodOverride(ctx, c.GetHeader(interfaces.HTTP_HEADER_METHOD_OVERRIDE))
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		return
	}
	// instant query 参数， time（即start 和 end），isInstantQuery, interval = 1
	//接收绑定参数
	query := interfaces.ActionQuery{}
	err = c.ShouldBindJSON(&query)
	if err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, oerrors.OntologyQuery_ObjectType_InvalidParameter).
			WithErrorDetails(fmt.Sprintf("Binding Paramter Failed:%s", err.Error()))

		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)
		return
	}

	query.KNID = knID
	query.Branch = branch
	query.ActionTypeID = otID
	query.CommonQueryParameters = objectsQueryParas

	err = validateActionQuery(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		return
	}

	// 执行查询
	result, err := r.ats.GetActionsByActionTypeID(ctx, &query)
	if err != nil {
		httpErr := err.(*rest.HTTPError)
		// 设置 trace 的错误信息的 attributes
		o11y.AddHttpAttrs4HttpError(span, httpErr)
		o11y.Error(ctx, fmt.Sprintf("%s. %v", httpErr.BaseError.Description,
			httpErr.BaseError.ErrorDetails))

		rest.ReplyError(c, httpErr)

		return
	}

	// 设置 trace 的成功信息的 attributes
	o11y.AddHttpAttrs4Ok(span, http.StatusOK)

	result.OverallMs = time.Now().UnixMilli() - startTime.UnixMilli()
	rest.ReplyOK(c, http.StatusOK, result)

}
