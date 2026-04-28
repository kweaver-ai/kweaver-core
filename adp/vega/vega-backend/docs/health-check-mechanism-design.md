# Catalog 健康检查机制设计文档

## 重要说明

**健康检查功能的适用范围**：
- 之前的 catalog 健康检查开关默认为 true，但功能在此需求中实现
- 因此，只有在此需求开发完成后创建和修改的 catalog，才会存在健康检查功能
- 对于在此需求开发之前已存在的 catalog，需要通过更新接口启用健康检查功能

## 1. 背景与问题

当前系统中的 catalog 健康检查机制是空壳实现，catalog 永远显示为 healthy 状态，缺乏真实的连接检测能力。这导致以下问题：

1. 无法及时发现数据源连接故障
2. 查询请求可能被发送到不可用的数据源，造成资源浪费和延迟
3. 缺乏自动恢复机制，需要人工干预处理故障
4. 系统整体可靠性降低

## 2. 设计目标

1. **真实连接检测**：`/test-connection` 端点应真实调用 connector 的 `Ping()` 或 `TestConnection()` 方法进行连接检测
2. **定期健康检查**：对启用了 `f_health_check_enabled` 的 catalog 实现后台定期健康检查
3. **查询前检查**：在执行查询前检查 catalog 健康状态，对不健康的 catalog 做 fail-fast

## 3. 核心设计

### 3.1 真实连接检测

#### 3.1.1 API 端点

**端点**: `POST /api/v1/catalogs/{id}/test-connection`

**说明**: 该接口为现有接口，无需修改传参和返回参数，仅需内部改造实现逻辑。

**现有返回结构** (`CatalogHealthCheckStatus`):
```go
type CatalogHealthCheckStatus struct {
    HealthCheckStatus string `json:"health_check_status"`  // 健康状态: healthy, degraded, unhealthy, offline, disabled
    LastCheckTime     int64  `json:"last_check_time"`       // 最后检查时间戳
    HealthCheckResult string `json:"health_check_result"`   // 检查结果描述
}
```

#### 3.1.2 实现逻辑改造

**当前实现问题**:
- `catalog_service.go` 中的 `TestConnection` 方法直接返回 catalog 中存储的 `CatalogHealthCheckStatus`，未进行真实连接检测

**改造方案**:
1. 根据 catalog 配置创建对应的 connector 实例
2. 调用 connector 的 `TestConnection(ctx)` 方法进行真实连接检测
3. 记录检测耗时和结果
4. 更新 catalog 的健康状态 (`HealthCheckStatus`)
5. 更新最后检查时间 (`LastCheckTime`)
6. 更新检查结果描述 (`HealthCheckResult`)
7. **将测试结果持久化到数据库**

**改造位置**:
- `server/logics/catalog/catalog_service.go` 中的 `TestConnection` 方法

**关键代码示例**:
```go
func (cs *catalogService) TestConnection(ctx context.Context, catalog *interfaces.Catalog) (*interfaces.CatalogHealthCheckStatus, error) {
    ctx, span := ar_trace.Tracer.Start(ctx, "Test catalog connection")
    defer span.End()

    if catalog == nil {
        span.SetStatus(codes.Error, "Catalog not found")
        return nil, rest.NewHTTPError(ctx, http.StatusNotFound, verrors.VegaBackend_Catalog_NotFound)
    }

    startTime := time.Now()
    
    // 创建 connector 实例
    connector, err := cs.connectorFactory.CreateConnector(catalog.ConnectorType, catalog.ConnectorCfg)
    if err != nil {
        span.SetStatus(codes.Error, "Failed to create connector")
        return &interfaces.CatalogHealthCheckStatus{
            HealthCheckStatus: interfaces.CatalogHealthStatusUnhealthy,
            LastCheckTime:     time.Now().Unix(),
            HealthCheckResult: fmt.Sprintf("Failed to create connector: %v", err),
        }, nil
    }
    
    // 执行真实连接检测
    err = connector.TestConnection(ctx)
    latency := time.Since(startTime).Milliseconds()
    
    status := interfaces.CatalogHealthCheckStatus{
        LastCheckTime: time.Now().Unix(),
    }

    // 定义延迟阈值（可配置）
    degradedThreshold := int64(1000) // 1秒
    
    if err != nil {
        // 判断错误类型以确定状态
        if isTimeoutError(err) || isNetworkUnreachableError(err) {
            status.HealthCheckStatus = interfaces.CatalogHealthStatusOffline
            status.HealthCheckResult = fmt.Sprintf("Connection timeout or network unreachable (latency: %dms): %v", latency, err)
        } else {
            status.HealthCheckStatus = interfaces.CatalogHealthStatusUnhealthy
            status.HealthCheckResult = fmt.Sprintf("Connection failed (latency: %dms): %v", latency, err)
        }
        span.SetStatus(codes.Error, status.HealthCheckResult)
    } else {
        // 根据延迟判断健康状态
        if latency > degradedThreshold {
            status.HealthCheckStatus = interfaces.CatalogHealthStatusDegraded
            status.HealthCheckResult = fmt.Sprintf("Connection successful but degraded (latency: %dms, threshold: %dms)", latency, degradedThreshold)
            span.SetStatus(codes.Warning, "Connection degraded due to high latency")
        } else {
            status.HealthCheckStatus = interfaces.CatalogHealthStatusHealthy
            status.HealthCheckResult = fmt.Sprintf("Connection successful (latency: %dms)", latency)
            span.SetStatus(codes.Ok, "Connection successful")
        }
    }
    
    // 将测试结果持久化到数据库
    err = cs.repo.UpdateCatalogHealthStatus(ctx, catalog.ID, status)
    if err != nil {
        // 记录错误但不影响返回结果，因为连接测试已经完成
        logger.Errorf("Failed to update catalog health status in database: %v", err)
    }
    
    return &status, nil
}
```

**辅助函数**:
```go
// isTimeoutError 判断错误是否为超时错误
func isTimeoutError(err error) bool {
    if err == nil {
        return false
    }
    // 检查是否为 context.DeadlineExceeded
    if errors.Is(err, context.DeadlineExceeded) {
        return true
    }
    // 检查是否为 net.Error 且超时
    if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
        return true
    }
    // 可以根据实际 connector 返回的错误类型添加更多判断
    return false
}

// isNetworkUnreachableError 判断错误是否为网络不可达错误
func isNetworkUnreachableError(err error) bool {
    if err == nil {
        return false
    }
    // 检查常见的网络不可达错误
    errMsg := strings.ToLower(err.Error())
    return strings.Contains(errMsg, "network unreachable") ||
           strings.Contains(errMsg, "connection refused") ||
           strings.Contains(errMsg, "no route to host")
}
```

### 3.2 定期健康检查机制

#### 3.2.1 配置参数

**健康检查配置结构** (存储在 catalog 的 `f_health_check_config` 字段中，与 `connector_config` 平齐):

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

1. **strategy (检查策略)**
   - `interval` (duration): 定期检查的间隔时间，格式如 "30s"、"1m"、"5m"
     - 默认值: "30s"
     - 最小值: "10s"
     - 说明: 调度器每隔此时间间隔执行一次健康检查
   - `timeout` (duration): 单次检查的超时时间，格式如 "5s"、"10s"
     - 默认值: "5s"
     - 最小值: "1s"
     - 说明: 单次连接检测的超时时间，超时后标记为 "offline"

3. **thresholds (状态判定阈值)**
   - `latency_warning_ms` (integer): 延迟警告阈值（毫秒）
     - 默认值: 500
     - 说明: 连接成功但延迟超过此值时，状态标记为 "degraded"
   - `latency_critical_ms` (integer): 延迟严重阈值（毫秒）
     - 默认值: 2000
     - 说明: 连接成功但延迟超过此值时，状态标记为 "unhealthy"
   - `failure_threshold` (integer): 连续失败阈值
     - 默认值: 3
     - 最小值: 1
     - 说明: 连续失败次数达到此值时，状态保持为 "unhealthy"
   - `recovery_threshold` (integer): 恢复阈值
     - 默认值: 2
     - 最小值: 1
     - 说明: 连续成功次数达到此值时，状态恢复为 "healthy" 或 "degraded"（根据延迟）

4. **retry (重试策略)**
   - `max_attempts` (integer): 单次检查失败时的最大重试次数
     - 默认值: 3
     - 最小值: 0
     - 最大值: 10
     - 说明: 连接失败时的重试次数，0 表示不重试
   - `backoff` (string): 重试退避策略
     - 可选值: "fixed"（固定延迟）、"exponential"（指数退避）
     - 默认值: "exponential"
     - 说明: 
       - fixed: 每次重试使用相同的延迟时间
       - exponential: 每次重试的延迟时间按指数增长
   - `initial_delay` (duration): 重试初始延迟时间
     - 默认值: "1s"
     - 最小值: "100ms"
     - 说明: 第一次重试的延迟时间
   - `max_delay` (duration): 重试最大延迟时间
     - 默认值: "30s"
     - 最小值: "1s"
     - 说明: 重试延迟的最大值（仅在 exponential 模式下生效）

**探活方式**: 健康检查直接使用各个连接器的 `TestConnection` 方法进行连接测试，无需额外配置探活方式。各连接器根据自身特性实现具体的连接测试逻辑（如 AnyShare 连接器会验证配置的路径和文档库类型）。

**配置优先级**: Catalog 级别的配置优先于全局配置（application.yaml），当 Catalog 未配置某项参数时，使用全局配置的默认值。

#### 3.2.2 实现方式

1. **健康检查调度器**：创建后台 goroutine 作为调度器
2. **检查任务队列**：使用 channel 管理待检查的 catalog
3. **并发控制**：使用 worker pool 限制并发检查数量
4. **状态管理**：维护每个 catalog 的健康状态和连续失败/成功计数

```go
// HealthCheckScheduler 健康检查调度器
type HealthCheckScheduler struct {
    workerPool    chan struct{}           // 并发控制
    checkQueue    chan *CatalogCheckTask  // 检查任务队列
    statusStore   *HealthStatusStore      // 状态存储
    appSetting    *common.AppSetting      // 应用配置
    catalogAccess interfaces.CatalogAccess // 数据访问
}

// CatalogCheckTask catalog检查任务
type CatalogCheckTask struct {
    CatalogID   string
    Catalog     *interfaces.Catalog
    CheckConfig *HealthCheckConfig
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
    Strategy   HealthCheckStrategy
    Thresholds HealthCheckThresholds
    Retry      HealthCheckRetry
}

// HealthCheckStrategy 检查策略
type HealthCheckStrategy struct {
    Interval time.Duration
    Timeout  time.Duration
}

// HealthCheckThresholds 状态判定阈值
type HealthCheckThresholds struct {
    LatencyWarningMs  int64
    LatencyCriticalMs int64
    FailureThreshold  int
    RecoveryThreshold  int
}

// HealthCheckRetry 重试策略
type HealthCheckRetry struct {
    MaxAttempts  int
    Backoff      string // "fixed" | "exponential"
    InitialDelay time.Duration
    MaxDelay     time.Duration
}

// HealthStatus 健康状态
type HealthStatus struct {
    CatalogID        string
    Status           string // healthy | degraded | unhealthy | offline | disabled
    LastCheckTime    time.Time
    LastSuccessTime  time.Time
    FailureCount     int
    SuccessCount     int
    LastLatencyMs    int64
}

// HealthStatusStore 健康状态存储
type HealthStatusStore struct {
    sync.RWMutex
    statuses map[string]*HealthStatus
}
```

#### 3.2.3 检查流程

1. 调度器定期扫描启用了 `f_health_check_enabled` 的 catalog
2. 根据每个 catalog 的 `health_check.strategy.interval` 配置，将需要检查的 catalog 放入检查队列
3. Worker 从队列取出 catalog 执行连接检测：
   - 调用连接器的 `TestConnection` 方法进行连接测试
   - 应用 `health_check.retry` 配置进行重试
   - 记录检测耗时
4. 根据检测结果和阈值更新健康状态：
   - 延迟 > `latency_critical_ms`: 标记为 unhealthy
   - 延迟 > `latency_warning_ms`: 标记为 degraded
   - 延迟 <= `latency_warning_ms`: 标记为 healthy
   - 连接失败: 根据错误类型标记为 unhealthy 或 offline
5. 维护连续失败/成功计数：
   - 检查成功: SuccessCount++, FailureCount=0
   - 检查失败: FailureCount++, SuccessCount=0
6. 应用阈值判定：
   - FailureCount >= `failure_threshold`: 状态保持为 unhealthy
   - SuccessCount >= `recovery_threshold`: 状态恢复为 healthy
7. 将健康状态持久化到数据库

### 3.3 查询前健康检查

#### 3.3.1 实现位置

在 catalog service 的查询入口处添加健康检查逻辑：

```go
func (s *CatalogService) Query(catalogID string, query string) (*QueryResult, error) {
    // 1. 检查 catalog 健康状态
    status, err := s.statusStore.Get(catalogID)
    if err != nil {
        return nil, fmt.Errorf("failed to get catalog health status: %v", err)
    }
    
    // 2. Fail-fast: 不健康的 catalog 直接返回错误
    if status.Status == interfaces.CatalogHealthStatusUnhealthy ||
       status.Status == interfaces.CatalogHealthStatusOffline {
        return nil, fmt.Errorf("catalog %s is not healthy (status: %s)", catalogID, status.Status)
    }
    
    // 3. degraded 状态记录警告日志
    if status.Status == interfaces.CatalogHealthStatusDegraded {
        logger.Warnf("catalog %s is in degraded state (latency: %dms)", catalogID, status.LastLatencyMs)
    }
    
    // 4. 执行查询
    result, err := s.connector.Query(query)
    
    return result, err
}
```

#### 3.3.2 Fail-fast 策略

1. 检查 catalog 健康状态，不健康（unhealthy/offline）则直接返回错误
2. degraded 状态允许查询，但记录警告日志
3. 查询失败时不需要立即更新健康状态（由定期健康检查机制处理）
4. 查询成功时不需要立即更新健康状态（由定期健康检查机制维护）



## 4. 数据模型变更

### 4.1 Catalog 表字段说明

Catalog 表已包含健康检查相关字段，无需新增字段，只需确保在连接测试后正确更新这些字段：

```sql
-- 健康检查配置字段
f_health_check_enabled BOOLEAN DEFAULT true;  -- 健康检查是否启用
f_health_check_config TEXT;                    -- 健康检查配置（JSON格式），与 connector_config 平齐

-- 健康检查状态字段（由健康检查流程更新）
f_health_check_status VARCHAR(20);              -- 健康状态: healthy, degraded, unhealthy, offline, disabled
f_last_check_time BIGINT;                       -- 最后检查时间戳
f_last_success_time BIGINT;                     -- 最后成功检查时间戳
f_failure_count INT DEFAULT 0;                  -- 连续失败次数
f_success_count INT DEFAULT 0;                  -- 连续成功次数
f_last_latency_ms BIGINT;                       -- 最后一次检查延迟（毫秒）
f_health_check_result TEXT;                     -- 检查结果描述
```

**更新逻辑**:
- `f_health_check_status`: 根据连接测试结果和阈值更新为以下状态之一：
  - `healthy`: 连接测试成功，延迟 <= latency_warning_ms
  - `degraded`: 连接测试成功，延迟 > latency_warning_ms 且 <= latency_critical_ms
  - `unhealthy`: 连接测试失败或延迟 > latency_critical_ms
  - `offline`: 连接超时或网络不可达
  - `disabled`: 健康检查被禁用（f_health_check_enabled=false）
- `f_last_check_time`: 更新为当前时间戳
- `f_last_success_time`: 检查成功时更新为当前时间戳
- `f_failure_count`: 检查失败时递增，检查成功时重置为 0
- `f_success_count`: 检查成功时递增，检查失败时重置为 0
- `f_last_latency_ms`: 更新为最后一次检查的延迟（毫秒）
- `f_health_check_result`: 更新为包含延迟、错误信息和状态描述的字符串

**阈值判定**:
- 当 `f_failure_count >= failure_threshold` 时，状态保持为 `unhealthy`
- 当 `f_success_count >= recovery_threshold` 时，状态恢复为 `healthy` 或 `degraded`

**注意**: 
- `f_health_check_enabled` 字段由用户通过创建/更新 catalog 接口设置
- 健康检查配置（strategy、thresholds、retry）存储在独立的 `f_health_check_config` 字段中，与 `connector_config` 平齐
- 这些配置不在健康检查流程中修改
- **历史 catalog 处理**：对于在此需求开发之前已存在的 catalog，需要通过更新接口启用健康检查功能，这些 catalog 的 `f_health_check_enabled` 字段默认为 false

## 5. API 变更

### 5.1 POST /api/v1/catalogs - 创建 Catalog

**变更内容**：
- 请求体中新增 `health_check_config` 字段，用于配置健康检查参数
- 创建成功后，初始化健康检查相关字段：
  - `f_health_check_enabled`: 根据请求参数设置
  - `f_health_check_status`: 初始值为 "healthy"
  - `f_last_check_time`: 当前时间戳
  - `f_last_success_time`: 当前时间戳
  - `f_failure_count`: 0
  - `f_success_count`: 0
  - `f_last_latency_ms`: 0
  - `f_health_check_result`: "Initial health check"

**请求示例**：
```json
{
  "name": "mysql_prod",
  "type": "physical",
  "connector_type": "mysql",
  "connector_config": {
    "host": "db.example.com",
    "port": 3306,
    "database": "production",
    "username": "admin",
    "password": "******"
  },
  "health_check_enabled": true,
  "health_check_config": {
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
}
```

### 5.2 PUT /api/v1/catalogs/{id} - 更新 Catalog

**变更内容**：
- 请求体中支持更新 `health_check_enabled` 字段
- 请求体中支持更新 `health_check_config` 字段
- 更新健康检查配置时，不会重置健康检查状态
- 如果将 `health_check_enabled` 从 true 改为 false，状态将变为 "disabled"
- **历史 catalog 启用**：对于在此需求开发之前已存在的 catalog，可以通过此接口设置 `health_check_enabled=true` 来启用健康检查功能

**请求示例**：
```json
{
  "name": "mysql_prod",
  "description": "Production MySQL database",
  "health_check_enabled": true,
  "health_check_config": {
    "strategy": {
      "interval": "60s",
      "timeout": "10s"
    },
    "thresholds": {
      "latency_warning_ms": 1000,
      "latency_critical_ms": 3000,
      "failure_threshold": 5,
      "recovery_threshold": 3
    },
    "retry": {
      "max_attempts": 5,
      "backoff": "exponential",
      "initial_delay": "2s",
      "max_delay": "60s"
    }
  }
}
```

### 5.3 GET /api/v1/catalogs/{id} - 获取 Catalog 详情

**变更内容**：
- 返回体中新增完整的健康检查状态信息
- 包括所有健康检查相关字段的当前值

**返回示例**：
```json
{
  "id": "catalog_001",
  "name": "mysql_prod",
  "type": "physical",
  "connector_type": "mysql",
  "connector_config": {
    "host": "db.example.com",
    "port": 3306,
    "database": "production",
    "username": "admin"
  },
  "health_check_enabled": true,
  "health_check_status": "healthy",
  "last_check_time": 1712345678900,
  "last_success_time": 1712345678900,
  "failure_count": 0,
  "success_count": 5,
  "last_latency_ms": 123,
  "health_check_result": "Connection successful (latency: 123ms)",
  "health_check_config": {
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
  },
  "create_time": 1712000000000,
  "update_time": 1712345678900
}
```

### 5.4 GET /api/v1/catalogs - 获取 Catalog 列表

**变更内容**：
- 查询参数中新增 `health_check_status` 过滤条件
- 返回体中包含每个 catalog 的健康检查状态信息

**查询参数**：
- `health_check_status`: 可选，过滤指定健康状态的 catalog（healthy/degraded/unhealthy/offline/disabled）

**请求示例**：
```
GET /api/v1/catalogs?health_check_status=healthy&page=1&page_size=20
```

**返回示例**：
```json
{
  "items": [
    {
      "id": "catalog_001",
      "name": "mysql_prod",
      "type": "physical",
      "connector_type": "mysql",
      "health_check_enabled": true,
      "health_check_status": "healthy",
      "last_check_time": 1712345678900,
      "last_success_time": 1712345678900,
      "failure_count": 0,
      "success_count": 5,
      "last_latency_ms": 123,
      "health_check_result": "Connection successful (latency: 123ms)"
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

## 6. 监控与告警

### 6.1 监控指标

1. **健康检查指标**
   - catalog_health_check_total - 健康检查总次数
   - catalog_health_check_success - 健康检查成功次数
   - catalog_health_check_failure - 健康检查失败次数
   - catalog_health_check_latency_ms - 健康检查延迟

### 6.2 告警规则

1. catalog 健康检查连续失败 3 次
2. 健康检查延迟超过阈值
3. 不健康 catalog 数量超过阈值

## 7. 附录

### 7.1 配置示例

#### 7.1.1 全局配置 (application.yaml)

**作用说明**:
- `application.yaml` 是 vega-backend 应用的全局配置文件
- `health_check` 配置项控制健康检查服务的全局行为
- 这些配置项为所有 catalog 提供默认值，当 catalog 的 `health_check_config` 中未指定具体值时使用

```yaml
# application.yaml
health_check:
  # ========== 基础配置 ==========
  default_interval: "30s"     # 默认检查间隔（catalog 未配置时使用）
  default_timeout: "5s"      # 默认单次检查超时（catalog 未配置时使用）
  default_max_workers: 10    # 默认最大并发检查数（worker pool 大小）
  
  # ========== 默认状态判定阈值（catalog 未配置时使用） ==========
  default_latency_warning_ms: 500      # 延迟警告阈值（毫秒），超过此值标记为 degraded
  default_latency_critical_ms: 2000     # 延迟严重阈值（毫秒），超过此值标记为 unhealthy
  default_failure_threshold: 3           # 连续失败阈值，达到此值将状态标记为 unhealthy
  default_recovery_threshold: 2          # 恢复阈值，连续成功次数达到此值将状态恢复为 healthy
  
  # ========== 默认重试策略（catalog 未配置时使用） ==========
  default_retry_max_attempts: 3         # 单次检查失败时的最大重试次数
  default_retry_backoff: "exponential"   # 重试退避策略：fixed（固定延迟）或 exponential（指数退避）
  default_retry_initial_delay: "1s"      # 重试初始延迟时间
  default_retry_max_delay: "30s"          # 重试最大延迟时间
```

#### 7.1.2 Catalog 配置 (通过 API 创建/更新 catalog)

```json
{
  "name": "mysql_prod",
  "type": "physical",
  "connector_config": {
    "type": "mysql",
    "host": "db.example.com",
    "port": 3306,
    "database": "production",
    "username": "admin",
    "password": "******"
  },
    
  "health_check_config": {
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
}
```

### 7.2 术语表

| 术语 | 定义 |
|------|------|
| Catalog | 数据源目录，包含数据源连接信息和元数据 |
| Connector | 连接器，用于与特定类型的数据源进行交互 |
| Health Check | 健康检查，用于检测数据源是否可用，包括定期检查和手动测试 |
| Health Status | 健康状态，表示数据源的可用性状态：healthy（健康）、degraded（降级）、unhealthy（不健康）、offline（离线）、disabled（禁用）|
| Strategy | 检查策略，定义健康检查的间隔和超时时间 |
| Thresholds | 状态判定阈值，包括延迟警告阈值、延迟严重阈值、失败阈值和恢复阈值 |
| Retry | 重试策略，定义检查失败时的重试次数和退避策略 |
| Fail-fast | 快速失败策略，在检测到数据源不可用时立即返回错误，避免无效等待 |
| Worker Pool | 工作线程池，用于控制并发健康检查的数量 |
| Latency | 延迟，表示健康检查的响应时间（毫秒）|
