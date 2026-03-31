package sessionredisacc

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
)

// GetSession 获取session，返回session是否存在及startTime
// conversationID: 会话ID
// 返回: exists-是否存在, startTime-会话开始时间, err-错误
func (s *sessionRedisAcc) GetSession(ctx context.Context, conversationID string) (exists bool, startTime int64, err error) {
	key := s.getRedisKey(conversationID)
	client := s.redisCmp.GetClient()

	defer func() {
		if err != nil {
			startTime = 0
			exists = false
		}
	}()

	// 1. 获取value（startTime）
	startTime, err = client.Get(ctx, key).Int64()
	if err != nil {
		// redis.Nil 表示key不存在，不是错误
		if errors.Is(err, redis.Nil) {
			err = nil
		}

		return
	}

	exists = true

	return
}

// GetSessionWithTTL 获取session，返回session是否存在、startTime及剩余TTL
// conversationID: 会话ID
// 返回: exists-是否存在, startTime-会话开始时间, ttl-剩余过期时间(秒), err-错误
func (s *sessionRedisAcc) GetSessionWithTTL(ctx context.Context, conversationID string) (exists bool, startTime int64, ttl int, err error) {
	key := s.getRedisKey(conversationID)
	client := s.redisCmp.GetClient()

	defer func() {
		if err != nil {
			startTime = 0
			ttl = 0
			exists = false
		}
	}()

	// 1. 获取value（startTime）
	startTime, err = client.Get(ctx, key).Int64()
	if err != nil {
		// redis.Nil 表示key不存在，不是错误
		if errors.Is(err, redis.Nil) {
			err = nil
		}

		return
	}

	exists = true

	// 2. 获取剩余TTL
	ttlDuration, err := client.TTL(ctx, key).Result()
	if err != nil {
		return
	}

	// TTL返回的是time.Duration，需要转换为秒
	// 如果TTL为-1表示没有设置过期时间，-2表示key不存在
	if ttlDuration > 0 {
		ttl = int(ttlDuration.Seconds())
	}

	return
}

// GetSessionTTL 仅获取session的剩余TTL
// conversationID: 会话ID
// 返回: ttl-剩余过期时间(秒), err-错误（如果key不存在返回0和nil）
func (s *sessionRedisAcc) GetSessionTTL(ctx context.Context, conversationID string) (ttl int, err error) {
	key := s.getRedisKey(conversationID)
	client := s.redisCmp.GetClient()

	// 获取剩余TTL
	ttlDuration, err := client.TTL(ctx, key).Result()
	if err != nil {
		return
	}

	// TTL返回的是time.Duration，需要转换为秒
	// 如果TTL为-1表示没有设置过期时间，-2表示key不存在
	if ttlDuration > 0 {
		ttl = int(ttlDuration.Seconds())
	}

	return
}
