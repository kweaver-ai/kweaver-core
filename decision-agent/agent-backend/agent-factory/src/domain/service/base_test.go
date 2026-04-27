package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewSvcBase(t *testing.T) {
	t.Parallel()

	svcBase := NewSvcBase()

	assert.NotNil(t, svcBase)
	assert.NotNil(t, svcBase.Logger)
}

func TestNewSvcBase_CreatesNewInstance(t *testing.T) {
	t.Parallel()

	svcBase1 := NewSvcBase()
	svcBase2 := NewSvcBase()

	assert.NotNil(t, svcBase1)
	assert.NotNil(t, svcBase2)
	// Logger is a singleton, so both instances will have the same logger
	assert.Equal(t, svcBase1.Logger, svcBase2.Logger)
	// But the service base instances themselves are different
	assert.NotSame(t, svcBase1, svcBase2)
}

func TestSvcBase_StructFields(t *testing.T) {
	t.Parallel()

	svcBase := &SvcBase{}

	assert.NotNil(t, svcBase)
	// Logger is nil when using struct literal
	assert.Nil(t, svcBase.Logger)
}

func TestGetMockedDlm(t *testing.T) {
	t.Parallel()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	dlm := getMockedDlm(ctrl)
	assert.NotNil(t, dlm)

	mu := dlm.NewMutex("test-key")
	assert.NotNil(t, mu)

	err := mu.Lock(t.Context())
	assert.NoError(t, err)

	unlockErr := mu.Unlock()
	assert.NoError(t, unlockErr)
}
