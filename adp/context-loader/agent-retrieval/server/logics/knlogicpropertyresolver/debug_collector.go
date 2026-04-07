// Copyright The kweaver.ai Authors.
//
// Licensed under the Apache License, Version 2.0.
// See the LICENSE file in the project root for details.

// Package knlogicpropertyresolver provides logic property resolver service for knowledge network.
package knlogicpropertyresolver

import (
	"github.com/kweaver-ai/adp/context-loader/agent-retrieval/server/interfaces"
)

// DebugCollector Debug 信息收集器
type DebugCollector struct {
	propertyTypes  map[string]string
	agentRequests  map[string]interfaces.AgentRequestDebugInfo
	agentResponses map[string]*interfaces.AgentResponseDebugInfo
	dynamicParams  map[string]any
	nowMs          int64
	traceID        string
	warnings       []string
}

// NewDebugCollector 创建 Debug 信息收集器
func NewDebugCollector() *DebugCollector {
	return &DebugCollector{
		propertyTypes:  make(map[string]string),
		agentRequests:  make(map[string]interfaces.AgentRequestDebugInfo),
		agentResponses: make(map[string]*interfaces.AgentResponseDebugInfo),
		dynamicParams:  make(map[string]any),
		warnings:       make([]string, 0),
	}
}

// AddPropertyType 添加属性类型
func (dc *DebugCollector) AddPropertyType(propertyName, propertyType string) {
	dc.propertyTypes[propertyName] = propertyType
}

// RecordMetricAgentRequest 记录 Metric Agent 请求（直接存储 Agent 请求结构）
func (dc *DebugCollector) RecordMetricAgentRequest(
	propertyName string,
	agentReq *interfaces.MetricDynamicParamsGeneratorReq,
) {
	// 直接存储 Agent 请求结构
	dc.agentRequests[propertyName] = agentReq
}

// RecordOperatorAgentRequest 记录 Operator Agent 请求（直接存储 Agent 请求结构）
func (dc *DebugCollector) RecordOperatorAgentRequest(
	propertyName string,
	agentReq *interfaces.OperatorDynamicParamsGeneratorReq,
) {
	// 直接存储 Agent 请求结构
	dc.agentRequests[propertyName] = agentReq
}

// RecordAgentResponseSuccess 记录 Agent 成功响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseSuccess(propertyName string, dynamicParams map[string]any) {
	// 直接存储 Agent 响应：成功时返回 dynamicParams
	dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
		DynamicParams: dynamicParams,
	}

	// 同时收集到 dynamic_params
	dc.dynamicParams[propertyName] = dynamicParams
}

// RecordAgentResponseMissingParams 记录 Agent 缺参响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseMissingParams(
	propertyName string,
	missingParams *interfaces.MissingPropertyParams,
) {
	errorMsg := ""
	if missingParams != nil {
		errorMsg = missingParams.ErrorMsg
	}

	// 直接存储 Agent 响应：失败时返回 _error 字段
	dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
		Error: errorMsg,
	}
}

// RecordAgentResponseError 记录 Agent 错误响应（直接存储 Agent 响应）
func (dc *DebugCollector) RecordAgentResponseError(propertyName, errorMsg string) {
	// 直接存储 Agent 响应：失败时返回 _error 字段
	dc.agentResponses[propertyName] = &interfaces.AgentResponseDebugInfo{
		Error: errorMsg,
	}
}

// SetNowMs 设置当前时间戳
func (dc *DebugCollector) SetNowMs(nowMs int64) {
	dc.nowMs = nowMs
}

// SetTraceID 设置追踪 ID
func (dc *DebugCollector) SetTraceID(traceID string) {
	dc.traceID = traceID
}

// AddWarning 添加警告信息
func (dc *DebugCollector) AddWarning(warning string) {
	dc.warnings = append(dc.warnings, warning)
}

// BuildDebugInfo 构建最终的 Debug 信息
func (dc *DebugCollector) BuildDebugInfo() *interfaces.ResolveDebugInfo {
	// 将 property_types、agent_requests、agent_responses 合并为 agent_info
	agentInfo := make(map[string]*interfaces.AgentInfo)
	for propertyName, propertyType := range dc.propertyTypes {
		var request interfaces.AgentRequestDebugInfo
		if req, exists := dc.agentRequests[propertyName]; exists {
			request = req
		}
		agentInfo[propertyName] = &interfaces.AgentInfo{
			PropertyType: propertyType,
			Request:      request,
			Response:     dc.agentResponses[propertyName],
		}
	}

	return &interfaces.ResolveDebugInfo{
		DynamicParams: dc.dynamicParams,
		AgentInfo:     agentInfo,
		NowMs:         dc.nowMs,
		Warnings:      dc.warnings,
		TraceID:       dc.traceID,
	}
}
