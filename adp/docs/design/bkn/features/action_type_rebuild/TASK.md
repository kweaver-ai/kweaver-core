# 行动类（ActionType）重构 — 实施任务清单（writing-plans）

> 依据：[DESIGN.md](./DESIGN.md)（#/288）  
> 用途：拆 PR、排期、验收对照；完成后将对应项勾为 `[x]`。

---

## 阶段 0：契约与对齐（阻塞后续开发）


| ID   | 任务                                                                                                                     | 验收标准                                                       | 依赖   |
| ---- | ---------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- | ---- |
| P0-1 | 与 `**bkn-specification**`（及发布节奏）对齐 `BknActionType`：`action_intent`、`impact_contracts`（数组）、扩展后的 `AffectObject` / 单行影响结构 | SDK 可解析/序列化；示例 `.bkn` 片段与 DESIGN §5.5、§7 一致；`go mod` 版本可构建 | 无    |
| P0-2 | 冻结 `**action_intent` 枚举**与 `**action_type` 完全一致**（`add` / `modify` / `delete`）；文档化与代码常量单一来源                            | 常量或生成表与 OpenAPI `enum` 一致；与 DESIGN §7.1 无歧义                | P0-1 |
| P0-3 | 明确 **本期不落地** 字段清单：`missing_root_policy`、`context_scope`、`navigation_boundary`（见 DESIGN §7.3）；规范或反序列化策略（忽略 vs 拒绝）经评审落字  | ADR 或 PR 说明；无「半实现」歧义                                       | 无    |


---

## 阶段 1：`bkn-backend` — `interfaces` 与 DTO


| ID   | 任务                                                                                                                                       | 验收标准                                                       | 依赖   |
| ---- | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------- | ---- |
| B1-1 | 在 `ActionTypeWithKeyField`（或等价 squash 结构）增加 `**action_intent`**（`string`）                                                                | JSON/`mapstructure` tag 与 OpenAPI 字段名一致；零值语义与回填策略在 B1-4 体现 | P0-2 |
| B1-2 | 定义 `**ImpactContractItem**`（或项目约定命名）：含 `object_type_id`、`expected_operation`、`description`、`affected_fields`、及 DESIGN 要求的 `scope`（若本期纳入） | 字段与 §7.2、§5.4.4 对齐；omitempty 规则明确                          | P0-1 |
| B1-3 | 增加 `**impact_contracts []ImpactContractItem**`（切片，序列化为 JSON 数组）                                                                          | 空切片与 `nil` 在 API 表现约定清楚（省略 vs `[]`）                        | B1-2 |
| B1-4 | 扩展 `**ActionAffect**`：增加 `**expected_operation**`、`**affected_fields**`（及与 `ImpactContractItem` 对齐的其余键）                                  | 旧数据无新字段时反序列化不失败；与 §7.2 一致                                  | B1-2 |
| B1-5 | （可选）抽取 `**action_intent` 与 `action_type` 合法值**共享校验入口，避免魔法字符串散落                                                                           | `validate_action_type` / service 复用同一函数                    | P0-2 |


---

## 阶段 2：`bkn-backend` — 校验（`validate_action_type` 等）


| ID   | 任务                                                                                                           | 验收标准                    | 依赖         |
| ---- | ------------------------------------------------------------------------------------------------------------ | ----------------------- | ---------- |
| V2-1 | `**action_intent`**：非空时校验属于 P0-2 枚举                                                                          | 非法值返回现有错误码风格 + 可测用例     | B1-1       |
| V2-2 | **双写一致**：请求体同时带 `action_type` 与 `action_intent` 时**值必须相同**；否则 400                                            | 单测：一致通过 / 不一致失败         | B1-1       |
| V2-3 | **缺省回填**（三选一，与实现 PR 描述一致）：仅 `action_type` → 补 `action_intent`；或仅 `action_intent` → 补 `action_type`；或二者必须同时出现 | 与 §7.1「兼容存量」一致；单测覆盖     | V2-1, V2-2 |
| V2-4 | `**impact_contracts`**：校验数组元素必填项、 `expected_operation` 枚举（若独立枚举）、`affected_fields` 类型（`[]string` 等）          | 非法结构有明确错误信息             | B1-3       |
| V2-5 | **strict 模式**：`**impact_contracts`** 每条 `**object_type_id**` 在 **当前 KN/分支** 存在（**允许任意对象类**，不限于绑定根）           | 单测：存在通过、不存在失败；与 §7.2 一致 | B1-3       |
| V2-6 | **扩展后 `affect`**：当仅使用 `affect` 时校验新增字段形态；与 `impact_contracts` 互斥/并存规则按设计实现                                   | 单测覆盖单行 `affect` 全量字段    | B1-4       |
| V2-7 | 确认 `**cond` / `parameters**` 校验路径**无行为变化**（§7.4）                                                             | 回归或显式单测「未改分支」           | 无          |


---

## 阶段 3：`bkn-backend` — 互转与 BKN 映射（`bkn_convert`）


| ID   | 任务                                                                                | 验收标准                                    | 依赖         |
| ---- | --------------------------------------------------------------------------------- | --------------------------------------- | ---------- |
| C3-1 | `**ToBKNActionType`**：写出 `action_intent`、`impact_contracts`；扩展 `Affect` 对应 BKN 字段 | 与 `bknsdk.SerializeActionType` 无冲突      | P0-1, B1-3 |
| C3-2 | `**ToADPActionType**`：读入上述字段；**缺 `action_intent`** 时用 `action_type` 回填            | 单测：旧 YAML 仍可导入                          | C3-1       |
| C3-3 | 实现 `**affect` ⇄ `impact_contracts**`：单行 ⇄ **长度为 1** 的数组；多元素仅存 `impact_contracts`  | 单测：`ToADP`→`ToBKN`→`ToADP` 允许等价（约定默认字段） | C3-1, B1-4 |
| C3-4 | `**BKNRawContent`** 写入路径在 Create/Update 中仍由序列化触发；含新字段后不损坏旧客户端                     | 集成或单测覆盖一条完整行动类                          | C3-2       |


---

## 阶段 4：`bkn-backend` — Service / Access / DB


| ID   | 任务                                                                        | 验收标准                             | 依赖         |
| ---- | ------------------------------------------------------------------------- | -------------------------------- | ---------- |
| S4-1 | `**CreateActionTypes` / `UpdateActionType**`：在持久化前应用 V2-3 回填与 V2-2 一致性强约束 | 无绕过 validate 的路径                 | V2-2, V2-3 |
| S4-2 | `**ValidateActionTypes**`：纳入 `impact_contracts` 的对象类存在性（若启用 strict）       | 与 §4.3「记录完整性」一致；**不把契约当运行时拦截依据** | V2-5       |
| S4-3 | `**drivenadapters`**（行动类仓储）：若除 `bkn_raw` 外还有列拆分，补齐迁移脚本与读写映射               | `go test` / 冒烟导入；迁移可回滚或文档说明      | B1-3       |
| S4-4 | **OpenSearch / 索引**（若索引行动类正文）：评估新字段是否入索引；若入则重建或增量文档                       | worker/索引任务无误报                   | C3-4       |


---

## 阶段 5：OpenAPI 与对外 API 文档


| ID   | 任务                                                                                                                                                    | 验收标准                               | 依赖         |
| ---- | ----------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------- | ---------- |
| A5-1 | 在 `**adp/docs/api/bkn/bkn-backend-api/bkn.yaml`**（及复用同一 schema 的组件）为 `**action_type**` 增加 `**deprecated: true**`，`description` 指向 `**action_intent**` | Redoc/Swagger 可见删除线或 deprecated 提示 | B1-1       |
| A5-2 | 为 `**affect**`（或内联对象）增加 `**deprecated: true**`，指向 `**impact_contracts**`                                                                              | 与 §7.5 一致                          | B1-3       |
| A5-3 | 新增 `**action_intent**`、`**impact_contracts**` schema 定义；示例请求体更新                                                                                       | 与运行时 JSON 一致                       | A5-1, A5-2 |
| A5-4 | 若 **ontology-query** 或他服务**独立发布**含 ActionType 的 OpenAPI，同步 **deprecated** 与新字段                                                                        | `ontology-query.yaml` 等 diff 可审    | A5-3       |


---

## 阶段 6：测试与回归


| ID   | 任务                                                                           | 验收标准                             | 依赖         |
| ---- | ---------------------------------------------------------------------------- | -------------------------------- | ---------- |
| T6-1 | `**bkn_convert` 单测**：`action_intent` / `impact_contracts` / 扩展 `affect` 往返   | `go test` 通过                     | C3-3       |
| T6-2 | `**validate_action_type` 单测**：双写一致、非法枚举、`impact_contracts` 元素校验、strict 对象类引用 | 覆盖边界：空数组、单条、多条                   | V2-4, V2-5 |
| T6-3 | **集成/导入**：旧 BKN（仅 `action_type` + 旧 `affect`）导入成功；导出再导入含新字段                  | 与 CI 策略一致（`-tags=integration` 等） | C3-4, S4-1 |
| T6-4 | **显式负例**：请求体带 `context_scope` / `missing_root_policy` 等本期不落地字段时的行为符合 P0-3    | 单测或集成                            | P0-3       |


---

## 建议 PR 切分（降低一次改 >3 文件的心智负担）

1. **PR-1**：P0 + `bkn-spec` 版本 bump + `interfaces`（B1）+ 常量（B1-5）
2. **PR-2**：校验 V2 + `validate_action_type` 单测（T6-2 子集）
3. **PR-3**：`bkn_convert` + 单测（C3 + T6-1）
4. **PR-4**：Service / Access / 索引（S4）+ 集成测（T6-3）
5. **PR-5**：OpenAPI（A5）+ 文档链接 TASK/DESIGN
6. **PR-6**（可选）：ontology-query OpenAPI 同步（A5-4）

若必须单 PR，建议按 **B1 → V2 → C3 → S4 → A5 → T6** 顺序提交（commit）便于 review。

---

## 失败条件（整体验收不通过）

- `**action_intent` 与 `action_type` 可长期不一致**且无校验、无回填策略。  
- `**impact_contracts` 与扩展 `affect` 无法互转**或 `BKNRawContent` 往返丢失新语义。  
- **strict 下 `impact_contracts` 引用了不存在对象类**仍入库（与 §7.2 冲突）。  
- **OpenAPI 未对 `action_type`、`affect` 标注 `deprecated`**，或与 §7.5 叙事矛盾。  
- **本期误实现 `missing_root_policy` / 子图绑定结构**（超出 §7.3）。  
- `**cond` / `parameters` 校验或存储行为被无意改变**（违反 §7.4）。

---

## 参考路径（实现时自查）


| 区域       | 路径提示                                                                   |
| -------- | ---------------------------------------------------------------------- |
| DTO / 常量 | `adp/bkn/bkn-backend/server/interfaces/action_type.go`                 |
| 校验       | `adp/bkn/bkn-backend/server/driveradapters/validate_action_type.go`    |
| BKN 转换   | `adp/bkn/bkn-backend/server/logics/bkn_convert.go`（及 `_test.go`）       |
| 业务服务     | `adp/bkn/bkn-backend/server/logics/action_type/action_type_service.go` |
| OpenAPI  | `adp/docs/api/bkn/bkn-backend-api/bkn.yaml`                            |


