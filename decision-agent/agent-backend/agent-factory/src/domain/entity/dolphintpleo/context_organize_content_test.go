package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextOrganizeContent(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)

	assert.NotNil(t, content)
	assert.Equal(t, otherTpl, content.OtherTplStruct)
	assert.False(t, content.IsEnable)
	assert.Empty(t, content.DocRetrieveContent)
	assert.Empty(t, content.GraphRetrieveContent)
}

func TestContextOrganizeContent_LoadFromConfig_WithDocRetrieve(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}
	otherTpl.DocRetrieve.IsEnable = true

	content := NewContextOrganizeContent(otherTpl)

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.DocRetrieveContent)
	assert.NotEmpty(t, content.Other)
}

func TestContextOrganizeContent_LoadFromConfig_WithGraphRetrieve(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}
	otherTpl.GraphRetrieve.IsEnable = true

	content := NewContextOrganizeContent(otherTpl)

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.GraphRetrieveContent)
	assert.NotEmpty(t, content.Other)
}

func TestContextOrganizeContent_LoadFromConfig_DocRetrieveDisabled(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}
	otherTpl.DocRetrieve.IsEnable = true

	content := NewContextOrganizeContent(otherTpl)

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{
			{Key: cdaenum.DolphinTplKeyDocRetrieve, Enabled: false},
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.Empty(t, content.DocRetrieveContent)
}

func TestContextOrganizeContent_LoadFromConfig_GraphRetrieveDisabled(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}
	otherTpl.GraphRetrieve.IsEnable = true

	content := NewContextOrganizeContent(otherTpl)

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{
			{Key: cdaenum.DolphinTplKeyGraphRetrieve, Enabled: false},
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.Empty(t, content.GraphRetrieveContent)
}

func TestContextOrganizeContent_LoadFromConfig_NoReferences(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)

	config := &daconfvalobj.Config{}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Other)
	assert.Contains(t, content.Other, "query")
}

func TestContextOrganizeContent_ToString_WithReference(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}
	otherTpl.DocRetrieve.IsEnable = true

	content := NewContextOrganizeContent(otherTpl)
	content.DocRetrieveContent = "doc content"
	content.Other = "other content"
	content.referenceEnable = true

	result := content.ToString()

	assert.Contains(t, result, "如果有参考文档")
	assert.Contains(t, result, "doc content")
	assert.Contains(t, result, "other content")
}

func TestContextOrganizeContent_ToString_WithoutReference(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.Other = "other content"
	content.referenceEnable = false

	result := content.ToString()

	assert.NotContains(t, result, "如果有参考文档")
	assert.Contains(t, result, "other content")
}

func TestContextOrganizeContent_ToString_WithTempZoneContent(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.TempZoneContent = "temp zone content"
	content.Other = "other content"

	result := content.ToString()

	assert.Contains(t, result, "temp zone content")
	assert.Contains(t, result, "other content")
}

func TestContextOrganizeContent_ToString_WithGraphRetrieveContent(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.GraphRetrieveContent = "graph content"
	content.Other = "other content"

	result := content.ToString()

	assert.Contains(t, result, "graph content")
	assert.Contains(t, result, "other content")
}

func TestContextOrganizeContent_ToString_AllContentTypes(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.DocRetrieveContent = "doc content"
	content.GraphRetrieveContent = "graph content"
	content.TempZoneContent = "temp zone content"
	content.Other = "other content"
	content.referenceEnable = true

	result := content.ToString()

	assert.Contains(t, result, "如果有参考文档")
	assert.Contains(t, result, "doc content")
	assert.Contains(t, result, "graph content")
	assert.Contains(t, result, "temp zone content")
	assert.Contains(t, result, "other content")
}

func TestContextOrganizeContent_ToDolphinTplEo(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.DocRetrieveContent = "test content"
	content.referenceEnable = true

	eo := content.ToDolphinTplEo()

	require.NotNil(t, eo)
	assert.Equal(t, cdaenum.DolphinTplKeyContextOrganize, eo.Key)
	assert.NotEmpty(t, eo.Name)
	assert.Contains(t, eo.Value, "test content")
}

func TestContextOrganizeContent_LoadFromConfig_DataSourceWithDoc(t *testing.T) {
	t.Parallel()

	docSource := &datasourcevalobj.DocSource{}
	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.OtherTplStruct.DocRetrieve.IsEnable = true

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{},
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Doc: []*datasourcevalobj.DocSource{docSource},
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Other)
}

func TestContextOrganizeContent_LoadFromConfig_DataSourceWithGraph(t *testing.T) {
	t.Parallel()

	kgSource := &datasourcevalobj.KgSource{}
	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)
	content.OtherTplStruct.GraphRetrieve.IsEnable = true

	config := &daconfvalobj.Config{
		PreDolphin: []*daconfvalobj.DolphinTpl{},
		DataSource: &datasourcevalobj.RetrieverDataSource{
			Kg: []*datasourcevalobj.KgSource{kgSource},
		},
	}

	content.LoadFromConfig(config)

	assert.True(t, content.IsEnable)
	assert.NotEmpty(t, content.Other)
}

func TestContextOrganizeContent_ToString_EmptyContent(t *testing.T) {
	t.Parallel()

	otherTpl := &OtherTplStruct{
		DocRetrieve:    NewDocRetrieveContent(),
		GraphRetrieve:  NewGraphRetrieveContent(),
		MemoryRetrieve: NewMemoryRetrieveContent(),
	}

	content := NewContextOrganizeContent(otherTpl)

	result := content.ToString()

	assert.Empty(t, result)
}
