package v2agentexecutoraccess

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"go.opentelemetry.io/otel/attribute"
)

// Resume 恢复 Agent 执行（中断后恢复）
func (ae *v2AgentExecutorHttpAcc) Resume(ctx context.Context, req *v2agentexecutordto.AgentResumeReq) (chan string, chan error, error) {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID),
		attribute.String("action", req.ResumeInfo.Action),
	)

	url := fmt.Sprintf("%s/api/agent-executor/v2/agent/resume", ae.privateAddress)

	headers := make(map[string]string)

	messages, errs, err := ae.streamClient.StreamPost(ctx, url, headers, req)

	return messages, errs, err
}
