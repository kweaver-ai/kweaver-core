package sqlhelper2

import (
	"errors"
)

type DeleteBuilder struct {
	fromTable string
	*WhereBuilder
}

func NewDeleteBuilder() *DeleteBuilder {
	return &DeleteBuilder{
		WhereBuilder: NewWhereBuilder(),
	}
}

func (d *DeleteBuilder) SetWhereBuilder(wb *WhereBuilder) *DeleteBuilder {
	d.WhereBuilder = wb
	return d
}

func (d *DeleteBuilder) From(table string) *DeleteBuilder {
	d.fromTable = table
	return d
}

func (d *DeleteBuilder) ToDeleteSQL() (sqlStr string, args []interface{}, err error) {
	args = make([]interface{}, 0)

	if d.fromTable == "" {
		//nolint:goerr113
		err = errors.New("fromTable is empty")
		return
	}

	sqlStr = "delete from " + d.fromTable

	whereSQLStr, whereArgs, whereErr := d.WhereBuilder.ToWhereSQL()
	if whereErr != nil {
		err = whereErr
		return
	}

	if whereSQLStr != "" {
		sqlStr += " where " + whereSQLStr

		args = append(args, whereArgs...)
	}

	return
}
