# VEGA Resource 数据查询限流控制设计文档

**作者**: Sisyphus (AI Agent)  
**日期**: 2026-04-29  
**状态**: ✅ 已完成  
**版本**: v3.0 (当前只实现并发控制)

---

## 1. 背景与目标

### 1.1 当前状态分析

通过对 VEGA Backend 代码库的深入探索，我们发现当前系统**缺乏生产级别的限流控制机制**：

| 限流维度 | 当前状态 | 实现位置 | 说明 |
|---------|---------|---------|------|
| **并发控制** | ❌ 无 | - | 无限制同时执行的查询数量 |
| **QPS 限流** | ❌ 无 | - | 无每秒查询数限制 |
| **内存限制** | ⚠️ 基础 | `query_service.go:29` | 仅行数限制 (10000 行)，无内存预算 |
| **吞吐量控制** | ❌ 无 | - | 无数据量/字节数限制 |
| **查询复杂度** | ❌ 无 | - | 无复杂度评分机制 |
| **工作负载隔离** | ❌ 无 | - | 无优先级/配额隔离 |

**现有基础限制**：
```go
// query_service.go:29
const (
    QueryLimitMax = 10000  // 仅作为结果截断，非真正的限流
)

// sql_query_service.go:283-302
// SQL 执行前自动添加 LIMIT 10000
// 但这只是结果限制，无法防止长耗时查询占用资源
```

### 1.2 问题陈述

随着查询负载增加，当前设计面临以下**生产风险**：

1. **资源耗尽风险**
   - 大量并发查询可能耗尽数据库连接池
   - 单个大查询可能占用全部内存
   - 无连接池配置（当前为每次查询新建连接）

2. **查询饥饿问题**
   - 长耗时查询阻塞其他查询执行
   - 无优先级机制，关键业务查询可能被阻塞
   - 无超时强制终止机制

3. **系统稳定性风险**
   - 单个客户端可发起无限并发查询
   - 无降级机制，过载时可能导致雪崩
   - 无熔断保护，故障数据源持续影响系统

4. **多租户不公平**
   - 无法对不同用户/团队实施差异化配额
   - 无资源隔离，嘈杂邻居问题严重

### 1.3 设计目标

本设计旨在为 VEGA Resource 数据查询添加**全面的限流控制**，实现以下目标：

| 目标 | 描述 | 优先级 | 验收标准 |
|------|------|--------|----------|
| **并发控制** | 限制同时执行的查询数量，防止资源耗尽 | P0 | 系统在高并发下保持稳定 |
| **QPS 限流** | 限制每秒查询数，保护下游服务 | P0 | 查询速率可配置、可监控 |
| **内存限制** | 限制单个查询和全局内存使用 | P1 | 单查询内存不超过配额 |
| **吞吐量控制** | 限制数据返回量 (bytes/rows) | P1 | 数据吞吐量可配置 |
| **可观测性** | 提供限流相关的指标和日志 | P0 | 所有限流事件可追踪 |
| **优雅降级** | 限流时返回明确错误和重试建议 | P0 | HTTP 429/503 响应规范 |
| **工作负载隔离** | 不同优先级查询的资源隔离 | P2 | 关键查询不受低优先级影响 |

---

## 2. 限流方案对比与选型

### 2.1 算法对比

基于业界最佳实践研究（Cloudflare、GitHub、Uber、Apache Pinot 等），我们对比以下算法：

| 算法 | 优点 | 缺点 | 适用场景 | 推荐度 |
|------|------|------|----------|--------|
| **令牌桶 (Token Bucket)** | 允许可控突发，O(1) 内存，实现简单 | 边界突发放大 | API 限流、通用场景 | ⭐⭐⭐⭐⭐ |
| **滑动窗口计数器** | 精确控制，无边界突发问题，O(1) 内存 | 实现稍复杂 | 高并发场景 (Cloudflare 标准) | ⭐⭐⭐⭐⭐ |
| **漏桶 (Leaky Bucket)** | 严格限速，输出平滑 | 不支持突发，增加延迟 | 下游保护、流量整形 | ⭐⭐⭐ |
| **固定窗口** | 实现简单 | 边界突发 2x 问题 | 不推荐用于生产 | ⭐ |
| **滑动窗口日志** | 精确到毫秒 | O(n) 内存，性能差 | 极低频高精度场景 | ⭐⭐ |

**推荐方案**: **令牌桶 + 滑动窗口计数器** 组合
- **令牌桶**: 用于允许突发流量（用户友好）
- **滑动窗口计数器**: 用于严格 QPS 限制（精确控制）

### 2.2 并发控制方案

| 方案 | 实现 | 优点 | 缺点 | 推荐度 |
|------|------|------|------|--------|
| **信号量 (Semaphore)** | Go `sync.Semaphore` 或缓冲通道 | 简单高效，原生支持 | 单机限制 | ⭐⭐⭐⭐⭐ |
| **加权信号量** | `golang.org/x/sync/semaphore` | 支持不同资源成本 | 需要预估成本 | ⭐⭐⭐⭐ |
| **连接池** | `database/sql` 连接池 | 直接控制 DB 连接 | 仅适用于 DB | ⭐⭐⭐⭐ |
| **分布式信号量** | Redis Lua 脚本 | 跨实例统一限制 | 依赖外部存储 | ⭐⭐⭐ |

**推荐方案**: **本地加权信号量 + 全局 Redis 分布式信号量**（二阶段）
- **Phase 1**: 单机信号量（快速路径，无网络依赖）
- **Phase 2**: 分布式信号量（多实例部署时）

### 2.3 内存控制方案

基于 Apache Pinot、TiDB、Presto 等生产系统设计：

| 策略 | 实现 | 优点 | 缺点 | 推荐度 |
|------|------|------|------|--------|
| **行数限制** | `len(result) > maxRows` | 简单，已有实现 | 不准确（字段大小不一） | ⭐⭐⭐ |
| **字节估算** | 预估每行字节数 | 更准确 | 需要计算开销 | ⭐⭐⭐⭐ |
| **内存追踪器** | 类似 TiDB Tracker | 精确控制 | 实现复杂 | ⭐⭐⭐⭐ |
| **溢出到磁盘** | 内存不足时写磁盘 | 支持大查询 | I/O 开销大 | ⭐⭐ |

**推荐方案**: **行数限制 + 字节估算** 组合
- 保持现有 10000 行限制作为硬性上限
- 新增字节估算作为软性限制（可配置告警）

---

## 3. 多层级限流架构设计

### 3.1 整体架构

```
┌──────────────────────────────────────────────────────────────────┐
│                     Query Request Flow                        │
│                                                                  │
│  1. HTTP Handler 层                                              │
│     ├─ 工作负载分类 (X-Workload-Priority header)                │
│     └─ 认证/授权                                                │
│         ↓                                                       │
│  2. 全局 QPS 限流 (令牌桶，Redis 分布式)                          │
│     ├─ 全局 QPS 限制 (例如 1000 QPS)                              │
│     └─ 按租户/用户 QPS 限制                                        │
│         ↓                                                       │
│  3. 全局并发限流 (信号量)                                         │
│     ├─ 全局并发查询数限制 (例如 100)                              │
│     └─ 等待队列 (FIFO，超时 30s)                                  │
│         ↓                                                       │
│  4. Catalog 级限流 (令牌桶 + 信号量)                              │
│     ├─ 按 Catalog 的 QPS 限制                                       │
│     └─ 按 Catalog 的并发限制                                       │
│         ↓                                                       │
│  5. 资源预算检查                                                  │
│     ├─ 内存预算检查 (基于行数估算)                                │
│     ├─ 查询复杂度评分 (JOIN 数量、子查询等)                        │
│     └─ 超时时间检查                                              │
│         ↓                                                       │
│  6. 查询执行 (带内存监控)                                         │
│     ├─ 连接池获取 (受限于 DB 配置)                                  │
│     ├─ 实际查询执行                                              │
│     └─ 结果集大小监控                                            │
│         ↓                                                       │
│  7. 结果返回 / 限流错误                                           │
│     ├─ 成功：返回数据 + 限流指标头                                │
│     └─ 失败：HTTP 429/503 + Retry-After                          │
└──────────────────────────────────────────────────────────────────┘
```

### 3.2 限流层次说明

```
Layer 1: HTTP Handler (query_handler.go)
  ├─ 工作负载分类 (critical/high/normal/low)
  └─ 用户/租户识别 (从 JWT/Context)

Layer 2: Global QPS Limiter (rate_limiter.go)
  ├─ 全局令牌桶 (golang.org/x/time/rate)
  └─ Redis 分布式令牌桶 (Lua 脚本)

Layer 3: Global Concurrency Limiter (concurrency_limiter.go)
  ├─ 全局信号量 (sync.Semaphore)
  └─ 等待队列 (channel + context 超时)

Layer 4: Catalog Limiter (catalog_limiter.go)
  ├─ 每 Catalog 令牌桶 (map[catalogID]*rate.Limiter)
  └─ 每 Catalog 信号量 (map[catalogID]*sync.Semaphore)

Layer 5: Resource Budget Checker (budget_checker.go)
  ├─ 内存预算检查 (估算结果集大小)
  ├─ 复杂度评分 (SQL 解析)
  └─ 超时时间验证

Layer 6: Query Executor (query_service.go)
  └─ 实际执行，带内存追踪
```

---

## 4. 详细设计

### 4.1 并发控制 (Concurrency Limiting)

#### 4.1.1 设计目标

**核心原则：全局控制与 Catalog 级控制是互补关系，而非互斥关系**

| 控制层级 | 保护对象 | 配置依据 | 典型值 | 实现状态 |
|---------|---------|---------|--------|---------|
| **全局并发控制** | VEGA 服务自身稳定性 | VEGA 服务资源容量（内存、CPU、连接池） | 10-100 | ✅ 已实现 |
| **Catalog 级并发控制** | 下游数据源稳定性 | 下游数据库/服务实际能力 | 5-50 | ✅ 已实现 |

**互补关系说明**：
- **全局限制**防止 VEGA 服务过载（即使下游能处理更多）
- **Catalog 限制**防止压垮特定下游服务（即使 VEGA 还能处理）
- **两者同时生效**，取更严格的限制（AND 逻辑，非 OR 逻辑）

```
并发许可 = 全局许可 AND Catalog 许可
         = 全局信号量可用 AND Catalog 信号量可用
```

#### 4.1.2 配置参数

**全局并发控制**（配置文件）：

```yaml
# server/config/vega-backend-config.yaml
rateLimiting:
  concurrency:
    enabled: true
    
    # === 全局并发控制：保护 VEGA 服务自身 ===
    global:
      # 基于 VEGA 服务资源容量设置
      max_concurrent_queries: 10       # VEGA 最大并发处理能力
```

**Catalog 级并发控制**（从 Catalog 配置获取）：

Catalog 的 `ConnectorCfg` 中包含并发配置：

| 键名 | 类型 | 说明 | 默认值 |
|------|------|------|--------|
| `concurrent` | float64 | Catalog 最大并发查询数 | 0（不限制） |

**说明**：
- `concurrent = 0` 或未设置：不限制，仅使用全局限制
- `concurrent > 0`：按此数值限制
- Catalog 配置在运行时动态读取，支持热更新

#### 4.1.3 接口定义

**当前实现**（基于实际代码）：

```go
// server/logics/rate/concurrency_limiter.go

package rate

// ConcurrencyLimiter 并发控制器接口
type ConcurrencyLimiter interface {
    // Acquire 尝试获取执行许可
    // 返回：(释放函数，错误)
    // 错误类型：ErrGlobalLimitExceeded | ErrCatalogLimitExceeded
    Acquire(params AcquireParams) (release ReleaseFunc, err error)
    
    // Close 释放所有资源
    Close()
}

// AcquireParams 获取许可的参数
type AcquireParams struct {
    CatalogID            string // Catalog ID（必填）
    MaxConcurrentQueries int64  // Catalog 最大并发查询数
}

// ReleaseFunc 释放许可的函数
// 注意：必须同时释放全局和 Catalog 许可
type ReleaseFunc func()

// 限流错误类型
var (
    ErrGlobalLimitExceeded  = errors.New("global concurrency limit exceeded")
    ErrCatalogLimitExceeded = errors.New("catalog concurrency limit exceeded")
)
```

**实现特点**：
1. **无阻塞获取**：使用 `TryAcquire` 而非 `Acquire`，立即返回失败而非等待
2. **无等待队列**：当前实现不支持排队等待，直接拒绝超限请求
3. **简化接口**：移除了工作负载优先级、超时控制等 Phase 2 功能
4. **同步释放**：释放函数在查询完成后同步调用

#### 4.1.4 实现方案

**当前实现**（基于实际代码）：

```go
// server/logics/rate/concurrency_limiter.go

package rate

import (
    "fmt"
    "sync"
    
    "golang.org/x/sync/semaphore"
)

// concurrencyLimiter 实现两级并发控制
type concurrencyLimiter struct {
    cfg ConcurrencyConfig
    
    // 全局信号量：保护 VEGA 服务自身
    globalSem *semaphore.Weighted
    
    // Catalog 信号量：保护下游服务
    // Key: catalogID, Value: *catalogSemaphore
    catalogSems sync.Map
}

// catalogSemaphore 封装 Catalog 信号量
type catalogSemaphore struct {
    sem *semaphore.Weighted
}

// NewConcurrencyLimiter 创建并发限流器
func NewConcurrencyLimiter(cfg ConcurrencyConfig) ConcurrencyLimiter {
    cl := &concurrencyLimiter{
        cfg:       cfg,
        globalSem: semaphore.NewWeighted(int64(cfg.Global.MaxConcurrentQueries)),
    }
    
    return cl
}

// Acquire 尝试获取执行许可（非阻塞）
func (cl *concurrencyLimiter) Acquire(params AcquireParams) (ReleaseFunc, error) {
    // 1. 尝试获取全局许可（保护 VEGA 服务）
    if !cl.globalSem.TryAcquire(1) {
        return nil, NewRateLimitError(ErrGlobalLimitExceeded,
            fmt.Sprintf("global concurrency limit exceeded, max=%d", cl.cfg.Global.MaxConcurrentQueries))
    }
    
    // 2. 尝试获取 Catalog 许可（保护下游服务）
    var catalogSem *catalogSemaphore
    
    if params.CatalogID != "" && params.MaxConcurrentQueries > 0 {
        catalogSem = cl.getOrCreateCatalogSemaphore(params.CatalogID, params.MaxConcurrentQueries)
        
        if !catalogSem.sem.TryAcquire(1) {
            // Catalog 限流：释放全局许可
            cl.globalSem.Release(1)
            
            return nil, NewRateLimitError(ErrCatalogLimitExceeded,
                fmt.Sprintf("catalog concurrency limit exceeded for catalog=%s, max=%d",
                    params.CatalogID, params.MaxConcurrentQueries))
        }
    }
    
    // 3. 成功，返回释放函数
    return func() {
        // 释放 Catalog 许可（如果获取了）
        if catalogSem != nil {
            catalogSem.sem.Release(1)
        }
        
        // 释放全局许可
        cl.globalSem.Release(1)
    }, nil
}

// getOrCreateCatalogSemaphore 获取或创建 Catalog 信号量
func (cl *concurrencyLimiter) getOrCreateCatalogSemaphore(catalogID string, maxConcurrentQueries int64) *catalogSemaphore {
    if sem, ok := cl.catalogSems.Load(catalogID); ok {
        return sem.(*catalogSemaphore)
    }
    
    // 创建新的 Catalog 信号量
    newSem := &catalogSemaphore{
        sem: semaphore.NewWeighted(maxConcurrentQueries),
    }
    
    // 使用 LoadOrStore 避免竞争条件
    actual, loaded := cl.catalogSems.LoadOrStore(catalogID, newSem)
    if loaded {
        return actual.(*catalogSemaphore)
    }
    
    return newSem
}
```

**实现特点**：

| 特性 | 实现方式 | 说明 |
|------|---------|------|
| **获取方式** | `TryAcquire` | 非阻塞，立即返回失败 |
| **等待队列** | 不支持 | 超限直接拒绝，不排队 |
| **工作负载隔离** | 未实现 | Phase 2 功能 |
| **超时控制** | 未实现 | Phase 2 功能 |
| **统计指标** | 未实现 | Phase 2 功能 |
| **动态更新** | 不支持 | 信号量创建后不可变 |

#### 4.1.5 互补关系验证（当前实现）

**场景 1：全局限制更严格**
```
全局限制：10 个并发
Catalog A 限制：20 个并发

实际并发：
- Catalog A: max 10 (受全局限制)
```

**场景 2：Catalog 限制更严格**
```
全局限制：100 个并发
Catalog A 限制：5 个并发（下游能力弱）

实际并发：
- Catalog A: max 5 (受 Catalog 限制)
```

**场景 3：两级同时生效**
```
全局限制：100 个并发
Catalog A 限制：20 个并发
Catalog B 限制：30 个并发
Catalog C 限制：40 个并发

实际并发：
- Catalog A: max 20
- Catalog B: max 30
- Catalog C: max 40
- 总计：90 < 100 (受 Catalog 限制总和，全局限制未达上限)

如果 Catalog A、B、C 同时达到上限：
- 总计：90 个并发
- 全局剩余：10 个配额可用于其他 Catalog
```

**当前实现限制**：

| 限制 | 说明 | 改进计划 |
|------|------|---------|
| **LogicView不支持catalog级别并发限制** | LogicView的catalog在查询入口难获取，暂时只支持全局并发限制 | 待定 |
| **无等待队列** | 超限直接拒绝 | Phase 2 支持排队 |
| **无超时控制** | 立即返回失败 | Phase 2 支持超时 |
| **无优先级** | 所有请求平等 | Phase 2 支持工作负载隔离 |
| **无统计指标** | 无法监控限流效果 | Phase 2 添加 Prometheus 指标 |

### 4.2 QPS 限流 (Rate Limiting) - Phase 2

*当前未实现，计划在 Phase 2 添加*

#### 4.2.1 设计目标
- 限制每秒查询数
- 支持突发流量（令牌桶）
- 支持分布式部署（Redis）

#### 4.2.2 配置参数
```yaml
rate_limiting:
  qps:
    enabled: true
    
    # 全局 QPS 限制
    global:
      requests_per_second: 1000     # 全局 QPS 限制
      burst_size: 2000              # 允许突发大小
    
    # 按租户 QPS 限制
    per_tenant:
      enabled: true
      default:
        requests_per_second: 100
        burst_size: 200
      tiers:                         # 不同租户等级的限制
        enterprise:
          requests_per_second: 500
          burst_size: 1000
        professional:
          requests_per_second: 200
          burst_size: 400
        free:
          requests_per_second: 50
          burst_size: 100
    
    # 按 Catalog QPS 限制
    per_catalog:
      enabled: true
      default:
        requests_per_second: 200
        burst_size: 500
    
    # 分布式配置
    distributed:
      enabled: false                # Phase 2 启用
      redis:
        cluster_mode: true
        key_prefix: "vega:ratelimit:"
        sync_interval_ms: 100       # 批量同步间隔
```

#### 4.2.3 接口定义
```go
// server/logics/rate/qps_limiter.go

package rate

import "context"

// QPSLimiter QPS 限流器接口
type QPSLimiter interface {
    // Allow 检查是否允许请求
    Allow(ctx context.Context, params RateLimitParams) error
    
    // GetStatus 获取限流状态
    GetStatus(ctx context.Context, params RateLimitParams) RateLimitStatus
}

// RateLimitParams 限流参数
type RateLimitParams struct {
    UserID    string // 用户 ID
    TenantID  string // 租户 ID
    CatalogID string // Catalog ID
}

// RateLimitStatus 限流状态
type RateLimitStatus struct {
    Limit       int     // 限制值
    Remaining   int     // 剩余配额
    ResetTime   int64   // 重置时间 (Unix timestamp)
    RetryAfter  float64 // 建议重试时间 (秒)
}
```

#### 4.2.4 实现方案
```go
// server/logics/rate/qps_limiter.go

type qpsLimiter struct {
    cfg QPSConfig
    
    // 本地令牌桶 (快速路径)
    globalLimiter   *rate.Limiter
    tenantLimiters  sync.Map // map[string]*rate.Limiter
    catalogLimiters sync.Map // map[string]*rate.Limiter
    
    // 分布式限流器 (Phase 2)
    distributedLimiter *DistributedQPSLimiter
}

func (ql *qpsLimiter) Allow(ctx context.Context, params RateLimitParams) error {
    // 1. 全局 QPS 检查（本地）
    if !ql.globalLimiter.Allow() {
        return ql.createRateLimitError("global_qps", params)
    }
    
    // 2. 租户 QPS 检查
    if ql.cfg.PerTenant.Enabled && params.TenantID != "" {
        limiter := ql.getTenantLimiter(params.TenantID)
        if !limiter.Allow() {
            return ql.createRateLimitError("tenant_qps", params)
        }
    }
    
    // 3. Catalog QPS 检查
    if ql.cfg.PerCatalog.Enabled && params.CatalogID != "" {
        limiter := ql.getCatalogLimiter(params.CatalogID)
        if !limiter.Allow() {
            return ql.createRateLimitError("catalog_qps", params)
        }
    }
    
    // 4. 分布式 QPS 检查 (Phase 2)
    if ql.cfg.Distributed.Enabled {
        if err := ql.distributedLimiter.Allow(ctx, params); err != nil {
            return err
        }
    }
    
    return nil
}

func (ql *qpsLimiter) getTenantLimiter(tenantID string) *rate.Limiter {
    if limiter, ok := ql.tenantLimiters.Load(tenantID); ok {
        return limiter.(*rate.Limiter)
    }
    
    // 创建新的限流器
    cfg := ql.cfg.PerTenant.Default
    limiter := rate.NewLimiter(rate.Limit(cfg.RequestsPerSecond), cfg.BurstSize)
    
    ql.tenantLimiters.Store(tenantID, limiter)
    return limiter
}
```

### 4.3 内存限制 (Memory Limiting)

*当前未实现，依赖现有行数限制*

#### 4.3.1 设计目标
- 估算查询结果集大小
- 防止单个查询耗尽内存
- 提供内存使用监控

#### 4.3.2 配置参数
```yaml
rate_limiting:
  memory:
    enabled: true
    
    # 单查询内存限制
    per_query:
      max_memory_bytes: 1073741824       # 1 GB
      max_rows: 50000                    # 单查询最大行数（软限制）
      avg_row_size_bytes: 1024           # 平均每行估算大小 (1KB)
    
    # 全局内存限制
    global:
      max_total_memory_bytes: 8589934592 # 8 GB
      warning_threshold: 0.8             # 80% 时告警
      eviction_policy: "reject_new"      # reject_new | cancel_longest
    
    # 内存监控
    monitoring:
      enabled: true
      check_interval_ms: 1000            # 检查间隔
      gc_before_check: true              # 检查前强制 GC
```

#### 4.3.3 实现方案
```go
// server/logics/rate/memory_checker.go

package rate

type MemoryChecker struct {
    cfg MemoryConfig
}

// CheckBudget 检查内存预算
func (mc *MemoryChecker) CheckBudget(ctx context.Context, estimatedRows int64) error {
    // 1. 单查询检查
    estimatedMemory := estimatedRows * mc.cfg.PerQuery.AvgRowSizeBytes
    if estimatedMemory > mc.cfg.PerQuery.MaxMemoryBytes {
        return NewRateLimitError("memory_budget_exceeded", 
            fmt.Sprintf("estimated memory %d bytes exceeds limit %d bytes", 
                estimatedMemory, mc.cfg.PerQuery.MaxMemoryBytes))
    }
    
    // 2. 全局内存检查
    if mc.cfg.Global.Enabled {
        var m runtime.MemStats
        runtime.ReadMemStats(&m)
        
        if m.Alloc > uint64(mc.cfg.Global.MaxTotalMemoryBytes) {
            return NewRateLimitError("system_memory_high",
                fmt.Sprintf("system memory usage too high: %d bytes", m.Alloc))
        }
    }
    
    return nil
}

// EstimateResultSize 估算结果集大小
func (mc *MemoryChecker) EstimateResultSize(columns []Column, estimatedRows int64) int64 {
    var totalSize int64
    
    for _, col := range columns {
        size := mc.estimateColumnSize(col)
        totalSize += size * estimatedRows
    }
    
    return totalSize
}

func (mc *MemoryChecker) estimateColumnSize(col Column) int64 {
    switch col.Type {
    case "integer", "int64":
        return 8
    case "float", "float64":
        return 8
    case "boolean":
        return 1
    case "string", "text":
        return int64(mc.cfg.PerQuery.AvgRowSizeBytes / 3) // 假设 3 个字符串字段
    default:
        return int64(mc.cfg.PerQuery.AvgRowSizeBytes / 5) // 默认值
    }
}
```

### 4.4 查询复杂度限制 (Query Complexity)

*当前未实现*

#### 4.4.1 设计目标
- 基于 SQL 特征计算复杂度分数
- 阻止过于复杂的查询
- 提供优化建议

#### 4.4.2 复杂度评分规则
```go
// server/logics/rate/complexity_scorer.go

type ComplexityScorer struct {
    cfg ComplexityConfig
}

// Score 计算查询复杂度
func (cs *ComplexityScorer) Score(sql string) (int, error) {
    score := 0
    
    // 1. JOIN 数量 (每个 +20 分)
    joinCount := countKeywords(sql, "JOIN")
    score += joinCount * 20
    
    // 2. 子查询数量 (每个 +15 分)
    subqueryCount := countSubqueries(sql)
    score += subqueryCount * 15
    
    // 3. 聚合函数 (每个 +10 分)
    aggCount := countKeywords(sql, "GROUP BY", "SUM(", "COUNT(", "AVG(", "MAX(", "MIN(")
    score += aggCount * 10
    
    // 4. DISTINCT ( +10 分)
    if strings.Contains(strings.ToUpper(sql), "DISTINCT") {
        score += 10
    }
    
    // 5. 排序 ( +5 分)
    if strings.Contains(strings.ToUpper(sql), "ORDER BY") {
        score += 5
    }
    
    // 6. 跨数据源 ( +5 分)
    // TODO: 实现跨数据源检测
    
    return score, nil
}

// 复杂度等级
const (
    ComplexityLow    = 50   // 低复杂度
    ComplexityMedium = 100  // 中等复杂度
    ComplexityHigh   = 150  // 高复杂度
    ComplexityCritical = 200 // 临界复杂度
)
```

### 4.5 工作负载隔离 (Workload Isolation)

*当前未实现*

#### 4.5.1 设计目标
- 按工作负载类型分配不同资源配额
- 高优先级查询优先执行
- 防止低优先级查询影响关键业务

#### 4.5.2 配置参数
```yaml
rate_limiting:
  workload_isolation:
    enabled: true
    
    workloads:
      critical:
        description: "关键业务查询"
        concurrency: 30
        qps: 300
        memory: 4294967296  # 4 GB
        priority: 100
        timeout: 60s
        query_types: ["standard", "stream"]
        
      high:
        description: "高优先级查询"
        concurrency: 40
        qps: 400
        memory: 2147483648  # 2 GB
        priority: 75
        timeout: 120s
        query_types: ["standard", "stream"]
        
      normal:
        description: "普通查询"
        concurrency: 20
        qps: 200
        memory: 1073741824  # 1 GB
        priority: 50
        timeout: 300s
        query_types: ["standard", "stream"]
        
      low:
        description: "低优先级/批量查询"
        concurrency: 10
        qps: 100
        memory: 536870912   # 512 MB
        priority: 25
        timeout: 600s
        query_types: ["standard"]
```

#### 4.5.3 工作负载分类
```go
// server/logics/rate/workload_classifier.go

func classifyQuery(ctx context.Context, req *QueryRequest) string {
    // 1. 从请求头获取（显式指定）
    if priority := ctx.Value("X-Workload-Priority"); priority != nil {
        return priority.(string)
    }
    
    // 2. 根据用户等级自动分类
    if user := getCurrentUser(ctx); user != nil {
        switch user.Tier {
        case "enterprise":
            return "high"
        case "professional":
            return "normal"
        case "free":
            return "low"
        }
    }
    
    // 3. 根据查询类型自动分类
    if req.QueryType == "stream" {
        return "normal"
    }
    
    // 4. 默认分类
    return "normal"
}
```

---

## 5. 集成方案

### 5.1 代码修改位置

#### 5.1.1 服务初始化
```go
// server/logics/resource_data/resource_data_service.go

func NewResourceDataService(appSetting *common.AppSetting) interfaces.ResourceDataService {
    rdServiceOnce.Do(func() {
        rdService = &resourceDataService{
            appSetting: appSetting,
            ds:         dataset.NewDatasetService(appSetting),
            cs:         catalog.NewCatalogService(appSetting),
            rs:         resource.NewResourceService(appSetting),
            lvs:        logic_view.NewLogicViewService(appSetting),
        }

        // 初始化并发限流器（如果启用）
        if appSetting.RateLimitingSetting.Concurrency.Enabled && 
           appSetting.RateLimitingSetting.Concurrency.Global.MaxConcurrentQueries > 0 {
            cfg := rate.ConcurrencyConfig{
                Enabled: appSetting.RateLimitingSetting.Concurrency.Enabled,
                Global: rate.GlobalConcurrencyConfig{
                    MaxConcurrentQueries: appSetting.RateLimitingSetting.Concurrency.Global.MaxConcurrentQueries,
                },
            }

            rdService.(*resourceDataService).cl = rate.NewConcurrencyLimiter(cfg)
        }
    })
    return rdService
}
```

#### 5.1.2 查询执行层
```go
// server/logics/resource_data/resource_data_service.go

func (rds *resourceDataService) Query(ctx context.Context, resource *interfaces.Resource, 
    params *interfaces.ResourceDataQueryParams) ([]map[string]any, int64, error) {
    
    // 获取 Catalog 信息
    catalog, err := rds.cs.GetByID(ctx, resource.CatalogID, true)
    if err != nil {
        return nil, 0, err
    }
    
    // 并发控制
    var release func()
    if rds.cl != nil {
        // 从 Catalog 配置获取并发限制
        maxConcurrentQueries := int64(0)
        if concurrent, exists := catalog.ConnectorCfg["concurrent"]; exists {
            maxConcurrentQueries = int64(concurrent.(float64))
        }
        
        // 获取并发许可
        release, err = rds.cl.Acquire(rate.AcquireParams{
            CatalogID:            resource.CatalogID,
            MaxConcurrentQueries: maxConcurrentQueries,
        })
        if err != nil {
            // 限流错误处理
            if rateErr, ok := err.(*rate.RateLimitError); ok {
                return nil, 0, rest.NewHTTPError(ctx, rateErr.HTTPStatus, 
                    verrors.VegaBackend_Query_ConcurrencyLimitExceeded).
                    WithErrorDetails(rateErr.Message)
            }
            return nil, 0, rest.NewHTTPError(ctx, http.StatusTooManyRequests, 
                verrors.VegaBackend_Query_ConcurrencyLimitExceeded).
                WithErrorDetails("Query concurrency limit exceeded, please retry later")
        }
        defer release() // 查询完成后释放许可
    }
    
    // 原有查询逻辑...
    // ...
}
```

**集成特点**：

| 集成点 | 说明 |
|--------|------|
| **初始化** | 服务启动时创建限流器实例 |
| **配置读取** | 从 `appSetting.RateLimitingSetting` 读取配置 |
| **Catalog 配置** | 从 `catalog.ConnectorCfg["concurrent"]` 动态读取 |
| **错误处理** | 返回 HTTP 503 或 429 错误 |
| **释放时机** | `defer release()` 确保查询完成后释放 |

---

## 6. 可观测性设计

### 6.1 指标 (Metrics)

使用 Prometheus 暴露以下指标：

```go
// metrics/rate_limit_metrics.go

var (
    // 查询总量
    QueryTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "vega_queries_total",
            Help: "Total number of queries",
        },
        []string{"catalog_id", "workload", "status"},
    )
    
    // 当前并发查询数
    QueryConcurrent = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vega_queries_concurrent",
            Help: "Current number of concurrent queries",
        },
        []string{"catalog_id", "workload"},
    )
    
    // 限流命中次数
    RateLimitHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "vega_rate_limit_hits_total",
            Help: "Total rate limit hits",
        },
        []string{"limit_type", "catalog_id", "workload"},
    )
    
    // 等待队列长度
    QueryQueueSize = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vega_query_queue_size",
            Help: "Current query queue size",
        },
        []string{"catalog_id"},
    )
    
    // 队列等待时间
    QueryQueueWaitTime = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "vega_query_queue_wait_seconds",
            Help:    "Time spent waiting in queue",
            Buckets: prometheus.ExponentialBuckets(0.001, 2, 10),
        },
        []string{"catalog_id"},
    )
    
    // 内存使用
    QueryMemoryBytes = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vega_query_memory_bytes",
            Help: "Memory usage per query",
        },
        []string{"query_id"},
    )
    
    // 查询复杂度分数
    QueryComplexityScore = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "vega_query_complexity_score",
            Help:    "Query complexity score",
            Buckets: prometheus.LinearBuckets(0, 10, 20),
        },
        []string{"catalog_id"},
    )
    
    // QPS
    QueryQPS = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vega_query_qps",
            Help: "Queries per second",
        },
        []string{"catalog_id", "tenant_id"},
    )
)
```

### 6.2 日志 (Logging)

限流触发时的结构化日志：

```go
// 限流错误日志
logger.Warn("Rate limit exceeded",
    "limit_type", "concurrency",
    "catalog_id", catalogID,
    "workload", workload,
    "user_id", userID,
    "current_concurrent", currentConcurrent,
    "max_concurrent", maxConcurrent,
    "queue_length", queueLength,
)

// 接近限制告警日志
logger.Info("Approaching rate limit",
    "limit_type", "qps",
    "usage_ratio", 0.85,
    "limit", 1000,
    "current", 850,
)
```

### 6.3 响应头

限流相关的 HTTP 响应头：

```go
// 成功响应
X-RateLimit-Limit: 1000        // 限制值
X-RateLimit-Remaining: 950     // 剩余配额
X-RateLimit-Reset: 1714387200  // 重置时间 (Unix timestamp)

// 限流响应 (HTTP 429)
Retry-After: 45                // 建议重试时间 (秒)
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 0
X-RateLimit-Reset: 1714387200
```

---

## 7. 配置示例

### 7.1 当前配置文件

**实际配置**（基于现有代码）：

```yaml
# server/config/vega-backend-config.yaml

server:
  runMode: release
  httpPort: 13014
  readTimeOut: 600
  writeTimeOut: 600

# 并发限流配置
rateLimiting:
  concurrency:
    enabled: true
    global:
      max_concurrent_queries: 10       # 全局最大并发查询数
```

**Catalog 级配置**（在 Catalog 创建时设置）：

Catalog 的 `ConnectorCfg` JSON 字段中包含并发配置：

```json
{
  "connectorType": "mysql",
  "connectorCfg": {
    "host": "localhost",
    "port": 3306,
    "database": "test",
    "concurrent": 5          // Catalog 级并发限制
  }
}
```

**配置说明**：

| 配置项 | 位置 | 说明 | 默认值 |
|--------|------|------|--------|
| `rateLimiting.concurrency.enabled` | 配置文件 | 是否启用并发控制 | `false` |
| `rateLimiting.concurrency.global.max_concurrent_queries` | 配置文件 | 全局最大并发查询数 | `0`（不限制） |
| `catalog.ConnectorCfg.concurrent` | Catalog 配置 | Catalog 级并发限制 | `0`（不限制） |

**注意事项**：

1. **配置层级**：
   - 全局配置在 `vega-backend-config.yaml` 中
   - Catalog 配置在每个 Catalog 的 `ConnectorCfg` 中

2. **生效条件**：
   - 必须同时启用 `enabled: true` 且 `max_concurrent_queries > 0`
   - Catalog 级限制为 0 时不生效，仅使用全局限制

3. **热更新**：
   - 全局配置需要重启服务生效
   - Catalog 配置在下次查询时自动读取（无需重启）

---

## 8. 实施计划

### 8.1 分阶段实施

**Phase 1 范围：并发控制**（已完成）

| 阶段 | 内容 | 状态 |
|------|------|------|
| **Phase 1** | 并发控制 (全局 + Catalog 两级信号量) | ✅ 已完成 |

**Phase 1 完成情况**：

| 任务 | 文件 | 状态 |
|------|------|------|
| 1. 定义接口 | `server/logics/rate/concurrency_limiter.go` | ✅ 完成 |
| 2. 实现全局信号量 | `server/logics/rate/concurrency_limiter.go` | ✅ 完成 |
| 3. 实现 Catalog 信号量 | `server/logics/rate/concurrency_limiter.go` | ✅ 完成 |
| 4. 集成到查询服务 | `server/logics/resource_data/resource_data_service.go` | ✅ 完成 |
| 5. 配置支持 | `server/config/vega-backend-config.yaml` | ✅ 完成 |
| 6. AT 测试 | `tests/at/rate/concurrency_test.go` | ✅ 通过 |

---

## 9. API 变更

### 9.1 新增错误响应

**HTTP 503 Service Unavailable (并发限制)**:
```json
{
  "error": "concurrency_limit_exceeded",
  "message": "Query concurrency limit exceeded, please retry later",
  "details": {
    "limit_type": "concurrency",
    "limit_subtype": "global",
    "retry_after_seconds": 5
  }
}
```

**HTTP 429 Too Many Requests (Catalog 限制)**:
```json
{
  "error": "concurrency_limit_exceeded",
  "message": "catalog concurrency limit exceeded for catalog=catalog-123, max=20",
  "details": {
    "limit_type": "concurrency",
    "limit_subtype": "catalog",
    "catalog_id": "catalog-123",
    "max_concurrent": 20,
    "retry_after_seconds": 10
  }
}
```

**说明**：
- 当前实现返回 HTTP 503 或 429，取决于错误类型
- `Retry-After` 响应头可选添加，建议重试时间
- 错误详情包含限流类型和当前限制值

### 9.2 响应头（可选）

| Header 名称 | 说明 | 示例值 |
|-------------|------|----------|
| `Retry-After` | 建议重试时间 (秒) | `5` |
| `X-RateLimit-Type` | 限流类型 | `concurrency` |
| `X-RateLimit-Limit` | 当前限制值 | `100` |
| `X-RateLimit-Remaining` | 剩余可用配额 | `50` |

---

## 10. 测试计划

### 10.1 单元测试（已完成）

**当前测试**（基于实际代码）：

| 测试文件 | 测试内容 | 状态 |
|----------|----------|------|
| `server/logics/rate/concurrency_limiter_test.go` | 信号量获取/释放、两级限流 | ✅ 完成 |
| `tests/at/rate/concurrency_test.go` | 集成测试，并发控制验证 | ✅ 完成 |

**测试覆盖**：

1. **全局限流测试**: 验证超过全局限制时拒绝请求
2. **Catalog 限流测试**: 验证单个 Catalog 超限
3. **两级互补测试**: 验证 AND 逻辑（同时生效）
4. **释放函数测试**: 验证 defer release() 正确释放

### 10.2 集成测试（已完成）

**当前测试场景**：

1. **并发限制测试**: 模拟超过全局限制的并发查询
2. **Catalog 限制测试**: 模拟超过 Catalog 限制的并发查询
3. **正常查询测试**: 验证在限制内查询正常执行

### 10.3 压力测试（计划）

**建议测试工具**：

```bash
# 使用 wrk 或 hey 进行并发测试
# 测试并发限制
hey -n 1000 -c 50 http://localhost:13014/api/vega-backend/v1/query

# 验证限流效果
# 观察 503/429 错误比例
```

---

## 11. 风险与缓解措施

### 11.1 技术风险

| 风险 | 影响 | 概率 | 缓解措施 |
|------|------|------|----------|
| 限流器成为性能瓶颈 | 高 | 中 | 使用无锁数据结构，避免同步阻塞 |
| 内存限制不准确 | 中 | 高 | 结合行数限制和结果集大小估算 |
| 配置错误导致误限流 | 高 | 中 | 提供配置验证，默认关闭限流 |
| 分布式部署时限流不一致 | 高 | 高 | Phase 1 先支持单机，后续用 Redis 实现 |

### 11.2 运维风险

| 风险 | 影响 | 缓解措施 |
|------|------|----------|
| 限流阈值设置过低 | 合法查询被拒绝 | 提供监控指标，支持动态调整阈值 |
| 限流阈值设置过高 | 起不到保护作用 | 基于历史数据设置合理的默认值 |
| 配置热更新失败 | 服务重启 | 实现配置版本管理，支持回滚 |

---

## 12. 未来扩展 (Phase 2+)

### 12.1 等待队列支持

**当前问题**：超限直接拒绝，用户体验不佳

**改进方案**：

```go
// 添加等待队列支持
type AcquireParams struct {
    CatalogID            string
    MaxConcurrentQueries int64
    Timeout              time.Duration  // 新增：等待超时时间
    WaitQueueEnabled     bool           // 新增：是否启用等待队列
}

// Acquire 改为阻塞式（可选）
func (cl *concurrencyLimiter) Acquire(ctx context.Context, params AcquireParams) (ReleaseFunc, error) {
    // 如果启用等待队列，使用 Acquire 而非 TryAcquire
    // 支持上下文超时控制
}
```

### 12.2 工作负载优先级隔离

**当前问题**：所有查询同等对待，关键业务可能受低优先级影响

**改进方案**：

```go
// 按工作负载类型分配不同配额
type WorkloadConfig struct {
    Name               string
    MaxConcurrentQueries int64
    Priority           int  // 1-100, 越高越优先
    Timeout            time.Duration
}

// 工作负载分类
type WorkloadType string
const (
    WorkloadCritical WorkloadType = "critical"
    WorkloadHigh     WorkloadType = "high"
    WorkloadNormal   WorkloadType = "normal"
    WorkloadLow      WorkloadType = "low"
)
```

### 12.3 统计指标与监控

**当前问题**：缺乏限流命中统计和监控

**改进方案**：

```go
// 添加统计指标
type ConcurrencyStats struct {
    GlobalCurrentConcurrent int64
    GlobalLimitHits         int64
    CatalogLimitHits        map[string]int64
    AverageWaitTimeMs       int64
}

// Prometheus 指标
var (
    ConcurrencyLimitHits = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "vega_concurrency_limit_hits_total",
            Help: "Total concurrency limit hits",
        },
        []string{"limit_type", "catalog_id"},
    )
    
    CurrentConcurrentQueries = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "vega_concurrent_queries",
            Help: "Current number of concurrent queries",
        },
        []string{"catalog_id"},
    )
)
```

### 12.4 分布式限流

**当前问题**：单机限流，多实例部署时不一致

**改进方案**：

- **Phase 2**: Redis 分布式信号量
- **Phase 3**: 基于 Raft 的一致性限流

### 12.5 动态限流

**当前问题**：配置静态，无法根据系统负载自动调整

**改进方案**：

```go
// 基于系统负载动态调整
type AdaptiveLimiter struct {
    baseLimit          int64
    minLimit           int64
    maxLimit           int64
    cpuThreshold       float64
    memoryThreshold    float64
}

// 定期采样系统负载，自动调整限流阈值
func (al *AdaptiveLimiter) AdjustLimit() {
    cpuUsage := getCPUUsage()
    memoryUsage := getMemoryUsage()
    
    if cpuUsage > al.cpuThreshold || memoryUsage > al.memoryThreshold {
        // 降低限流阈值
        newLimit := int64(float64(al.baseLimit) * 0.8)
        al.updateLimit(newLimit)
    }
}
```

---

## 13. 参考资料

1. **Apache Pinot Workload Isolation**: https://docs.pinot.apache.org/operate-pinot/tuning/workload-query-isolation
2. **Apache Impala Admission Control**: https://impala.apache.org/docs/build/html/topics/impala_admission.html
3. **SingleStore Resource Pools**: https://docs.singlestore.com/cloud/reference/sql-reference/resource-pool-commands/create-resource-pool.md
4. **Uber Global Rate Limiter**: https://www.uber.com/blog/ubers-rate-limiting-system/
5. **Cloudflare Rate Limiting**: https://blog.cloudflare.com/counting-things-a-lot-of-different-ways/
6. **Go Time/Rate Package**: https://pkg.go.dev/golang.org/x/time/rate
7. **Go Sync/Semaphore**: https://pkg.go.dev/golang.org/x/sync/semaphore
8. **Token Bucket vs Leaky Bucket**: https://konghq.com/blog/how-to-design-a-scalable-rate-limiting-algorithm/
9. **Google SRE Book - Load Testing**: https://sre.google/workbook/load-testing/
10. **Netflix Cinnamon - Adaptive Shedding**: https://netflixtechblog.com/adaptive-shedding-at-netflix-82c796f67f9c

---

## 14. 附录

### 14.1 术语表

| 术语 | 定义 |
|------|------|
| **QPS** | Queries Per Second，每秒查询数 |
| **令牌桶** | Token Bucket，允许突发流量的限流算法 |
| **漏桶** | Leaky Bucket，严格限速的流量整形算法 |
| **信号量** | Semaphore，限制并发数量的同步原语 |
| **工作负载** | Workload，具有相同优先级和资源配额的查询 |
| **复杂度分数** | Complexity Score，基于 SQL 特征的查询复杂度指标 |
| **背压** | Backpressure，下游压力向上传播的机制 |

### 14.2 配置参数速查表

| 参数路径 | 默认值 | 描述 |
|----------|--------|------|
| `rate_limiting.enabled` | `true` | 是否启用限流 |
| `rate_limiting.concurrency.global.max_concurrent_queries` | `100` | 全局最大并发查询数 |
| `rate_limiting.qps.global.requests_per_second` | `1000` | 全局 QPS 限制 |
| `rate_limiting.memory.per_query.max_memory_bytes` | `1GB` | 单查询内存限制 |
| `rate_limiting.complexity.max_score` | `100` | 最大查询复杂度分数 |

---

**文档结束**
