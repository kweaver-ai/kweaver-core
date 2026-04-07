package agentconfigreq

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCopyReq_StructFields(t *testing.T) {
	t.Parallel()

	req := CopyReq{
		Name: "TestAgentCopy",
	}

	assert.Equal(t, "TestAgentCopy", req.Name)
}

func TestCopyReq_EmptyName(t *testing.T) {
	t.Parallel()

	req := CopyReq{}

	assert.Empty(t, req.Name)
}

func TestCopyReq_GetErrMsgMap(t *testing.T) {
	t.Parallel()

	req := CopyReq{}

	errMsgMap := req.GetErrMsgMap()

	assert.NotNil(t, errMsgMap)
	// Name is optional, so error map should be empty
	assert.Empty(t, errMsgMap)
}

func TestCopyReq_ReqCheck_ValidName(t *testing.T) {
	t.Parallel()

	req := CopyReq{
		Name: "ValidAgentName",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopyReq_ReqCheck_EmptyName(t *testing.T) {
	t.Parallel()

	req := CopyReq{
		Name: "",
	}

	err := req.ReqCheck()

	assert.NoError(t, err) // Empty name is valid (will be auto-generated)
}

func TestCopyReq_ReqCheck_NameTooLong(t *testing.T) {
	t.Parallel()

	// Create a name that exceeds the max length
	longName := strings.Repeat("a", 51) // Max is 50
	req := CopyReq{
		Name: longName,
	}

	err := req.ReqCheck()

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "agent名称长度不能超过")
}

func TestCopyReq_ReqCheck_NameAtMaxLength(t *testing.T) {
	t.Parallel()

	// Create a name that is exactly at the max length
	maxName := strings.Repeat("a", 50) // Max is 50
	req := CopyReq{
		Name: maxName,
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopyReq_ReqCheck_NameWithSpecialCharacters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "name with spaces",
			input:   "My Agent Name",
			wantErr: false,
		},
		{
			name:    "name with numbers",
			input:   "Agent123",
			wantErr: false,
		},
		{
			name:    "name with hyphens",
			input:   "my-agent-name",
			wantErr: false,
		},
		{
			name:    "name with underscores",
			input:   "my_agent_name",
			wantErr: false,
		},
		{
			name:    "name with Chinese characters",
			input:   "我的智能助手",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := CopyReq{
				Name: tt.input,
			}

			err := req.ReqCheck()

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCopyReq_ReqCheck_UnicodeLength(t *testing.T) {
	t.Parallel()

	// Test that rune length is calculated correctly for Unicode characters
	req := CopyReq{
		Name: "我的智能助手", // 6 Chinese characters
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
}

func TestCopyReq_WithEmptyString(t *testing.T) {
	t.Parallel()

	req := CopyReq{
		Name: "",
	}

	err := req.ReqCheck()

	assert.NoError(t, err)
	assert.Empty(t, req.Name)
}
