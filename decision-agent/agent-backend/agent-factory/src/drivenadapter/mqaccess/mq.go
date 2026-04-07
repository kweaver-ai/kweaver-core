package mqaccess

import (
	"context"
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/ctopicenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/mqcmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cglobal"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/imqaccess"
)

var (
	mqOnce sync.Once
	mqImpl imqaccess.IMqAccess
)

type mqCmp struct {
	mqClient icmp.IMQClient
}

var _ imqaccess.IMqAccess = &mqCmp{}

func NewMqAccess() imqaccess.IMqAccess {
	mqOnce.Do(func() {
		mqCfgPath := cglobal.GConfig.MqCfgPath

		mqImpl = &mqCmp{
			mqClient: mqcmp.NewMQClientWithPath(mqCfgPath),
		}
	})

	return mqImpl
}

func (p *mqCmp) Publish(ctx context.Context, topic ctopicenum.MqTopic, msg []byte) (err error) {
	err = p.mqClient.Publish(string(topic), msg)
	return
}

func (p *mqCmp) Subscribe(ctx context.Context, topic ctopicenum.MqTopic, fun func([]byte) error) (err error) {
	err = p.mqClient.Subscribe(string(topic), "data_agent", fun)

	return
}
