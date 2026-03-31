package writer

import (
	"database/sql"
)

// Database type constants
const (
	DatabaseTypeMySQL      = "mysql"
	DatabaseTypeMariaDB    = "mariadb"
	DatabaseTypeMaria      = "maria"
	DatabaseTypePostgreSQL = "postgresql"
	DatabaseTypePostgres   = "postgres"
	DatabaseTypeDM8        = "dm8"
	DatabaseTypeDM         = "dameng"
	DatabaseTypeKDB        = "kingbase"
	DatabaseTypeSQLServer  = "sqlserver"
	DatabaseTypeMSSQL      = "mssql"
	DatabaseTypeOracle     = "oracle"
)

// Database connection constants
const (
	DefaultMaxIdleConns     = 10
	DefaultMaxOpenConns     = 100
	DefaultBatchSize        = 1000
	DefaultVarcharLength    = 255
	DefaultDecimalPrecision = 2
	DefaultDecimalScale     = 10
	DefaultEnumLength       = 50
	PrimaryKeyFlag          = 1
)

// Mathematical constants
const (
	DecimalBase           = 10 // Base for decimal number system
	MinBatchDataThreshold = 10 // Minimum data length for batch operations
)

// Database operation constants
const (
	OperationInsert = "insert"
	OperationUpdate = "update"
	OperationDelete = "delete"
	OperationAppend = "append"
)

// Error type constants
const (
	ErrorTypeDuplicateKey            = "duplicate_key"
	ErrorTypeForeignKeyConstraint    = "foreign_key_constraint"
	ErrorTypeNullConstraint          = "null_constraint"
	ErrorTypeDataTooLong             = "data_too_long"
	ErrorTypeDataTypeMismatch        = "data_type_mismatch"
	ErrorTypeFieldNotExist           = "field_not_exist"
	ErrorTypeTimeout                 = "timeout"
	ErrorTypeConnectionError         = "connection_error"
	ErrorTypeTransactionCommitFailed = "transaction_commit_failed"
	ErrorTypeUnknown                 = "unknown_error"
)

// DBConn 连接信息
type DBConn struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Schema   string            `json:"schema,omitempty"`
	Params   map[string]string `json:"params,omitempty"`
}

// FieldAttr 字段元信息
type FieldAttr struct {
	Name       string `json:"name"`
	Comment    string `json:"comment,omitempty"`
	DataLenth  int    `json:"data_lenth,omitempty"` // DECIMAL类型复用为总位数(scale)
	DataType   string `json:"data_type,omitempty"`
	IsNullable string `json:"is_nullable,omitempty"` // "YES"/"NO"
	PrimaryKey int    `json:"primary_key,omitempty"` // 0/1
	Precision  int    `json:"precision,omitempty"`   // DECIMAL类型的小数位数精度
}

// ColumnInfo 列信息结构（用于表字段查询）
type ColumnInfo struct {
	Name         string `json:"name"`                    // 列名
	DataType     string `json:"data_type"`               // 数据类型
	DataLength   int    `json:"data_length,omitempty"`   // 数据长度
	IsNullable   string `json:"is_nullable"`             // 是否可为空 YES/NO
	PrimaryKey   int    `json:"primary_key,omitempty"`   // 是否主键 0/1
	DefaultValue string `json:"default_value,omitempty"` // 默认值
	Comment      string `json:"comment,omitempty"`       // 注释
	Precision    int    `json:"precision,omitempty"`     // 精度（DECIMAL类型）
	Scale        int    `json:"scale,omitempty"`         // 小数位数（DECIMAL类型）
}

// nullStringScanner 用于处理可能为NULL的字符串字段
type nullStringScanner struct {
	value sql.NullString
}

func (ns *nullStringScanner) Scan(value interface{}) error {
	return ns.value.Scan(value)
}

func (ns *nullStringScanner) String() string {
	if ns.value.Valid {
		return ns.value.String
	}
	return ""
}

// nullIntScanner 用于处理可能为NULL的整数字段
type nullIntScanner struct {
	value sql.NullInt64
}

func (ni *nullIntScanner) Scan(value interface{}) error {
	return ni.value.Scan(value)
}

func (ni *nullIntScanner) Int() int {
	if ni.value.Valid {
		return int(ni.value.Int64)
	}
	return 0
}

// FieldMapping 字段映射（source -> target）
type FieldMapping struct {
	Target FieldAttr `json:"target"`
	Source FieldAttr `json:"source"`
}

// SyncOptions 额外写入配置
type SyncOptions struct {
	BatchSize           int  `json:"batch_size,omitempty"`
	TruncateBeforeWrite bool `json:"truncate_before_write,omitempty"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string            `json:"host"`
	Port     int               `json:"port"`
	Username string            `json:"username"`
	Password string            `json:"password"`
	Database string            `json:"database"`
	Driver   string            `json:"driver"`
	Params   map[string]string `json:"params,omitempty"`
}

// TableInfo 表信息
type TableInfo struct {
	DatasourceType string         `json:"datasource_type"`
	TableName      string         `json:"table_name"`
	TableExist     bool           `json:"table_exist,omitempty"`
	Conn           *DBConn        `json:"conn"`
	Fields         []FieldMapping `json:"sync_model_fields,omitempty"`
	Options        *SyncOptions   `json:"sync_options,omitempty"`
}

// TableMetadata 表元数据信息
type TableMetadata struct {
	Name    string `json:"name"`
	Schema  string `json:"schema,omitempty"`
	Type    string `json:"type,omitempty"` // TABLE, VIEW, etc.
	Comment string `json:"comment,omitempty"`
}

// ExecutionResult 执行结果
type ExecutionResult struct {
	AffectedRows   int64                    `json:"affected_rows"`
	Operation      string                   `json:"operation"`
	Table          string                   `json:"table"`
	Success        bool                     `json:"success"`
	BeforeCount    int64                    `json:"before_count,omitempty"`
	AfterCount     int64                    `json:"after_count,omitempty"`
	SuccessCount   int64                    `json:"success_count"`
	FailedCount    int64                    `json:"failed_count"`
	TotalProcessed int64                    `json:"total_processed"`
	FailedRecords  []map[string]interface{} `json:"failed_records"`
	FailureReasons map[string]int           `json:"failure_reasons"`
	Message        string                   `json:"message,omitempty"`
}
