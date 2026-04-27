package persrecvalid

import (
	"testing"

	"github.com/go-playground/validator/v10"
)

func TestCheckName(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	validate.RegisterValidation("checkName", CheckName) //nolint:errcheck

	type TestStruct struct {
		Name string `validate:"checkName"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "空字符串",
			value:   "",
			wantErr: false,
		},
		{
			name:    "有效名称",
			value:   "ValidName123",
			wantErr: false,
		},
		{
			name:    "包含特殊字符",
			value:   "Invalid?Name",
			wantErr: true,
		},
		{
			name:    "包含斜杠",
			value:   "Invalid/Name",
			wantErr: true,
		},
		{
			name:    "过长名称",
			value:   "WayTooLongNameWayTooLongNameWayTooLongNameWayTooLongNameWayTooLongName",
			wantErr: true,
		},
		{
			name:    "开头有空格",
			value:   " LeadingSpace",
			wantErr: true,
		},
		{
			name:    "结尾有空格",
			value:   "TrailingSpace ",
			wantErr: true,
		},
		{
			name:    "有效的带空格名称",
			value:   "Valid Name",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := TestStruct{Name: tt.value}

			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckName(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestCheckCode(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	validate.RegisterValidation("checkCode", CheckCode) //nolint:errcheck

	type TestStruct struct {
		Code string `validate:"checkCode"`
	}

	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:    "空字符串",
			value:   "",
			wantErr: false,
		},
		{
			name:    "有效代码",
			value:   "code123",
			wantErr: false,
		},
		{
			name:    "包含特殊字符",
			value:   "code#123",
			wantErr: true,
		},
		{
			name:    "包含空格",
			value:   "code 123",
			wantErr: true,
		},
		{
			name:    "过长代码",
			value:   "WayTooLongCodeWayTooLongCodeWayTooLongCodeWayTooLongCodeWayTooLongCode",
			wantErr: true,
		},
		{
			name:    "有效的包含@符号",
			value:   "code@test",
			wantErr: false,
		},
		{
			name:    "有效的包含-符号",
			value:   "code-test",
			wantErr: false,
		},
		{
			name:    "有效的包含.符号",
			value:   "code.test",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := TestStruct{Code: tt.value}

			err := validate.Struct(s)
			if (err != nil) != tt.wantErr {
				t.Errorf("CheckCode(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestGenNameErrMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "基本前缀",
			target: "前缀",
			want:   "前缀" + NameErrMsg,
		},
		{
			name:   "空前缀",
			target: "",
			want:   NameErrMsg,
		},
		{
			name:   "中英文混合",
			target: "姓名",
			want:   "姓名" + NameErrMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GenNameErrMsg(tt.target)
			if got != tt.want {
				t.Errorf("GenNameErrMsg(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}

func TestGenCodeErrMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		want   string
	}{
		{
			name:   "基本前缀",
			target: "前缀",
			want:   "前缀" + CodeErrMsg,
		},
		{
			name:   "空前缀",
			target: "",
			want:   CodeErrMsg,
		},
		{
			name:   "中英文混合",
			target: "代码",
			want:   "代码" + CodeErrMsg,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := GenCodeErrMsg(tt.target)
			if got != tt.want {
				t.Errorf("GenCodeErrMsg(%q) = %q, want %q", tt.target, got, tt.want)
			}
		})
	}
}
