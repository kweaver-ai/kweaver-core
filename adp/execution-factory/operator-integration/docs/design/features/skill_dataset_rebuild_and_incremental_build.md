# 🏗️ Design Doc: Skill DataSet 重建与增量构建

> 状态: In Review  
> 负责人: 待确认  
> Reviewers: 待确认  
> 关联 PRD: ../../product/prd/skill_dataset_rebuild_and_incremental_build.md  

---

# 📌 1. 概述（Overview）

## 1.1 背景

- 当前 `operator-integration` 已支持 Skill 在线双写 Vega Skill DataSet。
- 在线双写只能覆盖发布、更新、下架的主路径，无法提供统一的离线补偿能力。
- 当在线双写漏写、漏删、历史存量未回灌时，需要一个可手动触发、可查询、可重试的离线构建能力。
- 离线构建必须以 `t_skill_repository` 为扫描源，并结合 `t_skill_release` 解析当前对外可见态，避免 `editing` 场景误删仍有发布快照的 Skill 文档。
- 任务状态需要落库到 `t_skill_index_build_task`，以支持查询、审计、恢复和周期性后台处理。

## 1.2 目标

- 支持 Skill DataSet 的 `full` 与 `incremental` 构建任务。
- 支持任务创建、列表查询、详情查询、取消、重试。
- 通过 `repository + release` 组合解析可见态，确保：
  - `editing + release存在` 时使用发布快照写入 DataSet。
  - `offline/unpublish/is_deleted=true` 时执行删除补偿。
- 复用现有 `index_sync` 中的 DataSet 初始化、embedding 和文档写入能力。
- 通过进程内 worker 自动推进 `pending` 任务，并提供周期性全量构建、陈旧任务恢复、过期任务清理能力。

## 1.3 非目标（Out of Scope）

- 不替代现有在线双写逻辑。
- 不复用或改造 Vega 通用 DataSet build task。
- 不引入独立消息队列或 Asynq 任务编排。
- 不提供任务暂停能力。
- 不新增 Skill DataSet 多版本索引能力。
- 不实现前台页面，仅提供内部 API。

## 1.4 术语说明（Optional）

| 术语 | 说明 |
|------|------|
| Skill DataSet | Skill 在 Vega/OpenSearch 中的统一检索索引 |
| Full 构建 | 全量扫描 `t_skill_repository` 并逐条对齐 DataSet |
| Incremental 构建 | 基于 `(f_update_time, f_skill_id)` 游标扫描增量变更 |
| 发布快照 | `t_skill_release` 中的当前对外可见 Skill 版本 |
| 任务表 | `t_skill_index_build_task`，用于持久化离线构建任务状态 |
| 分配锁 | Redis 锁，仅用于避免多实例重复抢占同一个 `pending` 任务 |

---

# 🏗️ 2. 整体设计（HLD）

> 本章节关注系统“怎么搭建”，不涉及具体实现细节

---

## 🌍 2.1 系统上下文（C4 - Level 1）

### 参与者
- 用户：运维人员、执行工厂研发
- 外部系统：
  - Vega Backend
  - mf-model-manage
  - mf-model-api
- 第三方基础设施：
  - Redis

### 系统关系

```mermaid
flowchart LR
    U[运维 / 研发] --> OI[operator-integration]

    subgraph EFDB[Execution Factory DB]
        SR[(t_skill_repository)]
        SL[(t_skill_release)]
        ST[(t_skill_index_build_task)]
    end

    subgraph CACHE[协调基础设施]
        R[(Redis Lock)]
    end

    subgraph VEGA[Vega]
        VB[vega-backend]
        DS[(Skill DataSet / OpenSearch)]
    end

    subgraph MODEL[模型服务]
        MM[mf-model-manage]
        MA[mf-model-api]
    end

    OI --> SR
    OI --> SL
    OI --> ST
    OI --> R
    OI --> MM
    OI --> MA
    OI --> VB
    VB --> DS
```

---

## 🧱 2.2 容器架构（C4 - Level 2）

| 容器 | 技术栈 | 职责 |
|------|--------|------|
| Skill Build API | Go + Gin | 创建任务、列表查询、详情查询、取消任务、重试任务 |
| Skill Build Worker | Go + Goroutine | 轮询 `pending` 任务，执行 full/incremental 构建 |
| Skill Registry DB | MySQL | 存储主表、发布快照、构建任务 |
| Redis Lock | Redis | 仅负责分布式任务分配锁 |
| Vega Backend | Go | 写入、更新、删除 Skill DataSet 文档 |
| Model Services | Go/Python | 提供 embedding 维度和向量生成能力 |

### 容器交互

```mermaid
flowchart LR
    API[Skill Build API] --> DB[(MySQL)]
    Worker[Skill Build Worker] --> DB
    Worker --> Lock[(Redis Lock)]
    Worker --> ModelMgr[mf-model-manage]
    Worker --> ModelAPI[mf-model-api]
    Worker --> Vega[vega-backend]
    Vega --> DS[(Skill DataSet)]
```

---

## 🧩 2.3 组件设计（C4 - Level 3）

### Skill Build 组件

| 组件 | 职责 |
|------|------|
| Build Handler | 解析 header、uri、json，调用 service |
| Build Service | 创建任务、查询任务、取消任务、重试任务、驱动构建流程 |
| Build Worker | 每秒轮询一次，恢复陈旧任务、调度周期任务、清理过期任务、推进 `pending` 任务 |
| Pending Task Assigner | 基于 Redis 锁确保同一 `task_id` 只被一个实例抢占 |
| Visible Resolver | 解析主表状态和 release 快照，计算可见态 |
| Document Builder | 根据 repository/release 构造 DataSet 文档 |
| Index Sync | 调用 vega 与模型服务，执行 create/update/delete |
| BuildTask Repo | 维护 `t_skill_index_build_task` |

---

## 🔄 2.4 数据流（Data Flow）

### 主流程

```mermaid
flowchart TD
    A[创建构建任务] --> B[写 t_skill_index_build_task pending]
    B --> C[Worker 轮询 pending 任务]
    C --> D[Redis 分配锁抢占 task_id]
    D --> E[任务置为 running]
    E --> F[读取主表与 release]
    F --> G[解析可见态]
    G -- 是 --> H[构造文档并 upsert DataSet]
    G -- 否 --> I[delete DataSet 文档]
    H --> J[更新任务统计与游标]
    I --> J
    J --> K[任务 completed / pending重试 / failed / canceled]
```

### 子流程

```mermaid
sequenceDiagram
    participant Worker as Build Worker
    participant Task as t_skill_index_build_task
    participant Repo as t_skill_repository
    participant Release as t_skill_release
    participant Index as index_sync
    participant Vega as vega-backend

    Worker->>Task: 查询 running/pending 任务
    Worker->>Task: 抢占成功后置 running
    Worker->>Repo: 按游标分页读取主表
    Worker->>Release: 按 skill_id 查询发布快照
    alt visible
        Worker->>Index: UpsertSkill(payload)
        Index->>Vega: create/update document
    else invisible
        Worker->>Index: DeleteSkill(skill_id)
        Index->>Vega: delete document
    end
    Worker->>Task: 更新统计、游标和状态
```

---

## ⚖️ 2.5 关键设计决策（Design Decisions）

| 决策 | 说明 |
|------|------|
| 扫描源使用 `t_skill_repository` | 能覆盖下架漏删和主表变更场景 |
| `editing` 必须查 `t_skill_release` | 避免误删仍有发布快照的 Skill 文档 |
| 任务状态落 MySQL 任务表 | 需要可查询、可审计、可恢复的任务实体 |
| Redis 仅用于分配锁 | 不承担任务真相源，不承担队列语义 |
| `full` 采用逐条 reconcile | 避免先清空再回灌带来的检索空窗 |
| 增量游标使用 `(f_update_time, f_skill_id)` | 保证稳定排序与断点续跑 |
| 复用现有 `index_sync` | 避免重复实现 embedding 和 Vega 写入逻辑 |
| 任务执行采用进程内 worker | 当前需求只需要轻量后台任务能力，不引入独立队列编排系统 |

---

## 🚀 2.6 部署架构（Deployment）

- 部署环境：K8s
- 拓扑结构：`operator-integration` API 与 worker 运行在同一服务进程内，由 `server/main.go` 启动 worker 后台循环
- 扩展策略：
  - API 可水平扩展
  - worker 通过 MySQL 单运行任务约束和 Redis 分配锁避免同一任务被多实例重复执行

---

## 🔐 2.7 非功能设计

### 性能
- 单批处理数量固定为 `200`
- 增量扫描通过主表排序字段和游标实现稳定断点续跑
- 单条失败不阻断整批任务

### 可用性
- 任务状态持久化到 MySQL
- 在线双写与离线构建解耦
- worker 会周期性恢复陈旧 `running` 任务，避免永久悬挂

### 安全
- 权限模型：沿用内部接口现有 `x-business-domain` 和 `user_id` 校验机制
- 数据保护：不新增敏感字段；沿用现有 Skill 索引字段

### 可观测性
- tracing：支持
- logging：支持
- metrics：支持

---

# 🔧 3. 详细设计（LLD）

> 本章节关注“如何实现”，开发可直接参考

---

## 🌐 3.1 API 设计

### 创建 Skill DataSet 构建任务

**Endpoint:** `POST /api/agent-operator-integration/internal-v1/skills/index/build`

**Headers:**

- `x-business-domain`: required
- `user_id`: optional
- `x-account-id`: optional
- `x-account-type`: optional

**Request:**

```json
{
  "execute_type": "full"
}
```

**Response:**

```json
{
  "task_id": "skill-index-build-xxx",
  "status": "pending",
  "execute_type": "full"
}
```

**HTTP Status:**

- `200`: success
- `400`: invalid header or body
- `409`: running task exists
- `500`: internal error

---

### 查询 Skill DataSet 构建任务列表

**Endpoint:** `GET /api/agent-operator-integration/internal-v1/skills/index/build`

**Headers:**

- `x-business-domain`: required
- `user_id`: optional

**Query:**

- `page`
- `page_size`
- `all`
- `sort_by`: `create_time` / `update_time` / `name`
- `sort_order`: `asc` / `desc`
- `status`: `pending` / `running` / `completed` / `failed` / `canceled`
- `execute_type`: `full` / `incremental`
- `create_user`

**Response:**

```json
{
  "total": 2,
  "page": 1,
  "page_size": 10,
  "total_pages": 1,
  "has_next": false,
  "has_prev": false,
  "data": [
    {
      "task_id": "task-1",
      "status": "pending",
      "execute_type": "full",
      "queue_state": "",
      "retry_count": 0,
      "max_retry": 3
    }
  ]
}
```

**HTTP Status:**

- `200`: success
- `400`: invalid header or query
- `500`: internal error

---

### 查询 Skill DataSet 构建任务详情

**Endpoint:** `GET /api/agent-operator-integration/internal-v1/skills/index/build/{task_id}`

**Headers:**

- `x-business-domain`: required
- `user_id`: required

**Response:**

```json
{
  "task_id": "skill-index-build-xxx",
  "status": "running",
  "execute_type": "incremental",
  "queue_state": "",
  "total_count": 120,
  "success_count": 100,
  "delete_count": 18,
  "failed_count": 2,
  "retry_count": 1,
  "max_retry": 3,
  "cursor_update_time": 1760000000000000000,
  "cursor_skill_id": "skill-123",
  "error_msg": "",
  "create_time": 1760000000000000000,
  "update_time": 1760000001000000000,
  "last_finished_time": 0
}
```

**HTTP Status:**

- `200`: success
- `400`: invalid header or path param
- `404`: task not found
- `500`: internal error

---

### 取消 Skill DataSet 构建任务

**Endpoint:** `POST /api/agent-operator-integration/internal-v1/skills/index/build/{task_id}/cancel`

**Headers:**

- `x-business-domain`: required
- `user_id`: required

**Response:**

```json
{
  "task_id": "skill-index-build-xxx",
  "action": "cancel_task"
}
```

**HTTP Status:**

- `200`: success
- `400`: invalid header or path param
- `404`: task not found
- `409`: task is not cancellable
- `500`: internal error

---

### 重试 Skill DataSet 构建任务

**Endpoint:** `POST /api/agent-operator-integration/internal-v1/skills/index/build/{task_id}/retry`

**Headers:**

- `x-business-domain`: required
- `user_id`: required

**Response:**

```json
{
  "source_task_id": "skill-index-build-xxx",
  "task_id": "skill-index-build-xxx",
  "status": "pending",
  "execute_type": "incremental"
}
```

**HTTP Status:**

- `200`: success
- `400`: invalid header or path param
- `404`: task not found
- `409`: task status is not retryable or another task is running
- `500`: internal error

---

## 🗂️ 3.2 数据模型

### CreateSkillIndexBuildTaskReq

| 字段 | 类型 | 说明 |
|------|------|------|
| `business_domain_id` | string | 来源于 `x-business-domain` |
| `user_id` | string | 来源于 `user_id` |
| `execute_type` | string | `full` 或 `incremental` |

### GetSkillIndexBuildTaskReq

| 字段 | 类型 | 说明 |
|------|------|------|
| `business_domain_id` | string | 来源于 `x-business-domain` |
| `user_id` | string | 来源于 `user_id` |
| `task_id` | string | 路径参数 |

### QuerySkillIndexBuildTaskListReq

| 字段 | 类型 | 说明 |
|------|------|------|
| `business_domain_id` | string | 来源于 `x-business-domain` |
| `user_id` | string | 来源于 `user_id` |
| `status` | string | 可选，`pending/running/completed/failed/canceled` |
| `execute_type` | string | 可选，`full/incremental` |
| `create_user` | string | 可选 |
| `page` | int | 页码 |
| `page_size` | int | 每页数量 |
| `all` | bool | 是否全量返回 |
| `sort_by` | string | `create_time/update_time/name` |
| `sort_order` | string | `asc/desc` |

### CancelSkillIndexBuildTaskReq / RetrySkillIndexBuildTaskReq

| 字段 | 类型 | 说明 |
|------|------|------|
| `business_domain_id` | string | 来源于 `x-business-domain` |
| `user_id` | string | 来源于 `user_id` |
| `task_id` | string | 路径参数 |

### SkillIndexBuildTaskDB

| 字段 | 类型 | 说明 |
|------|------|------|
| `f_id` | bigint | 自增主键 |
| `f_task_id` | varchar(64) | 任务 ID |
| `f_status` | varchar(32) | `pending/running/completed/failed/canceled` |
| `f_execute_type` | varchar(32) | `full/incremental` |
| `f_total_count` | bigint | 已扫描总数 |
| `f_success_count` | bigint | upsert 成功数 |
| `f_delete_count` | bigint | delete 成功数 |
| `f_failed_count` | bigint | 失败数 |
| `f_retry_count` | bigint | 当前任务已执行的内部重试次数 |
| `f_max_retry` | bigint | 最大内部重试次数 |
| `f_cursor_update_time` | bigint | 当前游标时间 |
| `f_cursor_skill_id` | varchar(64) | 当前游标 skill_id |
| `f_error_msg` | text | 任务错误信息 |
| `f_create_user` | varchar(64) | 触发用户 |
| `f_create_time` | bigint | 创建时间 |
| `f_update_time` | bigint | 更新时间 |
| `f_last_finished_time` | bigint | 最近一次结束时间 |

---

## 💾 3.3 存储设计

- 存储类型：MySQL + OpenSearch + Redis
- 数据分布：
  - `t_skill_repository`：Skill 主表
  - `t_skill_release`：发布态快照
  - `t_skill_index_build_task`：离线构建任务
  - Skill DataSet：OpenSearch 中的 Vega DataSet
- 索引设计：
  - `t_skill_index_build_task`：
    - `uk_task_id(f_task_id)`
    - `idx_status_create_time(f_status, f_create_time)`
    - `idx_exec_status_finish_time(f_execute_type, f_status, f_last_finished_time)`
- Redis 只保存 `task_id` 维度的短时分配锁：
  - 锁 key：`eexecution-factory-lock:skill:index:build:assign:{task_id}`
  - 锁过期时间：`5s`

任务表 DDL：

```sql
CREATE TABLE `t_skill_index_build_task` (
  `f_id` bigint NOT NULL AUTO_INCREMENT,
  `f_task_id` varchar(64) NOT NULL,
  `f_status` varchar(32) NOT NULL,
  `f_execute_type` varchar(32) NOT NULL,
  `f_total_count` bigint NOT NULL DEFAULT 0,
  `f_success_count` bigint NOT NULL DEFAULT 0,
  `f_delete_count` bigint NOT NULL DEFAULT 0,
  `f_failed_count` bigint NOT NULL DEFAULT 0,
  `f_retry_count` bigint NOT NULL DEFAULT 0,
  `f_max_retry` bigint NOT NULL DEFAULT 0,
  `f_cursor_update_time` bigint NOT NULL DEFAULT 0,
  `f_cursor_skill_id` varchar(64) NOT NULL DEFAULT '',
  `f_error_msg` text,
  `f_create_user` varchar(64) NOT NULL DEFAULT '',
  `f_create_time` bigint NOT NULL,
  `f_update_time` bigint NOT NULL,
  `f_last_finished_time` bigint NOT NULL DEFAULT 0,
  PRIMARY KEY (`f_id`),
  UNIQUE KEY `uk_task_id` (`f_task_id`),
  KEY `idx_status_create_time` (`f_status`, `f_create_time`),
  KEY `idx_exec_status_finish_time` (`f_execute_type`, `f_status`, `f_last_finished_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

---

## 🔁 3.4 核心流程（详细）

### 创建构建任务流程

1. Handler 绑定 header 和 json。
2. Service 校验 `execute_type`。
3. Service 查询是否已有 `running` 任务。
4. 若不存在运行中任务，创建 `t_skill_index_build_task`，状态为 `pending`。
5. 若为 `incremental`，尝试继承最近一次成功增量任务的游标。
6. 返回 `task_id`，等待后台 worker 轮询执行。

### Worker 主循环

1. 服务启动时初始化 `skillIndexBuildWorker`。
2. `Start()` 后在后台启动 goroutine。
3. worker 每秒执行一次循环：
   - `recoverStaleRunningTask`
   - `schedulePeriodicFullTask`
   - `cleanupExpiredFinishedTasks`
   - `processPendingTask`
4. 若收到 stop 信号，则退出循环。

### Pending 任务抢占流程

1. worker 查询当前可执行任务。
2. 若存在 `pending` 任务，则以 `task_id` 申请 Redis 分配锁。
3. 抢锁成功后再次读任务，确认仍为 `pending`。
4. 将任务状态更新为 `running`。
5. 调用 `runTask(task_id)` 执行构建。

### 列表与详情查询流程

1. Handler 绑定 header、query 或 path。
2. Service 读取 `t_skill_index_build_task`。
3. Service 直接返回任务状态、统计、游标、错误信息。
4. `queue_state` 固定为空字符串，仅用于兼容响应结构，不代表真实队列状态。

### 取消任务流程

1. Handler 绑定 header 和 `task_id`。
2. Service 查询任务表。
3. 仅 `pending` 与 `running` 任务允许取消。
4. 直接将任务状态更新为 `canceled`，`error_msg` 置为 `"task canceled by user"`。
5. 设置 `last_finished_time` 并返回 `action=cancel_task`。

### 重试任务流程

1. Handler 绑定 header 和 `task_id`。
2. Service 查询原任务。
3. 仅 `failed`、`canceled`、`completed` 任务允许重试。
4. 若存在其它 `running` 任务，则返回冲突。
5. 重置原任务为 `pending`，清空统计、错误和游标。
6. 若为 `incremental`，重新继承最近一次成功增量任务的游标。
7. 返回原 `task_id`，`source_task_id` 与 `task_id` 相同。

### Full 构建流程

1. `runTask` 读取任务并确保状态为 `running`。
2. 调用 `EnsureInitialized` 确保 Skill DataSet 已存在。
3. 按 `SelectSkillBuildPage(cursor_update_time, cursor_skill_id, 200)` 分页扫描 `t_skill_repository`。
4. 对每条 Skill 查询 `t_skill_release`。
5. 执行可见态解析。
6. 可见则构建文档并 `upsert`，不可见则 `delete`。
7. 每条处理后更新统计与游标。
8. 批次结束后回写任务表。
9. 扫描完成后更新任务为 `completed`。

### Incremental 构建流程

1. `runTask` 读取任务并确保状态为 `running`。
2. 起始游标来自任务表中的 `cursor_update_time/cursor_skill_id`。
3. 首次增量或重试增量时，游标来源于最近一次成功增量任务。
4. 按 `(f_update_time, f_skill_id)` 过滤条件分页扫描 `t_skill_repository`。
5. 对每条 Skill 查询 `t_skill_release`。
6. 执行可见态解析。
7. 可见则 `upsert`，不可见则 `delete`。
8. 每条处理后更新统计与游标。
9. 批次结束后回写任务表。
10. 扫描完成后更新任务为 `completed`。

### 周期性后台流程

#### 周期性 Full 构建

1. 若 `enable_periodic_full_scan=false`，直接跳过。
2. 若当前存在运行中任务，直接跳过。
3. 查询最近一次成功的 full 任务。
4. 若距离上次 full 完成时间未超过 `periodic_full_scan_interval`，跳过。
5. 否则由系统用户 `system` 创建新的 full 任务。

#### 陈旧 Running 任务恢复

1. 查询当前 `running` 任务。
2. 若任务 `update_time` 距今超过 `30m`，视为陈旧任务。
3. 将任务标记为 `failed`。
4. 设置错误信息为 `"stale running task recovered as failed"`。

#### 过期完成任务清理

1. 若 `enable_task_cleanup=false`，直接跳过。
2. 计算 `now - task_retention` 的截止时间。
3. 删除在截止时间之前已结束的任务记录。

---

## 🧠 3.5 关键逻辑设计

### 可见态解析逻辑
- 若 `is_deleted=true`，直接不可见。
- 若 `status=offline` 或 `status=unpublish`，直接不可见。
- 若 `status=editing`：
  - release 存在：可见，内容来源为 release。
  - release 不存在：不可见。
- 若 `status=published`：
  - release 存在：可见，内容来源为 release。
  - release 不存在：可见，内容来源为 repository 兜底。

### 增量游标逻辑
- 主游标：`f_update_time`
- 次游标：`f_skill_id`
- 增量条件：
  - `f_update_time > cursor_update_time`
  - 或 `f_update_time = cursor_update_time AND f_skill_id > cursor_skill_id`

### 文档构建逻辑
- `_id = skill_id`
- 若来源为 release，则使用 release 的名称、描述、版本等字段
- 若来源为 repository，则使用主表字段
- embedding 输入统一由最终写入的 `name + description` 生成

### 任务推进逻辑
- `pending -> running`
- `running -> completed`
- `running -> failed`
- `pending/running -> canceled`
- `failed/canceled/completed -> pending`（重试时复用原任务）

### 失败与内部重试逻辑
- 当任务级异常发生时，优先进入 `failTask`
- 若 `retry_count < max_retry`：
  - 任务重置为 `pending`
  - 清空统计、错误和结束时间
  - 游标重置为 0 或最近一次成功增量游标
  - `retry_count + 1`
- 否则：
  - 任务置为 `failed`
  - `error_msg` 记录最终错误

---

## ❗ 3.6 错误处理

- 参数校验失败：返回 `400`
- 运行中任务冲突：返回 `409`
- 任务不存在：返回 `404`
- 任务状态不允许取消或重试：返回 `409`
- release 查询失败：当前 Skill 记失败，不执行 delete
- embedding 调用失败：当前 Skill 记失败
- vega 写入失败：当前 Skill 记失败
- 批次级异常：
  - 若仍有内部重试次数，则任务重置为 `pending`
  - 否则任务置为 `failed`，写入 `error_msg`

---

## ⚙️ 3.7 配置设计

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `skill_index_build.enable_periodic_full_scan` | `false` | 是否开启周期性全量构建 |
| `skill_index_build.periodic_full_scan_interval` | `168h` | 周期性全量构建间隔 |
| `skill_index_build.enable_task_cleanup` | `false` | 是否开启过期任务清理 |
| `skill_index_build.task_retention` | `720h` | 已结束任务保留时长 |
| `skill_index_build_batch_size` | `200` | 单批处理数量，当前为代码常量 |
| `skill_index_build_assign_lock_expiry` | `5s` | 任务分配锁过期时间，当前为代码常量 |
| `skill_index_build_running_timeout` | `30m` | 陈旧 running 任务超时时间，当前为代码常量 |

---

## 📊 3.8 可观测性实现

- tracing：
  - 在创建任务、worker 轮询、任务执行、单批扫描、单条写入/删除处打点
  - span 包含 `task_id`、`execute_type`、`skill_id`

- metrics：
  - `skill_index_build_task_created_total`
  - `skill_index_build_task_completed_total`
  - `skill_index_build_task_failed_total`
  - `skill_index_build_task_canceled_total`
  - `skill_index_build_record_upsert_total`
  - `skill_index_build_record_delete_total`
  - `skill_index_build_record_failed_total`

- logging：
  - 结构化日志字段包含：
    - `task_id`
    - `execute_type`
    - `skill_id`
    - `repository_status`
    - `visible_source`
    - `cursor_update_time`
    - `cursor_skill_id`

---

# ⚠️ 4. 风险与权衡（Risks & Trade-offs）

| 风险 | 影响 | 解决方案 |
|------|------|----------|
| `t_skill_release` 查询异常 | `editing` 场景误判风险 | 查询失败记单条失败，不按无 release delete |
| `mf-model-api` 抖动 | 单条 upsert 失败 | 记录失败数，支持后续增量补偿 |
| `vega-backend` 抖动 | DataSet 写入或删除失败 | 记录失败数，任务可重试 |
| 全量任务耗时长 | 对排障时效有影响 | 采用任务化、分批执行、任务状态可查询 |
| 进程内 worker 随实例启动 | 无独立队列削峰能力 | 当前需求以轻量补偿为主，通过单任务互斥控制复杂度 |
| Redis 不可用 | 多实例任务抢占保护失效 | 锁获取失败时放弃本轮抢占，由后续轮询继续尝试 |

---

# 🧪 5. 测试策略（Testing Strategy）

- 单元测试：
  - 可见态解析
  - 增量游标过滤
  - 任务状态推进
  - 文档构建来源切换
  - 任务取消与重试语义
  - 陈旧任务恢复
  - 周期性 full 创建与过期任务清理

- 集成测试：
  - 创建 `full` 任务
  - 创建 `incremental` 任务
  - 冲突任务返回 `409`
  - 查询任务返回完整字段
  - worker 轮询 `pending` 任务并调用 `index_sync`

- 回归测试：
  - `queue_state` 保持空字符串
  - 重试复用原 `task_id`
  - `editing + release存在` 走 release upsert
  - `editing + release不存在` 走 delete

---

# 📅 6. 发布与回滚（Release Plan）

### 发布步骤
1. 执行数据库变更，创建 `t_skill_index_build_task`
2. 发布 `operator-integration` 新版本
3. 验证内部接口可创建和查询任务
4. 确认 worker 随服务启动并进入轮询
5. 触发一次测试环境 `full` 构建
6. 验证任务状态、DataSet 变更和日志指标

### 回滚方案
- 若接口或 worker 异常，回滚 `operator-integration` 版本
- 保留任务表不删，不影响旧逻辑
- 由于在线双写链路未替换，回滚不会影响已有在线同步能力

---

# 🔗 7. 附录（Appendix）

## 相关文档
- PRD: ../../product/prd/skill_dataset_rebuild_and_incremental_build.md
- 其他设计: ../../design/features/skill_write_vega_dataset.md

## 参考资料
