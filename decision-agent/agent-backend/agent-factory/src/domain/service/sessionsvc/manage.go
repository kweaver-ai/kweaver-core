package sessionsvc

import (
	"context"
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/httpaccess/agentexecutoraccess/agentexecutoraccreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/session/sessionreq"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/session/sessionresp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/ctype"
)

func (s *sessionSvc) Manage(ctx context.Context, req sessionreq.ManageReq, visitorInfo *ctype.VisitorInfo) (resp sessionresp.ManageResp, err error) {
	var startTime int64

	var ttl int

	switch req.Action {
	case sessionreq.SessionManageActionGetInfoOrCreate:
		startTime, ttl, err = s.HandleGetInfoOrCreate(ctx, req, visitorInfo, true)
		if err != nil {
			return
		}

	case sessionreq.SessionManageActionRecoverLifetimeOrCreate:
		startTime, ttl, err = s.HandleRecoverLifetimeOrCreate(ctx, req, visitorInfo, true)
		if err != nil {
			return
		}

	default:
		err = fmt.Errorf("unsupported action: %s", req.Action)
		return
	}

	// 生成session id: {conversation_id}-{start_time}
	resp = sessionresp.ManageResp{
		ConversationSessionID: fmt.Sprintf("%s-%d", req.ConversationID, startTime),
		TTL:                   ttl,
		StartTime:             startTime,
	}

	return
}

// triggerAgentCacheUpsert 触发agent缓存的创建或更新
func (s *sessionSvc) triggerAgentCacheUpsert(ctx context.Context, req sessionreq.ManageReq, visitorInfo *ctype.VisitorInfo) error {
	cacheReq := &agentexecutoraccreq.AgentCacheManageReq{
		AgentID:      req.AgentID,
		AgentVersion: req.AgentVersion,
		Action:       agentexecutoraccreq.AgentCacheActionUpsert,
	}

	_, err := s.agentExecutorV1.AgentCacheManage(ctx, cacheReq, visitorInfo)

	return err
}
