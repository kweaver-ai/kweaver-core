package panichelper

import (
	"errors"
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/cmp/icmp/cmpmock"

	//"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/test/mock_log"
	"go.uber.org/mock/gomock"

	"github.com/stretchr/testify/assert"
)

func ForRecovery(logger icmp.Logger) {
	defer Recovery(logger)
	panic("test Recovery")
}

func TestRecovery(t *testing.T) {
	t.Parallel()

	t.Run("recovery with panic", func(t *testing.T) {
		t.Parallel()
		ctl := gomock.NewController(t)
		logger := cmpmock.NewMockLogger(ctl)
		logger.EXPECT().Errorln(gomock.Any()).DoAndReturn(func(args ...interface{}) interface{} {
			t.Log(args...)
			return nil
		})

		ForRecovery(logger)
	})

	t.Run("recovery without panic", func(t *testing.T) {
		t.Parallel()
		ctl := gomock.NewController(t)
		logger := cmpmock.NewMockLogger(ctl)
		logger.EXPECT().Errorln(gomock.Any()).Times(0) // No panic should be logged

		// Function that does not panic
		noPanicFunc := func() {
			defer Recovery(logger)
			// No panic here
		}

		noPanicFunc()
	})
}

func TestRecoveryNoPanic(t *testing.T) {
	t.Parallel()

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	logger := cmpmock.NewMockLogger(ctl)

	// Call Recovery directly without a panic
	// This should do nothing and just return
	Recovery(logger)
	// No panic occurred, so no error should be logged
	// The test completes successfully
}

func ForRecoveryAndSetErr(logger icmp.Logger, err *error) {
	defer RecoveryAndSetErr(logger, err)
	panic("test RecoveryAndSetErr")
}

func TestRecoveryAndSetErr(t *testing.T) {
	t.Parallel()

	ctl := gomock.NewController(t)
	logger := cmpmock.NewMockLogger(ctl)
	logger.EXPECT().Errorln(gomock.Any()).DoAndReturn(func(args ...interface{}) interface{} {
		t.Log(args...)
		return nil
	})

	var err error

	ForRecoveryAndSetErr(logger, &err)

	// 1.检查err的值是否是"test RecoveryAndSetErr"
	assert.Equal(t, "test RecoveryAndSetErr", err.Error())
}

type customErr struct {
	msg string
}

func (c *customErr) Error() string {
	return c.msg
}

func ForRecoveryAndSetErrCustomErr(logger icmp.Logger, err *error) {
	defer RecoveryAndSetErr(logger, err)

	_err := &customErr{msg: "test RecoveryAndSetErr2"}
	panic(_err)
}

func TestRecoveryAndSetErrCustomErr(t *testing.T) {
	t.Parallel()

	ctl := gomock.NewController(t)
	logger := cmpmock.NewMockLogger(ctl)
	logger.EXPECT().Errorln(gomock.Any()).DoAndReturn(func(args ...interface{}) interface{} {
		t.Log(args...)
		return nil
	})

	var err error

	ForRecoveryAndSetErrCustomErr(logger, &err)

	// 1.检查err的值是否是"test RecoveryAndSetErr2"
	assert.Equal(t, "test RecoveryAndSetErr2", err.Error())

	// 2.检查err是否是customErr类型
	var _customErr *customErr
	ok := errors.As(err, &_customErr)
	assert.True(t, ok)
	assert.Equal(t, "test RecoveryAndSetErr2", _customErr.msg)
}

func TestRecovery_DebugMode(t *testing.T) {
	// t.Parallel() - 移除：此测试使用 t.Setenv() 修改环境变量，不能与 t.Parallel() 同时使用
	// Set debug mode environment variable
	t.Setenv("AGENT_FACTORY_DEBUG_MODE", "true")

	ctl := gomock.NewController(t)
	logger := cmpmock.NewMockLogger(ctl)
	logger.EXPECT().Errorln(gomock.Any()).DoAndReturn(func(args ...interface{}) interface{} {
		t.Log(args...)
		return nil
	})

	func() {
		defer Recovery(logger)
		panic("debug mode test")
	}()
}
