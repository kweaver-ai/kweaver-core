package writer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	mysqld "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// MySQLDriver MySQL数据库驱动实现
type MySQLDriver struct{}

func (d *MySQLDriver) GetDriverName() string { return "mysql" }

func (d *MySQLDriver) BuildDSN(config *DatabaseConfig) string {
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

func (d *MySQLDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	// MySQL specific connection setup
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection to %s:%d as %s (database: %s): %w",
			config.Host, config.Port, config.Username, config.Database, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dial := mysqld.New(mysqld.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM MySQL connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	return gormDB, nil
}

// CreateTableIfNotExists 创建表（如果不存在）- MySQL版本
func (d *MySQLDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate MySQL create table SQL: %w", err)
	}

	// 执行建表SQL
	if err := dbConn.Exec(createSQL).Error; err != nil {
		// 如果是表已存在的错误，忽略
		if strings.Contains(err.Error(), "already exists") ||
			strings.Contains(err.Error(), "already_exists") ||
			strings.Contains(err.Error(), "1050") { // MySQL table exists error code
			return nil
		}
		return fmt.Errorf("failed to create MySQL table: %w", err)
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - MySQL版本
func (d *MySQLDriver) GetFullTableName(tableInfo *TableInfo) string {
	// MySQL 支持 schema，但通常使用 database 概念
	// 如果指定了 schema，使用 schema.table 格式
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		return fmt.Sprintf("%s.%s", tableInfo.Conn.Schema, tableInfo.TableName)
	}

	// 否则只返回表名
	return tableInfo.TableName
}

// EscapeIdentifier 转义MySQL标识符（表名、字段名等）
// MySQL中，如果标识符包含特殊字符或不符合命名规则，需要用反引号包围
func (d *MySQLDriver) EscapeIdentifier(identifier string) string {
	// 检查是否需要转义
	// 1. 包含特殊字符
	// 2. 以数字开头
	// 3. 包含空格
	// 4. 是MySQL保留字
	if strings.ContainsAny(identifier, "`@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		(len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
		strings.Contains(identifier, " ") ||
		d.isMySQLReservedWord(strings.ToUpper(identifier)) {
		return "`" + strings.ReplaceAll(identifier, "`", "``") + "`"
	}

	return identifier
}

// isMySQLReservedWord 检查是否为MySQL保留字
func (d *MySQLDriver) isMySQLReservedWord(word string) bool {
	reservedWords := []string{
		"ADD", "ALL", "ALTER", "AND", "AS", "ASC", "AUTO_INCREMENT",
		"BETWEEN", "BIGINT", "BINARY", "BLOB", "BOTH", "BY",
		"CASE", "CHANGE", "CHAR", "COLUMN", "COLUMNS", "CONSTRAINT",
		"CREATE", "CROSS", "CURRENT_DATE", "CURRENT_TIME", "CURRENT_TIMESTAMP",
		"DATABASE", "DATABASES", "DATE", "DECIMAL", "DEFAULT", "DELAYED",
		"DELETE", "DESC", "DESCRIBE", "DISTINCT", "DOUBLE", "DROP",
		"ELSE", "ENCLOSED", "ESCAPED", "EXISTS", "EXPLAIN", "FALSE",
		"FLOAT", "FOR", "FOREIGN", "FROM", "FULL", "FULLTEXT",
		"GRANT", "GROUP", "HAVING", "HIGH_PRIORITY", "HOSTS",
		"IF", "IGNORE", "IN", "INDEX", "INFILE", "INNER", "INSERT",
		"INT", "INTEGER", "INTERVAL", "INTO", "IS", "JOIN", "KEY", "KEYS",
		"LEADING", "LEFT", "LIKE", "LIMIT", "LINES", "LOAD", "LOCAL",
		"LOCK", "LONGTEXT", "LOW_PRIORITY",
		"MATCH", "MEDIUMINT", "MEDIUMTEXT", "MIDDLEINT",
		"NATURAL", "NOT", "NULL", "NUMERIC",
		"ON", "OPTIMIZE", "OPTION", "OPTIONALLY", "OR", "ORDER", "OUTER", "OUTFILE",
		"PARTIAL", "PRIMARY", "PRIVILEGES", "PROCEDURE", "PURGE",
		"READ", "REAL", "REFERENCES", "REGEXP", "RENAME", "REPLACE", "REQUIRE",
		"RESTRICT", "RIGHT", "RLIKE", "SCHEMA", "SCHEMAS",
		"SELECT", "SET", "SHOW", "SMALLINT", "SONAME", "SQL_BIG_RESULT",
		"SQL_CALC_FOUND_ROWS", "SQL_SMALL_RESULT", "STARTING", "STRAIGHT_JOIN",
		"TABLE", "TABLES", "TERMINATED", "TEXT", "THEN", "TIME", "TIMESTAMP",
		"TINYINT", "TINYTEXT", "TO", "TRAILING", "TRUE", "TRUNCATE",
		"UNION", "UNIQUE", "UNLOCK", "UNSIGNED", "UPDATE", "USAGE", "USE",
		"USING", "VALUES", "VARBINARY", "VARCHAR", "VARYING",
		"WHEN", "WHERE", "WITH", "WRITE",
		"YEAR", "ZEROFILL",
	}

	for _, reserved := range reservedWords {
		if word == reserved {
			return true
		}
	}
	return false
}

// escapeCommentString 转义MySQL注释字符串中的特殊字符
func (d *MySQLDriver) escapeCommentString(comment string) string {
	// MySQL字符串字面量中的转义规则：
	// 1. 单引号需要转义为两个单引号
	// 2. 反斜杠需要转义为两个反斜杠

	// 先转义反斜杠
	result := strings.ReplaceAll(comment, "\\", "\\\\")
	// 再转义单引号
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

func (d *MySQLDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
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

		// Handle MySQL UNSIGNED types
		isUnsigned := strings.Contains(dataType, "UNSIGNED")
		baseDataType := strings.TrimSuffix(dataType, " UNSIGNED")
		baseDataType = strings.TrimSuffix(baseDataType, "UNSIGNED")

		if mappedType, exists := mapping[baseDataType]; exists {
			dataType = mappedType
			if isUnsigned {
				dataType += " UNSIGNED"
			}
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

func (d *MySQLDriver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		"TINYINT":    "TINYINT",
		"SMALLINT":   "SMALLINT",
		"MEDIUMINT":  "MEDIUMINT",
		"INT":        "INT",
		"INTEGER":    "INT",
		"BIGINT":     "BIGINT",
		"DECIMAL":    "DECIMAL",
		"FLOAT":      "FLOAT",
		"DOUBLE":     "DOUBLE",
		"CHAR":       "CHAR",
		"VARCHAR":    "VARCHAR",
		"STRING":     "VARCHAR",
		"TEXT":       "TEXT",
		"TINYTEXT":   "TINYTEXT",
		"MEDIUMTEXT": "MEDIUMTEXT",
		"LONGTEXT":   "LONGTEXT",
		"BINARY":     "BINARY",
		"VARBINARY":  "VARBINARY",
		"BLOB":       "BLOB",
		"TINYBLOB":   "TINYBLOB",
		"MEDIUMBLOB": "MEDIUMBLOB",
		"LONGBLOB":   "LONGBLOB",
		"DATE":       "DATE",
		"DATETIME":   "DATETIME",
		"TIMESTAMP":  "TIMESTAMP",
		"TIME":       "TIME",
		"YEAR":       "YEAR",
		"BOOL":       "TINYINT(1)",
		"BOOLEAN":    "TINYINT(1)",
		"ENUM":       "VARCHAR",
		"SET":        "VARCHAR",
		"JSON":       "JSON",
		"NUMERIC":    "DECIMAL",
		"BIT":        "BIT",
	}
}

func (d *MySQLDriver) SupportSchema() bool      { return false }
func (d *MySQLDriver) SupportBatchInsert() bool { return true }

func (d *MySQLDriver) GetDefaultParams() map[string]string {
	return map[string]string{
		"charset":   "utf8mb4",
		"parseTime": "true",
	}
}

// ListTables 列出数据库中的所有表 - MySQL版本
func (d *MySQLDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// MySQL中，schema就是数据库名
	// 如果没有指定schema，使用当前数据库
	if schema == "" {
		schema = dbConn.Migrator().CurrentDatabase()
	}

	query := `
		SELECT
			TABLE_NAME as name,
			TABLE_SCHEMA as table_schema,
			TABLE_TYPE as type,
			COALESCE(TABLE_COMMENT, '') as comment
		FROM information_schema.TABLES
		WHERE TABLE_SCHEMA = ?
		ORDER BY TABLE_NAME
	`

	rows, err := dbConn.Raw(query, schema).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query MySQL tables: %w", err)
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
			return nil, fmt.Errorf("failed to scan MySQL table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// ListTableColumns 列出指定表的字段信息 - MySQL版本
func (d *MySQLDriver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	// MySQL中，schema就是数据库名
	// 如果没有指定schema，使用当前数据库
	if schema == "" {
		schema = dbConn.Migrator().CurrentDatabase()
	}

	query := `
		SELECT
			COLUMN_NAME as name,
			DATA_TYPE as data_type,
			COALESCE(CHARACTER_MAXIMUM_LENGTH, NUMERIC_PRECISION) as data_length,
			IS_NULLABLE as is_nullable,
			COLUMN_KEY = 'PRI' as primary_key,
			COLUMN_DEFAULT as default_value,
			COLUMN_COMMENT as comment,
			NUMERIC_PRECISION as numeric_precision,
			NUMERIC_SCALE as numeric_scale
		FROM information_schema.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := dbConn.Raw(query, schema, tableName).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query MySQL table columns: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var col ColumnInfo
		var primaryKeyInt int
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
			&primaryKeyInt,
			&defaultValueScanner,
			&commentScanner,
			&precisionScanner,
			&scaleScanner,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan MySQL column row: %w", err)
		}

		// 将NULL值转换为适当的格式
		col.DefaultValue = defaultValueScanner.String()
		col.Comment = commentScanner.String()
		col.DataLength = dataLengthScanner.Int()
		col.Precision = precisionScanner.Int()
		col.Scale = scaleScanner.Int()
		col.PrimaryKey = primaryKeyInt

		// 将MySQL的数据类型转换为更易理解的格式
		col.DataType = normalizeMySQLDataType(col.DataType)

		columns = append(columns, col)
	}

	return columns, nil
}

// normalizeMySQLDataType 将MySQL的数据类型转换为更易理解的格式
func normalizeMySQLDataType(dataType string) string {
	switch strings.ToLower(dataType) {
	case "int":
		return "int"
	case "tinyint":
		return "tinyint"
	case "smallint":
		return "smallint"
	case "mediumint":
		return "mediumint"
	case "bigint":
		return "bigint"
	case "decimal":
		return "decimal"
	case "float":
		return "float"
	case "double":
		return "double"
	case "bit":
		return "bit"
	case "char":
		return "char"
	case "varchar":
		return "varchar"
	case "binary":
		return "binary"
	case "varbinary":
		return "varbinary"
	case "tinyblob":
		return "tinyblob"
	case "blob":
		return "blob"
	case "mediumblob":
		return "mediumblob"
	case "longblob":
		return "longblob"
	case "tinytext":
		return "tinytext"
	case "text":
		return "text"
	case "mediumtext":
		return "mediumtext"
	case "longtext":
		return "longtext"
	case "enum":
		return "enum"
	case "set":
		return "set"
	case "date":
		return "date"
	case "datetime":
		return "datetime"
	case "timestamp":
		return "timestamp"
	case "time":
		return "time"
	case "year":
		return "year"
	default:
		return dataType
	}
}
