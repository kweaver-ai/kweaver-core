package dainject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSessionSvc_ReturnsService(t *testing.T) {
	// t.Parallel() - 移除：此测试调用单例初始化函数，在并发环境下会导致 sync.Once 死锁
	// The NewSessionSvc function initializes with mock dependencies in test mode
	svc := NewSessionSvc()
	assert.NotNil(t, svc)
}
