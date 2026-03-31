package icmp

import (
	"github.com/go-redis/redis/v8"
)

//go:generate mockgen -source=./redis.go -destination ./cmpmock/redis_mock.go -package cmpmock
type RedisCmp interface {
	GetClient() (client redis.UniversalClient)
}
