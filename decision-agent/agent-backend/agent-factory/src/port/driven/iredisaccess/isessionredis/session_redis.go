package isessionredis

import (
	"context"
)

//go:generate mockgen -source=./session_redis.go -destination ./isessionredismock/session_redis.go -package isessionredismock
type ISessionRedisAcc interface {
	// SetSession 设置session，返回是否是新创建的session
	// conversationID: 会话ID
	// startTime: 会话开始时间（作为value存储）
	// ttl: 过期时间（秒）
	SetSession(ctx context.Context, conversationID string, startTime int64, ttl int) (isNew bool, err error)
	// GetSession 获取session，返回session是否存在及startTime
	// conversationID: 会话ID
	// 返回: exists-是否存在, startTime-会话开始时间, err-错误
	GetSession(ctx context.Context, conversationID string) (exists bool, startTime int64, err error)
	// GetSessionWithTTL 获取session，返回session是否存在、startTime及剩余TTL
	// conversationID: 会话ID
	// 返回: exists-是否存在, startTime-会话开始时间, ttl-剩余过期时间(秒), err-错误
	GetSessionWithTTL(ctx context.Context, conversationID string) (exists bool, startTime int64, ttl int, err error)
	// GetSessionTTL 仅获取session的剩余TTL
	// conversationID: 会话ID
	// 返回: ttl-剩余过期时间(秒), err-错误（如果key不存在返回0和nil）
	GetSessionTTL(ctx context.Context, conversationID string) (ttl int, err error)
	// RefreshSession 刷新session的过期时间，返回session是否存在及startTime
	// conversationID: 会话ID
	// ttl: 过期时间（秒）
	// 返回: exists-是否存在, startTime-会话开始时间, err-错误
	RefreshSession(ctx context.Context, conversationID string, ttl int) (exists bool, startTime int64, err error)
}
