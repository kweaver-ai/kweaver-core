package chatlogrecord

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/otellog"
	"go.opentelemetry.io/otel/log"
)

func LogSuccessExecution(ctx context.Context, req *agentreq.ChatReq, progressAns []*agentrespvo.Progress, totalTime float64, totalTokens int64) {
	progressJsonStr, _ := json.Marshal(progressAns)

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

	otellog.LogInfo(ctx, "After process success",
		log.String("agent_id", req.AgentID),
		log.String("agent_version", req.AgentVersion),
		log.String("user_id", req.UserID),
		log.String("conversation_id", req.ConversationID),
		log.String("session_id", req.ConversationSessionID),
		log.String("call_type", string(req.CallType)),
		log.String("progress", string(progressJsonStr)),
		log.String("run_id", req.AgentRunID),
		log.Float64("ttft", float64(req.TTFT)),
		log.Float64("total_time", totalTime*1000),
		log.Int64("total_tokens", totalTokens),
		log.String("status", "success"),
		log.String("input_message", req.Query),
		log.Int64("start_time", req.ReqStartTime),
		log.Int64("end_time", cutil.GetCurrentMSTimestamp()),
		log.Int("tool_call_count", toolCallCount),
		log.Int("tool_call_failed_count", toolCallFailedCount),
	)
}