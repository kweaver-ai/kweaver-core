package lock

import (
	"context"
	"fmt"
	"time"

	"github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/store"
)

const internalTime = 3 * time.Second

type DistributeLock interface {
	GetLockKey() string
	Lock(ctx context.Context, expirtTime time.Duration) error
	TryLock(ctx context.Context, expirtTime time.Duration, isNeedSchedule bool) error
	Release() (bool, error)
	GetErrChannel() chan error
	ResetErrChannel(args ...chan error)
	ClearErrChannel()
}

const (
	// distributed redis lock
	defaultLockKeyPrefix = "DistributeLock"
	defaultExpiry        = 30 * time.Second
	defaultSpinsMinTime  = 10 * time.Second
	errChanBufferSize    = 10
)

// DistributeLock 分布式锁结构体
type distributeLock struct {
	rds    store.RDB
	config *configOption
	lock   *lock
}

// ConfigOption 配置信息
type configOption struct {
	lockKeyPrefix string
}

// DistLock 锁信息
type lock struct {
	// redis key
	lockName     string
	field        string
	expiry       time.Duration
	spinsMinTime time.Duration
	errChan      chan error
	cancel       context.CancelFunc
}

// NewDistributeLock new distributeLock instance
func NewDistributeLock(rds store.RDB, lockName, field string) DistributeLock {
	cf := &configOption{
		lockKeyPrefix: defaultLockKeyPrefix,
	}
	dlock := &lock{
		lockName:     defaultLockKeyPrefix + ":" + lockName,
		field:        field,
		expiry:       defaultExpiry,
		spinsMinTime: defaultSpinsMinTime,
		errChan:      make(chan error, errChanBufferSize),
	}
	dl := &distributeLock{
		rds:    rds,
		config: cf,
		lock:   dlock,
	}
	return dl
}

// GetLockKey get lock key
func (dl *distributeLock) GetLockKey() string {
	return dl.lock.lockName
}

// Lock is a normal and will not have cas to compete lock,
// expirtTime default expirtTime 30s.
func (dl *distributeLock) Lock(ctx context.Context, expirtTime time.Duration) error {
	dl.lock.expiry = expirtTime
	return dl.tryLock(ctx, false)
}

// TryLock is a relatively fair lock and a retry mechanism.
// if the lock is successful, it will return true.
// If the lock fails, it will enter cas or it will return false if it times out.
// if param isNeedSchedule false, expirtTime default expirtTime 30s,
// but if isNeedSchedule is true, it will open an additional thread to ensure that the lock will not expire in advance,
// which means that you must release the lock manually, otherwise a deadlock will occur.
func (dl *distributeLock) TryLock(ctx context.Context, expirtTime time.Duration, isNeedSchedule bool) error {
	dl.lock.expiry = expirtTime
	err := dl.tryLock(ctx, isNeedSchedule)

	if err == nil {
		return nil
	}
	return dl.cas(ctx, isNeedSchedule)
}

func (dl *distributeLock) Release() (bool, error) {
	return dl.rds.Unlock(dl.lock.lockName, dl.lock.field)
}

func (dl *distributeLock) tryLock(ctx context.Context, isNeedSchedule bool) error {
	err := dl.rds.TryLock(dl.lock.lockName, dl.lock.field, dl.lock.expiry)
	if err == nil && isNeedSchedule {
		// renew ground
		go dl.scheduleExpirationRenewal(ctx)
	}
	return err
}

// scheduleExpirationRenewal renew lock
func (dl *distributeLock) scheduleExpirationRenewal(ctx context.Context) {
	if dl.lock.cancel != nil {
		dl.lock.cancel()
	}

	renewCtx, cancel := context.WithCancel(ctx)
	dl.lock.cancel = cancel

	go func(ctx context.Context) {
		ticker := time.NewTicker(dl.lock.expiry / 3)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				lockValue, err := dl.rds.Get(dl.lock.lockName)
				// key exist or redis server abnormal or key is nil
				if err != nil {
					dl.sendError(fmt.Errorf("[scheduleExpirationRenewal] Faild to get key, err: %s", err.Error()))
					continue
				}
				if lockValue != dl.lock.field {
					continue
				}
				success, err := dl.rds.Renew(dl.lock.lockName, dl.lock.field, dl.lock.expiry/time.Millisecond)
				if err != nil {
					dl.sendError(fmt.Errorf("[scheduleExpirationRenewal] Renew failed, err: %s", err.Error()))
				} else if !success {
					dl.sendError(fmt.Errorf("[scheduleExpirationRenewal] Lock expired or released"))
					return
				}
			}
		}
	}(renewCtx)
}

// cas Spin grab lock
func (dl *distributeLock) cas(ctx context.Context, isNeedSchedule bool) error {
	var timer *time.Timer
	for {
		var sleepTime time.Duration
		ttl, err := dl.rds.GetExpireTime(dl.lock.lockName)
		if err != nil {
			return err
		}
		if ttl < 0 {
			ttl = dl.lock.expiry
		}
		if ttl <= dl.lock.spinsMinTime {
			sleepTime = time.Duration(ttl)
		} else {
			sleepTime = time.Duration(ttl * 2 / 3)
		}
		if timer == nil {
			timer = time.NewTimer(sleepTime)
			defer timer.Stop()
		} else {
			timer.Reset(sleepTime)
		}

		err = dl.tryLock(ctx, isNeedSchedule)
		if err == nil {
			return nil
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timer.C:
		}
	}
}

// sendError 非阻塞发送错误
func (dl *distributeLock) sendError(err error) {
	select {
	case dl.lock.errChan <- err:
	default:
	}
}

// GetErrChannel return err msg channel
func (dl *distributeLock) GetErrChannel() chan error {
	return dl.lock.errChan
}

// ResetErrChannel reset err channel
func (dl *distributeLock) ResetErrChannel(args ...chan error) {
	oldChan := dl.lock.errChan
	if oldChan != nil {
		close(oldChan)
		// 清空可能残留的消息
		go func() {
			for range oldChan {
			}
		}()
	}
	if len(args) == 0 {
		dl.lock.errChan = make(chan error, errChanBufferSize)
		return
	}
	dl.lock.errChan = args[0]
}

// ClearErrChannel 清空错误channel
func (dl *distributeLock) ClearErrChannel() {
	timeout := time.After(internalTime)
	for {
		select {
		case <-timeout:
			return
		case _, ok := <-dl.lock.errChan:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
