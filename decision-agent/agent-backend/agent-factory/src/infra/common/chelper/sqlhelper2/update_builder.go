package sqlhelper2

import (
	"errors"
	"sort"
	"strings"
)

type UpdateBuilder struct {
	updateFieldKVPairs map[string]interface{}
	fromTable          string
	whereBuilder       *WhereBuilder
	tag                string
	updateFieldMap     map[string]struct{}

	err error
}

func NewUpdateBuilder() *UpdateBuilder {
	return &UpdateBuilder{
		updateFieldKVPairs: map[string]interface{}{},
		whereBuilder:       NewWhereBuilder(),
		tag:                "db",
	}
}

func (u *UpdateBuilder) Tag(tag string) *UpdateBuilder {
	u.tag = tag
	return u
}

func (u *UpdateBuilder) SetWhereBuilder(whereBuilder *WhereBuilder) *UpdateBuilder {
	u.whereBuilder = whereBuilder
	return u
}

// Update 将键值对添加到更新键值对列表中
func (u *UpdateBuilder) Update(updateFieldKVPairs map[string]interface{}) *UpdateBuilder {
	u.updateFieldKVPairs = updateFieldKVPairs
	return u
}

// UpdateByStruct 将结构体转换为更新的键值对
func (u *UpdateBuilder) UpdateByStruct(s interface{}) *UpdateBuilder {
	u.updateFieldKVPairs, u.err = struct2SQLValPairsMapByTag(s, u.tag)
	return u
}

func (u *UpdateBuilder) SetUpdateFields(fields []string) *UpdateBuilder {
	u.updateFieldMap = make(map[string]struct{}, len(fields))
	for _, field := range fields {
		u.updateFieldMap[field] = struct{}{}
	}

	return u
}

func (u *UpdateBuilder) From(table string) *UpdateBuilder {
	u.fromTable = table
	return u
}

func (u *UpdateBuilder) Where(key string, op Operator, value interface{}) *UpdateBuilder {
	u.whereBuilder.Where(key, op, value)
	return u
}

func (u *UpdateBuilder) Or(key string, op Operator, value interface{}) *UpdateBuilder {
	u.whereBuilder.Or(key, op, value)
	return u
}

func (u *UpdateBuilder) In(key string, value interface{}) *UpdateBuilder {
	u.whereBuilder.In(key, value)
	return u
}

func (u *UpdateBuilder) NotIn(key string, value interface{}) *UpdateBuilder {
	u.whereBuilder.NotIn(key, value)
	return u
}

func (u *UpdateBuilder) ToUpdateSQL() (sql string, args []interface{}, err error) {
	if u.err != nil {
		err = u.err
		return
	}

	// 1. 拷贝一份，避免修改原始数据
	updateKvPairs := make(map[string]interface{})
	for k, v := range u.updateFieldKVPairs {
		updateKvPairs[k] = v
	}

	// 2. 如果指定了更新字段，则只更新指定的字段
	if len(u.updateFieldMap) > 0 {
		//	更改updateKvPairs
		for k := range updateKvPairs {
			if _, ok := u.updateFieldMap[k]; !ok {
				delete(updateKvPairs, k)
			}
		}
	}

	if len(updateKvPairs) == 0 {
		//nolint:goerr113
		err = errors.New("no fields to update")
		return
	}

	// 3. 拼接sql
	sql = "update " + u.fromTable + " set "

	sortedKeys := make([]string, 0, len(updateKvPairs))

	for k := range updateKvPairs {
		sortedKeys = append(sortedKeys, k)
	}

	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		sql += k + " = ?, "
		args = append(args, updateKvPairs[k])
	}

	sql = strings.TrimRight(sql, ", ")

	whereSQL, whereArgs, err := u.whereBuilder.ToWhereSQL()
	if err != nil {
		return
	}

	sql += " where " + whereSQL

	args = append(args, whereArgs...)

	return
}
