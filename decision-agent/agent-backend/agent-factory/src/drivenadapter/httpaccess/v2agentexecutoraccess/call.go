package v2agentexecutoraccess

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant/otelconst"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/v2agentexecutoraccess/v2agentexecutordto"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/otel/oteltrace"
)

// mapCarrier adapts map[string]string to propagation.TextMapCarrier.
type mapCarrier map[string]string

func (c mapCarrier) Get(key string) string { return c[key] }
func (c mapCarrier) Set(key, value string) { c[key] = value }
func (c mapCarrier) Keys() []string {
	keys := make([]string, 0, len(c))
	for k := range c {
		keys = append(keys, k)
	}

	return keys
}

func (ae *v2AgentExecutorHttpAcc) Call(ctx context.Context, req *v2agentexecutordto.V2AgentCallReq) (chan string, chan error, error) {
	var err error

	ctx, _ = oteltrace.StartInternalSpan(ctx)
	defer oteltrace.EndSpan(ctx, err)
	oteltrace.SetAttributes(ctx,
		attribute.String("agent_call_req", fmt.Sprintf("%+v", req)),
		attribute.String(otelconst.AttrUserID, req.UserID),
		attribute.String(otelconst.AttrGenAIAgentRunID, req.AgentOptions.AgentRunID),
		attribute.String(otelconst.AttrGenAIAgentID, req.AgentID),
	)
	oteltrace.SetConversationID(ctx, req.AgentOptions.ConversationID)

	var url string
	if req.CallType == constant.DebugChat {
		url = fmt.Sprintf("%s/api/agent-executor/v2/agent/debug", ae.privateAddress)
	} else {
		url = fmt.Sprintf("%s/api/agent-executor/v2/agent/run", ae.privateAddress)
	}

	headers := make(map[string]string)

	// NOTE: 内部接口传递x-account-id，值为userID
	// NOTE: 内部接口传递x-account-type给executor，如果是应用账号使用的是app而不是business
	// NOTE: x-account-type 枚举值是 app 和 user 和 anonymous
	chelper.SetAccountInfoToHeaderMap(headers, req.XAccountID, req.XAccountType)
	// headers["x-account-id"] = req.UserID
	// if req.VisitorType == constant.Business {
	// 	headers["x-account-type"] = "app"
	// } else if req.VisitorType == constant.RealName {
	// 	headers["x-account-type"] = "user"
	// } else {
	// 	headers["x-account-type"] = "anonymous"
	// }
	if req.Token != "" {
		headers["token"] = req.Token
		headers["Authorization"] = "Bearer " + req.Token
	}

	headers["x-business-domain"] = req.XBusinessDomainID

	// 注入 OTel trace context（traceparent/tracestate）到请求头，实现跨服务 trace 关联
	otel.GetTextMapPropagator().Inject(ctx, mapCarrier(headers))

	messages, errs, err := ae.streamClient.StreamPost(ctx, url, headers, req)

	return messages, errs, err
}
