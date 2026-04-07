package ormhelper

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
)

// SelectBuilder SELECT语句构建器
type SelectBuilder struct {
	db      *DB
	columns []string
	table   string
	joins   []string
	where   *WhereBuilder
	groupBy []string
	having  *WhereBuilder
	orderBy []string
	limit   int
	offset  int
}

// From 指定表名
func (s *SelectBuilder) From(table string) *SelectBuilder {
	s.table = fmt.Sprintf("`%s`.`%s`", s.db.dbName, table)
	return s
}

// Join 内连接
func (s *SelectBuilder) Join(table, condition string) *SelectBuilder {
	fullTable := fmt.Sprintf("`%s`.`%s`", s.db.dbName, table)
	s.joins = append(s.joins, fmt.Sprintf("JOIN %s ON %s", fullTable, condition))
	return s
}

// LeftJoin 左连接
func (s *SelectBuilder) LeftJoin(table, condition string) *SelectBuilder {
	fullTable := fmt.Sprintf("`%s`.`%s`", s.db.dbName, table)
	s.joins = append(s.joins, fmt.Sprintf("LEFT JOIN %s ON %s", fullTable, condition))
	return s
}

// RightJoin 右连接
func (s *SelectBuilder) RightJoin(table, condition string) *SelectBuilder {
	fullTable := fmt.Sprintf("`%s`.`%s`", s.db.dbName, table)
	s.joins = append(s.joins, fmt.Sprintf("RIGHT JOIN %s ON %s", fullTable, condition))
	return s
}

// InnerJoin 内连接（Join的别名）
func (s *SelectBuilder) InnerJoin(table, condition string) *SelectBuilder {
	return s.Join(table, condition)
}

// Where 添加WHERE条件
func (s *SelectBuilder) Where(field, op string, value interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.Condition(field, op, value)
	return s
}

// WhereEq 等于条件的简写
func (s *SelectBuilder) WhereEq(field string, value interface{}) *SelectBuilder {
	return s.Where(field, "=", value)
}

// WhereNe 不等于条件
func (s *SelectBuilder) WhereNe(field string, value interface{}) *SelectBuilder {
	return s.Where(field, "!=", value)
}

// WhereGt 大于条件
func (s *SelectBuilder) WhereGt(field string, value interface{}) *SelectBuilder {
	return s.Where(field, ">", value)
}

// WhereGte 大于等于条件
func (s *SelectBuilder) WhereGte(field string, value interface{}) *SelectBuilder {
	return s.Where(field, ">=", value)
}

// WhereLt 小于条件
func (s *SelectBuilder) WhereLt(field string, value interface{}) *SelectBuilder {
	return s.Where(field, "<", value)
}

// WhereLte 小于等于条件
func (s *SelectBuilder) WhereLte(field string, value interface{}) *SelectBuilder {
	return s.Where(field, "<=", value)
}

// WhereIn IN条件
func (s *SelectBuilder) WhereIn(field string, values ...interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.In(field, values...)
	return s
}

// WhereNotIn NOT IN条件
func (s *SelectBuilder) WhereNotIn(field string, values ...interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.NotIn(field, values...)
	return s
}

// WhereLike LIKE条件
func (s *SelectBuilder) WhereLike(field, pattern string) *SelectBuilder {
	return s.Where(field, "LIKE", pattern)
}

// WhereNotLike NOT LIKE条件
func (s *SelectBuilder) WhereNotLike(field, pattern string) *SelectBuilder {
	return s.Where(field, "NOT LIKE", pattern)
}

// WhereBetween BETWEEN条件
func (s *SelectBuilder) WhereBetween(field string, start, end interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.Between(field, start, end)
	return s
}

// WhereNotBetween NOT BETWEEN条件
func (s *SelectBuilder) WhereNotBetween(field string, start, end interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.NotBetween(field, start, end)
	return s
}

// WhereNull IS NULL条件
func (s *SelectBuilder) WhereNull(field string) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.IsNull(field)
	return s
}

// WhereNotNull IS NOT NULL条件
func (s *SelectBuilder) WhereNotNull(field string) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.IsNotNull(field)
	return s
}

// And 开始AND条件组
func (s *SelectBuilder) And(fn func(*WhereBuilder)) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.And(fn)
	return s
}

// Or 开始OR条件组
func (s *SelectBuilder) Or(fn func(*WhereBuilder)) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.Or(fn)
	return s
}

// WhereRaw 原始WHERE条件
func (s *SelectBuilder) WhereRaw(condition string, args ...interface{}) *SelectBuilder {
	if s.where == nil {
		s.where = NewWhere()
	}
	s.where.Raw(condition, args...)
	return s
}

// GroupBy 分组
func (s *SelectBuilder) GroupBy(columns ...string) *SelectBuilder {
	s.groupBy = append(s.groupBy, columns...)
	return s
}

// Having HAVING条件
func (s *SelectBuilder) Having(field, op string, value interface{}) *SelectBuilder {
	if s.having == nil {
		s.having = NewWhere()
	}
	s.having.Condition(field, op, value)
	return s
}

// HavingEq HAVING等于条件
func (s *SelectBuilder) HavingEq(field string, value interface{}) *SelectBuilder {
	return s.Having(field, "=", value)
}

// HavingGt HAVING大于条件
func (s *SelectBuilder) HavingGt(field string, value interface{}) *SelectBuilder {
	return s.Having(field, ">", value)
}

// HavingLt HAVING小于条件
func (s *SelectBuilder) HavingLt(field string, value interface{}) *SelectBuilder {
	return s.Having(field, "<", value)
}

// OrderBy 排序（升序）
func (s *SelectBuilder) OrderBy(column string) *SelectBuilder {
	s.orderBy = append(s.orderBy, column)
	return s
}

// OrderByDesc 降序排序
func (s *SelectBuilder) OrderByDesc(column string) *SelectBuilder {
	s.orderBy = append(s.orderBy, column+" DESC")
	return s
}

// OrderByAsc 升序排序（OrderBy的别名）
func (s *SelectBuilder) OrderByAsc(column string) *SelectBuilder {
	s.orderBy = append(s.orderBy, column+" ASC")
	return s
}

// Limit 限制数量
func (s *SelectBuilder) Limit(limit int) *SelectBuilder {
	s.limit = limit
	return s
}

// Offset 偏移量
func (s *SelectBuilder) Offset(offset int) *SelectBuilder {
	s.offset = offset
	return s
}

// Pagination 应用分页参数
func (s *SelectBuilder) Pagination(pagination *PaginationParams) *SelectBuilder {
	if pagination != nil && pagination.Page > 0 && pagination.PageSize > 0 {
		offset := (pagination.Page - 1) * pagination.PageSize
		s.Limit(pagination.PageSize).Offset(offset)
	}
	return s
}

// Sort 应用排序参数
func (s *SelectBuilder) Sort(sort *SortParams) *SelectBuilder {
	if sort != nil && len(sort.Fields) > 0 {
		for _, field := range sort.Fields {
			if field.Order.ToUpper() == SortOrderAsc {
				s.OrderByAsc(field.Field)
			} else {
				s.OrderByDesc(field.Field)
			}
		}
	}
	return s
}

// Cursor 应用游标参数
func (s *SelectBuilder) Cursor(cursor *CursorParams) *SelectBuilder {
	if cursor != nil && cursor.Field != "" && cursor.Value != nil {
		if cursor.Direction.ToUpper() == SortOrderDesc {
			s.WhereLt(cursor.Field, cursor.Value)
		} else {
			s.WhereGt(cursor.Field, cursor.Value)
		}
	}
	return s
}

// Build 构建SQL语句
func (s *SelectBuilder) Build() (query string, args []interface{}) {
	// 构建SELECT部分
	columns := "*"
	if len(s.columns) > 0 {
		columns = strings.Join(s.columns, ", ")
	}

	query = fmt.Sprintf("SELECT %s FROM %s", columns, s.table)

	// JOIN
	if len(s.joins) > 0 {
		query += " " + strings.Join(s.joins, " ")
	}

	// WHERE
	if s.where != nil {
		whereClause, whereArgs := s.where.Build()
		if whereClause != "" {
			query += " WHERE " + whereClause
			args = append(args, whereArgs...)
		}
	}

	// GROUP BY
	if len(s.groupBy) > 0 {
		query += " GROUP BY " + strings.Join(s.groupBy, ", ")
	}

	// HAVING
	if s.having != nil {
		havingClause, havingArgs := s.having.Build()
		if havingClause != "" {
			query += " HAVING " + havingClause
			args = append(args, havingArgs...)
		}
	}

	// ORDER BY
	if len(s.orderBy) > 0 {
		query += " ORDER BY " + strings.Join(s.orderBy, ", ")
	}

	// LIMIT
	if s.limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", s.limit)
	}

	// OFFSET
	if s.offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", s.offset)
	}

	return query, args
}

// Execute 执行查询
func (s *SelectBuilder) Execute(ctx context.Context) (*sql.Rows, error) {
	query, args := s.Build()
	return s.db.executor.QueryContext(ctx, query, args...)
}

// First 查询第一条记录
func (s *SelectBuilder) First(ctx context.Context, dest interface{}) error {
	s.Limit(1)

	// 使用QueryContext代替QueryRowContext，以获取列信息
	rows, err := s.Execute(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()

	// 检查是否有数据
	if !rows.Next() {
		return sql.ErrNoRows
	}

	// 获取列信息
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	// 使用反射进行字段映射扫描
	destValue := reflect.ValueOf(dest)
	if destValue.Kind() != reflect.Ptr {
		return fmt.Errorf("dest must be a pointer")
	}

	destValue = destValue.Elem()
	if destValue.Kind() != reflect.Struct {
		return fmt.Errorf("dest must be a pointer to struct")
	}

	destType := destValue.Type()

	// 创建字段映射
	fieldMap := buildFieldMap(destType)

	// 准备扫描目标
	scanTargets := prepareScanTargets(destValue, columns, fieldMap)

	// 扫描数据
	err = rows.Scan(scanTargets...)
	if err != nil {
		return err
	}

	// 检查遍历过程中是否有错误
	return rows.Err()
}

// Get 查询多条记录
func (s *SelectBuilder) Get(ctx context.Context, dest interface{}) error {
	rows, err := s.Execute(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = rows.Close()
	}()
	err = s.db.scanner.ScanMany(rows, dest)
	if err != nil {
		return err
	}
	return rows.Err()
}

// Count 统计数量
func (s *SelectBuilder) Count(ctx context.Context) (int64, error) {
	// 重新构建COUNT查询
	countBuilder := &SelectBuilder{
		db:      s.db,
		columns: []string{"COUNT(*)"},
		table:   s.table,
		joins:   s.joins,
		where:   s.where,
		groupBy: s.groupBy,
		having:  s.having,
	}

	query, args := countBuilder.Build()
	var count int64
	row := s.db.executor.QueryRowContext(ctx, query, args...)
	err := row.Scan(&count)
	return count, err
}

// Exists 检查是否存在
func (s *SelectBuilder) Exists(ctx context.Context) (bool, error) {
	count, err := s.Count(ctx)
	return count > 0, err
}
