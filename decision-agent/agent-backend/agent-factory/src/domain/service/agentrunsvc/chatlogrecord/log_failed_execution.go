package chatlogrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kweaver-ai/TelemetrySDK-Go/span/v2/field"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	o11y "github.com/kweaver-ai/kweaver-go-lib/observability"
)

// NOTE: 失败日志上报
func LogFailedExecution(ctx context.Context, req *agentreq.ChatReq, err error, resp *agentresp.ChatResp) {
	options := []field.LogOptionFunc{}

	toolCallCount := 0

	toolCallFailedCount := 0

	if req.ConversationSessionID == "" {
		timestamp := cutil.GetCurrentMSTimestamp()
		req.ConversationSessionID = fmt.Sprintf("%s-%s", req.ConversationID, strconv.FormatInt(timestamp, 10))
	}
	// // 如果传入的resp是nil或者是空，则赋为零值
	if resp == nil {
		totalTimeAttr := field.NewAttribute("total_time", field.MallocJsonField(0))
		totalTokensAttr := field.NewAttribute("total_tokens", field.MallocJsonField(0))
		// 空json字符串
		progressAttr := field.NewAttribute("progress", field.MallocJsonField("[]"))
		toolCallCountAttr := field.NewAttribute("tool_call_count", field.MallocJsonField(0))
		toolCallFailedCountAttr := field.NewAttribute("tool_call_failed_count", field.MallocJsonField(0))
		options = append(options, field.WithAttribute(progressAttr))
		options = append(options, field.WithAttribute(totalTokensAttr))
		options = append(options, field.WithAttribute(totalTimeAttr))
		options = append(options, field.WithAttribute(toolCallCountAttr))
		options = append(options, field.WithAttribute(toolCallFailedCountAttr))
	} else {
		// NNOTE:获取progress 避免panic
		var totaltime float64

		var totalTokens int64

		if assistantContent, ok := resp.Message.Content.(map[string]interface{}); ok {
			// NOTE: 获取middler answer 的progress
			// NOTE: 1.判断是否存在middle_answer
			if middleAnswerVal, ok := assistantContent["middle_answer"]; ok {
				// NOTE: 2.断言middle_answer
				if middleAnswer, ok := middleAnswerVal.(map[string]interface{}); ok {
					// NOTE: 3.判断middle_answer中是否存在progress
					if val, ok := middleAnswer["progress"]; ok {
						progressJsonStr, _ := json.Marshal(val)
						options = append(options, field.WithAttribute(field.NewAttribute("progress", field.MallocJsonField(string(progressJsonStr)))))

						progresses, ok := val.([]interface{})
						if ok {
							for _, val := range progresses {
								progress, ok := val.(map[string]interface{})
								if !ok {
									continue
								}

								if stage, ok := progress["stage"].(string); ok {
									if stage == "skill" {
										toolCallCount++

										if status, ok := progress["status"].(string); ok && status == "failed" {
											toolCallFailedCount++
										}
									}
								}
							}
						}
					} else {
						options = append(options, field.WithAttribute(field.NewAttribute("progress", field.MallocJsonField([]agentrespvo.Progress{}))))
					}
				} else {
					options = append(options, field.WithAttribute(field.NewAttribute("progress", field.MallocJsonField([]agentrespvo.Progress{}))))
				}
			} else {
				options = append(options, field.WithAttribute(field.NewAttribute("progress", field.MallocJsonField([]agentrespvo.Progress{}))))
			}
		} else {
			options = append(options, field.WithAttribute(field.NewAttribute("progress", field.MallocJsonField([]agentrespvo.Progress{}))))
		}

		if resp.Message.Ext != nil {
			totaltime = resp.Message.Ext.TotalTime
			totalTokens = resp.Message.Ext.TotalTokens
		}
		// total time单位是s，存进去变成毫秒
		totalTimeAttr := field.NewAttribute("total_time", field.MallocJsonField(totaltime*1000))
		totalTokensAttr := field.NewAttribute("total_tokens", field.MallocJsonField(totalTokens))
		options = append(options, field.WithAttribute(totalTokensAttr))
		options = append(options, field.WithAttribute(totalTimeAttr))
	}

	agentIdAttr := field.NewAttribute("agent_id", field.MallocJsonField(req.AgentID))
	agentVersionAttr := field.NewAttribute("agent_version", field.MallocJsonField(req.AgentVersion))
	userIDAttr := field.NewAttribute("user_id", field.MallocJsonField(req.UserID))
	conversationIDAttr := field.NewAttribute("conversation_id", field.MallocJsonField(req.ConversationID))
	sessionIDAttr := field.NewAttribute("session_id", field.MallocJsonField(req.ConversationSessionID))
	callTypeAttr := field.NewAttribute("call_type", field.MallocJsonField(req.CallType))
	runIDAttr := field.NewAttribute("run_id", field.MallocJsonField(req.AgentRunID))
	ttftAttr := field.NewAttribute("ttft", field.MallocJsonField(req.TTFT))
	executeStatus := field.NewAttribute("status", field.MallocJsonField("failed"))

	inputMessage := field.NewAttribute("input_message", field.MallocJsonField(req.Query))
	startTime := field.NewAttribute("start_time", field.MallocJsonField(req.ReqStartTime))
	endTime := field.NewAttribute("end_time", field.MallocJsonField(cutil.GetCurrentMSTimestamp()))

	toolCallCountAttr := field.NewAttribute("tool_call_count", field.MallocJsonField(toolCallCount))
	toolCallFailedCountAttr := field.NewAttribute("tool_call_failed_count", field.MallocJsonField(toolCallFailedCount))
	options = append(options, field.WithAttribute(toolCallCountAttr))
	options = append(options, field.WithAttribute(toolCallFailedCountAttr))

	options = append(options, field.WithAttribute(agentIdAttr))
	options = append(options, field.WithAttribute(agentVersionAttr))
	options = append(options, field.WithAttribute(userIDAttr))
	options = append(options, field.WithAttribute(conversationIDAttr))
	options = append(options, field.WithAttribute(sessionIDAttr))
	options = append(options, field.WithAttribute(callTypeAttr))
	options = append(options, field.WithAttribute(runIDAttr))
	options = append(options, field.WithAttribute(ttftAttr))
	options = append(options, field.WithAttribute(executeStatus))
	options = append(options, field.WithAttribute(inputMessage))
	options = append(options, field.WithAttribute(startTime))
	options = append(options, field.WithAttribute(endTime))

	o11y.InfoWithAttr(ctx, "After process failed", options...)
}
