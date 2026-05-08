# BKN 范式指标导入/导出 — 任务清单

> **依据**：[DESIGN.md](./DESIGN.md)  
> **关联工单**：#377  
> **路径约定**：相对于仓库根目录 `kweaver-core`；后端代码在 `adp/bkn/bkn-backend/`。

实现顺序建议自上而下；同一 Phase 内条目可按依赖微调。**`- [x]` 表示已在当前仓库落地或已有等价实现；`- [ ]` 表示未完成或仅部分完成（见备注）。**

---

## 进度摘要（对照代码库当前实现）

| Phase | 说明 |
| ----- | ---- |
| **0** | `replace`、映射文档已在本文件附录补充；规范分支逐项核对 / 独立 SDK harness 仍依赖人工或 CI。 |
| **1～5** | **核心导入导出链路已实现**：`KN.Metrics`、`bkn_convert` 指标转换、`UploadBKN` 填充与校验、`CreateKN` 事务内 `CreateMetrics`、`GetKNByID(Export)` `ListMetrics`、`ToBKNNetWork`/`ExportToTar`、`ImportMode_Overwrite` 下指标分支、`ValidateKN`/`ValidateMetrics`。 |
| **1.3** | **未做**：`batchindex` 仍未纳入指标 ID / scope（当前 `UploadBKN` 使用 `strictMode=false`，与设计「严格模式顺序依赖」风险并存）。 |
| **6** | **指标概念索引**：`CreateMetrics` 内写 dataset；`concept_syncer` 已有 `handleMetrics` 对账；**KN 单条概念文档**仍不嵌套 `metrics` 字段（与 OT 等同层级索引分离）。 |
| **7～8** | **集成测试 / CI replace 策略 / 发版去 replace / 主计划勾链 / 状态收尾** 多为流程项，代码侧尚未全部闭合。 |

---

## Phase 0 — 前置：规范/SDK 对账（阻塞后续）

- [ ] **0.1** 在本地 `bkn-specification` **spec 指标分支**（如 `feature/7-issue`）核对：tar 内指标文件命名、`metric` 语法、`BknNetwork` / Go 导出字段名（对齐 DESIGN 1.5 三条）。*（人工 / PR 评审勾选）*
- [x] **0.2** 产出 **映射表**：见本文 **附录 A**（SDK 类型名为 `BknMetric`，与 DESIGN 草案中的 `BknMetricDefinition` 表述对齐）。
- [x] **0.3** 在 `adp/bkn/bkn-backend/server/go.mod` 配置 **`replace`**：`github.com/kweaver-ai/bkn-specification/sdk/golang` → 本机 `sdk/golang`（路径 `../../../../../bkn-specification/sdk/golang`，依赖仓库与 `kweaver-core` 同级布局）；本地 **`go build` / `go test` 需配合 `I18N_MODE_UT=true`**。
- [ ] **0.4** 单独验证 SDK：`LoadNetworkFromTar` / `WriteNetworkToTar` / `SerializeBknNetwork` 对 **Metrics** 的 round-trip（最小示例或可合并进 Phase 7 集成用例）。*（可与 7.3 合并）*

---

## Phase 1 — 领域模型与载荷

- [x] **1.1** `server/interfaces/knowledge_network.go`：`KN` 增加 `Metrics []*MetricDefinition`（`json` / `mapstructure` 与现有 KN 字段一致）。
- [ ] **1.2** 全局检索 `KN` **构造与拷贝**：按需补齐 `Metrics` 传递（全仓逐项审计未完成；新增路径以 `UploadBKN` → `CreateKN`、`GetKNByID` export 为主已贯通）。
- [ ] **1.3**（按需）`batchindex`：`CollectKNFromPayload` 等纳入 **指标 ID** 与 **scope_ref**。*（未实现；若开启 Upload `strictMode=true` 建议优先补）*

---

## Phase 2 — BKN ↔ ADP 转换

- [x] **2.1** `server/logics/bkn_convert.go`：实现 **`ToBKNMetricDefinition`**（导出侧嵌入 `ToBKNNetWork`）。
- [x] **2.2** 同上：实现 **`ToADPMetricDefinition`**（导入侧由 `bkn_handler` 调用）。
- [x] **2.3** 单元测试：`bkn_convert_test.go` 中 `Test_ToADPMetricDefinition_MinimalAtomic`、`Test_ToBKNMetricDefinition_RoundTrip_KeyFields`。*（`nil`/空切片可加例强化，见备注）*

---

## Phase 3 — 导出：`GetKNByID` + `ExportToTar`

- [x] **3.1** `knowledge_network_service.go`：`GetKNByID(..., Mode_Export)` 调用 **`MetricService.ListMetrics`**（`Limit: -1`），填充 `kn.Metrics`。
- [x] **3.2** `bkn_service.go`：`ExportToTar` 通过 **`ToBKNNetWork(kn)`** 写出 **Metrics**（含各 `ToBKN*`），与 SDK `WriteNetworkToTar` 一致。
- [ ] **3.3** 自动化或脚本：含指标的 KN **下载 tar → `LoadNetworkFromTar`**，对账条数与关键字段。*（待 Phase 7）*

---

## Phase 4 — 校验链（与 REST 对齐）

- [x] **4.1** `knowledge_network_service.ValidateKN`：编排 **`ValidateMetrics`**（与 OT/RT/AT 等同层）。
- [x] **4.2** `driveradapters/bkn_handler.go`：`UploadBKN` 填充 **`kn.Metrics`** 并调用 **`ValidateMetrics`**（`strictMode=false`，与当前 `CreateKN(..., false)` 一致）。
- [ ] **4.3** **顺序与严格模式**：设计上应先 OT/CG 再校验指标 `scope_ref`；当前 Upload 路径 **`strictMode=false`**，`ValidateMetrics` 不严格外依赖 DB。*（开启 strict 时需结合 1.3）*
- [ ] **4.4** 负例测试：`strict_mode` 下非法 `scope_ref` 等与 REST 批量指标 API **对齐**。*（待补充用例）*

---

## Phase 5 — 导入：`CreateKN` 事务内写指标

- [x] **5.1** `CreateKN`：在 **isCreate / isUpdate** 分支中，于 OT/RT/AT/Risk 之后调用 **`MetricService.CreateMetrics(tx, kn.Metrics, ...)`**，失败回滚事务。
- [x] **5.2** **`ImportMode_Overwrite`**：`metric_service.handleMetricImportMode` 支持 **creates + updates**，与 OT 覆盖语义对齐（原「overwrite 不支持」已移除）。
- [x] **5.3** **UpdateKN 路径**：与 Upload 相关的 **isUpdate** 分支同样调用 **`CreateMetrics`**。
- [x] **5.4** **`SerializeBknNetwork`**：`CreateKN` 开头对 **`ToBKNNetWork(kn)`** 序列化时，`kn.Metrics` 若在调用前已填充则会进入 tar 语义缓存。**说明**：`InsertDatasetData` 写入的 **KN 概念文档**仍不序列化 `metrics` 数组；指标由 **`CreateMetrics`/worker** 写独立 metric 文档。

---

## Phase 6 — 概念索引与异步一致

- [x] **6.1** **指标**写入概念数据集：`MetricService.CreateMetrics` → **`InsertDatasetData`（指标）**；与 KN 主文档分离属当前架构。
- [x] **6.2** **`concept_syncer`**：已有 **`handleMetrics`**，按 DB 与 dataset 对账增量。*（与设计文档第 3.1 节是否完全一致，可按联调再勾选「已验收」）*

---

## Phase 7 — 集成与回归测试

- [ ] **7.1** fixture：含 **≥1** `MetricDefinition` 的 BKN tar（或 examples 目录）。
- [ ] **7.2** `tests/integration_tests/bkn/bkn_test.go`：**POST .../bkns** → **GET .../metrics** 对账。
- [ ] **7.3** **导出 → 再导入 round-trip**。
- [ ] **7.4** CI：`replace` 策略文档化（clone `bkn-specification`、相对路径或 `go.work`）。

---

## Phase 8 — 收尾与依赖清理

- [ ] **8.1** `bkn-specification` **正式发版**后：去掉 **`replace`**，升级 **require** 版本号。
- [ ] **8.2** 更新 [`IMPLEMENTATION_PLAN.md`](../bkn_native_metrics/IMPLEMENTATION_PLAN.md) 导入导出勾选项或指向本文档。
- [ ] **8.3** [DESIGN.md](./DESIGN.md) / 本文 **状态**改为「已实现」并链接合并 PR。

---

## 验收核对（发布前快速扫）

对照 [DESIGN.md 第 4 节](./DESIGN.md)：

- [ ] `go test ./...`（约定 SDK + `I18N_MODE_UT=true`）通过。
- [ ] 含指标 tar **导入**成功且库内一致。
- [ ] 导出 tar **`LoadNetworkFromTar`** 指标完整。
- [ ] `strict_mode` 负例与 REST 指标 API 一致。

---

## 附录 A：`bkn.BknMetric` ↔ `interfaces.MetricDefinition`（实现对照）

> SDK 结构体名为 **`BknMetric`**（见 `bkn-specification/sdk/golang/bkn`）；下列对应 **`ToADPMetricDefinition` / `ToBKNMetricDefinition`**（`server/logics/bkn_convert.go`）。

| BknMetric（含 Frontmatter / Attributes） | MetricDefinition |
| ---------------------------------------- | ---------------- |
| `ID`, `Name`, `Tags` | `ID`, `Name`, `Tags` |
| `Summary` / `Description`（优先 Description） | `Comment` |
| `MetricAttributes.MetricType`，缺省时 `Formula.Kind` | `MetricType` |
| `MetricAttributes.UnitType`, `Unit` | `UnitType`, `Unit` |
| `ScopeType`, `ScopeRef` | `ScopeType`, `ScopeRef` |
| `Formula.Atomic`（condition / aggregation / group_by / order_by / having） | `CalculationFormula` |
| `TimeDimensions[0]` | `TimeDimension`（单行聚合） |
| `AnalysisDimensions[]` | `AnalysisDimensions[]` |
| `RawContent` | `BKNRawContent` |

**未做单层 round-trip 的复杂形态**：`CondCfg.sub_conditions` 等与 SDK `MetricCondition` 叶子形态不一致时，仍以 REST/DB 校验为准；导入路径当前覆盖原子公式常用字段。
