package dainject

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConversationSvc_SingletonAndConstruct(t *testing.T) {
	// t.Parallel() - 移除：此测试调用单例初始化函数，在并发环境下会导致 sync.Once 死锁
	initInjectGlobalConfig(t)
	resetInjectSingletons()

	first := NewConversationSvc()
	second := NewConversationSvc()

	assert.NotNil(t, first)
	assert.Same(t, first, second)
}
