package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewGraphRetrieveContent(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	assert.NotNil(t, content)
	assert.Empty(t, content.Content)
	assert.False(t, content.IsEnable)
}

func TestGraphRetrieveContent_LoadFromConfig_BuiltInGraphQAAgent(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	config := &daconfvalobj.Config{}
	isBuiltInGraphQAAgent := true

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.Content)
}

func TestGraphRetrieveContent_LoadFromConfig_WithKgDataSource(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	kgSource := &datasourcevalobj.KgSource{}
	config := &daconfvalobj.Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: []*datasourcevalobj.KgSource{kgSource},
		},
	}
	isBuiltInGraphQAAgent := false

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Content)
	assert.Contains(t, content.Content, "graph_qa")
	assert.Contains(t, content.Content, "graph_retrieval_res")
}

func TestGraphRetrieveContent_LoadFromConfig_NoKgDataSource(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	config := &daconfvalobj.Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: []*datasourcevalobj.KgSource{},
		},
	}
	isBuiltInGraphQAAgent := false

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.Content)
}

func TestGraphRetrieveContent_LoadFromConfig_NilDataSource(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	config := &daconfvalobj.Config{
		DataSource: nil,
	}
	isBuiltInGraphQAAgent := false

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.Content)
}

func TestGraphRetrieveContent_LoadFromConfig_NilKgField(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	config := &daconfvalobj.Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: nil,
		},
	}
	isBuiltInGraphQAAgent := false

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.False(t, content.IsEnable)
	assert.Empty(t, content.Content)
}

func TestGraphRetrieveContent_LoadFromConfig_MultipleKgSources(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	kgSource1 := &datasourcevalobj.KgSource{}
	kgSource2 := &datasourcevalobj.KgSource{}
	config := &daconfvalobj.Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: []*datasourcevalobj.KgSource{kgSource1, kgSource2},
		},
	}
	isBuiltInGraphQAAgent := false

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Content)
}

func TestGraphRetrieveContent_ToString_Empty(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	result := content.ToString()

	assert.Empty(t, result)
}

func TestGraphRetrieveContent_ToString_WithContent(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()
	content.Content = "test graph content"

	result := content.ToString()

	assert.Equal(t, "test graph content", result)
}

func TestGraphRetrieveContent_ToDolphinTplEo(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()
	content.Content = "test value"

	eo := content.ToDolphinTplEo()

	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.DolphinTplKeyGraphRetrieve, eo.Key)
	assert.NotEmpty(t, eo.Name)
	assert.Equal(t, "test value", eo.Value)
}

func TestGraphRetrieveContent_ToDolphinTplEo_EmptyContent(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	eo := content.ToDolphinTplEo()

	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.DolphinTplKeyGraphRetrieve, eo.Key)
	assert.NotEmpty(t, eo.Name)
	assert.Empty(t, eo.Value)
}

func TestGraphRetrieveContent_LoadFromConfig_BuiltInTakesPrecedence(t *testing.T) {
	t.Parallel()

	content := NewGraphRetrieveContent()

	kgSource := &datasourcevalobj.KgSource{}
	config := &daconfvalobj.Config{
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: []*datasourcevalobj.KgSource{kgSource},
		},
	}
	isBuiltInGraphQAAgent := true

	content.LoadFromConfig(config, isBuiltInGraphQAAgent)

	// Built-in agent should disable regardless of data source
	assert.False(t, content.IsEnable)
	assert.Empty(t, content.Content)
}
