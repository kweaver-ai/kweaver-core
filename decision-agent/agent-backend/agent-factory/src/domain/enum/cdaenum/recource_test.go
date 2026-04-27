package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceType_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		rt      ResourceType
		wantErr bool
	}{
		{
			name:    "valid data agent",
			rt:      ResourceTypeDataAgent,
			wantErr: false,
		},
		{
			name:    "valid data agent tpl",
			rt:      ResourceTypeDataAgentTpl,
			wantErr: false,
		},
		{
			name:    "invalid resource type",
			rt:      ResourceType("invalid"),
			wantErr: true,
		},
		{
			name:    "empty resource type",
			rt:      ResourceType(""),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.rt.EnumCheck()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResourceType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rt       ResourceType
		expected string
	}{
		{
			name:     "data agent string",
			rt:       ResourceTypeDataAgent,
			expected: "agent",
		},
		{
			name:     "data agent tpl string",
			rt:       ResourceTypeDataAgentTpl,
			expected: "agent_tpl",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.rt.String()
			assert.Equal(t, tt.expected, got)
		})
	}
}
