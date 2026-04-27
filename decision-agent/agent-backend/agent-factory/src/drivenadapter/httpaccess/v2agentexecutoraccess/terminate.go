package v2agentexecutoraccess

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"go.opentelemetry.io/otel/attribute"
)

// Terminate 终止 Agent 执行
func (ae *v2AgentExecutorHttpAcc) Terminate(ctx context.Context, req *v2agentexecutordto.AgentTerminateReq) error {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx, attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentRunID))

	url := fmt.Sprintf("%s/api/agent-executor/v2/agent/terminate", ae.privateAddress)

	headers := make(map[string]string)

	// 使用 streamClient 发起请求并等待完成
	messages, errs, err := ae.streamClient.StreamPost(ctx, url, headers, req)
	if err != nil {
		return err
	}

	// 等待响应完成
	for {
		select {
		case _, ok := <-messages:
			if !ok {
				return nil
			}
		case err, ok := <-errs:
			if ok && err != nil {
				return err
			}

			return nil
		}
	}
}
