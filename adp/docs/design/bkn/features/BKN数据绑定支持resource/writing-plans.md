# BKN 数据绑定支持 resource — 实施任务清单（writing-plans）

> 依据：[DESIGN.md](./DESIGN.md)（#336）  
> 用途：拆 PR、排期、验收对照；完成后将对应项勾为 `[x]`。

---

## 阶段 0：契约与常量（阻塞后续开发）

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| P0-1 | 冻结 `data_source.type` / `backing_data_source.type` 枚举值 **`resource`**（与 JSON/API 文档一致） | OpenAPI/对外文档或代码常量中可见 `data_view` \| `resource` | 无 |
| P0-2 | 产出 **QueryResourceData** 与 vega `ResourceDataQueryParams` / 响应体的 **字段对照表**（含 JSON tag） | 文档或 PR 描述可逐字段核对；与 `vega/.../resource_data.go`、handler 行为一致 | 无 |
| P0-3 | 固化 HTTP 约定：**`POST`** + **`Content-Type: application/json`** + **`x-http-method-override: GET`** | bkn-backend / ontology-query 适配层注释或封装函数内显式设置 | P0-2 |

---

## 阶段 1：bkn-backend — 接口与适配层

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| B1-1 | 在 `interfaces` 增加 `ResourceDataQueryParams` / 响应 DTO（与 ontology-query **同构**）及 `VegaBackendAccess.QueryResourceData` | `go build` 通过；mock 接口可生成 | P0-2 |
| B1-2 | `drivenadapters/vega_backend` 实现 `QueryResourceData`（baseUrl、account header、超时、错误映射） | 单测或集成测覆盖成功/4xx/5xx；header 符合 P0-3 | B1-1 |
| B1-3 | `go generate` / 更新 `mock_vega_backend_access` | mock 与接口一致 | B1-1 |

---

## 阶段 2：bkn-backend — 对象类与关系类业务

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| B2-1 | **ObjectType**：`data_source.type == resource` 时 `GetResourceByID` 补全 `name`、schema 驱动 `MappedField` 展示名/类型、条件操作符等与 view 对称 | 单测覆盖详情接口；非法 resource id 行为明确 | B1-2 |
| B2-2 | **ObjectType** 创建/更新校验：映射字段在 **resource schema** 中存在 | 缺字段报错语义符合 DESIGN §2.3 决策 6 | B2-1 |
| B2-3 | **RelationType** 间接映射：`backing_data_source.type == resource` 时走 **GetResourceByID** 校验与名称回填，不再强依赖 `GetDataViewByID` | 单测覆盖；与 data_view 用例对称 | B1-2 |

---

## 阶段 3：bkn-backend — 索引 Worker

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| B3-1 | `ObjectTypeTask`（或等价 worker）：`type == resource` 时不再 early-return；用 **QueryResourceData** 分页拉取 | 与视图路径一样写入 OpenSearch；批处理日志可观测 | B1-2, B2-2 |
| B3-2 | **增量任务**：明确 resource 是否支持 `JobTypeIncremental`；不支持则校验拒绝或文档化 | 无静默跳过；行为可测 | B3-1 |
| B3-3 | 大结果集：分页/`search_after` 与 vega 限流错误 **透传或包装** | 与设计文档错误语义一致 | B3-1 |

---

## 阶段 4：bkn-backend — 审计与其它

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| B4-1 | 对象类/关系类绑定变更审计中记录 **`data_source.type` + `id`**（resource 与 view 同模型） | 审计日志或单测断言 | B2-2, B2-3 |
| B4-2 | 血缘/依赖（若现网有）：resource id 同等记录 | 与产品对血缘范围约定一致 | 视现网能力 |

---

## 阶段 5：ontology-query — 配置与客户端

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| O5-1 | 配置：`VegaBackendUrl`（或项目内等价名）+ 超时；`main` 注入 | 配置样例写入 `ontology-query-config` 示例 | P0-3 |
| O5-2 | 新建 `drivenadapters/vega_backend`，DTO **与 bkn-backend B1-1 一致** | 代码评审可 diff 对齐；禁止随意改字段名 | B1-1（对齐快照） |
| O5-3 | 实现 `QueryResourceData`（及按需 `GetResourceByID`） | 单测 mock HTTP | O5-2 |

---

## 阶段 6：ontology-query — 查询逻辑

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| O6-1 | **ObjectType 查询**：`data_source.type == resource` 且无索引（或设计规定走源）→ **QueryResourceData**；条件重写与 `ViewQuery` 对齐策略明确 | 单测；与设计双路径一致 | O5-3, B2-1 |
| O6-2 | **有索引** 路径：仍走 **OpenSearch**，不因 resource 破坏 | 回归单测 | O6-1 |
| O6-3 | **Knowledge network / 间接关联**：`backing_data_source` 为 resource 时批量拉数走 **QueryResourceData**，后续复用 `viewData` 映射逻辑 | 单测或集成场景 | O5-3, B2-3 |

---

## 阶段 7：测试与文档收口

| ID | 任务 | 验收标准 | 依赖 |
|----|------|----------|------|
| T7-1 | 覆盖 DESIGN §4.1 摘要用例（详情、非法 id、缺映射列、分页、5xx） | `bkn-backend` / `ontology-query` `go test` 通过 | 阶段 1–6 |
| T7-2 | 对外 API：枚举 **`resource`** 说明；运维：vega 不可用报错与重试 | 文档或 OpenAPI 已更新 | T7-1 |
| T7-3 | **DESIGN.md** 状态改为「已评审/已实现」（按团队规范） | 与发布版本一致 | T7-1 |

---

## 建议 PR 切分（降低一次改 >3 文件的心智负担）

1. **PR-A**：P0 + B1（契约 + VegaBackendAccess + adapter + mock）  
2. **PR-B**：B2 + B4（业务校验 + 审计）  
3. **PR-C**：B3（worker 索引）  
4. **PR-D**：O5 + O6（ontology-query）  
5. **PR-E**：T7（测试与文档）

若必须单 PR，按 **B1 → B2 → B3 → O5 → O6** 顺序提交（commit）便于 review。

---

## 失败条件（整体验收不通过）

- `type=resource` 仅配置层可见，**索引或查询任一链路未通**。  
- bkn-backend 与 ontology-query **DTO 不一致**导致线上解析错误。  
- 未设置 **`x-http-method-override: GET`** 导致 vega **400** 且未在文档/代码中固化。  
- 审计未记录 **type + id**，与现网 data_view 不对齐。
