package agentsvc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOriginalChatResp_ValidJSON(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`{"answer":"Hello","think":"Testing"}`)
	result, err := agentSvc.originalChatResp(data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "Hello", result["answer"])
	assert.Equal(t, "Testing", result["think"])
}

func TestOriginalChatResp_EmptyJSON(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`{}`)
	result, err := agentSvc.originalChatResp(data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestOriginalChatResp_InvalidJSON(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`{invalid json}`)
	_, err := agentSvc.originalChatResp(data)

	// sonic.Unmarshal returns error for invalid JSON
	assert.Error(t, err)
}

func TestOriginalChatResp_EmptyData(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(``)
	result, err := agentSvc.originalChatResp(data)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOriginalChatResp_NestedJSON(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`{"outer":{"inner":"value"},"array":[1,2,3]}`)
	result, err := agentSvc.originalChatResp(data)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result["outer"])
	assert.NotNil(t, result["array"])
}

func TestOriginalChatResp_NullData(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`null`)
	result, err := agentSvc.originalChatResp(data)

	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestOriginalChatResp_ArrayData(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`[1,2,3]`)
	result, err := agentSvc.originalChatResp(data)

	// sonic.Unmarshal will fail for array into map[string]interface{}
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOriginalChatResp_StringData(t *testing.T) {
	t.Parallel()

	agentSvc := &agentSvc{}

	data := []byte(`"just a string"`)
	result, err := agentSvc.originalChatResp(data)

	// sonic.Unmarshal will fail for string into map[string]interface{}
	assert.Error(t, err)
	assert.Nil(t, result)
}
