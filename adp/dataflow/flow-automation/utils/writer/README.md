# 数据库写入开发指南

本文档提供了为 flow-automation 项目添加新数据库写入支持的完整开发指南。

## 目录

1. [系统架构概述](#系统架构概述)
2. [核心接口介绍](#核心接口介绍)
3. [开发步骤](#开发步骤)
4. [实现示例](#实现示例)
5. [测试指南](#测试指南)
6. [最佳实践](#最佳实践)
7. [常见问题](#常见问题)

## 系统架构概述

flow-automation 的数据库写入系统采用插件化架构，支持多种数据库类型。主要组件包括：

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│ DatabaseWriter  │    │ DatabaseDriver  │    │ DatabaseExecutor│
│   (统一入口)    │    │   (数据库驱动)   │    │  (执行器实现)   │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │   Registry      │
                    │ (驱动注册管理)   │
                    └─────────────────┘
```

### 核心设计原则

1. **插件化**: 每个数据库类型都是独立的插件
2. **统一接口**: 通过接口抽象屏蔽数据库差异
3. **可扩展性**: 易于添加新的数据库类型支持
4. **容错性**: 完善的错误处理和恢复机制

## 核心接口介绍

### DatabaseDriver 接口

数据库驱动接口定义了数据库相关的核心功能：

```go
type DatabaseDriver interface {
    // 连接相关
    GetDriverName() string
    BuildDSN(config *DatabaseConfig) string
    GetConnection(config *DatabaseConfig) (*gorm.DB, error)

    // 表管理
    CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error
    GetFullTableName(tableInfo *TableInfo) string
    ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error)

    // SQL生成相关
    GenerateCreateTableSQL(tableInfo *TableInfo) (string, error)
    GetDataTypeMapping() map[string]string

    // 特性支持
    SupportSchema() bool
    SupportBatchInsert() bool
    GetDefaultParams() map[string]string
}
```

### DatabaseExecutor 接口

数据库执行器接口定义了数据操作的核心功能：

```go
type DatabaseExecutor interface {
    ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error)
    ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error)
    ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error)
}
```

## 开发步骤

### 步骤 1: 创建驱动文件

为新的数据库类型创建驱动文件 `newdb_driver.go`：

```go
package writer

import (
    "database/sql"
    "fmt"
    "strings"
    "time"

    // 导入对应的 GORM 驱动
    // newdbd "gorm.io/driver/newdb"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// NewDBDriver 新数据库驱动实现
type NewDBDriver struct{}

func (d *NewDBDriver) GetDriverName() string { return "newdb" }

func (d *NewDBDriver) BuildDSN(config *DatabaseConfig) string {
    // 实现 DSN 构建逻辑
    dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
        config.Host, config.Port, config.Username, config.Password, config.Database)

    // 添加自定义参数
    for k, v := range config.Params {
        dsn += fmt.Sprintf(" %s=%s", k, v)
    }

    return dsn
}

func (d *NewDBDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := d.BuildDSN(config)

    sqlDB, err := sql.Open("newdb", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open NewDB connection: %w", err)
    }

    // 配置连接池
    sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
    sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Hour)

    // 创建 GORM 连接
    dial := newdbd.New(newdbd.Config{Conn: sqlDB})
    gormDB, err := gorm.Open(dial, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
    if err != nil {
        sqlDB.Close()
        return nil, fmt.Errorf("failed to create GORM NewDB connection: %w", err)
    }

    return gormDB, nil
}

// 实现其他接口方法...
```

### 步骤 2: 创建执行器文件

创建对应的执行器文件 `newdb_executor.go`：

```go
package writer

import (
    "context"
    "fmt"
    "strings"

    "gorm.io/gorm"
)

// NewDBExecutor 新数据库执行器实现
type NewDBExecutor struct{}

func (e *NewDBExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
    return e.executeNewDBInsert(ctx, dbConn, tableInfo, driver, data)
}

func (e *NewDBExecutor) ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
    return e.executeNewDBUpdate(ctx, dbConn, tableInfo, driver, data, where)
}

func (e *NewDBExecutor) ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
    return e.executeNewDBDelete(ctx, dbConn, tableInfo, driver, where)
}

// executeNewDBInsert 实现插入逻辑
func (e *NewDBExecutor) executeNewDBInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    // 检查表是否存在
    var count int64
    if err := dbConn.Table(fullTableName).Count(&count).Error; err != nil {
        return nil, fmt.Errorf("failed to check table existence: %w", err)
    }

    // 批量插入逻辑
    batchSize := DefaultBatchSize
    if tableInfo.Options != nil && tableInfo.Options.BatchSize > 0 {
        batchSize = tableInfo.Options.BatchSize
    }

    successCount, failedRecords, failureReasons := e.executeBatchInsertWithDetails(ctx, dbConn, tableInfo, driver, data, batchSize)

    // 验证结果
    var newCount int64
    if err := dbConn.Table(fullTableName).Count(&newCount).Error; err != nil {
        return nil, fmt.Errorf("failed to verify write result: %w", err)
    }

    return &ExecutionResult{
        AffectedRows:   successCount,
        Operation:      OperationInsert,
        Table:          fullTableName,
        Success:        len(failedRecords) == 0,
        BeforeCount:    count,
        AfterCount:     newCount,
        SuccessCount:   successCount,
        FailedCount:    int64(len(failedRecords)),
        TotalProcessed: int64(len(data)),
        FailedRecords:  failedRecords,
        FailureReasons: failureReasons,
    }, nil
}

// executeNewDBUpdate 实现更新逻辑
func (e *NewDBExecutor) executeNewDBUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    if where == nil {
        return nil, fmt.Errorf("where condition is required for update operation")
    }

    if len(data) == 0 {
        return nil, fmt.Errorf("data cannot be empty for update operation")
    }

    updateData := data[0]
    result := dbConn.Table(fullTableName).Where(where).Updates(updateData)
    if result.Error != nil {
        return nil, fmt.Errorf("NewDB update failed: %w", result.Error)
    }

    return &ExecutionResult{
        AffectedRows:   result.RowsAffected,
        Operation:      OperationUpdate,
        Table:          fullTableName,
        Success:        true,
        SuccessCount:   result.RowsAffected,
        FailedCount:    0,
        TotalProcessed: 1,
        FailedRecords:  []map[string]interface{}{},
        FailureReasons: map[string]int{},
    }, nil
}

// executeNewDBDelete 实现删除逻辑
func (e *NewDBExecutor) executeNewDBDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    if where == nil {
        return nil, fmt.Errorf("where condition is required for delete operation")
    }

    result := dbConn.Table(fullTableName).Where(where).Delete(nil)
    if result.Error != nil {
        return nil, fmt.Errorf("NewDB delete failed: %w", result.Error)
    }

    return &ExecutionResult{
        AffectedRows:   result.RowsAffected,
        Operation:      OperationDelete,
        Table:          fullTableName,
        Success:        true,
        SuccessCount:   result.RowsAffected,
        FailedCount:    0,
        TotalProcessed: 0, // 删除操作不需要统计处理的数据量
        FailedRecords:  []map[string]interface{}{},
        FailureReasons: map[string]int{},
    }, nil
}

// executeBatchInsertWithDetails 批量插入实现
func (e *NewDBExecutor) executeBatchInsertWithDetails(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, batchSize int) (int64, []map[string]interface{}, map[string]int) {
    fullTableName := driver.GetFullTableName(tableInfo)
    var successCount int64
    var failedRecords []map[string]interface{}
    failureReasons := make(map[string]int)

    // 分批处理数据
    for i := 0; i < len(data); i += batchSize {
        end := i + batchSize
        if end > len(data) {
            end = len(data)
        }

        batch := data[i:end]

        // 使用 GORM 的批量插入
        result := dbConn.Table(fullTableName).CreateInBatches(batch, batchSize)
        if result.Error != nil {
            // 记录失败的记录
            for _, record := range batch {
                failedRecords = append(failedRecords, record)
            }
            failureReasons[result.Error.Error()] += len(batch)
        } else {
            successCount += result.RowsAffected
        }
    }

    return successCount, failedRecords, failureReasons
}
```

### 步骤 3: 实现核心方法

#### 数据类型映射

```go
func (d *NewDBDriver) GetDataTypeMapping() map[string]string {
    return map[string]string{
        "TINYINT":     "SMALLINT",
        "SMALLINT":    "SMALLINT",
        "INT":         "INTEGER",
        "BIGINT":      "BIGINT",
        "DECIMAL":     "DECIMAL",
        "FLOAT":       "REAL",
        "DOUBLE":      "DOUBLE PRECISION",
        "CHAR":        "CHAR",
        "VARCHAR":     "VARCHAR",
        "TEXT":        "TEXT",
        "DATE":        "DATE",
        "DATETIME":    "TIMESTAMP",
        "TIMESTAMP":   "TIMESTAMP",
        "TIME":        "TIME",
        "BOOL":        "BOOLEAN",
        "JSON":        "JSON",
        // 添加新数据库特有的数据类型映射
        "UNIQUEIDENTIFIER": "VARCHAR(36)", // 示例：SQL Server 的 GUID 类型
    }
}
```

#### SQL 生成

```go
func (d *NewDBDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
    if len(tableInfo.Fields) == 0 {
        return "", fmt.Errorf("no field mappings provided")
    }

    fullTableName := d.GetFullTableName(tableInfo)

    var sqlBuilder strings.Builder
    sqlBuilder.WriteString("CREATE TABLE IF NOT EXISTS ")
    sqlBuilder.WriteString(fullTableName)
    sqlBuilder.WriteString(" (")

    columns := make([]string, 0, len(tableInfo.Fields))
    for i := range tableInfo.Fields {
        field := &tableInfo.Fields[i]
        if field.Target.Name == "" {
            continue
        }

        var columnBuilder strings.Builder
        columnBuilder.WriteString(d.EscapeIdentifier(field.Target.Name))
        columnBuilder.WriteString(" ")

        // 使用数据类型映射
        mapping := d.GetDataTypeMapping()
        dataType := field.Target.DataType
        if mappedType, exists := mapping[strings.ToUpper(dataType)]; exists {
            dataType = mappedType
        }

        // 处理长度限制
        if (strings.HasPrefix(strings.ToUpper(dataType), "VARCHAR") ||
            strings.HasPrefix(strings.ToUpper(dataType), "CHAR")) &&
            field.Target.DataLenth > 0 {
            dataType = fmt.Sprintf("%s(%d)", dataType, field.Target.DataLenth)
        }

        // 处理 DECIMAL 精度
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

        columnBuilder.WriteString(dataType)

        // 处理约束
        if field.Target.IsNullable == "NO" {
            columnBuilder.WriteString(" NOT NULL")
        }

        if field.Target.PrimaryKey == PrimaryKeyFlag {
            columnBuilder.WriteString(" PRIMARY KEY")
        }

        columns = append(columns, columnBuilder.String())
    }

    sqlBuilder.WriteString(strings.Join(columns, ", "))
    sqlBuilder.WriteString(")")

    return sqlBuilder.String(), nil
}
```

#### 标识符转义

```go
func (d *NewDBDriver) EscapeIdentifier(identifier string) string {
    // 检查是否需要转义
    if strings.ContainsAny(identifier, "\"@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
        (len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
        strings.Contains(identifier, " ") ||
        d.isReservedWord(strings.ToUpper(identifier)) {
        return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
    }
    return identifier
}

func (d *NewDBDriver) isReservedWord(word string) bool {
    // 返回新数据库的保留字列表
    reservedWords := []string{
        "SELECT", "INSERT", "UPDATE", "DELETE", "CREATE", "DROP",
        "TABLE", "INDEX", "VIEW", "DATABASE", "SCHEMA",
        // 添加更多保留字...
    }

    for _, reserved := range reservedWords {
        if word == reserved {
            return true
        }
    }
    return false
}
```

#### 表名处理

```go
func (d *NewDBDriver) GetFullTableName(tableInfo *TableInfo) string {
    if tableInfo.Conn != nil && tableInfo.Conn.Schema != "" && d.SupportSchema() {
        return fmt.Sprintf("%s.%s", d.EscapeIdentifier(tableInfo.Conn.Schema), d.EscapeIdentifier(tableInfo.TableName))
    }
    return d.EscapeIdentifier(tableInfo.TableName)
}
```

#### 特性支持

```go
func (d *NewDBDriver) SupportSchema() bool {
    // 返回新数据库是否支持 schema
    return true // 或 false，根据数据库特性决定
}

func (d *NewDBDriver) SupportBatchInsert() bool {
    // 返回新数据库是否支持批量插入
    return true
}

func (d *NewDBDriver) GetDefaultParams() map[string]string {
    // 返回新数据库的默认连接参数
    return map[string]string{
        "charset": "utf8mb4",
        "parseTime": "true",
        // 添加其他默认参数...
    }
}
```

### 步骤 4: 注册新驱动

在 `writer.go` 的 `init()` 函数中注册新的数据库驱动：

```go
func init() {
    // ... 现有注册代码 ...

    // 注册新数据库驱动
    newDBDriver := &NewDBDriver{}
    newDBExecutor := &NewDBExecutor{}
    globalRegistry.Register(DatabaseTypeNewDB, newDBDriver, newDBExecutor)
}
```

### 步骤 5: 添加数据库类型常量

在 `types.go` 中添加新的数据库类型常量：

```go
// Database type constants
const (
    DatabaseTypeMySQL      = "mysql"
    DatabaseTypePostgreSQL = "postgresql"
    // ... 其他现有类型 ...
    DatabaseTypeNewDB      = "newdb"  // 添加新数据库类型
)
```

## 实现示例

### 完整实现示例：SQLite 支持

以下是一个完整的 SQLite 数据库支持实现示例：

#### sqlite_driver.go

```go
package writer

import (
    "database/sql"
    "fmt"
    "strings"
    "time"

    sqlite "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

// SQLiteDriver SQLite数据库驱动实现
type SQLiteDriver struct{}

func (d *SQLiteDriver) GetDriverName() string { return "sqlite" }

func (d *SQLiteDriver) BuildDSN(config *DatabaseConfig) string {
    // SQLite 使用文件路径作为 DSN
    if config.Database == "" {
        return ":memory:" // 内存数据库
    }
    return config.Database
}

func (d *SQLiteDriver) GetConnection(config *DatabaseConfig) (*gorm.DB, error) {
    dsn := d.BuildDSN(config)

    sqlDB, err := sql.Open("sqlite3", dsn)
    if err != nil {
        return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
    }

    // SQLite 连接池配置
    sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
    sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
    sqlDB.SetConnMaxLifetime(time.Hour)

    dial := sqlite.New(sqlite.Config{Conn: sqlDB})
    gormDB, err := gorm.Open(dial, &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
    if err != nil {
        sqlDB.Close()
        return nil, fmt.Errorf("failed to create GORM SQLite connection: %w", err)
    }

    return gormDB, nil
}

func (d *SQLiteDriver) CreateTableIfNotExists(dbConn *gorm.DB, tableInfo *TableInfo) error {
    if tableInfo.TableExist {
        return nil
    }

    createSQL, err := d.GenerateCreateTableSQL(tableInfo)
    if err != nil {
        return fmt.Errorf("failed to generate SQLite create table SQL: %w", err)
    }

    if err := dbConn.Exec(createSQL).Error; err != nil {
        if !strings.Contains(err.Error(), "already exists") {
            return fmt.Errorf("failed to create SQLite table: %w", err)
        }
    }

    return nil
}

func (d *SQLiteDriver) GetFullTableName(tableInfo *TableInfo) string {
    // SQLite 不支持 schema，直接返回表名
    return d.EscapeIdentifier(tableInfo.TableName)
}

func (d *SQLiteDriver) EscapeIdentifier(identifier string) string {
    // SQLite 使用双引号或方括号转义标识符
    if strings.ContainsAny(identifier, "\"@#$%^&*()-+=[]{}|\\:;\"'<>?,./") ||
        (len(identifier) > 0 && identifier[0] >= '0' && identifier[0] <= '9') ||
        strings.Contains(identifier, " ") ||
        d.isReservedWord(strings.ToUpper(identifier)) {
        return "\"" + strings.ReplaceAll(identifier, "\"", "\"\"") + "\""
    }
    return identifier
}

func (d *SQLiteDriver) isReservedWord(word string) bool {
    reservedWords := []string{
        "ABORT", "ACTION", "ADD", "AFTER", "ALL", "ALTER", "ANALYZE", "AND", "AS",
        "ASC", "ATTACH", "AUTOINCREMENT", "BEFORE", "BEGIN", "BETWEEN", "BY",
        "CASCADE", "CASE", "CAST", "CHECK", "COLLATE", "COLUMN", "COMMIT",
        "CONFLICT", "CONSTRAINT", "CREATE", "CROSS", "CURRENT_DATE", "CURRENT_TIME",
        "CURRENT_TIMESTAMP", "DATABASE", "DEFAULT", "DEFERRABLE", "DEFERRED",
        "DELETE", "DESC", "DETACH", "DISTINCT", "DROP", "EACH", "ELSE", "END",
        "ESCAPE", "EXCEPT", "EXCLUSIVE", "EXISTS", "EXPLAIN", "FAIL", "FOR",
        "FOREIGN", "FROM", "FULL", "GLOB", "GROUP", "HAVING", "IF", "IGNORE",
        "IMMEDIATE", "IN", "INDEX", "INDEXED", "INITIALLY", "INNER", "INSERT",
        "INSTEAD", "INTERSECT", "INTO", "IS", "ISNULL", "JOIN", "KEY", "LEFT",
        "LIKE", "LIMIT", "MATCH", "NATURAL", "NO", "NOT", "NOTNULL", "NULL",
        "OF", "OFFSET", "ON", "OR", "ORDER", "OUTER", "PLAN", "PRAGMA", "PRIMARY",
        "QUERY", "RAISE", "RECURSIVE", "REFERENCES", "REGEXP", "REINDEX",
        "RELEASE", "RENAME", "REPLACE", "RESTRICT", "RIGHT", "ROLLBACK", "ROW",
        "SAVEPOINT", "SELECT", "SET", "TABLE", "TEMP", "TEMPORARY", "THEN",
        "TO", "TRANSACTION", "TRIGGER", "UNION", "UNIQUE", "UPDATE", "USING",
        "VACUUM", "VALUES", "VIEW", "VIRTUAL", "WHEN", "WHERE", "WITH", "WITHOUT",
    }

    for _, reserved := range reservedWords {
        if word == reserved {
            return true
        }
    }
    return false
}

func (d *SQLiteDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
    if len(tableInfo.Fields) == 0 {
        return "", fmt.Errorf("no field mappings provided")
    }

    fullTableName := d.GetFullTableName(tableInfo)

    var sqlBuilder strings.Builder
    sqlBuilder.WriteString("CREATE TABLE IF NOT EXISTS ")
    sqlBuilder.WriteString(fullTableName)
    sqlBuilder.WriteString(" (")

    columns := make([]string, 0, len(tableInfo.Fields))
    for i := range tableInfo.Fields {
        field := &tableInfo.Fields[i]
        if field.Target.Name == "" {
            continue
        }

        var columnBuilder strings.Builder
        columnBuilder.WriteString(d.EscapeIdentifier(field.Target.Name))
        columnBuilder.WriteString(" ")

        mapping := d.GetDataTypeMapping()
        dataType := field.Target.DataType
        if mappedType, exists := mapping[strings.ToUpper(dataType)]; exists {
            dataType = mappedType
        }

        // SQLite 特殊处理
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

        columnBuilder.WriteString(dataType)

        if field.Target.IsNullable == "NO" {
            columnBuilder.WriteString(" NOT NULL")
        }

        if field.Target.PrimaryKey == PrimaryKeyFlag {
            columnBuilder.WriteString(" PRIMARY KEY")
        }

        columns = append(columns, columnBuilder.String())
    }

    sqlBuilder.WriteString(strings.Join(columns, ", "))
    sqlBuilder.WriteString(")")

    return sqlBuilder.String(), nil
}

func (d *SQLiteDriver) GetDataTypeMapping() map[string]string {
    return map[string]string{
        "TINYINT":     "INTEGER",
        "SMALLINT":    "INTEGER",
        "INT":         "INTEGER",
        "BIGINT":      "INTEGER",
        "DECIMAL":     "REAL",    // SQLite 使用 REAL 存储浮点数
        "FLOAT":       "REAL",
        "DOUBLE":      "REAL",
        "CHAR":        "TEXT",
        "VARCHAR":     "TEXT",
        "TEXT":        "TEXT",
        "DATE":        "TEXT",    // SQLite 没有专门的日期类型
        "DATETIME":    "TEXT",
        "TIMESTAMP":   "TEXT",
        "TIME":        "TEXT",
        "BOOL":        "INTEGER", // SQLite 使用 0/1 表示布尔值
        "BOOLEAN":     "INTEGER",
        "JSON":        "TEXT",
    }
}

func (d *SQLiteDriver) SupportSchema() bool      { return false }
func (d *SQLiteDriver) SupportBatchInsert() bool { return true }

func (d *SQLiteDriver) GetDefaultParams() map[string]string {
    return map[string]string{
        "_journal_mode": "WAL",
        "_synchronous":  "NORMAL",
    }
}

func (d *SQLiteDriver) ListTables(dbConn *gorm.DB, schema string) ([]TableMetadata, error) {
    var tables []TableMetadata

    query := `
        SELECT name, '', 'table', ''
        FROM sqlite_master
        WHERE type='table' AND name NOT LIKE 'sqlite_%'
        ORDER BY name
    `

    rows, err := dbConn.Raw(query).Rows()
    if err != nil {
        return nil, fmt.Errorf("failed to query SQLite tables: %w", err)
    }
    defer rows.Close()

    for rows.Next() {
        var table TableMetadata
        err := rows.Scan(&table.Name, &table.Schema, &table.Type, &table.Comment)
        if err != nil {
            return nil, fmt.Errorf("failed to scan SQLite table row: %w", err)
        }
        tables = append(tables, table)
    }

    return tables, nil
}
```

#### sqlite_executor.go

```go
package writer

import (
    "context"
    "fmt"

    "gorm.io/gorm"
)

// SQLiteExecutor SQLite执行器实现
type SQLiteExecutor struct{}

func (e *SQLiteExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
    return e.executeSQLiteInsert(ctx, dbConn, tableInfo, driver, data)
}

func (e *SQLiteExecutor) ExecuteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
    return e.executeSQLiteUpdate(ctx, dbConn, tableInfo, driver, data, where)
}

func (e *SQLiteExecutor) ExecuteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
    return e.executeSQLiteDelete(ctx, dbConn, tableInfo, driver, where)
}

func (e *SQLiteExecutor) executeSQLiteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    var count int64
    if err := dbConn.Table(fullTableName).Count(&count).Error; err != nil {
        return nil, fmt.Errorf("failed to check table existence: %w", err)
    }

    batchSize := DefaultBatchSize
    if tableInfo.Options != nil && tableInfo.Options.BatchSize > 0 {
        batchSize = tableInfo.Options.BatchSize
    }

    successCount, failedRecords, failureReasons := e.executeBatchInsertWithDetails(ctx, dbConn, tableInfo, driver, data, batchSize)

    var newCount int64
    if err := dbConn.Table(fullTableName).Count(&newCount).Error; err != nil {
        return nil, fmt.Errorf("failed to verify write result: %w", err)
    }

    return &ExecutionResult{
        AffectedRows:   successCount,
        Operation:      OperationInsert,
        Table:          fullTableName,
        Success:        len(failedRecords) == 0,
        BeforeCount:    count,
        AfterCount:     newCount,
        SuccessCount:   successCount,
        FailedCount:    int64(len(failedRecords)),
        TotalProcessed: int64(len(data)),
        FailedRecords:  failedRecords,
        FailureReasons: failureReasons,
    }, nil
}

func (e *SQLiteExecutor) executeSQLiteUpdate(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, where interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    if where == nil {
        return nil, fmt.Errorf("where condition is required for update operation")
    }

    if len(data) == 0 {
        return nil, fmt.Errorf("data cannot be empty for update operation")
    }

    updateData := data[0]
    result := dbConn.Table(fullTableName).Where(where).Updates(updateData)
    if result.Error != nil {
        return nil, fmt.Errorf("SQLite update failed: %w", result.Error)
    }

    return &ExecutionResult{
        AffectedRows:   result.RowsAffected,
        Operation:      OperationUpdate,
        Table:          fullTableName,
        Success:        true,
        SuccessCount:   result.RowsAffected,
        FailedCount:    0,
        TotalProcessed: 1,
        FailedRecords:  []map[string]interface{}{},
        FailureReasons: map[string]int{},
    }, nil
}

func (e *SQLiteExecutor) executeSQLiteDelete(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, where interface{}) (*ExecutionResult, error) {
    fullTableName := driver.GetFullTableName(tableInfo)

    if where == nil {
        return nil, fmt.Errorf("where condition is required for delete operation")
    }

    result := dbConn.Table(fullTableName).Where(where).Delete(nil)
    if result.Error != nil {
        return nil, fmt.Errorf("SQLite delete failed: %w", result.Error)
    }

    return &ExecutionResult{
        AffectedRows:   result.RowsAffected,
        Operation:      OperationDelete,
        Table:          fullTableName,
        Success:        true,
        SuccessCount:   result.RowsAffected,
        FailedCount:    0,
        TotalProcessed: 0,
        FailedRecords:  []map[string]interface{}{},
        FailureReasons: map[string]int{},
    }, nil
}

func (e *SQLiteExecutor) executeBatchInsertWithDetails(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}, batchSize int) (int64, []map[string]interface{}, map[string]int) {
    fullTableName := driver.GetFullTableName(tableInfo)
    var successCount int64
    var failedRecords []map[string]interface{}
    failureReasons := make(map[string]int)

    for i := 0; i < len(data); i += batchSize {
        end := i + batchSize
        if end > len(data) {
            end = len(data)
        }

        batch := data[i:end]

        result := dbConn.Table(fullTableName).CreateInBatches(batch, batchSize)
        if result.Error != nil {
            for _, record := range batch {
                failedRecords = append(failedRecords, record)
            }
            failureReasons[result.Error.Error()] += len(batch)
        } else {
            successCount += result.RowsAffected
        }
    }

    return successCount, failedRecords, failureReasons
}
```

## 测试指南

### 单元测试

创建测试文件 `newdb_driver_test.go`：

```go
package writer

import (
    "testing"

    "github.com/stretchr/testify/assert"
)

func TestNewDBDriver_GetDriverName(t *testing.T) {
    driver := &NewDBDriver{}
    assert.Equal(t, "newdb", driver.GetDriverName())
}

func TestNewDBDriver_BuildDSN(t *testing.T) {
    driver := &NewDBDriver{}
    config := &DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "user",
        Password: "pass",
        Database: "testdb",
        Params: map[string]string{
            "charset": "utf8mb4",
        },
    }

    dsn := driver.BuildDSN(config)
    expected := "host=localhost port=3306 user=user password=pass dbname=testdb charset=utf8mb4"
    assert.Equal(t, expected, dsn)
}

func TestNewDBDriver_GetDataTypeMapping(t *testing.T) {
    driver := &NewDBDriver{}
    mapping := driver.GetDataTypeMapping()

    assert.Equal(t, "INTEGER", mapping["INT"])
    assert.Equal(t, "VARCHAR", mapping["STRING"])
    assert.Equal(t, "TIMESTAMP", mapping["DATETIME"])
}

func TestNewDBDriver_EscapeIdentifier(t *testing.T) {
    driver := &NewDBDriver{}

    // 测试正常标识符
    assert.Equal(t, "column_name", driver.EscapeIdentifier("column_name"))

    // 测试需要转义的标识符
    assert.Equal(t, "\"select\"", driver.EscapeIdentifier("select"))
    assert.Equal(t, "\"column-name\"", driver.EscapeIdentifier("column-name"))
    assert.Equal(t, "\"123column\"", driver.EscapeIdentifier("123column"))
}
```

### 集成测试

```go
package writer

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestNewDBExecutor_ExecuteInsert(t *testing.T) {
    // 创建测试数据库连接
    config := &DatabaseConfig{
        Host:     "localhost",
        Port:     3306,
        Username: "testuser",
        Password: "testpass",
        Database: "testdb",
    }

    registry := NewDatabaseDriverRegistry()
    driver := &NewDBDriver{}
    executor := &NewDBExecutor{}
    registry.Register(DatabaseTypeNewDB, driver, executor)

    writer := NewUniversalDatabaseWriter(registry)

    // 测试表信息
    tableInfo := &TableInfo{
        DatasourceType: DatabaseTypeNewDB,
        TableName:      "test_table",
        Conn: &DBConn{
            Host:     config.Host,
            Port:     config.Port,
            Username: config.Username,
            Password: config.Password,
            Database: config.Database,
        },
        Fields: []FieldMapping{
            {
                Target: FieldAttr{
                    Name:     "id",
                    DataType: "INT",
                    IsNullable: "NO",
                    PrimaryKey: PrimaryKeyFlag,
                },
            },
            {
                Target: FieldAttr{
                    Name:     "name",
                    DataType: "VARCHAR",
                    DataLenth: 255,
                },
            },
        },
    }

    // 测试数据
    testData := []map[string]interface{}{
        {"id": 1, "name": "Test User 1"},
        {"id": 2, "name": "Test User 2"},
    }

    // 执行插入操作
    result, err := writer.Execute(context.Background(), tableInfo, testData, nil, OperationInsert)

    require.NoError(t, err)
    assert.True(t, result.Success)
    assert.Equal(t, int64(2), result.SuccessCount)
    assert.Equal(t, OperationInsert, result.Operation)
}
```

## 最佳实践

### 1. 错误处理

- 始终返回详细的错误信息，包括上下文
- 使用 `fmt.Errorf` 包装底层错误，不要丢失错误链
- 对于已存在的表等非错误情况，使用特定的错误信息标识

```go
// 推荐的错误处理方式
if err := dbConn.Exec(createSQL).Error; err != nil {
    // 检查是否是表已存在的错误
    if strings.Contains(err.Error(), "already exists") ||
        strings.Contains(err.Error(), "already_exists") ||
        strings.Contains(err.Error(), "1050") { // MySQL 错误码
        return nil // 表已存在不是错误
    }
    return fmt.Errorf("failed to create table: %w", err)
}
```

### 2. 连接管理

- 正确设置连接池参数
- 及时关闭连接，避免资源泄漏
- 处理连接超时和重连逻辑

```go
sqlDB.SetMaxIdleConns(DefaultMaxIdleConns)
sqlDB.SetMaxOpenConns(DefaultMaxOpenConns)
sqlDB.SetConnMaxLifetime(time.Hour)

// 确保在函数返回时关闭连接
defer func() {
    if sqlDB != nil {
        sqlDB.Close()
    }
}()
```

### 3. SQL 注入防护

- 使用参数化查询，不要拼接 SQL
- 正确转义标识符（表名、字段名）
- 验证输入数据的合法性

```go
// 正确的方式：使用参数化查询
result := dbConn.Table(tableName).Where("id = ?", id).First(&record)

// 错误的方式：字符串拼接（有 SQL 注入风险）
query := "SELECT * FROM " + tableName + " WHERE id = " + id
```

### 4. 批量操作优化

- 根据数据库特性选择合适的批量大小
- 处理批量操作中的部分失败
- 提供详细的执行结果统计

```go
func executeBatchInsertWithDetails(...) (int64, []map[string]interface{}, map[string]int) {
    var successCount int64
    var failedRecords []map[string]interface{}
    failureReasons := make(map[string]int)

    // 分批处理，记录失败详情
    for i := 0; i < len(data); i += batchSize {
        // ... 批量处理逻辑 ...
    }

    return successCount, failedRecords, failureReasons
}
```

### 5. 数据类型映射

- 提供完整的通用数据类型映射
- 处理数据库特有的数据类型
- 支持长度、精度等参数

```go
func (d *NewDBDriver) GetDataTypeMapping() map[string]string {
    return map[string]string{
        // 标准映射
        "INT":       "INTEGER",
        "VARCHAR":   "VARCHAR",
        "DECIMAL":   "DECIMAL",

        // 数据库特有类型
        "UNIQUEIDENTIFIER": "VARCHAR(36)", // SQL Server GUID
        "MONEY":           "DECIMAL(19,4)", // SQL Server Money
    }
}
```

## 常见问题

### Q: 如何处理数据库特有的 SQL 语法？

A: 在驱动的相应方法中检查数据库类型，并使用条件分支处理不同的语法：

```go
func (d *MySQLDriver) GenerateCreateTableSQL(tableInfo *TableInfo) (string, error) {
    // MySQL 特有的语法处理
    sql := "CREATE TABLE IF NOT EXISTS " + tableName + " (...) ENGINE=InnoDB"
    return sql, nil
}
```

### Q: 如何支持事务操作？

A: 在执行器中添加事务支持：

```go
func (e *NewDBExecutor) ExecuteInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, data []map[string]interface{}) (*ExecutionResult, error) {
    return dbConn.Transaction(func(tx *gorm.DB) error {
        // 在事务中执行操作
        result, err := e.executeInsertInTransaction(ctx, tx, tableInfo, driver, data)
        return err
    })
}
```

### Q: 如何处理大文件或大数据量的写入？

A: 实现流式处理和内存优化：

```go
func (e *NewDBExecutor) ExecuteLargeInsert(ctx context.Context, dbConn *gorm.DB, tableInfo *TableInfo, driver DatabaseDriver, dataChannel <-chan map[string]interface{}) (*ExecutionResult, error) {
    // 使用 channel 处理大数据流
    // 实现内存友好的批量处理
}
```

### Q: 如何添加新的数据库特性支持？

A: 
1. 在接口中添加新方法
2. 在所有现有驱动中实现该方法
3. 在执行器中使用新特性
4. 更新文档和测试

```go
// 在接口中添加新方法
type DatabaseDriver interface {
    // ... 现有方法 ...
    SupportUpsert() bool
    GenerateUpsertSQL(tableInfo *TableInfo, conflictColumns []string) (string, error)
}
```