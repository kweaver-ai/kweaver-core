package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdateBuilder_Update(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})

	assert.Equal(t, "value1", updateBuilder.updateFieldKVPairs["key1"], "updateBuilder.Update() failed")
}

func TestUpdateBuilder_From(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.From("table1")

	assert.Equal(t, "table1", updateBuilder.fromTable, "updateBuilder.From() failed")
}

func TestUpdateBuilder_Where(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Where("key1", OperatorEq, "value1")

	assert.Equal(t, 1, len(updateBuilder.whereBuilder.AndClauses), "updateBuilder.Where() failed")
}

func TestUpdateBuilder_Or(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Or("key1", OperatorEq, "value1")

	assert.Equal(t, 1, len(updateBuilder.whereBuilder.OrClauses), "updateBuilder.Or() failed")
}

func TestUpdateBuilder_ToUpdateSql(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value1")
	updateBuilder.Or("key2", OperatorEq, "value2")
	sql, args, err := updateBuilder.ToUpdateSQL()

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, "update table1 set key1 = ?, key2 = ? where key1 = ? or key2 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value2", "value1", "value2"}, args, "updateBuilder.ToUpdateSQL() failed")
}

func TestUpdateBuilder_UpdateByStruct_ToUpdateSql(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()

	// 1.当struct中的字段是string或者number时
	updateBuilder.UpdateByStruct(struct {
		Key1 string `db:"key1"`
		Key2 int    `db:"key2"`
	}{
		Key1: "value1",
		Key2: 1,
	})
	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value3")
	updateBuilder.Or("key2", OperatorEq, 2)
	sql, args, err := updateBuilder.ToUpdateSQL()

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, "update table1 set key1 = ?, key2 = ? where key1 = ? or key2 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, []interface{}{"value1", 1, "value3", 2}, args, "updateBuilder.ToUpdateSQL() failed")

	// 2.当struct中的字段不是string或者number时
	updateBuilder.UpdateByStruct(struct {
		Key1 string `db:"key1"`
		Key2 bool   `db:"key2"`
	}{
		Key1: "value1",
		Key2: true,
	})

	sql, args, err = updateBuilder.ToUpdateSQL() //nolint:staticcheck,ineffassign

	assert.NotNil(t, err, "updateBuilder.ToUpdateSQL() failed")
	assert.Equal(t, err.Error(), "only support string number *string *number, but field Key2 is bool", "updateBuilder.ToUpdateSQL() failed")
}

func TestUpdateBuilder_ToUpdateSqlFullFeatures(t *testing.T) {
	t.Parallel()

	// 1、updateBuilder.Update
	updateBuilder := NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})

	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value3")
	updateBuilder.Or("key2", OperatorEq, "value4")
	updateBuilder.In("key3", []int{1, 2, 3})
	updateBuilder.NotIn("key4", []int{4, 5, 6})
	sql, args, err := updateBuilder.ToUpdateSQL()

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, "update table1 set key1 = ?, key2 = ? where key1 = ? and key3 in (?,?,?) and key4 not in (?,?,?) or key2 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value2", "value3", 1, 2, 3, 4, 5, 6, "value4"}, args, "updateBuilder.ToUpdateSQL() failed")

	// use UpdateByStruct to set updateFieldKVPairs
	// 2、updateBuilder.UpdateByStruct
	updateBuilder.UpdateByStruct(struct {
		Key1 string `db:"key1"`
		Key2 string `db:"key2"`
		Key3 string `db:"key3"`
		Key4 string `json:"key4"` // will be ignored, because tag is not "db"
		Key5 int    `db:"key5"`
		Key6 int    `db:"key6"`
	}{
		Key1: "value11",
		Key2: "value22",
		Key3: "", // empty string, will be ignored
		Key4: "value4",
		Key5: 5,
		Key6: 0, // will be ignored, because value is 0
	})

	sql, args, err = updateBuilder.ToUpdateSQL()

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, "update table1 set key1 = ?, key2 = ?, key5 = ? where key1 = ? and key3 in (?,?,?) and key4 not in (?,?,?) or key2 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, []interface{}{"value11", "value22", 5, "value3", 1, 2, 3, 4, 5, 6, "value4"}, args, "updateBuilder.ToUpdateSQL() failed")

	//	3、指定更新字段
	// 3.1 updateBuilder.Update
	updateBuilder = NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value3")

	updateBuilder.SetUpdateFields([]string{"key1"})
	sql, args, err = updateBuilder.ToUpdateSQL()

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	// sql语句中只有key1=value1，key2=value2被忽略
	assert.Equal(t, "update table1 set key1 = ? where key1 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value3"}, args, "updateBuilder.ToUpdateSQL() failed")

	// updateBuilder.updateFieldKVPairs不变
	assert.Equal(t, "value1", updateBuilder.updateFieldKVPairs["key1"], "updateBuilder.Update() failed")
	assert.Equal(t, "value2", updateBuilder.updateFieldKVPairs["key2"], "updateBuilder.Update() failed")

	// 3.2 updateBuilder.UpdateByStruct
	updateBuilder = NewUpdateBuilder()
	updateBuilder.UpdateByStruct(struct {
		Key1 string `db:"key1"`
		Key2 string `db:"key2"`
	}{
		Key1: "value11",
		Key2: "value22",
	})

	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value3")

	updateBuilder.SetUpdateFields([]string{"key1"})

	sql, args, err = updateBuilder.ToUpdateSQL() //nolint:staticcheck,ineffassign

	assert.Nil(t, err, "updateBuilder.ToUpdateSQL() failed")

	assert.Equal(t, "update table1 set key1 = ? where key1 = ?", sql, "updateBuilder.ToUpdateSQL() failed")

	// updateBuilder.updateFieldKVPairs不变
	assert.Equal(t, "value11", updateBuilder.updateFieldKVPairs["key1"], "updateBuilder.Update() failed")
	assert.Equal(t, "value22", updateBuilder.updateFieldKVPairs["key2"], "updateBuilder.Update() failed")
}

func TestUpdateBuilder_Tag(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Tag("custom")

	assert.Equal(t, "custom", updateBuilder.tag)
}

func TestUpdateBuilder_SetWhereBuilder(t *testing.T) {
	t.Parallel()

	wb := NewWhereBuilder()
	wb.Where("id", OperatorEq, 1)

	updateBuilder := NewUpdateBuilder()
	updateBuilder.SetWhereBuilder(wb)

	assert.Equal(t, wb, updateBuilder.whereBuilder)
}

func TestUpdateBuilder_TagWithStruct(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Tag("json")

	updateBuilder.UpdateByStruct(struct {
		Key1 string `json:"key1"`
		Key2 int    `json:"key2"`
	}{
		Key1: "value1",
		Key2: 1,
	})

	assert.Equal(t, "value1", updateBuilder.updateFieldKVPairs["key1"])
	assert.Equal(t, 1, updateBuilder.updateFieldKVPairs["key2"])
}

func TestUpdateBuilder_ToUpdateSQL_NoFieldsToUpdate(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
	})
	updateBuilder.From("table1")
	updateBuilder.Where("key1", OperatorEq, "value3")

	// Set update fields to exclude the only field
	updateBuilder.SetUpdateFields([]string{"key2"})

	sql, args, err := updateBuilder.ToUpdateSQL()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no fields to update")
	assert.Empty(t, sql)
	assert.Empty(t, args)
}

func TestUpdateBuilder_ToUpdateSQL_WhereBuilderError(t *testing.T) {
	t.Parallel()

	updateBuilder := NewUpdateBuilder()
	updateBuilder.Update(map[string]interface{}{
		"key1": "value1",
	})
	updateBuilder.From("table1")

	// Use In with an unsupported type (not a slice) to cause an error
	updateBuilder.In("key1", "not a slice")

	sql, args, err := updateBuilder.ToUpdateSQL()

	// Print to see what's happening
	t.Logf("err: %v, sql: %s, args: %v", err, sql, args)

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "OperatorNotIn and OperatorIn only support")
}
