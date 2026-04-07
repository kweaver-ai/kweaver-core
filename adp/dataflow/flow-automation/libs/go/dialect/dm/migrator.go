package dm

import (
	"database/sql"
	"fmt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Migrator struct {
	migrator.Migrator
	Dialector
}

func (m Migrator) AutoMigrate(dst ...interface{}) error {
	return m.Migrator.AutoMigrate(dst...)
}

func (m Migrator) CurrentDatabase() (name string) {
	m.DB.Raw("SELECT SYS_CONTEXT('USERENV', 'CURRENT_SCHEMA');").Row().Scan(&name)
	return
}

func (m Migrator) FullDataTypeOf(field *schema.Field) clause.Expr {
	expr := m.Migrator.FullDataTypeOf(field)

	if value, ok := field.TagSettings["COMMENT"]; ok {
		expr.SQL += " COMMENT " + m.Dialector.Explain("?", value)
	}

	return expr
}

func (m Migrator) GetTypeAliases(databaseTypeName string) []string {
	// super
	return m.Migrator.GetTypeAliases(databaseTypeName)
}

func (m Migrator) CreateTable(values ...interface{}) error {
	// 将`gorm:"default:true"`转为`gorm:"default:1"`
	for _, value := range values {
		if err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
			for _, v := range stmt.Schema.Fields {
				if v.HasDefaultValue {
					if vv, ok := v.DefaultValueInterface.(bool); ok {
						if vv {
							v.DefaultValueInterface = int64(1)
						} else {
							v.DefaultValueInterface = int64(0)
						}
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	// super
	return m.Migrator.CreateTable(values...)
}

func (m Migrator) DropTable(values ...interface{}) error {
	values = m.ReorderModels(values, false)
	for i := len(values) - 1; i >= 0; i-- {
		tx := m.DB.Session(&gorm.Session{})
		if err := m.RunWithValue(values[i], func(stmt *gorm.Statement) error {
			return tx.Exec("DROP TABLE IF EXISTS ? CASCADE", m.CurrentTable(stmt)).Error
		}); err != nil {
			return err
		}
	}
	return nil
}

func (m Migrator) HasTable(value interface{}) bool {
	tableSql := `SELECT /*+ MAX_OPT_N_TABLES(5) */ COUNT(TABS.NAME) FROM
(SELECT ID, PID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCH' AND NAME = ?) SCHEMAS,
(SELECT ID, SCHID, NAME FROM SYS.SYSOBJECTS WHERE
NAME = ? AND TYPE$ = 'SCHOBJ' AND SUBTYPE$ IN ('UTAB', 'STAB', 'VIEW', 'SYNOM')
AND ((SUBTYPE$ ='UTAB' AND CAST((INFO3 & 0x00FF & 0x003F) AS INT) not in (9, 27, 29, 25, 12, 7, 21, 23, 18, 5))
OR SUBTYPE$ in ('STAB', 'VIEW', 'SYNOM'))) TABS
WHERE TABS.SCHID = SCHEMAS.ID AND SF_CHECK_PRIV_OPT(UID(), CURRENT_USERTYPE(), TABS.ID, SCHEMAS.PID, -1, TABS.ID) = 1;`

	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw(tableSql, m.CurrentDatabase(), stmt.Table).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) RenameTable(oldName, newName interface{}) error {
	// super
	return m.Migrator.RenameTable(oldName, newName)
}

func (m Migrator) GetTables() (tableList []string, err error) {
	tableSql := `SELECT /*+ MAX_OPT_N_TABLES(5) */ TABS.NAME FROM
(SELECT ID, PID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCH' AND NAME = ?) SCHEMAS,
(SELECT ID, SCHID, NAME FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCHOBJ' AND SUBTYPE$ IN ('UTAB', 'STAB', 'VIEW', 'SYNOM')
AND ((SUBTYPE$ ='UTAB' AND CAST((INFO3 & 0x00FF & 0x003F) AS INT) not in (9, 27, 29, 25, 12, 7, 21, 23, 18, 5))
OR SUBTYPE$ in ('STAB', 'VIEW', 'SYNOM'))) TABS
WHERE TABS.SCHID = SCHEMAS.ID AND SF_CHECK_PRIV_OPT(UID(), CURRENT_USERTYPE(), TABS.ID, SCHEMAS.PID, -1, TABS.ID) = 1;`

	err = m.DB.Raw(tableSql, m.CurrentDatabase()).Scan(&tableList).Error
	return
}

func (m Migrator) AddColumn(dst interface{}, field string) error {
	// super
	return m.Migrator.AddColumn(dst, field)
}

func (m Migrator) DropColumn(dst interface{}, field string) error {
	// super
	return m.Migrator.DropColumn(dst, field)
}

func (m Migrator) AlterColumn(value interface{}, field string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if field := stmt.Schema.LookUpField(field); field != nil {
			return m.DB.Exec(
				"ALTER TABLE ? MODIFY ? ?",
				clause.Table{Name: stmt.Table},
				clause.Column{Name: field.DBName},
				m.FullDataTypeOf(field),
			).Error
		}
		return fmt.Errorf("failed to look up field with name: %s", field)
	})
}

func (m Migrator) MigrateColumn(dst interface{}, field *schema.Field, columnType gorm.ColumnType) error {
	// super
	return m.Migrator.MigrateColumn(dst, field, columnType)
}

func (m Migrator) HasColumn(value interface{}, field string) bool {
	columnSql := `SELECT /*+ MAX_OPT_N_TABLES(5) */ COUNT(DISTINCT COLS.NAME) FROM
(SELECT ID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCH' AND NAME = ?) SCHS,
(SELECT ID, SCHID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCHOBJ' AND SUBTYPE$ IN ('UTAB', 'STAB', 'VIEW') AND NAME = ?) TABS,
(SELECT NAME, ID FROM SYS.SYSCOLUMNS WHERE NAME = ?) COLS
WHERE TABS.ID = COLS.ID AND SCHS.ID = TABS.SCHID;`

	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Raw(columnSql, m.CurrentDatabase(), stmt.Table, field).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) RenameColumn(value interface{}, oldName, newName string) error {
	// super
	return m.Migrator.RenameColumn(value, oldName, newName)
}

func (m Migrator) ColumnTypes(dst interface{}) ([]gorm.ColumnType, error) {
	// super
	return m.Migrator.ColumnTypes(dst)
}

func (m Migrator) CreateView(name string, option gorm.ViewOption) error {
	// super, not support
	return m.Migrator.CreateView(name, option)
}

func (m Migrator) DropView(name string) error {
	// super, not support
	return m.Migrator.DropView(name)
}

func (m Migrator) CreateConstraint(dst interface{}, name string) error {
	// super
	return m.Migrator.CreateConstraint(dst, name)
}

func (m Migrator) DropConstraint(value interface{}, name string) error {
	// super
	return m.Migrator.DropConstraint(value, name)
}

func (m Migrator) HasConstraint(value interface{}, name string) bool {
	conSql := `select count(CON_OBJ.NAME) from
(select ID from SYSOBJECTS where TYPE$='SCH' and NAME = ?) SCH_OBJ, 
(select ID, SCHID from SYSOBJECTS where TYPE$='SCHOBJ' and SUBTYPE$ like '_TAB') TAB_OBJ, 
(select ID, NAME from SYSOBJECTS where SUBTYPE$ = 'CONS' and NAME=?) CON_OBJ,
SYSCONS CONS
where CON_OBJ.ID=CONS.ID and TAB_OBJ.ID=CONS.TABLEID and TAB_OBJ.SCHID=SCH_OBJ.ID;`

	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		constraint, chk, _ := m.GuessConstraintAndTable(stmt, name)
		if constraint != nil {
			name = constraint.Name
		} else if chk != nil {
			name = chk.Name
		}
		return m.DB.Raw(conSql, m.CurrentDatabase(), name).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) CreateIndex(dst interface{}, name string) error {
	// super
	return m.Migrator.CreateIndex(dst, name)
}

func (m Migrator) DropIndex(value interface{}, name string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}

		return m.DB.Exec("DROP INDEX ?", clause.Column{Name: name}).Error
	})
}

func (m Migrator) HasIndex(value interface{}, name string) bool {
	indexSql := `WITH USERS(ID) AS (SELECT ID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCH' AND NAME = ?),
TAB(ID,SCHID) AS (SELECT ID, SCHID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCHOBJ' AND SUBTYPE$ = 'UTAB' AND NAME = ?)
SELECT COUNT(DISTINCT INDEX_NAME) FROM (
SELECT /*+ MAX_OPT_N_TABLES(5) */ OBJ_INDS.NAME AS INDEX_NAME FROM USERS, TAB, SYS.SYSINDEXES AS INDS, SYS.SYSCOLUMNS AS COLS,
(SELECT ID, PID, NAME FROM SYS.SYSOBJECTS WHERE SUBTYPE$='INDEX' AND NAME = ?) OBJ_INDS
WHERE TAB.ID =COLS.ID AND TAB.ID =OBJ_INDS.PID AND INDS.ID=OBJ_INDS.ID AND TAB.SCHID= USERS.ID
AND SF_COL_IS_IDX_KEY(INDS.KEYNUM, INDS.KEYINFO, COLS.COLID)=1
UNION SELECT OBJ_INDS.NAME AS INDEX_NAME FROM USERS, TAB, SYSCONTEXTINDEXES AS OBJ_INDS, SYS.SYSCOLUMNS AS COLS
WHERE TAB.ID = COLS.ID AND TAB.ID = OBJ_INDS.TABLEID AND COLS.COLID = OBJ_INDS.COLID AND TAB.SCHID = USERS.ID AND OBJ_INDS.NAME = ?)`

	var count int64
	m.RunWithValue(value, func(stmt *gorm.Statement) error {
		if idx := stmt.Schema.LookIndex(name); idx != nil {
			name = idx.Name
		}
		return m.DB.Raw(indexSql, m.CurrentDatabase(), stmt.Schema.Table, name, name).Row().Scan(&count)
	})
	return count > 0
}

func (m Migrator) RenameIndex(value interface{}, oldName, newName string) error {
	return m.RunWithValue(value, func(stmt *gorm.Statement) error {
		return m.DB.Exec(
			"ALTER INDEX ? RENAME TO ?",
			clause.Column{Name: oldName}, clause.Column{Name: newName},
		).Error
	})
}

func (m Migrator) GetIndexes(value interface{}) ([]gorm.Index, error) {
	indexSql := `WITH USERS(ID) AS (SELECT ID FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCH' AND NAME = ?),
TAB(ID,SCHID,NAME) AS (SELECT ID, SCHID, NAME FROM SYS.SYSOBJECTS WHERE TYPE$ = 'SCHOBJ' AND SUBTYPE$ = 'UTAB' AND NAME = ?)
SELECT /*+ MAX_OPT_N_TABLES(5) */ TAB.NAME AS TABLE_NAME, COLS.NAME AS COLUMN_NAME,
OBJ_INDS.NAME AS INDEX_NAME, CASE INDS.ISUNIQUE WHEN 'Y' THEN 0 ELSE 1 END AS NON_UNIQUE,
CASE OBJ_INDS.TYPE$ WHEN 'P' THEN 1 ELSE 0 END AS IS_PRIMARY FROM USERS, TAB, SYS.SYSINDEXES AS INDS, SYS.SYSCOLUMNS AS COLS,
(SELECT INDS.ID, INDS.PID, INDS.NAME, CONS.TYPE$ FROM SYS.SYSOBJECTS AS INDS LEFT JOIN SYS.SYSCONS AS CONS ON CONS.INDEXID=INDS.ID AND SUBTYPE$='INDEX') OBJ_INDS
WHERE TAB.ID =COLS.ID AND TAB.ID =OBJ_INDS.PID AND INDS.ID=OBJ_INDS.ID AND TAB.SCHID=USERS.ID AND SF_COL_IS_IDX_KEY(INDS.KEYNUM, INDS.KEYINFO, COLS.COLID)=1
UNION SELECT TAB.NAME AS TABLE_NAME, COLS.NAME AS COLUMN_NAME, OBJ_INDS.NAME AS INDEX_NAME, 1 AS NON_UNIQUE, 0 AS IS_PRIMARY FROM
USERS, TAB, SYSCONTEXTINDEXES AS OBJ_INDS, SYS.SYSCOLUMNS AS COLS WHERE
TAB.ID = COLS.ID AND TAB.ID = OBJ_INDS.TABLEID AND COLS.COLID = OBJ_INDS.COLID AND TAB.SCHID = USERS.ID;`

	indexes := make([]gorm.Index, 0)
	err := m.RunWithValue(value, func(stmt *gorm.Statement) error {
		result := make([]*Index, 0)
		if scanErr := m.DB.Raw(indexSql, m.CurrentDatabase(), stmt.Table).Scan(&result).Error; scanErr != nil {
			return scanErr
		}
		indexMap := groupByIndexName(result)
		for _, idx := range indexMap {
			tempIdx := &migrator.Index{
				TableName: idx[0].TableName,
				NameValue: idx[0].IndexName,
				PrimaryKeyValue: sql.NullBool{
					Bool:  idx[0].Primary,
					Valid: true,
				},
				UniqueValue: sql.NullBool{
					Bool:  idx[0].NonUnique,
					Valid: true,
				},
			}
			for _, x := range idx {
				tempIdx.ColumnList = append(tempIdx.ColumnList, x.ColumnName)
			}
			indexes = append(indexes, tempIdx)
		}
		return nil
	})
	return indexes, err
}

type Index struct {
	TableName  string `gorm:"column:TABLE_NAME"`
	ColumnName string `gorm:"column:COLUMN_NAME"`
	IndexName  string `gorm:"column:INDEX_NAME"`
	NonUnique  bool   `gorm:"column:NON_UNIQUE"`
	Primary    bool   `gorm:"column:IS_PRIMARY"`
}

func groupByIndexName(indexList []*Index) map[string][]*Index {
	columnIndexMap := make(map[string][]*Index, len(indexList))
	for _, idx := range indexList {
		columnIndexMap[idx.IndexName] = append(columnIndexMap[idx.IndexName], idx)
	}
	return columnIndexMap
}
