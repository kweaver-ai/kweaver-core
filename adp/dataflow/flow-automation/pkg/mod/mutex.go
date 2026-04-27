package mod

import (
	"context"
	"time"
)

// DistributedMutex
type DistributedMutex interface {
	Lock(ctx context.Context, ops ...LockOptionOp) error
	Unlock(ctx context.Context) error
}

// LockOption
type LockOption struct {
	TTL               time.Duration
	ReentrantIdentity string
	SpinInterval      time.Duration
}

type LockOptionOp func(option *LockOption)

// NewLockOption
func NewLockOption(ops []LockOptionOp) *LockOption {
	opt := &LockOption{
		TTL:               30 * time.Second,
		ReentrantIdentity: "",
		SpinInterval:      time.Millisecond * 100,
	}

	for _, op := range ops {
		op(opt)
	}
	return opt
}

// LockTTL configured lock ttl, default values: 30s
func LockTTL(d time.Duration) LockOptionOp {
	return func(option *LockOption) {
		if d != 0 {
			option.TTL = d
		}
	}
}

// Reentrant mean lock it reentrant
func Reentrant(identity string) LockOptionOp {
	return func(option *LockOption) {
		option.ReentrantIdentity = identity
	}
}
