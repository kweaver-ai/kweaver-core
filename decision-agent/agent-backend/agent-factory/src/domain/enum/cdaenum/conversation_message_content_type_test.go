package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConversationMsgContentType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       ConversationMsgContentType
		wantErr bool
	}{
		{
			name:    "Text类型",
			t:       MsgText,
			wantErr: false,
		},
		{
			name:    "Explore类型",
			t:       ExploreMsgType,
			wantErr: false,
		},
		{
			name:    "Prompt类型",
			t:       PromptMsgType,
			wantErr: false,
		},
		{
			name:    "Other类型",
			t:       OtherMsgType,
			wantErr: false,
		},
		{
			name:    "无效类型",
			t:       ConversationMsgContentType("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       ConversationMsgContentType(""),
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
