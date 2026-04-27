package daresvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/chat_enum/chatresenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/agentrespvo"
	"github.com/stretchr/testify/assert"
)

func TestNewDocRetrievalManager(t *testing.T) {
	t.Parallel()

	manager := NewDocRetrievalManager()

	assert.NotNil(t, manager)
	assert.NotNil(t, manager.strategies)
}

func TestDocRetrievalManager_RegisterStrategy(t *testing.T) {
	t.Parallel()

	manager := NewDocRetrievalManager()
	initialLen := len(manager.strategies)

	// Create a mock strategy
	strategy := &mockDocRetrievalStrategy{
		name: "test_strategy",
	}

	manager.RegisterStrategy(strategy)

	assert.Len(t, manager.strategies, initialLen+1)
}

func TestDocRetrievalManager_ProcessResult_StandardStrategy(t *testing.T) {
	t.Parallel()

	manager := NewDocRetrievalManager()

	// Standard strategy expects a map with result and full_result structure
	answer := map[string]interface{}{
		"result": "Test answer content",
		"full_result": map[string]interface{}{
			"text":       "Full text content",
			"references": []interface{}{},
		},
	}
	result, err := manager.ProcessResult(answer, chatresenum.DocRetrievalStrategyStandard)

	// The standard strategy should be registered
	assert.NoError(t, err)
	assert.NotEmpty(t, result.Result)
	assert.Equal(t, "Test answer content", result.Result)
}

func TestDocRetrievalManager_ProcessResult_InvalidStrategy(t *testing.T) {
	t.Parallel()

	manager := NewDocRetrievalManager()

	answer := "Test answer content"
	_, err := manager.ProcessResult(answer, "invalid_strategy")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "未找到合适的处理策略")
}

// mockDocRetrievalStrategy is a mock implementation of DocRetrievalResultStrategy
type mockDocRetrievalStrategy struct {
	name string
}

func (m *mockDocRetrievalStrategy) GetStrategyName() chatresenum.DocRetrievalStrategy {
	return chatresenum.DocRetrievalStrategyStandard // Return a valid strategy for testing
}

func (m *mockDocRetrievalStrategy) Process(answer interface{}) (agentrespvo.DocRetrievalAnswer, error) {
	return agentrespvo.DocRetrievalAnswer{
		Result: answer.(string),
	}, nil
}
