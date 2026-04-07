package daconfvalobj

import "testing"

func TestPresetQuestion_ValObjCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		p       *PresetQuestion
		wantErr bool
	}{
		{
			name: "有效问题",
			p: &PresetQuestion{
				Question: "这是一个问题",
			},
			wantErr: false,
		},
		{
			name: "空问题",
			p: &PresetQuestion{
				Question: "",
			},
			wantErr: true,
		},
		{
			name:    "空结构体",
			p:       &PresetQuestion{},
			wantErr: true,
		},
		{
			name: "单字符问题",
			p: &PresetQuestion{
				Question: "?",
			},
			wantErr: false,
		},
		{
			name: "长问题",
			p: &PresetQuestion{
				Question: "这是一个非常长的问题，用于测试边界情况，确保系统能够处理长文本内容",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.p.ValObjCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("ValObjCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
