// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package data_view

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/bytedance/sonic/decoder"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	dvAccessOnce sync.Once
	dvAccess     interfaces.DataViewAccess
)

type dataViewAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
}

func NewDataViewAccess(appSetting *common.AppSetting) interfaces.DataViewAccess {
	dvAccessOnce.Do(func() {
		dvAccess = &dataViewAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
		}
	})
	return dvAccess
}

// 根据 id 获取视图
func (dva *dataViewAccess) GetDataViewByID(ctx context.Context, id string) (*interfaces.DataView, error) {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Get views by IDs from data-model service", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("view_id").String(id))

	httpUrl := fmt.Sprintf("%s/data-views/%s", dva.appSetting.DataViewUrl, id)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		"X-Language":                        rest.GetLanguageByCtx(ctx),
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := dva.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("get [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("GetDataViewByID http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get view failed")

		return nil, fmt.Errorf("get request method failed: %s", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("data view [%s] not exists", id)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("data view [%s] not found", id))

		return nil, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get data view failed: %s", respData)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(respData, &baseError); err != nil {
			logger.Errorf("Unmalshal baesError failed: %s", err)
			o11y.Error(ctx, err.Error())
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal baseError failed")
			return nil, err
		}

		o11y.Error(ctx, fmt.Sprintf("%s. %v", baseError.Description, baseError.ErrorDetails))
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, fmt.Errorf("GetDataViewByIDs failed: %s", baseError.ErrorDetails)
	}

	var views []*interfaces.DataView
	if err = sonic.Unmarshal(respData, &views); err != nil {
		logger.Errorf("Unmarshal data view failed: %s", err)
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal data view info failed")
		return nil, err
	}

	if len(views) == 0 {
		return nil, nil
	}

	// 将字段转成 map 结构
	fieldsMap := make(map[string]*interfaces.ViewField)
	for _, f := range views[0].Fields {
		field := f
		fieldsMap[f.Name] = field
	}
	views[0].FieldsMap = fieldsMap

	o11y.AddHttpAttrs4Ok(span, respCode)
	return views[0], nil
}

// 分批获取视图数据
func (dva *dataViewAccess) GetDataStart(ctx context.Context, id string,
	incKey string, incValue any, limit int) (*interfaces.ViewQueryResult, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: GetDataStart",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("view_id").String(id))

	httpUrl := fmt.Sprintf("%s/data-views/%s?include_view=true&timeout=5m", dva.appSetting.UniQueryUrl, id)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	params := map[string]any{
		"need_total":       true,
		"use_search_after": true,
		"limit":            limit,
	}

	if incKey != "" {
		params["sort"] = []map[string]any{
			{
				"field":     incKey,
				"direction": "asc",
			},
		}
		if incValue != nil {
			params["filters"] = map[string]any{
				"field":     incKey,
				"operation": ">",
				"value":     incValue,
			}
		}
	}

	respCode, respData, err := dva.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, params)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("GetDataStart http request failed: response code is [%d], result is [%s], error is [%v]", respCode, respData, err)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http post failed")
		return nil, err
	}

	if respCode != http.StatusOK {
		err = fmt.Errorf("DataPlatform get_data_start error: response code is [%d], result is [%s]", respCode, respData)
		logger.Error(err.Error())
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, err
	}

	var result interfaces.ViewQueryResult
	d := decoder.NewDecoder(string(respData))
	d.UseInt64()
	if err = d.Decode(&result); err != nil {
		errDetails := fmt.Sprintf("GetDataStart unmarshal result failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal result failed")
		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return &result, nil
}

// 分批获取视图数据
func (dva *dataViewAccess) GetDataNext(ctx context.Context, id string,
	searchAfter []any, limit int) (*interfaces.ViewQueryResult, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: GetDataNext",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(attr.Key("view_id").String(id))

	httpUrl := fmt.Sprintf("%s/data-views/%s?timeout=5m", dva.appSetting.UniQueryUrl, id)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	params := map[string]any{
		"use_search_after": true,
		"search_after":     searchAfter,
		"limit":            limit,
	}

	respCode, respData, err := dva.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, params)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("GetDataNext http request failed: response code is [%d], result is [%s], error is [%v]", respCode, respData, err)
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http post failed")
		return nil, err
	}

	if respCode != http.StatusOK {
		err = fmt.Errorf("DataPlatform get_data_next error: response code is [%d], result is [%s]", respCode, respData)
		logger.Error(err.Error())
		o11y.Error(ctx, err.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status code is not 200")
		return nil, err
	}

	var result interfaces.ViewQueryResult
	d := decoder.NewDecoder(string(respData))
	d.UseInt64()
	if err = d.Decode(&result); err != nil {
		errDetails := fmt.Sprintf("GetDataNext unmarshal result failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal result failed")
		return nil, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return &result, nil
}
