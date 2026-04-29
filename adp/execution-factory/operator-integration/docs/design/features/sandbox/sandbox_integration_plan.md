# Sandbox 沙箱对接方案说明文档 (Feature-796721)

## 1. 概述
本方案旨在实现 `operator-integration` 与沙箱环境（`kweaver-ai/sandbox`）的高效对接。沙箱环境提供安全的 Python 代码执行能力，本系统通过会话池（Session Pool）管理、统一代理（Proxy）以及确定性 ID 策略，实现了资源最小化占用、高并发支持以及故障自动修复。

## 2. 架构设计

### 2.1 核心组件
- **[session_pool.go](/adp/execution-factory/operator-integration/server/logics/sandbox/session_pool.go)**: 会话池逻辑，负责生命周期管理、任务分配及扩缩容。
- **[sandbox_control_plane.go](/adp/execution-factory/operator-integration/server/drivenadapters/sandbox_control_plane.go)**: 驱动适配器，封装了沙箱 REST API 的调用（创建、查询、删除、执行）。
- **[proxy.go](/adp/execution-factory/operator-integration/server/driveradapters/common/proxy.go)**: 入站代理，接收前端执行请求，调用元数据服务获取代码并提交给会话池。

### 2.2 确定性 ID 策略
- 使用固定前缀 `sess_aoi_` 配合索引（如 `sess_aoi_0`, `sess_aoi_1`）作为 Session ID。
- **优势**: 方便在系统重启或异常后进行状态同步和资源回收。

## 3. 会话池生命周期管理

### 3.1 任务分配 (堆叠策略)
- **逻辑**: 优先选择当前 `RunningTasks` 负载最高但未满载的 Session 进行任务分配。
- **目的**: 尽可能填满一个 Session 后再启用下一个，确保空闲 Session 能被及时缩容。

### 3.2 预热与空闲维持 (ActiveSessions)
- **初始化恢复**: 服务启动时，会首先检查并恢复（Recover）所有状态为 `Running` 的现有会话，将其纳入池管理。
- **启动预热**: 初始化恢复后，若可用会话数不足 `activeSessions`，则自动创建补足。
- **水位维持**: 后台任务每分钟检查一次，若健康 Session 数少于 `activeSessions` 则自动创建（预热），若多余且空闲则进行缩容。

### 3.3 健康检查与修复
- 后台任务定期调用沙箱 `QuerySession` 接口。
- **自动修复**: 发现状态非 `Running` 的 Session 立即从池中移除并清理远程资源，下次请求时自动重建。

### 3.4 容错重试机制
- 当 `sess_aoi_0` 创建失败时，系统会自动尝试 `sess_aoi_1`，直到遍历所有可用槽位。
- 只有当所有槽位均尝试失败后，才会返回错误。

## 4. 接口定义与对接

### 4.1 核心 API 交互
| 接口功能 | 接口路径 | 说明 |
| :--- | :--- | :--- |
| 创建会话 | `POST /api/v1/sessions` | 根据确定性 ID 创建沙箱环境 |
| 查询会话 | `GET /api/v1/sessions/{id}` | 获取状态及资源详情 |
| 删除会话 | `DELETE /api/v1/sessions/{id}` | 回收资源 |
| 同步执行 | `POST /api/v1/executions/sessions/{id}/execute-sync` | 提交代码（支持动态依赖）并获取 Stdout/Stderr/Result |

### 4.2 配置参数 (SandboxControlPlaneConfig)
- `max_sessions`: 最大物理会话数（槽位上限）。
- `max_concurrent_tasks`: 单个会话支持的并发任务上限。
- `active_sessions`: 即使无任务也需保持活跃的最小会话数。

## 5. 异常处理
- **池满载**: 若 `max_sessions * max_concurrent_tasks` 均满载，返回 `503 Service Unavailable`。
- **执行失败**: 若请求沙箱发生网络错误，该 Session 会被标记为失效并异步移除。
- **超时控制**: 默认执行超时 30 秒，预热超时 40 秒。

## 6. 依赖管理

系统支持在沙箱环境中动态管理和使用 Python 依赖库。

### 6.1 动态依赖安装
- 在执行代码请求 (`ExecuteCodeReq`) 中，支持携带 `dependencies` 列表和 `dependencies_url`。
- **执行流程**:
    1. 检查请求中是否包含依赖信息。
    2. 如果存在，调用沙箱接口安装指定依赖。
    3. 依赖安装成功后，继续执行用户代码。
    4. 若安装失败，记录错误日志并尝试继续执行（或根据策略报错）。

### 6.2 预置依赖
系统默认预置了常用的数据科学和工具库，包括但不限于：
- `requests==2.28.1`
- `pandas==1.5.2`
- `numpy==1.23.4`
- `matplotlib==3.7.1`
- `mock==0.0.1`

