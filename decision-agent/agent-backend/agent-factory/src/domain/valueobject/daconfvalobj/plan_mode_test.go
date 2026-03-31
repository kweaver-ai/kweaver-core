package daconfvalobj

import "testing"

func TestNewPlanMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		isEnabled     bool
		wantIsEnabled bool
	}{
		{
			name:          "启用",
			isEnabled:     true,
			wantIsEnabled: true,
		},
		{
			name:          "禁用",
			isEnabled:     false,
			wantIsEnabled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := NewPlanMode(tt.isEnabled)
			if got.IsEnabled != tt.wantIsEnabled {
				t.Errorf("NewPlanMode() IsEnabled = %v, want %v", got.IsEnabled, tt.wantIsEnabled)
			}
		})
	}
}

func TestPlanMode_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		p    *PlanMode
	}{
		{
			name: "启用模式",
			p: &PlanMode{
				IsEnabled: true,
			},
		},
		{
			name: "禁用模式",
			p: &PlanMode{
				IsEnabled: false,
			},
		},
		{
			name: "空结构体",
			p:    &PlanMode{},
		},
		{
			name: "nil指针",
			p:    nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if tt.p == nil {
				t.Skip("nil指针不需要测试")
			}

			err := tt.p.ValObjCheck()
			if err != nil {
				t.Errorf("ValObjCheck() error = %v, want nil", err)
			}
		})
	}
}
