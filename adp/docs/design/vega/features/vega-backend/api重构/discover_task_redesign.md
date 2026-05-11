# DiscoverTask API 收敛 技术设计文档

> **状态**：草案
> **负责人**：@待补
> **日期**：2026-05-07
> **相关 Ticket**：待补

---

## 1. 背景与目标

### 背景

vega-backend 当前 DiscoverTask 的 HTTP 端点设计有以下问题：

- **`by-id` / `by-schedule` 前缀冗余**：`GET /discover-tasks/by-id/{id}` 和 `GET /discover-tasks/by-schedule/{scheduleId}` 把"按什么字段查"写进 path，与 RESTful 风格冲突。`by-id` 应当退化为 `/discover-tasks/{id}`，`by-schedule` 应当退化为 list 的过滤参数 `?schedule_id=`。
- **缺 DELETE 端点**：DiscoverTask 是任务执行历史，会持续累积。当前没有清理入口，运维只能直接动数据库。`DiscoverTaskService` 也没有 `Delete` 方法。
- **list 过滤维度不全**：`DiscoverTaskQueryParams` 已支持 `catalog_id` / `status` / `trigger_type`，但缺 `schedule_id`——即"看某个定时任务历史触发了哪些 DiscoverTask"。这恰是 `by-schedule` 端点试图回答的问题，应当通过 list filter 表达。

### 目标

DiscoverTask 是**执行审计记录**，不是用户可控的执行实例（区别于 BuildTask）：

- 状态机（`pending → running → completed/failed`）由 worker 单向推进
- 创建路径只有两条：手动触发（`POST /catalogs/{id}/discover` 动作端点）或定时触发（worker 内部从 `ScheduledDiscoverTask.ExecuteTask` 调起）
- 用户对 task 唯一能做的事：看历史（list/get）、清理（delete）

因此对外端点收敛到**仅 list / get / delete 三类**：

1. 移除 `by-id` / `by-schedule` 路径前缀，详情走 `/discover-tasks/{id}`，按 schedule 过滤走 list filter。
2. 新增 `DELETE /discover-tasks/{ids}`，对齐 BuildTask 的整体事务批量删除语义。
3. `DiscoverTaskQueryParams` 新增 `ScheduleId` 字段，service 层 `GetByScheduledID` 退化为 list 一种特例（接口可保留也可删，详见 §3.4）。

### 非目标

- **不改触发动作**：`POST /catalogs/{id}/discover` 保留。规则允许动作类端点保留嵌套形态，且 `/discover` 作为动词比 `/discover-tasks` 子资源更自然。
- **不改状态机**：`pending / running / completed / failed` 不变；状态转移由 worker 内部驱动。
- **不引入 task 控制语义**：不加 cancel / retry / restart 端点。重做一次 discover 的方式仍是再次 `POST /catalogs/{id}/discover`。
- **不动 ScheduledDiscoverTask**：定时配置的重构在另一份 design doc 处理。

## 2. 方案概览

### 2.1 端点变化总览

#### 新增

```
DELETE /api/vega-backend/v1/discover-tasks/{ids}                # 整体事务，?ignore_missing=true 可放宽
```

#### 调整

```
GET    /api/vega-backend/v1/discover-tasks/{id}                 # 替代 /by-id/{id}
GET    /api/vega-backend/v1/discover-tasks                      # 新增过滤参数 ?schedule_id=
                                                                  # 替代 /by-schedule/{scheduleId} 的功能
```

#### 弃用

```
- GET    /discover-tasks/by-id/{id}
- GET    /discover-tasks/by-schedule/{scheduleId}
```

#### 保留不变

```
POST   /api/vega-backend/v1/catalogs/{id}/discover              # 手动触发，对 catalog 的动作
GET    /api/vega-backend/v1/discover-tasks                      # 列表（已存在，仅扩展 filter）
```

### 2.2 资源关系图

```mermaid
graph LR
    Catalog[Catalog] -->|POST /catalogs/{id}/discover 手动触发| DiscoverTask
    Schedule[ScheduledDiscoverTask] -->|worker 定时触发| DiscoverTask[DiscoverTask]
    DiscoverTask -->|状态机| StateMachine[pending → running → completed/failed]

    User[用户] -->|GET 列表/详情| DiscoverTask
    User -->|DELETE 清理历史| DiscoverTask
```

DiscoverTask 是 Catalog 的孩子（每条 task 关联 1 个 catalog），但实体本身已是独立资源（全局唯一 ID、独立状态机、独立查询）。HTTP 层将其暴露为顶层资源 `/discover-tasks`，仅提供只读 + 清理三类操作。

## 3. 详细设计

### 3.1 端点边界

#### GET `/discover-tasks`

支持过滤：

| 参数 | 说明 |
|---|---|
| `catalog_id` | 按 catalog 归属（已支持） |
| `schedule_id` | **新增**，按 ScheduledDiscoverTask 归属过滤；本字段对 trigger_type=manual 的 task 恒为空 |
| `status` | `pending` / `running` / `completed` / `failed` |
| `trigger_type` | `manual` / `scheduled` |
| `offset / limit / sort / direction` | 分页与排序 |

#### GET `/discover-tasks/{id}`

- 200：DiscoverTask 实体
- 404 `VegaBackend.DiscoverTask.NotFound`：不存在

#### DELETE `/discover-tasks/{ids}`

`{ids}` 为逗号分隔列表，单条即长度为 1 的退化情形。**整体事务语义**：所有 id 通过预校验后才进入删除阶段；任一预校验失败则整批不删。

预校验顺序：

| 条件 | HTTP | errcode | error_details |
|---|---|---|---|
| 任一 id 处于 `pending` / `running` | 409 | `VegaBackend.DiscoverTask.HasRunningExecution` | `{ running_ids: [...] }` |
| 任一 id 不存在（且 `ignore_missing` 未启用） | 404 | `VegaBackend.DiscoverTask.NotFound` | `{ missing_ids: [...] }` |
| 全部通过 | 204 | — | — |

可选 query 参数：

- `ignore_missing=true`：忽略不存在的 id，其它 id 正常删。**不影响** `pending/running` 拦截——状态拦截不可绕过，避免删除掉 worker 正在写入的 task 留下孤儿数据。

`pending/running` 检查**先于**缺失检查；要删活跃 task 必须等其完成（或失败）。

### 3.2 错误码

新增：

| HTTP | errcode | 触发场景 |
|---|---|---|
| 404 | `VegaBackend.DiscoverTask.NotFound` | GET / DELETE 找不到 task |
| 409 | `VegaBackend.DiscoverTask.HasRunningExecution` | DELETE 时 task 处于 pending/running |
| 500 | `VegaBackend.DiscoverTask.InternalError.DeleteFailed` | 删除失败 |
| 500 | `VegaBackend.DiscoverTask.InternalError.GetFailed` | 查询失败（如已存在沿用） |

中英文 i18n 同步补充。检查 `errors/task.go` 现有项是否已涵盖（`DiscoverTask.NotFound` 可能已存在）。

### 3.3 数据模型变更

**无表结构变更。** 仅扩展查询参数：

```go
type DiscoverTaskQueryParams struct {
    PaginationQueryParams
    CatalogID   string
    ScheduleId  string  // 新增
    Status      string
    TriggerType string
}
```

`drivenadapters/discover_task` 的 SQL where 子句新增 `f_scheduled_id = ?` 分支。

### 3.4 service 层接口调整

**移除 `GetByScheduledID`**：

调研显示该方法的调用方只有即将删除的旧 handler（`GetDiscoverTaskByScheduleId` / `...ByIn`），无 worker / scheduler 在用。移除范围：

- `interfaces/discover_task_service.go`：从 `DiscoverTaskService` 接口删除 `GetByScheduledID`
- `interfaces/discover_task_access.go`：从 `DiscoverTaskAccess` 接口删除 `GetByScheduledID`
- `logics/discover_task/discover_task_service.go`：删除实现
- `drivenadapters/discover_task/discover_task_access.go`：删除实现 + 对应 SQL

按 schedule 过滤的能力由 `List(ctx, {ScheduleId: ...})` 完整替代。

**新增** `DiscoverTaskService.Delete`：

```go
// Delete atomically deletes discover tasks by IDs.
// Pre-validates: any missing id returns 404 unless ignoreMissing=true;
// any pending/running id returns 409 (cannot be skipped).
Delete(ctx context.Context, ids []string, ignoreMissing bool) error
```

`drivenadapters/discover_task` 数据访问层补 `Delete(ctx, id) error`。

### 3.5 与 BuildTask 的对称性

DELETE 形态和 BuildTask 完全对齐：

| 操作 | DiscoverTask | BuildTask |
|---|---|---|
| 列表 | `GET /discover-tasks?...` | `GET /build-tasks?...` |
| 详情 | `GET /discover-tasks/{id}` | `GET /build-tasks/{id}` |
| 删除 | `DELETE /discover-tasks/{ids}?ignore_missing=` | `DELETE /build-tasks/{ids}?ignore_missing=` |
| 创建 | （动作端点 `POST /catalogs/{id}/discover`） | `POST /build-tasks` |
| 启停 | （无外部启停） | `POST /build-tasks/{id}/start\|stop` |

差异来自业务语义：DiscoverTask 是审计记录（无外部启停 / 创建走动作端点），BuildTask 是可控执行实例（显式 CRUD + start/stop）。

## 4. 边界情况与风险

| 类型 | 描述 | 应对 |
|---|---|---|
| 并发 | 两个请求同时 DELETE 同一 task | 第二次 DELETE 在预校验阶段会发现 task 不存在；若未启用 `ignore_missing` 返回 404 |
| 状态竞态 | DELETE 通过预校验时 status=completed，进入删除阶段时 worker 把 status 改回 running | 不可能——worker 不会回退状态机。completed/failed 是终态 |
| 兼容性 | 弃用旧 `/by-id` / `/by-schedule` 路由破坏现有客户端 | 与外部依赖方对齐时间窗，发布时通知 + CHANGELOG 标 BREAKING |
| 数据迁移 | 无 schema 变更 | — |
| 性能 | 新增 `f_scheduled_id` 过滤 | 检查现有索引，必要时补 composite |
| 审计 | DELETE 是历史清理操作，应当审计 | 沿用 `audit.NewWarnLog`，与 BuildTask 删除同套路 |

## 5. 替代方案

### 方案 A：保留 `by-id` / `by-schedule` 端点

只新增 DELETE，不动既有读路径。

**优点**：兼容性最好。
**缺点**：与刚定的端点设计规则（[vega-backend/CLAUDE.md](../../../../vega/vega-backend/CLAUDE.md)）冲突——禁止 `by-X` 这种把"按什么查"写进 path 的形态。仓库内的 RESTful 风格分裂。

**结论**：放弃。规则刚定就违反，会让规则失效。

### 方案 B：一并把 `POST /catalogs/{id}/discover` 改为 `POST /discover-tasks` body 含 `catalog_id`

形态与 BuildTask 完全对齐。

**优点**：所有 task 类资源使用同一种创建形态。
**缺点**：discover 在语义上是**对 catalog 的动作**，而非"创建一条 task"——task 是动作的副产物（审计记录）。把 RPC 风格强行 RESTful 化，反而失真。规则也允许动作端点保留嵌套形态。

**结论**：放弃。RPC vs REST 在动作类端点上 RPC 更自然。

### 最终方案：list / get / delete 三件套 + 触发动作端点保留

详见第 2、3 节。

## 6. 任务拆分

按 [adp/CLAUDE.md](../../../../../CLAUDE.md) 规则 5 拆分。每批前按规则 1 描述细节方案 + 验收清单 + 失败条件待批准。

- [ ] **批 1：错误码 + i18n**
  - `errors/task.go`：新增 `VegaBackend.DiscoverTask.NotFound` / `HasRunningExecution` / `InternalError.DeleteFailed`（如已存在则跳过）
  - `locale/task.{en-US,zh-CN}.toml`：补对应条目
  - 不动调用点

- [ ] **批 2：service + 数据访问层 Delete + ScheduleId filter + 移除 GetByScheduledID**
  - `interfaces/discover_task.go`：`DiscoverTaskQueryParams` 新增 `ScheduleId string`
  - `interfaces/discover_task_service.go`：`DiscoverTaskService` 删除 `GetByScheduledID`，新增 `Delete(ctx, ids, ignoreMissing) error`
  - `interfaces/discover_task_access.go`：删除 `GetByScheduledID`，新增 `Delete(ctx, id) error`（如不存在）
  - `drivenadapters/discover_task/`：实现数据层 Delete + list where 子句加 `f_scheduled_id = ?`，删除 `GetByScheduledID` 实现
  - `logics/discover_task/`：实现 service 层 Delete（含预校验），删除 `GetByScheduledID` 实现
  - 同步更新对应 mock 文件

- [ ] **批 3：handler + router**
  - 新建 `driveradapters/discover_task_handler.go`（如尚未独立成文件，从 catalog_handler 拆出 `DiscoverTask` 相关三个旧 handler）
  - 新增 `DeleteDiscoverTasksByEx/ByIn/deleteDiscoverTasks`
  - 重写 `GetDiscoverTask` 从 `c.Param("id")` 直接取（不再走 `by-id` 路径段）
  - list handler 新增 `schedule_id` query 解析
  - `router.go`：删 `/by-id/:id` / `/by-schedule/:scheduleId` 路由，新增 `GET /:id` / `DELETE /:ids`
  - 旧 `GetDiscoverTaskByScheduleId` handler 删除

- [ ] **批 4：OpenAPI yaml**
  - `adp/docs/api/vega/vega-backend-api/discover-task.yaml`

## 7. 已决定事项

1. **DELETE 是否拒绝活跃 task**：**拒绝**。pending/running 返回 409 `HasRunningExecution`，对齐 BuildTask。`ignore_missing` 不能绕过此拦截。

2. **`POST /catalogs/{id}/discover` 是否一并重命名**：**不改**。是动作端点，RPC 形态比 RESTful 子资源形态更自然。

3. **是否新增 retrigger / cancel 等控制端点**：**不加**。DiscoverTask 是审计记录，不是可控实例。重做一次走 `POST /catalogs/{id}/discover`，没有外部 cancel 概念。

4. **`GetByScheduledID` service 方法**：**移除**。调用方只有即将删除的旧 handler，无内部依赖。按 schedule 过滤的能力由 `List(ctx, {ScheduleId: ...})` 替代。

5. **删除批量响应**：**204 + 无 body**，与 BuildTask 一致。任一预校验失败整批不删并返回 4xx + `error_details` 含 `running_ids` 或 `missing_ids`。

6. **路径形态**：`DELETE /discover-tasks/{ids}` 走 path 逗号分隔，与 `/build-tasks/{ids}` / `/catalogs/{ids}` / `/resources/{ids}` 对齐，不用 `?ids=` query 形态。

## 8. Follow-up（不在本次重构范围）

### 8.1 `ScheduledDiscoverTask` 实体重命名

**问题**：`ScheduledDiscoverTask` 是过去分词 + 名词 + Task 的复合命名，在 OOP / REST 命名规范上属于 anti-pattern：

- "Scheduled" 是形容词，类名应该用名词
- 它的真实语义是"DiscoverTask 的 schedule 配置"——是触发 task 的**配方**，不是另一种 task；名字里带 "Task" 反而误导
- FK 字段 `f_scheduled_id` / API 字段 `schedule_id` 的命名拉锯（见本文档 §3.4 / §6 历史）正是源于这个错误的实体名——只要实体名不规范，FK 简写就没有"对"的写法

**建议方向**：把 `ScheduledDiscoverTask` 重命名为 **`DiscoverSchedule`**：

- 名词，符合规范
- FK 字段在 `DiscoverTask` 里自然写为 `schedule_id`（指向 `DiscoverSchedule`），不再是简写
- DB 列名可同步迁移 `f_scheduled_id` → `f_schedule_id`
- 路由路径 `/catalogs/{id}/scheduled-discover` → `/discover-schedules`（顶层化），与本文档主体重构同构

**为什么本次不做**：

1. 本次 PR 已经聚焦 DiscoverTask 端点收敛，scope 已大
2. ScheduledDiscoverTask 的端点本身也有双主键 path / 命名不复数 / start/stop 名实不符等问题，需要单独一份 design doc 做完整重构
3. DB 列名迁移涉及 schema migration，需要协调发布窗口

**入口条件**：建议在 ScheduledDiscoverTask 端点重构时一并处理；等到那时本文档的 §3.4 命名决策（`schedule_id`）会自然对齐到正确状态。
