package imqaccess

import (
	"context"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/ctopicenum"
)

//go:generate mockgen -source=./mq.go -destination ../ihttpaccessmock/mq.go -package imqaccessmock
type IMqAccess interface {
	Publish(ctx context.Context, topic ctopicenum.MqTopic, msg []byte) (err error)
	Subscribe(ctx context.Context, topic ctopicenum.MqTopic, fun func([]byte) error) (err error)
}
