package writer

import (
	"fmt"
	"strings"

	oracle "github.com/godoes/gorm-oracle"
	"gorm.io/gorm"
)

// OracleDriver Oracle数据库驱动实现
type OracleDriver struct{}

func (d *OracleDriver) GetDriverName() string { return "oracle" }

func (d *OracleDriver) BuildDSN(config *DatabaseConfig) string {
	connStr := oracle.BuildUrl(config.Host, config.Port, config.Database, config.Username, config.Password, map[string]string{
		"SSL": "false",
	})
	return connStr
}

func (d *OracleDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
	dsn := d.BuildDSN(config)

	dialector := oracle.New(oracle.Config{
		DSN:                     dsn,
		IgnoreCase:              true,
		NamingCaseSensitive:     true,
		VarcharSizeIsCharLength: true,
	})

	// // 配置连接池
	// sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
	// sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
	// sqlDB.SetConnMaxLifetime(time.Hour)

	// // 创建自定义的 Oracle dialector
	// // 由于 go-ora 没有专门的 GORM dialector，我们使用通用的方法
	// dialector := &oracleDialector{db: sqlDB}

	gormDB, err := gorm.Open(dialector, &gorm.Config{
		Logger: getLogger(),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create GORM Oracle connection to %s:%d: %w",
			config.Host, config.Port, err)
	}

	// set session parameters
	if sqlDB, err := gormDB.DB(); err == nil {
		_, _ = oracle.AddSessionParams(sqlDB, map[string]string{
			"TIME_ZONE":               "+08:00",                       // ALTER SESSION SET TIME_ZONE = '+08:00';
			"NLS_DATE_FORMAT":         "YYYY-MM-DD",                   // ALTER SESSION SET NLS_DATE_FORMAT = 'YYYY-MM-DD';
			"NLS_TIME_FORMAT":         "HH24:MI:SSXFF",                // ALTER SESSION SET NLS_TIME_FORMAT = 'HH24:MI:SS.FF3';
			"NLS_TIMESTAMP_FORMAT":    "YYYY-MM-DD HH24:MI:SSXFF",     // ALTER SESSION SET NLS_TIMESTAMP_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3';
			"NLS_TIME_TZ_FORMAT":      "HH24:MI:SS.FF TZR",            // ALTER SESSION SET NLS_TIME_TZ_FORMAT = 'HH24:MI:SS.FF3 TZR';
			"NLS_TIMESTAMP_TZ_FORMAT": "YYYY-MM-DD HH24:MI:SSXFF TZR", // ALTER SESSION SET NLS_TIMESTAMP_TZ_FORMAT = 'YYYY-MM-DD HH24:MI:SS.FF3 TZR';
		})
	}
	return gormDB, nil
}

// CreateTableIfNotExists 创建表（如果不存在）- Oracle版本
func (d *OracleDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
	// 如果明确指定表已存在，跳过创建
	if tableInfo.TableExist {
		return nil
	}

	// 生成建表SQL
	createSQL, err := d.GenerateCreateTableSQL(tableInfo)
	if err != nil {
		return fmt.Errorf("failed to generate Oracle create table SQL: %w", err)
	}

	res := dbConn.Exec(createSQL)

	// 执行建表SQL
	if res.Error != nil {
		// Oracle表已存在错误码: ORA-00955
		errStr := strings.ToLower(res.Error.Error())
		if strings.Contains(errStr, "already exists") ||
			strings.Contains(errStr, "already_exists") ||
			strings.Contains(errStr, "ora-00955") { // Oracle table exists error code
			return nil
		}
		return fmt.Errorf("failed to create Oracle table: %w", err)
	}

	return nil
}

// GetFullTableName 获取包含schema的完整表名 - Oracle版本
func (d *OracleDriver) GetFullTableName(tableInfo *TableInfo) string {
	// Oracle默认使用用户名作为schema
	schema := ""
	if tableInfo.Conn != nil && tableInfo.Conn.Schema == "" {
		schema = tableInfo.Conn.Username
	}
	if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" {
		schema = tableInfo.Conn.Schema
	}

	// Oracle使用双引号转义标识符
	return fmt.Sprintf("\"%s\".\"%s\"", strings.ToUpper(schema), tableInfo.TableName)
}

// GenerateCreateTableSQL 生成建表SQL - Oracle版本
func (d *OracleDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
	if len(tableInfo.Fields) == 0 {
		return "", fmt.Errorf("no fields specified for table creation")
	}

	var sqlBuilder strings.Builder
	fullTableName := d.GetFullTableName(tableInfo)

	// 简单地生成建表语句，表存在性检查由调用方处理
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

	return sqlBuilder.String(), nil
}

// generateFieldSQL 生成字段SQL - Oracle版本
func (d *OracleDriver) generateFieldSQL(field FieldAttr, isFirst bool) (string, error) {
	var sqlBuilder strings.Builder

	// 字段名（使用双引号转义）
	sqlBuilder.WriteString("                \"")
	sqlBuilder.WriteString(field.Name)
	sqlBuilder.WriteString("\" ")

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
func (d *OracleDriver) mapDataTypeByMapping(field FieldAttr) string {
	mapping := d.GetDataTypeMapping()
	dataTypeKey := strings.ToLower(field.DataType)

	// 首先尝试从映射表中获取基础数据类型
	baseType, exists := mapping[dataTypeKey]
	if !exists {
		// 如果映射表中没有找到，使用默认的NVARCHAR2
		baseType = "NVARCHAR2"
	}

	// 处理需要长度参数的数据类型
	switch strings.ToUpper(field.DataType) {
	case "VARCHAR", "NVARCHAR", "STRING":
		if field.DataLenth <= 0 {
			// 不带长度或长度无效，使用NCLOB
			return "NCLOB"
		} else if field.DataLenth < 2000 {
			// 长度小于2000，使用NVARCHAR2
			return fmt.Sprintf("NVARCHAR2(%d)", field.DataLenth)
		} else if field.DataLenth < 4000 {
			// 长度小于4000大于等于2000，使用VARCHAR2
			return fmt.Sprintf("VARCHAR2(%d)", field.DataLenth)
		} else {
			// 长度大于等于4000，使用NCLOB
			return "NCLOB"
		}

	case "CHAR", "NCHAR":
		if field.DataLenth > 0 && field.DataLenth <= 2000 {
			return fmt.Sprintf("%s(%d)", baseType, field.DataLenth)
		}
		return fmt.Sprintf("%s(1)", baseType)

	case "DECIMAL", "NUMERIC":
		if field.Precision > 0 {
			scale := DefaultDecimalScale
			if field.Precision > 0 && field.Precision <= scale {
				scale = field.Precision
			}
			return fmt.Sprintf("NUMBER(%d, %d)", DefaultDecimalPrecision, scale)
		}
		return fmt.Sprintf("NUMBER")

	default:
		// 对于其他类型，直接使用映射表的结果
		// 如果是默认的NVARCHAR2且没有长度，添加默认长度255
		if baseType == "NVARCHAR2" && !strings.Contains(baseType, "(") {
			return "NVARCHAR2(255)"
		}
		return baseType
	}
}

// GetDataTypeMapping 获取数据类型映射 - Oracle版本
func (d *OracleDriver) GetDataTypeMapping() map[string]string {
	return map[string]string{
		// 通用字符串类型映射
		"string":   "NVARCHAR2",
		"varchar":  "NVARCHAR2",
		"nvarchar": "NVARCHAR2",
		"text":     "CLOB",

		// 整型映射
		"tinyint":  "NUMBER(3)",
		"smallint": "NUMBER(5)",
		"int":      "NUMBER(10)",
		"integer":  "NUMBER(10)",
		"bigint":   "NUMBER(19)",
		"long":     "NUMBER(19)",

		// 浮点型映射
		"float":         "BINARY_FLOAT",
		"real":          "BINARY_FLOAT",
		"double":        "BINARY_DOUBLE",
		"binary_float":  "BINARY_FLOAT",
		"binary_double": "BINARY_DOUBLE",

		// 定点型映射
		"decimal": "NUMBER",
		"numeric": "NUMBER",

		// 布尔型映射
		"bit":     "NUMBER",
		"bool":    "NUMBER",
		"boolean": "NUMBER",

		// 日期和时间类型映射
		"date":                     "DATE",
		"time":                     "DATE",
		"datetime":                 "TIMESTAMP",
		"datetime2":                "TIMESTAMP",
		"timestamp":                "TIMESTAMP(3)",
		"timestamptz":              "TIMESTAMP(3) WITH TIME ZONE",
		"timestamp with time zone": "TIMESTAMP(3) WITH TIME ZONE",

		// 字符串类型映射
		"char":      "CHAR",
		"nchar":     "NCHAR", // 虽然表格说不支持，但保留映射
		"varchar2":  "VARCHAR2",
		"nvarchar2": "NVARCHAR2",
		"clob":      "CLOB",
		"nclob":     "NCLOB",

		// 二进制数据类型映射
		"blob":     "BLOB",
		"binary":   "BLOB",
		"raw":      "BLOB",
		"long raw": "BLOB",
		"bfile":    "BFILE",
	}
}

// SupportSchema 是否支持schema - Oracle版本
func (d *OracleDriver) SupportSchema() bool {
	return true // Oracle支持schema（用户）
}

// SupportBatchInsert 是否支持批量插入 - Oracle版本
func (d *OracleDriver) SupportBatchInsert() bool {
	return true
}

// GetDefaultParams 获取默认连接参数 - Oracle版本
func (d *OracleDriver) GetDefaultParams() map[string]string {
	return map[string]string{
		"SSL": "false",
		// Oracle doesn't use query parameters in the same way as SQL Server
	}
}

// ListTables 列出数据库中的所有表 - Oracle版本
func (d *OracleDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
	var tables []TableMetadata

	// Oracle中，如果没有指定schema，使用当前用户
	if schema == "" {
		schema = "SYS" // Oracle默认使用SYS用户，或者可以通过查询获取当前用户
	}

	query := fmt.Sprintf("SELECT t.TABLE_NAME as name, t.OWNER as table_schema, 'TABLE' as type, NVL(c.COMMENTS, '') as table_comment FROM ALL_TABLES t LEFT JOIN ALL_TAB_COMMENTS c ON c.OWNER = t.OWNER AND c.TABLE_NAME = t.TABLE_NAME WHERE t.OWNER = UPPER('%s') ORDER BY t.TABLE_NAME", schema)

	rows, err := dbConn.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query Oracle tables: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var table TableMetadata
		var commentScanner nullStringScanner
		err := rows.Scan(
			&table.Name,
			&table.Schema,
			&table.Type,
			&commentScanner,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan Oracle table row: %w", err)
		}
		table.Comment = commentScanner.String()
		tables = append(tables, table)
	}

	return tables, nil
}

// ListTableColumns 列出指定表的字段信息 - Oracle版本
func (d *OracleDriver) ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error) {
	var columns []ColumnInfo

	query := fmt.Sprintf(`
		SELECT
			ATC.COLUMN_NAME as name,
			ATC.DATA_TYPE as data_type,
			ATC.DATA_LENGTH as data_length,
			ATC.NULLABLE as is_nullable,
			CASE WHEN PKC.CONSTRAINT_NAME IS NOT NULL THEN 1 ELSE 0 END as primary_key,
			ATC.DATA_DEFAULT as default_value,
			CM.COMMENTS as column_comment,
			ATC.DATA_PRECISION as precision,
			ATC.DATA_SCALE as scale
		FROM ALL_TAB_COLUMNS ATC
		LEFT JOIN (
			SELECT DISTINCT ACC.OWNER, ACC.TABLE_NAME, ACC.COLUMN_NAME, ACC.CONSTRAINT_NAME
			FROM ALL_CONS_COLUMNS ACC
			INNER JOIN ALL_CONSTRAINTS AC
				ON ACC.CONSTRAINT_NAME = AC.CONSTRAINT_NAME
				AND ACC.OWNER = AC.OWNER
				AND AC.CONSTRAINT_TYPE = 'P'
		) PKC
			ON ATC.OWNER = PKC.OWNER
			AND ATC.TABLE_NAME = PKC.TABLE_NAME
			AND ATC.COLUMN_NAME = PKC.COLUMN_NAME
		LEFT JOIN ALL_COL_COMMENTS CM
			ON ATC.OWNER = CM.OWNER
			AND ATC.TABLE_NAME = CM.TABLE_NAME
			AND ATC.COLUMN_NAME = CM.COLUMN_NAME
		WHERE ATC.OWNER = '%s' AND ATC.TABLE_NAME = '%s'
		ORDER BY ATC.COLUMN_ID
	`, strings.ToUpper(schema), strings.ToUpper(tableName))

	rows, err := dbConn.Raw(query).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to query Oracle table columns: %w", err)
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
			return nil, fmt.Errorf("failed to scan Oracle column row: %w", err)
		}

		// 将NULL值转换为适当的格式
		col.DefaultValue = defaultValueScanner.String()
		col.Comment = commentScanner.String()
		col.DataLength = dataLengthScanner.Int()
		col.Precision = precisionScanner.Int()
		col.Scale = scaleScanner.Int()

		// 将Oracle的数据类型转换为更易理解的格式
		col.DataType = normalizeOracleDataType(col.DataType)

		columns = append(columns, col)
	}

	return columns, nil
}

// normalizeOracleDataType 将Oracle的数据类型转换为更易理解的格式
func normalizeOracleDataType(dataType string) string {
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
	default:
		return dataType
	}
}

// EscapeIdentifier 转义Oracle标识符（表名、字段名等）
func (d *OracleDriver) EscapeIdentifier(identifier string) string {
	// Oracle中使用双引号转义标识符
	if strings.ContainsAny(identifier, "\"") {
		// 如果已经包含双引号，进行转义
		return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
	}
	// 检查是否需要转义（包含特殊字符或空格）
	if strings.ContainsAny(identifier, " \t\n\r\f\v`~!@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
		strings.ToUpper(identifier) != identifier { // 如果不是全大写，可能是保留字
		return "\"" + identifier + "\""
	}
	return identifier
}

// isOracleReservedWord 检查是否为Oracle保留字
func (d *OracleDriver) isOracleReservedWord(word string) bool {
	reservedWords := []string{
		"ACCESS", "ADD", "ALL", "ALTER", "AND", "ANY", "AS", "ASC", "AUDIT", "BETWEEN",
		"BY", "CHAR", "CHECK", "CLUSTER", "COLUMN", "COMMENT", "COMPRESS", "CONNECT",
		"CREATE", "CURRENT", "DATE", "DECIMAL", "DEFAULT", "DELETE", "DESC", "DISTINCT",
		"DROP", "ELSE", "EXCLUSIVE", "EXISTS", "FILE", "FLOAT", "FOR", "FROM", "GRANT",
		"GROUP", "HAVING", "IDENTIFIED", "IMMEDIATE", "IN", "INCREMENT", "INDEX", "INITIAL",
		"INSERT", "INTEGER", "INTERSECT", "INTO", "IS", "LEVEL", "LIKE", "LOCK", "LONG",
		"MAXEXTENTS", "MINUS", "MLSLABEL", "MODE", "MODIFY", "NOAUDIT", "NOCOMPRESS",
		"NOT", "NOWAIT", "NULL", "NUMBER", "OF", "OFFLINE", "ON", "ONLINE", "OPTION",
		"OR", "ORDER", "PCTFREE", "PRIOR", "PRIVILEGES", "PUBLIC", "RAW", "RENAME",
		"RESOURCE", "REVOKE", "ROW", "ROWID", "ROWNUM", "ROWS", "SELECT", "SESSION",
		"SET", "SHARE", "SIZE", "SMALLINT", "START", "SUCCESSFUL", "SYNONYM", "SYSDATE",
		"TABLE", "THEN", "TO", "TRIGGER", "UID", "UNION", "UNIQUE", "UPDATE", "USER",
		"VALIDATE", "VALUES", "VARCHAR", "VARCHAR2", "VIEW", "WHENEVER", "WHERE", "WITH",
	}

	wordUpper := strings.ToUpper(word)
	for _, reserved := range reservedWords {
		if wordUpper == reserved {
			return true
		}
	}
	return false
}
