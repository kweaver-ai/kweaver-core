package v3agentconfigsvc

import (
	"testing"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/domain/service"
	"github.com/stretchr/testify/assert"
)

func TestDataAgentConfigSvc_CheckPreSetQuestionResFormat(t *testing.T) {
	t.Parallel()

	svc := &dataAgentConfigSvc{SvcBase: service.NewSvcBase()}

	tests := []struct {
		name      string
		content   string
		wantOk    bool
		wantCount int
	}{
		{
			name:      "valid JSON array",
			content:   `["问题1","问题2","问题3"]`,
			wantOk:    true,
			wantCount: 3,
		},
		{
			name:    "empty array",
			content: `[]`,
			wantOk:  false,
		},
		{
			name:    "invalid JSON",
			content: `not json`,
			wantOk:  false,
		},
		{
			name:    "empty string",
			content: ``,
			wantOk:  false,
		},
		{
			name:    "JSON object not array",
			content: `{"key":"value"}`,
			wantOk:  false,
		},
		{
			name:      "single question",
			content:   `["only one question"]`,
			wantOk:    true,
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			questions, ok := svc.checkPreSetQuestionResFormat(tt.content)
			assert.Equal(t, tt.wantOk, ok)

			if tt.wantOk {
				assert.Len(t, questions, tt.wantCount)
			}
		})
	}
}
