package rediscmp

import (
	"sync"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/redishelper"

	"github.com/go-redis/redis/v8"
)

var (
	redisCmpOnce sync.Once
	redisCmpImpl icmp.RedisCmp
)

type redisCmp struct{}

var _ icmp.RedisCmp = &redisCmp{}

func NewRedisCmp() icmp.RedisCmp {
	redisCmpOnce.Do(func() {
		redisCmpImpl = &redisCmp{}
	})

	return redisCmpImpl
}

func (r *redisCmp) GetClient() (client redis.UniversalClient) {
	return redishelper.GetRedisClientUniversal()
}
