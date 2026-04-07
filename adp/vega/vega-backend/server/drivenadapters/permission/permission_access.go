package permission

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"vega-backend/common"
	"vega-backend/interfaces"
)

var (
	pAccessOnce sync.Once
	pAccess     interfaces.PermissionAccess
)

type permissionAccess struct {
	appSetting    *common.AppSetting
	permissionUrl string
	httpClient    rest.HTTPClient
}

type PermissionError struct {
	Code        string `json:"code"`
	Description string `json:"description"`
	Cause       any    `json:"cause"`
}

func NewPermissionAccess(appSetting *common.AppSetting) interfaces.PermissionAccess {
	pAccessOnce.Do(func() {
		pAccess = &permissionAccess{
			appSetting:    appSetting,
			permissionUrl: appSetting.PermissionUrl,
			httpClient:    common.NewHTTPClient(),
		}
	})

	return pAccess
}

func (pa *permissionAccess) CheckPermission(ctx context.Context, check interfaces.PermissionCheck) (bool, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "请求策略的决策接口", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(check.Accessor.ID),
		attr.Key("resource_id").String(check.Resource.ID),
		attr.Key("Operation").StringSlice(check.Operations),
	)

	httpUrl := fmt.Sprintf("%s/operation-check", pa.permissionUrl)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:            httpUrl,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodGet,
	})

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	check.Method = http.MethodGet
	respCode, result, err := pa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, check)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Post operation-check request failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("Post operation-check request failed: %v", err))
		return false, fmt.Errorf("post operation-check request failed: %v", err)
	}
	if respCode != http.StatusOK {
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			logger.Errorf("unmalshal PermissionError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal PermissionError failed: %v", err))
			return false, err
		}
		httpErr := &rest.HTTPError{
			HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  permissionError.Description,
				ErrorDetails: permissionError.Cause,
			}}
		logger.Errorf("operation-check error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Post operation-check failed: %v", httpErr))
		return false, httpErr
	}

	if result == nil {
		o11y.AddHttpAttrs4Ok(span, respCode)
		o11y.Warn(ctx, "Http response body is null")
		return false, nil
	}

	var checkResult interfaces.PermissionCheckResult
	if err := sonic.Unmarshal(result, &checkResult); err != nil {
		logger.Errorf("unmalshal operation-check result failed: %v\n", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal operation-check result failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal operation-check result failed: %v", err))
		return false, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)

	return checkResult.Result, nil
}

func (pa *permissionAccess) CreateResources(ctx context.Context, policies []interfaces.PermissionPolicy) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "请求创建决策接口", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(policies[0].Accessor.ID),
		attr.Key("resource_id").String(policies[0].Resource.ID),
		attr.Key("Operation").String(fmt.Sprintf("%v", policies[0].Operations)),
	)

	httpUrl := fmt.Sprintf("%s/policy", pa.permissionUrl)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:            httpUrl,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodGet,
	})

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	respCode, result, err := pa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, policies)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Post create policy request failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("Post create policy request failed: %v", err))
		return fmt.Errorf("post create policy request failed: %v", err)
	}
	if respCode != http.StatusNoContent {
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			logger.Errorf("unmalshal PermissionError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal PermissionError failed: %v", err))
			return err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  permissionError.Description,
				ErrorDetails: permissionError.Cause,
			}}
		logger.Errorf("create policy error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Post create policy failed: %v", httpErr))
		return httpErr
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return nil
}

func (pa *permissionAccess) DeleteResources(ctx context.Context, res []interfaces.PermissionResource) error {
	ctx, span := ar_trace.Tracer.Start(ctx, "请求删除决策接口", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	createUrl := fmt.Sprintf("%s/policy-delete", pa.permissionUrl)

	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:            createUrl,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodDelete,
	})

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	st := map[string]any{
		"method":    http.MethodDelete,
		"resources": res,
	}

	respCode, result, err := pa.httpClient.PostNoUnmarshal(ctx, createUrl, headers, st)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", createUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Post delete policy request failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("Post delete policy request failed: %v", err))
		return fmt.Errorf("post delete policy request failed: %v", err)
	}
	if respCode != http.StatusNoContent {
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			logger.Errorf("unmalshal PermissionError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal PermissionError failed: %v", err))
			return err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  permissionError.Description,
				ErrorDetails: permissionError.Cause,
			}}
		logger.Errorf("delete policy error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Post delete policy failed: %v", httpErr))
		return httpErr
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return nil
}

func (pa *permissionAccess) FilterResources(ctx context.Context, filter interfaces.PermissionResourcesFilter) ([]interfaces.PermissionResourceOps, error) {
	ctx, span := ar_trace.Tracer.Start(ctx, "请求资源过滤接口", trace.WithSpanKind(trace.SpanKindClient))
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(filter.Accessor.ID),
		attr.Key("Operation").StringSlice(filter.Operations),
	)

	httpUrl := fmt.Sprintf("%s/resource-filter", pa.permissionUrl)
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:            httpUrl,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodGet,
	})

	var ops []interfaces.PermissionResourceOps

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	filter.Method = http.MethodGet
	respCode, result, err := pa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, filter)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		logger.Errorf("Post operation-check request failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		o11y.Error(ctx, fmt.Sprintf("Post operation-check request failed: %v", err))
		return ops, fmt.Errorf("post operation-check request failed: %v", err)
	}
	if respCode != http.StatusOK {
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			logger.Errorf("unmalshal PermissionError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal PermissionError failed: %v", err))
			return ops, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  permissionError.Description,
				ErrorDetails: permissionError.Cause,
			}}
		logger.Errorf("operation-filter error: %v", httpErr.Error())
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("Post operation-filter failed: %v", httpErr))
		return ops, httpErr
	}

	if result == nil {
		o11y.AddHttpAttrs4Ok(span, respCode)
		o11y.Warn(ctx, "Http response body is null")
		return ops, nil
	}

	if err := sonic.Unmarshal(result, &ops); err != nil {
		logger.Errorf("unmalshal operation-check result failed: %v\n", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal operation-filter result failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal operation-filter result failed: %v", err))
		return ops, err
	}

	o11y.AddHttpAttrs4Ok(span, respCode)

	return ops, nil
}
