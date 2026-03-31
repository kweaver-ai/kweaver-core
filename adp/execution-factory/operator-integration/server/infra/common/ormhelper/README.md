# 🚀 ORM Helper

一个轻量级、高性能的Go语言ORM工具库，专为简化数据库操作而设计。支持链式调用、事务管理、日志监控等功能。

## 📋 目录

- [✨ 特性](#-特性)
- [🚀 快速入门](#-快速入门)
  - [安装](#安装)
  - [基础使用](#基础使用)
  - [5分钟上手](#5分钟上手)
- [📖 基础教程](#-基础教程)
  - [查询操作](#查询操作)
  - [插入操作](#插入操作)
  - [更新操作](#更新操作)
  - [删除操作](#删除操作)
- [🔧 高级功能](#-高级功能)
  - [事务管理](#事务管理)
  - [日志功能](#日志功能)
  - [复杂查询](#复杂查询)
  - [DAO模式集成](#dao模式集成)
- [⚡ 性能优化](#-性能优化)
- [🛠️ 最佳实践](#️-最佳实践)
- [❓ 常见问题](#-常见问题)
- [🏗️ 架构设计](#️-架构设计)

## ✨ 特性

- 🔗 **链式调用**：流畅的API设计，代码更简洁
- 🔄 **事务支持**：完整的事务管理，确保数据一致性
- 📊 **日志监控**：内置SQL执行日志，支持慢查询检测
- 🎯 **类型安全**：基于反射的结构体映射，编译时类型检查
- ⚡ **高性能**：最小化反射使用，优化的SQL构建
- 🔌 **兼容性强**：支持标准database/sql和sqlx
- 🧪 **易测试**：支持Mock测试，便于单元测试

## 🚀 快速入门

### 安装

```bash
go get github.com/kweaver-ai/operator-hub/operator-integration/server/infra/common/ormhelper
```

### 基础使用

```go
import (
    "github.com/kweaver-ai/operator-hub/operator-integration/server/infra/common/ormhelper"
    "github.com/kweaver-ai/operator-hub/operator-integration/server/infra/logger"
)

// 1. 创建ORM实例
dbPool := db.NewDBPool() // 你的数据库连接池
orm := ormhelper.New(dbPool, "your_database_name")

// 2. 定义数据模型
type User struct {
    ID       string `json:"f_id" db:"f_id"`
    Name     string `json:"f_name" db:"f_name"`
    Email    string `json:"f_email" db:"f_email"`
    Status   string `json:"f_status" db:"f_status"`
    CreateTime int64 `json:"f_create_time" db:"f_create_time"`
}

// 3. 开始使用
ctx := context.Background()

// 查询用户
user := &User{}
err := orm.Select().From("t_users").WhereEq("f_id", "123").First(ctx, user)

// 插入用户
_, err = orm.Insert().Into("t_users").Values(map[string]interface{}{
    "f_id":    "new-user-id",
    "f_name":  "张三",
    "f_email": "zhangsan@example.com",
    "f_status": "active",
}).Execute(ctx)
```

### 5分钟上手

以下是一个完整的使用示例，展示了ORM Helper的主要功能：

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/kweaver-ai/operator-hub/operator-integration/server/infra/common/ormhelper"
    "github.com/kweaver-ai/operator-hub/operator-integration/server/infra/logger"
)

// 定义数据模型
type MCPConfig struct {
    ID          string `json:"f_id" db:"f_id"`
    Name        string `json:"f_name" db:"f_name"`
    Description string `json:"f_description" db:"f_description"`
    Status      string `json:"f_status" db:"f_status"`
    CreateTime  int64  `json:"f_create_time" db:"f_create_time"`
    UpdateTime  int64  `json:"f_update_time" db:"f_update_time"`
}

func main() {
    // 1. 初始化数据库连接和ORM
    dbPool := initDB() // 你的数据库初始化函数

    // 创建带日志的ORM实例
    logger := logger.NewLogger(logger.LevelInfo)
    logConfig := ormhelper.LogConfig{
        Level:              ormhelper.LogLevelInfo,
        SlowQueryThreshold: 100, // 100毫秒
        LogSlowQuery:       true,
        LogAllQueries:      false,
    }

    orm := ormhelper.NewWithLogger(dbPool, "your_database", logger, logConfig)
    ctx := context.Background()

    // 2. 插入数据
    newConfig := &MCPConfig{
        ID:          "config-001",
        Name:        "示例配置",
        Description: "这是一个示例配置",
        Status:      "active",
        CreateTime:  time.Now().UnixNano(),
        UpdateTime:  time.Now().UnixNano(),
    }

    _, err := orm.Insert().Into("t_mcp_server_config").Values(map[string]interface{}{
        "f_id":          newConfig.ID,
        "f_name":        newConfig.Name,
        "f_description": newConfig.Description,
        "f_status":      newConfig.Status,
        "f_create_time": newConfig.CreateTime,
        "f_update_time": newConfig.UpdateTime,
    }).Execute(ctx)

    if err != nil {
        log.Printf("插入失败: %v", err)
        return
    }
    log.Println("✅ 数据插入成功")

    // 3. 查询单条数据
    config := &MCPConfig{}
    err = orm.Select().
        From("t_mcp_server_config").
        WhereEq("f_id", "config-001").
        First(ctx, config)

    if err != nil {
        log.Printf("查询失败: %v", err)
        return
    }
    log.Printf("✅ 查询成功: %+v", config)

    // 4. 查询多条数据
    configs := []*MCPConfig{}
    err = orm.Select().
        From("t_mcp_server_config").
        WhereEq("f_status", "active").
        OrderByDesc("f_create_time").
        Limit(10).
        Get(ctx, &configs)

    if err != nil {
        log.Printf("查询列表失败: %v", err)
        return
    }
    log.Printf("✅ 查询到 %d 条记录", len(configs))

    // 5. 更新数据
    _, err = orm.Update("t_mcp_server_config").
        Set("f_description", "更新后的描述").
        Set("f_update_time", time.Now().UnixNano()).
        WhereEq("f_id", "config-001").
        Execute(ctx)

    if err != nil {
        log.Printf("更新失败: %v", err)
        return
    }
    log.Println("✅ 数据更新成功")

    // 6. 统计数据
    count, err := orm.Select().
        From("t_mcp_server_config").
        WhereEq("f_status", "active").
        Count(ctx)

    if err != nil {
        log.Printf("统计失败: %v", err)
        return
    }
    log.Printf("✅ 活跃配置数量: %d", count)

    // 7. 事务操作
    tx, err := dbPool.Begin()
    if err != nil {
        log.Printf("开启事务失败: %v", err)
        return
    }
    defer tx.Rollback()

    txORM := orm.WithTx(tx)

    // 在事务中执行多个操作
    _, err = txORM.Update("t_mcp_server_config").
        Set("f_status", "inactive").
        WhereEq("f_id", "config-001").
        Execute(ctx)

    if err != nil {
        log.Printf("事务操作失败: %v", err)
        return
    }

    // 提交事务
    if err = tx.Commit(); err != nil {
        log.Printf("提交事务失败: %v", err)
        return
    }
    log.Println("✅ 事务执行成功")
}
```

运行这个示例，你将看到类似以下的日志输出：

```
✅ 数据插入成功
INF SQL执行 | SQL: SELECT f_id, f_name, f_description, f_status, f_create_time, f_update_time FROM `your_database`.`t_mcp_server_config` WHERE f_id = ? LIMIT 1 | 参数: ['config-001'] | 执行时间: 5ms
✅ 查询成功: &{ID:config-001 Name:示例配置 Description:这是一个示例配置 Status:active CreateTime:1703123456789 UpdateTime:1703123456789}
✅ 查询到 1 条记录
✅ 数据更新成功
✅ 活跃配置数量: 1
✅ 事务执行成功
```

🎉 **恭喜！** 你已经掌握了ORM Helper的基本用法。接下来可以查看详细的功能介绍和高级用法。

## 📖 基础教程

### 查询操作

#### 查询单条记录

```go
// 根据ID查询
user := &User{}
err := orm.Select().From("t_users").WhereEq("f_id", "123").First(ctx, user)

// 查询指定字段
user := &User{}
err := orm.Select("f_id", "f_name", "f_email").
    From("t_users").
    WhereEq("f_id", "123").
    First(ctx, user)

// 处理未找到的情况
if err == sql.ErrNoRows {
    // 记录不存在
    return nil, nil
}
```

#### 查询多条记录

```go
// 查询列表
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").
    OrderByDesc("f_create_time").
    Get(ctx, &users)

// 分页查询
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").
    OrderByDesc("f_create_time").
    Limit(20).
    Offset(40). // 第3页，每页20条
    Get(ctx, &users)

// 统计数量
count, err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").
    Count(ctx)
```

#### 复杂查询条件

```go
// 多个条件
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").
    WhereLike("f_name", "%张%").
    WhereGt("f_create_time", startTime).
    WhereLt("f_create_time", endTime).
    Get(ctx, &users)

// IN 查询
userIDs := []interface{}{"id1", "id2", "id3"}
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereIn("f_id", userIDs...).
    Get(ctx, &users)

// 复杂条件组合
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").
    And(func(w *ormhelper.WhereBuilder) {
        w.Like("f_name", "%admin%").Or(func(w2 *ormhelper.WhereBuilder) {
            w2.Eq("f_role", "admin").Eq("f_department", "IT")
        })
    }).
    Get(ctx, &users)
```

### 插入操作

#### 单条插入

```go
// 使用map插入
data := map[string]interface{}{
    "f_id":          "user-001",
    "f_name":        "张三",
    "f_email":       "zhangsan@example.com",
    "f_status":      "active",
    "f_create_time": time.Now().UnixNano(),
}

result, err := orm.Insert().Into("t_users").Values(data).Execute(ctx)
if err != nil {
    return err
}

// 获取插入的行数
rowsAffected, _ := result.RowsAffected()
```

#### 批量插入

```go
// 批量插入多条记录
columns := []string{"f_id", "f_name", "f_email", "f_status", "f_create_time"}
values := [][]interface{}{
    {"user-001", "张三", "zhangsan@example.com", "active", time.Now().UnixNano()},
    {"user-002", "李四", "lisi@example.com", "active", time.Now().UnixNano()},
    {"user-003", "王五", "wangwu@example.com", "active", time.Now().UnixNano()},
}

_, err := orm.Insert().
    Into("t_users").
    BatchValues(columns, values).
    Execute(ctx)
```

### 更新操作

#### 根据条件更新

```go
// 更新单个字段
_, err := orm.Update("t_users").
    Set("f_status", "inactive").
    WhereEq("f_id", "user-001").
    Execute(ctx)

// 更新多个字段
_, err := orm.Update("t_users").
    Set("f_name", "新名称").
    Set("f_email", "newemail@example.com").
    Set("f_update_time", time.Now().UnixNano()).
    WhereEq("f_id", "user-001").
    Execute(ctx)

// 批量更新
affectedRows, err := orm.Update("t_users").
    Set("f_status", "inactive").
    Set("f_update_time", time.Now().UnixNano()).
    WhereIn("f_id", "user-001", "user-002", "user-003").
    ExecuteAndReturnAffected(ctx)
```

### 删除操作

```go
// 根据ID删除
_, err := orm.Delete().From("t_users").WhereEq("f_id", "user-001").Execute(ctx)

// 批量删除
_, err := orm.Delete().
    From("t_users").
    WhereEq("f_status", "inactive").
    WhereLt("f_create_time", oldTime).
    Execute(ctx)

// 软删除（推荐）
_, err := orm.Update("t_users").
    Set("f_status", "deleted").
    Set("f_delete_time", time.Now().UnixNano()).
    WhereEq("f_id", "user-001").
    Execute(ctx)
```

## 🔧 高级功能

### 事务管理

#### 基础事务使用

```go
// 开启事务
tx, err := dbPool.Begin()
if err != nil {
    return err
}
defer tx.Rollback() // 确保异常时回滚

// 使用事务ORM
txORM := orm.WithTx(tx)

// 在事务中执行操作
_, err = txORM.Insert().Into("t_users").Values(userData).Execute(ctx)
if err != nil {
    return err
}

_, err = txORM.Insert().Into("t_user_profiles").Values(profileData).Execute(ctx)
if err != nil {
    return err
}

// 提交事务
return tx.Commit()
```

#### 事务函数封装

```go
func (dao *UserDAO) CreateUserWithProfile(ctx context.Context, user *User, profile *UserProfile) error {
    return dao.withTransaction(ctx, func(txORM *ormhelper.DB) error {
        // 插入用户
        _, err := txORM.Insert().Into("t_users").Values(map[string]interface{}{
            "f_id":          user.ID,
            "f_name":        user.Name,
            "f_email":       user.Email,
            "f_create_time": time.Now().UnixNano(),
        }).Execute(ctx)
        if err != nil {
            return fmt.Errorf("插入用户失败: %w", err)
        }

        // 插入用户资料
        _, err = txORM.Insert().Into("t_user_profiles").Values(map[string]interface{}{
            "f_user_id":     user.ID,
            "f_avatar":      profile.Avatar,
            "f_bio":         profile.Bio,
            "f_create_time": time.Now().UnixNano(),
        }).Execute(ctx)
        if err != nil {
            return fmt.Errorf("插入用户资料失败: %w", err)
        }

        return nil
    })
}

// 事务辅助函数
func (dao *UserDAO) withTransaction(ctx context.Context, fn func(*ormhelper.DB) error) error {
    tx, err := dao.dbPool.Begin()
    if err != nil {
        return err
    }
    defer tx.Rollback()

    txORM := dao.orm.WithTx(tx)
    if err := fn(txORM); err != nil {
        return err
    }

    return tx.Commit()
}
```

### 日志功能

#### 日志配置详解

```go
// 开发环境配置 - 详细日志
devLogger := logger.NewLogger(logger.LevelDebug)
devLogConfig := ormhelper.LogConfig{
    Level:              ormhelper.LogLevelDebug, // 记录所有日志
    SlowQueryThreshold: 50,                      // 50毫秒慢查询阈值
    LogSlowQuery:       true,                    // 记录慢查询
    LogAllQueries:      true,                    // 记录所有查询
}

// 生产环境配置 - 精简日志
prodLogger := logger.NewLogger(logger.LevelWarn)
prodLogConfig := ormhelper.LogConfig{
    Level:              ormhelper.LogLevelWarn, // 只记录警告和错误
    SlowQueryThreshold: 200,                    // 200毫秒慢查询阈值
    LogSlowQuery:       true,                   // 记录慢查询
    LogAllQueries:      false,                  // 不记录所有查询
}

// 创建ORM实例
orm := ormhelper.NewWithLogger(dbPool, "database_name", prodLogger, prodLogConfig)
```

#### 动态日志控制

```go
// 运行时启用调试日志
func (dao *UserDAO) EnableDebugLogging() {
    debugLogger := logger.NewLogger(logger.LevelDebug)
    debugConfig := ormhelper.LogConfig{
        Level:         ormhelper.LogLevelDebug,
        LogAllQueries: true,
    }
    dao.orm.EnableLogging(debugLogger, debugConfig)
}

// 运行时禁用日志
func (dao *UserDAO) DisableLogging() {
    dao.orm.DisableLogging()
}

// 临时启用日志进行调试
func (dao *UserDAO) DebugQuery(ctx context.Context, userID string) (*User, error) {
    // 临时启用调试日志
    originalEnabled := dao.orm.IsLoggingEnabled()
    if !originalEnabled {
        dao.EnableDebugLogging()
        defer dao.DisableLogging()
    }

    // 执行需要调试的查询
    user := &User{}
    err := dao.orm.Select().From("t_users").WhereEq("f_id", userID).First(ctx, user)

    return user, err
}
```

### 复杂查询

#### JOIN 查询

```go
// 左连接查询
type UserWithProfile struct {
    UserID     string `db:"user_id"`
    UserName   string `db:"user_name"`
    UserEmail  string `db:"user_email"`
    Avatar     string `db:"avatar"`
    Bio        string `db:"bio"`
}

func (dao *UserDAO) GetUsersWithProfiles(ctx context.Context) ([]*UserWithProfile, error) {
    var results []*UserWithProfile

    err := dao.orm.Select(
        "u.f_id AS user_id",
        "u.f_name AS user_name",
        "u.f_email AS user_email",
        "p.f_avatar AS avatar",
        "p.f_bio AS bio",
    ).
        From("t_users u").
        LeftJoin("t_user_profiles p", "u.f_id = p.f_user_id").
        WhereEq("u.f_status", "active").
        OrderByDesc("u.f_create_time").
        Get(ctx, &results)

    return results, err
}
```

#### 聚合查询

```go
// 统计查询
type UserStats struct {
    Status string `db:"status"`
    Count  int64  `db:"count"`
}

func (dao *UserDAO) GetUserStatsByStatus(ctx context.Context) ([]*UserStats, error) {
    var stats []*UserStats

    err := dao.orm.Select("f_status AS status", "COUNT(*) AS count").
        From("t_users").
        GroupBy("f_status").
        Having("COUNT(*) > 0").
        OrderByDesc("count").
        Get(ctx, &stats)

    return stats, err
}

// 复杂聚合
func (dao *UserDAO) GetMonthlyUserStats(ctx context.Context, year int) (map[string]int64, error) {
    type MonthlyStats struct {
        Month string `db:"month"`
        Count int64  `db:"count"`
    }

    var monthlyStats []*MonthlyStats
    err := dao.orm.Select(
        "DATE_FORMAT(FROM_UNIXTIME(f_create_time/1000000000), '%Y-%m') AS month",
        "COUNT(*) AS count",
    ).
        From("t_users").
        WhereGt("f_create_time", time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()).
        WhereLt("f_create_time", time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC).UnixNano()).
        GroupBy("month").
        OrderByAsc("month").
        Get(ctx, &monthlyStats)

    if err != nil {
        return nil, err
    }

    // 转换为map
    result := make(map[string]int64)
    for _, stat := range monthlyStats {
        result[stat.Month] = stat.Count
    }

    return result, nil
}
```

### DAO模式集成

#### 完整的DAO实现

```go
type UserDAO struct {
    orm    *ormhelper.DB
    dbPool *sqlx.DB
    logger interfaces.Logger
}

func NewUserDAO(dbName string, logger interfaces.Logger) *UserDAO {
    dbPool := db.NewDBPool()

    // 配置日志
    logConfig := ormhelper.LogConfig{
        Level:              ormhelper.LogLevelInfo,
        SlowQueryThreshold: 100,
        LogSlowQuery:       true,
        LogAllQueries:      false,
    }

    return &UserDAO{
        orm:    ormhelper.NewWithLogger(dbPool, dbName, logger, logConfig),
        dbPool: dbPool,
        logger: logger,
    }
}

// 查询方法
func (dao *UserDAO) GetByID(ctx context.Context, id string) (*User, error) {
    user := &User{}
    err := dao.orm.Select().From("t_users").WhereEq("f_id", id).First(ctx, user)

    if err == sql.ErrNoRows {
        return nil, nil
    }
    if err != nil {
        dao.logger.Errorf("查询用户失败, id=%s, error=%v", id, err)
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }

    return user, nil
}

// 分页查询
func (dao *UserDAO) GetList(ctx context.Context, filters *UserFilters, page, pageSize int) ([]*User, int64, error) {
    query := dao.orm.Select().From("t_users")

    // 动态构建查询条件
    if filters.Status != "" {
        query = query.WhereEq("f_status", filters.Status)
    }
    if filters.Name != "" {
        query = query.WhereLike("f_name", "%"+filters.Name+"%")
    }
    if filters.Email != "" {
        query = query.WhereLike("f_email", "%"+filters.Email+"%")
    }

    // 查询总数
    total, err := query.Count(ctx)
    if err != nil {
        return nil, 0, fmt.Errorf("查询用户总数失败: %w", err)
    }

    // 分页查询
    var users []*User
    offset := (page - 1) * pageSize
    err = query.
        OrderByDesc("f_create_time").
        Limit(pageSize).
        Offset(offset).
        Get(ctx, &users)

    if err != nil {
        dao.logger.Errorf("分页查询用户失败, page=%d, pageSize=%d, error=%v", page, pageSize, err)
        return nil, 0, fmt.Errorf("分页查询用户失败: %w", err)
    }

    return users, total, nil
}

// 创建用户
func (dao *UserDAO) Create(ctx context.Context, user *User) error {
    now := time.Now().UnixNano()
    user.CreateTime = now
    user.UpdateTime = now

    _, err := dao.orm.Insert().Into("t_users").Values(map[string]interface{}{
        "f_id":          user.ID,
        "f_name":        user.Name,
        "f_email":       user.Email,
        "f_status":      user.Status,
        "f_create_time": user.CreateTime,
        "f_update_time": user.UpdateTime,
    }).Execute(ctx)

    if err != nil {
        dao.logger.Errorf("创建用户失败, user=%+v, error=%v", user, err)
        return fmt.Errorf("创建用户失败: %w", err)
    }

    dao.logger.Infof("创建用户成功, id=%s, name=%s", user.ID, user.Name)
    return nil
}

// 更新用户
func (dao *UserDAO) Update(ctx context.Context, user *User) error {
    user.UpdateTime = time.Now().UnixNano()

    _, err := dao.orm.Update("t_users").
        Set("f_name", user.Name).
        Set("f_email", user.Email).
        Set("f_status", user.Status).
        Set("f_update_time", user.UpdateTime).
        WhereEq("f_id", user.ID).
        Execute(ctx)

    if err != nil {
        dao.logger.Errorf("更新用户失败, user=%+v, error=%v", user, err)
        return fmt.Errorf("更新用户失败: %w", err)
    }

    return nil
}

// 删除用户（软删除）
func (dao *UserDAO) Delete(ctx context.Context, id string) error {
    _, err := dao.orm.Update("t_users").
        Set("f_status", "deleted").
        Set("f_delete_time", time.Now().UnixNano()).
        WhereEq("f_id", id).
        Execute(ctx)

    if err != nil {
        dao.logger.Errorf("删除用户失败, id=%s, error=%v", id, err)
        return fmt.Errorf("删除用户失败: %w", err)
    }

    return nil
}
```

## ⚡ 性能优化

### 查询优化

```go
// ✅ 好的做法：使用索引字段
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereEq("f_status", "active").    // f_status有索引
    WhereEq("f_department", "IT").    // f_department有索引
    OrderByDesc("f_create_time").     // f_create_time有索引
    Limit(20).                        // 限制结果数量
    Get(ctx, &users)

// ❌ 避免：全表扫描
users := []*User{}
err := orm.Select().
    From("t_users").
    WhereLike("f_description", "%keyword%"). // 没有索引的LIKE查询
    Get(ctx, &users) // 没有LIMIT限制
```

### 批量操作优化

```go
// ✅ 批量插入
columns := []string{"f_id", "f_name", "f_email", "f_create_time"}
values := make([][]interface{}, 0, len(users))
for _, user := range users {
    values = append(values, []interface{}{
        user.ID, user.Name, user.Email, time.Now().UnixNano(),
    })
}

_, err := orm.Insert().Into("t_users").BatchValues(columns, values).Execute(ctx)

// ❌ 避免：循环单条插入
for _, user := range users {
    _, err := orm.Insert().Into("t_users").Values(map[string]interface{}{
        "f_id":          user.ID,
        "f_name":        user.Name,
        "f_email":       user.Email,
        "f_create_time": time.Now().UnixNano(),
    }).Execute(ctx)
}
```

### 连接池优化

```go
// 数据库连接池配置
func setupDBPool() *sqlx.DB {
    db := sqlx.MustConnect("mysql", dsn)

    // 设置连接池参数
    db.SetMaxOpenConns(100)        // 最大打开连接数
    db.SetMaxIdleConns(20)         // 最大空闲连接数
    db.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
    db.SetConnMaxIdleTime(time.Minute * 30) // 连接最大空闲时间

    return db
}
```

## 🛠️ 最佳实践

### 错误处理

```go
func (dao *UserDAO) SafeGetUser(ctx context.Context, id string) (*User, error) {
    if id == "" {
        return nil, errors.New("用户ID不能为空")
    }

    user := &User{}
    err := dao.orm.Select().From("t_users").WhereEq("f_id", id).First(ctx, user)

    if err == sql.ErrNoRows {
        return nil, nil // 明确返回nil表示未找到
    }

    if err != nil {
        dao.logger.Errorf("查询用户失败, id=%s, error=%v", id, err)
        return nil, fmt.Errorf("查询用户失败: %w", err)
    }

    return user, nil
}
```

### 参数验证

```go
func (dao *UserDAO) CreateUser(ctx context.Context, user *User) error {
    // 参数验证
    if user == nil {
        return errors.New("用户信息不能为空")
    }
    if user.Name == "" {
        return errors.New("用户名不能为空")
    }
    if user.Email == "" {
        return errors.New("邮箱不能为空")
    }

    // 检查邮箱格式
    if !isValidEmail(user.Email) {
        return errors.New("邮箱格式不正确")
    }

    // 检查用户是否已存在
    existing, err := dao.GetByEmail(ctx, user.Email)
    if err != nil {
        return fmt.Errorf("检查用户是否存在失败: %w", err)
    }
    if existing != nil {
        return errors.New("邮箱已被使用")
    }

    // 执行创建
    return dao.Create(ctx, user)
}
```

### 日志记录

```go
func (dao *UserDAO) UpdateUserWithLog(ctx context.Context, user *User) error {
    start := time.Now()

    // 记录操作开始
    dao.logger.Infof("开始更新用户, id=%s, name=%s", user.ID, user.Name)

    err := dao.Update(ctx, user)
    duration := time.Since(start)

    if err != nil {
        dao.logger.Errorf("更新用户失败, id=%s, duration=%v, error=%v",
            user.ID, duration, err)
        return err
    }

    dao.logger.Infof("更新用户成功, id=%s, duration=%v", user.ID, duration)
    return nil
}
```

## ❓ 常见问题

### Q: 如何处理数据库字段和结构体字段的映射？

A: 使用`db`标签进行映射：

```go
type User struct {
    ID         string `json:"id" db:"f_id"`           // 数据库字段：f_id
    Name       string `json:"name" db:"f_name"`       // 数据库字段：f_name
    Email      string `json:"email" db:"f_email"`     // 数据库字段：f_email
    CreateTime int64  `json:"create_time" db:"f_create_time"`
}
```

### Q: 如何处理NULL值？

A: 使用指针类型或sql.NullXXX类型：

```go
type User struct {
    ID          string         `db:"f_id"`
    Name        string         `db:"f_name"`
    Description *string        `db:"f_description"`    // 可为NULL
    Age         sql.NullInt64  `db:"f_age"`           // 可为NULL
}
```

### Q: 如何进行数据库迁移？

A: 建议使用专门的迁移工具，ORM Helper专注于数据操作：

```go
// 使用golang-migrate或其他迁移工具
// ORM Helper不提供DDL操作，专注于DML操作
```

### Q: 如何处理复杂的业务逻辑？

A: 在Service层处理业务逻辑，DAO层只负责数据访问：

```go
// Service层
type UserService struct {
    userDAO    *UserDAO
    profileDAO *UserProfileDAO
}

func (s *UserService) CreateUserWithProfile(ctx context.Context, req *CreateUserRequest) error {
    // 业务逻辑验证
    if err := s.validateCreateRequest(req); err != nil {
        return err
    }

    // 使用事务创建用户和资料
    return s.userDAO.withTransaction(ctx, func(txORM *ormhelper.DB) error {
        // 创建用户
        user := &User{
            ID:    generateUserID(),
            Name:  req.Name,
            Email: req.Email,
        }
        if err := s.userDAO.CreateWithTx(ctx, txORM, user); err != nil {
            return err
        }

        // 创建用户资料
        profile := &UserProfile{
            UserID: user.ID,
            Avatar: req.Avatar,
            Bio:    req.Bio,
        }
        return s.profileDAO.CreateWithTx(ctx, txORM, profile)
    })
}
```

### Q: 如何进行单元测试？

A: 使用测试数据库或Mock：

```go
func TestUserDAO_Create(t *testing.T) {
    // 设置测试数据库
    testDB := setupTestDB()
    defer testDB.Close()

    logger := logger.NewLogger(logger.LevelDebug)
    dao := NewUserDAO("test_db", logger)

    ctx := context.Background()
    user := &User{
        ID:    "test-user-001",
        Name:  "测试用户",
        Email: "test@example.com",
        Status: "active",
    }

    // 执行测试
    err := dao.Create(ctx, user)
    assert.NoError(t, err)

    // 验证结果
    created, err := dao.GetByID(ctx, user.ID)
    assert.NoError(t, err)
    assert.Equal(t, user.Name, created.Name)
    assert.Equal(t, user.Email, created.Email)
}
```

## 🏗️ 架构设计

ORM Helper采用模块化设计，详细的架构说明请参考 [ARCHITECTURE.md](./ARCHITECTURE.md)。

### 核心组件

- **ORM核心** (`orm.go`) - 主要的ORM类，提供统一的数据库操作接口
- **查询构建器** (`select.go`) - 负责构建SELECT查询语句
- **插入构建器** (`insert.go`) - 负责构建INSERT语句
- **更新构建器** (`update.go`) - 负责构建UPDATE语句
- **删除构建器** (`delete.go`) - 负责构建DELETE语句
- **条件构建器** (`where.go`) - 负责构建WHERE条件
- **结果扫描器** (`scanner.go`) - 负责将查询结果映射到结构体
- **日志系统** (`logger.go`) - 提供SQL执行日志和性能监控

### 设计原则

1. **简单易用** - API设计直观，学习成本低
2. **完全兼容** - 与现有`sqlx.DB`无缝集成
3. **非侵入式** - 可渐进式迁移，不影响现有代码
4. **类型安全** - 编译时检查，减少运行时错误
5. **高性能** - 最小化性能开销，接近原生SQL性能

---

🎉 **开始使用ORM Helper，让数据库操作变得简单、安全、高效！**

如果你在使用过程中遇到任何问题，欢迎提出Issue或贡献代码。

# ORM Helper - 统一分页和排序功能

`ormhelper` 包现在提供了统一的分页和排序功能，使得在业务层可以轻松实现一致的数据查询接口。

## 核心特性

### 1. 分页参数 (PaginationParams)

```go
type PaginationParams struct {
    Page     int `json:"page" validate:"min=1"`              // 页码，从1开始
    PageSize int `json:"page_size" validate:"min=1,max=100"` // 每页数量
}
```

### 2. 排序参数 (SortParams)

```go
type SortOrder string

const (
    SortOrderAsc  SortOrder = "asc"  // 升序
    SortOrderDesc SortOrder = "desc" // 降序
)

type SortField struct {
    Field string    `json:"field"` // 数据库字段名
    Order SortOrder `json:"order"` // 排序方向
}

type SortParams struct {
    Fields []SortField `json:"fields,omitempty"` // 支持多字段排序
}
```

### 3. 查询结果 (QueryResult)

```go
type QueryResult struct {
    Total      int64 `json:"total"`       // 总记录数
    Page       int   `json:"page"`        // 当前页码
    PageSize   int   `json:"page_size"`   // 每页数量
    TotalPages int   `json:"total_pages"` // 总页数
    HasNext    bool  `json:"has_next"`    // 是否有下一页
    HasPrev    bool  `json:"has_prev"`    // 是否有上一页
}
```

## SelectBuilder 新增方法

### Pagination()

应用分页参数到查询中：

```go
func (s *SelectBuilder) Pagination(pagination *PaginationParams) *SelectBuilder
```

### Sort()

应用排序参数到查询中：

```go
func (s *SelectBuilder) Sort(sort *SortParams) *SelectBuilder
```

## 使用示例

### 基本分页查询

```go
paginationParams := &PaginationParams{
    Page:     1,
    PageSize: 10,
}

var results []ExampleData
err := orm.Select().
    From("t_example").
    Pagination(paginationParams).
    Get(ctx, &results)
```

### 带排序的分页查询

```go
sortParams := &SortParams{
    Fields: []SortField{
        {Field: "name", Order: SortOrderAsc},
        {Field: "id", Order: SortOrderDesc},
    },
}

paginationParams := &PaginationParams{
    Page:     1,
    PageSize: 10,
}

err := orm.Select().
    From("t_example").
    Pagination(paginationParams).
    Sort(sortParams).
    Get(ctx, &results)
```

### 获取分页信息

```go
// 获取总数
totalCount, err := orm.Select().
    From("t_example").
    Count(ctx)

// 计算分页结果
queryResult := CalculateQueryResult(totalCount, paginationParams)

// queryResult 包含完整的分页信息
fmt.Printf("总记录数: %d\n", queryResult.Total)
fmt.Printf("当前页: %d/%d\n", queryResult.Page, queryResult.TotalPages)
fmt.Printf("是否有下一页: %v\n", queryResult.HasNext)
```

### 复杂查询示例

```go
sortParams := &SortParams{
    Fields: []SortField{
        {Field: "priority", Order: SortOrderDesc},
        {Field: "create_time", Order: SortOrderAsc},
    },
}

paginationParams := &PaginationParams{
    Page:     2,
    PageSize: 20,
}

err := orm.Select().
    From("t_example").
    WhereEq("status", "active").         // 条件过滤
    WhereLike("name", "%test%").         // 模糊查询
    Sort(sortParams).                    // 多字段排序
    Pagination(paginationParams).        // 分页
    Get(ctx, &results)
```

## 工具函数

### CalculateQueryResult()

计算查询结果的分页信息：

```go
func CalculateQueryResult(total int64, pagination *PaginationParams) *QueryResult
```

这个函数会自动计算：
- 总页数
- 是否有下一页/上一页
- 处理边界情况（如分页参数为空或无效）

## 注意事项

1. **字段名安全**: `SortField.Field` 中的字段名应该由调用方确保安全，避免SQL注入
2. **参数验证**: 建议在业务层对分页和排序参数进行验证
3. **性能考虑**: 大分页查询时注意性能影响
4. **索引**: 确保排序字段有适当的数据库索引

## 集成指南

这些功能设计为不影响现有业务逻辑，您可以：

1. 在现有的数据访问层方法中逐步添加分页和排序支持
2. 保持现有接口不变，通过可选参数方式添加新功能
3. 利用统一的 `QueryResult` 结构在API层返回一致的分页信息

## 示例文件

参考 `usage_example.go` 文件了解完整的使用示例。