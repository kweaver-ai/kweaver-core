package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSelectBuilder_Simple 简单的select示例
func TestSelectBuilder_Simple(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()

	sb.Select([]string{"id", "name"}).From("users").
		Where("id", OperatorEq, 1).
		Or("name", OperatorEq, "John").
		Limit(10).Offset(0)

	sqlStr, args, err := sb.ToSelectSQL()
	assert.Equal(t, nil, err)

	expectedSQL := "select id,name from users where id = ? or name = ? limit 10 offset 0"

	assert.Equal(t, expectedSQL, sqlStr)

	expectedArgs := []interface{}{1, "John"}

	assert.Equal(t, expectedArgs, args)
}

// TestSelectBuilder_Complex 复杂的select示例
func TestSelectBuilder_Complex(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()

	sb.Select([]string{"id", "name"}).From("users").
		Where("id", OperatorEq, 1).
		Or("name", OperatorEq, "John").
		Limit(10).Offset(0)

	sqlStr, args, err := sb.ToSelectSQL()
	assert.Equal(t, nil, err)

	expectedSQL := "select id,name from users where id = ? or name = ? limit 10 offset 0"

	assert.Equal(t, expectedSQL, sqlStr)

	expectedArgs := []interface{}{1, "John"}

	assert.Equal(t, expectedArgs, args)
}

func TestSelectBuilder_SelectFields(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"id", "name", "email"})

	fields := sb.SelectFields()
	assert.Equal(t, []string{"id", "name", "email"}, fields)
}

func TestSelectBuilder_GetFromTable(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.From("users")

	table := sb.GetFromTable()
	assert.Equal(t, "users", table)
}

func TestSelectBuilder_GetWhereBuilder(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	wb := sb.GetWhereBuilder()

	assert.NotNil(t, wb)
}

func TestSelectBuilder_Page(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		page          int
		size          int
		expectedLimit int
		expectedOff   int
	}{
		{"first page default size", 1, 0, 10, 0},
		{"second page", 2, 10, 10, 10},
		{"third page custom size", 3, 25, 25, 50},
		{"zero page uses default", 0, 10, 10, 0},
		{"zero size uses default", 5, 0, 10, 40},
		{"both zero use defaults", 0, 0, 10, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			sb := NewSelectBuilder()
			sb.Page(tt.page, tt.size)
			sb.From("users")

			sql, _, _ := sb.ToSelectSQL()
			assert.Contains(t, sql, "limit")
			assert.Contains(t, sql, "offset")
		})
	}
}

func TestSelectBuilder_Order(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").Order("id desc")

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Equal(t, "select * from users order by id desc", sql)
	assert.Empty(t, args)
}

func TestSelectBuilder_In(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").In("id", []int{1, 2, 3})

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Contains(t, sql, "in")
	assert.Equal(t, []interface{}{1, 2, 3}, args)
}

func TestSelectBuilder_Like(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").Like("name", "John")

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Contains(t, sql, "like")
	// Like adds % wildcards around the value
	assert.Equal(t, []interface{}{"%John%"}, args)
}

func TestSelectBuilder_NotIn(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").NotIn("status", []string{"deleted", "banned"})

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Contains(t, sql, "not in")
	assert.Equal(t, []interface{}{"deleted", "banned"}, args)
}

func TestSelectBuilder_SetWhereBuilder(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("id", OperatorEq, 1)

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").SetWhereBuilder(wb)

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Contains(t, sql, "where id = ?")
	assert.Equal(t, []interface{}{1}, args)
}

func TestSelectBuilder_WhereRaw(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").WhereRaw("id > ?", 100)

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	// WhereRaw adds parentheses around the condition
	assert.Contains(t, sql, "where (id > ?)")
	assert.Equal(t, []interface{}{100}, args)
}

func TestSelectBuilder_WhereOrRaw(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users").
		Where("id", OperatorEq, 1).
		WhereOrRaw("status = ?", "active")

	sql, args, err := sb.ToSelectSQL()
	assert.NoError(t, err)
	assert.Contains(t, sql, "or")
	assert.Equal(t, []interface{}{1, "active"}, args)
}

func TestSelectBuilder_ToSelectSQL_WhereBuilderError(t *testing.T) {
	t.Parallel()

	sb := NewSelectBuilder()
	sb.Select([]string{"*"}).From("users")

	// Use In with an unsupported type (not a slice) to cause an error
	sb.In("id", "not a slice")

	_, _, err := sb.ToSelectSQL()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "OperatorNotIn and OperatorIn only support")
}
