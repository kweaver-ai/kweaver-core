// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"

	"github.com/bytedance/sonic"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

type bknBackendAccess struct {
	logger     interfaces.Logger
	baseURL    string
	httpClient interfaces.HTTPClient
}

var (
	bknAccessOnce sync.Once
	bknAccess     interfaces.BknBackendAccess
)

// NewBknBackendAccess 创建 BknBackendAccess
func NewBknBackendAccess() interfaces.BknBackendAccess {
	bknAccessOnce.Do(func() {
		conf := config.NewConfigLoader()
		bknAccess = &bknBackendAccess{
			logger:     conf.GetLogger(),
			baseURL:    conf.BknBackend.BuildURL("/api/bkn-backend"),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return bknAccess
}

// GetKnowledgeNetworkDetail 获取知识网络详情（include_detail=true, mode=export）
// 对应 Python 的 _get_knowledge_network_detail
func (b *bknBackendAccess) GetKnowledgeNetworkDetail(ctx context.Context, knID string) (*interfaces.KnowledgeNetworkDetail, error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s", b.baseURL, knID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON

	queryValues := url.Values{}
	queryValues.Set("include_detail", "true")
	queryValues.Set("mode", "export")

	respCode, respBody, err := b.httpClient.GetNoUnmarshal(ctx, src, queryValues, header)

	result := &interfaces.KnowledgeNetworkDetail{ID: knID}
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] GetKnowledgeNetworkDetail request failed, err: %v", err)
		return result, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] GetKnowledgeNetworkDetail request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return result, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] GetKnowledgeNetworkDetail get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmarshal KnBaseError failed: %v\n", err)
			return result, err
		}

		return result, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return result, nil
	}

	if err := sonic.Unmarshal(respBody, result); err != nil {
		b.logger.Errorf("[BknBackendAccess] GetKnowledgeNetworkDetail unmarshal failed: %v\n", err)
		return result, err
	}

	return result, nil
}

// SearchObjectTypes 搜索对象类
func (b *bknBackendAccess) SearchObjectTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (objectTypes *interfaces.ObjectTypeConcepts, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/object-types", b.baseURL, query.KnID)
	header := common.GetHeaderFromCtx(ctx)
	header["Content-Type"] = "application/json"
	header["x-http-method-override"] = "GET"
	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, query)

	objectTypes = &interfaces.ObjectTypeConcepts{}
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] SearchObjectTypes request failed, err: %v", err)
		return objectTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] SearchObjectTypes request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return objectTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] SearchObjectTypes， get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return objectTypes, err
		}

		return objectTypes, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return objectTypes, nil
	}

	// 处理返回结果
	if err := sonic.Unmarshal(respBody, objectTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess] SearchObjectTypes unmarshal ObjectTypes failed: %v\n", err)
		return nil, err
	}

	return objectTypes, nil
}

// GetObjectTypeDetail 获取对象类详情
func (b *bknBackendAccess) GetObjectTypeDetail(ctx context.Context, knID string, otIds []string, includeDetail bool) ([]*interfaces.ObjectType, error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/object-types/%s", b.baseURL, knID, strings.Join(otIds, ","))
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	queryValues := url.Values{}
	queryValues.Set("include_detail", strconv.FormatBool(includeDetail))

	respCode, respBody, err := b.httpClient.GetNoUnmarshal(ctx, src, queryValues, header)

	var emptyObjectTypes []*interfaces.ObjectType
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] GetObjectTypeDetail request failed, err: %v", err)
		return emptyObjectTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] GetObjectTypeDetail request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return emptyObjectTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] GetObjectTypeDetail get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return emptyObjectTypes, err
		}

		return emptyObjectTypes, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return emptyObjectTypes, nil
	}

	// 处理返回结果 - 适配新的响应结构 {"entries": []}
	var response struct {
		Entries []*interfaces.ObjectType `json:"entries"`
	}
	if err := sonic.Unmarshal(respBody, &response); err != nil {
		b.logger.Errorf("[BknBackendAccess]GetObjectTypeDetail unmalshal ObjectTypes failed: %v\n", err)
		return emptyObjectTypes, err
	}

	return response.Entries, nil
}

// SearchRelationTypes 搜索关系类
func (b *bknBackendAccess) SearchRelationTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (releationTypes *interfaces.RelationTypeConcepts, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/relation-types", b.baseURL, query.KnID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, query)

	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] SearchRelationTypes request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] SearchRelationTypes request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] SearchRelationTypes get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return nil, err
		}

		return nil, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	releationTypes = &interfaces.RelationTypeConcepts{}
	if len(respBody) == 0 {
		return releationTypes, nil
	}

	// 处理返回结果
	if err := sonic.Unmarshal(respBody, releationTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess]SearchRelationTypes unmalshal RelationTypes failed: %v\n", err)
		return nil, err
	}

	return releationTypes, nil
}

// GetRelationTypeDetail 获取关系类详情
func (b *bknBackendAccess) GetRelationTypeDetail(ctx context.Context, knID string, rtIDs []string, includeDetail bool) ([]*interfaces.RelationType, error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/relation-types/%s", b.baseURL, knID, strings.Join(rtIDs, ","))
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	queryValues := url.Values{}
	queryValues.Set("include_detail", strconv.FormatBool(includeDetail))

	respCode, respBody, err := b.httpClient.GetNoUnmarshal(ctx, src, queryValues, header)

	var emptyRelationTypes []*interfaces.RelationType
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] GetRelationTypeDetail request failed, err: %v", err)
		return emptyRelationTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] GetRelationTypeDetail request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return emptyRelationTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] GetRelationTypeDetail get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return emptyRelationTypes, err
		}

		return emptyRelationTypes, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return emptyRelationTypes, nil
	}

	// 处理返回结果
	var releationTypes []*interfaces.RelationType
	if err := sonic.Unmarshal(respBody, &releationTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess]GetRelationTypeDetail unmalshal releationTypes failed: %v\n", err)
		return emptyRelationTypes, err
	}

	return releationTypes, nil
}

// SearchActionTypes 搜索行动类
func (b *bknBackendAccess) SearchActionTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (actionTypes *interfaces.ActionTypeConcepts, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/action-types", b.baseURL, query.KnID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, query)

	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] SearchActionTypes request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] SearchActionTypes request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] SearchActionTypes get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return nil, err
		}

		return nil, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	actionTypes = &interfaces.ActionTypeConcepts{}
	if len(respBody) == 0 {
		return actionTypes, nil
	}

	// 处理返回结果
	if err := sonic.Unmarshal(respBody, actionTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess]SearchActionTypes unmalshal actionTypes failed: %v\n", err)
		return nil, err
	}

	return actionTypes, nil
}

// SearchMetricTypes 搜索指标类
func (b *bknBackendAccess) SearchMetricTypes(ctx context.Context, query *interfaces.QueryConceptsReq) (metricTypes *interfaces.MetricTypeConcepts, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/metrics", b.baseURL, query.KnID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, query)

	metricTypes = &interfaces.MetricTypeConcepts{}
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] SearchMetricTypes request failed, err: %v", err)
		return metricTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] SearchMetricTypes request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return metricTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] SearchMetricTypes get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return nil, err
		}

		return nil, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return metricTypes, nil
	}

	if err := sonic.Unmarshal(respBody, metricTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess] SearchMetricTypes unmarshal metricTypes failed: %v\n", err)
		return nil, err
	}

	return metricTypes, nil
}

// GetActionTypeDetail 获取行动类详情
func (b *bknBackendAccess) GetActionTypeDetail(ctx context.Context, knID string, atIDs []string, includeDetail bool) ([]*interfaces.ActionType, error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/action-types/%s", b.baseURL, knID, strings.Join(atIDs, ","))
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	header["x-http-method-override"] = "GET"
	queryValues := url.Values{}
	queryValues.Set("include_detail", strconv.FormatBool(includeDetail))

	respCode, respBody, err := b.httpClient.GetNoUnmarshal(ctx, src, queryValues, header)

	var emptyActionTypes []*interfaces.ActionType
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] GetActionTypeDetail request failed, err: %v", err)
		return emptyActionTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] GetActionTypeDetail request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return emptyActionTypes, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] GetActionTypeDetail get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmalshal KnBaseError failed: %v\n", err)
			return emptyActionTypes, err
		}

		return emptyActionTypes, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	if len(respBody) == 0 {
		return emptyActionTypes, nil
	}

	// 处理返回结果
	var actionTypes []*interfaces.ActionType
	if err := sonic.Unmarshal(respBody, &actionTypes); err != nil {
		b.logger.Errorf("[BknBackendAccess]GetActionTypeDetail unmalshal actionTypes failed: %v\n", err)
		return emptyActionTypes, err
	}

	return actionTypes, nil
}

// CreateFullBuildOntologyJob Create a full ontology build job
func (b *bknBackendAccess) CreateFullBuildOntologyJob(ctx context.Context, knID string, req *interfaces.CreateFullBuildOntologyJobReq) (resp *interfaces.CreateJobResp, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/jobs", b.baseURL, knID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON

	// Build request body
	jobReq := map[string]any{
		"name":     req.Name,
		"job_type": interfaces.OntologyJobTypeFull,
	}

	respCode, respBody, err := b.httpClient.PostNoUnmarshal(ctx, src, header, jobReq)
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] CreateFullBuildOntologyJob request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] CreateFullBuildOntologyJob request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] CreateFullBuildOntologyJob get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmarshal KnBaseError failed: %v\n", err)
			return nil, err
		}

		return nil, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	resp = &interfaces.CreateJobResp{}
	if err := sonic.Unmarshal(respBody, resp); err != nil {
		b.logger.Errorf("[BknBackendAccess] CreateFullBuildOntologyJob unmarshal response failed: %v\n", err)
		return nil, err
	}

	return resp, nil
}

// ListOntologyJobs List ontology jobs with filters
func (b *bknBackendAccess) ListOntologyJobs(ctx context.Context, knID string, req *interfaces.ListOntologyJobsReq) (resp *interfaces.ListOntologyJobsResp, err error) {
	src := fmt.Sprintf("%s/in/v1/knowledge-networks/%s/jobs", b.baseURL, knID)
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON

	// Build query parameters
	queryValues := url.Values{}
	if req.NamePattern != "" {
		queryValues.Set("name_pattern", req.NamePattern)
	}
	if req.State != "" {
		queryValues.Set("state", string(req.State))
	}
	if req.JobType != "" {
		queryValues.Set("job_type", string(req.JobType))
	}
	if req.Limit > 0 {
		queryValues.Set("limit", strconv.Itoa(req.Limit))
	}
	if req.Direction != "" {
		queryValues.Set("direction", req.Direction)
	}
	if req.Offset > 0 {
		queryValues.Set("offset", strconv.Itoa(req.Offset))
	}

	respCode, respBody, err := b.httpClient.GetNoUnmarshal(ctx, src, queryValues, header)
	if err != nil {
		b.logger.WithContext(ctx).Errorf("[BknBackendAccess] ListOntologyJobs request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] ListOntologyJobs request failed, err: %v", err))
	}

	if respCode == http.StatusNotFound {
		b.logger.WithContext(ctx).Warnf("[BknBackendAccess] request not found, [%s]", src)
		return nil, infraErr.DefaultHTTPError(ctx, respCode,
			fmt.Sprintf("[BknBackendAccess] request not found, [%s]", src))
	}

	if (respCode < http.StatusOK) || (respCode >= http.StatusMultipleChoices) {
		b.logger.Errorf("[BknBackendAccess] ListOntologyJobs get resp failed, [%s], %v\n", src, respBody)

		var baseError interfaces.KnBaseError
		if err := sonic.Unmarshal(respBody, &baseError); err != nil {
			b.logger.Errorf("unmarshal KnBaseError failed: %v\n", err)
			return nil, err
		}

		return nil, &infraErr.HTTPError{
			HTTPCode:     respCode,
			Code:         baseError.ErrorCode,
			Description:  baseError.Description,
			Solution:     baseError.Solution,
			ErrorLink:    baseError.ErrorLink,
			ErrorDetails: baseError.ErrorDetails,
		}
	}

	resp = &interfaces.ListOntologyJobsResp{}
	if len(respBody) == 0 {
		return resp, nil
	}

	if err := sonic.Unmarshal(respBody, resp); err != nil {
		b.logger.Errorf("[BknBackendAccess] ListOntologyJobs unmarshal response failed: %v\n", err)
		return nil, err
	}

	return resp, nil
}
