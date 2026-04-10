package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum/builtinagentenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
	"github.com/stretchr/testify/assert"
)

func TestNewDolphinTplMapStruct(t *testing.T) {
	t.Parallel()

	m := NewDolphinTplMapStruct()

	assert.NotNil(t, m)
	assert.NotNil(t, m.MemoryRetrieve)
	assert.NotNil(t, m.DocRetrieve)
	assert.NotNil(t, m.GraphRetrieve)
	assert.NotNil(t, m.ContextOrganize)
	assert.NotNil(t, m.RelatedQuestions)
}

func TestDolphinTplMapStruct_LoadFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                     string
		config                   *daconfvalobj.Config
		builtInAgentKey          builtinagentenum.AgentKey
		isNeedHandleBuiltinAgent bool
	}{
		{
			name: "load from config with data source",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Doc: []*datasourcevalobj.DocSource{
						{DsID: "test-ds"},
					},
				},
			},
			builtInAgentKey:          builtinagentenum.AgentKeyDocQA,
			isNeedHandleBuiltinAgent: true,
		},
		{
			name: "load from config without builtin agent handling",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Doc: []*datasourcevalobj.DocSource{
						{DsID: "test-ds-2"},
					},
				},
			},
			builtInAgentKey:          builtinagentenum.AgentKeyDocQA,
			isNeedHandleBuiltinAgent: false,
		},
		{
			name: "load from config for graph qa agent",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Kg: []*datasourcevalobj.KgSource{
						{
							KgID: "kg-123",
						},
					},
				},
			},
			builtInAgentKey:          builtinagentenum.AgentKeyGraphQA,
			isNeedHandleBuiltinAgent: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := NewDolphinTplMapStruct()

			// This should not panic
			assert.NotPanics(t, func() {
				m.LoadFromConfig(tt.config, tt.builtInAgentKey, tt.isNeedHandleBuiltinAgent)
			})

			// Verify the struct is still valid after loading
			assert.NotNil(t, m.MemoryRetrieve)
			assert.NotNil(t, m.DocRetrieve)
			assert.NotNil(t, m.GraphRetrieve)
			assert.NotNil(t, m.ContextOrganize)
			assert.NotNil(t, m.RelatedQuestions)
		})
	}
}

func TestDolphinTplMapStruct_LoadFromConfig_NilConfig(t *testing.T) {
	t.Parallel()

	m := NewDolphinTplMapStruct()

	// LoadFromConfig panics when config is nil, so we test for that
	assert.Panics(t, func() {
		m.LoadFromConfig(nil, builtinagentenum.AgentKeyDocQA, false)
	})

	// Struct should still be valid after panic recovery
	assert.NotNil(t, m.MemoryRetrieve)
	assert.NotNil(t, m.DocRetrieve)
}

func TestDolphinTplMapStruct_Fields(t *testing.T) {
	t.Parallel()

	memoryRetrieve := NewMemoryRetrieveContent()
	docRetrieve := NewDocRetrieveContent()
	graphRetrieve := NewGraphRetrieveContent()
	contextOrganize := NewContextOrganizeContent(&OtherTplStruct{
		MemoryRetrieve: memoryRetrieve,
		DocRetrieve:    docRetrieve,
		GraphRetrieve:  graphRetrieve,
	})
	relatedQuestions := NewRelatedQuestionsContent()

	m := &DolphinTplMapStruct{
		MemoryRetrieve:   memoryRetrieve,
		DocRetrieve:      docRetrieve,
		GraphRetrieve:    graphRetrieve,
		ContextOrganize:  contextOrganize,
		RelatedQuestions: relatedQuestions,
	}

	assert.Equal(t, memoryRetrieve, m.MemoryRetrieve)
	assert.Equal(t, docRetrieve, m.DocRetrieve)
	assert.Equal(t, graphRetrieve, m.GraphRetrieve)
	assert.Equal(t, contextOrganize, m.ContextOrganize)
	assert.Equal(t, relatedQuestions, m.RelatedQuestions)
}
