// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package drivenadapters
// file: agent_app.go
// desc: 智能体App接口
package drivenadapters

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/common"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/config"
	infraErr "github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/errors"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/infra/rest"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/utils"
)

type agentClient struct {
	logger      interfaces.Logger
	baseURL     string
	httpClient  interfaces.HTTPClient
	DeployAgent config.DeployAgentConfig
}

var (
	agentOnce sync.Once
	ag        interfaces.AgentApp
)

const (
	// http://agent-factory:13020/api/agent-factory/internal/v1/app/{app_key}/api/chat/completion
	chatURI = "/internal/v1/app/%s/api/chat/completion"
)

// NewAgentAppClient 新建AgentAppClient
func NewAgentAppClient() interfaces.AgentApp {
	agentOnce.Do(func() {
		configLoader := config.NewConfigLoader()
		ag = &agentClient{
			logger:      configLoader.GetLogger(),
			baseURL:     configLoader.AgentApp.BuildURL("/api/agent-factory"),
			httpClient:  rest.NewHTTPClient(),
			DeployAgent: configLoader.DeployAgent,
		}
	})
	return ag
}

// APIChat 智能体API调用
func (a *agentClient) APIChat(ctx context.Context, req *interfaces.ChatRequest) (resp *interfaces.ChatResponse, err error) {
	url := fmt.Sprintf("%s%s", a.baseURL, fmt.Sprintf(chatURI, req.AgentKey))
	header := common.GetHeaderFromCtx(ctx)
	header[rest.ContentTypeKey] = rest.ContentTypeJSON
	_, respBody, err := a.httpClient.Post(ctx, url, header, req)
	if err != nil {
		a.logger.WithContext(ctx).Warnf("[AgentApp#ApiChat] ApiChat request failed, err: %v", err)
		return
	}

	resp = &interfaces.ChatResponse{}
	resultByt := utils.ObjectToByte(respBody)
	err = json.Unmarshal(resultByt, resp)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ApiChat] Unmarshal %s err:%v", string(resultByt), err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
	}
	return
}

// ConceptIntentionAnalysisAgent 概念意图识分析智能体 app
func (a *agentClient) ConceptIntentionAnalysisAgent(ctx context.Context,
	req *interfaces.ConceptIntentionAnalysisAgentReq) (queryUnderstandResult *interfaces.QueryUnderstanding, err error) {
	customQuerys := make(map[string]any)
	if len(req.PreviousQueries) > 0 {
		customQuerys["previous_queries"] = req.PreviousQueries
		customQuerys["kn_id"] = req.KnID
	}
	chatReq := &interfaces.ChatRequest{
		AgentKey:     a.DeployAgent.ConceptIntentionAnalysisAgentKey,
		Stream:       false,
		Query:        req.Query,
		CustomQuerys: customQuerys,
		AgentVersion: "latest",
	}
	result, err := a.APIChat(ctx, chatReq)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptIntentionAnalysisAgent] APIChat err:%v", err)
		return
	}

	// 输出内容判断
	var text string
	if result != nil && result.Message != nil && result.Message.Content != nil && result.Message.Content.FinalAnswer != nil && result.Message.Content.FinalAnswer.Answer != nil {
		text = result.Message.Content.FinalAnswer.Answer.Text
	}

	// 解析输出内容
	resultStr, err := parseResultFromAgentV1Answer(text)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptIntentionAnalysisAgent] parseResultFromAgentV1Answer err:%v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	queryUnderstandResult = &interfaces.QueryUnderstanding{}
	err = json.Unmarshal([]byte(resultStr), queryUnderstandResult)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptIntentionAnalysisAgent] Unmarshal %s err:%v", resultStr, err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	return queryUnderstandResult, nil
}

// ConceptRetrievalStrategistAgent 概念召回策略智能体 app
func (a *agentClient) ConceptRetrievalStrategistAgent(ctx context.Context,
	req *interfaces.ConceptRetrievalStrategistReq) (queryStrategys []*interfaces.SemanticQueryStrategy, err error) {
	customQuerys := make(map[string]any)
	if len(req.PreviousQueries) > 0 {
		customQuerys["previous_queries"] = req.PreviousQueries
		customQuerys["kn_id"] = req.KnID
	}
	chatReq := &interfaces.ChatRequest{
		AgentKey:     a.DeployAgent.ConceptRetrievalStrategistAgentKey,
		Stream:       false,
		Query:        utils.ObjectToJSON(req.QueryParam),
		CustomQuerys: customQuerys,
	}
	result, err := a.APIChat(ctx, chatReq)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptIntentionAnalysisAgent] APIChat err:%v", err)
		return
	}
	// 输出内容判断
	var text string
	if result != nil && result.Message != nil && result.Message.Content != nil && result.Message.Content.FinalAnswer != nil && result.Message.Content.FinalAnswer.Answer != nil {
		text = result.Message.Content.FinalAnswer.Answer.Text
	}
	// 解析输出内容
	resultStr, err := parseResultFromAgentV1Answer(text)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptRetrievalStrategistAgent] parseResultFromAgentV1Answer err:%v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	queryUnderstanding := &interfaces.QueryUnderstanding{}
	err = json.Unmarshal([]byte(resultStr), queryUnderstanding)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#ConceptRetrievalStrategistAgent] Unmarshal %s err:%v", resultStr, err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return
	}
	queryStrategys = queryUnderstanding.QueryStrategys
	return queryStrategys, nil
}

// MetricDynamicParamsGeneratorAgent Metric 动态参数生成智能体
func (a *agentClient) MetricDynamicParamsGeneratorAgent(
	ctx context.Context,
	req *interfaces.MetricDynamicParamsGeneratorReq,
) (dynamicParams map[string]any, missingParams *interfaces.MissingPropertyParams, err error) {
	// 📤 记录调用 Agent 的入参
	queryStr := utils.ObjectToJSON(req)
	a.logger.WithContext(ctx).Infof("  ├─ [Agent调用] Metric Agent 入参: query=%s", queryStr)

	chatReq := &interfaces.ChatRequest{
		AgentKey:     a.DeployAgent.MetricDynamicParamsGeneratorKey,
		Stream:       false,
		Query:        queryStr,
		AgentVersion: "latest",
	}

	result, err := a.APIChat(ctx, chatReq)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("  ├─ [Agent调用] ❌ APIChat 失败: %v", err)
		return nil, nil, err
	}

	// 提取输出内容
	var text string
	if result != nil && result.Message != nil && result.Message.Content != nil &&
		result.Message.Content.FinalAnswer != nil && result.Message.Content.FinalAnswer.Answer != nil {
		text = result.Message.Content.FinalAnswer.Answer.Text
	}

	// 解析输出内容
	resultStr, err := parseResultFromAgentV1Answer(text)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("  ├─ [Agent解析] ❌ 解析失败: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, nil, err
	}

	// 📥 记录 Agent 原始输出
	a.logger.WithContext(ctx).Debugf("  ├─ [Agent返回] 原始输出: %s", resultStr)

	// 解析 JSON 结果
	var rawResult map[string]any
	err = json.Unmarshal([]byte(resultStr), &rawResult)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("  ├─ [JSON解析] ❌ 失败: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, nil, err
	}

	// 检查是否是缺参错误
	if errorMsg, ok := rawResult["_error"].(string); ok {
		missingParams = parseMetricMissingParamsFromError(req.LogicProperty.Name, errorMsg)
		a.logger.WithContext(ctx).Warnf("  └─ [Agent结果] ⚠️ 缺参: %s", errorMsg)
		return nil, missingParams, nil
	}

	// 成功情况
	a.logger.WithContext(ctx).Debugf("  └─ [Agent结果] ✅ 成功: %+v", rawResult)
	return rawResult, nil, nil
}

// OperatorDynamicParamsGeneratorAgent Operator 动态参数生成智能体
func (a *agentClient) OperatorDynamicParamsGeneratorAgent(
	ctx context.Context,
	req *interfaces.OperatorDynamicParamsGeneratorReq,
) (dynamicParams map[string]any, missingParams *interfaces.MissingPropertyParams, err error) {
	// 📤 记录调用 Agent 的入参
	queryStr := utils.ObjectToJSON(req)
	a.logger.WithContext(ctx).Infof("  ├─ [Agent调用] Operator Agent 入参: property=%s, query=%s",
		req.LogicProperty.Name, req.Query)
	customQuerys := make(map[string]any)
	if req.OperatorID != "" {
		customQuerys["operator_id"] = req.OperatorID
	}
	chatReq := &interfaces.ChatRequest{
		AgentKey:     a.DeployAgent.OperatorDynamicParamsGeneratorKey,
		Stream:       false,
		Query:        queryStr,
		CustomQuerys: customQuerys,
		AgentVersion: "latest",
	}

	result, err := a.APIChat(ctx, chatReq)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("[AgentApp#OperatorDynamicParamsGeneratorAgent] APIChat err:%v", err)
		return nil, nil, err
	}

	// 提取输出内容
	var text string
	if result != nil && result.Message != nil && result.Message.Content != nil &&
		result.Message.Content.FinalAnswer != nil && result.Message.Content.FinalAnswer.Answer != nil {
		text = result.Message.Content.FinalAnswer.Answer.Text
	}

	// 解析输出内容
	resultStr, err := parseResultFromAgentV1Answer(text)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("  ├─ [Agent解析] ❌ 解析失败: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, nil, err
	}

	// 📥 记录 Agent 原始输出
	a.logger.WithContext(ctx).Debugf("  ├─ [Agent返回] 原始输出: %s", resultStr)

	// 解析 JSON 结果
	var rawResult map[string]any
	err = json.Unmarshal([]byte(resultStr), &rawResult)
	if err != nil {
		a.logger.WithContext(ctx).Errorf("  ├─ [JSON解析] ❌ 失败: %v", err)
		err = infraErr.DefaultHTTPError(ctx, http.StatusInternalServerError, err.Error())
		return nil, nil, err
	}

	// 检查是否是缺参错误
	if errorMsg, ok := rawResult["_error"].(string); ok {
		missingParams = parseOperatorMissingParamsFromError(req.LogicProperty.Name, errorMsg)
		a.logger.WithContext(ctx).Warnf("  └─ [Agent结果] ⚠️ 缺参: %s", errorMsg)
		return nil, missingParams, nil
	}

	// 成功情况
	a.logger.WithContext(ctx).Debugf("  └─ [Agent结果] ✅ 成功: %+v", rawResult)
	return rawResult, nil, nil
}

func parseResultFromAgentV1Answer(jsonStr string) (resultStr string, err error) {
	start := strings.Index(jsonStr, "{")
	end := strings.LastIndex(jsonStr, "}")
	if start == -1 || end == -1 {
		err = fmt.Errorf("invalid JSON format")
		return
	}

	jsonStr = jsonStr[start : end+1]

	// If the string contains escape characters, unescape them
	if strings.Contains(jsonStr, "\\n") || strings.Contains(jsonStr, "\\\"") {
		jsonStr = strings.ReplaceAll(jsonStr, "\\n", "\n")
		jsonStr = strings.ReplaceAll(jsonStr, "\\\"", "\"")
	}
	resultStr = jsonStr
	return
}

// parseMetricMissingParamsFromError 解析 metric agent 返回的缺参错误信息（简化版）
// 直接返回 Agent 生成的原始错误消息，不再解析具体参数信息
func parseMetricMissingParamsFromError(propertyName, errorMsg string) *interfaces.MissingPropertyParams {
	if errorMsg == "" {
		return &interfaces.MissingPropertyParams{
			Property: propertyName,
			ErrorMsg: "",
		}
	}

	// 直接返回 Agent 生成的错误消息，不再解析具体参数信息
	return &interfaces.MissingPropertyParams{
		Property: propertyName,
		ErrorMsg: errorMsg,
	}
}

// parseOperatorMissingParamsFromError 解析 operator agent 返回的缺参错误信息（简化版）
// 直接返回 Agent 生成的原始错误消息，不再解析具体参数信息
func parseOperatorMissingParamsFromError(propertyName, errorMsg string) *interfaces.MissingPropertyParams {
	// operator 和 metric 的缺参格式相同，直接复用
	return parseMetricMissingParamsFromError(propertyName, errorMsg)
}
