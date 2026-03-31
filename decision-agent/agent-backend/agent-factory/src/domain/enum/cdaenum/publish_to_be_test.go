package cdaenum

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPublishToBe_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       PublishToBe
		wantErr bool
	}{
		{
			name:    "API Agent",
			t:       PublishToBeAPIAgent,
			wantErr: false,
		},
		{
			name:    "Web SDK Agent",
			t:       PublishToBeWebSDKAgent,
			wantErr: false,
		},
		{
			name:    "Skill Agent",
			t:       PublishToBeSkillAgent,
			wantErr: false,
		},
		{
			name:    "DataFlow Agent",
			t:       PublishToBeDataFlowAgent,
			wantErr: false,
		},
		{
			name:    "无效值",
			t:       PublishToBe("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       PublishToBe(""),
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
