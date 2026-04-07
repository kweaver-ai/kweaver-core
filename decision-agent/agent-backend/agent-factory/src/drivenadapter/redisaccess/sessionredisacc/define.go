package sessionredisacc

import (
	"fmt"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/rediscmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/iredisaccess/isessionredis"
)

const (
	// SessionTTL session过期时间（秒）
	SessionTTL = 600
	// SessionRedisKeyPrefix redis key前缀
	SessionRedisKeyPrefix = "agent-app:conversation-session:"
)

type sessionRedisAcc struct {
	redisCmp icmp.RedisCmp
}

var _ isessionredis.ISessionRedisAcc = &sessionRedisAcc{}

func NewSessionRedisAcc() isessionredis.ISessionRedisAcc {
	return &sessionRedisAcc{
		redisCmp: rediscmp.NewRedisCmp(),
	}
}

func (s *sessionRedisAcc) getRedisKey(conversationID string) string {
	return fmt.Sprintf("%s%s", SessionRedisKeyPrefix, conversationID)
}
