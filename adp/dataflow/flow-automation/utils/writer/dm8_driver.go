package writer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	// DM8 and KDB drivers are now implemented in this package
	"github.com/go-sql-driver/mysql"
	dm "github.com/kweaver-ai/adp/autoflow/flow-automation/libs/go/dialect/dm"
	"gorm.io/gorm"
)

// DM8Driver DM8数据库驱动实现
type DM8Driver struct{}

func (d *DM8Driver) GetDriverName() string { return "DM8" }

func (d *DM8Driver) BuildDSN(config *DatabaseConfig) string {
	dsnConfig := mysql.NewConfig()
	dsnConfig.Addr = fmt.Sprintf("%s:%d", config.Host, config.Port)
	dsnConfig.User = config.Username
	dsnConfig.Passwd = config.Password
	dsnConfig.DBName = config.Database
	dsnConfig.Net = "tcp"

	params := d.GetDefaultParams()
	for k, v := range config.Params {
		params[k] = v
	}

	dsnConfig.Params = params
	return dsnConfig.FormatDSN()
}

func (d *DM8Driver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	// DM8 specific connection setup
	sqlDB, err := sql.Open("DM8", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open DM8 connection to %s:%d as %s (database: %s): %w",
			config.Host, config.Port, config.Username, config.Database, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dial := dm.New(dm.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM DM8 connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	return gormDB, nil
}

// CreateTableIfNotExists 创建表（如果不存在）- DM8版本
func (d *DM8Driver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate DM8 create table SQL: %w", err)
	}

	// 执行建表SQL
	if err := dbConn.Exec(createSQL).Error; err != nil {
		// 如果是表已存在的错误，忽略
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "already_exists") ||
			strings.Contains(err.Error(), "ORA-00955") || // DM8/Oracle name already used
			strings.Contains(err.Error(), "表已存在") ||
			strings.Contains(err.Error(), "对象已存在") {
			return nil
		}
		return fmt.Errorf("failed to create DM8 table: %w", err)
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - DM8版本
func (d *DM8Driver) GetFullTableName(tableInfo *TableInfo) string {
	// DM8 支持 schema
	// 如果指定了 schema，使用 schema.table 格式
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		return fmt.Sprintf("%s.%s", tableInfo.Conn.Schema, tableInfo.TableName)
	}

	// DM8 默认 schema 通常是用户名
	// 这里返回不带 schema 的表名，由连接的默认 schema 处理
	return tableInfo.TableName
}

// EscapeIdentifier 转义DM8标识符（表名、字段名等）
// DM8基于Oracle，如果标识符包含特殊字符或不符合命名规则，需要用双引号包围
func (d *DM8Driver) EscapeIdentifier(identifier string) string {
	// 检查是否需要转义
	// 1. 包含特殊字符
	// 2. 以数字开头
	// 3. 包含空格
	// 4. 是Oracle/DM8保留字
	// 5. 包含大写字母（Oracle默认大小写敏感）
	if strings.ContainsAny(identifier, "\"@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		(len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
		strings.Contains(identifier, " ") ||
		strings.ToLower(identifier) != identifier || // 包含大写字母
		d.isDM8ReservedWord(strings.ToUpper(identifier)) {
		return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
	}

	return identifier
}

// isDM8ReservedWord 检查是否为DM8/Oracle保留字
func (d *DM8Driver) isDM8ReservedWord(word string) bool {
	reservedWords := []string{
		"ACCESS", "ADD", "ALL", "ALTER", "AND", "ANY", "AS", "ASC", "AUDIT",
		"BETWEEN", "BY", "CASE", "CAST", "CHAR", "CHECK", "CLUSTER", "COLUMN",
		"COMMENT", "COMPRESS", "CONNECT", "CREATE", "CURRENT", "DATE", "DECIMAL",
		"DEFAULT", "DELETE", "DESC", "DISTINCT", "DROP", "ELSE", "EXCLUSIVE",
		"EXISTS", "FALSE", "FLOAT", "FOR", "FROM", "GRANT", "GROUP", "HAVING",
		"IDENTIFIED", "IMMEDIATE", "IN", "INCREMENT", "INDEX", "INITIAL", "INSERT",
		"INTEGER", "INTERSECT", "INTO", "IS", "LEVEL", "LIKE", "LOCK", "LONG",
		"MAXEXTENTS", "MINUS", "MODE", "MODIFY", "NOAUDIT", "NOCOMPRESS", "NOT",
		"NOWAIT", "NULL", "NUMBER", "OF", "OFFLINE", "ON", "ONLINE", "OPTION",
		"OR", "ORDER", "PCTFREE", "PRIOR", "PRIVILEGES", "PUBLIC", "RAW", "RENAME",
		"RESOURCE", "REVOKE", "ROW", "ROWID", "ROWNUM", "ROWS", "SELECT", "SESSION",
		"SET", "SHARE", "SIZE", "SMALLINT", "START", "SUCCESSFUL", "SYNONYM", "SYSDATE",
		"TABLE", "THEN", "TO", "TRIGGER", "TRUE", "UID", "UNION", "UNIQUE", "UPDATE",
		"USER", "VALIDATE", "VALUES", "VARCHAR", "VARCHAR2", "VIEW", "WHENEVER",
		"WHERE", "WITH",
	}

	for _, reserved := range reservedWords {
		if word == reserved {
			return true
		}
	}
	return false
}

// escapeCommentString 转义DM8注释字符串中的特殊字符
func (d *DM8Driver) escapeCommentString(comment string) string {
	// DM8字符串字面量中的转义规则：
	// 1. 单引号需要转义为两个单引号
	// 2. 反斜杠需要转义为两个反斜杠

	// 先转义反斜杠
	result := strings.ReplaceAll(comment, "\\", "\\\\")
	// 再转义单引号
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

func (d *DM8Driver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
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

		// Handle UNSIGNED types (DM8 doesn't support UNSIGNED, but for consistency)
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

func (d *DM8Driver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		// Numeric types
		"TINYINT":   "TINYINT",
		"SMALLINT":  "SMALLINT",
		"MEDIUMINT": "INTEGER",
		"INT":       "INTEGER",
		"INTEGER":   "INTEGER",
		"BIGINT":    "BIGINT",
		"DECIMAL":   "DECIMAL",
		"NUMERIC":   "DECIMAL",
		"FLOAT":     "FLOAT",
		"REAL":      "REAL",
		"DOUBLE":    "DOUBLE",

		// String types
		"CHAR":       "CHAR",
		"VARCHAR":    "VARCHAR",
		"STRING":     "VARCHAR",
		"TEXT":       "TEXT",
		"TINYTEXT":   "TEXT",
		"MEDIUMTEXT": "TEXT",
		"LONGTEXT":   "TEXT",

		// Binary types
		"BINARY":     "BLOB",
		"VARBINARY":  "BLOB",
		"BLOB":       "BLOB",
		"TINYBLOB":   "BLOB",
		"MEDIUMBLOB": "BLOB",
		"LONGBLOB":   "BLOB",

		// Date/Time types
		"DATE":      "DATE",
		"DATETIME":  "TIMESTAMP",
		"TIMESTAMP": "TIMESTAMP",
		"TIME":      "TIME",
		"YEAR":      "INTEGER",

		// Boolean types
		"BOOL":    "BIT",
		"BOOLEAN": "BIT",

		// Other types
		"ENUM": "VARCHAR",
		"SET":  "VARCHAR",
		"JSON": "TEXT", // DM8 doesn't have native JSON type

		// DM8 specific types
		"BIT":   "BIT",
		"IMAGE": "BLOB",
	}
}

func (d *DM8Driver) SupportSchema() bool      { return true }
func (d *DM8Driver) SupportBatchInsert() bool { return true }

func (d *DM8Driver) GetDefaultParams() map[string]string {
	return map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
	}
}

// ListTables 列出数据库中的所有表 - DM8版本
func (d *DM8Driver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// DM8中，如果没有指定schema，使用当前用户模式
	if schema == "" {
		// 使用USER_TABLES查询当前用户的表，避免关键字问题
		query := `
			SELECT
				t.TABLE_NAME as name,
				USER as schema_name,
				'TABLE' as type,
				NVL(c.COMMENTS, '') as table_comment
			FROM USER_TABLES t
			LEFT JOIN USER_TAB_COMMENTS c ON c.TABLE_NAME = t.TABLE_NAME
			ORDER BY t.TABLE_NAME
		`

		rows, err := dbConn.Raw(query).Rows()
		if err != nil {
			return nil, fmt.Errorf("failed to query DM8 user tables: %w", err)
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
				return nil, fmt.Errorf("failed to scan DM8 table row: %w", err)
			}
			tables = append(tables, table)
		}

		return tables, nil
	} else {
		// 查询指定schema的表
		query := `
			SELECT
				t.TABLE_NAME as name,
				t.OWNER as schema_name,
				'TABLE' as type,
				NVL(c.COMMENTS, '') as table_comment
			FROM ALL_TABLES t
			LEFT JOIN ALL_TAB_COMMENTS c ON c.OWNER = t.OWNER AND c.TABLE_NAME = t.TABLE_NAME
			WHERE t.OWNER = UPPER(?)
			ORDER BY t.TABLE_NAME
		`

		rows, err := dbConn.Raw(query, schema).Rows()
		if err != nil {
			return nil, fmt.Errorf("failed to query DM8 tables: %w", err)
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
				return nil, fmt.Errorf("failed to scan DM8 table row: %w", err)
			}
			tables = append(tables, table)
		}

		return tables, nil
	}
}

// ListTableColumns 列出指定表的字段信息 - DM8版本
func (d *DM8Driver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	// DM8中，如果没有指定schema，使用USER_TAB_COLUMNS查询当前用户的表
	if schema == "" {
		query := fmt.Sprintf(`
			SELECT
				COLUMN_NAME as name,
				DATA_TYPE as data_type,
				DATA_LENGTH as data_length,
				NULLABLE as is_nullable,
				0 as primary_key,
				DATA_DEFAULT as default_value,
				'' as column_comment,
				DATA_PRECISION as precision,
				DATA_SCALE as scale
			FROM USER_TAB_COLUMNS
			WHERE TABLE_NAME = '%s'
			ORDER BY COLUMN_ID
		`, strings.ToUpper(tableName))

		rows, err := dbConn.Raw(query).Rows()
		if err != nil {
			return nil, fmt.Errorf("failed to query DM8 user table columns: %w", err)
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
				return nil, fmt.Errorf("failed to scan DM8 table column row: %w", err)
			}

			col.DataLength = dataLengthScanner.Int()
			col.DefaultValue = defaultValueScanner.String()
			col.Comment = commentScanner.String()
			col.Precision = precisionScanner.Int()
			col.Scale = scaleScanner.Int()

			columns = append(columns, col)
		}

		return columns, nil
	} else {
		// 查询指定schema的表
		query := fmt.Sprintf(`
			SELECT
				COLUMN_NAME as name,
				DATA_TYPE as data_type,
				DATA_LENGTH as data_length,
				NULLABLE as is_nullable,
				0 as primary_key,
				DATA_DEFAULT as default_value,
				'' as column_comment,
				DATA_PRECISION as precision,
				DATA_SCALE as scale
			FROM ALL_TAB_COLUMNS
			WHERE OWNER = '%s' AND TABLE_NAME = '%s'
			ORDER BY COLUMN_ID
		`, strings.ToUpper(schema), strings.ToUpper(tableName))

		rows, err := dbConn.Raw(query).Rows()
		if err != nil {
			return nil, fmt.Errorf("failed to query DM8 table columns: %w", err)
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
				return nil, fmt.Errorf("failed to scan DM8 table column row: %w", err)
			}

			col.DataLength = dataLengthScanner.Int()
			col.DefaultValue = defaultValueScanner.String()
			col.Comment = commentScanner.String()
			col.Precision = precisionScanner.Int()
			col.Scale = scaleScanner.Int()

			columns = append(columns, col)
		}

		return columns, nil
	}
}

// normalizeDM8DataType 将DM8的数据类型转换为更易理解的格式
func normalizeDM8DataType(dataType string) string {
	switch strings.ToLower(dataType) {
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
	case "timestamp with time zone":
		return "timestamptz"
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
	case "integer":
		return "int"
	case "smallint":
		return "smallint"
	case "bigint":
		return "bigint"
	case "tinyint":
		return "tinyint"
	case "boolean":
		return "bool"
	case "decimal":
		return "decimal"
	case "numeric":
		return "decimal"
	case "real":
		return "real"
	case "double precision":
		return "double"
	case "text":
		return "text"
	case "time":
		return "time"
	case "timetz":
		return "timetz"
	default:
		return dataType
	}
}
