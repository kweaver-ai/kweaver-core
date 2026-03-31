package funcs

import (
	"context"
	"testing"
)

func TestMath(t *testing.T) {
	math := &Math{}
	ctx := context.Background()

	tests := []struct {
		name     string        // 测试用例名称
		op       string        // 操作名称
		args     []interface{} // 输入参数
		expected float64       // 期望结果
		wantErr  bool          // 是否期望错误
		errMsg   string        // 期望的错误信息(如果有)
	}{
		// 加法测试用例
		{
			name:     "加法-两个正数",
			op:       "add",
			args:     []interface{}{2.5, 3.5},
			expected: 6.0,
		},
		{
			name:     "加法-多个数字",
			op:       "add",
			args:     []interface{}{1, 2, 3, 4},
			expected: 10.0,
		},
		{
			name:     "加法-零值",
			op:       "add",
			args:     []interface{}{0, 0},
			expected: 0.0,
		},

		// 减法测试用例
		{
			name:     "减法-两个数字",
			op:       "sub",
			args:     []interface{}{10.5, 2.5},
			expected: 8.0,
		},
		{
			name:     "减法-多个数字",
			op:       "sub",
			args:     []interface{}{20, 5, 3},
			expected: 12.0,
		},
		{
			name:     "减法-无参数返回0",
			op:       "sub",
			args:     []interface{}{},
			expected: 0.0,
		},

		// 乘法测试用例
		{
			name:     "乘法-两个数字",
			op:       "mul",
			args:     []interface{}{3.0, 4.0},
			expected: 12.0,
		},
		{
			name:     "乘法-包含零值",
			op:       "mul",
			args:     []interface{}{5, 0, 10},
			expected: 0.0,
		},

		// 除法测试用例
		{
			name:     "除法-两个数字",
			op:       "div",
			args:     []interface{}{10.0, 2.0},
			expected: 5.0,
		},
		{
			name:     "除法-多个数字",
			op:       "div",
			args:     []interface{}{100.0, 2.0, 5.0},
			expected: 10.0,
		},
		{
			name:     "除法-无参数返回0",
			op:       "div",
			args:     []interface{}{},
			expected: 0.0,
		},
		{
			name:    "除法-除零错误",
			op:      "div",
			args:    []interface{}{10.0, 0.0},
			wantErr: true,
			errMsg:  "division by zero",
		},

		// 错误处理测试用例
		{
			name:    "无效操作",
			op:      "mod", // 不支持的模运算
			args:    []interface{}{10, 3},
			wantErr: true,
			errMsg:  "unsupported operation: mod",
		},
		{
			name:    "无效参数类型",
			op:      "add",
			args:    []interface{}{"not-a-number", 2},
			wantErr: true,
			errMsg:  "invalid argument 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, rets, err := math.Call(ctx, tt.op, 1, tt.args...)

			// 检查错误情况
			if tt.wantErr {
				if err == nil {
					t.Fatal("期望返回错误, 但实际没有错误")
				}
				if tt.errMsg != "" && err.Error()[:len(tt.errMsg)] != tt.errMsg {
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

			result, ok := rets[0].(float64)
			if !ok {
				t.Fatalf("返回结果不是float64类型: %T", rets[0])
			}

			if result != tt.expected {
				t.Errorf("期望结果 %v, 实际得到 %v", tt.expected, result)
			}
		})
	}
}
