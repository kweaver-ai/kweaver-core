package cdaenum

import "testing"

func TestOutputDefaultFormat_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		o       OutputDefaultFormat
		wantErr bool
	}{
		{
			name:    "JSON格式",
			o:       OutputDefaultFormatJson,
			wantErr: false,
		},
		{
			name:    "Markdown格式",
			o:       OutputDefaultFormatMarkdown,
			wantErr: false,
		},
		{
			name:    "无效格式",
			o:       OutputDefaultFormat("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			o:       OutputDefaultFormat(""),
			wantErr: true,
		},
		{
			name:    "未知格式",
			o:       OutputDefaultFormat("html"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.o.EnumCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
