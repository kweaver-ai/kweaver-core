package dbhelper2

import (
	"database/sql"
	"errors"
	"log"
	"reflect"

	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/cenvhelper"
	"github.com/kweaver-ai/kweaver-core/decision-agent/agent-backend/agent-factory/src/infra/common/chelper/sqlhelper2"
)

// FindOne 查询一条记录
// 【注意】：注意select字段和obj的字段要一一对应，否则可能导致数据对不上
func (q *SQLRunner) FindOne(obj interface{}) (err error) {
	q.sb.Offset(0)
	q.sb.Limit(1)

	if len(q.sb.SelectFields()) == 0 && q.rawSQL == "" {
		q.sb.Select(sqlhelper2.AllFieldsByStruct(obj, q.tag))
	}

	var (
		_sql string
		args []interface{}
	)

	if q.rawSQL != "" {
		_sql = q.rawSQL
		args = q.rawSQLArgs
	} else {
		_sql, args, err = q.sb.ToSelectSQL()
		if err != nil {
			return
		}
	}

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	opt := Option{
		SelectFields: q.sb.SelectFields(),
	}
	err = q.db.QueryRow(_sql, args...).
		Scan(q.struct2ScanArgsByTag(obj, opt)...)

	return
}

func (q *SQLRunner) Scan(obj ...interface{}) (err error) {
	if len(q.sb.SelectFields()) == 0 {
		panic("Scan must be called after Select")
	}

	var (
		_sql string
		args []interface{}
	)

	if q.rawSQL != "" {
		_sql = q.rawSQL
		args = q.rawSQLArgs
	} else {
		_sql, args, err = q.sb.ToSelectSQL()
		if err != nil {
			return
		}
	}

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	err = q.db.QueryRow(_sql, args...).
		Scan(obj...)

	return
}

// Find 查询多条记录
// 注意：
// 1. 暂不支持指针切片，如[]*dccpo.DccDocLibJoin
// 2. 注意select字段和objSlice元素的字段要一一对应，否则可能导致数据对不上
func (q *SQLRunner) Find(objSlice interface{}) (err error) {
	sliceType := reflect.TypeOf(objSlice)

	if sliceType.Kind() == reflect.Ptr {
		sliceType = sliceType.Elem()
	}

	if sliceType.Kind() != reflect.Slice {
		panic("objSlice must be a slice")
	}

	elementType := sliceType.Elem()

	var (
		_sql string
		args []interface{}
	)

	if q.rawSQL != "" {
		_sql = q.rawSQL
		args = q.rawSQLArgs
	} else {
		if len(q.sb.SelectFields()) == 0 && q.po != nil {
			q.sb.Select(sqlhelper2.AllFieldsByStruct(q.po, q.tag))
		}

		_sql, args, err = q.sb.ToSelectSQL()
		if err != nil {
			return
		}
	}

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	rows, err := q.db.Query(_sql, args...)
	defer chelper.CloseRows(rows, q.logger)

	if err != nil {
		return
	}

	opt := Option{
		SelectFields: q.sb.SelectFields(),
	}

	for rows.Next() {
		obj := reflect.New(elementType).Interface()
		// vv := reflect.New(elementType).Elem()
		// obj:=createStructWithInitializedAnonymousField(vv.Interface())
		err = rows.Scan(q.struct2ScanArgsByTag(obj, opt)...)
		if err != nil {
			return
		}

		reflect.ValueOf(objSlice).Elem().
			Set(reflect.Append(reflect.ValueOf(objSlice).Elem(), reflect.ValueOf(obj).Elem()))
	}

	return
}

// FindColumn 查询某一列的值
// 【注意】不确定是否有问题，慎用
// todo objSlice清空原有数据
func (q *SQLRunner) FindColumn(columnName string, objSlice interface{}) (err error) {
	typ := reflect.TypeOf(objSlice)

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	} else {
		panic("[FindColumn]: objSlice must be a pointer to slice")
	}

	if typ.Kind() != reflect.Slice {
		panic("[FindColumn]: objSlice must be a slice")
	}

	elementType := typ.Elem()

	var (
		_sql string
		args []interface{}
	)

	if q.rawSQL != "" {
		_sql = q.rawSQL
		args = q.rawSQLArgs
	} else {
		_sql, args, err = q.sb.Select([]string{columnName}).
			ToSelectSQL()
		if err != nil {
			return
		}
	}

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	rows, err := q.db.Query(_sql, args...)
	defer chelper.CloseRows(rows, q.logger)

	if err != nil {
		return
	}

	for rows.Next() {
		obj := reflect.New(elementType).Interface()

		err = rows.Scan(obj)
		if err != nil {
			return
		}

		reflect.ValueOf(objSlice).Elem().
			Set(reflect.Append(reflect.ValueOf(objSlice).Elem(), reflect.ValueOf(obj).Elem()))
	}

	return
}

func (q *SQLRunner) Count() (total int64, err error) {
	var (
		_sql string
		args []interface{}
	)

	if q.rawSQL != "" {
		_sql = q.rawSQL
		args = q.rawSQLArgs
	} else {
		_sql, args, err = q.sb.Select([]string{"count(*)"}).ToSelectSQL()
		if err != nil {
			return
		}
	}

	q.lastSql = _sql
	q.lastArgs = args

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	err = q.db.QueryRow(_sql, args...).
		Scan(&total)
	if errors.Is(err, sql.ErrNoRows) {
		err = nil
	}

	return
}

// Exists 判断记录是否存在
func (q *SQLRunner) Exists() (exists bool, err error) {
	q.sb.Select([]string{"1"})
	q.sb.Limit(1)

	_sql, args, err := q.sb.ToSelectSQL()
	if err != nil {
		return
	}

	q.lastSql = _sql
	q.lastArgs = args

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	var aa int

	err = q.db.QueryRow(_sql, args...).Scan(&aa)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}

		err = nil

		return
	}

	if aa == 1 {
		exists = true
	}

	return
}

func (q *SQLRunner) RawExists(rawSql string, args ...interface{}) (exists bool, err error) {
	q.lastSql = rawSql
	q.lastArgs = args

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", rawSql, args)
		log.Printf("filled sql: [%s]", FillSQL(rawSql, args...))
	}

	var aa int

	err = q.db.QueryRow(rawSql, args...).Scan(&aa)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return
		}

		err = nil

		return
	}

	if aa == 1 {
		exists = true
	}

	return
}
