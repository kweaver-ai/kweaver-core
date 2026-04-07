package customvalidator

import (
	"testing"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
)

// TestCheckAgentAndTplName_WithValidator 使用 validator.Struct 测试实际的 CheckAgentAndTplName 函数
// 与 agent_and_tpl_name_test.go 中直接测试 regex 不同，这里测试通过 validator 注册后的完整调用路径
func TestCheckAgentAndTplName_WithValidator(t *testing.T) {
	t.Parallel()

	validate := validator.New()
	_ = validate.RegisterValidation("checkAgentAndTplName", CheckAgentAndTplName)

	type testStruct struct {
		Name string `validate:"checkAgentAndTplName"`
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"empty string allowed", "", false},
		{"chinese chars", "测试名称", false},
		{"english chars", "testName", false},
		{"underscore start", "_test", false},
		{"with digits", "test123", false},
		{"digit start not allowed", "1test", true},
		{"special chars not allowed", "test@name", true},
		{"space not allowed", "test name", true},
		{"dash not allowed", "test-name", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := testStruct{Name: tt.input}
			err := validate.Struct(s)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
