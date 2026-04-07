package writer

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	postgresd "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgreSQLDriver PostgreSQL数据库驱动实现
type PostgreSQLDriver struct{}

func (d *PostgreSQLDriver) GetDriverName() string { return "postgres" }

func (d *PostgreSQLDriver) BuildDSN(config *DatabaseConfig) string {
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

func (d *PostgreSQLDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	// PostgreSQL specific connection setup
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection to %s:%d as %s (database: %s): %w",
			config.Host, config.Port, config.Username, config.Database, err)
	}

	// Configure connection pool
	sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Hour)

	dial := postgresd.New(postgresd.Config{Conn: sqlDB})
	gormDB, err := gorm.Open(dial, &gorm.Config{Logger: getLogger()})
	if err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("failed to create GORM PostgreSQL connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	return gormDB, nil
}

// CreateTableIfNotExists 创建表（如果不存在）- PostgreSQL版本
func (d *PostgreSQLDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate PostgreSQL create table SQL: %w", err)
	}

	// PostgreSQL可能包含多个SQL语句（建表+注释）
	sqlStatements := strings.Split(createSQL, "; ")
	for _, sql := range sqlStatements {
		sql = strings.TrimSpace(sql)
		if sql == "" {
			continue
		}

		if err := dbConn.Exec(sql).Error; err != nil {
			// 如果是表已存在的错误或列注释已存在的错误，忽略
			if strings.Contains(err.Error(), "already exists") ||
				strings.Contains(err.Error(), "already_exists") ||
				strings.Contains(err.Error(), "42P07") || // PostgreSQL table exists error code
				strings.Contains(err.Error(), "42710") { // PostgreSQL comment already exists
				continue
			}
			return fmt.Errorf("failed to execute PostgreSQL SQL: %s, error: %w", sql, err)
		}
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - PostgreSQL版本
func (d *PostgreSQLDriver) GetFullTableName(tableInfo *TableInfo) string {
	// PostgreSQL 原生支持 schema
	// 如果指定了 schema，使用 schema.table 格式
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		return fmt.Sprintf("%s.%s", tableInfo.Conn.Schema, tableInfo.TableName)
	}

	// 默认使用 public schema
	return fmt.Sprintf("public.%s", tableInfo.TableName)
}

// EscapeIdentifier 转义PostgreSQL标识符（表名、字段名等）
// PostgreSQL中，如果标识符包含特殊字符或不符合命名规则，需要用双引号包围
func (d *PostgreSQLDriver) EscapeIdentifier(identifier string) string {
	// 检查是否需要转义
	// 1. 包含特殊字符
	// 2. 以数字开头
	// 3. 包含空格
	// 4. 是PostgreSQL保留字
	// 5. 包含大写字母（PostgreSQL默认大小写敏感）
	if strings.ContainsAny(identifier, "\"@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		(len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
		strings.Contains(identifier, " ") ||
		strings.ToLower(identifier) != identifier || // 包含大写字母
		d.isPostgreSQLReservedWord(strings.ToUpper(identifier)) {
		return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
	}

	return identifier
}

// isPostgreSQLReservedWord 检查是否为PostgreSQL保留字
func (d *PostgreSQLDriver) isPostgreSQLReservedWord(word string) bool {
	reservedWords := []string{
		"ALL", "AND", "ANY", "ARRAY", "AS", "ASC", "ASYMMETRIC",
		"BETWEEN", "BIGINT", "BINARY", "BIT", "BOOLEAN", "BY",
		"CASE", "CAST", "CHAR", "CHARACTER", "CHECK", "COLLATE",
		"COLUMN", "CONSTRAINT", "CREATE", "CROSS", "CURRENT_CATALOG",
		"CURRENT_DATE", "CURRENT_ROLE", "CURRENT_SCHEMA", "CURRENT_TIME",
		"CURRENT_TIMESTAMP", "CURRENT_USER", "DECIMAL", "DEFAULT",
		"DELETE", "DESC", "DISTINCT", "DO", "DOUBLE", "DROP",
		"ELSE", "END", "EXCEPT", "EXISTS", "EXTRACT", "FALSE",
		"FETCH", "FLOAT", "FOR", "FOREIGN", "FROM", "FULL",
		"GRANT", "GROUP", "HAVING", "ILIKE", "IN", "INITIALLY",
		"INNER", "INSERT", "INT", "INTEGER", "INTERSECT", "INTERVAL",
		"INTO", "IS", "JOIN", "LATERAL", "LEADING", "LEFT",
		"LIKE", "LIMIT", "LOCAL", "NATURAL", "NOT", "NULL",
		"NUMERIC", "ON", "ONLY", "OR", "ORDER", "OUTER",
		"OVERLAPS", "PLACING", "POSITION", "PRECISION", "PRIMARY",
		"REAL", "REFERENCES", "RETURNING", "RIGHT", "SELECT",
		"SESSION_USER", "SIMILAR", "SMALLINT", "SOME", "SYMMETRIC",
		"TABLE", "THEN", "TIME", "TIMESTAMP", "TO", "TRAILING",
		"TREAT", "TRIM", "TRUE", "UNION", "UNIQUE", "UNKNOWN",
		"UPDATE", "USER", "USING", "VALUES", "VARCHAR", "VARIADIC",
		"WHEN", "WHERE", "WINDOW", "WITH", "XML", "XMLATTRIBUTES",
		"XMLCONCAT", "XMLELEMENT", "XMLEXISTS", "XMLFOREST", "XMLPARSE",
		"XMLPI", "XMLROOT", "XMLSERIALIZE",
	}

	for _, reserved := range reservedWords {
		if word == reserved {
			return true
		}
	}
	return false
}

// escapeCommentString 转义PostgreSQL注释字符串中的特殊字符
func (d *PostgreSQLDriver) escapeCommentString(comment string) string {
	// PostgreSQL字符串字面量中的转义规则：
	// 1. 单引号需要转义为两个单引号
	// 2. 反斜杠需要转义为两个反斜杠
	// 3. 其他特殊字符保持原样（在单引号包围的字符串中）

	// 先转义反斜杠
	result := strings.ReplaceAll(comment, "\\", "\\\\")
	// 再转义单引号
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

func (d *PostgreSQLDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
	if len(tableInfo.Fields) == 0 {
		return "", fmt.Errorf("no field mappings provided")
	}

	fullTableName := d.GetFullTableName(tableInfo)

	var sqls []string

	// 1. 生成建表语句（不包含注释）
	var createSQL strings.Builder
	createSQL.WriteString("CREATE TABLE IF NOT EXISTS ")
	createSQL.WriteString(fullTableName)
	createSQL.WriteString(" (")

	columns := make([]string, 0, len(tableInfo.Fields))
	for i := range tableInfo.Fields {
		field := &tableInfo.Fields[i]
		if field.Target.Name == "" {
			continue
		}

		var columnSQL strings.Builder
		columnSQL.WriteString(d.EscapeIdentifier(field.Target.Name))
		columnSQL.WriteString(" ")

		// 使用驱动的数据类型映射
		mapping := d.GetDataTypeMapping()
		dataType := strings.ToUpper(field.Target.DataType)

		// Handle UNSIGNED types (PostgreSQL doesn't support UNSIGNED, but for consistency)
		baseDataType := strings.TrimSuffix(dataType, " UNSIGNED")
		baseDataType = strings.TrimSuffix(baseDataType, "UNSIGNED")

		if mappedType, exists := mapping[baseDataType]; exists {
			dataType = mappedType
		} else {
			// If no mapping found, use the original type
			dataType = field.Target.DataType
		}

		// 处理长度限制
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

		columns = append(columns, columnSQL.String())
	}

	createSQL.WriteString(strings.Join(columns, ", "))
	createSQL.WriteString(")")
	sqls = append(sqls, createSQL.String())

	// 2. 生成注释语句
	for i := range tableInfo.Fields {
		field := &tableInfo.Fields[i]
		if field.Target.Name != "" && field.Target.Comment != "" {
			// 转义注释内容中的特殊字符
			escapedComment := d.escapeCommentString(field.Target.Comment)
			escapedTargetName := d.EscapeIdentifier(field.Target.Name)
			commentSQL := fmt.Sprintf("COMMENT ON COLUMN %s.%s IS '%s'",
				fullTableName,
				escapedTargetName,
				escapedComment)
			sqls = append(sqls, commentSQL)
		}
	}

	return strings.Join(sqls, "; "), nil
}

func (d *PostgreSQLDriver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		"TINYINT":    "SMALLINT", // PostgreSQL没有TINYINT，使用SMALLINT
		"SMALLINT":   "SMALLINT",
		"MEDIUMINT":  "INTEGER", // PostgreSQL没有MEDIUMINT，使用INTEGER
		"INT":        "INTEGER",
		"INTEGER":    "INTEGER",
		"BIGINT":     "BIGINT",
		"DECIMAL":    "DECIMAL",
		"FLOAT":      "REAL",
		"DOUBLE":     "DOUBLE PRECISION",
		"CHAR":       "CHAR",
		"VARCHAR":    "VARCHAR",
		"STRING":     "VARCHAR",
		"TEXT":       "TEXT",
		"TINYTEXT":   "TEXT",
		"MEDIUMTEXT": "TEXT",
		"LONGTEXT":   "TEXT",
		"BINARY":     "BYTEA",
		"VARBINARY":  "BYTEA",
		"BLOB":       "BYTEA",
		"TINYBLOB":   "BYTEA",
		"MEDIUMBLOB": "BYTEA",
		"LONGBLOB":   "BYTEA",
		"DATE":       "DATE",
		"DATETIME":   "TIMESTAMP",
		"TIMESTAMP":  "TIMESTAMP",
		"TIME":       "TIME",
		"YEAR":       "INTEGER",
		"BOOL":       "BOOLEAN",
		"BOOLEAN":    "BOOLEAN",
		"ENUM":       "VARCHAR",
		"SET":        "VARCHAR",
		"JSON":       "JSONB",
	}
}

func (d *PostgreSQLDriver) SupportSchema() bool      { return true }
func (d *PostgreSQLDriver) SupportBatchInsert() bool { return true }

func (d *PostgreSQLDriver) GetDefaultParams() map[string]string {
	return map[string]string{
		"sslmode": "disable",
	}
}

// ListTables 列出数据库中的所有表 - PostgreSQL版本
func (d *PostgreSQLDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// PostgreSQL中，如果没有指定schema，使用public
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			t.table_name as name,
			t.table_schema as schema,
			'TABLE' as type,
			COALESCE(d.description, '') as comment
		FROM information_schema.tables t
		LEFT JOIN pg_class c ON c.relname = t.table_name
		LEFT JOIN pg_description d ON d.objoid = c.oid AND d.objsubid = 0
		WHERE t.table_schema = $1
			AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name
	`

	rows, err := dbConn.Raw(query, schema).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query PostgreSQL tables: %w", err)
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
			return nil, fmt.Errorf("failed to scan PostgreSQL table row: %w", err)
		}
		tables = append(tables, table)
	}

	return tables, nil
}

// ListTableColumns 列出指定表的字段信息 - PostgreSQL版本
func (d *PostgreSQLDriver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	// PostgreSQL中，如果没有指定schema，使用public
	if schema == "" {
		schema = "public"
	}

	query := `
		SELECT
			c.column_name as name,
			c.data_type as data_type,
			COALESCE(c.character_maximum_length, c.numeric_precision) as data_length,
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
		return nil, fmt.Errorf("failed to query PostgreSQL table columns: %w", err)
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
			return nil, fmt.Errorf("failed to scan PostgreSQL column row: %w", err)
		}

		// 将NULL值转换为适当的格式
		col.DefaultValue = defaultValueScanner.String()
		col.Comment = commentScanner.String()
		col.DataLength = dataLengthScanner.Int()
		col.Precision = precisionScanner.Int()
		col.Scale = scaleScanner.Int()

		// 将PostgreSQL的数据类型转换为更易理解的格式
		col.DataType = normalizeDataType(col.DataType)

		columns = append(columns, col)
	}

	return columns, nil
}

// normalizeDataType 将PostgreSQL的数据类型转换为更易理解的格式
func normalizeDataType(dataType string) string {
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
		return "numeric"
	default:
		return dataType
	}
}
