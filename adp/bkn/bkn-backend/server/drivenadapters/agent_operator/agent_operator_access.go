// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

package agent_operator

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/otel/otellog"
	"github.com/kweaver-ai/kweaver-go-lib/otel/oteltrace"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"bkn-backend/common"
	"bkn-backend/interfaces"
)

var (
	aoAccessOnce sync.Once
	aoAccess     interfaces.AgentOperatorAccess
)

type agentOperatorAccess struct {
	appSetting       *common.AppSetting
	agentOperatorURL string
	httpClient       rest.HTTPClient
}

type OperatorError struct {
	Code        string      `json:"code"`        // 错误码
	Description string      `json:"description"` // 错误描述
	Detail      interface{} `json:"detail"`      // 详细内容
	Solution    interface{} `json:"solution"`    // 错误解决方案
	Link        interface{} `json:"link"`        // 错误链接
}

// NewAgentOperatorAccess returns a singleton AgentOperatorAccess for operator existence checks.
func NewAgentOperatorAccess(appSetting *common.AppSetting) interfaces.AgentOperatorAccess {
	aoAccessOnce.Do(func() {
		aoAccess = &agentOperatorAccess{
			appSetting:       appSetting,
			agentOperatorURL: appSetting.AgentOperatorUrl,
			httpClient:       common.NewHTTPClient(),
		}
	})
	return aoAccess
}

func (aoa *agentOperatorAccess) GetAgentOperatorByID(ctx context.Context, agentOperatorID string) (interfaces.AgentOperator, error) {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetAgentOperatorByID")
	defer span.End()

	operatorInfo := interfaces.AgentOperator{}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	// GET .../api/agent-operator-integration/internal-v1/operator/market/:operator_id
	url := fmt.Sprintf("%s/operator/market/%s", aoa.agentOperatorURL, agentOperatorID)
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get agent operator failed")
		otellog.LogError(ctx, "Get agent operator request failed", err)
		return operatorInfo, fmt.Errorf("get request method failed: %w", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal OperatorError failed")
			otellog.LogError(ctx, "Unmarshal OperatorError failed", err)
			return operatorInfo, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, fmt.Sprintf("get operator info %s return error", agentOperatorID), httpErr)

		return operatorInfo, fmt.Errorf("get operator info %s return error %v", agentOperatorID, httpErr.Error())
	}
	if result == nil {
		err := fmt.Errorf("get operator info %s return null body", agentOperatorID)
		otellog.LogError(ctx, "Get agent operator returned null body", err)
		return operatorInfo, err
	}
	if err = json.Unmarshal(result, &operatorInfo); err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal operator info failed")
		otellog.LogError(ctx, "Unmarshal operator info failed", err)
		return operatorInfo, err
	}
	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return operatorInfo, nil
}

// GetToolByID verifies tool-box tool exists (GET .../tool-box/{box_id}/tool/{tool_id}).
func (aoa *agentOperatorAccess) GetToolByID(ctx context.Context, boxID, toolID string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetToolByID")
	defer span.End()

	if boxID == "" || toolID == "" {
		err := fmt.Errorf("box_id and tool_id are required for tool binding check")
		otellog.LogError(ctx, "Invalid tool binding parameter", err)
		return err
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	url := fmt.Sprintf("%s/tool-box/%s/tool/%s", aoa.agentOperatorURL, boxID, toolID)
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get tool failed")
		otellog.LogError(ctx, "Tool binding check request failed", err)
		return fmt.Errorf("tool binding check failed: %w", err)
	}
	if respCode == http.StatusOK {
		oteltrace.AddHttpAttrs4Ok(span, respCode)
		return nil
	}
	if respCode == http.StatusNotFound {
		err := fmt.Errorf("tool not found: box_id=%s tool_id=%s", boxID, toolID)
		oteltrace.AddHttpAttrs4Error(span, respCode, "NotFound", "Tool not found")
		otellog.LogError(ctx, "Tool not found", err)
		return err
	}
	if respCode != http.StatusOK {
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unmarshal OperatorError failed")
			otellog.LogError(ctx, "Unmarshal OperatorError failed", err)
			return fmt.Errorf("tool binding check failed: %w", err)
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		oteltrace.AddHttpAttrs4HttpError(span, httpErr)
		otellog.LogError(ctx, "Tool binding check failed", httpErr)
		return fmt.Errorf("tool binding check failed: %v", httpErr.Error())
	}
	oteltrace.AddHttpAttrs4Ok(span, respCode)
	return nil
}

// CheckMCPToolBinding verifies MCP exposes a tool with toolName (GET .../mcp/proxy/{mcp_id}/tools).
func (aoa *agentOperatorAccess) GetMcpToolByName(ctx context.Context, mcpID, toolName string) error {
	ctx, span := oteltrace.StartNamedClientSpan(ctx, "GetMcpToolByName")
	defer span.End()

	if mcpID == "" || toolName == "" {
		err := fmt.Errorf("mcp_id and tool_name are required for MCP tool binding check")
		otellog.LogError(ctx, "Invalid MCP tool binding parameter", err)
		return err
	}

	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}
	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:        interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_ACCOUNT_ID:   accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE: accountInfo.Type,
	}

	url := fmt.Sprintf("%s/mcp/proxy/%s/tools", aoa.agentOperatorURL, mcpID)
	oteltrace.AddAttrs4InternalHttp(span, oteltrace.TraceAttrs{
		HttpUrl:         url,
		HttpMethod:      http.MethodGet,
		HttpContentType: rest.ContentTypeJson,
	})

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Http get MCP tools failed")
		otellog.LogError(ctx, "MCP tools list request failed", err)
		return fmt.Errorf("MCP tool binding check failed: %w", err)
	}
	if respCode != http.StatusOK {
		if respCode == http.StatusNotFound {
			err := fmt.Errorf("MCP server not found: mcp_id=%s", mcpID)
			oteltrace.AddHttpAttrs4Error(span, respCode, "NotFound", "MCP server not found")
			otellog.LogError(ctx, "MCP server not found", err)
			return err
		}

		var opError OperatorError
		if len(result) > 0 && json.Unmarshal(result, &opError) == nil && opError.Description != "" {
			err := fmt.Errorf("MCP tool binding check failed (status %d): %s", respCode, opError.Description)
			oteltrace.AddHttpAttrs4Error(span, respCode, opError.Code, opError.Description)
			otellog.LogError(ctx, "MCP tool binding check failed", err)
			return err
		}
		err := fmt.Errorf("MCP tool binding check failed: unexpected status %d", respCode)
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Unexpected MCP tool binding response status")
		otellog.LogError(ctx, "MCP tool binding check failed", err)
		return err
	}
	var list struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(result, &list); err != nil {
		oteltrace.AddHttpAttrs4Error(span, respCode, "InternalError", "Parse MCP tools response failed")
		otellog.LogError(ctx, "Parse MCP tools response failed", err)
		return fmt.Errorf("parse MCP tools response: %w", err)
	}
	want := strings.TrimSpace(toolName)
	for _, t := range list.Tools {
		if strings.TrimSpace(t.Name) == want {
			oteltrace.AddHttpAttrs4Ok(span, respCode)
			return nil
		}
	}
	err = fmt.Errorf("MCP tool not found: mcp_id=%s tool_name=%s", mcpID, toolName)
	otellog.LogError(ctx, "MCP tool not found", err)
	return err
}
