# ORM Helper 架构设计

## 🏗️ 整体架构

ORM Helper 采用模块化设计，每个组件专注单一职责，通过接口解耦，保持代码简洁和可维护性。

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Application   │    │     DAO Layer   │    │  ORM Helper     │
│                 │───▶│                 │───▶│                 │
│  业务逻辑层       │    │   数据访问层     │    │  数据库抽象层    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                       │
                                                       ▼
                                               ┌─────────────────┐
                                               │   sqlx.DB       │
                                               │   sql.Tx        │
                                               │  现有数据库连接   │
                                               └─────────────────┘
```

## 📦 核心组件

### 1. 接口层 (`interfaces.go`)

定义核心抽象接口，确保组件间的解耦和可测试性。

```go
// 执行器接口 - 兼容 sqlx.DB 和 sql.Tx
type Executor interface {
    ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
    QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
    QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// 结果扫描器接口 - 处理查询结果到结构体的映射
type Scanner interface {
    ScanStruct(rows *sql.Rows, dest interface{}) error
    ScanStructs(rows *sql.Rows, dest interface{}) error
}
```

**设计要点：**
- 最小化接口，只暴露必要方法
- 完全兼容标准库和 sqlx
- 支持事务和非事务场景

### 2. ORM 核心类 (`orm.go`)

负责管理数据库连接和创建各种构建器。

```go
type DB struct {
    executor Executor   // 数据库执行器
    dbName   string     // 数据库名称
    scanner  Scanner    // 结果扫描器
}
```

**核心方法：**
- `New(executor, dbName)` - 创建 ORM 实例
- `WithTx(tx)` - 事务模式切换
- `Select()` - 创建查询构建器
- `Insert()` - 创建插入构建器
- `Update(table)` - 创建更新构建器
- `Delete()` - 创建删除构建器

**设计要点：**
- 单一职责：只负责创建和管理构建器
- 无状态设计：每次操作创建新的构建器
- 事务透明：自动处理事务和非事务场景

### 3. 查询构建器 (`select.go`)

负责构建 SELECT 查询语句。

```go
type SelectBuilder struct {
    db       *DB
    columns  []string
    tables   []string
    joins    []JoinClause
    where    *WhereBuilder
    groupBy  []string
    having   *WhereBuilder
    orderBy  []OrderClause
    limit    *int
    offset   *int
}
```

**主要功能：**
- 列选择：`Select(columns...)`
- 表连接：`From()`, `Join()`, `LeftJoin()`
- 条件过滤：`Where...()` 系列方法
- 分组聚合：`GroupBy()`, `Having()`
- 排序分页：`OrderBy()`, `Limit()`, `Offset()`

**设计要点：**
- 流畅 API：支持方法链式调用
- 延迟执行：构建和执行分离
- 类型安全：结构体自动映射

### 4. 条件构建器 (`where.go`)

专门处理 WHERE 和 HAVING 条件。

```go
type WhereBuilder struct {
    conditions []Condition
    logicOp    string // AND 或 OR
}
```

**条件类型：**
- 基本比较：`Eq`, `Ne`, `Gt`, `Lt`, `Gte`, `Lte`
- NULL 检查：`IsNull`, `IsNotNull`
- 范围条件：`In`, `NotIn`, `Between`
- 模糊匹配：`Like`, `NotLike`
- 复杂组合：`And()`, `Or()`

**设计要点：**
- 类型安全的参数处理
- 自动 SQL 注入防护
- 支持嵌套条件组合

### 5. 修改构建器

#### 插入构建器 (`insert.go`)
```go
type InsertBuilder struct {
    db          *DB
    table       string
    columns     []string
    values      [][]interface{}
    onDuplicate map[string]interface{}
    ignore      bool
}
```

#### 更新构建器 (`update.go`)
```go
type UpdateBuilder struct {
    db     *DB
    table  string
    sets   map[string]interface{}
    where  *WhereBuilder
}
```

#### 删除构建器 (`delete.go`)
```go
type DeleteBuilder struct {
    db    *DB
    table string
    where *WhereBuilder
    limit *int
}
```

**设计要点：**
- 统一的 API 风格
- 参数验证和错误处理
- 支持批量操作

### 6. 结果扫描器 (`scanner.go`)

负责将查询结果映射到 Go 结构体。

```go
type DefaultScanner struct{}

func (s *DefaultScanner) ScanStruct(rows *sql.Rows, dest interface{}) error
func (s *DefaultScanner) ScanStructs(rows *sql.Rows, dest interface{}) error
```

**核心功能：**
- 基于反射的字段映射
- 支持 `db` 标签自定义映射
- 处理数据类型转换
- 错误处理和验证

**设计要点：**
- 高性能反射使用
- 类型安全转换
- 详细的错误信息

## 🔄 数据流程

### 查询流程
```
Application
    │
    ▼
orm.Select().From("table").WhereEq("id", 1)
    │
    ▼
SelectBuilder.Get(ctx, &result)
    │
    ▼
SQL 生成 + 参数绑定
    │
    ▼
Executor.QueryContext(ctx, sql, args...)
    │
    ▼
Scanner.ScanStruct(rows, &result)
    │
    ▼
返回结果给 Application
```

### 事务流程
```
Application
    │
    ▼
tx := dbPool.BeginTx(ctx, nil)
    │
    ▼
txORM := orm.WithTx(tx)
    │
    ▼
txORM.Insert().Into("table").Values(data).Execute(ctx)
    │
    ▼
tx.Commit() 或 tx.Rollback()
```

## 🎯 设计原则

### 1. 简单性
- API 设计贴近 SQL 语法
- 最小化学习成本
- 避免过度抽象

### 2. 兼容性
- 完全兼容现有 `sqlx.DB`
- 支持渐进式迁移
- 保持向后兼容

### 3. 类型安全
- 编译时错误检查
- 结构体自动映射
- 参数类型验证

### 4. 性能
- 延迟 SQL 生成
- 高效的反射使用
- 最小化内存分配

### 5. 可扩展性
- 接口驱动设计
- 组件松耦合
- 易于添加新功能

## 🔧 扩展指南

### 添加新的条件类型
```go
// 在 WhereBuilder 中添加新方法
func (w *WhereBuilder) WhereRegex(column, pattern string) *WhereBuilder {
    return w.addCondition("REGEXP", column, pattern)
}
```

### 自定义扫描器
```go
// 实现 Scanner 接口
type CustomScanner struct {
    DefaultScanner
}

func (s *CustomScanner) ScanStruct(rows *sql.Rows, dest interface{}) error {
    // 自定义扫描逻辑
    return s.DefaultScanner.ScanStruct(rows, dest)
}

// 使用自定义扫描器
orm := ormhelper.NewWithScanner(executor, dbName, &CustomScanner{})
```

### 添加查询钩子
```go
// 可以在构建器中添加钩子机制
type QueryHook interface {
    BeforeQuery(sql string, args []interface{}) (string, []interface{})
    AfterQuery(result interface{}, err error) error
}
```

## 💡 最佳实践

1. **DAO 层封装**：将 ORM 操作封装在 DAO 层中
2. **事务管理**：统一事务处理模式
3. **错误处理**：详细的错误信息和恰当的错误类型
4. **性能优化**：选择必要列，使用索引，合理分页
5. **代码组织**：按业务模块组织 DAO 和相关结构体

## 🚀 未来规划

- 查询缓存支持
- 读写分离
- 连接池监控
- SQL 日志和性能分析
- 更多数据库方言支持