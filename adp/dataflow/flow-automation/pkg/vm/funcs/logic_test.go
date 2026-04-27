package funcs

import (
	"context"
	"testing"
)

func TestLogicCall(t *testing.T) {
	l := &Logic{}
	ctx := context.Background()

	tests := []struct {
		name     string        // 测试用例名称
		op       string        // 操作类型
		args     []interface{} // 输入参数
		expected bool          // 期望结果
	}{
		// AND 操作测试
		{"AND空参数列表", "and", []interface{}{}, true},
		{"AND全真参数", "and", []interface{}{true, 1, "text"}, true},
		{"AND包含假值参数", "and", []interface{}{true, false, 1}, false},
		{"AND全假参数", "and", []interface{}{false, 0, ""}, false},

		// OR 操作测试
		{"OR空参数列表", "or", []interface{}{}, false},
		{"OR全真参数", "or", []interface{}{true, 1, "text"}, true},
		{"OR包含真值参数", "or", []interface{}{false, 0, "text"}, true},
		{"OR全假参数", "or", []interface{}{false, 0, ""}, false},

		// NOT 操作测试
		{"NOT无参数", "not", []interface{}{}, true},
		{"NOT真值参数", "not", []interface{}{true}, false},
		{"NOT假值参数", "not", []interface{}{false}, true},
		{"NOT多个参数取反第一个", "not", []interface{}{true, false}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait, rets, err := l.Call(ctx, tt.op, 1, tt.args...)
			if err != nil {
				t.Errorf("调用 Call() 出错: %v", err)
				return
			}
			if wait {
				t.Error("Call() 返回 wait = true, 期望 false")
			}
			if len(rets) != 1 {
				t.Errorf("Call() 返回 %d 个结果, 期望 1", len(rets))
				return
			}
			if result, ok := rets[0].(bool); !ok || result != tt.expected {
				t.Errorf("Call() = %v, 期望 %v", rets[0], tt.expected)
			}
		})
	}

	// 测试无效操作
	t.Run("测试无效操作", func(t *testing.T) {
		invalidOps := []string{"xor", "nand", "", "AND"}
		for _, op := range invalidOps {
			t.Run(op, func(t *testing.T) {
				wait, rets, err := l.Call(ctx, op, 1, true, false)
				if err == nil {
					t.Error("期望无效操作返回错误")
				}
				if wait {
					t.Error("Call() 返回 wait = true, 期望 false")
				}
				if len(rets) != 0 {
					t.Errorf("Call() 返回 %d 个结果, 期望 0", len(rets))
				}
			})
		}
	})

	// 测试 isTruthy 的各种情况
	truthyTests := []struct {
		name     string
		arg      interface{}
		expected bool
	}{
		{"nil值", nil, false},
		{"bool false", false, false},
		{"bool true", true, true},
		{"数字0", 0, false},
		{"数字1", 1, true},
		{"空字符串", "", false},
		{"非空字符串", "hello", true},
		{"空切片", []int{}, false},
		{"非空切片", []int{1}, true},
	}

	for _, tt := range truthyTests {
		t.Run("测试isTruthy_"+tt.name, func(t *testing.T) {
			// 通过NOT操作来测试isTruthy的结果
			_, rets, err := l.Call(ctx, "not", 1, tt.arg)
			if err != nil {
				t.Errorf("调用 Call() 出错: %v", err)
				return
			}
			if result, ok := rets[0].(bool); !ok || result == tt.expected {
				t.Errorf("isTruthy(%v) = %v, 期望 %v", tt.arg, !result, tt.expected)
			}
		})
	}
}

func TestIsFalsy(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  bool
	}{
		// nil case
		{"nil", nil, true},

		// bool cases
		{"true bool", true, false},
		{"false bool", false, true},

		// integer cases
		{"positive int", 42, false},
		{"zero int", 0, true},
		{"negative int", -1, false},
		{"int8 zero", int8(0), true},
		{"int16 zero", int16(0), true},
		{"int32 zero", int32(0), true},
		{"int64 zero", int64(0), true},

		// unsigned integer cases
		{"uint zero", uint(0), true},
		{"uint positive", uint(42), false},
		{"uint8 zero", uint8(0), true},
		{"uint16 zero", uint16(0), true},
		{"uint32 zero", uint32(0), true},
		{"uint64 zero", uint64(0), true},

		// float cases
		{"float64 zero", 0.0, true},
		{"float64 non-zero", 3.14, false},
		{"float32 zero", float32(0.0), true},
		{"float32 non-zero", float32(1.23), false},
		{"float64 negative", -0.5, false},

		// string cases
		{"empty string", "", true},
		{"non-empty string", "hello", false},

		// slice/array cases
		{"empty slice", []interface{}{}, true},
		{"non-empty slice", []interface{}{1, "two"}, false},
		{"empty array", [0]int{}, true},
		{"non-empty array", [2]int{1, 2}, false},

		// map cases
		{"empty map", map[interface{}]interface{}{}, true},
		{"non-empty map", map[interface{}]interface{}{"key": "value"}, false},

		// other types (should be falsy)
		{"struct", struct{}{}, false},
		{"pointer to zero", new(int), false}, // pointer to 0 is not nil
		{"channel", make(chan int), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isFalsy(tt.input); got != tt.want {
				t.Errorf("isFalsy(%v) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
