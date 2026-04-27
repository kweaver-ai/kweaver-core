package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryRetrieveContent(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	assert.NotNil(t, content)
	assert.Empty(t, content.RelevantMemory)
	assert.Empty(t, content.Other)
	assert.False(t, content.IsEnable)
}

func TestMemoryRetrieveContent_LoadFromConfig_WithMemoryEnabled(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	isEnabled := true
	config := &daconfvalobj.Config{
		MemoryCfg: &daconfvalobj.MemoryCfg{
			IsEnabled: isEnabled,
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.RelevantMemory)
	assert.Contains(t, content.RelevantMemory, "search_memory")
	assert.Contains(t, content.RelevantMemory, "relevant_memories")
	assert.NotEmpty(t, content.Other)
}

func TestMemoryRetrieveContent_LoadFromConfig_WithMemoryDisabled(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	isEnabled := false
	config := &daconfvalobj.Config{
		MemoryCfg: &daconfvalobj.MemoryCfg{
			IsEnabled: isEnabled,
		},
	}

	content.LoadFromConfig(config)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.RelevantMemory)
	assert.Empty(t, content.Other)
}

func TestMemoryRetrieveContent_LoadFromConfig_NilMemoryCfg(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	config := &daconfvalobj.Config{
		MemoryCfg: nil,
	}

	content.LoadFromConfig(config)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.RelevantMemory)
	assert.Empty(t, content.Other)
}

func TestMemoryRetrieveContent_LoadFromConfig_AppendsToOther(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()
	content.Other = "existing content"

	isEnabled := true
	config := &daconfvalobj.Config{
		MemoryCfg: &daconfvalobj.MemoryCfg{
			IsEnabled: isEnabled,
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.Contains(t, content.Other, "existing content")
	assert.Contains(t, content.Other, "relevant_memories")
}

func TestMemoryRetrieveContent_ToString_Empty(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	result := content.ToString()

	assert.Empty(t, result)
}

func TestMemoryRetrieveContent_ToString_WithRelevantMemory(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()
	content.RelevantMemory = "memory content"

	result := content.ToString()

	assert.Equal(t, "memory content", result)
}

func TestMemoryRetrieveContent_ToString_WithOther(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()
	content.Other = "other content"

	result := content.ToString()

	assert.Equal(t, "other content", result)
}

func TestMemoryRetrieveContent_ToString_WithBoth(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()
	content.RelevantMemory = "memory content"
	content.Other = "other content"

	result := content.ToString()

	assert.Contains(t, result, "memory content")
	assert.Contains(t, result, "other content")
}

func TestMemoryRetrieveContent_ToDolphinTplEo(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()
	content.RelevantMemory = "test value"

	eo := content.ToDolphinTplEo()

	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.DolphinTplKeyMemoryRetrieve, eo.Key)
	assert.NotEmpty(t, eo.Name)
	assert.Equal(t, "test value", eo.Value)
}

func TestMemoryRetrieveContent_ToDolphinTplEo_EmptyContent(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	eo := content.ToDolphinTplEo()

	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.DolphinTplKeyMemoryRetrieve, eo.Key)
	assert.NotEmpty(t, eo.Name)
	assert.Empty(t, eo.Value)
}

func TestMemoryRetrieveContent_LoadFromConfig_MultipleCalls(t *testing.T) {
	t.Parallel()

	content := NewMemoryRetrieveContent()

	isEnabled := true
	config := &daconfvalobj.Config{
		MemoryCfg: &daconfvalobj.MemoryCfg{
			IsEnabled: isEnabled,
		},
	}

	// First call
	content.LoadFromConfig(config)
	firstOther := content.Other

	// Second call - should append
	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Other)
	assert.NotEqual(t, firstOther, content.Other)
}
