package daresvo

import (
	"errors"
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

// MockDocRetrievalResultStrategy is a mock implementation of DocRetrievalResultStrategy
type MockDocRetrievalResultStrategy struct {
	ProcessCalled         bool
	ProcessAnswer         interface{}
	ProcessReturnAnswer   agentrespvo.DocRetrievalAnswer
	ProcessReturnError    error
	GetStrategyNameResult chatresenum.DocRetrievalStrategy
}

func (m *MockDocRetrievalResultStrategy) Process(answer interface{}) (agentrespvo.DocRetrievalAnswer, error) {
	m.ProcessCalled = true
	m.ProcessAnswer = answer

	return m.ProcessReturnAnswer, m.ProcessReturnError
}

func (m *MockDocRetrievalResultStrategy) GetStrategyName() chatresenum.DocRetrievalStrategy {
	return m.GetStrategyNameResult
}

func TestDocRetrievalResultStrategy_Interface(t *testing.T) {
	t.Parallel()

	// Test that MockDocRetrievalResultStrategy implements DocRetrievalResultStrategy
	var _ DocRetrievalResultStrategy = &MockDocRetrievalResultStrategy{}

	assert.True(t, true)
}

func TestMockDocRetrievalResultStrategy_Process_Success(t *testing.T) {
	t.Parallel()

	mock := &MockDocRetrievalResultStrategy{
		ProcessReturnAnswer: agentrespvo.DocRetrievalAnswer{
			Result: "success",
		},
		ProcessReturnError: nil,
	}

	answer := map[string]interface{}{"key": "value"}
	result, err := mock.Process(answer)

	assert.NoError(t, err)
	assert.True(t, mock.ProcessCalled)
	assert.Equal(t, answer, mock.ProcessAnswer)
	assert.Equal(t, "success", result.Result)
}

func TestMockDocRetrievalResultStrategy_Process_Error(t *testing.T) {
	t.Parallel()

	mock := &MockDocRetrievalResultStrategy{
		ProcessReturnError: errors.New("process error"),
	}

	answer := "test answer"
	_, err := mock.Process(answer)

	assert.Error(t, err)
	assert.True(t, mock.ProcessCalled)
}

func TestMockDocRetrievalResultStrategy_GetStrategyName(t *testing.T) {
	t.Parallel()

	mock := &MockDocRetrievalResultStrategy{
		GetStrategyNameResult: chatresenum.DocRetrievalStrategyStandard,
	}

	strategy := mock.GetStrategyName()

	assert.Equal(t, chatresenum.DocRetrievalStrategyStandard, strategy)
}

func TestDocRetrievalResultStrategy_Process_NilAnswer(t *testing.T) {
	t.Parallel()

	mock := &MockDocRetrievalResultStrategy{
		ProcessReturnAnswer: agentrespvo.DocRetrievalAnswer{},
		ProcessReturnError:  nil,
	}

	result, err := mock.Process(nil)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDocRetrievalResultStrategy_GetStrategyName_Custom(t *testing.T) {
	t.Parallel()

	mock := &MockDocRetrievalResultStrategy{
		GetStrategyNameResult: chatresenum.DocRetrievalStrategy("custom"),
	}

	strategy := mock.GetStrategyName()

	assert.Equal(t, chatresenum.DocRetrievalStrategy("custom"), strategy)
}
