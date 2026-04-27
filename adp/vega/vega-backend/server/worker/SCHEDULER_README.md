# 定时发现Catalog资源功能说明

## 概述

定时发现功能允许用户按照预定义的时间表自动执行catalog资源的发现操作。该功能使用cron表达式来定义任务的执行时间，并通过worker中的调度器来管理这些定时任务。

## 架构

### 核心组件

1. **ScheduledDiscoverTask** - 定时发现任务的数据模型
   - 存储在数据库中
   - 包含catalog_id、cron表达式、开始/结束时间等信息
   - 支持启用/禁用状态

2. **ScheduledDiscoverTaskService** - 定时任务的业务逻辑层
   - 管理定时任务的CRUD操作
   - 负责执行定时发现任务
   - 更新任务的执行状态

3. **Scheduler** - 调度器
   - 使用robfig/cron库实现cron调度
   - 管理所有启用的定时任务
   - 负责任务的加载、调度和执行

## API接口

### 创建定时发现任务

**外部API**
```
POST /api/vega-backend/v1/catalogs/:id/scheduled-discover
```

**内部API**
```
POST /api/vega-backend/in/v1/catalogs/:id/scheduled-discover
```

**请求体**
```json
{
  "catalog_id": "catalog-123",
  "cron_expr": "0 0 * * *",
  "start_time": 1234567890000,  // 可选，开始时间戳（毫秒）
  "end_time": 1234567890000     // 可选，结束时间戳（毫秒）
}
```

**响应**
```json
{
  "task_id": "task-abc123",
  "message": "Scheduled discover task created successfully"
}
```

## Cron表达式格式

使用标准的cron表达式格式（支持秒级精度）：

```
秒 分 时 日 月 周
*  *  *  *  *
```

示例：
- `0 0 * * *` - 每天午夜执行
- `0 */6 * * *` - 每6小时执行一次
- `0 0 * * 0` - 每周日午夜执行
- `0 0 0 1 *` - 每月1号午夜执行

## 使用示例

### 1. 创建一个每天午夜执行的定时任务

```bash
curl -X POST http://localhost:8080/api/vega-backend/v1/catalogs/catalog-123/scheduled-discover   -H "Content-Type: application/json"   -H "Authorization: Bearer YOUR_TOKEN"   -d '{
    "catalog_id": "catalog-123",
    "cron_expr": "0 0 * * *"
  }'
```

### 2. 创建一个每小时执行一次的定时任务

```bash
curl -X POST http://localhost:8080/api/vega-backend/v1/catalogs/catalog-123/scheduled-discover   -H "Content-Type: application/json"   -H "Authorization: Bearer YOUR_TOKEN"   -d '{
    "catalog_id": "catalog-123",
    "cron_expr": "0 * * * *"
  }'
```

### 3. 创建一个带时间限制的定时任务

```bash
curl -X POST http://localhost:8080/api/vega-backend/v1/catalogs/catalog-123/scheduled-discover   -H "Content-Type: application/json"   -H "Authorization: Bearer YOUR_TOKEN"   -d '{
    "catalog_id": "catalog-123",
    "cron_expr": "0 */6 * * *",
    "start_time": 1234567890000,
    "end_time": 1234567890000
  }'
```

## 调度器管理

### 启动调度器

在应用启动时初始化并启动调度器：

```go
import (
    "vega-backend/common"
    "vega-backend/logics/scheduled_discover_task"
    "vega-backend/worker"
)

func main() {
    appSetting := common.LoadAppSetting()

    // 初始化服务
    dts := discover_task.NewDiscoverTaskService(appSetting)
    sdtService := scheduled_discover_task.NewScheduledDiscoverTaskService(appSetting, dts)

    // 初始化并启动调度器
    scheduler := worker.NewScheduler(appSetting, sdtService)
    if err := scheduler.Start(); err != nil {
        log.Fatalf("Failed to start scheduler: %v", err)
    }

    // 应用其他逻辑...
}
```

### 重载调度器

当需要重新加载所有定时任务时：

```go
if err := scheduler.Reload(); err != nil {
    log.Errorf("Failed to reload scheduler: %v", err)
}
```

### 停止调度器

在应用关闭时停止调度器：

```go
scheduler.Stop()
```

## 注意事项

1. **任务过期**：当定时任务到达结束时间后，会自动被禁用并从调度器中移除
2. **任务状态**：可以通过查询任务状态来查看任务的执行历史和下次执行时间
3. **并发执行**：调度器支持多个任务并发执行，但每个catalog同一时间只能有一个发现任务在执行
4. **错误处理**：任务执行失败会记录日志，但不会影响其他任务的执行

## 数据库表结构

定时发现任务需要以下数据库表：

```sql
CREATE TABLE scheduled_discover_tasks (
    id VARCHAR(64) PRIMARY KEY,
    catalog_id VARCHAR(64) NOT NULL,
    cron_expr VARCHAR(100) NOT NULL,
    start_time BIGINT,
    end_time BIGINT,
    enabled BOOLEAN DEFAULT TRUE,
    last_run BIGINT,
    next_run BIGINT,
    creator_id VARCHAR(64),
    creator_type VARCHAR(32),
    create_time BIGINT,
    updater_id VARCHAR(64),
    updater_type VARCHAR(32),
    update_time BIGINT,
    INDEX idx_catalog_id (catalog_id),
    INDEX idx_enabled (enabled),
    INDEX idx_next_run (next_run)
);
```

## 待实现功能

以下功能已预留接口，但尚未完全实现：

1. 定时任务的查询和列表接口
2. 定时任务的更新和删除接口
3. 定时任务的启用/禁用接口
4. 定时任务执行历史记录
5. 任务执行失败通知
6. 更精细的cron表达式验证和提示

## 依赖

- github.com/robfig/cron/v3 - Cron调度库
- asynq - 异步任务队列
