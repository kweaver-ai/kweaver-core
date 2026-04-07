package daconfvalobj

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/enum/cdaenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOutput_ValObjCheck_DolphinMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		output      *Output
		expectError bool
	}{
		{
			name: "valid output with all fields set",
			output: &Output{
				Variables: &VariablesS{
					AnswerVar:           "answer",
					DocRetrievalVar:     "doc_res",
					GraphRetrievalVar:   "graph_res",
					RelatedQuestionsVar: "questions",
				},
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: false,
		},
		{
			name: "valid output with default values - dolphin mode",
			output: &Output{
				Variables: &VariablesS{
					AnswerVar: "answer",
				},
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: false,
		},
		{
			name: "invalid - missing answer_var in dolphin mode",
			output: &Output{
				Variables: &VariablesS{
					AnswerVar: "",
				},
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: true,
		},
		{
			name: "invalid - invalid default format",
			output: &Output{
				Variables: &VariablesS{
					AnswerVar: "answer",
				},
				DefaultFormat: "invalid_format",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.output.ValObjCheck(true) // dolphin mode
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEqual(t, "", tt.output.Variables.AnswerVar)
				assert.NotEqual(t, "", tt.output.Variables.DocRetrievalVar)
				assert.NotEqual(t, "", tt.output.Variables.GraphRetrievalVar)
				assert.NotEqual(t, "", tt.output.Variables.RelatedQuestionsVar)
			}
		})
	}
}

func TestOutput_ValObjCheck_NonDolphinMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		output      *Output
		expectError bool
	}{
		{
			name: "valid output - sets default answer var",
			output: &Output{
				Variables:     &VariablesS{},
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: false,
		},
		{
			name: "valid output - nil variables",
			output: &Output{
				Variables:     nil,
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: false,
		},
		{
			name: "valid with all fields set",
			output: &Output{
				Variables: &VariablesS{
					AnswerVar:           "custom_answer",
					DocRetrievalVar:     "custom_doc",
					GraphRetrievalVar:   "custom_graph",
					RelatedQuestionsVar: "custom_questions",
				},
				DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
			},
			expectError: false,
		},
		{
			name: "invalid - invalid format",
			output: &Output{
				Variables:     &VariablesS{},
				DefaultFormat: "invalid",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.output.ValObjCheck(false) // non-dolphin mode
			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, tt.output.Variables)
			}
		})
	}
}

func TestOutput_ValObjCheck_DefaultValues(t *testing.T) {
	t.Parallel()

	output := &Output{
		Variables:     &VariablesS{},
		DefaultFormat: cdaenum.OutputDefaultFormatJson,
	}

	err := output.ValObjCheck(false)
	require.NoError(t, err)

	assert.Equal(t, "answer", output.Variables.AnswerVar)
	assert.Equal(t, "doc_retrieval_res", output.Variables.DocRetrievalVar)
	assert.Equal(t, "graph_retrieval_res", output.Variables.GraphRetrievalVar)
	assert.Equal(t, "related_questions", output.Variables.RelatedQuestionsVar)
}

func TestVariablesS_Fields(t *testing.T) {
	t.Parallel()

	vars := &VariablesS{
		AnswerVar:           "answer",
		DocRetrievalVar:     "doc",
		GraphRetrievalVar:   "graph",
		RelatedQuestionsVar: "questions",
		OtherVars:           []string{"var1", "var2"},
		MiddleOutputVars:    []string{"mid1", "mid2"},
	}

	assert.Equal(t, "answer", vars.AnswerVar)
	assert.Equal(t, "doc", vars.DocRetrievalVar)
	assert.Equal(t, "graph", vars.GraphRetrievalVar)
	assert.Equal(t, "questions", vars.RelatedQuestionsVar)
	assert.Len(t, vars.OtherVars, 2)
	assert.Len(t, vars.MiddleOutputVars, 2)
}

func TestOutput_Fields(t *testing.T) {
	t.Parallel()

	vars := &VariablesS{
		AnswerVar: "ans",
	}

	output := &Output{
		Variables:     vars,
		DefaultFormat: cdaenum.OutputDefaultFormatMarkdown,
	}

	assert.Equal(t, vars, output.Variables)
	assert.Equal(t, cdaenum.OutputDefaultFormatMarkdown, output.DefaultFormat)
}
