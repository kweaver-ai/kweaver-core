package chatlogrecord

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// NOTE: 成功日志上报
func LogSuccessExecution(ctx context.Context, req *agentreq.ChatReq, progressAns []*agentrespvo.Progress, totalTime float64, totalTokens int64) {
	// 1. 构造属性
	agentIdAttr := field.NewAttribute("agent_id", field.MallocJsonField(req.AgentID))
	agentVersionAttr := field.NewAttribute("agent_version", field.MallocJsonField(req.AgentVersion))
	userIDAttr := field.NewAttribute("user_id", field.MallocJsonField(req.UserID))
	conversationIDAttr := field.NewAttribute("conversation_id", field.MallocJsonField(req.ConversationID))
	sessionIDAttr := field.NewAttribute("session_id", field.MallocJsonField(req.ConversationSessionID))
	callTypeAttr := field.NewAttribute("call_type", field.MallocJsonField(req.CallType))

	// NOTE 转为json字符串
	progressJsonStr, _ := json.Marshal(progressAns)
	progressAttr := field.NewAttribute("progress", field.MallocJsonField(string(progressJsonStr)))

	runIDAttr := field.NewAttribute("run_id", field.MallocJsonField(req.AgentRunID))
	ttftAttr := field.NewAttribute("ttft", field.MallocJsonField(req.TTFT))

	totalTimeAttr := field.NewAttribute("total_time", field.MallocJsonField(totalTime*1000))
	totalTokensAttr := field.NewAttribute("total_tokens", field.MallocJsonField(totalTokens))

	executeStatus := field.NewAttribute("status", field.MallocJsonField("success"))

	inputMessage := field.NewAttribute("input_message", field.MallocJsonField(req.Query))

	startTime := field.NewAttribute("start_time", field.MallocJsonField(req.ReqStartTime))
	endTime := field.NewAttribute("end_time", field.MallocJsonField(cutil.GetCurrentMSTimestamp()))

	// 2. 构造日志选项
	options := []field.LogOptionFunc{}
	options = append(options, field.WithAttribute(agentIdAttr))
	options = append(options, field.WithAttribute(agentVersionAttr))
	options = append(options, field.WithAttribute(userIDAttr))
	options = append(options, field.WithAttribute(conversationIDAttr))
	options = append(options, field.WithAttribute(sessionIDAttr))
	options = append(options, field.WithAttribute(callTypeAttr))
	options = append(options, field.WithAttribute(progressAttr))
	options = append(options, field.WithAttribute(runIDAttr))
	options = append(options, field.WithAttribute(ttftAttr))
	options = append(options, field.WithAttribute(totalTokensAttr))
	options = append(options, field.WithAttribute(totalTimeAttr))
	options = append(options, field.WithAttribute(executeStatus))
	options = append(options, field.WithAttribute(inputMessage))
	options = append(options, field.WithAttribute(startTime))
	options = append(options, field.WithAttribute(endTime))

	// 2.1 构造工具调用次数和失败次数
	toolCallCount := 0
	toolCallFailedCount := 0

	for _, progress := range progressAns {
		if progress.Stage == "skill" {
			toolCallCount++

			if progress.Status == "failed" {
				toolCallFailedCount++
			}
		}
	}

	toolCallCountAttr := field.NewAttribute("tool_call_count", field.MallocJsonField(toolCallCount))
	toolCallFailedCountAttr := field.NewAttribute("tool_call_failed_count", field.MallocJsonField(toolCallFailedCount))
	options = append(options, field.WithAttribute(toolCallCountAttr))
	options = append(options, field.WithAttribute(toolCallFailedCountAttr))

	// 3. 记录日志
	o11y.InfoWithAttr(ctx, "After process success", options...)
}
