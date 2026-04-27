package funcs

import (
	"context"
	"testing"
)

func TestCompare(t *testing.T) {
	f := &compare{}
	ctx := context.Background()

	tests := []struct {
		name     string
		op       string
		args     []interface{}
		expected interface{}
	}{
		// eq 操作测试
		{"比较两个nil是否相等", "eq", []interface{}{nil, nil}, true},
		{"比较相同数值是否相等", "eq", []interface{}{5, 5}, true},
		{"比较不同数值是否相等", "eq", []interface{}{5, 6}, false},
		{"比较不同类型是否相等", "eq", []interface{}{5, "5"}, false},
		{"比较单个参数是否等于nil", "eq", []interface{}{5}, false},

		// ne 操作测试
		{"比较两个nil是否不等", "ne", []interface{}{nil, nil}, false},
		{"比较相同数值是否不等", "ne", []interface{}{5, 5}, false},
		{"比较不同数值是否不等", "ne", []interface{}{5, 6}, true},

		// lt 操作测试
		{"比较小数是否小于大数", "lt", []interface{}{4, 5}, true},
		{"比较相等数值是否小于", "lt", []interface{}{5, 5}, false},
		{"比较大数是否小于小数", "lt", []interface{}{6, 5}, false},
		{"比较非数字类型是否小于", "lt", []interface{}{"a", "b"}, false},

		// lte 操作测试
		{"比较小数是否小于等于大数", "lte", []interface{}{4, 5}, true},
		{"比较相等数值是否小于等于", "lte", []interface{}{5, 5}, true},
		{"比较大数是否小于等于小数", "lte", []interface{}{6, 5}, false},

		// gt 操作测试
		{"比较小数是否大于大数", "gt", []interface{}{4, 5}, false},
		{"比较相等数值是否大于", "gt", []interface{}{5, 5}, false},
		{"比较大数是否大于小数", "gt", []interface{}{6, 5}, true},

		// gte 操作测试
		{"比较小数是否大于等于大数", "gte", []interface{}{4, 5}, false},
		{"比较相等数值是否大于等于", "gte", []interface{}{5, 5}, true},
		{"比较大数是否大于等于小数", "gte", []interface{}{6, 5}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wait, rets, err := f.Call(ctx, tt.op, 1, tt.args...)
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
			if rets[0] != tt.expected {
				t.Errorf("Call() = %v, 期望 %v", rets[0], tt.expected)
			}
		})
	}

	// 测试无效操作
	t.Run("测试无效操作", func(t *testing.T) {
		wait, rets, err := f.Call(ctx, "invalid", 1, 1, 2)
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

	// 测试无参数情况
	t.Run("测试无参数比较", func(t *testing.T) {
		wait, rets, err := f.Call(ctx, "eq", 1)
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
		if rets[0] != true { // 比较 nil 与 nil
			t.Errorf("Call() = %v, 期望 true", rets[0])
		}
	})
}
