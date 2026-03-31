package sessionredisacc

import (
	"context"
	"time"
)

// SetSession 设置session，返回是否是新创建的session以及startTime
// conversationID: 会话ID
// startTime: 会话开始时间（作为value存储）
// ttl: 过期时间（秒）
func (s *sessionRedisAcc) SetSession(ctx context.Context, conversationID string, startTime int64, ttl int) (isNew bool, err error) {
	key := s.getRedisKey(conversationID)
	client := s.redisCmp.GetClient()

	defer func() {
		if err != nil {
			isNew = false
		}
	}()

	// 1. 使用SetNX来判断是否是新创建的
	isNew, err = client.SetNX(ctx, key, startTime, time.Duration(ttl)*time.Second).Result()
	if err != nil {
		return
	}

	// 2. 如果key已存在，更新过期时间
	if !isNew {
		err = client.Expire(ctx, key, time.Duration(ttl)*time.Second).Err()
		if err != nil {
			return
		}
	}

	return
}
