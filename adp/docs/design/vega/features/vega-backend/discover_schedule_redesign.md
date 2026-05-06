# DiscoverSchedule 顶层资源化与 API 重构 技术设计文档

> **状态**：草案
> **负责人**：@待补
> **日期**：2026-04-29
> **相关 Ticket**：待补

---

## 1. 背景与目标

### 背景

vega-backend 当前 `scheduled-discover`（catalog 资源发现的定时调度）模块在 HTTP 层与底层数据模型之间存在明显错位：

- **数据模型**：`ScheduledDiscoverTask` 拥有全局唯一主键 `id`，service 接口（`GetByID/Update/Delete/Enable/Disable`）只用 `id` 定位，独立 DB 表，与 catalog 是 N:1 关系。
- **HTTP 层**：路由强制挂在 `/catalogs/{id}/scheduled-discover/{tid}/...` 下，要求双主键 path，且把"列出全部调度"端点（实质跨 catalog）`GET /catalogs/scheduled-discover` 也挂在 catalog 子路径，URI 形态与功能错位。
- **词典撕裂**：service 层用 `Enable / Disable`，handler 包装成 `start / stop`，与仓库内其它资源（ConnectorType 用 enable/disable）不一致。
- **缺端点**：service 提供了 `Delete` / `GetByID`，HTTP 没暴露。

此外存在两处规范化问题：

- `DiscoverTask.scheduled_id` 命名不顺（过去分词当前缀），与新顶层资源 `DiscoverSchedule` 的命名约定不一致。
- `discover-tasks` 路由用了 `/by-id/{id}` `/by-schedule/{sid}` 的非 RESTful 形态，应规范化为 `/{id}` + query string filter。

### 目标

1. 把"发现调度配置"建模为顶层独立资源 `DiscoverSchedule`，URI、service、DB 三层主键对齐。
2. 状态切换统一为动作端点 `enable / disable`，与 ConnectorType 风格一致。
3. 重构 `discover-tasks` 路由为标准 REST 形态。
4. 修正 `scheduled_id` 字段命名为 `schedule_id`，DB / Go / JSON 同步迁移。

### 非目标

- 不抽象成"通用 Schedule + target_type"模型（业务确认非 discover 的周期任务不会出现，其它任务管理走自己的属性）。
- 不引入 schedule 模板（多 catalog 复用同一调度）能力。
- 不重做 cron 调度引擎（仍用 `worker.Scheduler` + `robfig/cron/v3`）。
- 不改 catalog / connector-type 等其它资源的 API 形态。
- 不改一次性发现入口 `POST /catalogs/{cid}/discover` 的路径或行为。

## 2. 方案概览

### 2.1 资源关系图

```mermaid
graph LR
    Catalog[Catalog] -->|0..N| Schedule[DiscoverSchedule]
    Schedule -->|0..N| Task[DiscoverTask]
    Catalog -->|0..N 手动触发| Task

    User[用户] -->|手动 POST /catalogs/cid/discover<br/>带 strategies| Catalog
    Cron[Cron 调度器] -->|按 cron 触发<br/>schedule_id 落入 task| Task
```

`DiscoverSchedule` 是纯**配置**资源（cron + strategies + 起止时间 + 启停状态），不直接承载执行。`DiscoverTask` 是执行实例集合，由两个入口产出：

- **手动触发**：`POST /catalogs/{cid}/discover`，task.schedule_id 为空。
- **cron 自动触发**：调度器到点产出，task.schedule_id 指向触发它的 schedule。

### 2.2 端点变化总览

#### 新增

```
POST   /api/vega-backend/v1/discover-schedules                  # 创建
GET    /api/vega-backend/v1/discover-schedules                  # ?catalog_id= ?enabled= 过滤
GET    /api/vega-backend/v1/discover-schedules/{sid}            # 详情
PUT    /api/vega-backend/v1/discover-schedules/{sid}            # 严格全量替换；不修改 enabled
DELETE /api/vega-backend/v1/discover-schedules/{sid}            # 删除调度，不影响已产生的 DiscoverTask
POST   /api/vega-backend/v1/discover-schedules/{sid}/enable     # 启用
POST   /api/vega-backend/v1/discover-schedules/{sid}/disable    # 停用

GET    /api/vega-backend/v1/catalogs/{cid}/discover-schedules   # 便利只读视图，等价 ?catalog_id={cid}
```

#### 变更（路径或字段名）

```
GET    /discover-tasks/by-id/{id}             →  GET /discover-tasks/{id}
GET    /discover-tasks/by-schedule/{sid}      →  GET /discover-tasks?schedule_id={sid}

字段重命名:
DiscoverTask.scheduled_id (JSON / Go / DB)    →  schedule_id
DiscoverTaskQueryParams 新增 ScheduleID 字段
```

#### 弃用

```
- POST   /catalogs/{cid}/scheduled-discover
- POST   /catalogs/{cid}/scheduled-discover/{tid}/start
- POST   /catalogs/{cid}/scheduled-discover/{tid}/stop
- PUT    /catalogs/{cid}/scheduled-discover/{tid}
- GET    /catalogs/scheduled-discover
```

#### 保留不变

```
POST   /api/vega-backend/v1/catalogs/{cid}/discover    # 手动触发，唯一手动入口
```

## 3. 详细设计

### 3.1 核心约束

#### 3.1.1 PUT 严格语义

`PUT /discover-schedules/{sid}` 是**全量替换**，但只允许变更"配置"字段。下表的"只读字段"如果在 body 中携带且与当前值不一致，返回 409。

| 字段 | 类型 | 可改 |
|---|---|---|
| `id` | string | 只读，与 path 不一致 → 409 `*.IdMismatch` |
| `catalog_id` | string | 只读（schedule 创建后归属不可改）→ 409 `*.CatalogMismatch` |
| `cron_expr` | string | 可改 |
| `strategies` | string[] | 可改 |
| `start_time` | int64 | 可改 |
| `end_time` | int64 | 可改 |
| `enabled` | bool | 只读（状态切换走 enable/disable） → 409 `*.EnabledFieldNotAllowed` |
| `last_run` / `next_run` | int64 | 只读（系统维护），携带时静默忽略 |
| `creator` / `create_time` / `updater` / `update_time` | - | 只读（系统维护），携带时静默忽略 |

#### 3.1.2 状态切换的唯一入口

`enabled` 字段在响应里只读返回；变更走 `POST .../enable` `POST .../disable`。
- 端点幂等：对已 enable 的资源再 enable 不报错（也可选返回 409 `AlreadyEnabled`，需团队拍板）。
- 与 ConnectorType 风格一致，避免仓库内启停命名词典分裂。

#### 3.1.3 DELETE 与执行历史的解耦

`DELETE /discover-schedules/{sid}`：

- 从 cron 引擎摘除该 schedule 的注册条目。
- DB 中删除 schedule 记录。
- **不删除已有 DiscoverTask 历史记录**——schedule_id 字段保留为孤儿引用，便于审计追溯"该 schedule 历史触发了哪些任务"。
- 不影响正在执行（status=running）的 DiscoverTask；任务执行时已经持有所需上下文，不再依赖 schedule 配置。

### 3.2 数据模型变更

#### 3.2.1 DiscoverTask 字段重命名

```sql
-- migrations/mariadb/<新版本>_rename_scheduled_id_to_schedule_id.sql
ALTER TABLE t_discover_task
  CHANGE COLUMN f_scheduled_id f_schedule_id VARCHAR(64);
-- （索引 / 外键随列名同步）
```

Go 实体：

```go
type DiscoverTask struct {
    ...
-   ScheduledId string `json:"scheduled_id"`
+   ScheduleID  string `json:"schedule_id"`
    ...
}

type DiscoverTaskQueryParams struct {
    PaginationQueryParams
    CatalogID   string
+   ScheduleID  string
    Status      string
    TriggerType string
}
```

#### 3.2.2 DiscoverSchedule 表（沿用现有 t_scheduled_discover_task）

字段不变（已有 `id / catalog_id / cron_expr / strategies / start_time / end_time / enabled / last_run / next_run / creator / create_time / updater / update_time`），仅 service / handler / 路由层做调整。

是否同时把表重命名为 `t_discover_schedule` 以对齐资源命名？**待团队确认**，本设计倾向"先不改表名"以减小迁移风险，HTTP 层资源名 (`discover-schedules`) 与表名 (`t_scheduled_discover_task`) 之间允许命名差异。

#### 3.2.3 错误码新增

```
DiscoverSchedule.NotFound
DiscoverSchedule.InvalidCronExpr
DiscoverSchedule.InvalidStrategies
DiscoverSchedule.IdMismatch
DiscoverSchedule.CatalogMismatch
DiscoverSchedule.EnabledFieldNotAllowed
DiscoverSchedule.AlreadyEnabled        # 可选，看幂等策略
DiscoverSchedule.AlreadyDisabled       # 可选
```

中英文 i18n 同步补充。

### 3.3 接口定义

OpenAPI yaml 草稿待生成（路径 `adp/docs/api/vega/vega-backend-api/discover-schedule.yaml`、`discover-task.yaml`），格式延续 `connector-type.yaml` / `query.yaml` 的风格：

- OpenAPI 3.1.1
- 共用 `Error` schema 与 responses 块
- 内部接口前缀 `/api/vega-backend/in/v1` 在 `info.description` 中统一描述，不重复列出 path

### 3.4 cron 引擎集成

- `Create` / `Update` / `Enable` / `Disable` / `Delete` 均触发 `Scheduler.Reload()`（或局部 add/remove 单条），保证 cron 注册与 DB 状态实时一致。当前实现需 review 是否已经做到。
- 多实例部署时同一 schedule 会被多副本注册，由 catalog 级互斥（README 已声明"每个 catalog 同一时间只能有一个发现任务在执行"）防重。本次重构不改这一行为。

## 4. 边界情况与风险

| 类型 | 描述 | 应对 |
|---|---|---|
| 并发 | schedule.enabled 状态在两次请求间被改 | enable/disable 端点幂等；客户端不需要先读取状态 |
| 并发 | 删除 schedule 时正好 cron 触发 | 删除时先从 cron 引擎注销，再从 DB 删除；已经在执行的 task 不受影响 |
| 数据一致性 | DB 删 schedule 但 cron 引擎未刷新 | 启动时 `Scheduler.Start()` 重载所有 enabled schedule；变更时 `Reload()` |
| 字段重命名 | `scheduled_id` 与 `schedule_id` 双客户端同时调用 | 不做兼容期：本次发布作为破坏性版本；前端 / 调用方一次性切换 |
| DB 迁移 | 重命名列时锁表 | 表规模评估；数据量大时可拆为"新增列 + 双写 + 切流 + 删旧列"三步迁移（看 DBA 建议） |
| 向后兼容 | 弃用旧路径会破坏现有客户端 | 与外部依赖方对齐时间窗，通知 + 双发布期；本次设计**不**保留旧路由作 deprecated 别名 |
| 性能 | `GET /discover-tasks?schedule_id=` 查询性能 | DB 列上加索引（与原 `f_scheduled_id` 索引同步迁移） |

## 5. 替代方案

### 方案 A：保持 catalog 子资源，仅修内部问题

将 `start/stop` 改为 `enable/disable`，路径仍为 `/catalogs/{cid}/scheduled-discover-tasks/{tid}/...`，跨 catalog 列表剥离到 `/scheduled-discover-tasks`。

**优点**：URI 表达"任务从属于 catalog"的归属关系。
**缺点**：双主键 path 冗余（task_id 已经全局唯一）；同一资源的"列表 vs 单条操作"路径风格分裂；本质上仍是顶层资源被强行嵌套。

**结论**：放弃。归属关系靠资源字段 `catalog_id` 表达即可，不需要 URI 来表达。

### 方案 B：把多份调度合并为 catalog 上的"rules 数组"属性

每个 catalog 上挂一个 `discover_schedule.rules[]` JSON 数组，元素含 cron / strategies / enabled。

**优点**：路径形态最简，仅 `/catalogs/{cid}/discover-schedule`。
**缺点**：

- 数组元素需要独立的 URI / 状态 / 历史 / 审计，最终被迫在元素上造 `rule_id` 字段，等于把 1:N 模型藏进数组里、丢掉 REST 的所有好处。
- 并发编辑（两人同时操作不同 rule）会触发数组级 last-write-wins，必须额外引入 ETag。
- 无法简洁表达"按 rule 反查 task"。

**结论**：放弃。详细论证见前期讨论稿。

### 方案 C：通用 `Schedule` + `target_type/target_id` 抽象

把 schedule 抽成一个能调度任意 target（discover / sync / health-check / ...）的顶层概念。

**优点**：未来扩展性最强。
**缺点**：业务确认非 discover 的周期任务不会出现，过度抽象会引入没用的多态字段。

**结论**：放弃。

### 最终方案：DiscoverSchedule 顶层独立资源 + 字段命名规范化

详见第 2、3 节。

## 6. 任务拆分

按 [adp/CLAUDE.md](../../../../../CLAUDE.md) 规则 5（单批 ≤ 3 文件原则）拆为 5 批：

- [ ] **批次 1：`scheduled_id` → `schedule_id` 重命名 + DB 迁移**
  - DB migration 脚本
  - `interfaces/discover_task.go` 字段重命名 + 新增 `ScheduleID` query 参数
  - `drivenadapters` SQL 字段映射 / DAO
  - mock 同步
  - locale / 错误码不涉及

- [ ] **批次 2：DiscoverSchedule 顶层资源（service + handler + router）**
  - 新增 `interfaces/discover_schedule_service.go`（或在原 `scheduled_discover_task_service.go` 调整）
  - 新增 `driveradapters/discover_schedule_handler.go`
  - 新增 `driveradapters/validate_discover_schedule.go`
  - router 注册新端点（暂不删旧）
  - 错误码 + locale
  - mock 同步
  - 单元测试（PUT 主键冲突 / enable / disable 等核心路径）

- [ ] **批次 3：discover-tasks 路由形态规范化 + `?schedule_id=` 过滤**
  - router：`/by-id/{id}` → `/{id}`，`/by-schedule/{sid}` 删除
  - handler 改造、access SQL 增加 `schedule_id` 过滤
  - 单元测试

- [ ] **批次 4：弃用旧 scheduled-discover 路由**
  - router 删除旧路由
  - handler 中废弃的方法移除
  - CHANGELOG 标注 BREAKING

- [ ] **批次 5：OpenAPI yaml 草稿生成**
  - `adp/docs/api/vega/vega-backend-api/discover-schedule.yaml`
  - `adp/docs/api/vega/vega-backend-api/discover-task.yaml`

每批前按 CLAUDE.md 规则 1 单独描述方案待批准。

## 7. 待团队确认的开放问题

1. **是否同时把 DB 表 `t_scheduled_discover_task` 重命名为 `t_discover_schedule`**？本设计建议先不改，HTTP/Go 层与表名命名差异可接受。
2. **enable/disable 幂等策略**：对已 enable 的资源再 enable 是返回 204（幂等成功）还是 409 `AlreadyEnabled`？建议 204（幂等）。
3. **DB 列重命名一次完成 vs 三步迁移**：取决于线上 t_discover_task 数据量与运维窗口，DBA 建议为准。
4. **是否保留旧路由短期作为 deprecated 别名**？建议不保留（破坏性发布 + 外部对齐），但若有历史合作方需要灰度，可以暂留 1 个版本。
