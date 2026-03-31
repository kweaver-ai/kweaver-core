package sessionredisacc

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
)

// RefreshSession 刷新session的过期时间，返回session是否存在及startTime
// conversationID: 会话ID
// ttl: 过期时间（秒）
// 返回: exists-是否存在, startTime-会话开始时间, err-错误
func (s *sessionRedisAcc) RefreshSession(ctx context.Context, conversationID string, ttl int) (exists bool, startTime int64, err error) {
	key := s.getRedisKey(conversationID)
	client := s.redisCmp.GetClient()

	defer func() {
		if err != nil {
			startTime = 0
			exists = false
		}
	}()

	// 1. 先获取value（startTime）
	startTime, err = client.Get(ctx, key).Int64()
	if err != nil {
		// redis.Nil 表示key不存在，不是错误
		if errors.Is(err, redis.Nil) {
			err = nil
		}

		return
	}

	// 2. 刷新过期时间
	err = client.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
	if err != nil {
		return
	}

	exists = true

	return
}
