# 关系类「分侧过滤全连接」（filtered_cross_join）技术设计文档

> **状态**：草案  
> **版本**：0.2.1  
> **日期**：2026-04-08  
> **相关 Ticket**：#450

---

## 1. 背景与目标 (Context & Goals)

当前 BKN 关系类支持两种关联类型：

- **`direct`**：通过起点、终点对象类上的属性对建立映射，在查询展开时等价于 **INNER JOIN**（仅键能匹配的实例对形成边）。
- **`data_view`**：通过数据视图与 `source_mapping_rules` / `target_mapping_rules` 间接映射，展开时同样以 **键匹配** 为核心，语义上仍属于 INNER JOIN 类。

上述方式无法表达「在起点侧、终点侧分别施加过滤条件后，将**每一**起点实例与**每一**终点实例建立联系」的场景；该类语义在关系代数上对应 **在两侧 condition 约束下的笛卡尔积（cross join）**，而非按键连接。

**目标**：

1. 引入与 `direct`、`data_view` **平级**的第三种关系类型 **`filtered_cross_join`**（中文名：**分侧过滤全连接**）。
2. 在关系类定义中承载 **起点对象类**、**终点对象类** 的过滤条件（rule），用于子图查询展开时生成边。
3. 对展开结果规模设置 **默认可配置上限**（默认配额 **10_000** 条边），超限 **静默截断**；**仅**约束子图查询展开路径。
4. **不对**该关系类做关系实例物化（不落库边表）。

**非目标**：
- 不在本设计范围内规定「截断顺序」的具体实现算法细节以外的工程要求（但必须 **确定且稳定**，见第 3 节）。
- 不扩展其他 API（如关系类列表、通用图遍历）的默认行为以套用同一配额，除非另起需求。

---

## 2. 方案概览 (High-Level Design)

### 2.1 核心思路

1. **类型常量**：在关系类 `type` 上新增取值 `filtered_cross_join`，与 `direct`、`data_view` 并列；`mapping_rules` 的 JSON 结构 **仅** 在该类型下定义，不与 `[]Mapping`、`InDirectMapping` 混用。
2. **语义**：设起点条件为 \(C_s\)，终点条件为 \(C_t\)，在各自对象类实例上得到集合 \(S\)、\(T\)。展开边集为 **\(S \times T\)**：对任意 \(s \in S\)、\(t \in T\) 生成一条有向边 \((s, t)\)（方向遵循业务上起点→终点）。
3. **无匹配字段**：本类型 **不声明** 用于键对齐的字段对；「建立联系」的规则即为 **全连接**。
4. **规模**：设 \(N = |S| \times |T|\)。若 \(N\) 大于 **配额** \(Q\)（默认 10_000，可由 ConfigMap 配置），则 **仅保留至多 \(Q\) 条边**，其余 **静默丢弃**；**不返回错误**。
5. **作用域**：配额与截断 **仅** 应用于 **子图查询展开** 路径；其他接口不受该配额约束（需在实现中明确列出入口，避免误用）。

本方案 **不** 在 `data_view` 上叠加 outer/笛卡尔类标志（避免与键映射语义混杂、校验与文档难维护），**不** 通过 `direct` 虚构「恒真」键强行全连接（避免误导为 INNER JOIN、且扭曲属性语义）；而是采用 **独立类型 `filtered_cross_join`**，以 **分侧 condition（`CondCfg`）+ 子图展开配额 + 静默截断 + 不物化** 作为唯一表达方式，与 `direct` / `data_view` 清晰分离。

### 2.2 总体数据流（概念）

```text
RelationType(type=filtered_cross_join)
  mapping_rules: { source_condition, target_condition, ... }
                    |
                    v
子图查询展开（仅此处）
  1) 解析对象类，求 S = { instances | C_s }
  2) 求 T = { instances | C_t }
  3) 生成 S×T 的边序列（顺序稳定、可截断）
  4) 若 |S|×|T| > Q → 截断至 Q 条（静默）
```

### 2.3 关键决策

| 决策 | 说明 |
|------|------|
| 独立类型 | 与 `direct` / `data_view` 同级，避免在现有 `mapping_rules` 结构内嵌套导致兼容与校验复杂化。 |
| 不物化 | 关系类不存物理边；展开为查询时计算，降低存储与一致性成本。 |
| 静默截断 | 超限不报错，避免打断子图查询；需 **稳定截断顺序** 以保证可复现（见 3.4）。 |
| 配额可配置 | 默认 10_000，通过 **ConfigMap**（或等价配置）覆盖，便于多环境调优。 |
| 仅子图展开 | 控制面与性能面集中在子图场景，其他路径保持原有行为。 |

---

## 3. 详细设计 (Detailed Design)

### 3.1 类型与命名

| 项 | 值 |
|----|-----|
| 类型常量（建议） | `filtered_cross_join` |
| 中文名（并列展示） | **分侧过滤全连接** |
| 对外说明 | 在起点、终点集合上分别过滤后，对两集合做笛卡尔积式连边；非按键 INNER/OUTER JOIN。 |

### 3.2 `mapping_rules` 形态（草案）

在 `type === "filtered_cross_join"` 时，`mapping_rules` 为独立结构体（字段名以最终实现为准），至少包含：

| 字段 | 说明 |
|------|------|
| `source_condition` | 针对 **起点对象类**（`source_object_type_id`）实例的过滤条件，结构与平台现有 **condition** 一致（见 3.2.1）。 |
| `target_condition` | 针对 **终点对象类**（`target_object_type_id`）实例的过滤条件，同上。 |

**不包含** `source_property` / `target_property` 对键映射。

#### 3.2.1 `source_condition` / `target_condition` 与现有 condition 对齐

两侧条件 **复用** 当前对象检索等能力中已使用的 **condition** 结构，实现上与 `ontology-query`（及对齐场景下的 `bkn-backend`）中的 **`CondCfg`** 一致，便于复用解析、校验与 `NewCondition` 转换链路。

**代码参考**：`bkn/ontology-query/server/common/condition/entity.go` 中的 `CondCfg`。

**JSON 形态要点**（与现有 API 一致，非穷举；以代码为准）：

| JSON 字段 | 说明 |
|-----------|------|
| `operation` | 逻辑与/或：`and`、`or`；叶子条件为具体比较符（如 `==`、`in`、`range` 等，与 `condition` 包常量一致）。 |
| `field` | 属性技术名（与对象类字段对应）。 |
| `sub_conditions` | 嵌套子条件数组，用于 `and` / `or` 组合。 |
| `value` / `value_from` / `real_value` | 与 `ValueOptCfg` 一致，表示取值来源与比较值。 |
| `object_type_id` | 在行动等场景用于标注作用对象类；**关系类分侧条件**下通常由关系上的 `source_object_type_id` / `target_object_type_id` 限定作用域，若实现中允许省略则需在校验层写清规则。 |

对象检索请求体中的 `condition` 与上述结构同一套语义（参见 `ObjectQueryBaseOnObjectType` 中的 `condition` 与解析后的 `ActualCondition *CondCfg`）。

**校验**（与 `direct` / `data_view` 分支并列）：

- `source_object_type_id`、`target_object_type_id` 必填且与知识网络内对象类一致。
- `source_condition` / `target_condition` 须能被解析为合法 `CondCfg`，且其中 **field** 所引用的属性属于对应侧对象类（或平台允许的同一套规则）。
- 若条件非法或无法解析，在 **关系类创建/更新** 时拒绝（具体错误码与文案由接口规范统一）。

### 3.3 配额与配置

| 项 | 说明 |
|----|------|
| 配置项 | 例如 `filtered_cross_join.max_edge_expand`（或项目统一命名规范下的键名），默认 **10000**。 |
| 配置来源 | **ConfigMap**（或与现有 `bkn-backend-config.yaml` / 服务配置一致的方式），支持按环境调整。 |
| 语义 | 表示 **单关系类、单子图展开请求** 中，该类型可生成的 **最大边条数**（截断上限）。 |

### 3.4 截断顺序（必须）

静默截断 **必须** 使用 **确定性** 的全序，否则同一查询多次结果不一致。建议约定（实现时选其一并文档化）：

- 先按 **起点实例主键**（或稳定 id）升序，再按 **终点实例主键** 升序，依次取边直至达到 \(Q\)；或  
- 等价的全序定义，保证与排序键可复现。

**不要求**向调用方返回「已截断」标志（当前为静默）；若运维/排障需要，可在内部可选 **debug 日志** 或指标中记录 `truncated`、`raw_cartesian_size`（不纳入本设计对外 API 承诺）。

### 3.5 行为示例

| 场景 | 行为 |
|------|------|
| \(\|S\|=100, \|T\|=100\) | \(N=10000\)，若 \(Q=10000\)，不截断。 |
| \(\|S\|=101, \|T\|=100\) | \(N=10100\)，若 \(Q=10000\)，**静默截断为 10000 条**，**不失败**。 |
| \(S=\emptyset\) 或 \(T=\emptyset\) | \(N=0\)，无边，不视为超限。 |
| 配置 \(Q=5000\) | 同上逻辑，截断至 5000。 |

### 3.6 与「子图查询」的边界

- **必须**：仅在 **子图查询展开** 路径读取 `filtered_cross_join` 的 `mapping_rules` 并执行 \(S \times T\) 与截断。
- **必须**：其他接口（如仅返回关系类元数据、不做实例展开）**不**应用该配额逻辑。
- 若存在多个服务（如 `bkn-backend` 与 `ontology-query`），**配置项默认值与语义应对齐**，或明确单一配置源，避免环境间行为不一致。

---

## 4. 风险与边界 (Risks & Edge Cases)

| 风险 / 边界 | 说明与缓解 |
|-------------|------------|
| **静默截断导致认知偏差** | 用户可能误以为图上仅有截断后的边；产品文档需说明「分侧过滤全连接」与配额；必要时在内部指标中暴露截断事实。 |
| **单侧集合过大** | 即使 \(N\) 被截断，求 \(S\) 或 \(T\) 仍可能扫描大量实例；需实现上考虑 **流式/分页** 与 **上限**（可另起性能项，非本设计必须）。 |
| **配置不一致** | 多副本、多服务若未共享同一配置，可能出现「同请求不同截断」；通过统一配置源与发布流程缓解。 |
| **条件复杂度** | 复杂 condition 可能导致展开超时；与通用查询超时、限流策略一致处理。 |
| **一对多 / 重复** | 笛卡尔积下每对 \((s,t)\) 至多一条边；若业务需要多重边，不在本类型范围内。 |

---

## 5. 任务拆分 (Milestones)

- [ ] 在接口层（`interfaces`）增加 `RELATION_TYPE_FILTERED_CROSS_JOIN` 及 `mapping_rules` 结构体定义。  
- [ ] `validate_relation_type`（及等价校验）增加分支：`source_condition` / `target_condition` 与 `CondCfg` 合法性、对象类引用。  
- [ ] 子图查询展开路径：识别该类型，计算 \(S\)、\(T\)，生成 \(S \times T\)，按配置截断。  
- [ ] ConfigMap / 配置文件增加配额项，默认 10000，并接入读取逻辑。  
- [ ] 单元测试：截断边界、空集、配置覆盖；集成测试：子图查询仅在此路径生效。  
- [ ] 文档：对外 API / 用户手册中补充「分侧过滤全连接」与配额、静默截断说明。

---

## 参考

- 现有关系类型与映射结构：`bkn/bkn-backend/server/interfaces/relation_type.go`、`bkn/ontology-query/server/interfaces/relation_type.go`  
- Condition 结构（`CondCfg`）：`bkn/ontology-query/server/common/condition/entity.go`；`bkn-backend` 侧：`bkn/bkn-backend/server/common/condition/entity.go`（与数据视图等场景对齐时使用）。  
- 对象检索中的 condition 用法：`bkn/ontology-query/server/interfaces/object_type.go`（`ObjectQueryBaseOnObjectType`）。  
- 同类设计文档格式参考：`docs/design/bkn/features/bkn_docs/DESIGN.md`  

---

## 验收清单（建议）

- [ ] `type` 为 `filtered_cross_join` 时，可独立保存 `mapping_rules`，且不与 `direct`/`data_view` 结构混淆。  
- [ ] 子图查询展开在 \(N > Q\) 时返回 **至多 \(Q\)** 条边，**无错误**；\(N \le Q\) 时**不截断**。  
- [ ] 配额可通过配置修改（如改为 5000），行为随之变化。  
- [ ] 非子图路径**不**应用该截断逻辑。  
- [ ] 关系类**无**物化边存储（与现有「不物化」决策一致）。  

## 失败条件（建议）

- 子图展开在 \(N > Q\) 时返回 **错误**（与「静默截断」冲突）→ **失败**。  
- 截断顺序 **随机** 或 **不可复现** → **失败**。  
- 其他 API 误用该配额导致非预期截断或限流 → **失败**（按产品定义修正边界）。  
