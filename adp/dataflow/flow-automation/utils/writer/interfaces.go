package writer

import (
	"context"

	"gorm.io/gorm"
)

// DatabaseDriver 数据库驱动接口
type DatabaseDriver interface {
	// 连接相关
	GetDriverName() string
	BuildDSN(config *DatabaseConfig) string

	GetConnection(config *DatabaseConfig) (*gorm.DB, error)

	// 表管理 (可选实现)
	CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error
	GetFullTableName(tableInfo *TableInfo) string
	ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error)
	ListTableColumns(dbConn *gorm.DB, tableName, schema string) ([]ColumnInfo, error)

	// SQL生成相关
	GenerateCreateTableSQL(tableInfo *TableInfo) (string, error)
	GetDataTypeMapping() map[string]string

	// 特性支持
	SupportSchema() bool
	SupportBatchInsert() bool
	GetDefaultParams() map[string]string
}

// DatabaseExecutor 数据库执行器接口
type DatabaseExecutor interface {
	ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error)
	ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error)
	ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error)
}

// DatabaseWriter 数据库写入器接口
type DatabaseWriter interface {
	// 连接管理
	GetDBConnection(config *DatabaseConfig) (*gorm.DB, error)

	// 表管理
	CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error

	// 数据操作
	Execute(ctx context.Context, tableInfo *TableInfo, data []map[string]interface{}, where interface{}, operation string) (*ExecutionResult, error)
}
