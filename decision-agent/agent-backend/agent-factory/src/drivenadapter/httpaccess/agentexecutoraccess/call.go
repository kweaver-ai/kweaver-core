package agentexecutoraccess

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutordto"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/otel/oteltrace"
	"go.opentelemetry.io/otel/attribute"
)

func (ae *agentExecutorHttpAcc) Call(ctx context.Context, req *agentexecutordto.AgentCallReq) (chan string, chan error, error) {
	var err error

	ctx, span := oteltrace.StartInternalSpan(ctx)
	defer span.End()
	span.SetAttributes(attribute.String("agent_call_req", fmt.Sprintf("%+v", req)))
	span.SetAttributes(attribute.String("user_id", req.UserID))
	span.SetAttributes(attribute.String("agent_run_id", req.Config.SessionID))
	span.SetAttributes(attribute.String("agent_id", req.Config.AgentID))

	var url string
	if req.CallType == constant.DebugChat {
		url = fmt.Sprintf("%s/api/agent-executor/v1/agent/debug", ae.privateAddress)
	} else {
		url = fmt.Sprintf("%s/api/agent-executor/v1/agent/run", ae.privateAddress)
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

	messages, errs, err := ae.streamClient.StreamPost(ctx, url, headers, req)

	return messages, errs, err
}
