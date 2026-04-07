package dm

import (
	"database/sql"
	"gorm.io/gorm"
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"reflect"
)

func Create(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	if db.Statement.Schema != nil && !db.Statement.Unscoped {
		for _, c := range db.Statement.Schema.CreateClauses {
			db.Statement.AddClause(c)
		}
	}

	var onConflict clause.OnConflict
	var hasConflict bool
	if db.Statement.SQL.String() == "" {
		var (
			values = callbacks.ConvertToCreateValues(db.Statement)
			c      = db.Statement.Clauses["ON CONFLICT"]
		)
		onConflict, hasConflict = c.Expression.(clause.OnConflict)

		if hasConflict {
			if len(db.Statement.Schema.PrimaryFields) > 0 {
				columnsMap := map[string]bool{}
				for _, column := range values.Columns {
					columnsMap[column.Name] = true
				}

				for _, field := range db.Statement.Schema.PrimaryFields {
					if _, ok := columnsMap[field.DBName]; !ok {
						hasConflict = false
					}
				}
			} else {
				hasConflict = false
			}
		}

		if hasConflict {
			MergeCreate(db, onConflict, values)
		} else {
			setIdentityInsert := false

			if db.Statement.Schema != nil {
				if field := db.Statement.Schema.PrioritizedPrimaryField; field != nil && field.AutoIncrement {
					switch db.Statement.ReflectValue.Kind() {
					case reflect.Struct:
						_, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
						setIdentityInsert = !isZero
					case reflect.Slice, reflect.Array:
						for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
							obj := db.Statement.ReflectValue.Index(i)
							if reflect.Indirect(obj).Kind() == reflect.Struct {
								_, isZero := field.ValueOf(db.Statement.Context, db.Statement.ReflectValue.Index(i))
								setIdentityInsert = !isZero
							}
							break
						}
					}

					if setIdentityInsert && !db.DryRun && db.Error == nil {
						db.Statement.SQL.Reset()
						db.Statement.WriteString("SET IDENTITY_INSERT ")
						db.Statement.WriteQuoted(db.Statement.Table)
						db.Statement.WriteString(" ON;")
						_, err := db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
						if db.AddError(err) != nil {
							return
						}
						defer func() {
							db.Statement.SQL.Reset()
							db.Statement.WriteString("SET IDENTITY_INSERT ")
							db.Statement.WriteQuoted(db.Statement.Table)
							db.Statement.WriteString(" OFF;")
							db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
						}()
					}
				}
			}

			db.Statement.SQL.Reset()
			db.Statement.AddClauseIfNotExists(clause.Insert{})
			db.Statement.Build("INSERT")
			db.Statement.WriteByte(' ')

			db.Statement.AddClause(values)
			if values, ok := db.Statement.Clauses["VALUES"].Expression.(clause.Values); ok {
				if len(values.Columns) > 0 {
					db.Statement.WriteByte('(')
					for idx, column := range values.Columns {
						if idx > 0 {
							db.Statement.WriteByte(',')
						}
						db.Statement.WriteQuoted(column)
					}
					db.Statement.WriteByte(')')

					//outputInserted(db)

					db.Statement.WriteString(" VALUES ")

					for idx, value := range values.Values {
						if idx > 0 {
							db.Statement.WriteByte(',')
						}

						db.Statement.WriteByte('(')
						db.Statement.AddVar(db.Statement, value...)
						db.Statement.WriteByte(')')
					}

					db.Statement.WriteString(";")
				} else {
					db.Statement.WriteString("DEFAULT VALUES;")
				}
			}
		}
	}

	if !db.DryRun && db.Error == nil {
		var (
			rows           *sql.Rows
			result         sql.Result
			err            error
			updateInsertID bool  // 是否需要更新主键自增列
			insertID       int64 // 主键自增列最新值
		)
		if hasConflict {
			rows, err = db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if db.AddError(err) != nil {
				return
			}
			defer rows.Close()
			if rows.Next() {
				rows.Scan(&insertID)
				if insertID > 0 {
					updateInsertID = true
				}
			}
		} else {
			result, err = db.Statement.ConnPool.ExecContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
			if db.AddError(err) != nil {
				return
			}
			db.RowsAffected, _ = result.RowsAffected()
			if db.RowsAffected != 0 && db.Statement.Schema != nil &&
				db.Statement.Schema.PrioritizedPrimaryField != nil &&
				db.Statement.Schema.PrioritizedPrimaryField.HasDefaultValue {
				insertID, err = result.LastInsertId()
				insertOk := err == nil && insertID > 0
				if !insertOk {
					db.AddError(err)
					return
				}
				updateInsertID = true
			}
		}

		if updateInsertID {
			switch db.Statement.ReflectValue.Kind() {
			case reflect.Slice, reflect.Array:
				//if config.LastInsertIDReversed {
				for i := db.Statement.ReflectValue.Len() - 1; i >= 0; i-- {
					rv := db.Statement.ReflectValue.Index(i)
					if reflect.Indirect(rv).Kind() != reflect.Struct {
						break
					}

					_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv)
					if isZero {
						db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
						insertID -= db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
					}
				}
				//} else {
				//	for i := 0; i < db.Statement.ReflectValue.Len(); i++ {
				//		rv := db.Statement.ReflectValue.Index(i)
				//		if reflect.Indirect(rv).Kind() != reflect.Struct {
				//			break
				//		}
				//
				//		if _, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, rv); isZero {
				//			db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, rv, insertID))
				//			insertID += db.Statement.Schema.PrioritizedPrimaryField.AutoIncrementIncrement
				//		}
				//	}
				//}
			case reflect.Struct:
				_, isZero := db.Statement.Schema.PrioritizedPrimaryField.ValueOf(db.Statement.Context, db.Statement.ReflectValue)
				if isZero {
					db.AddError(db.Statement.Schema.PrioritizedPrimaryField.Set(db.Statement.Context, db.Statement.ReflectValue, insertID))
				}
			}
		}
	}
}

func MergeCreate(db *gorm.DB, onConflict clause.OnConflict, values clause.Values) {
	db.Statement.WriteString("MERGE INTO ")
	db.Statement.WriteQuoted(db.Statement.Table)
	db.Statement.WriteString(" USING (")
	for idx, value := range values.Values {
		if idx > 0 {
			db.Statement.WriteString(" UNION ALL ")
		}

		db.Statement.WriteString("SELECT ")
		db.Statement.AddVar(db.Statement, value...)
		db.Statement.WriteString(" FROM DUAL")
	}

	db.Statement.WriteString(") AS \"excluded\" (")
	for idx, column := range values.Columns {
		if idx > 0 {
			db.Statement.WriteByte(',')
		}
		db.Statement.WriteQuoted(column.Name)
	}
	db.Statement.WriteString(") ON ")

	var where clause.Where
	for _, field := range db.Statement.Schema.PrimaryFields {
		where.Exprs = append(where.Exprs, clause.Eq{
			Column: clause.Column{Table: db.Statement.Table, Name: field.DBName},
			Value:  clause.Column{Table: "excluded", Name: field.DBName},
		})
	}
	where.Build(db.Statement)

	if len(onConflict.DoUpdates) > 0 {
		// 将UPDATE子句中出现在关联条件中的列去除（即上面的ON子句），否则会报错：-4064:不能更新关联条件中的列
		var withoutOnColumns = make([]clause.Assignment, 0, len(onConflict.DoUpdates))
		a : for _, assignment := range onConflict.DoUpdates {
			for _, field := range db.Statement.Schema.PrimaryFields {
				if assignment.Column.Name == field.DBName {
					continue a
				}
			}
			withoutOnColumns = append(withoutOnColumns, assignment)
		}
		onConflict.DoUpdates = clause.Set(withoutOnColumns)
		if len(onConflict.DoUpdates) > 0 {
			db.Statement.WriteString(" WHEN MATCHED THEN UPDATE SET ")
			onConflict.DoUpdates.Build(db.Statement)
		}
	}

	db.Statement.WriteString(" WHEN NOT MATCHED THEN INSERT (")

	written := false
	for _, column := range values.Columns {
		if db.Statement.Schema.PrioritizedPrimaryField == nil || !db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement || db.Statement.Schema.PrioritizedPrimaryField.DBName != column.Name {
			if written {
				db.Statement.WriteByte(',')
			}
			written = true
			db.Statement.WriteQuoted(column.Name)
		}
	}

	db.Statement.WriteString(") VALUES (")

	written = false
	for _, column := range values.Columns {
		if db.Statement.Schema.PrioritizedPrimaryField == nil || !db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement || db.Statement.Schema.PrioritizedPrimaryField.DBName != column.Name {
			if written {
				db.Statement.WriteByte(',')
			}
			written = true
			db.Statement.WriteQuoted(clause.Column{
				Table: "excluded",
				Name:  column.Name,
			})
		}
	}

	db.Statement.WriteString(")")
	//outputInserted(db)
	db.Statement.WriteString(";")

	// merge into 语句插入的记录，无法通过LastInsertID获取
	if db.Statement.Schema.PrioritizedPrimaryField != nil && db.Statement.Schema.PrioritizedPrimaryField.AutoIncrement {
		db.Statement.WriteString("SELECT ")
		db.Statement.WriteQuoted(db.Statement.Schema.PrioritizedPrimaryField.DBName)
		db.Statement.WriteString(" FROM ")
		db.Statement.WriteQuoted(db.Statement.Table)
		db.Statement.WriteString(" ORDER BY ")
		db.Statement.WriteQuoted(db.Statement.Schema.PrioritizedPrimaryField.DBName)
		db.Statement.WriteString(" DESC LIMIT 1;")
	}
}
