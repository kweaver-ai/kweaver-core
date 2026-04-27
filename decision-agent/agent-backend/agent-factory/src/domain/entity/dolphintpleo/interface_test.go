package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
)

// MockDolphinTpl is a mock implementation of IDolphinTpl for testing
type MockDolphinTpl struct {
	LoadFromConfigCalled bool
	ToStringCalled       bool
	Config               *daconfvalobj.Config
	ToStringResult       string
}

func (m *MockDolphinTpl) LoadFromConfig(config *daconfvalobj.Config) {
	m.LoadFromConfigCalled = true
	m.Config = config
}

func (m *MockDolphinTpl) ToString() string {
	m.ToStringCalled = true
	return m.ToStringResult
}

func TestIDolphinTpl_Interface(t *testing.T) {
	t.Parallel()

	// Test that MockDolphinTpl implements IDolphinTpl
	var _ IDolphinTpl = &MockDolphinTpl{}

	assert.True(t, true)
}

func TestMockDolphinTpl_LoadFromConfig(t *testing.T) {
	t.Parallel()

	mock := &MockDolphinTpl{}
	config := daconfvalobj.NewConfig()

	mock.LoadFromConfig(config)

	assert.True(t, mock.LoadFromConfigCalled)
	assert.Equal(t, config, mock.Config)
}

func TestMockDolphinTpl_ToString(t *testing.T) {
	t.Parallel()

	mock := &MockDolphinTpl{
		ToStringResult: "test result",
	}

	result := mock.ToString()

	assert.True(t, mock.ToStringCalled)
	assert.Equal(t, "test result", result)
}

func TestMockDolphinTpl_ToString_Empty(t *testing.T) {
	t.Parallel()

	mock := &MockDolphinTpl{}

	result := mock.ToString()

	assert.True(t, mock.ToStringCalled)
	assert.Empty(t, result)
}
