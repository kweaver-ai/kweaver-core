# 测试架构设计：基于 Commit 与多智能体

本文档描述以**研发 Commit 为触发**、由**多智能体协作**完成的测试架构，实现：用例增删改、多维度测试分析、按优先级筛选执行、测试报告生成。测试框架沿用现有 DIP AT 规范（见 [测试框架架构_全自动与智能体规范.md](测试框架架构_全自动与智能体规范.md)）。

---

## 一、设计目标与原则

### 1.1 设计目标

| 目标 | 说明 |
|------|------|
| **Commit 驱动** | 根据研发的 commit（变更路径、类型、受影响 API）触发用例维护与测试执行 |
| **智能体维护用例** | 结合智能体对现有 case 进行增、删、改，保证用例与代码/接口同步 |
| **多维度测试分析** | 除常规功能测试外，分析是否需要：契约测试、边界测试、压力测试、性能测试等 |
| **分级执行** | 先保证新增功能正常 → 再回归相关 case → 按评估决定是否执行边界/性能等 |
| **报告产出** | 测试完成后生成统一报告（Allure/JUnit 等） |

### 1.2 设计原则

- **框架复用**：用例结构、套件规范、`get_cases`、执行入口、报告机制均沿用现有 DIP AT 框架。
- **规范即契约**：智能体仅依据 `case_schema`、`suite_schema`、`apis.yaml`、`global_manifest`、`path_scope_mapping` 进行增删改与筛选。
- **单条用例粒度**：筛选与执行粒度到单条 case（name/tags/api_name/api_path），不整 suite 盲目执行。
- **多智能体分工**：不同职责由不同智能体承担，通过标准化输入输出衔接流水线。

---

## 二、整体架构与流水线

### 2.1 端到端流水线

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│  研发仓库 Push/Commit                                                                     │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│  【提交监听服务】 变更路径、变更类型、diff、受影响 API/文件                                 │
│  技术实现（对接 GitHub）：见 [提交监听服务_基于GitHub实现.md](提交监听服务_基于GitHub实现.md)   │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌──────────────────────────┐ ┌──────────────────────────┐ ┌──────────────────────────┐
│ 智能体 A：用例维护智能体   │ │ 智能体 B：测试维度分析智能体 │ │ （共用：path_scope_mapping  │
│ - 输入：commit 上下文 +   │ │ - 输入：commit + 受影响API  │ │  + 受影响 API 列表）       │
│   建议清单 + 现有 case    │ │   + 现有 case tags/类型    │ │                            │
│ - 输出：增/删/改 suites   │ │ - 输出：需执行的测试维度   │ │                            │
│   *.yaml（符合 Spec）     │ │   (functional/contract/   │ │                            │
│                          │ │   boundary/stress/perf)   │ │                            │
└──────────────────────────┘ └──────────────────────────┘ └──────────────────────────┘
                    │                   │
                    └───────────────────┼───────────────────┐
                                        ▼                   │
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 智能体 C：用例筛选与执行智能体                                                             │
│ - 输入：测试维度 + 受影响 API/scope/tags + 执行优先级策略                                   │
│ - 逻辑：先筛「新增功能相关」→ 再筛「回归相关」→ 按维度评估筛「边界/压力/性能」等              │
│ - 输出：CASE_NAMES 或 (SCOPE/TAGS/API_NAME/API_PATH) + 执行顺序/批次                        │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 【执行器】get_cases(...) 筛出单条用例 → pytest 执行（现有 main.py / pytest）                │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 智能体 D：报告生成智能体（可选增强）                                                         │
│ - 输入：Allure/JUnit 结果、执行维度、通过率、失败 case                                      │
│ - 输出：测试报告（HTML/汇总说明/结论建议）                                                  │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流概要

| 阶段 | 输入 | 输出 |
|------|------|------|
| 提交监听 | Git push / webhook | 变更路径、变更类型(added/modified/deleted)、diffs、**可选：变更文件完整内容**、受影响 api_name/api_path、scope、suggested_suites、need_add_cases（详见 [提交监听服务_基于GitHub实现](提交监听服务_基于GitHub实现.md)） |
| 智能体 A | 上述输出 + 本模块 apis/global_manifest/spec；**建议含 diff + 变更文件完整内容**以便理解提交后补用例；**写回前需校验分支 tip 是否仍为当前 commit_sha**，若已有新提交可放弃写回（详见《提交监听服务_基于GitHub实现》3.6 节） | 对 suites/*.yaml 的增/删/改（PR 或直接提交） |
| 智能体 B | commit + 受影响 API + **diff/变更文件内容** + case 的 tags/类型 | 本轮需执行的维度：functional, contract, boundary, stress, performance 等 |
| 智能体 C | 维度 + scope/tags/api + 优先级策略 | CASE_NAMES 或 SCOPE/TAGS/API_NAME/API_PATH，及执行批次顺序 |
| 执行器 | 筛选条件 | 只跑筛出的单条用例，产出 Allure/JUnit |
| 智能体 D | 执行结果文件 | 最终测试报告 |

---

## 三、多智能体职责与规范

### 3.1 智能体 A：用例维护智能体（Case Maintenance Agent）

**职责**：根据 commit 与建议清单，对现有 case 进行增、删、改。

| 项目 | 说明 |
|------|------|
| **触发** | 提交监听输出 `need_add_cases=true` 或变更类型为 added/modified/deleted 且命中 path_scope_mapping |
| **输入** | 变更路径、变更类型、diffs、**建议：变更文件完整内容（commit 后）**、受影响 api_name/api_path、suggested_suites、本模块 base_dir；无需默认提供完整代码仓库，详见《提交监听服务_基于GitHub实现》第七节 |
| **规范依据** | `testcase/_config/spec/case_schema.yaml`、`suite_schema.yaml`；本模块 `apis.yaml`、`global_manifest.yaml`、`path_scope_mapping.yaml` |
| **动作** | 新增 case：在 suggested_suites 对应 `suites/*.yaml` 的 cases 中追加，url 用 apis.yaml 的 name，变量用 global_manifest 的 ref；修改 case：按 name 替换同名字段；删除/下线：从 cases 中移除或 suite switch=n |
| **输出** | 变更后的 `suites/*.yaml`（符合 Spec），并为新 case 打上细粒度 tags（如 api:xxx、smoke、regression）便于后续筛选 |

**与现有框架的衔接**：完全遵循《测试框架架构_全自动与智能体规范》第八、九节；不引入未在 global_manifest 中声明的变量。

---

### 3.2 智能体 B：测试维度分析智能体（Test Dimension Analyzer Agent）

**职责**：分析本次 commit 除常规功能测试外，是否需要进行契约、边界、压力、性能等维度的测试。

| 维度 | 说明 | 识别依据（示例） | 对应 tags/scope 约定 |
|------|------|------------------|----------------------|
| **功能测试** | 常规 API 行为、业务场景 | 所有代码/接口变更 | smoke, regression |
| **契约测试** | 请求/响应结构、字段契约 | 接口定义变更、OpenAPI/Proto 变更、新增/修改请求体或响应体 | contract, regression；case 含 resp_schema |
| **边界测试** | 参数边界、长度、枚举临界值 | 参数校验逻辑、长度限制、枚举变更 | boundary；建议 case tags 增加 boundary |
| **压力测试** | 高并发、限流、熔断 | 并发/限流/熔断相关代码路径 | stress；建议 case/suite tags 增加 stress |
| **性能测试** | 时延、吞吐、资源占用 | 性能相关路径或 path_scope_mapping 中 performance_tags | performance；TAGS=performance 或 SCOPE=performance |

**输入**：commit 变更路径、变更类型、受影响 api_name/api_path、变更描述/diff 摘要；**建议包含变更文件完整内容（commit 后）**以便理解参数校验、接口契约等并判断维度；本模块 case 的 tags 及是否含 resp_schema/边界场景。无需默认提供完整代码仓库。

**输出**：本轮需执行的维度列表，例如：

```yaml
dimensions:
  - functional   # 必选
  - contract     # 若接口定义或响应结构有变更
  - boundary     # 若存在参数/长度/枚举相关变更
  - stress       # 若存在限流/并发相关变更
  - performance  # 若 path_scope_mapping 或改动涉及性能
```

**与现有框架的衔接**：  
- 在模块 `path_scope_mapping.yaml` 的 `test_type_tags` 中扩展（见 3.5 节），例如增加 `boundary`、`stress`。  
- 用例中通过 **tags** 标记维度（如 `boundary`、`stress`、`performance`），执行时通过 `get_cases(tags=[...])` 或环境变量 `TAGS=boundary,stress` 筛选。

---

### 3.3 智能体 C：用例筛选与执行智能体（Case Selector & Executor Agent）

**职责**：根据“先新增功能、再回归、再按需边界/压力/性能”的原则，筛出本轮要执行的 case，并给出执行顺序或批次。

**执行原则（优先级）**：

1. **第一优先级：新增功能相关**  
   与本次 commit 直接相关的 case（新增 API、修改过的 API 对应 case）。  
   筛选方式：`get_cases(api_name=..., api_path=..., names=[...])` 或提交监听给出的 affected_api_names/affected_api_paths → 得到「本次直接受影响」的 case 列表。

2. **第二优先级：回归相关**  
   与受影响模块/scope 相关的回归 case，排除已在前一步执行过的。  
   筛选方式：`get_cases(scope=..., tags=[regression])` 或 path_scope_mapping 的 scope_tags，再减去第一优先级已选 case。

3. **第三优先级：按维度评估执行**  
   由智能体 B 输出的维度（contract/boundary/stress/performance），评估是否执行：  
   - 若本次有接口契约变更 → 执行 tags 含 contract 或含 resp_schema 的 case；  
   - 若有参数/边界变更 → 执行 tags 含 boundary 的 case；  
   - 若有并发/限流变更 → 执行 tags 含 stress 的 case；  
   - 若有性能相关变更或配置建议 → 执行 `TAGS=performance` 或 `SCOPE=performance`。  

   筛选方式：`get_cases(tags=[contract|boundary|stress|performance])` 或对应 SCOPE/API_NAME 组合，且可与 scope 相交缩小范围。

**输入**：提交监听输出（scope、affected_api_names、affected_api_paths、suggested_suites）；智能体 B 输出的 dimensions；本模块 base_dir。

**输出**：  
- 方案一：**CASE_NAMES**（逗号分隔的 case name 列表），按优先级分批次时可为多组 CASE_NAMES。  
- 方案二：多轮执行参数，每轮一组 (SCOPE, TAGS, API_NAME, API_PATH)，由流水线依次执行并合并结果。

**与现有框架的衔接**：  
- 执行时 `CASE_NAMES=... pytest` 或 `TAGS=... SCOPE=... pytest`，见 README 与《测试框架架构_全自动与智能体规范》第九节。  
- 仅跑筛出的单条用例，不整 suite 全量执行。  
- **重要**：使用 CASE_NAMES 或 names 筛选时，必须包含所有 prev_case 依赖链上的用例，否则前置数据缺失会导致结果错误

---

### 3.4 智能体 D：报告生成智能体（Report Generator Agent，可选）

**职责**：在测试执行完成后，汇总结果并生成可读的测试报告。

| 项目 | 说明 |
|------|------|
| **输入** | Allure 结果（report/xml）、JUnit（report/junit_report.xml）、执行维度、批次信息、通过/失败 case 列表 |
| **逻辑** | 解析通过率、失败用例、失败原因；按维度（功能/契约/边界/压力/性能）汇总；给出结论与建议（如：是否需人工介入、是否建议补测） |
| **输出** | 测试报告（HTML 或 Markdown）、摘要、结论与后续建议 |

可与现有 Allure Report 并存：Allure 提供明细，智能体 D 提供高层汇总与自然语言结论。

---

### 3.5 测试维度在框架中的约定（扩展 path_scope_mapping）

在现有 `test_type_tags` 基础上，建议在各模块 `path_scope_mapping.yaml` 中扩展，供智能体 B/C 统一识别：

```yaml
# 测试类型与 tags 约定（智能体识别用）
test_type_tags:
  functional: [regression, smoke]
  contract: [regression, contract]      # 契约/结构校验（resp_schema 等）
  boundary: [boundary, regression]       # 边界测试
  stress: [stress]                       # 压力/并发测试
  performance: [performance]             # 性能测试
```

用例与套件中通过 **tags** 标记维度后，执行时即可用 `TAGS=boundary`、`TAGS=stress`、`TAGS=performance` 等与现有 `get_cases(tags=...)` 一致的方式筛选。

---

## 四、与现有测试框架的对接

### 4.1 复用内容

| 项目 | 说明 |
|------|------|
| 用例/套件规范 | `testcase/_config/spec/case_schema.yaml`、`suite_schema.yaml` |
| 模块配置 | `apis.yaml`、`global.yaml`、`global_manifest.yaml`、`path_scope_mapping.yaml`、`suite_manifest.yaml` |
| 筛选 API | `get_cases(base_dir, scope=..., tags=..., api_name=..., api_path=..., names=...)`，粒度到单条用例 |
| 执行入口 | `pytest`、`main.py`；环境变量 `CASE_NAMES`、`SCOPE`、`TAGS`、`API_NAME`、`API_PATH` |
| 报告 | Allure（report/xml → report/html）、JUnit（report/junit_report.xml） |
| 按 name 执行 | 智能体输出 CASE_NAMES 后，`CASE_NAMES=name1,name2 pytest` 或 `python main.py` |

### 4.2 需扩展或约定内容

| 项目 | 说明 |
|------|------|
| case/suite tags | 在现有 smoke、regression、performance 基础上，约定 **boundary**、**stress**、**contract**，并在 case_schema 说明中体现 |
| path_scope_mapping | 在 `test_type_tags` 中增加 boundary、stress（可选 contract），见 3.5 节 |
| 提交监听输出 | 需稳定提供：变更路径、变更类型、affected_api_names、affected_api_paths、scope、suggested_suites、need_add_cases，以便智能体 A/B/C 消费 |
| 执行顺序 | 若分多批次（先功能再回归再维度），流水线需支持多轮调用 get_cases + pytest，并合并 Allure/JUnit 结果供智能体 D 使用 |

---

## 五、执行策略小结

| 顺序 | 策略 | 筛选方式 | 说明 |
|------|------|----------|------|
| 1 | 新增功能 | api_name / api_path / names（与 commit 直接相关） | 先保证本次改动涉及的功能正确 |
| 2 | 回归 | scope + tags=regression，排除已执行 | 再保证相关模块回归 |
| 3 | 契约/边界/压力/性能 | 按智能体 B 的 dimensions，tags=contract/boundary/stress/performance 或 SCOPE=performance | 按评估决定是否跑、跑哪些 |

报告在全部执行完成后由 Allure 生成，并由智能体 D（可选）生成汇总与结论。

---

## 六、文档与规范索引

- **用例与套件规范、get_cases、执行方式**：[测试框架架构_全自动与智能体规范.md](测试框架架构_全自动与智能体规范.md)
- **运行方式、按条件运行与提取、智能体对接**：项目根目录 [README.md](../README.md)
- **本架构**：多智能体分工、测试维度分析、执行优先级、报告生成，与上述规范一致并在此基础上扩展。

---

## 七、小结

- **Commit 驱动**：提交监听产出变更与受影响 API，驱动用例维护与多维度分析。  
- **多智能体**：A 维护用例，B 分析测试维度，C 按优先级筛选用例并驱动执行，D 汇总报告。  
- **测试框架**：沿用现有 DIP AT 的 Spec、get_cases、单条用例粒度、CASE_NAMES/TAGS/SCOPE 执行与 Allure/JUnit 报告。  
- **多维度**：功能、契约、边界、压力、性能通过 tags 与 test_type_tags 约定，由 B 分析、C 筛选执行。  
- **执行原则**：先新增功能相关 case，再回归相关 case，最后按评估执行边界/压力/性能等维度；完成后生成报告。
