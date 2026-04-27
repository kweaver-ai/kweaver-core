package daresvo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStandardDocRetrievalStrategy_Process_Valid(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Valid answer with all required fields
	answer := map[string]interface{}{
		"result": "test result",
		"full_result": map[string]interface{}{
			"text": "test content",
		},
	}

	result, err := strategy.Process(answer)
	assert.NoError(t, err)
	assert.Equal(t, "test result", result.Result)
	assert.Equal(t, "test content", result.FullResult.Text)
}

func TestStandardDocRetrievalStrategy_Process_WithErrorCode(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer with error code
	answer := map[string]interface{}{
		"error_code": "ERR_001",
		"message":    "test error",
	}

	result, err := strategy.Process(answer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "文档召回错误")
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_MissingResult(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer missing result field
	answer := map[string]interface{}{
		"full_result": map[string]interface{}{
			"text": "test content",
		},
	}

	result, err := strategy.Process(answer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缺少必需字段: result")
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_MissingFullResult(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer missing full_result field
	answer := map[string]interface{}{
		"result": []interface{}{},
	}

	result, err := strategy.Process(answer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缺少必需字段: full_result")
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_FullResultNotMap(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer with full_result as non-map
	answer := map[string]interface{}{
		"result":      []interface{}{},
		"full_result": "not a map",
	}

	result, err := strategy.Process(answer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缺少必需字段: full_result")
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_MissingText(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer with full_result but missing text
	answer := map[string]interface{}{
		"result": []interface{}{},
		"full_result": map[string]interface{}{
			"other_field": "value",
		},
	}

	result, err := strategy.Process(answer)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "缺少必需字段: full_result.text")
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_NonMapAnswer(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer is not a map (non-standard structure)
	answer := "string answer"

	result, err := strategy.Process(answer)
	assert.NoError(t, err) // Non-standard structure doesn't error
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_Process_NilAnswer(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()

	// Answer is nil
	result, err := strategy.Process(nil)
	assert.NoError(t, err) // Non-standard structure doesn't error
	assert.Empty(t, result)
}

func TestStandardDocRetrievalStrategy_GetStrategyName(t *testing.T) {
	t.Parallel()

	strategy := NewStandardDocRetrievalStrategy()
	name := strategy.GetStrategyName()
	assert.NotEmpty(t, name)
}
