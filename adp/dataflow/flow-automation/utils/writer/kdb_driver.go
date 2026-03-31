package writer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	kdb "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/dialect/kdb"
	"gorm.io/gorm"
)

// KDBDriver KDB数据库驱动实现
type KDBDriver struct{}

func (d *KDBDriver) GetDriverName() string { return "kingbase" }

func (d *KDBDriver) BuildDSN(config *DatabaseConfig) string {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		config.Host, config.Port, config.Username, config.Password, config.Database)

	for k, v := range config.Params {
		dsn += fmt.Sprintf(" %s=%s", k, v)
	}

	// 默认设置sslmode为disable，除非明确指定
	if _, hasSSLMode := config.Params["sslmode"]; !hasSSLMode {
		dsn += " sslmode=disable"
	}

	return dsn
}

func (d *KDBDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	// KDB specific connection setup
	sqlDB, err := sql.Open("KDB", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open KDB connection to %s:%d as %s (database: %s): %w",
			config.Host, config.Port, config.Username, config.Database, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dial := kdb.New(kdb.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM KDB connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	return gormDB, nil
}

// CreateTableIfNotExists 创建表（如果不存在）- KDB版本
func (d *KDBDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate KDB create table SQL: %w", err)
	}

	// 执行建表SQL
	if err := dbConn.Exec(createSQL).Error; err != nil {
		// 如果是表已存在的错误，忽略
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "already_exists") ||
			strings.Contains(err.Error(), "table already exists") ||
			strings.Contains(err.Error(), "表已存在") {
			return nil
		}
		return fmt.Errorf("failed to create KDB table: %w", err)
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - KDB版本
func (d *KDBDriver) GetFullTableName(tableInfo *TableInfo) string {
	// KDB 支持 schema
	// 如果指定了 schema，使用 schema.table 格式
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		return fmt.Sprintf("%s.%s", tableInfo.Conn.Schema, tableInfo.TableName)
	}

	// KDB 默认 schema 处理
	return tableInfo.TableName
}

// EscapeIdentifier 转义KDB标识符（表名、字段名等）
// KDB中，如果标识符包含特殊字符或不符合命名规则，需要用反引号包围
func (d *KDBDriver) EscapeIdentifier(identifier string) string {
	// 检查是否需要转义
	// 1. 包含特殊字符
	// 2. 以数字开头
	// 3. 包含空格
	// 4. 是KDB保留字
	if strings.ContainsAny(identifier, "`@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		(len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
		strings.Contains(identifier, " ") ||
		d.isKDBReservedWord(strings.ToUpper(identifier)) {
		return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
	}

	return identifier
}

// isKDBReservedWord 检查是否为KDB保留字
func (d *KDBDriver) isKDBReservedWord(word string) bool {
	// KDB保留字相对较少，主要是一些SQL标准的保留字
	reservedWords := []string{
		"SELECT", "FROM", "WHERE", "INSERT", "UPDATE", "DELETE",
		"CREATE", "DROP", "ALTER", "TABLE", "INDEX", "VIEW",
		"PRIMARY", "FOREIGN", "KEY", "CONSTRAINT", "UNIQUE",
		"NOT", "NULL", "DEFAULT", "AUTO_INCREMENT",
		"AND", "OR", "IN", "LIKE", "BETWEEN", "IS",
		"ORDER", "BY", "GROUP", "HAVING", "LIMIT",
		"JOIN", "INNER", "LEFT", "RIGHT", "FULL", "OUTER",
		"UNION", "ALL", "DISTINCT", "AS", "ON",
	}

	for _, reserved := range reservedWords {
		if word == reserved {
			return true
		}
	}
	return false
}

// escapeCommentString 转义KDB注释字符串中的特殊字符
func (d *KDBDriver) escapeCommentString(comment string) string {
	// KDB字符串字面量中的转义规则：
	// 1. 单引号需要转义为两个单引号
	// 2. 反斜杠需要转义为两个反斜杠

	// 先转义反斜杠
	result := strings.ReplaceAll(comment, "\\", "\\\\")
	// 再转义单引号
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

func (d *KDBDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
	if len(tableInfo.Fields) == 0 {
		return "", fmt.Errorf("no field mappings provided")
	}

	// Get full table name
	fullTableName := d.GetFullTableName(tableInfo)

	var sql strings.Builder
	sql.WriteString("CREATE TABLE IF NOT EXISTS ")
	sql.WriteString(fullTableName)
	sql.WriteString(" (")

	columns := make([]string, 0, len(tableInfo.Fields))
	for i := range tableInfo.Fields {
		field := &tableInfo.Fields[i]
		if field.Target.Name == "" {
			continue
		}

		var columnSQL strings.Builder
		columnSQL.WriteString(d.EscapeIdentifier(field.Target.Name))
		columnSQL.WriteString(" ")

		// Use the driver's data type mapping
		mapping := d.GetDataTypeMapping()
		dataType := strings.ToUpper(field.Target.DataType)

		// Handle UNSIGNED types (KDB doesn't support UNSIGNED, but for consistency)
		baseDataType := strings.TrimSuffix(dataType, " UNSIGNED")
		baseDataType = strings.TrimSuffix(baseDataType, "UNSIGNED")

		if mappedType, exists := mapping[baseDataType]; exists {
			dataType = mappedType
		} else {
			// If no mapping found, use the original type
			dataType = field.Target.DataType
		}

		// Handle length for VARCHAR and CHAR types
		if (strings.HasPrefix(strings.ToUpper(dataType), "VARCHAR") ||
			strings.HasPrefix(strings.ToUpper(dataType), "CHAR")) &&
			field.Target.DataLenth > 0 {
			dataType = fmt.Sprintf("%s(%d)", dataType, field.Target.DataLenth)
		}

		// Handle DECIMAL precision
		if strings.EqualFold(dataType, "DECIMAL") && field.Target.Precision >= 0 {
			scale := field.Target.DataLenth
			if scale <= 0 {
				scale = DefaultDecimalScale
			}
			if scale < field.Target.Precision {
				scale = field.Target.Precision
			}
			dataType = fmt.Sprintf("DECIMAL(%d,%d)", scale, field.Target.Precision)
		}

		columnSQL.WriteString(dataType)

		if field.Target.IsNullable == "NO" {
			columnSQL.WriteString(" NOT NULL")
		}

		if field.Target.PrimaryKey == PrimaryKeyFlag {
			columnSQL.WriteString(" PRIMARY KEY")
		}

		if field.Target.Comment != "" {
			columnSQL.WriteString(" COMMENT '")
			columnSQL.WriteString(d.escapeCommentString(field.Target.Comment))
			columnSQL.WriteString("'")
		}

		columns = append(columns, columnSQL.String())
	}

	sql.WriteString(strings.Join(columns, ", "))
	sql.WriteString(")")

	return sql.String(), nil
}

func (d *KDBDriver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		"TINYINT":    "TINYINT",
		"SMALLINT":   "SMALLINT",
		"MEDIUMINT":  "INTEGER",
		"INT":        "INTEGER",
		"INTEGER":    "INTEGER",
		"BIGINT":     "BIGINT",
		"DECIMAL":    "DECIMAL",
		"FLOAT":      "FLOAT",
		"DOUBLE":     "DOUBLE",
		"CHAR":       "CHAR",
		"VARCHAR":    "VARCHAR",
		"STRING":     "VARCHAR",
		"TEXT":       "TEXT",
		"TINYTEXT":   "TEXT",
		"MEDIUMTEXT": "TEXT",
		"LONGTEXT":   "TEXT",
		"BINARY":     "BLOB",
		"VARBINARY":  "BLOB",
		"BLOB":       "BLOB",
		"TINYBLOB":   "BLOB",
		"MEDIUMBLOB": "BLOB",
		"LONGBLOB":   "BLOB",
		"DATE":       "DATE",
		"DATETIME":   "TIMESTAMP",
		"TIMESTAMP":  "TIMESTAMP",
		"TIME":       "TIME",
		"YEAR":       "INTEGER",
		"BOOL":       "BOOLEAN",
		"BOOLEAN":    "BOOLEAN",
		"ENUM":       "VARCHAR",
		"SET":        "VARCHAR",
		"JSON":       "TEXT", // KDB may not have native JSON type
	}
}

func (d *KDBDriver) SupportSchema() bool      { return true }
func (d *KDBDriver) SupportBatchInsert() bool { return true }

func (d *KDBDriver) GetDefaultParams() map[string]string {
	return map[string]string{
		"charset":   "utf8",
		"parseTime": "true",
	}
}

// ListTables 列出数据库中的所有表 - KDB版本
func (d *KDBDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// KDB中，schema就是数据库名
	// 如果没有指定schema，使用当前数据库
	if schema == "" {
		schema = dbConn.Migrator().CurrentDatabase()
	}

	query := `
		SELECT
			TABLE_NAME as name,
			TABLE_SCHEMA as schema,
			TABLE_TYPE as type,
			COALESCE(TABLE_COMMENT, '') as comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := dbConn.Raw(query, schema).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query KDB tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table TableMetadata
		err := rows.Scan(
			&table.Name,
			&table.Schema,
			&table.Type,
			&table.Comment,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan KDB table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// ListTableColumns 列出指定表的字段信息 - KDB版本
func (d *KDBDriver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	// KDB中，如果没有指定schema，使用public
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			c.column_name as name,
			c.data_type as data_type,
			COALESCE(CAST(c.character_maximum_length AS BIGINT), CAST(c.numeric_precision AS BIGINT)) as data_length,
			CASE WHEN c.is_nullable = 'YES' THEN 'YES' ELSE 'NO' END as is_nullable,
			CASE WHEN pk.constraint_name IS NOT NULL THEN 1 ELSE 0 END as primary_key,
			c.column_default as default_value,
			COALESCE(d.description, '') as comment,
			c.numeric_precision as precision,
			c.numeric_scale as scale
		FROM information_schema.columns c
		LEFT JOIN information_schema.key_column_usage kcu
			ON c.table_schema = kcu.table_schema
			AND c.table_name = kcu.table_name
			AND c.column_name = kcu.column_name
		LEFT JOIN information_schema.table_constraints pk
			ON kcu.constraint_name = pk.constraint_name
			AND kcu.table_schema = pk.table_schema
			AND kcu.table_name = pk.table_name
			AND pk.constraint_type = 'PRIMARY KEY'
		LEFT JOIN pg_class pc ON pc.relname = c.table_name
		LEFT JOIN pg_description d ON d.objoid = pc.oid AND d.objsubid = c.ordinal_position
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position
	`

	rows, err := dbConn.Raw(query, schema, tableName).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query KDB table columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var defaultValueScanner nullStringScanner
		var commentScanner nullStringScanner
		var dataLengthScanner nullIntScanner
		var precisionScanner nullIntScanner
		var scaleScanner nullIntScanner

		err := rows.Scan(
			&col.Name,
			&col.DataType,
			&dataLengthScanner,
			&col.IsNullable,
			&col.PrimaryKey,
			&defaultValueScanner,
			&commentScanner,
			&precisionScanner,
			&scaleScanner,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan KDB column row: %w", err)
		}

		// 将NULL值转换为适当的格式
		col.DefaultValue = defaultValueScanner.String()
		col.Comment = commentScanner.String()
		col.DataLength = dataLengthScanner.Int()
		col.Precision = precisionScanner.Int()
		col.Scale = scaleScanner.Int()

		// 将KDB的数据类型转换为更易理解的格式
		col.DataType = normalizeKDBDataType(col.DataType)

		columns = append(columns, col)
	}

	return columns, nil
}

// normalizeKDBDataType 将KDB的数据类型转换为更易理解的格式
func normalizeKDBDataType(dataType string) string {
	switch strings.ToLower(dataType) {
	case "character varying":
		return "varchar"
	case "character":
		return "char"
	case "double precision":
		return "double"
	case "timestamp without time zone":
		return "timestamp"
	case "timestamp with time zone":
		return "timestamptz"
	case "time without time zone":
		return "time"
	case "time with time zone":
		return "timetz"
	case "boolean":
		return "bool"
	case "integer":
		return "int"
	case "smallint":
		return "smallint"
	case "bigint":
		return "bigint"
	case "real":
		return "real"
	case "numeric":
		return "decimal"
	case "varchar2":
		return "varchar"
	case "nvarchar2":
		return "nvarchar"
	case "number":
		return "decimal"
	case "float":
		return "float"
	case "binary_float":
		return "float"
	case "binary_double":
		return "double"
	case "date":
		return "date"
	case "timestamp":
		return "timestamp"
	case "timestamp with local time zone":
		return "timestampltz"
	case "interval year to month":
		return "interval_ym"
	case "interval day to second":
		return "interval_ds"
	case "char":
		return "char"
	case "nchar":
		return "nchar"
	case "clob":
		return "clob"
	case "nclob":
		return "nclob"
	case "blob":
		return "blob"
	case "bfile":
		return "bfile"
	case "long":
		return "long"
	case "raw":
		return "raw"
	case "rowid":
		return "rowid"
	case "urowid":
		return "urowid"
	case "tinyint":
		return "tinyint"
	case "mediumint":
		return "mediumint"
	case "decimal":
		return "decimal"
	case "text":
		return "text"
	case "binary":
		return "binary"
	case "varbinary":
		return "varbinary"
	case "tinyblob":
		return "tinyblob"
	case "mediumblob":
		return "mediumblob"
	case "longblob":
		return "longblob"
	case "tinytext":
		return "tinytext"
	case "mediumtext":
		return "mediumtext"
	case "longtext":
		return "longtext"
	case "enum":
		return "enum"
	case "set":
		return "set"
	case "datetime":
		return "datetime"
	case "smalldatetime":
		return "smalldatetime"
	case "datetime2":
		return "datetime2"
	case "datetimeoffset":
		return "datetimeoffset"
	case "rowversion":
		return "rowversion"
	case "uniqueidentifier":
		return "uniqueidentifier"
	case "xml":
		return "xml"
	case "sql_variant":
		return "sql_variant"
	default:
		return dataType
	}
}
