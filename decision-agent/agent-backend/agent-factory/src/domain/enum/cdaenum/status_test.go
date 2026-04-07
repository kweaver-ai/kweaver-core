package cdaenum

import "testing"

func TestStatus_EnumCheck(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		t       Status
		wantErr bool
	}{
		{
			name:    "未发布",
			t:       StatusUnpublished,
			wantErr: false,
		},
		{
			name:    "已发布",
			t:       StatusPublished,
			wantErr: false,
		},
		{
			name:    "无效状态",
			t:       Status("invalid"),
			wantErr: true,
		},
		{
			name:    "空字符串",
			t:       Status(""),
			wantErr: true,
		},
		{
			name:    "未知状态",
			t:       Status("draft"),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.t.EnumCheck()
			if (err != nil) != tt.wantErr {
				t.Errorf("EnumCheck() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestStatus_IsPublished(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		t    Status
		want bool
	}{
		{
			name: "已发布",
			t:    StatusPublished,
			want: true,
		},
		{
			name: "未发布",
			t:    StatusUnpublished,
			want: false,
		},
		{
			name: "其他状态",
			t:    Status("draft"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.t.IsPublished(); got != tt.want {
				t.Errorf("IsPublished() = %v, want %v", got, tt.want)
			}
		})
	}
}
