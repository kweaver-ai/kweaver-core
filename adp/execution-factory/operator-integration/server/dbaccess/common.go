// Package dbaccess
// @file common.go
// @description: 数据库操作公共方法
package dbaccess

import (
	"database/sql"
	"fmt"
)

func checkAffected(rs sql.Result) (ok bool, err error) {
	var ac int64
	ac, err = rs.RowsAffected()
	if err != nil {
		err = fmt.Errorf("row end err:%v", err)
		return
	}
	if ac > 0 {
		ok = true
	}
	return
}

func checkHasQuery(query string, err error) (bool, error) {
	var has bool
	switch err {
	case nil:
		has = true
	case sql.ErrNoRows:
		err = nil
	default:
		err = fmt.Errorf("sql exec %s err:%v", query, err)
	}
	return has, err
}

func checkHasQueryErr(err error) (bool, error) {
	var has bool
	switch err {
	case nil:
		has = true
	case sql.ErrNoRows:
		err = nil
	default:
		err = fmt.Errorf("sql exec err:%v", err)
	}
	return has, err
}
