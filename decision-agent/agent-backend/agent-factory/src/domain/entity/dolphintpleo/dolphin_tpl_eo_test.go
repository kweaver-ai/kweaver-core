package dolphintpleo

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestDolphinTplEo_Fields_All(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  cdaenum.DolphinTplKey
	}{
		{"memory retrieve", cdaenum.DolphinTplKeyMemoryRetrieve},
		{"temp file process", cdaenum.DolphinTplKeyTempFileProcess},
		{"doc retrieve", cdaenum.DolphinTplKeyDocRetrieve},
		{"graph retrieve", cdaenum.DolphinTplKeyGraphRetrieve},
		{"context organize", cdaenum.DolphinTplKeyContextOrganize},
		{"related questions", cdaenum.DolphinTplKeyRelatedQuestions},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			eo := &DolphinTplEo{
				Key:   tt.key,
				Name:  tt.name + " name",
				Value: tt.name + " value",
			}

			assert.Equal(t, tt.key, eo.Key)
			assert.Contains(t, eo.Name, "name")
			assert.Contains(t, eo.Value, "value")
		})
	}
}
