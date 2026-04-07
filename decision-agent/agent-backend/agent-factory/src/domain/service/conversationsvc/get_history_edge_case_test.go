package conversationsvc

import (
	"testing"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/constant"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/comvalobj"
	"github.com/stretchr/testify/assert"
)

func TestGetHistory_EdgeCaseLogic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		history        []*comvalobj.LLMMessage
		limit          int
		expectedLength int
	}{
		{
			name:           "empty history",
			history:        []*comvalobj.LLMMessage{},
			limit:          4,
			expectedLength: 0,
		},
		{
			name: "limit greater than history length",
			history: []*comvalobj.LLMMessage{
				{Role: "user", Content: "msg1"},
				{Role: "assistant", Content: "msg2"},
			},
			limit:          10,
			expectedLength: 2,
		},
		{
			name: "limit equals history length",
			history: []*comvalobj.LLMMessage{
				{Role: "user", Content: "msg1"},
				{Role: "assistant", Content: "msg2"},
				{Role: "user", Content: "msg3"},
				{Role: "assistant", Content: "msg4"},
			},
			limit:          4,
			expectedLength: 4,
		},
		{
			name: "limit less than history length",
			history: []*comvalobj.LLMMessage{
				{Role: "user", Content: "msg1"},
				{Role: "assistant", Content: "msg2"},
				{Role: "user", Content: "msg3"},
				{Role: "assistant", Content: "msg4"},
				{Role: "user", Content: "msg5"},
				{Role: "assistant", Content: "msg6"},
				{Role: "user", Content: "msg7"},
				{Role: "assistant", Content: "msg8"},
			},
			limit:          2,
			expectedLength: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			limit := tt.limit
			history := tt.history

			if limit == 0 {
				limit = constant.DefaultHistoryLimit
			}

			if len(history) == 0 || limit == -1 {
				assert.Equal(t, tt.expectedLength, len(history))
				return
			}

			if limit >= len(history) {
				assert.Equal(t, tt.expectedLength, len(history))
				return
			}

			result := history[len(history)-limit:]
			assert.Equal(t, tt.expectedLength, len(result))
		})
	}
}
