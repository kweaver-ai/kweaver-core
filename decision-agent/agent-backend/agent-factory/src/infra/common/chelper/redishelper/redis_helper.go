package redishelper

import (
	"context"
	"errors"
	"log"
	"time"

	//"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	redis "github.com/go-redis/redis/v8"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cconstant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/cutil"

	"golang.org/x/sync/singleflight"
)

var ErrNotSupportInLocalEnv = errors.New("redishelper: not support in local env")

func SetStruct(rdb redis.Cmdable, key string, value interface{}, ttl time.Duration) (err error) {
	//if cenvhelper.IsLocalDev() {
	//	return nil
	//}
	// 创建一个带有超时的上下文
	redisCtx, cancel := context.WithTimeout(context.Background(), cconstant.RedisOpTimeout)
	defer cancel()

	jsonStr, err := cutil.JSON().MarshalToString(value)
	if err != nil {
		return
	}

	err = rdb.Set(redisCtx, key, jsonStr, ttl).Err()

	return
}

var sfgGetStruct singleflight.Group

func GetStruct(rdb redis.Cmdable, key string, value interface{}) (err error) {
	// 可根据需要打开或关闭
	//if cenvhelper.IsLocalDev() {
	//	return ErrNotSupportInLocalEnv
	//}
	// 创建一个带有超时的上下文
	redisCtx, cancel := context.WithTimeout(context.Background(), cconstant.RedisOpTimeout)
	defer cancel()

	jsonInter, err, shared := sfgGetStruct.Do(key, func() (interface{}, error) {
		return rdb.Get(redisCtx, key).Result()
	})

	if err != nil {
		return
	}

	if shared && cenvhelper.IsDebugMode() {
		log.Println("shared")
	}

	//nolint:forcetypeassert
	err = cutil.JSON().UnmarshalFromString(jsonInter.(string), value)

	return
}

func GetRedisClientUniversal() (uc redis.UniversalClient) {
	uc = RedisClient()
	if uc == nil {
		panic("redis client is not a universal client")
	}

	return
}
