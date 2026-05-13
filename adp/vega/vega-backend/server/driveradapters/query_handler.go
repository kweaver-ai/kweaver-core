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
	"github.com/kweaver-ai/kweaver-go-lib/hydra"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/errors"
	"vega-backend/interfaces"
	"vega-backend/logics/query"
)

// RawQueryByEx handles POST /api/vega-backend/v1/resources/query (External)
func (r *restHandler) RawQueryByEx(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 外网接口：校验token
	visitor, err := r.verifyOAuth(ctx, c)
	if err != nil {
		return
	}
	r.rawQuery(c, ctx, span, visitor)
}

// RawQueryByIn handles POST /api/vega-backend/in/v1/resources/query (Internal)
func (r *restHandler) RawQueryByIn(c *gin.Context) {
	ctx, span := oteltrace.StartServerSpan(c)
	defer span.End()

	// 内网接口：user_id从header中取
	visitor := GenerateVisitor(c)
	r.rawQuery(c, ctx, span, visitor)
}

// sqlQuery is the shared implementation for SQL query
func (r *restHandler) rawQuery(c *gin.Context, ctx context.Context, span trace.Span, visitor hydra.Visitor) {
	accountInfo := interfaces.AccountInfo{
		ID:   visitor.ID,
		Type: string(visitor.Type),
	}
	ctx = context.WithValue(ctx, interfaces.ACCOUNT_INFO_KEY, accountInfo)

	oteltrace.AddHttpAttrs4API(span, oteltrace.GetAttrsByGinCtx(c))

	var req interfaces.RawQueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errors.VegaBackend_InvalidParameter_RequestBody).
			WithErrorDetails(err.Error())
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验resource_type参数，必填，必须是当前统一查询接口支持的连接器类型
	if req.ResourceType == "" {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errors.VegaBackend_InvalidParameter_ResourceType).
			WithErrorDetails(fmt.Sprintf("resource_type is required and must be one of: %v", interfaces.GetSupportedConnectorTypesForQuery()))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	if !interfaces.IsConnectorTypeSupportedForQuery(req.ResourceType) {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errors.VegaBackend_InvalidParameter_ResourceType).
			WithErrorDetails(fmt.Sprintf("resource_type must be one of: %v, got: %s", interfaces.GetSupportedConnectorTypesForQuery(), req.ResourceType))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验stream_size参数，默认值为10000，最大值为10000，最小值为100
	if req.StreamSize == 0 {
		req.StreamSize = 10000 // 设置默认值
	} else if req.StreamSize < 100 || req.StreamSize > 10000 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errors.VegaBackend_InvalidParameter_StreamSize).
			WithErrorDetails(fmt.Sprintf("stream_size must be between 100 and 10000, got: %d", req.StreamSize))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	// 校验query_timeout参数，默认值为60，最大值为3600，最小值为1
	if req.QueryTimeout == 0 {
		req.QueryTimeout = 60 // 设置默认值
	} else if req.QueryTimeout < 1 || req.QueryTimeout > 3600 {
		httpErr := rest.NewHTTPError(ctx, http.StatusBadRequest, errors.VegaBackend_Query_InvalidParameter_QueryTimeout).
			WithErrorDetails(fmt.Sprintf("query_timeout must be between 1 and 3600, got: %d", req.QueryTimeout))
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	qs := query.NewRawQueryService(r.appSetting)
	resp, err := qs.Execute(ctx, &req)
	if err != nil {
		var httpErr *rest.HTTPError
		var ok bool
		if httpErr, ok = err.(*rest.HTTPError); !ok {
			// 如果不是HTTPError，则转换为内部服务器错误
			httpErr = rest.NewHTTPError(ctx, http.StatusInternalServerError, errors.VegaBackend_Query_ExecuteFailed).
				WithErrorDetails(err.Error())
		}
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		rest.ReplyError(c, httpErr)
		return
	}

	oteltrace.AddHttpAttrs4Ok(span, http.StatusOK)
	rest.ReplyOK(c, http.StatusOK, resp)
}
