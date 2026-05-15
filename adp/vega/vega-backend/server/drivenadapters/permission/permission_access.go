// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package permission

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	attr "go.opentelemetry.io/otel/attribute"

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
	Code        string `json:"code"`        // 错误码
	Message     string `json:"message"`     // 错误描述
	Description string `json:"description"` // 错误描述
	Cause       any    `json:"cause"`       // 原因
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

// 策略决策
func (pa *permissionAccess) CheckPermission(ctx context.Context, check interfaces.PermissionCheck) (bool, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "请求策略的决策接口")
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(check.Accessor.ID),
		attr.Key("resource_id").String(check.Resource.ID),
		attr.Key("Operation").StringSlice(check.Operations),
	)

	httpUrl := fmt.Sprintf("%s/operation-check", pa.permissionUrl)

	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
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
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		otellog.LogError(ctx, "Post operation-check request failed", err)

		return false, fmt.Errorf("post operation-check request failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			// 添加异常时的 trace 属性
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			// 记录异常日志
			otellog.LogError(ctx, "Unmalshal PermissionError failed", err)

			return false, err
		}

		description := permissionError.Message
		if description == "" {
			description = permissionError.Description
		}
		httpErr := &rest.HTTPError{
			HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  description,
				ErrorDetails: permissionError.Cause,
			}}

		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		otellog.LogError(ctx, "Post operation-check failed", httpErr)

		return false, httpErr
	}

	if result == nil {
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		otellog.LogWarn(ctx, "Http response body is null")

		return false, nil
	}

	// 处理返回结果 result
	var checkResult interfaces.PermissionCheckResult
	if err := sonic.Unmarshal(result, &checkResult); err != nil {
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal operation-check result failed")
		// 记录异常日志
		otellog.LogError(ctx, "Unmalshal operation-check result failed", err)

		return false, err
	}

	// 添加成功时的 trace 属性
	oteltrace.AddHttpAttrs4Ok(span, respCode)

	return checkResult.Result, nil
}

// 创建策略
func (pa *permissionAccess) CreateResources(ctx context.Context, policies []interfaces.PermissionPolicy) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "请求创建决策接口")
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(policies[0].Accessor.ID),
		attr.Key("resource_id").String(policies[0].Resource.ID),
		attr.Key("Operation").String(fmt.Sprintf("%v", policies[0].Operations)),
	)

	httpUrl := fmt.Sprintf("%s/policy", pa.permissionUrl)

	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
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
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		otellog.LogError(ctx, "Post create policy request failed", err)

		return fmt.Errorf("post create policy request failed: %v", err)
	}
	if respCode != http.StatusNoContent {
		// 转成 baseerror
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			// 添加异常时的 trace 属性
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			// 记录异常日志
			otellog.LogError(ctx, "Unmalshal PermissionError failed", err)

			return err
		}

		description := permissionError.Message
		if description == "" {
			description = permissionError.Description
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  description,
				ErrorDetails: permissionError.Cause,
			}}

		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		otellog.LogError(ctx, "Post create policy failed", httpErr)

		return httpErr
	}

	// 添加成功时的 trace 属性
	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return nil
}

// 删除资源策略
func (pa *permissionAccess) DeleteResources(ctx context.Context, res []interfaces.PermissionResource) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "请求删除决策接口")
	defer span.End()

	createUrl := fmt.Sprintf("%s/policy-delete", pa.permissionUrl)

	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
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
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		otellog.LogError(ctx, "Post delete policy request failed", err)

		return fmt.Errorf("post delete policy request failed: %v", err)
	}
	if respCode != http.StatusNoContent {
		// 转成 baseerror
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			// 添加异常时的 trace 属性
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			// 记录异常日志
			otellog.LogError(ctx, "Unmalshal PermissionError failed", err)

			return err
		}

		description := permissionError.Message
		if description == "" {
			description = permissionError.Description
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  description,
				ErrorDetails: permissionError.Cause,
			}}

		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		otellog.LogError(ctx, "Post delete policy failed", httpErr)

		return httpErr
	}

	// 添加成功时的 trace 属性
	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return nil
}

// 策略决策
func (pa *permissionAccess) FilterResources(ctx context.Context,
	filter interfaces.PermissionResourcesFilter) (map[string]interfaces.PermissionResourceOps, error) {

	ctx, span := oteltrace.StartNamedClientSpan(ctx, "请求资源过滤接口")
	defer span.End()

	span.SetAttributes(
		attr.Key("user_id").String(filter.Accessor.ID),
		attr.Key("Operation").StringSlice(filter.Operations),
	)

	httpUrl := fmt.Sprintf("%s/resource-filter", pa.permissionUrl)

	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:            httpUrl,
		HttpMethod:         http.MethodPost,
		HttpContentType:    rest.ContentTypeJson,
		HttpMethodOverride: http.MethodGet,
	})

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME: interfaces.CONTENT_TYPE_JSON,
	}

	filter.Method = http.MethodGet
	respCode, result, err := pa.httpClient.PostNoUnmarshal(ctx, httpUrl, headers, filter)
	logger.Debugf("post [%s] finished, response code is [%d], result is [%s], error is [%v]", httpUrl, respCode, result, err)

	if err != nil {
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Post Failed")
		// 记录异常日志
		otellog.LogError(ctx, "Post resource-filter request failed", err)

		return map[string]interfaces.PermissionResourceOps{}, fmt.Errorf("post resource-filter request failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var permissionError PermissionError
		if err := sonic.Unmarshal(result, &permissionError); err != nil {
			// 添加异常时的 trace 属性
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal PermissionError failed")
			// 记录异常日志
			otellog.LogError(ctx, "Unmalshal PermissionError failed", err)

			return map[string]interfaces.PermissionResourceOps{}, err
		}

		description := permissionError.Message
		if description == "" {
			description = permissionError.Description
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    permissionError.Code,
				Description:  description,
				ErrorDetails: permissionError.Cause,
			}}

		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		otellog.LogError(ctx, "Post resource-filter failed", httpErr)

		return map[string]interfaces.PermissionResourceOps{}, httpErr
	}

	if result == nil {
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		otellog.LogWarn(ctx, "Http response body is null")

		return map[string]interfaces.PermissionResourceOps{}, nil
	}

	allowOps := []struct {
		ResourceID string   `json:"id"`
		Operations []string `json:"allow_operation,omitempty"`
	}{}
	// 处理返回结果 result
	if err := sonic.Unmarshal(result, &allowOps); err != nil {
		// 添加异常时的 trace 属性
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal resource-filter result failed")
		// 记录异常日志
		otellog.LogError(ctx, "Unmalshal resource-filter result failed", err)

		return map[string]interfaces.PermissionResourceOps{}, err
	}

	// 添加成功时的 trace 属性
	oteltrace.AddHttpAttrs4Ok(span, respCode)

	ops := map[string]interfaces.PermissionResourceOps{}
	for _, op := range allowOps {
		ops[op.ResourceID] = interfaces.PermissionResourceOps{
			ResourceID: op.ResourceID,
			Operations: op.Operations,
		}
	}
	return ops, nil
}
