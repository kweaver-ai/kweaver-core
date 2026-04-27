package agentconfigvo

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/cdaenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/valueobject/daconfvalobj"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewOutputVariablesS(t *testing.T) {
	t.Parallel()

	vars := NewOutputVariablesS()
	assert.NotNil(t, vars)
	assert.IsType(t, &OutputVariablesS{}, vars)
}

func TestExtractOutputFromLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "包含->和output_",
			line: " -> output_test",
			want: "output_test",
		},
		{
			name: "包含>>和output_",
			line: ">> output_value",
			want: "output_value",
		},
		{
			name: "不包含output_",
			line: " some text",
			want: "",
		},
		{
			name: "空行",
			line: "",
			want: "",
		},
		{
			name: "只有->没有output_",
			line: " -> something",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractOutputFromLine(tt.line)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestOutputVariablesS_LoadFromConfig_PanicsWithNilConfig(t *testing.T) {
	t.Parallel()

	v := &OutputVariablesS{}

	assert.Panics(t, func() {
		_ = v.LoadFromConfig(nil)
	})
}

func TestOutputVariablesS_LoadFromConfig_PanicsWithNilOutput(t *testing.T) {
	t.Parallel()

	v := &OutputVariablesS{}

	config := &daconfvalobj.Config{
		Output: nil,
	}

	assert.Panics(t, func() {
		_ = v.LoadFromConfig(config)
	})
}

func TestExtractOutputsFromText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		text      string
		wantCount int
	}{
		{
			name:      "单个output",
			text:      " -> output_test",
			wantCount: 1,
		},
		{
			name:      "多个output",
			text:      " -> output_1\n -> output_2\n>> output_3",
			wantCount: 3,
		},
		{
			name:      "混合内容",
			text:      "some text\n -> output_test\nmore text",
			wantCount: 1,
		},
		{
			name:      "空文本",
			text:      "",
			wantCount: 0,
		},
		{
			name:      "没有output",
			text:      "just some text",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputs, err := ExtractOutputsFromText(tt.text)
			assert.NoError(t, err)
			assert.Len(t, outputs, tt.wantCount)
		})
	}
}

func TestExtractOutputsFromText_Errors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
	}{
		{
			name: "空白符",
			text: "   ",
		},
		{
			name: "多行空白符",
			text: "  \n  \n  ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outputs, err := ExtractOutputsFromText(tt.text)
			assert.NoError(t, err)
			assert.Len(t, outputs, 0)
		})
	}
}

func TestOutputVariablesS_LoadFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  *daconfvalobj.Config
		wantErr bool
		check   func(t *testing.T, v *OutputVariablesS)
	}{
		{
			name: "load from config with output variables",
			config: &daconfvalobj.Config{
				Output: &daconfvalobj.Output{
					Variables: &daconfvalobj.VariablesS{
						AnswerVar:           "answer",
						DocRetrievalVar:     "doc_res",
						GraphRetrievalVar:   "graph_res",
						RelatedQuestionsVar: "questions",
						OtherVars:           []string{"var1", "var2"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, v *OutputVariablesS) {
				assert.Equal(t, "answer", v.AnswerVar)
				assert.Equal(t, "doc_res", v.DocRetrievalVar)
				assert.Equal(t, []string{"var1", "var2"}, v.OtherVars)
			},
		},
		{
			name: "load from config with dolphin mode",
			config: &daconfvalobj.Config{
				IsDolphinMode: cdaenum.DolphinModeEnabled,
				Dolphin:       " -> output_test1\n -> output_test2\nsome text\n -> output_test3",
				Output: &daconfvalobj.Output{
					Variables: &daconfvalobj.VariablesS{
						AnswerVar: "answer",
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, v *OutputVariablesS) {
				assert.Equal(t, "answer", v.AnswerVar)
				assert.Len(t, v.MiddleOutputVars, 3)
				assert.Contains(t, v.MiddleOutputVars, "output_test1")
				assert.Contains(t, v.MiddleOutputVars, "output_test2")
				assert.Contains(t, v.MiddleOutputVars, "output_test3")
			},
		},
		{
			name: "load from config without dolphin mode",
			config: &daconfvalobj.Config{
				IsDolphinMode: cdaenum.DolphinModeDisabled,
				Dolphin:       " -> output_test1",
				Output: &daconfvalobj.Output{
					Variables: &daconfvalobj.VariablesS{
						AnswerVar: "answer",
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, v *OutputVariablesS) {
				assert.Equal(t, "answer", v.AnswerVar)
				assert.Empty(t, v.MiddleOutputVars)
			},
		},
		{
			name: "load from config with empty middle output vars",
			config: &daconfvalobj.Config{
				Output: &daconfvalobj.Output{
					Variables: &daconfvalobj.VariablesS{
						AnswerVar:        "answer",
						MiddleOutputVars: []string{},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, v *OutputVariablesS) {
				assert.Equal(t, "answer", v.AnswerVar)
				assert.Empty(t, v.MiddleOutputVars)
			},
		},
		{
			name: "load from config with existing middle output vars (should not extract)",
			config: &daconfvalobj.Config{
				IsDolphinMode: cdaenum.DolphinModeEnabled,
				Dolphin:       " -> output_test1\n -> output_test2",
				Output: &daconfvalobj.Output{
					Variables: &daconfvalobj.VariablesS{
						AnswerVar:        "answer",
						MiddleOutputVars: []string{"existing_var1", "existing_var2"},
					},
				},
			},
			wantErr: false,
			check: func(t *testing.T, v *OutputVariablesS) {
				assert.Equal(t, "answer", v.AnswerVar)
				// MiddleOutputVars should remain as is, not extracted from dolphin
				assert.Len(t, v.MiddleOutputVars, 2)
				assert.Contains(t, v.MiddleOutputVars, "existing_var1")
				assert.Contains(t, v.MiddleOutputVars, "existing_var2")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			v := &OutputVariablesS{}

			err := v.LoadFromConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)

				if tt.check != nil {
					tt.check(t, v)
				}
			}
		})
	}
}

func TestOutputVariablesS_ToVariable(t *testing.T) {
	t.Parallel()

	v := &OutputVariablesS{
		AnswerVar:           "answer",
		DocRetrievalVar:     "doc_res",
		GraphRetrievalVar:   "graph_res",
		RelatedQuestionsVar: "questions",
		OtherVars:           []string{"var1", "var2"},
		MiddleOutputVars:    []string{"mid1", "mid2"},
	}

	variable, err := v.ToVariable()
	require.NoError(t, err)
	assert.NotNil(t, variable)
	assert.Equal(t, "answer", variable.AnswerVar)
	assert.Equal(t, "doc_res", variable.DocRetrievalVar)
	assert.Equal(t, []string{"var1", "var2"}, variable.OtherVars)
	assert.Equal(t, []string{"mid1", "mid2"}, variable.MiddleOutputVars)
}

func TestOutputVariablesS_ToVariable_Empty(t *testing.T) {
	t.Parallel()

	v := &OutputVariablesS{}

	variable, err := v.ToVariable()
	require.NoError(t, err)
	assert.NotNil(t, variable)
	assert.Empty(t, variable.AnswerVar)
	assert.Nil(t, variable.OtherVars)
}

func TestOutputVariablesS_LoadFromConfig_EmptyDolphinWithMode(t *testing.T) {
	t.Parallel()

	// Test when dolphin is empty but mode is enabled
	config := &daconfvalobj.Config{
		IsDolphinMode: cdaenum.DolphinModeEnabled,
		Dolphin:       "", // Empty dolphin
		Output: &daconfvalobj.Output{
			Variables: &daconfvalobj.VariablesS{
				AnswerVar: "answer",
			},
		},
	}

	v := &OutputVariablesS{}
	err := v.LoadFromConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, "answer", v.AnswerVar)
	// MiddleOutputVars should remain empty since dolphin is empty
	assert.Empty(t, v.MiddleOutputVars)
}

func TestExtractOutputFromLine_ComplexPatterns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		line string
		want string
	}{
		{
			name: "混合->和>>",
			line: "text -> >> output_test",
			want: "output_test", // 第一个匹配的->之后的内容
		},
		{
			name: "多个output_",
			line: " -> output_1 >> output_2",
			want: "output_1", // 第一个匹配的
		},
		{
			name: "output_带数字",
			line: ">> output_123",
			want: "output_123",
		},
		{
			name: "output_带下划线",
			line: " -> output_test_value",
			want: "output_test_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extractOutputFromLine(tt.line)
			assert.Equal(t, tt.want, got)
		})
	}
}
