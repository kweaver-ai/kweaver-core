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
	"sync"
	"time"

	"github.com/kweaver-ai/kweaver-go-lib/logger"
	"github.com/kweaver-ai/kweaver-go-lib/rest"

	"ontology-query/common"
	"ontology-query/interfaces"
)

var (
	aoAccessOnce sync.Once
	aoAccess     interfaces.AgentOperatorAccess
)

type agentOperatorAccess struct {
	appSetting       *common.AppSetting
	agentOperatorUrl string
	httpClient       rest.HTTPClient
}

type OperatorError struct {
	Code        string      `json:"code"`        // 错误码
	Description string      `json:"description"` // 错误描述
	Detail      interface{} `json:"detail"`      // 详细内容
	Solution    interface{} `json:"solution"`    // 错误解决方案
	Link        interface{} `json:"link"`        // 错误链接
}

func NewAgentOperatorAccess(appSetting *common.AppSetting) interfaces.AgentOperatorAccess {
	aoAccessOnce.Do(func() {
		aoAccess = &agentOperatorAccess{
			appSetting:       appSetting,
			agentOperatorUrl: appSetting.AgentOperatorUrl,
			httpClient:       common.NewHTTPClient(),
		}
	})

	return aoAccess
}

func (aoa *agentOperatorAccess) GetAgentOperatorByID(ctx context.Context,
	agentOperatorID string) (interfaces.AgentOperator, error) {

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

	// http://{host}:{port}/api/agent-operator-integration/internal-v1/operator/market/:operator_id
	url := fmt.Sprintf("%s/market/%s", aoa.agentOperatorUrl, agentOperatorID)

	start := time.Now().UnixMilli()
	respCode, result, err = aoa.httpClient.GetNoUnmarshal(ctx, url, nil, headers)
	logger.Debugf("get [%s] with headers[%v] finished, request is [%v] response code is [%d],  error is [%v], 耗时: %dms",
		url, headers, respCode, err, time.Now().UnixMilli()-start)

	operatorInfo := interfaces.AgentOperator{}

	if err != nil {
		logger.Errorf("get request method failed: %v", err)

		return operatorInfo, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var baseError rest.BaseError
		if err = json.Unmarshal(result, &baseError); err != nil {
			logger.Errorf("unmalshal BaesError failed: %v\n", err)
			return operatorInfo, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode, BaseError: baseError}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		return operatorInfo, fmt.Errorf("get operator info %s return error %v", agentOperatorID, httpErr.Error())
	}

	if result == nil {
		return operatorInfo, fmt.Errorf("get operator info %v return null", agentOperatorID)
	}

	if err = json.Unmarshal(result, &operatorInfo); err != nil {
		logger.Errorf("Unmarshal operator info failed, %s", err)

		return operatorInfo, err
	}

	return operatorInfo, nil
}

type operatorExecuteResult struct {
	StatusCode int            `json:"status_code"`
	Headers    map[string]any `json:"headers"`
	Body       any            `json:"body"` // Can receive any type (string, map, array, etc.)
	Error      string         `json:"error"`
	DurationMs int            `json:"duration_ms"`
}

// 执行算子获取结果
func (aoa *agentOperatorAccess) ExecuteOperator(ctx context.Context, agentOperatorID string,
	execRequest interfaces.OperatorExecutionRequest) (any, error) {

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
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_METHOD_OVERRIDE: http.MethodGet,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}
	// http://{host}:{port}/api/agent-operator-integration/internal-v1/operator/proxy/{operator_id}
	url := fmt.Sprintf("%s/proxy/%s", aoa.agentOperatorUrl, agentOperatorID)

	start := time.Now().UnixMilli()
	respCode, result, err = aoa.httpClient.PostNoUnmarshal(ctx, url, headers, execRequest)
	logger.Debugf("post [%s] with headers[%v] finished, request is [%v] response code is [%d],  error is [%v], 耗时: %dms",
		url, headers, execRequest, respCode, err, time.Now().UnixMilli()-start)

	operatorResult := operatorExecuteResult{}

	if err != nil {
		logger.Errorf("get request method failed: %v", err)

		return operatorResult, fmt.Errorf("get request method failed: %v", err)
	}
	if respCode != http.StatusOK {
		// 转成 baseerror
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			logger.Errorf("unmalshal OperatorError failed: %v\n", err)
			return operatorResult, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		logger.Errorf("Formula invalid: %v", httpErr.Error())

		return operatorResult, fmt.Errorf("execute operator %s return error %v", agentOperatorID, httpErr.Error())
	}

	if result == nil {
		return operatorResult, fmt.Errorf("execute operator %v return null", agentOperatorID)
	}

	if err := json.Unmarshal(result, &operatorResult); err != nil {
		logger.Errorf("Unmarshal execute operator result failed, %s", err)

		return operatorResult, err
	}

	// status_code 在100-300间才算成功
	if http.StatusContinue <= operatorResult.StatusCode &&
		operatorResult.StatusCode < http.StatusMultipleChoices {

		return operatorResult.Body, nil
	} else {
		resByte, err := json.Marshal(operatorResult)
		if err != nil {
			logger.Errorf("unmalshal OperatorError failed: %v\n", err)
			return operatorResult, err
		}

		return nil, fmt.Errorf("execute operator failed: %v", string(resByte))
	}
}

// ExecuteTool executes a tool via tool-box API
// API: POST /tool-box/{box_id}/proxy/{tool_id}
func (aoa *agentOperatorAccess) ExecuteTool(ctx context.Context, boxID string,
	toolID string, execRequest interfaces.ToolExecutionRequest) (any, error) {

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

	// http://{host}:{port}/api/agent-operator-integration/internal-v1/tool-box/{box_id}/proxy/{tool_id}
	url := fmt.Sprintf("%s/%s/proxy/%s", aoa.appSetting.ToolBoxUrl, boxID, toolID)

	start := time.Now().UnixMilli()
	respCode, result, err = aoa.httpClient.PostNoUnmarshal(ctx, url, headers, execRequest)
	logger.Debugf("post [%s] with headers[%v] finished, request is [%v] response code is [%d], error is [%v], 耗时: %dms",
		url, headers, execRequest, respCode, err, time.Now().UnixMilli()-start)

	toolResult := operatorExecuteResult{}

	if err != nil {
		logger.Errorf("Tool execution request failed: %v", err)
		return toolResult, fmt.Errorf("tool execution request failed: %v", err)
	}

	if respCode != http.StatusOK {
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			logger.Errorf("unmarshal ToolError failed: %v", err)
			return toolResult, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		logger.Errorf("Tool execution failed: %v", httpErr.Error())
		return toolResult, fmt.Errorf("execute tool %s/%s return error %v", boxID, toolID, httpErr.Error())
	}

	if result == nil {
		return toolResult, fmt.Errorf("execute tool %s/%s return null", boxID, toolID)
	}

	if err := json.Unmarshal(result, &toolResult); err != nil {
		logger.Errorf("Unmarshal tool execution result failed, %s", err)
		return toolResult, err
	}

	// status_code 在100-300间才算成功
	if http.StatusContinue <= toolResult.StatusCode &&
		toolResult.StatusCode < http.StatusMultipleChoices {
		return toolResult.Body, nil
	} else {
		resByte, err := json.Marshal(toolResult)
		if err != nil {
			logger.Errorf("marshal tool result failed: %v", err)
			return toolResult, err
		}
		return nil, fmt.Errorf("execute tool failed: %v", string(resByte))
	}
}

// ExecuteMCP executes an MCP-based action through agent-operator-integration
// API: POST /mcp/proxy/{mcp_id}/tool/call
func (aoa *agentOperatorAccess) ExecuteMCP(ctx context.Context, mcpID string,
	toolName string, execRequest interfaces.MCPExecutionRequest) (any, error) {

	var (
		respCode int
		result   []byte
		err      error
	)

	// Get account info from context for user_id header
	accountInfo := interfaces.AccountInfo{}
	if ctx.Value(interfaces.ACCOUNT_INFO_KEY) != nil {
		accountInfo = ctx.Value(interfaces.ACCOUNT_INFO_KEY).(interfaces.AccountInfo)
	}

	// Get business domain from context (passed from request header)
	businessDomain := ""
	if ctx.Value(interfaces.BUSINESS_DOMAIN_KEY) != nil {
		businessDomain = ctx.Value(interfaces.BUSINESS_DOMAIN_KEY).(string)
	}

	headers := map[string]string{
		interfaces.CONTENT_TYPE_NAME:           interfaces.CONTENT_TYPE_JSON,
		interfaces.HTTP_HEADER_BUSINESS_DOMAIN: businessDomain,
		interfaces.HTTP_HEADER_ACCOUNT_ID:      accountInfo.ID,
		interfaces.HTTP_HEADER_ACCOUNT_TYPE:    accountInfo.Type,
	}

	// http://{host}:{port}/api/agent-operator-integration/internal-v1/mcp/proxy/{mcp_id}/tool/call
	url := fmt.Sprintf("%s/proxy/%s/tool/call", aoa.appSetting.MCPUrl, mcpID)

	start := time.Now().UnixMilli()
	respCode, result, err = aoa.httpClient.PostNoUnmarshal(ctx, url, headers, execRequest)
	logger.Debugf("post [%s] with headers[%v] finished, request is [%v] response code is [%d], error is [%v], 耗时: %dms",
		url, headers, execRequest, respCode, err, time.Now().UnixMilli()-start)

	mcpResult := operatorExecuteResult{}

	if err != nil {
		logger.Errorf("MCP execution request failed: %v", err)
		return mcpResult, fmt.Errorf("MCP execution request failed: %v", err)
	}

	if respCode != http.StatusOK {
		var opError OperatorError
		if err = json.Unmarshal(result, &opError); err != nil {
			logger.Errorf("unmarshal OperatorError failed: %v\n", err)
			return mcpResult, err
		}
		httpErr := &rest.HTTPError{HTTPCode: respCode,
			BaseError: rest.BaseError{
				ErrorCode:    opError.Code,
				Description:  opError.Description,
				ErrorDetails: opError.Detail,
			}}
		logger.Errorf("MCP execution failed: %v", httpErr.Error())
		return mcpResult, fmt.Errorf("execute MCP %s return error %v", mcpID, httpErr.Error())
	}

	if result == nil {
		return mcpResult, fmt.Errorf("execute MCP %s return null", mcpID)
	}

	if err := json.Unmarshal(result, &mcpResult); err != nil {
		logger.Errorf("Unmarshal MCP execution result failed, %s", err)
		return mcpResult, err
	}

	// status_code 在100-300间才算成功
	if http.StatusContinue <= mcpResult.StatusCode &&
		mcpResult.StatusCode < http.StatusMultipleChoices {
		return mcpResult.Body, nil
	} else {
		resByte, err := json.Marshal(mcpResult)
		if err != nil {
			logger.Errorf("marshal MCP result failed: %v\n", err)
			return mcpResult, err
		}
		return nil, fmt.Errorf("execute MCP failed: %v", string(resByte))
	}
}
