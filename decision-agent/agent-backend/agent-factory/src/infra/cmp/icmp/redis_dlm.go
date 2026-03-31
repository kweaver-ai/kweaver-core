package icmp

import "context"

//go:generate mockgen -source=./redis_dlm.go -destination ./cmpmock/redis_dlm_mock.go -package cmpmock
type RedisDlmCmp interface {
	NewMutex(name string) (mutex RedisDlmMutexCmp)
}

type RedisDlmMutexCmp interface {
	Lock(ctx context.Context) (err error)
	Unlock() (err error)
}
