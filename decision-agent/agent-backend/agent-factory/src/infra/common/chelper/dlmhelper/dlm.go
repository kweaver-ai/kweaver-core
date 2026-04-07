package dlmhelper

import (
	"context"
	"time"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/port/driven/idbaccess"
	"github.com/kweaver-ai/kweaver-go-lib/logger"

	dbaulid "github.com/kweaver-ai/decision-agent/agent-factory/src/drivenadapter/dbaccess/ulid"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/redisdlmcmp"

	"github.com/go-redsync/redsync/v4"
)

var uniqueIDRepo idbaccess.UlidRepo

// genRedisDlmUniqueValue generates a unique value for RedisDlm
// 【注意】使用数据库来保障唯一性，此处使用场景不用过于考虑性能问题
func genRedisDlmUniqueValue() (value string, err error) {
	if uniqueIDRepo == nil {
		uniqueIDRepo = dbaulid.NewUlidRepo()
	}

	ctx := context.Background()
	value, err = uniqueIDRepo.GenUniqID(ctx, cconstant.UniqueIDFlagRedisDlm)

	return
}

// delRedisDlmUniqueValue delete the value from db after RedisDlm unlock
func delRedisDlmUniqueValue(value string) (err error) {
	if uniqueIDRepo == nil {
		uniqueIDRepo = dbaulid.NewUlidRepo()
	}

	ctx := context.Background()
	err = uniqueIDRepo.DelUniqID(ctx, cconstant.UniqueIDFlagRedisDlm, value)

	return
}

// GetDefaultDlmConf 获取默认的RedisDlmCmpConf配置
// 【注意】：暂未经过考验，慎用
func GetDefaultDlmConf(redisKeyPrefix string) (dlmConf *redisdlmcmp.RedisDlmCmpConf) {
	expiry := 20 * time.Second
	delay := 20 * time.Millisecond
	maxTries := 500

	opts := []redsync.Option{
		redsync.WithExpiry(expiry),
		redsync.WithRetryDelay(delay),
		redsync.WithTries(maxTries), //【注意】：重试次数用完后，会返回错误（加锁失败）
		redsync.WithGenValueFunc(genRedisDlmUniqueValue),
	}

	dlmConf = &redisdlmcmp.RedisDlmCmpConf{
		Options:          opts,
		WatchDogInterval: expiry / 2,
		DeleteValueFunc:  delRedisDlmUniqueValue,
		RedisKeyPrefix:   redisKeyPrefix,
		Logger:           logger.GetLogger(),
	}

	return
}
