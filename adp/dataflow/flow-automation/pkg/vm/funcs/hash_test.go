package funcs

import (
	"context"
	"testing"

	"github.com/go-playground/assert/v2"
)

func TestGet(t *testing.T) {
	f := &Hash{}
	ctx := context.Background()

	_, v, _ := f.Call(ctx, "get", 1, nil, []any{})
	assert.Equal(t, v[0], nil)

	// 路径为空时返回输入值
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{})
	assert.Equal(t, v[0], "value")

	// 获取字符串的指定字符
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{1})
	assert.Equal(t, v[0], "a")

	// 超出字符串长度
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{5})
	assert.Equal(t, v[0], "")

	// 获取字符串的指定字符, 下标为文本时先尝试转为数字, 否则返回 ""
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{"1"})
	assert.Equal(t, v[0], "a")

	// 获取字符串的指定字符, 下标为文本时先尝试转为数字, 否则返回 ""
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{"1", 0})
	assert.Equal(t, v[0], "a")

	// "value"[1] = a, a[1] = ""
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{"1", 1})
	assert.Equal(t, v[0], "")

	// 下标不能转为数字返回 ""
	_, v, _ = f.Call(ctx, "get", 1, "value", []any{"a"})
	assert.Equal(t, v[0], "")

	// 路径为空返回原始值
	_, v, _ = f.Call(ctx, "get", 1, []any{"abc", "def", 123}, []any{})
	assert.Equal(t, v[0], []any{"abc", "def", 123})

	// 获取切片值
	_, v, _ = f.Call(ctx, "get", 1, []any{"abc", "def", 123}, []any{1})
	assert.Equal(t, v[0], "def")

	// 递归获取
	_, v, _ = f.Call(ctx, "get", 1, []any{"abc", "def", 123}, []any{1, 2})
	assert.Equal(t, v[0], "f")

	// 类型
	_, v, _ = f.Call(ctx, "get", 1, []any{"abc", "def", 123}, []any{2})
	assert.Equal(t, v[0], 123)

	// 获取 map 值
	_, v, _ = f.Call(ctx, "get", 1, map[string]any{
		"a": 1,
		"b": 2,
	}, []any{"a"})
	assert.Equal(t, v[0], 1)

	// 嵌套多层
	_, v, _ = f.Call(ctx, "get", 1, map[string]any{
		"abc": []any{
			map[string]any{
				"def": map[string]any{
					"ghi": []any{"jkl"},
				},
			},
		},
	}, []any{"abc", 0, "def", "ghi", 0, 1})
	assert.Equal(t, v[0], "k")

	// 获取不可索引的类型
	_, v, _ = f.Call(ctx, "get", 1, 1, []any{"a", 1})
	assert.Equal(t, v[0], nil)
}

func TestSet(t *testing.T) {
	f := &Hash{}
	ctx := context.Background()

	_, v, _ := f.Call(ctx, "set", 1, nil)
	assert.Equal(t, v[0], nil)

	_, v, _ = f.Call(ctx, "set", 1, "hello")
	assert.Equal(t, v[0], "hello")

	_, v, _ = f.Call(ctx, "set", 1, "hello", []any{1}, "a")
	assert.Equal(t, v[0], "hallo")

	_, v, _ = f.Call(ctx, "set", 1, "hello", []any{1, 1}, "a")
	assert.Equal(t, v[0], "hallo")

	_, v, _ = f.Call(ctx, "set", 1, "", []any{1}, "a")
	assert.Equal(t, v[0], " a")

	_, v, _ = f.Call(ctx, "set", 1, []any{}, []any{1}, "a")
	assert.Equal(t, v[0], []any{nil, "a"})
}
