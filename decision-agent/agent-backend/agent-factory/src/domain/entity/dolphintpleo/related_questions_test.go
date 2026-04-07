package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/daconfvalobj"
)

func TestNewRelatedQuestionsContent(t *testing.T) {
	t.Parallel()

	content := NewRelatedQuestionsContent()
	if content == nil { //nolint:staticcheck
		t.Error("NewRelatedQuestionsContent() should return non-nil")
	}

	if content.Content != "" { //nolint:staticcheck
		t.Errorf("Content should be empty, got %q", content.Content)
	}

	if content.IsEnable {
		t.Error("IsEnable should be false")
	}
}

func TestRelatedQuestionsContent_LoadFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		config       *daconfvalobj.Config
		wantIsEnable bool
	}{
		{
			name: "related question enabled",
			config: &daconfvalobj.Config{
				RelatedQuestion: &daconfvalobj.RelatedQuestion{
					IsEnabled: true,
				},
			},
			wantIsEnable: true,
		},
		{
			name: "related question disabled",
			config: &daconfvalobj.Config{
				RelatedQuestion: &daconfvalobj.RelatedQuestion{
					IsEnabled: false,
				},
			},
			wantIsEnable: false,
		},
		{
			name: "nil related question",
			config: &daconfvalobj.Config{
				RelatedQuestion: nil,
			},
			wantIsEnable: false,
		},
		{
			name:         "nil config",
			config:       nil,
			wantIsEnable: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			content := NewRelatedQuestionsContent()
			content.LoadFromConfig(tt.config)

			if content.IsEnable != tt.wantIsEnable {
				t.Errorf("IsEnable = %v, want %v", content.IsEnable, tt.wantIsEnable)
			}
		})
	}
}

func TestRelatedQuestionsContent_ToString(t *testing.T) {
	t.Parallel()

	content := NewRelatedQuestionsContent()
	content.Content = "test content"

	result := content.ToString()
	if result != "test content" {
		t.Errorf("ToString() = %q, want %q", result, "test content")
	}
}

func TestRelatedQuestionsContent_ToDolphinTplEo(t *testing.T) {
	t.Parallel()

	content := NewRelatedQuestionsContent()
	content.Content = "test"

	eo := content.ToDolphinTplEo()
	if eo.Key != cdaenum.DolphinTplKeyRelatedQuestions {
		t.Errorf("Key = %v, want %v", eo.Key, cdaenum.DolphinTplKeyRelatedQuestions)
	}

	if eo.Value != "test" {
		t.Errorf("Value = %q, want %q", eo.Value, "test")
	}
}
