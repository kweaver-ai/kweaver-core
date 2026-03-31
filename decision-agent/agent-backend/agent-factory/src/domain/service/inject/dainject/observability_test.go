package dainject

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	v3portdrivermock "github.com/kweaver-ai/decision-agent/agent-factory/src/port/driver/iv3portdriver/v3portdrivermock"
)

func TestNewObservabilitySvc_SingletonAndConstruct(t *testing.T) {
	// t.Parallel() - 移除：此测试调用单例初始化函数，在并发环境下会导致 sync.Once 死锁
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	initInjectGlobalConfig(t)
	resetInjectSingletons()

	// Pre-inject squareSvc mock to avoid real DB repo constructors panicking
	squareSvcOnce.Do(func() {
		squareSvcImpl = v3portdrivermock.NewMockISquareSvc(ctrl)
	})

	first := NewObservabilitySvc()
	second := NewObservabilitySvc()

	assert.NotNil(t, first)
	assert.Same(t, first, second)
}
