package sqlhelper2

import (
	"fmt"
	"strings"
)

type SelectBuilder struct {
	selectFields []string
	fromTable    string
	limit        *int
	offset       *int
	order        string
	whereBuilder *WhereBuilder
}

func (s *SelectBuilder) SelectFields() []string {
	return s.selectFields
}

func NewSelectBuilder() *SelectBuilder {
	return &SelectBuilder{
		selectFields: []string{},
		whereBuilder: NewWhereBuilder(),
		limit:        nil,
		offset:       nil,
		order:        "",
	}
}

func (s *SelectBuilder) Select(fields []string) *SelectBuilder {
	// 这样做的目的：防止在外面对selectFields进行修改
	_fields := make([]string, len(fields))
	copy(_fields, fields)

	s.selectFields = _fields

	return s
}

func (s *SelectBuilder) From(table string) *SelectBuilder {
	s.fromTable = table
	return s
}

func (s *SelectBuilder) GetFromTable() string {
	return s.fromTable
}

func (s *SelectBuilder) GetWhereBuilder() *WhereBuilder {
	return s.whereBuilder
}

func (s *SelectBuilder) Limit(limit int) *SelectBuilder {
	l := limit
	s.limit = &l

	return s
}

func (s *SelectBuilder) Offset(offset int) *SelectBuilder {
	o := offset
	s.offset = &o

	return s
}

func (s *SelectBuilder) Page(page, size int) *SelectBuilder {
	if page < 1 {
		page = 1
	}

	if size < 1 {
		size = 10
	}

	o := (page - 1) * size
	s.offset = &o
	l := size
	s.limit = &l

	return s
}

func (s *SelectBuilder) Order(order string) *SelectBuilder {
	s.order = order
	return s
}

func (s *SelectBuilder) Where(key string, op Operator, value interface{}) *SelectBuilder {
	s.whereBuilder.Where(key, op, value)
	return s
}

func (s *SelectBuilder) Or(key string, op Operator, value interface{}) *SelectBuilder {
	s.whereBuilder.Or(key, op, value)
	return s
}

func (s *SelectBuilder) In(key string, value interface{}) *SelectBuilder {
	s.whereBuilder.In(key, value)
	return s
}

func (s *SelectBuilder) Like(key string, value interface{}) *SelectBuilder {
	s.whereBuilder.Where(key, OperatorLike, value)
	return s
}

func (s *SelectBuilder) NotIn(key string, value interface{}) *SelectBuilder {
	s.whereBuilder.NotIn(key, value)
	return s
}

// ToSelectSQL 生成select语句
// 返回值:
// sqlStr: select语句
// args: 参数列表
// err: 错误
// 示例见select_builder_test.go
func (s *SelectBuilder) ToSelectSQL() (sqlStr string, args []interface{}, err error) {
	whereSQL, args, err := s.whereBuilder.ToWhereSQL()
	if err != nil {
		return
	}

	sqlStr = fmt.Sprintf("select %s from %s", strings.Join(s.selectFields, ","), s.fromTable)

	if whereSQL != "" {
		sqlStr += fmt.Sprintf(" where %s", whereSQL)
	}

	if s.order != "" {
		sqlStr += fmt.Sprintf(" order by %s", s.order)
	}

	if s.limit != nil {
		sqlStr += fmt.Sprintf(" limit %d", *s.limit)
	}

	if s.offset != nil {
		sqlStr += fmt.Sprintf(" offset %d", *s.offset)
	}

	return
}

func (s *SelectBuilder) SetWhereBuilder(whereBuilder *WhereBuilder) *SelectBuilder {
	s.whereBuilder = whereBuilder
	return s
}

func (s *SelectBuilder) WhereRaw(condition string, arg ...interface{}) *SelectBuilder {
	s.whereBuilder.WhereRaw(condition, arg...)
	return s
}

func (s *SelectBuilder) WhereOrRaw(condition string, arg ...interface{}) *SelectBuilder {
	s.whereBuilder.WhereOrRaw(condition, arg...)
	return s
}
