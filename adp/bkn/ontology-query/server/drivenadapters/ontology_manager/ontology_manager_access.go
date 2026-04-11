// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package ontology_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/kweaver-ai/TelemetrySDK-Go/exporter/v2/ar_trace"
	"github.com/kweaver-ai/kweaver-go-lib/logger"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
	"github.com/kweaver-ai/kweaver-go-lib/rest"
	"go.opentelemetry.io/otel/trace"

	"ontology-query/common"
	"ontology-query/interfaces"
)

var (
	omAccessOnce sync.Once
	omAccess     interfaces.OntologyManagerAccess
)

type ontologyManagerAccess struct {
	appSetting         *common.AppSetting
	ontologyManagerUrl string
	httpClient         rest.HTTPClient
}

func NewOntologyManagerAccess(appSetting *common.AppSetting) interfaces.OntologyManagerAccess {
	omAccessOnce.Do(func() {
		omAccess = &ontologyManagerAccess{
			appSetting:         appSetting,
			ontologyManagerUrl: appSetting.BKNBackendUrl,
			httpClient:         common.NewHTTPClient(),
		}
	})
	return omAccess
}

// 获取对象类信息
func (oma *ontologyManagerAccess) GetObjectType(ctx context.Context, knID string, branch string, otID string) (interfaces.ObjectType, bool, error) {

	httpUrl := fmt.Sprintf("%s/%s/object-types/%s?branch=%s", oma.ontologyManagerUrl, knID, otID, branch)
	// http client 发送请求时，在 RoundTrip 时是用 transport 在 RoundTrip，此时的 transport 是 otelhttp.NewTransport 的，
	// otelhttp.NewTransport 的 RoundTrip 时会对 propagator 做 inject, 即 t.propagators.Inject
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 ontology-manager 获取对象类信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	var emptyObjectType interfaces.ObjectType
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get object type request failed: %v", err))

		return emptyObjectType, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("object type %s not exists", otID)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("Metric model [%s] not found", otID))

		return emptyObjectType, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get object type failed: %v", result)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return emptyObjectType, false, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get object type failed: %v", httpErr))

		return emptyObjectType, false, fmt.Errorf("get object type failed: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyObjectType, false, nil
	}

	// 处理返回结果 result
	// var objectTypes []interfaces.ObjectType
	var response struct {
		ObjectTypes []interfaces.ObjectType `json:"entries"`
	}

	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("unmalshal object type info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal object type info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal object type info failed: %v", err))

		return emptyObjectType, false, err
	}

	if len(response.ObjectTypes) == 0 {
		return emptyObjectType, false, nil
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return response.ObjectTypes[0], true, nil
}

func (oma *ontologyManagerAccess) GetRelationTypePathsBaseOnSource(ctx context.Context, knID string,
	branch string, query interfaces.PathsQueryBaseOnSource) ([]interfaces.RelationTypePath, error) {

	url := fmt.Sprintf("%s/%s/relation-type-paths?branch=%s", oma.ontologyManagerUrl, knID, branch)

	ctx, span := ar_trace.Tracer.Start(ctx, "请求 ontology-manager 获取关系类路径信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodPost,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

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

	start := time.Now().UnixMilli()
	respCode, result, err := oma.httpClient.PostNoUnmarshal(ctx, url, headers, query)
	logger.Debugf("post [%s] with headers[%v] finished, request is [%v] response code is [%d],  error is [%v], 耗时: %dms",
		url, headers, query, respCode, err, time.Now().UnixMilli()-start)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get relation type paths request failed: %v", err))

		return nil, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("get relation type paths failed: %v", result)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return nil, err
		}

		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get relation type paths failed: %v", httpErr))

		return nil, fmt.Errorf("get relation type paths failed: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return nil, nil
	}

	// 处理返回结果 result
	// var typePaths []interfaces.RelationTypePath
	var response struct {
		TypePaths []interfaces.RelationTypePath `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("unmalshal object type info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal object type info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal relation type paths info failed: %v", err))

		return nil, err
	}

	if len(response.TypePaths) == 0 {
		return nil, nil
	}

	// 对关系类的映射在这转
	for i := range response.TypePaths {
		// 生成路径id，简单编号即可，在当前查询中唯一即可，为后续的路径配额使用
		response.TypePaths[i].ID = i
		for j := range response.TypePaths[i].TypeEdges {
			switch response.TypePaths[i].TypeEdges[j].RelationType.Type {
			case interfaces.RELATION_TYPE_DIRECT:
				var directMapping []interfaces.Mapping
				jsonData, err := json.Marshal(response.TypePaths[i].TypeEdges[j].RelationType.MappingRules)
				if err != nil {
					return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
				}
				err = json.Unmarshal(jsonData, &directMapping)
				if err != nil {
					return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
				}
				response.TypePaths[i].TypeEdges[j].RelationType.MappingRules = directMapping
			case interfaces.RELATION_TYPE_DATA_VIEW:
				var inDirectMapping interfaces.InDirectMapping
				jsonData, err := json.Marshal(response.TypePaths[i].TypeEdges[j].RelationType.MappingRules)
				if err != nil {
					return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
				}
				err = json.Unmarshal(jsonData, &inDirectMapping)
				if err != nil {
					return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
				}
				response.TypePaths[i].TypeEdges[j].RelationType.MappingRules = inDirectMapping
			case interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN:
				var fcj interfaces.FilteredCrossJoinMapping
				jsonData, err := json.Marshal(response.TypePaths[i].TypeEdges[j].RelationType.MappingRules)
				if err != nil {
					return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
				}
				err = json.Unmarshal(jsonData, &fcj)
				if err != nil {
					return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
				}
				response.TypePaths[i].TypeEdges[j].RelationType.MappingRules = fcj
			}
		}
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return response.TypePaths, nil
}

func (oma *ontologyManagerAccess) GetRelationType(ctx context.Context, knID string,
	branch string, rtID string) (interfaces.RelationType, bool, error) {

	httpUrl := fmt.Sprintf("%s/%s/relation-types/%s?branch=%s", oma.ontologyManagerUrl, knID, rtID, branch)
	// http client 发送请求时，在 RoundTrip 时是用 transport 在 RoundTrip，此时的 transport 是 otelhttp.NewTransport 的，
	// otelhttp.NewTransport 的 RoundTrip 时会对 propagator 做 inject, 即 t.propagators.Inject
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 ontology-manager 获取关系类信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	var emptyRelationType interfaces.RelationType
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get relation type request failed: %v", err))

		return emptyRelationType, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("relation type %s not exists", rtID)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("relation type [%s] not found", rtID))

		return emptyRelationType, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get relation type failed: %v", result)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return emptyRelationType, false, err
		}

		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get relation type failed: %v", httpErr))

		return emptyRelationType, false, fmt.Errorf("get relation type failed: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyRelationType, false, nil
	}

	// 处理返回结果 result
	// var relationTypes []interfaces.RelationType
	var response struct {
		RelationTypes []interfaces.RelationType `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("unmalshal relation type info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal relation type info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal relation type info failed: %v", err))

		return emptyRelationType, false, err
	}

	if len(response.RelationTypes) == 0 {
		return emptyRelationType, false, nil
	}

	switch response.RelationTypes[0].Type {
	case interfaces.RELATION_TYPE_DIRECT:
		var directMapping []interfaces.Mapping
		jsonData, err := json.Marshal(response.RelationTypes[0].MappingRules)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Marshal error: %s", err.Error())
		}
		err = json.Unmarshal(jsonData, &directMapping)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
		}
		response.RelationTypes[0].MappingRules = directMapping
	case interfaces.RELATION_TYPE_DATA_VIEW:
		var inDirectMapping interfaces.InDirectMapping
		jsonData, err := json.Marshal(response.RelationTypes[0].MappingRules)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Marshal error: %s", err.Error())
		}
		err = json.Unmarshal(jsonData, &inDirectMapping)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
		}
		response.RelationTypes[0].MappingRules = inDirectMapping
	case interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN:
		var fcj interfaces.FilteredCrossJoinMapping
		jsonData, err := json.Marshal(response.RelationTypes[0].MappingRules)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Marshal error: %s", err.Error())
		}
		err = json.Unmarshal(jsonData, &fcj)
		if err != nil {
			return emptyRelationType, false, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
		}
		response.RelationTypes[0].MappingRules = fcj
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return response.RelationTypes[0], true, nil
}

func (oma *ontologyManagerAccess) ListRelationTypes(ctx context.Context, knID string,
	branch string, query interfaces.RelationTypesQuery) ([]interfaces.RelationType, error) {

	httpUrl := fmt.Sprintf("%s/%s/relation-types?branch=%s&limit=-1", oma.ontologyManagerUrl, knID, branch)

	// 支持多个对象类型ID查询
	if len(query.SourceObjectTypeIDs) > 0 {
		for _, otID := range query.SourceObjectTypeIDs {
			httpUrl += fmt.Sprintf("&source_object_type_id=%s", otID)
		}
	}

	if len(query.TargetObjectTypeIDs) > 0 {
		for _, otID := range query.TargetObjectTypeIDs {
			httpUrl += fmt.Sprintf("&target_object_type_id=%s", otID)
		}
	}

	ctx, span := ar_trace.Tracer.Start(ctx, "请求 ontology-manager 获取关系类列表", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, result, err = oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	logger.Debugf("get [%s] with headers[%v] finished,response code is [%d],  error is [%v]",
		httpUrl, headers, respCode, err)

	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		o11y.Error(ctx, fmt.Sprintf("List relation types request failed: %v", err))
		return nil, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("list relation types failed: %v", result)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaseError failed: %v\n", err)
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaseError failed")
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaseError failed: %v", err))
			return nil, err
		}

		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		o11y.Error(ctx, fmt.Sprintf("List relation types failed: %v", httpErr))
		return nil, fmt.Errorf("list relation types failed: %v", httpErr.Error())
	}

	if result == nil {
		o11y.AddHttpAttrs4Ok(span, respCode)
		o11y.Warn(ctx, "Http response body is null")
		return []interfaces.RelationType{}, nil
	}

	var response struct {
		RelationTypes []interfaces.RelationType `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("unmalshal relation types info failed: %v\n", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal relation types info failed")
		o11y.Error(ctx, fmt.Sprintf("Unmalshal relation types info failed: %v", err))
		return nil, err
	}

	if len(response.RelationTypes) == 0 {
		return []interfaces.RelationType{}, nil
	}

	// 转换每个关系类的MappingRules
	for i := range response.RelationTypes {
		switch response.RelationTypes[i].Type {
		case interfaces.RELATION_TYPE_DIRECT:
			var directMapping []interfaces.Mapping
			jsonData, err := json.Marshal(response.RelationTypes[i].MappingRules)
			if err != nil {
				return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
			}
			err = json.Unmarshal(jsonData, &directMapping)
			if err != nil {
				return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
			}
			response.RelationTypes[i].MappingRules = directMapping
		case interfaces.RELATION_TYPE_DATA_VIEW:
			var inDirectMapping interfaces.InDirectMapping
			jsonData, err := json.Marshal(response.RelationTypes[i].MappingRules)
			if err != nil {
				return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
			}
			err = json.Unmarshal(jsonData, &inDirectMapping)
			if err != nil {
				return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
			}
			response.RelationTypes[i].MappingRules = inDirectMapping
		case interfaces.RELATION_TYPE_FILTERED_CROSS_JOIN:
			var fcj interfaces.FilteredCrossJoinMapping
			jsonData, err := json.Marshal(response.RelationTypes[i].MappingRules)
			if err != nil {
				return nil, fmt.Errorf("derived Config Marshal error: %s", err.Error())
			}
			err = json.Unmarshal(jsonData, &fcj)
			if err != nil {
				return nil, fmt.Errorf("derived Config Unmarshal error: %s", err.Error())
			}
			response.RelationTypes[i].MappingRules = fcj
		}
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return response.RelationTypes, nil
}

func (oma *ontologyManagerAccess) GetActionType(ctx context.Context, knID string,
	branch string, atID string) (interfaces.ActionType, map[string]any, bool, error) {

	httpUrl := fmt.Sprintf("%s/%s/action-types/%s?branch=%s", oma.ontologyManagerUrl, knID, atID, branch)
	// http client 发送请求时，在 RoundTrip 时是用 transport 在 RoundTrip，此时的 transport 是 otelhttp.NewTransport 的，
	// otelhttp.NewTransport 的 RoundTrip 时会对 propagator 做 inject, 即 t.propagators.Inject
	ctx, span := ar_trace.Tracer.Start(ctx, "请求 ontology-manager 获取行动类信息", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	var (
		respCode int
		result   []byte
		err      error
	)

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}
	// httpClient 的请求新增参数支持上下文的处理请求的函数
	respCode, result, err = oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)

	var emptyActionType interfaces.ActionType
	if err != nil {
		logger.Errorf("get request method failed: %v", err)
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get action type request failed: %v", err))

		return emptyActionType, nil, false, fmt.Errorf("get request method failed: %v", err)
	}

	if respCode == http.StatusNotFound {
		logger.Errorf("action type %s not exists", atID)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, fmt.Sprintf("action type [%s] not found", atID))

		return emptyActionType, nil, false, nil
	}

	if respCode != http.StatusOK {
		logger.Errorf("get action type failed: %v", result)

		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)

			// 添加异常时的 trace 属性
			o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal BaesError failed")
			// 记录异常日志
			o11y.Error(ctx, fmt.Sprintf("Unmalshal BaesError failed: %v", err))

			return emptyActionType, nil, false, err
		}

		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http status is not 200")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Get action type failed: %v", httpErr))

		return emptyActionType, nil, false, fmt.Errorf("get action type failed: %v", httpErr.Error())
	}

	if result == nil {
		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Ok(span, respCode)
		// 记录模型不存在的日志
		o11y.Warn(ctx, "Http response body is null")

		return emptyActionType, nil, false, nil
	}

	// 处理返回结果 result - 解析为结构体
	var response struct {
		ActionTypes []interfaces.ActionType `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("unmalshal action type info failed: %v\n", err)

		// 添加异常时的 trace 属性
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmalshal action type info failed")
		// 记录异常日志
		o11y.Error(ctx, fmt.Sprintf("Unmalshal action type info failed: %v", err))

		return emptyActionType, nil, false, err
	}

	if len(response.ActionTypes) == 0 {
		return emptyActionType, nil, false, nil
	}

	// 同时解析原始 JSON 为 map[string]any，保留完整数据
	var rawResponse struct {
		Entries []map[string]any `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &rawResponse); err != nil {
		logger.Errorf("unmalshal raw action type info failed: %v\n", err)
		// 仍然返回解析后的结构体，但原始数据为 nil
		o11y.AddHttpAttrs4Ok(span, respCode)
		return response.ActionTypes[0], nil, true, nil
	}

	var rawActionType map[string]any
	if len(rawResponse.Entries) > 0 {
		rawActionType = rawResponse.Entries[0]
	}

	// 添加成功时的 trace 属性
	o11y.AddHttpAttrs4Ok(span, respCode)

	return response.ActionTypes[0], rawActionType, true, nil
}

// GetRiskTypesByIDs 按 ID 批量获取风险类（调用 bkn-backend 内部 API）
func (oma *ontologyManagerAccess) GetRiskTypesByIDs(ctx context.Context, knID string, branch string, riskTypeIDs []string) ([]interfaces.RiskType, error) {
	if len(riskTypeIDs) == 0 {
		return []interfaces.RiskType{}, nil
	}

	idsStr := strings.Join(riskTypeIDs, ",")
	httpUrl := fmt.Sprintf("%s/%s/risk-types?risk_type_ids=%s&branch=%s", oma.ontologyManagerUrl, knID, idsStr, branch)

	ctx, span := ar_trace.Tracer.Start(ctx, "请求 bkn-backend 获取风险类", trace.WithSpanKind(trace.SpanKindClient))
	o11y.AddAttrs4InternalHttp(span, o11y.TraceAttrs{
		HttpUrl:         httpUrl,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})
	defer span.End()

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	respCode, result, err := oma.httpClient.GetNoUnmarshal(ctx, httpUrl, nil, headers)
	if err != nil {
		logger.Errorf("GetRiskTypesByIDs request failed: %v", err)
		o11y.AddHttpAttrs4Error(span, respCode, "InternalError", "Http Get Failed")
		return nil, fmt.Errorf("get risk types failed: %w", err)
	}

	if respCode != http.StatusOK {
		logger.Errorf("GetRiskTypesByIDs failed: %v", result)
		var baseError rest.BaseError
		if err = sonic.Unmarshal(result, &baseError); err != nil {
			return nil, fmt.Errorf("get risk types failed: %s", string(result))
		}
		return nil, fmt.Errorf("get risk types failed: %s", baseError.Description)
	}

	var response struct {
		Entries []interfaces.RiskType `json:"entries"`
	}
	if err = sonic.Unmarshal(result, &response); err != nil {
		logger.Errorf("Unmarshal risk types failed: %v", err)
		return nil, fmt.Errorf("unmarshal risk types failed: %w", err)
	}

	o11y.AddHttpAttrs4Ok(span, respCode)
	return response.Entries, nil
}
