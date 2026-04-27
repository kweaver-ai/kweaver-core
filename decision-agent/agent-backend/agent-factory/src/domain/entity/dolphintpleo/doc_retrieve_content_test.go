package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj/datasourcevalobj"
)

func TestNewDocRetrieveContent(t *testing.T) {
	t.Parallel()

	content := NewDocRetrieveContent()
	if content == nil { //nolint:staticcheck
		t.Error("NewDocRetrieveContent() should return non-nil")
	}

	if content.Content != "" { //nolint:staticcheck
		t.Errorf("Content should be empty, got %q", content.Content)
	}

	if content.IsEnable {
		t.Error("IsEnable should be false")
	}
}

func TestDocRetrieveContent_LoadFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                string
		config              *daconfvalobj.Config
		isBuiltInDocQAAgent bool
		wantIsEnable        bool
		wantContent         string
	}{
		{
			name: "built-in doc QA agent",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Doc: []*datasourcevalobj.DocSource{{}},
				},
			},
			isBuiltInDocQAAgent: true,
			wantIsEnable:        false,
			wantContent:         "",
		},
		{
			name: "doc data source enabled",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Doc: []*datasourcevalobj.DocSource{{}},
				},
			},
			isBuiltInDocQAAgent: false,
			wantIsEnable:        true,
			wantContent: `
/judge/(tools=["doc_qa"], history=True)判断【$query】是否需要到文档中召回，如果不需要召回，则直接返回\"不需要文档召回\"，否则执行工具对【$query】进行召回 -> doc_retrieval_res
`,
		},
		{
			name: "no doc data source",
			config: &daconfvalobj.Config{
				DataSource: &datasourcevalobj.RetrieverDataSource{
					Doc: []*datasourcevalobj.DocSource{},
				},
			},
			isBuiltInDocQAAgent: false,
			wantIsEnable:        false,
			wantContent:         "",
		},
		{
			name: "nil data source",
			config: &daconfvalobj.Config{
				DataSource: nil,
			},
			isBuiltInDocQAAgent: false,
			wantIsEnable:        false,
			wantContent:         "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := NewDocRetrieveContent()
			content.LoadFromConfig(tt.config, tt.isBuiltInDocQAAgent)

			if content.IsEnable != tt.wantIsEnable {
				t.Errorf("IsEnable = %v, want %v", content.IsEnable, tt.wantIsEnable)
			}

			if content.Content != tt.wantContent {
				t.Errorf("Content = %q, want %q", content.Content, tt.wantContent)
			}
		})
	}
}

func TestDocRetrieveContent_ToString(t *testing.T) {
	t.Parallel()

	content := NewDocRetrieveContent()
	content.Content = "test content"

	result := content.ToString()
	if result != "test content" {
		t.Errorf("ToString() = %q, want %q", result, "test content")
	}
}

func TestDocRetrieveContent_ToDolphinTplEo(t *testing.T) {
	t.Parallel()

	content := NewDocRetrieveContent()
	content.Content = "test"

	eo := content.ToDolphinTplEo()
	if eo.Key != cdaenum.DolphinTplKeyDocRetrieve {
		t.Errorf("Key = %v, want %v", eo.Key, cdaenum.DolphinTplKeyDocRetrieve)
	}

	if eo.Value != "test" {
		t.Errorf("Value = %q, want %q", eo.Value, "test")
	}
}
