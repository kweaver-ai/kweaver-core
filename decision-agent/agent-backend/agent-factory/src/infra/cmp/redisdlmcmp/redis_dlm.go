package redisdlmcmp

import (
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v8"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/redishelper"
)

type redisDlmCmp struct {
	conf    *RedisDlmCmpConf
	redSync *redsync.Redsync
}

type RedisDlmCmpConf struct {
	Options          []redsync.Option
	WatchDogInterval time.Duration
	// 当value来自于数据库等，在unlock后可以调用该函数来从db中删除此value
	DeleteValueFunc func(value string) error
	Logger          icmp.Logger
	RedisKeyPrefix  string
}

var _ icmp.RedisDlmCmp = &redisDlmCmp{}

func NewRedisDlmCmp(conf *RedisDlmCmpConf) icmp.RedisDlmCmp {
	pool := goredis.NewPool(redishelper.GetRedisClientUniversal())

	redisDlmCmpImpl := &redisDlmCmp{
		redSync: redsync.New(pool),
		conf:    conf,
	}

	return redisDlmCmpImpl
}

func (r *redisDlmCmp) NewMutex(name string) (mutex icmp.RedisDlmMutexCmp) {
	name = r.conf.RedisKeyPrefix + ":dlm:" + name

	// 1. create mutex
	mutex = &redisDlmMutex{
		redSyncMutex:     r.redSync.NewMutex(name, r.conf.Options...),
		watchDogInterval: r.conf.WatchDogInterval,
		deleteValueFunc:  r.conf.DeleteValueFunc,
		logger:           r.conf.Logger,
	}

	return
}
