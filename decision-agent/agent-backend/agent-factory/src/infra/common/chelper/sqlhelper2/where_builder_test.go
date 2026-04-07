package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWhereBuilder_ToWhereSql(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("name", OperatorEq, "John")
	wb.Where("city", OperatorIn, []string{"New York", "Los Angeles"})
	wb.Or("age", OperatorGt, 30)
	wb.Or("country", OperatorIn, []string{"USA", "Canada"})

	wb.In("city2", []string{"New York2", "Los Angeles2"})

	sqlStr, args, err := wb.ToWhereSQL()
	assert.Equal(t, nil, err)

	assert.Equal(t, "name = ? and city in (?,?) and city2 in (?,?) or age > ? or country in (?,?)", sqlStr)

	assert.Equal(t, []interface{}{"John", "New York", "Los Angeles", "New York2", "Los Angeles2", 30, "USA", "Canada"}, args)
}

func TestWhereBuilder_WhereEqual_WhereNotEqual(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.WhereEqual("name", "John")
	wb.WhereNotEqual("country", "USA")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.Equal(t, nil, err)
	assert.Equal(t, "name = ? and country <> ?", sqlStr)
	assert.Equal(t, []interface{}{"John", "USA"}, args)
}

func TestWhereBuilder_WhereRaw(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.WhereRaw("name = ? and country <> ?", "John", "USA")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.Equal(t, nil, err)
	assert.Equal(t, "(name = ? and country <> ?)", sqlStr)
	assert.Equal(t, []interface{}{"John", "USA"}, args)
}

func TestWhereBuilder_WhereRaw2(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("age", OperatorEq, 30)
	wb.WhereRaw("name = ? and country <> ?", "John", "USA")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.Equal(t, nil, err)
	assert.Equal(t, "age = ? and (name = ? and country <> ?)", sqlStr)
	assert.Equal(t, []interface{}{30, "John", "USA"}, args)
}

func TestWhereBuilder_hasOuterParentheses(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: false,
		},
		{
			name:     "单个字符",
			input:    "a",
			expected: false,
		},
		{
			name:     "简单条件无括号",
			input:    "a=b",
			expected: false,
		},
		{
			name:     "完整的外层括号",
			input:    "(a=b)",
			expected: true,
		},
		{
			name:     "复杂条件有完整外层括号",
			input:    "(a=b and c=d)",
			expected: true,
		},
		{
			name:     "嵌套括号有完整外层括号",
			input:    "((a=b) and (c=d))",
			expected: true,
		},
		{
			name:     "没有完整外层括号的复杂条件",
			input:    "(a=b and c=d) or (e=f)",
			expected: false,
		},
		{
			name:     "括号不匹配的情况",
			input:    "(a=b and c=d",
			expected: false,
		},
		{
			name:     "只有右括号",
			input:    "a=b)",
			expected: false,
		},
		{
			name:     "只有左括号",
			input:    "(a=b",
			expected: false,
		},
		{
			name:     "多个独立括号组",
			input:    "(a=b) and (c=d) or (e=f)",
			expected: false,
		},
		{
			name:     "开头结尾有括号但中间断开",
			input:    "(a=b) or c=d or (e=f)",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := wb.hasOuterParentheses(tt.input)
			assert.Equal(t, tt.expected, result, "测试用例: %s", tt.name)
		})
	}
}

func TestWhereBuilder_hlCondition(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "空字符串",
			input:    "",
			expected: "",
		},
		{
			name:     "简单条件",
			input:    "a=b",
			expected: "(a=b)",
		},
		{
			name:     "已有完整外层括号",
			input:    "(a=b)",
			expected: "(a=b)",
		},
		{
			name:     "复杂条件有完整外层括号",
			input:    "(a=b and c=d)",
			expected: "(a=b and c=d)",
		},
		{
			name:     "需要添加括号的复杂条件",
			input:    "(a=b and c=d) or (e=f)",
			expected: "((a=b and c=d) or (e=f))",
		},
		{
			name:     "嵌套括号有完整外层括号",
			input:    "((a=b) and (c=d))",
			expected: "((a=b) and (c=d))",
		},
		{
			name:     "多个独立括号组需要添加外层括号",
			input:    "(a=b) and (c=d) or (e=f)",
			expected: "((a=b) and (c=d) or (e=f))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := wb.hlCondition(tt.input)
			assert.Equal(t, tt.expected, result, "测试用例: %s", tt.name)
		})
	}
}

func TestWhereBuilder_Like(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Like("name", "John")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "name like ?", sqlStr)
	// Like adds % wildcards around the value
	assert.Equal(t, []interface{}{"%John%"}, args)
}

func TestWhereBuilder_OrEqual(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("id", OperatorEq, 1)
	wb.OrEqual("status", "active")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "id = ? or status = ?", sqlStr)
	assert.Equal(t, []interface{}{1, "active"}, args)
}

func TestWhereBuilder_OrNotEqual(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("id", OperatorEq, 1)
	wb.OrNotEqual("status", "deleted")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "id = ? or status <> ?", sqlStr)
	assert.Equal(t, []interface{}{1, "deleted"}, args)
}

func TestWhereBuilder_OrLike(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("id", OperatorEq, 1)
	wb.OrLike("name", "John")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "id = ? or name like ?", sqlStr)
	// OrLike adds % wildcards around the value
	assert.Equal(t, []interface{}{1, "%John%"}, args)
}

func TestWhereBuilder_OrLike_MultipleConditions(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.WhereEqual("category", "user")
	wb.OrLike("name", "admin")
	wb.OrNotEqual("status", "inactive")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "category = ? or name like ? or status <> ?", sqlStr)
	assert.Equal(t, []interface{}{"user", "%admin%", "inactive"}, args)
}

func TestWhereBuilder_OrNotEqual_OnlyOrConditions(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.OrNotEqual("status", "deleted")
	wb.OrNotEqual("status", "archived")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "status <> ? or status <> ?", sqlStr)
	assert.Equal(t, []interface{}{"deleted", "archived"}, args)
}

func TestWhereBuilder_OrLike_OnlyOrConditions(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.OrLike("name", "John")
	wb.OrLike("description", "test")

	sqlStr, args, err := wb.ToWhereSQL()
	assert.NoError(t, err)
	assert.Equal(t, "name like ? or description like ?", sqlStr)
	assert.Equal(t, []interface{}{"%John%", "%test%"}, args)
}
