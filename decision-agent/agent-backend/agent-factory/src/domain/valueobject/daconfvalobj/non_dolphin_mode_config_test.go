package daconfvalobj

import "testing"

func TestNonDolphinModeConfig_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    *NonDolphinModeConfig
	}{
		{
			name: "全部开启",
			p: &NonDolphinModeConfig{
				DisableHistoryInAConversation: true,
				DisableLLMCache:               true,
			},
		},
		{
			name: "全部关闭",
			p: &NonDolphinModeConfig{
				DisableHistoryInAConversation: false,
				DisableLLMCache:               false,
			},
		},
		{
			name: "空结构体",
			p:    &NonDolphinModeConfig{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.p.ValObjCheck()
			if err != nil {
				t.Errorf("ValObjCheck() error = %v, want nil", err)
			}
		})
	}
}
