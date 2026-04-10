package agentconfigreq

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/enum/daenum"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/cenum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParams_StructFields(t *testing.T) {
	t.Parallel()

	params := &Params{
		Name:    "TestAgent",
		Profile: "Test profile",
		Skills:  []string{"skill1", "skill2"},
		Sources: []string{"source1", "source2"},
	}

	assert.Equal(t, "TestAgent", params.Name)
	assert.Equal(t, "Test profile", params.Profile)
	assert.Len(t, params.Skills, 2)
	assert.Len(t, params.Sources, 2)
}

func TestParams_Empty(t *testing.T) {
	t.Parallel()

	params := &Params{}

	assert.Empty(t, params.Name)
	assert.Empty(t, params.Profile)
	assert.Nil(t, params.Skills)
	assert.Nil(t, params.Sources)
}

func TestParams_ReqCheck_AllFieldsProvided(t *testing.T) {
	t.Parallel()

	params := &Params{
		Name:    "TestAgent",
		Profile: "Test profile",
		Skills:  []string{"skill1"},
		Sources: []string{"source1"},
	}

	err := params.ReqCheck()

	assert.NoError(t, err)
}

func TestParams_ReqCheck_OnlyName(t *testing.T) {
	t.Parallel()

	params := &Params{
		Name: "TestAgent",
	}

	err := params.ReqCheck()

	assert.NoError(t, err)
}

func TestParams_ReqCheck_OnlyProfile(t *testing.T) {
	t.Parallel()

	params := &Params{
		Profile: "Test profile",
	}

	err := params.ReqCheck()

	assert.NoError(t, err)
}

func TestParams_ReqCheck_OnlySkills(t *testing.T) {
	t.Parallel()

	params := &Params{
		Skills: []string{"skill1", "skill2"},
	}

	err := params.ReqCheck()

	assert.NoError(t, err)
}

func TestParams_ReqCheck_OnlySources(t *testing.T) {
	t.Parallel()

	params := &Params{
		Sources: []string{"source1"},
	}

	err := params.ReqCheck()

	assert.NoError(t, err)
}

func TestParams_ReqCheck_AllEmpty(t *testing.T) {
	t.Parallel()

	params := &Params{}

	err := params.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "至少有一个不能为空")
}

func TestParams_ReqCheck_EmptyArrays(t *testing.T) {
	t.Parallel()

	params := &Params{
		Name:    "",
		Profile: "",
		Skills:  []string{},
		Sources: []string{},
	}

	err := params.ReqCheck()

	assert.Error(t, err)
}

func TestAiAutogenReq_StructFields(t *testing.T) {
	t.Parallel()

	params := &Params{
		Name: "TestAgent",
	}
	req := &AiAutogenReq{
		Language:    "zh-CN",
		Params:      params,
		From:        daenum.AiAutogenFromSystemPrompt,
		Stream:      true,
		UserID:      "user-123",
		AccountType: cenum.AccountTypeUser,
	}

	assert.Equal(t, "zh-CN", req.Language)
	assert.Equal(t, params, req.Params)
	assert.Equal(t, daenum.AiAutogenFromSystemPrompt, req.From)
	assert.True(t, req.Stream)
	assert.Equal(t, "user-123", req.UserID)
	assert.Equal(t, cenum.AccountTypeUser, req.AccountType)
}

func TestAiAutogenReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	assert.Equal(t, "参数对象(params)不能为空", errMsgMap["Params.required"])
	assert.Equal(t, "内容来源类型无效，只能为preset_question、system_prompt或opening_remarks", errMsgMap["From.oneof"])
}

func TestAiAutogenReq_IsNotStream(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		from     daenum.AiAutogenFrom
		expected bool
	}{
		{
			name:     "preset question - not stream",
			from:     daenum.AiAutogenFromPreSetQuestion,
			expected: true,
		},
		{
			name:     "system prompt - is stream",
			from:     daenum.AiAutogenFromSystemPrompt,
			expected: false,
		},
		{
			name:     "opening remarks - is stream",
			from:     daenum.AiAutogenFromOpeningRemarks,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := &AiAutogenReq{
				From: tt.from,
			}

			result := req.IsNotStream()

			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAiAutogenReq_ReqCheck_Valid(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: &Params{
			Name: "TestAgent",
		},
		From: daenum.AiAutogenFromSystemPrompt,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestAiAutogenReq_ReqCheck_NilParams(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: nil,
		From:   daenum.AiAutogenFromSystemPrompt,
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params is required")
}

func TestAiAutogenReq_ReqCheck_InvalidParams(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: &Params{
			// All empty
		},
		From: daenum.AiAutogenFromSystemPrompt,
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "params is invalid")
}

func TestAiAutogenReq_ReqCheck_InvalidFrom(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: &Params{
			Name: "TestAgent",
		},
		From: "invalid_from",
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from is invalid")
}

func TestAiAutogenReq_WithDifferentFromTypes(t *testing.T) {
	t.Parallel()

	fromTypes := []daenum.AiAutogenFrom{
		daenum.AiAutogenFromSystemPrompt,
		daenum.AiAutogenFromOpeningRemarks,
		daenum.AiAutogenFromPreSetQuestion,
	}

	for _, fromType := range fromTypes {
		req := &AiAutogenReq{
			Params: &Params{
				Name: "TestAgent",
			},
			From: fromType,
		}

		err := req.ReqCheck()

		assert.NoError(t, err, "Should validate successfully for from type: %s", fromType)
	}
}

func TestAiAutogenReq_StreamAndNotStream(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: &Params{
			Name: "TestAgent",
		},
		From:   daenum.AiAutogenFromSystemPrompt,
		Stream: true,
	}

	err := req.ReqCheck()
	require.NoError(t, err)

	assert.True(t, req.Stream)
	assert.False(t, req.IsNotStream())
}

func TestAiAutogenReq_WithAccountTypes(t *testing.T) {
	t.Parallel()

	accountTypes := []cenum.AccountType{
		cenum.AccountTypeUser,
		cenum.AccountTypeApp,
		cenum.AccountTypeAnonymous,
	}

	for _, accType := range accountTypes {
		req := &AiAutogenReq{
			Params:      &Params{Name: "TestAgent"},
			From:        daenum.AiAutogenFromSystemPrompt,
			UserID:      "user-test",
			AccountType: accType,
		}

		err := req.ReqCheck()
		assert.NoError(t, err, "Should validate successfully for account type: %s", accType)
		assert.Equal(t, accType, req.AccountType)
	}
}

func TestAiAutogenReq_ComplexParams(t *testing.T) {
	t.Parallel()

	req := &AiAutogenReq{
		Params: &Params{
			Name:    "ComplexAgent",
			Profile: "A complex agent with multiple skills",
			Skills:  []string{"writing", "analysis", "coding"},
			Sources: []string{"database", "api", "files"},
		},
		From:        daenum.AiAutogenFromOpeningRemarks,
		Stream:      false,
		Language:    "en-US",
		UserID:      "user-complex",
		AccountType: cenum.AccountTypeApp,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Len(t, req.Params.Skills, 3)
	assert.Len(t, req.Params.Sources, 3)
}
