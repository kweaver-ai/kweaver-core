package v3agentconfigsvc

import (
	"context"
	"encoding/json"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/ctopicenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/mqvo"
	"github.com/pkg/errors"
)

func (s *dataAgentConfigSvc) handleUpdateNameMq(agentID, agentName string) (err error) {
	msg := mqvo.NewUpdateAgentNameMqMsg(agentID, agentName)

	msgBys, err := json.Marshal(msg)
	if err != nil {
		err = errors.Wrapf(err, "marshal msg failed")
		return
	}

	ctx := context.Background()

	err = s.mqAccess.Publish(ctx, ctopicenum.AgentNameModifyForAuthorizationPlatform, msgBys)
	if err != nil {
		err = errors.Wrapf(err, "publish msg failed")
		return
	}

	return
}
