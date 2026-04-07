package redisdlmcmp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/cmp/icmp"
)

type redisDlmMutex struct {
	redSyncMutex     *redsync.Mutex
	done             chan bool
	watchDogInterval time.Duration
	deleteValueFunc  func(value string) error

	logger icmp.Logger
}

var _ icmp.RedisDlmMutexCmp = &redisDlmMutex{}

// watchDog periodically extends the lock
func (m *redisDlmMutex) watchDog() {
	interval := m.watchDogInterval
	if interval == 0 {
		panic("WatchDogInterval is not set")
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ok, err := m.redSyncMutex.Extend()
			if err != nil {
				m.logger.Errorln("Failed to extend lock:", err)
			} else if !ok {
				m.logger.Errorln("Failed to extend lock: not successes")
			} else {
				m.logger.Debugln("Extend lock success")
			}
		case <-m.done:
			m.logger.Debugln("WatchDog exit")
			return
		}
	}
}

func (m *redisDlmMutex) Lock(ctx context.Context) (err error) {
	err = m.redSyncMutex.LockContext(ctx)
	if err != nil {
		err = fmt.Errorf("[redisDlmMutex]: failed to lock: %w", err)
		return
	}

	m.done = make(chan bool)

	go m.watchDog()

	return
}

// Unlock unlocks the mutex
func (m *redisDlmMutex) Unlock() (err error) {
	// 1. close watchDog
	close(m.done)

	// 2. unlock
	ok, err := m.redSyncMutex.Unlock()
	if err != nil {
		err = fmt.Errorf("[redisDlmCmp]: failed to unlock: %w", err)
		return
	}

	if !ok {
		//nolint:goerr113
		err = errors.New("[redisDlmCmp]: failed to unlock: not successes")
		return
	}

	// 3. delete value from db
	value := m.redSyncMutex.Value()
	if m.deleteValueFunc != nil {
		err = m.deleteValueFunc(value)
		if err != nil {
			err = fmt.Errorf("[redisDlmCmp]: failed to delete value after unlock: %w", err)
			return
		}
	}

	return
}
