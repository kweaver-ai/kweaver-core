# Catalog 健康检查机制设计文档

## 重要说明

**健康检查功能的适用范围**：
- 之前的 catalog 健康检查开关默认为 true，但功能在此需求中实现
- 因此，只有在此需求开发完成后创建和修改的 catalog，才会存在健康检查功能
- 对于在此需求开发之前已存在的 catalog，需要通过更新接口启用健康检查功能

## 1. 背景与问题

当前系统中的 catalog 健康检查机制是空壳实现，catalog 永远显示为 healthy 状态，缺乏真实的连接检测能力。这导致以下问题：

1. 无法及时发现数据源连接故障
2. 查询请求发送到不可用的数据源后，缺乏自动恢复机制
3. 查询成功后无法自动更新不正常的健康状态
4. 系统整体可靠性降低

## 2. 设计目标

1. **真实连接检测**：`/test-connection` 端点应真实调用 connector 的 `Ping()` 或 `TestConnection()` 方法进行连接检测
2. **定期健康检查**：对启用了 `f_health_check_enabled` 的 catalog 实现后台定期健康检查
3. **查询与健康状态联动**：查询操作不受健康状态限制，查询成功后自动将不正常状态更新为正常状态

## 3. 核心设计

### 3.1 真实连接检测

#### 3.1.1 API 端点

**端点**: `POST /api/v1/catalogs/{id}/test-connection`

**说明**: 该接口为现有接口，无需修改传参和返回参数，仅需内部改造实现逻辑。

**返回体契约**（与现有 `CatalogHealthCheckStatus` 一致）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `health_check_status` | string | healthy, degraded, unhealthy, offline, disabled |
| `last_check_time` | int64 | 最后检查时间戳 |
| `health_check_result` | string | 检查结果描述 |

#### 3.1.2 行为约定（非实现细节）

- **Connector 侧**：各连接器通过既有能力完成连通性验证（如 `TestConnection(ctx)`）；具体错误语义由各 connector 实现决定。
- **Catalog Service 侧**：根据 catalog 创建 connector → 调用连通性检测 → 依据耗时与错误类型映射到上述状态枚举 → **将结果持久化**。
- **改造位置**（实现时参考）：`server/logics/catalog/catalog_service.go` 中的 `TestConnection`。

**状态映射原则**（实现时可配置阈值）：

- 成功：按延迟区分 healthy / degraded（超阈值可视为 degraded）
- 失败：区分一般不可用（unhealthy）与超时/网络不可达（offline）等，规则与全局及 catalog 级阈值对齐

### 3.2 定期健康检查机制

#### 3.2.1 配置参数

**存储位置**：catalog 的 `f_health_check_config`（JSON），与 `connector_config` 平齐。

**JSON 形态**（字段语义定义，默认值与校验由实现与全局配置兜底）：

```json
{
  "strategy": {
    "interval": "30s",
    "timeout": "5s"
  },
  "thresholds": {
    "latency_warning_ms": 500,
    "latency_critical_ms": 2000,
    "failure_threshold": 3,
    "recovery_threshold": 2
  },
  "retry": {
    "max_attempts": 3,
    "backoff": "exponential",
    "initial_delay": "1s",
    "max_delay": "30s"
  }
}
```

**配置层级说明**:

1. **strategy**：`interval`（检查周期）、`timeout`（单次检查超时，超时倾向判为 offline 等，与实现一致即可）
2. **thresholds**：延迟分级、连续失败/成功次数与状态翻转相关阈值
3. **retry**：单次检查失败时的重试次数与退避策略（`fixed` / `exponential`）

**探活方式**：统一使用各连接器的 `TestConnection`；无需在配置中再指定“探活类型”。

**配置优先级**：Catalog 级配置优先；缺项使用全局（如 `application.yaml`）默认值。

#### 3.2.2 组件职责与接口边界（不规定具体类名/存储实现）

| 组件 | 职责 |
|------|------|
| 调度器 | 按启用项与 `interval` 触发检查；控制与 worker pool 的协作 |
| 任务队列 | 传递待检查的 catalog 标识及解析后的配置 |
| Worker / 执行层 | 调用 connector `TestConnection`、应用 retry、记录耗时 |
| 状态聚合 | 维护连续失败/成功计数，与 thresholds 结合产出最终状态 |
| 持久化 | 将状态与计数字段写回 catalog 存储 |

**内部数据结构**：由实现选型（内存缓存、仅 DB 等），本设计只要求语义等价：**能表达当前状态、最后检查/成功时间、连续计数、最近延迟**。

#### 3.2.3 检查流程

1. 调度器定期纳入启用了 `f_health_check_enabled` 的 catalog
2. 按各 catalog 的 `strategy.interval` 决定何时入队
3. 执行层：`TestConnection` + retry + 记录耗时
4. 按阈值将延迟映射为 healthy / degraded / unhealthy，失败时区分 unhealthy / offline 等
5. 更新连续失败/成功计数，并按 `failure_threshold` / `recovery_threshold` 决定是否保持或恢复状态
6. 持久化

### 3.3 查询与健康状态联动机制

#### 3.3.1 设计原则

1. **查询不受健康状态限制**：任意当前状态均允许查询
2. **查询成功可恢复状态**：若当前为 unhealthy/offline（以及实现约定内的其他“需恢复”状态），成功后可异步更新为 healthy（及清零相关计数等，与定期检查结果一致）
3. **查询失败不反向写健康状态**：仍由定期健康检查主导故障发现

#### 3.3.2 接入点约定

- 在 **catalog 查询入口**（具体路径由现有分层决定）在查询完成后根据成功/失败执行上述策略；状态写库宜异步，避免阻塞查询路径。

#### 3.3.3 状态更新策略

1. **查询成功**：满足条件时更新为 healthy，刷新成功相关时间与计数；失败仅记日志
2. **查询失败**：不更新健康状态

## 4. 数据模型变更

### 4.1 Catalog 表字段说明

Catalog 表已包含健康检查相关字段，无需新增字段，仅需在流程中正确读写：

| 字段 | 说明 |
|------|------|
| `f_health_check_enabled` | 是否启用（历史 catalog 需通过更新接口开启） |
| `f_health_check_config` | 健康检查 JSON 配置 |
| `f_health_check_status` | healthy / degraded / unhealthy / offline / disabled |
| `f_last_check_time` | 最后检查时间戳 |
| `f_last_success_time` | 最后成功时间戳 |
| `f_failure_count` | 连续失败次数 |
| `f_success_count` | 连续成功次数 |
| `f_last_latency_ms` | 最近延迟（毫秒） |
| `f_health_check_result` | 描述文案 |

**状态与计数更新语义**：与第 3.2、3.3 节一致；`disabled` 与 `f_health_check_enabled=false` 对齐。

## 5. API 变更

### 5.1 POST /api/v1/catalogs - 创建 Catalog

- 请求体支持 `health_check_config`
- 创建时初始化健康相关字段（启用标志、初始状态、计数与时间戳、初始 `health_check_result` 等由产品/实现统一约定）

**请求体中与契约相关的片段示例**：

```json
{
  "health_check_enabled": true,
  "health_check_config": {
    "strategy": { "interval": "30s", "timeout": "5s" },
    "thresholds": {
      "latency_warning_ms": 500,
      "latency_critical_ms": 2000,
      "failure_threshold": 3,
      "recovery_threshold": 2
    },
    "retry": {
      "max_attempts": 3,
      "backoff": "exponential",
      "initial_delay": "1s",
      "max_delay": "30s"
    }
  }
}
```

### 5.2 PUT /api/v1/catalogs/{id} - 更新 Catalog

- 可更新 `health_check_enabled`、`health_check_config`
- 更新配置不强制重置健康状态；`health_check_enabled` 置为 false 时状态为 disabled
- 历史 catalog 可通过 `health_check_enabled=true` 启用

### 5.3 GET /api/v1/catalogs/{id} - 获取 Catalog 详情

- 返回体包含完整健康检查状态与 `health_check_config`（字段名以实际 API 为准）

### 5.4 GET /api/v1/catalogs - 列表

- 查询参数可选：`health_check_status`（过滤 healthy / degraded / unhealthy / offline / disabled）
- 列表项携带与健康相关的摘要字段

## 6. 监控与告警

### 6.1 监控指标（命名供实现参考）

- 检查次数、成功/失败次数、延迟分布或直方图等

### 6.2 告警规则（策略级）

- 连续失败、延迟超阈、不健康实例过多等，阈值与告警通道由运维策略定义

## 7. 附录

### 7.1 全局配置（application.yaml）

全局项为所有 catalog 提供默认值；键名与类型与现有工程约定一致，例如：

- `default_interval`、`default_timeout`、`default_max_workers`
- 默认 thresholds、默认 retry 参数

（具体 YAML 键值以实现为准，本设计只约定「存在全局兜底」这一层次。）

### 7.2 术语表

| 术语 | 定义 |
|------|------|
| Catalog | 数据源目录，包含数据源连接信息和元数据 |
| Connector | 连接器，用于与特定类型的数据源进行交互 |
| Health Check | 健康检查，包括手动测试与定期检查 |
| Health Status | healthy、degraded、unhealthy、offline、disabled |
| Strategy / Thresholds / Retry | 检查周期与超时、判定阈值、重试策略 |
| 状态联动 | 查询不受阻；成功后可将异常状态恢复为正常 |
