// 本方言包基于gorm v1.24.2开发，需要配合达梦数据库驱动使用
package dm

import (
	"database/sql"
	"fmt"

	_ "gitee.com/chunanyong/dm" // 引入dm数据库驱动包
	"gorm.io/gorm"              // 引入gorm v2包
	"gorm.io/gorm/callbacks"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

type Config struct {
	DriverName        string
	DSN               string
	Conn              gorm.ConnPool
	DefaultStringSize uint
}

type Dialector struct {
	*Config
}

var (
	// CreateClauses create clauses
	CreateClauses = []string{"INSERT", "VALUES", "ON CONFLICT"}
	// QueryClauses query clauses
	QueryClauses = []string{}
	// UpdateClauses update clauses
	UpdateClauses = []string{"UPDATE", "SET", "WHERE", "ORDER BY", "LIMIT"}
	// DeleteClauses delete clauses
	DeleteClauses = []string{"DELETE", "FROM", "WHERE", "ORDER BY", "LIMIT"}

	defaultDatetimePrecision = 3
)

func Open(dsn string) gorm.Dialector {
	return &Dialector{Config: &Config{DSN: dsn}}
}

func New(config Config) gorm.Dialector {
	return &Dialector{Config: &config}
}

func (d Dialector) Name() string {
	return "dm"
}

func (d Dialector) Initialize(db *gorm.DB) (err error) {
	if d.DriverName == "" {
		d.DriverName = "dm"
	}

	if d.Conn != nil {
		db.ConnPool = d.Conn
	} else {
		db.ConnPool, err = sql.Open(d.DriverName, d.DSN)
		if err != nil {
			return
		}
	}

	// register callbacks
	callbackConfig := &callbacks.Config{
		CreateClauses: CreateClauses,
		QueryClauses:  QueryClauses,
		UpdateClauses: UpdateClauses,
		DeleteClauses: DeleteClauses,
	}
	callbacks.RegisterDefaultCallbacks(db, callbackConfig)
	db.Callback().Create().Replace("gorm:create", Create)

	return
}

func (d Dialector) DefaultValueOf(*schema.Field) clause.Expression {
	// return clause.Expr{SQL: "DEFAULT VALUES"}
	// 和gorm v1不一样，gorm v1是 INSERT INTO XXXX DEFAULT VALUES;
	// gorm v2是 INSERT INTO XXXX(C1, C2) VALUES(XX, NULL);
	return clause.Expr{SQL: "NULL"}
}

func (d Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	return Migrator{
		Migrator: migrator.Migrator{
			Config: migrator.Config{
				DB:                          db,
				Dialector:                   d,
				CreateIndexAfterCreateTable: true,
			},
		},
		Dialector: d,
	}
}

func (d Dialector) BindVarTo(writer clause.Writer, stmt *gorm.Statement, v interface{}) {
	writer.WriteByte('?')
}

func (d Dialector) QuoteTo(writer clause.Writer, str string) {
	var (
		underQuoted, selfQuoted bool
		continuousBacktick      int8
		shiftDelimiter          int8
	)

	for _, v := range []byte(str) {
		switch v {
		case '"':
			continuousBacktick++
			if continuousBacktick == 2 {
				writer.WriteString(`""`)
				continuousBacktick = 0
			}
		case '.':
			if continuousBacktick > 0 || !selfQuoted {
				shiftDelimiter = 0
				underQuoted = false
				continuousBacktick = 0
				writer.WriteByte('"')
			}
			writer.WriteByte(v)
			continue
		default:
			if shiftDelimiter-continuousBacktick <= 0 && !underQuoted {
				writer.WriteByte('"')
				underQuoted = true
				if selfQuoted = continuousBacktick > 0; selfQuoted {
					continuousBacktick -= 1
				}
			}

			for ; continuousBacktick > 0; continuousBacktick -= 1 {
				writer.WriteString(`""`)
			}

			writer.WriteByte(v)
		}
		shiftDelimiter++
	}

	if continuousBacktick > 0 && !selfQuoted {
		writer.WriteString(`""`)
	}
	writer.WriteByte('"')
}

func (d Dialector) Explain(sql string, vars ...interface{}) string {
	return logger.ExplainSQL(sql, nil, `'`, vars...)
}

func (d Dialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "BIT"
	case schema.Int, schema.Uint:
		return d.getSchemaIntAndUnitType(field)
	case schema.Float:
		return d.getSchemaFloatType(field)
	case schema.String:
		return d.getSchemaStringType(field)
	case schema.Time:
		return d.getSchemaTimeType(field)
	case schema.Bytes:
		return d.getSchemaBytesType(field)
	default:
		return string(field.DataType)
		// what oracle do:
		//notNull, _ := field.TagSettings["NOT NULL"]
		//unique, _ := field.TagSettings["UNIQUE"]
		//additionalType := fmt.Sprintf("%s %s", notNull, unique)
		//if value, ok := field.TagSettings["DEFAULT"]; ok {
		//	additionalType = fmt.Sprintf("%s %s %s%s", "DEFAULT", value, additionalType, func() string {
		//		if value, ok := field.TagSettings["COMMENT"]; ok {
		//			return " COMMENT " + value
		//		}
		//		return ""
		//	}())
		//}
		//sqlType = fmt.Sprintf("%v %v", sqlType, additionalType)
	}
}

func (d Dialector) getSchemaIntAndUnitType(field *schema.Field) string {
	constraint := func(sqlType string) string {
		//if field.NotNull {
		//	sqlType += " NOT NULL"
		//}
		if field.AutoIncrement {
			sqlType += " IDENTITY(1,1)"
		}
		return sqlType
	}

	switch {
	case field.Size <= 8:
		return constraint("TINYINT")
	case field.Size <= 16:
		return constraint("SMALLINT")
	case field.Size <= 32:
		return constraint("INT")
	default:
		return constraint("BIGINT")
	}
}

func (d Dialector) getSchemaFloatType(field *schema.Field) string {
	if field.Precision > 0 {
		return fmt.Sprintf("DECIMAL(%d, %d)", field.Precision, field.Scale)
	}

	return "DOUBLE"
}

func (d Dialector) getSchemaStringType(field *schema.Field) string {
	size := field.Size

	if size == 0 {
		if d.DefaultStringSize > 0 {
			size = int(d.DefaultStringSize)
		} else {
			hasIndex := field.TagSettings["INDEX"] != "" || field.TagSettings["UNIQUE"] != ""
			// TEXT, GEOMETRY or JSON column can't have a default value
			if field.PrimaryKey || field.HasDefaultValue || hasIndex {
				size = 255 // mysql:191, dm not support utf8mb4
			}
		}
	}

	if size <= 0 {
		return "VARCHAR"
	} else if size >= 32768 {
		return "CLOB"
	} else {
		return fmt.Sprintf("VARCHAR(%d)", size)
	}
}

func (d Dialector) getSchemaTimeType(field *schema.Field) string {
	sqlType := "TIMESTAMP WITH TIME ZONE"
	//if field.NotNull || field.PrimaryKey {
	//	sqlType += " NOT NULL"
	//}
	return sqlType
}

func (d Dialector) getSchemaBytesType(field *schema.Field) string {
	if field.Size > 0 && field.Size < 32768 {
		return fmt.Sprintf("VARBINARY(%d)", field.Size)
	}

	return "BLOB"
}

func (d Dialector) SavePoint(tx *gorm.DB, name string) error {
	return tx.Exec("SAVEPOINT " + name).Error
}

func (d Dialector) RollbackTo(tx *gorm.DB, name string) error {
	return tx.Exec("ROLLBACK TO SAVEPOINT " + name).Error
}
