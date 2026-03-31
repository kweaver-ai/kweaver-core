package agentsvc

import (
	"testing"

	agentreq "github.com/kweaver-ai/decision-agent/agent-factory/src/driveradapter/api/rdto/agent/req"
	"github.com/stretchr/testify/assert"
)

func TestIsNormalChat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *agentreq.ChatReq
		want bool
	}{
		{
			name: "all empty IDs returns true",
			req: &agentreq.ChatReq{
				RegenerateAssistantMsgID:  "",
				InterruptedAssistantMsgID: "",
				RegenerateUserMsgID:       "",
			},
			want: true,
		},
		{
			name: "regenerate assistant msg ID returns false",
			req: &agentreq.ChatReq{
				RegenerateAssistantMsgID:  "msg-123",
				InterruptedAssistantMsgID: "",
				RegenerateUserMsgID:       "",
			},
			want: false,
		},
		{
			name: "interrupted assistant msg ID returns false",
			req: &agentreq.ChatReq{
				RegenerateAssistantMsgID:  "",
				InterruptedAssistantMsgID: "msg-456",
				RegenerateUserMsgID:       "",
			},
			want: false,
		},
		{
			name: "regenerate user msg ID returns false",
			req: &agentreq.ChatReq{
				RegenerateAssistantMsgID:  "",
				InterruptedAssistantMsgID: "",
				RegenerateUserMsgID:       "msg-789",
			},
			want: false,
		},
		{
			name: "multiple IDs set returns false",
			req: &agentreq.ChatReq{
				RegenerateAssistantMsgID:  "msg-123",
				InterruptedAssistantMsgID: "msg-456",
				RegenerateUserMsgID:       "msg-789",
			},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := IsNormalChat(tt.req)
			assert.Equal(t, tt.want, result)
		})
	}
}
