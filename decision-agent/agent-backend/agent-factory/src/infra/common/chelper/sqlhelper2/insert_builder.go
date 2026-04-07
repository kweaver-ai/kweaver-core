package sqlhelper2

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

type (
	insertFieldKVPairs map[string]interface{}
	InsertBuilder      struct {
		// insertFieldKVPairs map[string]interface{}
		rows      []insertFieldKVPairs
		fromTable string
		tag       string

		err error
	}
)

func NewInsertBuilder() *InsertBuilder {
	return &InsertBuilder{
		rows: make([]insertFieldKVPairs, 0),
		tag:  "db",
	}
}

func (ib *InsertBuilder) From(table string) *InsertBuilder {
	ib.fromTable = table
	return ib
}

func (ib *InsertBuilder) Tag(tag string) *InsertBuilder {
	ib.tag = tag
	return ib
}

func (ib *InsertBuilder) Insert(insertFieldKVPairs map[string]interface{}) *InsertBuilder {
	ib.rows = append(ib.rows, insertFieldKVPairs)
	return ib
}

// InsertStruct 将结构体转换为插入的键值对
// 结构体处理逻辑见struct2InsertPairsMapByTag的注释
func (ib *InsertBuilder) InsertStruct(obj interface{}) *InsertBuilder {
	if ib.tag == "" {
		panic("tag must be set before calling InsertStruct")
	}

	pairs, err := struct2SQLValPairsMapByTag(obj, ib.tag)
	if err != nil {
		ib.err = err
		return ib
	}

	// InsertStruct is called only once, so we can clear the rows
	ib.rows = make([]insertFieldKVPairs, 0)
	ib.rows = append(ib.rows, pairs)

	return ib
}

// InsertStructs 将结构体数组转换为插入的键值对数组
// 结构体处理逻辑见struct2InsertPairsMapByTag的注释
func (ib *InsertBuilder) InsertStructs(objs interface{}) *InsertBuilder {
	if ib.tag == "" {
		panic("tag must be set before calling InsertStruct")
	}

	v := reflect.ValueOf(objs)
	if v.Kind() != reflect.Slice {
		panic("objs must be a slice")
	}

	if v.Len() == 0 {
		//nolint:goerr113
		ib.err = errors.New("[InsertStructs]: no rows to insert")
		return ib
	}

	// InsertStructs is called only once, so we can clear the rows
	ib.rows = make([]insertFieldKVPairs, 0, v.Len())

	for i := 0; i < v.Len(); i++ {
		pairs, err := struct2SQLValPairsMapByTag(v.Index(i).Interface(), ib.tag)
		if err != nil {
			ib.err = err
			return ib
		}

		ib.rows = append(ib.rows, pairs)
	}

	return ib
}

func (ib *InsertBuilder) ToInsertSQL() (sql string, args []interface{}, err error) {
	if ib.err != nil {
		err = ib.err
		return
	}

	if len(ib.rows) == 0 {
		//nolint:goerr113
		err = errors.New("no rows to insert")
		return
	}

	kVPairs := ib.rows[0]
	args = make([]interface{}, 0, len(kVPairs))

	sql = "insert into " + ib.fromTable + " ("

	sortedKeys := make([]string, 0, len(kVPairs))

	for k := range kVPairs {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		sql += k + ","
		args = append(args, kVPairs[k])
	}

	sql = sql[:len(sql)-1] + ") values ("

	for range sortedKeys {
		sql += "?,"
	}

	sql = sql[:len(sql)-1] + ")"

	return
}

func (ib *InsertBuilder) toBatchInsertSQLCheck() {
	if ib.err != nil {
		return
	}

	//	1. 检查ib.rows中的每一条插入记录的字段的数量是否一致
	for i, _kVPairs := range ib.rows {
		if len(_kVPairs) != len(ib.rows[0]) {
			// 输出详细的错误信息方便排查
			fmt.Printf("[InsertBuilder] Error: insert rows have different number of fields\n")
			fmt.Printf("  - Row[0] has %d fields: %v\n", len(ib.rows[0]), getMapKeys(ib.rows[0]))
			fmt.Printf("  - Row[%d] has %d fields: %v\n", i, len(_kVPairs), getMapKeys(_kVPairs))
			//nolint:goerr113
			ib.err = errors.New("insert rows have different number of fields")

			return
		}
	}
	//	2. 检查ib.rows中的每一条插入记录的字段是否一致
	for i, _kVPairs := range ib.rows {
		for k := range _kVPairs {
			if _, ok := ib.rows[0][k]; !ok {
				// 输出详细的错误信息方便排查
				fmt.Printf("[InsertBuilder] Error: insert rows have different fields\n")
				fmt.Printf("  - Row[0] fields: %v\n", getMapKeys(ib.rows[0]))
				fmt.Printf("  - Row[%d] fields: %v\n", i, getMapKeys(_kVPairs))
				fmt.Printf("  - Missing field in Row[0]: %s\n", k)
				//nolint:goerr113
				ib.err = errors.New("insert rows have different fields")

				return
			}
		}
	}
}

// getMapKeys 获取map的所有key，用于调试输出
func getMapKeys(m insertFieldKVPairs) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}

func (ib *InsertBuilder) ToBatchInsertSQL() (sql string, args []interface{}, err error) {
	// 检查插入的记录是否正确
	ib.toBatchInsertSQLCheck()

	if ib.err != nil {
		err = ib.err
		return
	}

	kVPairs := ib.rows[0]
	args = make([]interface{}, 0, len(kVPairs)*len(ib.rows))

	sql = "insert into " + ib.fromTable + " ("

	sortedKeys := make([]string, 0, len(kVPairs))

	for k := range kVPairs {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		sql += k + ","
	}

	sql = sql[:len(sql)-1] + ") values "

	for _, _kVPairs := range ib.rows {
		sql += "("
		for _, k := range sortedKeys {
			sql += "?,"

			args = append(args, _kVPairs[k])
		}

		sql = sql[:len(sql)-1] + "),"
	}

	sql = sql[:len(sql)-1]

	return
}
