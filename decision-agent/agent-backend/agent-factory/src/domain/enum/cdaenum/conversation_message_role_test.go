package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationMsgRole_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       ConversationMsgRole
		wantErr bool
	}{
		{
			name:    "用户角色",
			t:       MsgRoleUser,
			wantErr: false,
		},
		{
			name:    "助手角色",
			t:       MsgRoleAssistant,
			wantErr: false,
		},
		{
			name:    "无效角色",
			t:       ConversationMsgRole("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       ConversationMsgRole(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.t.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err, "expected error")
			} else {
				assert.NoError(t, err, "expected no error")
			}
		})
	}
}
