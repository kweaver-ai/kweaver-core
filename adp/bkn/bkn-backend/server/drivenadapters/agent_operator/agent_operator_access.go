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

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		logger.Errorf("get request failed: %v", err)
		return operatorInfo, fmt.Errorf("get request method failed: %w", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			logger.Errorf("unmalshal OperatorError failed: %v\n", err)
			return operatorInfo, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		logger.Errorf("get operator info %s return error %v", agentOperatorID, httpErr.Error())

		return operatorInfo, fmt.Errorf("get operator info %s return error %v", agentOperatorID, httpErr.Error())
	}
	if result == nil {
		return operatorInfo, fmt.Errorf("get operator info %s return null body", agentOperatorID)
	}
	if err = json.Unmarshal(result, &operatorInfo); err != nil {
		logger.Errorf("Unmarshal operator info failed: %s", err)
		return operatorInfo, err
	}
	return operatorInfo, nil
}

// GetToolByID verifies tool-box tool exists (GET .../tool-box/{box_id}/tool/{tool_id}).
func (aoa *agentOperatorAccess) GetToolByID(ctx context.Context, boxID, toolID string) error {
	if boxID == "" || toolID == "" {
		return fmt.Errorf("box_id and tool_id are required for tool binding check")
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

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		logger.Errorf("tool binding check request failed: %v", err)
		return fmt.Errorf("tool binding check failed: %w", err)
	}
	if respCode == http.StatusOK {
		return nil
	}
	if respCode == http.StatusNotFound {
		return fmt.Errorf("tool not found: box_id=%s tool_id=%s", boxID, toolID)
	}
	if respCode != http.StatusOK {
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			logger.Errorf("unmalshal OperatorError failed: %v\n", err)
			return fmt.Errorf("tool binding check failed: %w", err)
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		logger.Errorf("tool binding check failed: %v", httpErr.Error())
		return fmt.Errorf("tool binding check failed: %v", httpErr.Error())
	}
	return nil
}

// CheckMCPToolBinding verifies MCP exposes a tool with toolName (GET .../mcp/proxy/{mcp_id}/tools).
func (aoa *agentOperatorAccess) GetMcpToolByName(ctx context.Context, mcpID, toolName string) error {
	if mcpID == "" || toolName == "" {
		return fmt.Errorf("mcp_id and tool_name are required for MCP tool binding check")
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

	start := time.Now().UnixMilli()
	respCode, result, err := aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] response code [%d], took %dms, err %v",
		url, respCode, time.Now().UnixMilli()-start, err)

	if err != nil {
		logger.Errorf("MCP tools list request failed: %v", err)
		return fmt.Errorf("MCP tool binding check failed: %w", err)
	}
	if respCode != http.StatusOK {
		if respCode == http.StatusNotFound {
			return fmt.Errorf("MCP server not found: mcp_id=%s", mcpID)
		}

		var opError OperatorError
		if len(result) > 0 && json.Unmarshal(result, &opError) == nil && opError.Description != "" {
			return fmt.Errorf("MCP tool binding check failed (status %d): %s", respCode, opError.Description)
		}
		return fmt.Errorf("MCP tool binding check failed: unexpected status %d", respCode)
	}
	var list struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	if err := json.Unmarshal(result, &list); err != nil {
		return fmt.Errorf("parse MCP tools response: %w", err)
	}
	want := strings.TrimSpace(toolName)
	for _, t := range list.Tools {
		if strings.TrimSpace(t.Name) == want {
			return nil
		}
	}
	return fmt.Errorf("MCP tool not found: mcp_id=%s tool_name=%s", mcpID, toolName)
}
