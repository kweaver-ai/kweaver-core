// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package drivenadapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

type operatorIntegrationClient struct {
	logger     interfaces.Logger
	baseURL    string
	httpClient interfaces.HTTPClient
}

var (
	operatorIntegrationOnce sync.Once
	operatorIntegration     interfaces.DrivenOperatorIntegration
)

const (
	// https://{host}:{port}/api/agent-operator-integration/internal-v1/tool-box/:box_id/tool/:tool_id
	getToolDetailURI = "/internal-v1/tool-box/%s/tool/%s"
	// https://{host}:{port}/api/agent-operator-integration/internal-v1/mcp/proxy/:mcp_id/tools
	getMCPToolListURI = "/internal-v1/mcp/proxy/%s/tools"
	// https://{host}:{port}/api/agent-operator-integration/internal-v1/mcp/proxy/:mcp_id/tool/call
	callMCPToolURI = "/internal-v1/mcp/proxy/%s/tool/call"
	// https://{host}:{port}/api/agent-operator-integration/internal-v1/impex/intcomp/import/toolbox
	syncToolDependencyPackageURI = "/internal-v1/impex/intcomp/import/toolbox"
)

// NewOperatorIntegrationClient 创建 OperatorIntegrationClient
func NewOperatorIntegrationClient() interfaces.DrivenOperatorIntegration {
	operatorIntegrationOnce.Do(func() {
		configLoader := config.NewConfigLoader()
		operatorIntegration = &operatorIntegrationClient{
			logger:     configLoader.GetLogger(),
			baseURL:    configLoader.OperatorIntegration.BuildURL("/api/agent-operator-integration"),
			httpClient: rest.NewHTTPClient(),
		}
	})
	return operatorIntegration
}

// GetToolDetail 获取工具详情
func (o *operatorIntegrationClient) GetToolDetail(ctx context.Context, req *interfaces.GetToolDetailRequest) (resp *interfaces.GetToolDetailResponse, err error) {
	uri := fmt.Sprintf(getToolDetailURI, req.BoxID, req.ToolID)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 记录请求日志
	o.logger.WithContext(ctx).Debugf("[OperatorIntegration#GetToolDetail] URL: %s", url)

	header := common.GetHeaderFromCtx(ctx)

	_, respBody, err := o.httpClient.Get(ctx, url, nil, header)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#GetToolDetail] Request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, fmt.Sprintf("工具详情接口调用失败: %v", err))
	}

	resp = &interfaces.GetToolDetailResponse{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#GetToolDetail] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析工具详情响应失败: %v", err))
		return nil, err
	}

	// 记录响应日志
	o.logger.WithContext(ctx).Debugf("[OperatorIntegration#GetToolDetail] Tool: %s, Name: %s", resp.ToolID, resp.Name)

	return resp, nil
}

// GetMCPToolDetail 获取 MCP 工具详情
func (o *operatorIntegrationClient) GetMCPToolDetail(ctx context.Context, req *interfaces.GetMCPToolDetailRequest) (*interfaces.GetMCPToolDetailResponse, error) {
	uri := fmt.Sprintf(getMCPToolListURI, req.McpID)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 记录请求日志
	o.logger.WithContext(ctx).Debugf("[OperatorIntegration#GetMCPToolDetail] URL: %s", url)

	header := common.GetHeaderFromCtx(ctx)
	_, respBody, err := o.httpClient.Get(ctx, url, nil, header)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#GetMCPToolDetail] Request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, fmt.Sprintf("MCP工具列表接口调用失败: %v", err))
	}

	var listResp struct {
		Tools []interfaces.GetMCPToolDetailResponse `json:"tools"`
	}

	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, &listResp)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#GetMCPToolDetail] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析MCP工具列表响应失败: %v", err))
	}

	for _, tool := range listResp.Tools {
		if tool.Name == req.ToolName {
			// 记录响应日志
			o.logger.WithContext(ctx).Debugf("[OperatorIntegration#GetMCPToolDetail] Found Tool: %s", tool.Name)
			return &tool, nil
		}
	}

	return nil, infraErr.DefaultHTTPError(ctx, http.StatusNotFound, fmt.Sprintf("未找到指定工具: %s", req.ToolName))
}

// CallMCPTool 调用 MCP 工具
func (o *operatorIntegrationClient) CallMCPTool(ctx context.Context, req *interfaces.CallMCPToolRequest) (map[string]interface{}, error) {
	uri := fmt.Sprintf(callMCPToolURI, req.McpID)
	url := fmt.Sprintf("%s%s", o.baseURL, uri)

	// 记录请求日志
	o.logger.WithContext(ctx).Debugf("[OperatorIntegration#CallMCPTool] URL: %s, Tool: %s", url, req.ToolName)

	header := common.GetHeaderFromCtx(ctx)

	// 构建请求体
	reqBody := map[string]interface{}{
		"tool_name":  req.ToolName,
		"parameters": req.Parameters,
	}

	_, respBody, err := o.httpClient.Post(ctx, url, header, reqBody)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#CallMCPTool] Request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, fmt.Sprintf("MCP工具调用接口失败: %v", err))
	}

	var result map[string]interface{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, &result)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#CallMCPTool] Unmarshal failed, body: %s, err: %v", string(resultByt), err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析MCP工具调用响应失败: %v", err))
	}

	return result, nil
}

// SyncToolDependencyPackage 同步工具依赖包
func (o *operatorIntegrationClient) SyncToolDependencyPackage(ctx context.Context, req *interfaces.SyncToolDependencyPackageRequest) (*interfaces.SyncToolDependencyPackageResponse, error) {
	url := fmt.Sprintf("%s%s", o.baseURL, syncToolDependencyPackageURI)
	o.logger.WithContext(ctx).Debugf("[OperatorIntegration#SyncToolDependencyPackage] URL: %s, Mode: %s, PackageVersion: %s",
		url, req.Mode, req.PackageVersion)

	reqBody, contentType, err := buildToolDependencyMultipartRequest(req)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#SyncToolDependencyPackage] Build multipart request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("构建工具依赖导入请求失败: %v", err))
	}

	headers := common.GetHeaderFromCtx(ctx)
	headers["Content-Type"] = contentType

	respCode, respBody, err := o.httpClient.PostNoUnmarshal(ctx, url, headers, reqBody)
	if err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#SyncToolDependencyPackage] Request failed, err: %v", err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusBadGateway, fmt.Sprintf("工具依赖导入接口调用失败: %v", err))
	}
	if respCode < http.StatusOK || respCode >= http.StatusMultipleChoices {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#SyncToolDependencyPackage] Request failed, status: %d, body: %s", respCode, string(respBody))
		return nil, infraErr.DefaultHTTPError(ctx, respCode, fmt.Sprintf("工具依赖导入接口返回异常: %s", string(respBody)))
	}
	if len(respBody) == 0 {
		return &interfaces.SyncToolDependencyPackageResponse{}, nil
	}

	resp := &interfaces.SyncToolDependencyPackageResponse{}
	if err = json.Unmarshal(respBody, resp); err != nil {
		o.logger.WithContext(ctx).Errorf("[OperatorIntegration#SyncToolDependencyPackage] Unmarshal failed, body: %s, err: %v", string(respBody), err)
		return nil, infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, fmt.Sprintf("解析工具依赖导入响应失败: %v", err))
	}
	return resp, nil
}

func buildToolDependencyMultipartRequest(req *interfaces.SyncToolDependencyPackageRequest) ([]byte, string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	if req.Mode != "" {
		if err := writer.WriteField("mode", req.Mode); err != nil {
			return nil, "", err
		}
	}
	if err := writer.WriteField("package_version", req.PackageVersion); err != nil {
		return nil, "", err
	}
	part, err := writer.CreateFormFile("data", "execution_factory_tools.adp")
	if err != nil {
		return nil, "", err
	}
	if _, err = part.Write(req.PackageData); err != nil {
		return nil, "", err
	}
	if err = writer.Close(); err != nil {
		return nil, "", err
	}
	return body.Bytes(), writer.FormDataContentType(), nil
}
