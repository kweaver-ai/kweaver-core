package dbhelper2

import (
	"database/sql"
	"log"
	"reflect"

	"github.com/kweaver-ai/decision-agent/agent-factory/src/infra/common/chelper/cenvhelper"
)

func (q *SQLRunner) Delete() (res sql.Result, err error) {
	_sql, args, err := q.deleteBuilder.From(q.sb.GetFromTable()).ToDeleteSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

func (q *SQLRunner) RawExec(sqlStr string, args ...interface{}) (res sql.Result, err error) {
	res, err = q.exec(sqlStr, args)

	return
}

func (q *SQLRunner) Insert(insertFieldKVPairs map[string]interface{}) (res sql.Result, err error) {
	_sql, args, err := q.ib.Insert(insertFieldKVPairs).ToInsertSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

// InsertStruct 根据结构体来插入数据
// 结构体处理逻辑见struct2InsertPairsMapByTag的注释
func (q *SQLRunner) InsertStruct(obj interface{}) (res sql.Result, err error) {
	_sql, args, err := q.ib.InsertStruct(obj).ToInsertSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

// InsertStructs 根据结构体切片来插入数据
func (q *SQLRunner) InsertStructs(objs interface{}) (res sql.Result, err error) {
	_sql, args, err := q.ib.InsertStructs(objs).ToBatchInsertSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

//	将objs分多次插入，每次插入batchSize条记录
//
// inserts value in batches of batchSize
func (q *SQLRunner) InsertStructsInBatches(objs interface{}, batchSize int) (err error) {
	//	1. 判断objs是否是一个slice
	v := reflect.ValueOf(objs)
	if v.Kind() != reflect.Slice {
		panic("objs must be a slice")
	}

	//	2. 获取objs的长度
	length := v.Len()

	//	3. 分批插入
	for i := 0; i < length; i += batchSize {
		end := i + batchSize
		if end > length {
			end = length
		}

		batch := v.Slice(i, end)

		var (
			_sql string
			args []interface{}
		)

		_sql, args, err = q.ib.InsertStructs(batch.Interface()).ToBatchInsertSQL()
		if err != nil {
			return
		}

		_, err = q.exec(_sql, args)
		if err != nil {
			return
		}
	}

	return
}

func (q *SQLRunner) Update(updateFieldKVPairs map[string]interface{}) (res sql.Result, err error) {
	_sql, args, err := q.ub.Update(updateFieldKVPairs).ToUpdateSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

func (q *SQLRunner) UpdateByStruct(obj interface{}) (res sql.Result, err error) {
	_sql, args, err := q.ub.UpdateByStruct(obj).ToUpdateSQL()
	if err != nil {
		return
	}

	res, err = q.exec(_sql, args)

	return
}

// exec 执行sql语句
func (q *SQLRunner) exec(_sql string, args []interface{}) (res sql.Result, err error) {
	q.lastSql = _sql
	q.lastArgs = args

	if cenvhelper.IsSQLPrint() {
		log.Printf("sql: %s, \n args: %v\n", _sql, args)
		log.Printf("filled sql: [%s]", FillSQL(_sql, args...))
	}

	res, err = q.db.Exec(_sql, args...)

	return
}
