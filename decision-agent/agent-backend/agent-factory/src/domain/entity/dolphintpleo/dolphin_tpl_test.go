package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
)

func TestDolphinTplEo(t *testing.T) {
	t.Parallel()

	eo := &DolphinTplEo{
		Key:   cdaenum.DolphinTplKeyDocRetrieve,
		Name:  "Doc Retrieve Template",
		Value: "/judge/doc_qa",
	}

	if eo.Key != cdaenum.DolphinTplKeyDocRetrieve {
		t.Errorf("Key = %v, want %v", eo.Key, cdaenum.DolphinTplKeyDocRetrieve)
	}

	if eo.Name != "Doc Retrieve Template" {
		t.Errorf("Name = %q, want %q", eo.Name, "Doc Retrieve Template")
	}

	if eo.Value != "/judge/doc_qa" {
		t.Errorf("Value = %q, want %q", eo.Value, "/judge/doc_qa")
	}
}

func TestDolphinTplEo_Empty(t *testing.T) {
	t.Parallel()

	eo := &DolphinTplEo{}

	if eo.Key != "" {
		t.Errorf("Key should be empty")
	}

	if eo.Name != "" {
		t.Errorf("Name should be empty")
	}

	if eo.Value != "" {
		t.Errorf("Value should be empty")
	}
}

func TestDolphinTplEo_DifferentKeys(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		key   cdaenum.DolphinTplKey
		value string
	}{
		{
			name:  "Doc Retrieve Key",
			key:   cdaenum.DolphinTplKeyDocRetrieve,
			value: "/doc_retrieve",
		},
		{
			name:  "Related Questions Key",
			key:   cdaenum.DolphinTplKeyRelatedQuestions,
			value: "/related_questions",
		},
		{
			name:  "Temp File Process Key",
			key:   cdaenum.DolphinTplKeyTempFileProcess,
			value: "/temp_file_process",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo := &DolphinTplEo{
				Key:   tt.key,
				Name:  tt.name,
				Value: tt.value,
			}

			if eo.Key != tt.key {
				t.Errorf("Key = %v, want %v", eo.Key, tt.key)
			}

			if eo.Value != tt.value {
				t.Errorf("Value = %q, want %q", eo.Value, tt.value)
			}
		})
	}
}
