package funcs

import (
	"context"
	"strings"
	"testing"
)

func TestStr(t *testing.T) {
	s := &str{}
	ctx := context.Background()

	tests := []struct {
		name     string        // 测试用例名称
		method   string        // 调用的方法名
		args     []interface{} // 输入参数
		expected string        // 期望的字符串结果
		wantErr  bool          // 是否期望错误
		errMsg   string        // 期望的错误信息
	}{
		// 基础类型转换测试
		{
			name:     "空参数返回空字符串",
			method:   "str",
			args:     []interface{}{},
			expected: "",
		},
		{
			name:     "nil参数转换为空字符串",
			method:   "str",
			args:     []interface{}{nil},
			expected: "",
		},
		{
			name:     "整数转换",
			method:   "str",
			args:     []interface{}{42},
			expected: "42",
		},
		{
			name:     "浮点数转换",
			method:   "str",
			args:     []interface{}{3.14},
			expected: "3.14",
		},
		{
			name:     "布尔值转换",
			method:   "str",
			args:     []interface{}{true, false},
			expected: "truefalse",
		},
		{
			name:     "字符串直接拼接",
			method:   "str",
			args:     []interface{}{"Hello", " ", "World"},
			expected: "Hello World",
		},

		// 复合类型转换测试
		{
			name:     "数组转换为JSON字符串",
			method:   "str",
			args:     []interface{}{[]int{1, 2, 3}},
			expected: "[1,2,3]",
		},
		{
			name:     "切片转换为JSON字符串",
			method:   "str",
			args:     []interface{}{[]string{"a", "b", "c"}},
			expected: `["a","b","c"]`,
		},
		{
			name:     "map转换为JSON字符串",
			method:   "str",
			args:     []interface{}{map[string]int{"one": 1, "two": 2}},
			expected: `{"one":1,"two":2}`,
		},
		{
			name:     "混合类型参数",
			method:   "str",
			args:     []interface{}{"ID:", 123, ", Active:", true},
			expected: "ID:123, Active:true",
		},

		// 错误情况测试
		{
			name:    "不支持的方法名",
			method:  "toString",
			args:    []interface{}{"test"},
			wantErr: true,
			errMsg:  "unknown method: toString",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rets, err := s.Call(ctx, tt.method, 1, tt.args...)

			// 检查错误情况
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误, 但实际没有错误")
				}
				if !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("期望错误信息包含: %q, 实际得到: %q", tt.errMsg, err.Error())
				}
				return
			}

			// 检查正常情况
			if err != nil {
				t.Fatalf("不期望错误但得到: %v", err)
			}

			if len(rets) != 1 {
				t.Fatalf("期望返回1个结果, 实际得到 %d 个", len(rets))
			}

			result, ok := rets[0].(string)
			if !ok {
				t.Fatalf("返回结果不是string类型: %T", rets[0])
			}

			if result != tt.expected {
				t.Errorf("期望结果 %q, 实际得到 %q", tt.expected, result)
			}
		})
	}
}
