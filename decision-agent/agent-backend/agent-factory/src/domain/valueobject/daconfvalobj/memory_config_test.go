package daconfvalobj

import "testing"

func TestMemoryCfg_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    *MemoryCfg
	}{
		{
			name: "启用记忆",
			p: &MemoryCfg{
				IsEnabled: true,
			},
		},
		{
			name: "禁用记忆",
			p: &MemoryCfg{
				IsEnabled: false,
			},
		},
		{
			name: "空结构体",
			p:    &MemoryCfg{},
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
