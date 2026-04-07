package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
)

func TestDolphinTpl_ValObjCheck_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tpl  *DolphinTpl
	}{
		{
			name: "valid doc retrieve template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyDocRetrieve,
				Name:    "Document Retrieval",
				Value:   "SELECT * FROM docs WHERE id = {{.id}}",
				Enabled: true,
				Edited:  false,
			},
		},
		{
			name: "valid graph retrieve template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyGraphRetrieve,
				Name:    "Graph Retrieval",
				Value:   "MATCH (n) RETURN n",
				Enabled: true,
				Edited:  false,
			},
		},
		{
			name: "valid context organize template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyContextOrganize,
				Name:    "Context Organization",
				Value:   "Organize context: {{.context}}",
				Enabled: true,
				Edited:  false,
			},
		},
		{
			name: "valid related questions template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyRelatedQuestions,
				Name:    "Related Questions",
				Value:   "Generate related questions",
				Enabled: true,
				Edited:  false,
			},
		},
		{
			name: "valid memory retrieve template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyMemoryRetrieve,
				Name:    "Memory Retrieve",
				Value:   "Retrieve from memory",
				Enabled: true,
				Edited:  false,
			},
		},
		{
			name: "valid temp file process template",
			tpl: &DolphinTpl{
				Key:     cdaenum.DolphinTplKeyTempFileProcess,
				Name:    "Temp File Process",
				Value:   "Process temp files",
				Enabled: false,
				Edited:  true,
			},
		},
		{
			name: "valid with empty name",
			tpl: &DolphinTpl{
				Key:   cdaenum.DolphinTplKeyDocRetrieve,
				Name:  "",
				Value: "Template value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.tpl.ValObjCheck()
			assert.NoError(t, err)
		})
	}
}

func TestDolphinTpl_ValObjCheck_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		tpl         *DolphinTpl
		expectedErr string
	}{
		{
			name: "empty key",
			tpl: &DolphinTpl{
				Key:   "",
				Value: "some value",
			},
			expectedErr: "[DolphinTpl]: key is required",
		},
		{
			name: "empty value",
			tpl: &DolphinTpl{
				Key:   cdaenum.DolphinTplKeyDocRetrieve,
				Value: "",
			},
			expectedErr: "[DolphinTpl]: value is required",
		},
		{
			name: "invalid key",
			tpl: &DolphinTpl{
				Key:   "invalid_key",
				Value: "some value",
			},
			expectedErr: "[DolphinTpl]: key is invalid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.tpl.ValObjCheck()
			assert.Error(t, err)
			assert.Contains(t, err.Error(), tt.expectedErr)
		})
	}
}

func TestDolphinTpl_ValObjCheck_Nil(t *testing.T) {
	t.Parallel()

	var tpl *DolphinTpl
	// Nil pointer will panic, so we test for that
	assert.Panics(t, func() {
		tpl.ValObjCheck() //nolint:errcheck
	})
}

func TestDolphinTpl_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	tpl := &DolphinTpl{}
	errMap := tpl.GetErrMsgMap()

	assert.NotNil(t, errMap)
	assert.Len(t, errMap, 2)
	assert.Equal(t, `"key"不能为空`, errMap["Key.required"])
	assert.Equal(t, `"value"不能为空`, errMap["Value.required"])
}

func TestDolphinTpl_Fields(t *testing.T) {
	t.Parallel()

	tpl := &DolphinTpl{
		Key:     cdaenum.DolphinTplKeyDocRetrieve,
		Name:    "Test Template",
		Value:   "Test value",
		Enabled: true,
		Edited:  false,
	}

	assert.Equal(t, cdaenum.DolphinTplKeyDocRetrieve, tpl.Key)
	assert.Equal(t, "Test Template", tpl.Name)
	assert.Equal(t, "Test value", tpl.Value)
	assert.True(t, tpl.Enabled)
	assert.False(t, tpl.Edited)
}
