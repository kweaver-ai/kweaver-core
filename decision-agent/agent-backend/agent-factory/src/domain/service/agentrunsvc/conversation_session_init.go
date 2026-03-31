package agentsvc

// import (
// 	"context"

// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
// 	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
// 	agentresp "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/resp"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cenum"
// 	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
// 	"github.com/pkg/errors"
// )

// // ConversationSessionInit
// func (agentSvc *agentSvc) ConversationSessionInit(ctx context.Context, req *agentreq.ConversationSessionInitReq) (resp *agentresp.ConversationSessionInitResp, err error) {

// 	// 1. 获取Agent
// 	agent, err := agentSvc.agentFactory.GetAgent(ctx, req.AgentID, req.AgentVersion)
// 	if err != nil {
// 		err=errors.Wrapf(err, "get agent failed")
// 		return
// 	}

// 	// 2. 构建请求
// 	initReq := &agentexecutoraccreq.ConversationSessionInitReq{
// 		ConversationID:    req.ConversationID,
// 		AgentID:           req.AgentID,
// 		AgentVersion:      req.AgentVersion,
// 		AgentConfig:       agent.Config,
// 	}

// 	visitorInfo:=&ctype.VisitorInfo{
// 		XAccountID:        req.XAccountID,
// 		XAccountType:      req.XAccountType,
// 		XBusinessDomainID: cenum.BizDomainID(req.XBusinessDomainID),
// 	}

// 	// 3. 发起请求
// 	rt, err := agentSvc.agentExecutorV1.ConversationSessionInit(ctx, initReq,visitorInfo)
// 	if err != nil {
// 		err=errors.Wrapf(err, "conversation session init failed")
// 		return
// 	}

// 	// 4. 返回结果
// 	resp=&agentresp.ConversationSessionInitResp{
// 		ConversationSessionID: rt.ConversationSessionID,
// 		TTL:                   rt.TTL,
// 	}

// 	return
// }
