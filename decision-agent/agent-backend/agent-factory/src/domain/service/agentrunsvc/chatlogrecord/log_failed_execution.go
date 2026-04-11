package chatlogrecord

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/req"
	agentresp "github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/driveradapter/api/rdto/agent/resp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/otellog"
	"go.opentelemetry.io/otel/log"
)

func LogFailedExecution(ctx context.Context, req *agentreq.ChatReq, execErr error, resp *agentresp.ChatResp) {
	toolCallCount := 0
	toolCallFailedCount := 0

	if req.ConversationSessionID == "" {
		timestamp := cutil.GetCurrentMSTimestamp()
		req.ConversationSessionID = fmt.Sprintf("%s-%s", req.ConversationID, strconv.FormatInt(timestamp, 10))
	}

	var progressAttr log.KeyValue

	var totalTimeAttr log.KeyValue

	var totalTokensAttr log.KeyValue

	if resp == nil {
		totalTimeAttr = log.Float64("total_time", 0)
		totalTokensAttr = log.Int64("total_tokens", 0)
		progressAttr = log.String("progress", "[]")
		toolCallCountAttr := log.Int("tool_call_count", 0)
		toolCallFailedCountAttr := log.Int("tool_call_failed_count", 0)

		otellog.LogInfo(ctx, "After process failed",
			totalTimeAttr,
			totalTokensAttr,
			progressAttr,
			toolCallCountAttr,
			toolCallFailedCountAttr,
			log.String("agent_id", req.AgentID),
			log.String("agent_version", req.AgentVersion),
			log.String("user_id", req.UserID),
			log.String("conversation_id", req.ConversationID),
			log.String("session_id", req.ConversationSessionID),
			log.String("call_type", string(req.CallType)),
			log.String("run_id", req.AgentRunID),
			log.Float64("ttft", float64(req.TTFT)),
			log.String("status", "failed"),
			log.String("input_message", req.Query),
			log.Int64("start_time", req.ReqStartTime),
			log.Int64("end_time", cutil.GetCurrentMSTimestamp()),
		)

		return
	}

	var totaltime float64

	var totalTokens int64

	if assistantContent, ok := resp.Message.Content.(map[string]interface{}); ok {
		if middleAnswerVal, ok := assistantContent["middle_answer"]; ok {
			if middleAnswer, ok := middleAnswerVal.(map[string]interface{}); ok {
				if val, ok := middleAnswer["progress"]; ok {
					progressJsonStr, _ := json.Marshal(val)
					progressAttr = log.String("progress", string(progressJsonStr))

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
					progressJsonStr, _ := json.Marshal([]agentrespvo.Progress{})
					progressAttr = log.String("progress", string(progressJsonStr))
				}
			} else {
				progressJsonStr, _ := json.Marshal([]agentrespvo.Progress{})
				progressAttr = log.String("progress", string(progressJsonStr))
			}
		} else {
			progressJsonStr, _ := json.Marshal([]agentrespvo.Progress{})
			progressAttr = log.String("progress", string(progressJsonStr))
		}
	} else {
		progressJsonStr, _ := json.Marshal([]agentrespvo.Progress{})
		progressAttr = log.String("progress", string(progressJsonStr))
	}

	if resp.Message.Ext != nil {
		totaltime = resp.Message.Ext.TotalTime
		totalTokens = resp.Message.Ext.TotalTokens
	}

	totalTimeAttr = log.Float64("total_time", totaltime*1000)
	totalTokensAttr = log.Int64("total_tokens", totalTokens)

	otellog.LogInfo(ctx, "After process failed",
		totalTimeAttr,
		totalTokensAttr,
		progressAttr,
		log.Int("tool_call_count", toolCallCount),
		log.Int("tool_call_failed_count", toolCallFailedCount),
		log.String("agent_id", req.AgentID),
		log.String("agent_version", req.AgentVersion),
		log.String("user_id", req.UserID),
		log.String("conversation_id", req.ConversationID),
		log.String("session_id", req.ConversationSessionID),
		log.String("call_type", string(req.CallType)),
		log.String("run_id", req.AgentRunID),
		log.Float64("ttft", float64(req.TTFT)),
		log.String("status", "failed"),
		log.String("input_message", req.Query),
		log.Int64("start_time", req.ReqStartTime),
		log.Int64("end_time", cutil.GetCurrentMSTimestamp()),
	)
}
