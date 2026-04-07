// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package business_system

import (
	"context"
	"fmt"
	"net/http"
	"sync"

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
	bsAccessOnce sync.Once
	bsAccess     interfaces.BusinessSystemAccess
)

type businessSystemAccess struct {
	appSetting *common.AppSetting
	httpClient rest.HTTPClient
	bsUrl      string
}

// NewBusinessSystemAccess 创建业务系统访问实例
func NewBusinessSystemAccess(appSetting *common.AppSetting) interfaces.BusinessSystemAccess {
	bsAccessOnce.Do(func() {
		bsAccess = &businessSystemAccess{
			appSetting: appSetting,
			httpClient: common.NewHTTPClient(),
			bsUrl:      appSetting.BusinessSystemUrl,
		}
	})

	return bsAccess
}

func (bsa *businessSystemAccess) BindResource(ctx context.Context, bd_id string, rid string, rtype string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Bind resource to business system",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("business_system_id").String(bd_id),
		attr.Key("resource_id").String(rid),
		attr.Key("resource_type").String(rtype))

	httpUrl := fmt.Sprintf("%s/resource", bsa.bsUrl)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	body := map[string]string{
		"bd_id": bd_id,
		"id":    rid,
		"type":  rtype,
	}
	respCode, respData, err := bsa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, body)
	logger.Debugf("BindResource [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("BindResource http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http bind resource failed")
		return fmt.Errorf("BindResource http request failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("BindResource failed: %s", respData)
		err = fmt.Errorf("BindResource failed: %s", respData)
		return err
	}

	return nil
}

func (bsa *businessSystemAccess) UnbindResource(ctx context.Context, bd_id string, rid string, rtype string) error {

	ctx, span := ar_trace.Tracer.Start(ctx, "driven layer: Unbind resource from business system",
		trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("business_system_id").String(bd_id),
		attr.Key("resource_id").String(rid),
		attr.Key("resource_type").String(rtype))

	httpUrl := fmt.Sprintf("%s/resource?bd_id=%s&id=%s&type=%s", bsa.bsUrl, bd_id, rid, rtype)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodDelete,
		HttpContentType: rest.ContentTypeJson,
	})

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, respData, err := bsa.httpClient.DeleteNoUnmarshal(ctx, httpUrl, headers)
	logger.Debugf("UnbindResource [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, respData, err)

	if err != nil {
		errDetails := fmt.Sprintf("UnbindResource http request failed: %s", err.Error())
		logger.Error(errDetails)
		o11y.Error(ctx, errDetails)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http unbind resource failed")
		return fmt.Errorf("UnbindResource http request failed: %s", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("UnbindResource failed: %s", respData)
		err = fmt.Errorf("UnbindResource failed: %s", respData)
		return err
	}

	return nil
}
