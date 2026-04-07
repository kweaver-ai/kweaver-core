package writer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/microsoft/go-mssqldb"
	sqlserverd "gorm.io/driver/sqlserver"
	"gorm.io/gorm"
)

// SQLServerDriver SQL Server数据库驱动实现
type SQLServerDriver struct{}

func (d *SQLServerDriver) GetDriverName() string { return "sqlserver" }

func (d *SQLServerDriver) BuildDSN(config *DatabaseConfig) string {
	// SQL Server DSN format: sqlserver://username:password@host:port?database=database&param=value
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port)

	// Add database parameter
	params := d.GetDefaultParams()
	if config.Database != "" {
		params["database"] = config.Database
	}

	// Add additional params
	for k, v := range config.Params {
		params[k] = v
	}

	// Build query string
	if len(params) > 0 {
		dsn += "?"
		paramPairs := make([]string, 0, len(params))
		for k, v := range params {
			paramPairs = append(paramPairs, fmt.Sprintf("%s=%s", k, v))
		}
		dsn += strings.Join(paramPairs, "&")
	}

	return dsn
}

func (d *SQLServerDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	// SQL Server connection setup using official SQL Server driver
	sqlDB, err := sql.Open("sqlserver", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL Server connection to %s:%d as %s (database: %s): %w",
			config.Host, config.Port, config.Username, config.Database, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Use official GORM SQL Server driver
	dial := sqlserverd.New(sqlserverd.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM SQL Server connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	return gormDB, nil
}

// GetConnectionWithRetry SQL Server连接重试版本（可选使用）
func (d *SQLServerDriver) GetConnectionWithRetry(config *DatabaseConfig, maxRetries int) (*gorm.DB, error) {
	var lastErr error
	for i := 0; i <= maxRetries; i++ {
		db, err := d.GetConnection(config)
		if err == nil {
			return db, nil
		}
		lastErr = err

		// 如果不是最后一次尝试，等待后重试
		if i < maxRetries {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	return nil, fmt.Errorf("failed to connect after %d retries: %w", maxRetries, lastErr)
}

// CreateTableIfNotExists 创建表（如果不存在）- SQL Server版本
func (d *SQLServerDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate SQL Server create table SQL: %w", err)
	}

	// 执行建表SQL
	if err := dbConn.Exec(createSQL).Error; err != nil {
		// 如果是表已存在的错误，忽略
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "already exists") ||
			strings.Contains(errStr, "already_exists") ||
			strings.Contains(errStr, "there is already an object named") ||
			strings.Contains(errStr, "2714") { // SQL Server table exists error code
			return nil
		}
		return fmt.Errorf("failed to create SQL Server table: %w", err)
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - SQL Server版本
func (d *SQLServerDriver) GetFullTableName(tableInfo *TableInfo) string {
	// SQL Server支持schema，默认使用dbo schema
	schema := "dbo"
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		schema = tableInfo.Conn.Schema
	}

	return fmt.Sprintf("%s.%s", schema, tableInfo.TableName)
}

// GenerateCreateTableSQL 生成建表SQL - SQL Server版本
func (d *SQLServerDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
	if len(tableInfo.Fields) == 0 {
		return "", fmt.Errorf("no fields specified for table creation")
	}

	var sqlBuilder strings.Builder
	fullTableName := d.GetFullTableName(tableInfo)

	sqlBuilder.WriteString("IF NOT EXISTS (SELECT * FROM sysobjects WHERE name='")
	sqlBuilder.WriteString(tableInfo.TableName)
	sqlBuilder.WriteString("' AND xtype='U')\nBEGIN\n")

	sqlBuilder.WriteString("CREATE TABLE ")
	sqlBuilder.WriteString(fullTableName)
	sqlBuilder.WriteString(" (\n")

	fieldSQLs := make([]string, 0, len(tableInfo.Fields))

	for i, field := range tableInfo.Fields {
		fieldSQL, err := d.generateFieldSQL(field.Target, i == 0)
		if err != nil {
			return "", fmt.Errorf("failed to generate field SQL for %s: %w", field.Target.Name, err)
		}
		fieldSQLs = append(fieldSQLs, fieldSQL)
	}

	sqlBuilder.WriteString(strings.Join(fieldSQLs, ",\n"))
	sqlBuilder.WriteString("\n)")

	sqlBuilder.WriteString("\nEND")

	return sqlBuilder.String(), nil
}

// generateFieldSQL 生成字段SQL - SQL Server版本
func (d *SQLServerDriver) generateFieldSQL(field FieldAttr, isFirst bool) (string, error) {
	var sqlBuilder strings.Builder

	// 字段名（使用方括号转义）
	sqlBuilder.WriteString("    [")
	sqlBuilder.WriteString(field.Name)
	sqlBuilder.WriteString("] ")

	// 使用映射表进行数据类型映射
	dataType := d.mapDataTypeByMapping(field)
	sqlBuilder.WriteString(dataType)

	// 主键
	if field.PrimaryKey == PrimaryKeyFlag {
		sqlBuilder.WriteString(" PRIMARY KEY")
	}

	// 非空约束
	if field.IsNullable == "NO" {
		sqlBuilder.WriteString(" NOT NULL")
	}

	return sqlBuilder.String(), nil
}

// mapDataTypeByMapping 按照映射表实现数据类型映射
func (d *SQLServerDriver) mapDataTypeByMapping(field FieldAttr) string {
	mapping := d.GetDataTypeMapping()
	dataTypeKey := strings.ToLower(field.DataType)

	// 首先尝试从映射表中获取数据类型
	dataType, exists := mapping[dataTypeKey]
	if exists {
		// 处理需要动态参数的类型
		switch strings.ToUpper(field.DataType) {
		case "VARCHAR", "NVARCHAR", "TEXT", "STRING":
			if field.DataLenth > 0 && field.DataLenth <= 8000 {
				return fmt.Sprintf("NVARCHAR(%d)", field.DataLenth)
			}
			// 超过8000字符使用NVARCHAR(MAX)
			return "NVARCHAR(MAX)"

		case "CHAR", "NCHAR":
			if field.DataLenth > 0 && field.DataLenth <= 4000 {
				return fmt.Sprintf("NCHAR(%d)", field.DataLenth)
			}
			return "NCHAR(1)"

		case "DECIMAL", "NUMERIC":
			if field.Precision > 0 {
				scale := DefaultDecimalScale
				if field.Precision > 0 && field.Precision <= scale {
					scale = field.Precision
				}
				return fmt.Sprintf("DECIMAL(%d, %d)", DefaultDecimalPrecision, scale)
			}
			return fmt.Sprintf("DECIMAL(%d, %d)", DefaultDecimalPrecision, DefaultDecimalScale)

		case "BINARY":
			if field.DataLenth > 0 && field.DataLenth <= 8000 {
				return fmt.Sprintf("BINARY(%d)", field.DataLenth)
			}
			return "VARBINARY(MAX)"

		case "VARBINARY":
			if field.DataLenth > 0 && field.DataLenth <= 8000 {
				return fmt.Sprintf("VARBINARY(%d)", field.DataLenth)
			}
			return "VARBINARY(MAX)"

		default:
			// 对于直接映射的类型，直接返回映射结果
			return dataType
		}
	}

	// 如果映射表中没有找到，使用默认的NVARCHAR
	if field.DataLenth > 0 && field.DataLenth <= 8000 {
		return fmt.Sprintf("NVARCHAR(%d)", field.DataLenth)
	}
	return "NVARCHAR(255)"
}

// GetDataTypeMapping 获取数据类型映射 - SQL Server版本
func (d *SQLServerDriver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		"string":        "NVARCHAR",
		"varchar":       "NVARCHAR",
		"nvarchar":      "NVARCHAR",
		"text":          "NVARCHAR",
		"int":           "INT",
		"integer":       "INT",
		"bigint":        "BIGINT",
		"smallint":      "SMALLINT",
		"tinyint":       "TINYINT",
		"decimal":       "DECIMAL",
		"numeric":       "NUMERIC",
		"float":         "FLOAT",
		"real":          "REAL",
		"double":        "FLOAT",
		"bit":           "BIT",
		"bool":          "BIT",
		"boolean":       "BIT",
		"datetime":      "DATETIME",
		"timestamp":     "DATETIME",
		"date":          "DATE",
		"time":          "TIME",
		"datetime2":     "DATETIME2",
		"money":         "MONEY",
		"smallmoney":    "SMALLMONEY",
		"binary":        "VARBINARY",
		"varbinary":     "VARBINARY",
		"image":         "IMAGE",
		"uuid":          "UNIQUEIDENTIFIER",
		"xml":           "XML",
		"sql_variant":   "SQL_VARIANT",
		"smalldatetime": "DATETIME",
	}
}

// SupportSchema 是否支持schema - SQL Server版本
func (d *SQLServerDriver) SupportSchema() bool {
	return true // SQL Server支持schema
}

// SupportBatchInsert 是否支持批量插入 - SQL Server版本
func (d *SQLServerDriver) SupportBatchInsert() bool {
	return true
}

// GetDefaultParams 获取默认连接参数 - SQL Server版本
func (d *SQLServerDriver) GetDefaultParams() map[string]string {
	return map[string]string{}
}

// ListTables 列出数据库中的所有表 - SQL Server版本
func (d *SQLServerDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// SQL Server中，如果没有指定schema，使用dbo
	if schema == "" {
		schema = "dbo"
	}

	query := `
		SELECT
			t.name as name,
			s.name as [schema],
			'TABLE' as type,
			COALESCE(CAST(ep.value AS NVARCHAR(MAX)), '') as comment
		FROM sys.tables t
		INNER JOIN sys.schemas s ON t.schema_id = s.schema_id
		LEFT JOIN sys.extended_properties ep ON ep.major_id = t.object_id
			AND ep.minor_id = 0
			AND ep.class = 1
			AND ep.name = 'MS_Description'
		WHERE s.name = ?
		ORDER BY t.name
	`

	rows, err := dbConn.Raw(query, schema).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query SQL Server tables: %w", err)
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
			return nil, fmt.Errorf("failed to scan SQL Server table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// ListTableColumns 列出指定表的字段信息 - SQL Server版本
func (d *SQLServerDriver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	// SQL Server中，如果没有指定schema，使用dbo
	if schema == "" {
		schema = "dbo"
	}

	query := fmt.Sprintf(`
		SELECT
			c.COLUMN_NAME as name,
			c.DATA_TYPE as data_type,
			COALESCE(c.CHARACTER_MAXIMUM_LENGTH, c.NUMERIC_PRECISION) as data_length,
			c.IS_NULLABLE as is_nullable,
			CASE WHEN pk.COLUMN_NAME IS NOT NULL THEN 1 ELSE 0 END as primary_key,
			c.COLUMN_DEFAULT as default_value,
			ep.value as comment,
			c.NUMERIC_PRECISION as precision,
			c.NUMERIC_SCALE as scale
		FROM INFORMATION_SCHEMA.COLUMNS c
		LEFT JOIN (
			SELECT ku.COLUMN_NAME
			FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
			INNER JOIN INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
				ON ku.CONSTRAINT_NAME = tc.CONSTRAINT_NAME
				AND ku.TABLE_SCHEMA = tc.TABLE_SCHEMA
				AND ku.TABLE_NAME = tc.TABLE_NAME
			WHERE tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
				AND ku.TABLE_SCHEMA = '%s'
				AND ku.TABLE_NAME = '%s'
		) pk ON c.COLUMN_NAME = pk.COLUMN_NAME
		LEFT JOIN sys.extended_properties ep
			ON ep.major_id = OBJECT_ID(c.TABLE_SCHEMA + '.' + c.TABLE_NAME)
			AND ep.minor_id = c.ORDINAL_POSITION
			AND ep.name = 'MS_Description'
		WHERE c.TABLE_SCHEMA = '%s' AND c.TABLE_NAME = '%s'
		ORDER BY c.ORDINAL_POSITION
	`, schema, tableName, schema, tableName)

	rows, err := dbConn.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query SQL Server table columns: %w", err)
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
			return nil, fmt.Errorf("failed to scan SQL Server column row: %w", err)
		}

		// 将NULL值转换为适当的格式
		col.DefaultValue = defaultValueScanner.String()
		col.Comment = commentScanner.String()
		col.DataLength = dataLengthScanner.Int()
		col.Precision = precisionScanner.Int()
		col.Scale = scaleScanner.Int()

		// 将SQL Server的数据类型转换为更易理解的格式
		col.DataType = normalizeSQLServerDataType(col.DataType)

		columns = append(columns, col)
	}

	return columns, nil
}

// normalizeSQLServerDataType 将SQL Server的数据类型转换为更易理解的格式
func normalizeSQLServerDataType(dataType string) string {
	switch strings.ToLower(dataType) {
	case "int":
		return "int"
	case "smallint":
		return "smallint"
	case "tinyint":
		return "tinyint"
	case "bigint":
		return "bigint"
	case "bit":
		return "bit"
	case "decimal":
		return "decimal"
	case "numeric":
		return "decimal"
	case "money":
		return "money"
	case "smallmoney":
		return "smallmoney"
	case "float":
		return "float"
	case "real":
		return "real"
	case "char":
		return "char"
	case "varchar":
		return "varchar"
	case "nchar":
		return "nchar"
	case "nvarchar":
		return "nvarchar"
	case "text":
		return "text"
	case "ntext":
		return "ntext"
	case "binary":
		return "binary"
	case "varbinary":
		return "varbinary"
	case "image":
		return "image"
	case "datetime":
		return "datetime"
	case "datetime2":
		return "datetime2"
	case "smalldatetime":
		return "smalldatetime"
	case "date":
		return "date"
	case "time":
		return "time"
	case "datetimeoffset":
		return "datetimeoffset"
	case "timestamp":
		return "timestamp"
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

// EscapeIdentifier 转义SQL Server标识符（表名、字段名等）
func (d *SQLServerDriver) EscapeIdentifier(identifier string) string {
	// SQL Server中使用方括号转义标识符
	if strings.ContainsAny(identifier, "[]") {
		// 如果已经包含方括号，进行转义
		return "[" + strings.ReplaceAll(identifier, "]", "]]") + "]"
	}
	// 检查是否需要转义（包含特殊字符或空格）
	if strings.ContainsAny(identifier, " \t\n\r\f\v`~!@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		strings.ToUpper(identifier) != identifier { // 如果不是全大写，可能是保留字
		return "[" + identifier + "]"
	}
	return identifier
}

// isSQLServerReservedWord 检查是否为SQL Server保留字
func (d *SQLServerDriver) isSQLServerReservedWord(word string) bool {
	reservedWords := []string{
		"ADD", "ALL", "ALTER", "AND", "ANY", "AS", "ASC", "AUTHORIZATION", "BACKUP",
		"BEGIN", "BETWEEN", "BREAK", "BROWSE", "BULK", "BY", "CASCADE", "CASE",
		"CHECK", "CHECKPOINT", "CLOSE", "CLUSTERED", "COALESCE", "COLLATE", "COLUMN",
		"COMMIT", "COMPUTE", "CONSTRAINT", "CONTAINS", "CONTAINSTABLE", "CONTINUE",
		"CONVERT", "CREATE", "CROSS", "CURRENT", "CURRENT_DATE", "CURRENT_TIME",
		"CURRENT_TIMESTAMP", "CURRENT_USER", "CURSOR", "DATABASE", "DBCC", "DEALLOCATE",
		"DECLARE", "DEFAULT", "DELETE", "DENY", "DESC", "DISK", "DISTINCT", "DISTRIBUTED",
		"DOUBLE", "DROP", "DUMP", "ELSE", "END", "ERRLVL", "ESCAPE", "EXCEPT", "EXEC",
		"EXECUTE", "EXISTS", "EXIT", "EXTERNAL", "FETCH", "FILE", "FILLFACTOR", "FOR",
		"FOREIGN", "FREETEXT", "FREETEXTTABLE", "FROM", "FULL", "FUNCTION", "GOTO",
		"GRANT", "GROUP", "HAVING", "HOLDLOCK", "HOUR", "IDENTITY", "IDENTITYCOL",
		"IDENTITY_INSERT", "IF", "IN", "INDEX", "INNER", "INSERT", "INSTEAD", "INTERSECT",
		"INTO", "IS", "JOIN", "KEY", "KILL", "LEFT", "LIKE", "LINENO", "LOAD", "MERGE",
		"NATIONAL", "NOCHECK", "NONCLUSTERED", "NOT", "NULL", "NULLIF", "OF", "OFF",
		"OFFSETS", "ON", "OPEN", "OPENDATASOURCE", "OPENQUERY", "OPENROWSET", "OPENXML",
		"OPTION", "OR", "ORDER", "OUTER", "OVER", "PERCENT", "PIVOT", "PLAN", "PRECISION",
		"PRIMARY", "PRINT", "PROC", "PROCEDURE", "PUBLIC", "RAISERROR", "READ", "READTEXT",
		"RECONFIGURE", "REFERENCES", "REPLICATION", "RESTORE", "RESTRICT", "RETURN",
		"REVERT", "REVOKE", "RIGHT", "ROLLBACK", "ROWCOUNT", "ROWGUIDCOL", "RULE",
		"SAVE", "SCHEMA", "SECURITYAUDIT", "SELECT", "SEMANTICKEYPHRASETABLE",
		"SEMANTICSIMILARITYDETAILSTABLE", "SEMANTICSIMILARITYTABLE", "SESSION_USER",
		"SET", "SETUSER", "SHUTDOWN", "SOME", "STATISTICS", "SYSTEM_USER", "TABLE",
		"TABLESAMPLE", "TEXTSIZE", "THEN", "TO", "TOP", "TRAN", "TRANSACTION", "TRIGGER",
		"TRUNCATE", "TRY_CONVERT", "TSEQUAL", "UNION", "UNIQUE", "UNPIVOT", "UPDATE",
		"UPDATETEXT", "USE", "USER", "VALUES", "VARYING", "VIEW", "WAITFOR", "WHEN",
		"WHERE", "WHILE", "WITH", "WITHIN GROUP", "WRITETEXT", "YEAR",
	}

	wordUpper := strings.ToUpper(word)
	for _, reserved := range reservedWords {
		if wordUpper == reserved {
			return true
		}
	}
	return false
}
