package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteBuilder(t *testing.T) {
	t.Parallel()

	db := NewDeleteBuilder()
	db.From("table1")
	db.Where("key1", OperatorEq, "value1")
	db.Or("key2", OperatorEq, "value2")
	db.In("key3", []string{"value3", "value4"})

	sql, args, err := db.ToDeleteSQL()
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, "delete from table1 where key1 = ? and key3 in (?,?) or key2 = ?", sql, "db.ToDeleteSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value3", "value4", "value2"}, args, "db.ToDeleteSQL() failed")
}

func TestDeleteBuilder_SetWhereBuilder(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("key1", OperatorEq, "value1")

	db := NewDeleteBuilder()
	db.SetWhereBuilder(wb)
	db.From("table1")

	sql, args, err := db.ToDeleteSQL()
	assert.NoError(t, err)
	assert.Equal(t, "delete from table1 where key1 = ?", sql)
	assert.Equal(t, []interface{}{"value1"}, args)
}

func TestDeleteBuilder_EmptyFromTable(t *testing.T) {
	t.Parallel()

	db := NewDeleteBuilder()
	db.Where("key1", OperatorEq, "value1")

	sql, args, err := db.ToDeleteSQL()
	assert.Error(t, err)
	assert.Empty(t, sql)
	assert.Empty(t, args) // Empty slice, not nil
}

func TestDeleteBuilder_WithNoWhere(t *testing.T) {
	t.Parallel()

	db := NewDeleteBuilder()
	db.From("users")

	sql, args, err := db.ToDeleteSQL()
	assert.NoError(t, err)
	assert.Equal(t, "delete from users", sql)
	assert.Empty(t, args)
}

func TestDeleteBuilder_NewDeleteBuilder(t *testing.T) {
	t.Parallel()

	db := NewDeleteBuilder()
	assert.NotNil(t, db)
	assert.NotNil(t, db.WhereBuilder)
}

func TestDeleteBuilder_WhereBuilderError(t *testing.T) {
	t.Parallel()

	db := NewDeleteBuilder()
	db.From("users")

	// Use In with an unsupported type (not a slice) to cause an error
	db.In("id", "not a slice")

	_, _, err := db.ToDeleteSQL()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "OperatorNotIn and OperatorIn only support")
}
