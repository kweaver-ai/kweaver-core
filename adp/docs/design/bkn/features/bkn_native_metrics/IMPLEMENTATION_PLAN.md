# BKN 原生指标实现计划

> **依据**：[DESIGN.md](./DESIGN.md)（v0.9.8+）  
> **编写**：按 `writing-plans` 技能：面向 **零代码库上下文** 的工程师，任务可勾选跟踪（`- [ ]`）。

> **For agentic workers:** 实现时可配合 `superpowers:subagent-driven-development` 或 `superpowers:executing-plans`，**按 Task 顺序**推进；每 Task 完成后做接口/契约对账。

**Goal:** 在 BKN 内落地 **独立 MetricDefinition**；**契约拆分**：**bkn-backend** 侧 `**bkn-metrics.yaml`** 承载指标 **CRUD** OpenAPI；**ontology-query** 侧 `**ontology-query.yaml`** 仅新增 `**POST .../metrics/{metric_id}/data**`（查数）与 `**POST .../metrics/dry-run**`（试算）。**ontology-query** 实现上述查询与试算；**执行适配（vega Resource）**；**下线** 对象类 metric 逻辑属性与 **uniquery 指标模型** 主路径；**Context Loader** 暴露对接上述路径的工具。

**Architecture:** **bkn-backend** 持久化与 CRUD；**ontology-query** 编排「定义 + ObjectType.data_source + 请求体」→ **ResolvedMetricExecutionPayload** → **vega**（或后续适配器）；**Agent** 优先 **Context Loader** 调 ontology-query。与 [DESIGN §2.2.1](./DESIGN.md#221-context-loader-与-agent-查数) 一致。

**Tech Stack:** Go（bkn-backend、ontology-query、context-loader/agent-retrieval）、MariaDB/DM（迁移以仓库现有方言为准）、OpenAPI 3.x、JSON Schema（附录 B）、vega-backend Resource Data API。

**仓库根路径约定：** 下文路径相对于 `kweaver-core` 仓库；核心代码在 `adp/bkn/`、`adp/context-loader/`、`adp/docs/api/bkn/`。

---

## 1. 设计覆盖核对（Spec coverage）


| DESIGN 章节                                                | 本计划 Task       |
| -------------------------------------------------------- | -------------- |
| §3.2 信息模型 / 附录 B Schema                                  | Task 1、2、3     |
| §3.1 / §3.3 查询与试算（含 `**MetricData` 响应** §3.3.1.2、附录 B.5） | Task 1、6、7、8、9 |
| strict_mode / 概念分组 / 导入导出                                | Task 4、5       |
| §2.2.1 Agent / Context Loader                            | Task 10        |
| §4 迁移与 uniquery 退场                                       | Task 5、11      |
| vega 映射                                                  | Task 8、12      |


---

## 2. 子系统与目录映射


| 子系统                | 主要职责                                                                                                              | 典型路径（存在则改、无则建）                                                                                                                                                                                                                                             |
| ------------------ | ----------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **契约**             | OpenAPI、JSON Schema、错误码枚举                                                                                         | **指标 CRUD**：`adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml`（新建）；**指标查询 / 试算**：`adp/docs/api/bkn/ontology-query-ai/ontology-query.yaml`（仅增上述两 path）；**对象类**：`object-type.yaml`（及 query 侧对象类片段）**去掉** 逻辑属性 **指标（metric）** 相关声明，与设计 §4.2 一致；设计附录 B 可对齐生成物 |
| **bkn-backend**    | 表结构、Metric CRUD、校验、导入导出、对象类去 metric                                                                               | `adp/bkn/bkn-backend/server/`、`migrations/`                                                                                                                                                                                                                |
| **ontology-query** | 注册 `**POST /metrics/{metric_id}/data`**、`**POST /metrics/dry-run**`；合并载荷、调 vega；替代 `drivenadapters/uniquery` 指标路径 | `adp/bkn/ontology-query/server/`（参考 `interfaces/uniquery_access.go`、`drivenadapters/uniquery/`）                                                                                                                                                            |
| **vega**           | Resource Data 承载统一中间表示（联调契约，可能独立仓库）                                                                               | 对接文档 + `resource_data_query_analytics_schema.md`（同目录）                                                                                                                                                                                                      |
| **Context Loader** | MCP/HTTP 工具 `query_metric`、`dry_run_metric`                                                                       | `adp/context-loader/agent-retrieval/server/`（参考 `drivenadapters/bkn_backend.go`、现有 ontology-query mock）                                                                                                                                                    |


---

## 3. 任务分解

### Task 1: 冻结对外契约（OpenAPI + 类型）

**Files:**

- **Create:** `adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml` — **指标 CRUD**（Create/Read/Update/Delete/List 等 paths；`components.schemas` 含 `**MetricDefinition`** 及 CRUD 请求/响应所需类型；路径前缀与现有 `**/api/bkn-backend/v1/...**` 及同目录 `object-type.yaml`、`action-type.yaml` 等风格一致）。
- **Modify:** `adp/docs/api/bkn/bkn-backend-api/object-type.yaml` — **去掉对象类接口中关于「逻辑属性—指标」的声明**：从 `**LogicProperty`**、`**LogicSource**`、`**ObjectType**` 等相关 **schema**（如 `type` **枚举**含 `metric`、数据源 `**metric-model`**）、**description**、**示例**、`**Parameter4Metric`** 等片段中 **删除或改写**；`**strict_mode` 等文案** 若仍写「逻辑属性绑定的指标/指标模型」，改为与 **独立 `MetricDefinition`** 或 **已废弃** 语义一致（与 DESIGN §4.2、本计划 Task 5 对齐）。
- **Modify:** `adp/docs/api/bkn/ontology-query-ai/ontology-query.yaml` — **仅新增** 指标 **查询数据** 与 **试算** 两 path；**components/schemas** 增补 `MetricQueryRequest`、`MetricDryRun`、`MetricData`（**200**）、与 `Condition` 的 `$ref`；**不** 在本文件承载指标 CRUD。**同时**：若文中 **对象类** 请求/响应 **示例或 schema** 仍包含 `**logic_properties` + `type: metric`** 等，**同步删除或改写**，与 `**object-type.yaml`** 一致。
- Create（可选）: `adp/docs/design/bkn/features/bkn_native_metrics/openapi-diff-notes.md`（若需记录与现网 diff，**非必须**）

**Steps:**

- **Step 1（bkn-backend）：** 对照 [DESIGN §3.2 / 附录 B.2](./DESIGN.md) 在 `**bkn-metrics.yaml`** 中冻结 **指标 CRUD** 路径与 `**MetricDefinition`** 等 schema；与 **bkn-backend** 路由、鉴权头（如 `X-Business-Domain`）与同目录 OpenAPI **对齐**。
- **Step 1b（对象类契约）：** 按上表 **修订 `object-type.yaml`**，并 **扫一遍 `ontology-query.yaml`** 中对象类相关片段，**去掉逻辑属性—指标** 的 OpenAPI 声明（枚举、示例、专用子 schema 等）；与 **Task 5**「禁止 `type == metric` 逻辑属性」**对外契约一致**。（若产品要求迁移期保留 **deprecated** 字段说明，在对应 schema `description` 中注明并链到 `**bkn-metrics.yaml`**。）
- **Step 2（ontology-query）：** 对照 [DESIGN 附录 B](./DESIGN.md#附录-bjson-schema-示例草案级供-openapi--校验对齐) 与 §3.3.1 / **§3.3.1.2**，在 `**ontology-query.yaml`** 中仅增加：`**POST .../metrics/{metric_id}/data**`（请求体 `MetricQueryRequest`、**200** `**MetricData`**）、`**POST .../metrics/dry-run**`（`MetricDryRun`、**200** `**MetricData`**）。`**MetricData**` 与 **uniquery** `interfaces.MetricData` 字段一致（见 `adp/bkn/ontology-query/server/interfaces/uniquery_access.go`）。**components** 增补 `MetricData`、`MetricDryRun`、`MetricQueryRequestBody`（名称与现有一致）、`Condition` `**$ref`**；两 path 的 **200** 均 `**$ref: '#/components/schemas/MetricData'`**（或与现网组件名显式映射）。若 `**MetricDefinition**` 需在 query 服务文档中引用（例如错误体示例），可用 **相对路径 `$ref`** 指向 `**bkn-metrics.yaml**` 或 **复制最小片段**（与团队 **多文件 OpenAPI** 约定一致，避免重复定义冲突）。
- **Step 3:** 跑 **spectral** 或团队现有 **OpenAPI lint**（若 CI 已配置）；无则至少 **人工校验** `**bkn-metrics.yaml`、`object-type.yaml`、`ontology-query.yaml`** 可解析、**无互相矛盾的 metric 逻辑属性** 描述。

```bash
# 若已安装 spectral（示例）
spectral lint adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml
spectral lint adp/docs/api/bkn/bkn-backend-api/object-type.yaml
spectral lint adp/docs/api/bkn/ontology-query-ai/ontology-query.yaml
```

- **Step 4:** **Commit** `docs(api): add bkn-metrics.yaml, metric query paths, object-type metric cleanup`（或拆成多次提交：先 `bkn-metrics.yaml`，再 `object-type`/`ontology-query` 清理与查询 path）

---

### Task 2: 持久化 — `MetricDefinition` 表与迁移

**Files:**

- Create/Modify: `adp/bkn/bkn-backend/migrations/mariadb/0.7.0/`（本计划约定版本目录 `**0.7.0`**；若仓库已有更高版本目录，以**递增**为准并同步改本句）
- Create/Modify: 对应 `dm8/0.7.0/` 迁移（若双方言必跑）
- Modify: 访问层初始化处注册新表（以现有 `init.sql` 模式为准）

**Steps:**

- **Step 1:** 按 DESIGN §3.2.1 列 **列清单**（`id`、`kn_id`、`branch`、`scope_type`、`scope_ref`、JSON 列存 `calculation_formula` / `time_dimension` 等）。
- **Step 2:** 编写 **up** 迁移 SQL；本地 **空库执行** 验证无语法错误。
- **Step 3:** **Commit** `feat(bkn-backend): add metric_definition migration`

---

### Task 3: bkn-backend — Metric CRUD 与路由

**Files（示例，按现有分层微调）:**

- **契约（与 Task 1 一致）：** `adp/docs/api/bkn/bkn-backend-api/bkn-metrics.yaml` — 实现时 **path / operationId / schema** 与本文件 **对账**。
- Create: `adp/bkn/bkn-backend/server/interfaces/metric.go`（或并入现有 `interfaces/` 命名）
- Create: `adp/bkn/bkn-backend/server/interfaces/metric_service.go`
- Create: `adp/bkn/bkn-backend/server/driveradapters/metric_handler.go`
- Create: `adp/bkn/bkn-backend/server/drivenadapters/metric/metric_access.go`
- Modify: `adp/bkn/bkn-backend/server/main.go` 或路由注册处（wire handler）

**Steps:**

- **Step 1:** 定义 **Go struct** 与 DESIGN 字段一致；**JSON tag** 与 `**bkn-metrics.yaml`** 中 `**MetricDefinition**` 及 CRUD 请求/响应 **对账**。
- **Step 2:** 实现 **Create/Read/Update/Delete/List**；**单元测试** access 层（`go test ./adp/bkn/bkn-backend/server/drivenadapters/metric/... -count=1`）。
- **Step 3:** **Commit** `feat(bkn-backend): metric CRUD API`

---

### Task 4: bkn-backend — 校验链（`strict_mode`、scope、概念分组、`data_source`）

**Files:**

- Modify: `validate_*` 或 `validate_metric.go`（与 `validate_action_type.go` 模式对齐）
- Modify: `adp/bkn/bkn-backend/server/common/condition/`（若 `calculation_formula.condition` 复用校验）

**Steps:**

- **Step 1:** 实现 **scope_ref → ObjectType** 存在且属本 `kn_id`/`branch`。
- **Step 2:** `**data_source.type == resource`** 且可解析；`**data_view**` 拒绝创建指标（DESIGN §3.3.2）。
- **Step 3:** **表驱动测试** `strict_mode` true/false；`go test` 覆盖。
- **Step 4:** **Commit** `feat(bkn-backend): metric validation strict_mode and scope`

---

### Task 5: 对象类逻辑属性 metric 下线 + 导入导出

**Files:**

- Modify: `adp/bkn/bkn-backend/server/driveradapters/*object_type*`、`validate_object_type.go` 等
- Modify: 导出/导入 pipeline（概念分组、KN 包）
- **契约（与 Task 1 Step 1b 对账）：** `object-type.yaml`、`**ontology-query.yaml`** 中对象类片段已 **无** metric 逻辑属性对外声明（见 Task 1）。

**Steps:**

- **Step 1:** **禁止** 新建/保存 `logic_properties` 中 `type == metric`（与 DESIGN §4.2 一致）；**错误信息**可指向独立指标；与 **Task 1 Step 1b** OpenAPI **语义一致**。
- **Step 2:** **导出清单** 包含 `MetricDefinition`；**导入顺序** 对象类 → 分组 → 指标。
- **Step 3:** **Commit** `feat(bkn-backend): deprecate metric logic property and export metrics`

---

### Task 6: ontology-query — 拉取 MetricDefinition（适配器）

**Files:**

- Create: `adp/bkn/ontology-query/server/drivenadapters/metric/` 或复用 **HTTP 客户端** 调 bkn-backend
- Modify: `adp/bkn/ontology-query/server/interfaces/` 新增 `MetricDefinitionReader` 接口
- Modify: `adp/bkn/ontology-query/server/interfaces/driven.go` 或 `main.go` 注入

**Steps:**

- **Step 1:** 选定 **数据源**：仅 HTTP 拉 bkn-backend **或** 共享 DB（与架构组一致）；**禁止** 在 query 服务内直接写指标表若违反单一写入源。
- **Step 2:** 实现 **按 `metric_id` + kn/branch** 拉取定义；**mock 测试**。
- **Step 3:** **Commit** `feat(ontology-query): fetch MetricDefinition from bkn-backend`

---

### Task 7: ontology-query — 编排与 `ResolvedMetricExecutionPayload`

**Files:**

- Create: `adp/bkn/ontology-query/server/logics/metric/`（或 `metric_query/`）
- Modify: `adp/bkn/ontology-query/server/driveradapters/` 注册 HTTP 路由

**Steps:**

- **Step 1:** 注册 `**POST .../metrics/{metric_id}/data`** handler；路径参数 `metric_id` + 请求体 **MetricQueryRequest**（见 DESIGN §3.3）；**200** 序列化为 `**MetricData`**（DESIGN §3.3.1.2、附录 B.5）。
- **Step 2:** 按 DESIGN §3.1.2 实现 **加载 MetricDefinition → ObjectType → 合并请求体**。
- **Step 3:** 定义 **内部结构体** `ResolvedMetricExecutionPayload`（与 DESIGN 命名一致）。
- **Step 4:** **单元测试** 合并逻辑（无外部 IO 用 mock）。
- **Step 5:** **Commit** `feat(ontology-query): metric query POST /metrics/{metric_id}/data`

---

### Task 8: ontology-query — 执行适配（vega Resource）与 uniquery 切换

**Files:**

- Modify: `adp/bkn/ontology-query/server/drivenadapters/uniquery/uniquery_access.go` 调用链 **或** 新建 `vega_resource_adapter.go`
- 参考: [resource_data_query_analytics_schema.md](./resource_data_query_analytics_schema.md)

**Steps:**

- **Step 1:** 当 **ObjectType.data_source.type == resource** 时，将 **统一中间表示** 映射为 **vega Resource Data** 请求（与 vega 团队冻结字段）；**响应** 归一化为 `**MetricData`**（与 `GetMetricDataByID` 返回形状一致，DESIGN §3.3.1.2）。
- **Step 2:** **特性开关**：新指标走 BKN 路径；旧 **uniquery metric-models** 仅在迁移期保留（见 DESIGN §6）。
- **Step 3:** **集成测试**（需 vega mock 或测试环境）：`go test ./adp/bkn/ontology-query/server/... -count=1 -short` 或 tagged `integration`。
- **Step 4:** **Commit** `feat(ontology-query): vega adapter for metric execution`

---

### Task 9: ontology-query — `POST .../metrics/dry-run`（MetricDryRun）

**Files:**

- Modify: 同 Task 7 handler 包；**复用** Task 7 编排，**跳过** 持久化读 `metric_id`

**Steps:**

- **Step 1:** 实现 **从 `metric_config` 构造内存 MetricDefinition**；校验链 **复用** bkn-backend 规则 **或** 内嵌同一套 validator 包（避免重复实现时抽 **共享模块** `adp/bkn/common/...` 若允许）。
- **Step 2:** **200** 响应与 **正式查询** 相同，均为 `**MetricData`**（DESIGN §3.3.3、§3.3.1.2、附录 B.5）。
- **Step 3:** **Commit** `feat(ontology-query): metric dry-run endpoint`

---

### Task 10: 退场与清理

**Files:**

- Modify: `adp/bkn/bkn-backend/.../data_model`（`GetMetricModelByID` 等）
- Modify: `adp/bkn/ontology-query/server/logics/object_type/object_type_service.go`（metric 逻辑属性取值路径）

**Steps:**

- **Step 1:** **特性开关** 关闭旧路径；监控 **无调用** 后移除 dead code。
- **Step 2:** 更新 **ontology-query.yaml** 中 **deprecated** 说明。
- **Step 3:** **Commit** `chore: remove legacy metric-model and logic-property metric paths`

---

### Task 11: 文档、E2E、发布检查清单

**Files:**

- Modify: `adp/docs/design/bkn/features/bkn_native_metrics/DESIGN.md`（**状态** 改为「已落地」时）
- 新增/更新: 运维 Runbook（若需要）

**Steps:**

- **Step 1:** **E2E**：创建指标 → `**POST .../metrics/{metric_id}/data`** → `**POST .../metrics/dry-run**` → Agent/CL 工具（最小路径）。
- **Step 2:** **错误码表** 与 DESIGN §3.4.1 对齐。
- **Step 3:** **Commit** `docs: mark BKN native metrics rollout`

---

## 4. 自检（Self-Review）

- **Spec coverage：** DESIGN §2～§4、附录 B 均有对应 Task。
- **无占位符：** 本计划无「TBD 实现」；具体 URL/路径以**实现时** OpenAPI 为准若需微调。
- **契约分文件：** 指标 **CRUD** 在 `**bkn-metrics.yaml`**；**查询 / 试算** 在 `**ontology-query.yaml`**；二者职责不混放。
- **对象类契约：** `**object-type.yaml`**（及 `**ontology-query.yaml**` 内对象类相关 schema/示例）**已去掉** 逻辑属性 **指标（metric）** 声明，与 **Task 1 Step 1b**、**Task 5** 一致。
- **类型一致：** `**MetricDefinition`**（`bkn-metrics.yaml` / bkn-backend Go）与 `**MetricDryRun**`、`**MetricQueryRequest**`、`**MetricData**`、`**scope_context**`（`ontology-query.yaml` / ontology-query Go / CL）在 **OpenAPI / Go** 三处同名或显式映射表；正式查询路径 `**POST /metrics/{metric_id}/data`** 与 DESIGN 一致；`**MetricData**` 与 **uniquery** `interfaces.MetricData` 字段对齐。

---

## 5. 执行方式（Execution Handoff）

**计划文件路径：** `adp/docs/design/bkn/features/bkn_native_metrics/IMPLEMENTATION_PLAN.md`

**可选执行方式：**

1. **Subagent-Driven（推荐）** — 每 Task 派生子代理，Task 间做契约评审。
2. **Inline Execution** — 本会话按 Task 顺序实现，**Task 8/10** 建议设检查点（vega/CL 联调）。

需要我按某一 Task 展开到「逐文件 patch 清单」时，指定 **Task 编号** 即可。