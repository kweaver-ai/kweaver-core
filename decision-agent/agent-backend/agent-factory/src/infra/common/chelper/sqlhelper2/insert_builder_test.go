package sqlhelper2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsertBuilder_Insert(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.Insert(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})

	assert.Equal(t, "value1", ib.rows[0]["key1"], "ib.Insert() failed")
}

func TestInsertBuilder_From(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")

	assert.Equal(t, "table1", ib.fromTable, "ib.From() failed")
}

func TestInsertBuilder_Tag(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.Tag("json")

	assert.Equal(t, "json", ib.tag, "ib.Tag() failed")
}

func TestInsertBuilder_InsertStruct(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.Tag("json")
	ib.InsertStruct(struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}{
		Key1: "value1",
		Key2: "value2",
	})

	assert.Equal(t, "value1", ib.rows[0]["key1"], "ib.InsertStruct() failed")
}

func TestInsertBuilder_Insert_ToInsertSql(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.Insert(map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	})
	ib.From("table1")
	sql, args, err := ib.ToInsertSQL()

	assert.Nil(t, err, "ib.ToInsertSQL() failed")

	assert.Equal(t, "insert into table1 (key1,key2) values (?,?)", sql, "ib.ToInsertSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value2"}, args, "ib.ToInsertSQL() failed")
}

func TestInsertBuilder_InsertStruct_ToInsertSql(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	// 1.测试正常情况
	ib.InsertStruct(struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
	}{
		Key1: "value1",
		Key2: "value2",
	})

	sql, args, err := ib.ToInsertSQL()

	assert.Nil(t, err, "ib.ToInsertSQL() failed")

	assert.Equal(t, "insert into table1 (key1,key2) values (?,?)", sql, "ib.ToInsertSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value2"}, args, "ib.ToInsertSQL() failed")

	// 2.当struct中的字段不是string或者number时
	ib = NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")
	ib.InsertStruct(struct {
		Key1 string `json:"key1"`
		Key2 bool   `json:"key2"`
	}{
		Key1: "value1",
		Key2: true,
	})

	sql, args, err = ib.ToInsertSQL() //nolint:staticcheck,ineffassign
	assert.NotNil(t, err, "ib.ToInsertSQL() failed")
	assert.Equal(t, err.Error(), "only support string number *string *number, but field Key2 is bool", "ib.ToInsertSQL() failed")

	// 3.当struct中的字段为零值时
	ib = NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")
	ib.InsertStruct(struct {
		Key1 string `json:"key1"`
		Key2 string `json:"key2"`
		Key3 int    `json:"key3"`
	}{
		Key1: "value1",
		Key2: "",
		Key3: 0,
	})

	sql, args, err = ib.ToInsertSQL()
	assert.Nil(t, err, "ib.ToInsertSQL() failed")
	assert.Equal(t, "insert into table1 (key1) values (?)", sql, "ib.ToInsertSQL() failed")
	assert.Equal(t, []interface{}{"value1"}, args, "ib.ToInsertSQL() failed")

	// 4.当struct中的字段为指针时
	// 4.1 指针的值为零值
	ib = NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")
	ib.InsertStruct(struct {
		Key1 string  `json:"key1"`
		Key2 *string `json:"key2"`
	}{
		Key1: "value1",
		Key2: nil,
	})

	sql, args, err = ib.ToInsertSQL()
	assert.Nil(t, err, "ib.ToInsertSQL() failed")
	assert.Equal(t, "insert into table1 (key1) values (?)", sql, "ib.ToInsertSQL() failed")
	assert.Equal(t, []interface{}{"value1"}, args, "ib.ToInsertSQL() failed")

	// 4.2 指针底层的值为非字符串或者数字
	ib = NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")
	ib.InsertStruct(struct {
		Key1 string `json:"key1"`
		Key2 *bool  `json:"key2"`
	}{
		Key1: "value1",
		Key2: new(bool),
	})

	sql, args, err = ib.ToInsertSQL() //nolint:staticcheck,ineffassign
	assert.NotNil(t, err, "ib.ToInsertSQL() failed")
	assert.Equal(t, "only support string number *string *number, but field Key2 is ptr and underlying type is bool", err.Error(), "ib.ToInsertSQL() failed")

	// 4.2 指针底层的值为字符串或者数字

	ib = NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	value2 := "value2"

	ib.InsertStruct(struct {
		Key1 string  `json:"key1"`
		Key2 *string `json:"key2"`
	}{
		Key1: "value1",
		Key2: &value2,
	})

	sql, args, err = ib.ToInsertSQL()
	assert.Nil(t, err, "ib.ToInsertSQL() failed")
	assert.Equal(t, "insert into table1 (key1,key2) values (?,?)", sql, "ib.ToInsertSQL() failed")
	assert.Equal(t, []interface{}{"value1", "value2"}, args, "ib.ToInsertSQL() failed")
}

func TestInsertBuilder_InsertStructs_ToBatchInsertSql(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	ib.InsertStructs([]interface{}{
		struct {
			Key1 string `json:"key1"`
			Key2 string `json:"key2"`
		}{
			Key1: "value1",
			Key2: "value2",
		},
		struct {
			Key1 string `json:"key1"`
			Key2 string `json:"key2"`
		}{
			Key1: "value3",
			Key2: "value4",
		},
	})

	sql, args, err := ib.ToBatchInsertSQL()

	assert.Nil(t, err, "ib.ToBatchInsertSQL() failed")

	assert.Equal(t, "insert into table1 (key1,key2) values (?,?),(?,?)", sql, "ib.ToBatchInsertSQL() failed")

	assert.Equal(t, []interface{}{"value1", "value2", "value3", "value4"}, args, "ib.ToBatchInsertSQL() failed")
}

func TestInsertBuilder_InsertStructs_ToBatchInsertSql_toBatchInsertSqlCheck(t *testing.T) {
	t.Parallel()

	// 1. 检查ib.rows中的每一条插入记录的字段的数量是否一致
	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	ib.InsertStructs([]interface{}{
		struct {
			Key1 string `json:"key1"`
			Key2 string `json:"key2"`
			Key3 string `json:"key3"`
		}{
			Key1: "value1",
			Key2: "value2",
			Key3: "value3",
		},
		struct {
			Key1 string `json:"key1"`
			Key2 string `json:"key2"`
		}{
			Key1: "value3",
			Key2: "value4",
		},
	})

	_, _, err := ib.ToBatchInsertSQL()

	assert.NotNil(t, err, "insert rows have different number of fields")

	// 2. 检查ib.rows中的每一条插入记录的字段是否一致
	ib2 := NewInsertBuilder()
	ib2.From("table1")
	ib2.Tag("json")

	ib2.InsertStructs([]interface{}{
		struct {
			Key1 string `json:"key1"`
			Key3 string `json:"key3"`
		}{
			Key1: "value1",
			Key3: "value3",
		},
		struct {
			Key1 string `json:"key1"`
			Key2 string `json:"key2"`
		}{
			Key1: "value3",
			Key2: "value4",
		},
	})

	_, _, err = ib2.ToBatchInsertSQL()

	assert.NotNil(t, err, "insert rows have different fields")
}

func TestInsertBuilder_InsertStructs_EmptySlice(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	ib.InsertStructs([]interface{}{})

	_, _, err := ib.ToInsertSQL()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "no rows to insert")
}

func TestInsertBuilder_InsertStructs_UnsupportedType(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	ib.InsertStructs([]interface{}{
		struct {
			Key1 string `json:"key1"`
			Key2 bool   `json:"key2"`
		}{
			Key1: "value1",
			Key2: true,
		},
	})

	_, _, err := ib.ToInsertSQL()

	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "only support string number *string *number")
}

func TestInsertBuilder_InsertStructs_PanicWhenNoTag(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("") // Explicitly set tag to empty to trigger panic

	assert.Panics(t, func() {
		ib.InsertStructs([]interface{}{
			struct {
				Key1 string `json:"key1"`
			}{
				Key1: "value1",
			},
		})
	})
}

func TestInsertBuilder_InsertStructs_PanicWhenNotSlice(t *testing.T) {
	t.Parallel()

	ib := NewInsertBuilder()
	ib.From("table1")
	ib.Tag("json")

	assert.Panics(t, func() {
		ib.InsertStructs("not a slice")
	})
}
