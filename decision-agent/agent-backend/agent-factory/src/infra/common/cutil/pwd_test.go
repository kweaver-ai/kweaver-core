package cutil

import "testing"

func TestCheckPassword(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		password string
		want     bool
	}{
		{
			name:     "有效密码-包含所有必需字符",
			password: "TestPass123@",
			want:     true,
		},
		{
			name:     "无效密码-长度太短",
			password: "Abc123@",
			want:     false,
		},
		{
			name:     "无效密码-长度太长",
			password: "A1b" + string(make([]byte, 98)) + "@",
			want:     false,
		},
		{
			name:     "无效密码-缺少大写字母",
			password: "testpass123@",
			want:     false,
		},
		{
			name:     "无效密码-缺少小写字母",
			password: "TESTPASS123@",
			want:     false,
		},
		{
			name:     "无效密码-缺少数字",
			password: "TestPassword@",
			want:     false,
		},
		{
			name:     "无效密码-包含不允许的特殊字符",
			password: "TestPass123*&^",
			want:     false,
		},
		{
			name:     "有效密码-包含所有允许的特殊字符",
			password: "TestPass123~!%#$@-_.",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := CheckPassword(tt.password); got != tt.want {
				t.Errorf("CheckPassword() = %v, want %v", got, tt.want)
			}
		})
	}
}
