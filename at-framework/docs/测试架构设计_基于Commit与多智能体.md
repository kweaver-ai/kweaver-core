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
- **模板统一、vega 为产物**：所有模板统一放在 `testcase/_config/spec/` 下；`testcase/vega/` 为按模板生成的产物，其下配置与用例（含 `path_scope_mapping.yaml`、`suites/*.yaml`、`apis.yaml` 等）均由智能体通过模板生成或更新——**不存在则从模板创建或追加，存在则进行增删改**，无需人工手写维护。
- **多智能体分工**：不同职责由不同智能体承担，通过标准化输入输出衔接流水线。

---

## 二、整体架构与流水线

### 2.1 端到端流水线总览

端到端流水线分为**两段**：**主控 workflow**（拉代码、按模板生成配置、产出 listener_output）与**下游编排**（接收 listener_output → A/B 并行 → C → 部署 → 执行 → D）。全流程无人工介入。

**阶段一览**：

| 阶段 | 说明 |
|------|------|
| **一、主控 workflow** | 上游触发传入 repo/branch；主控拉取代码、同步测试仓库、按模板+脚本生成/更新各模块 path_scope_mapping（可选合并 mapping 智能体 supplement）、从 info.json 生成 listener_output 并发布。详见 2.2 节。 |
| **二、下游编排** | 内网 CI 接收 listener_output.json，按规则编排：need_add_cases 则启 A，始终启 B；A、B 并行 → 汇聚 → C 产出执行批次 → 拉包部署 → 按批次执行 pytest → D 产出报告。详见 2.3 节。 |

**流程图**：

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│  研发仓库 Push/Commit（GitHub）→ 上游触发（如内网镜像构建 pipeline）                          │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│  【阶段一：主控 workflow】 拉取代码（上游 repo/branch）→ 同步测试仓库 → 按模板生成/更新        │
│  path_scope_mapping（所有模块，可选 mapping 智能体 supplement）→ 从 info.json 生成 listener_output → 发布 │
│  格式见 2.7 节                                                                             │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│  【阶段二：下游编排】 接收 listener_output → A/B 并行 → C → 部署 → 执行 → D                   │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                    ┌───────────────────┼───────────────────┐
                    ▼                   ▼                   ▼
┌──────────────────────────┐ ┌──────────────────────────┐ ┌──────────────────────────┐
│ 智能体 A：用例维护智能体   │ │ 智能体 B：测试维度分析智能体 │ │ （共用：path_scope_mapping  │
│ - 输入：listener_output + │ │ - 输入：listener_output + │ │  + 受影响 API 列表）       │
│   建议清单 + 现有 case    │ │   case tags/类型          │ │                            │
│ - 输出：增/删/改 suites   │ │ - 输出：测试维度           │ │                            │
│   *.yaml（符合 Spec）     │ │   (functional/contract/   │ │                            │
│                          │ │   boundary/stress/perf)   │ │                            │
└──────────────────────────┘ └──────────────────────────┘ └──────────────────────────┘
                    │                   │
                    └───────────────────┼───────────────────┐
                                        ▼                   │
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 智能体 C：用例筛选与执行智能体                                                             │
│ - 输入：listener_output + B 的 dimensions + 执行优先级策略                                  │
│ - 输出：CASE_NAMES 或 (SCOPE/TAGS/API_NAME/API_PATH) + 执行顺序/批次                        │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 【自动化部署】按 listener_output 的 repo/branch/commit_sha 拉取构建产物 → 部署 → 健康检查    │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 【执行器】按 C 的 batches 调用 get_cases(...) → pytest 执行 → 结果写入 report/             │
└─────────────────────────────────────────────────────────────────────────────────────────┘
                                        │
                                        ▼
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│ 智能体 D：报告生成智能体（可选）                                                             │
│ - 输入：Allure/JUnit 结果、执行维度、通过率、失败 case                                      │
│ - 输出：测试报告（HTML/汇总说明/结论建议）                                                  │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### 2.2 主控 workflow（产出 listener_output）

主控 workflow 负责：拉取上游指定仓库与分支、同步测试代码仓库、按模板生成/更新**所有模块**的 path_scope_mapping、从 info.json 生成符合 2.7 节格式的 listener_output 并发布，供下游编排消费。**无人工介入**。

**步骤**：

| 步骤 | 说明 |
|------|------|
| 1. 拉取代码 | 按上游传入的 repo、branch 拉取被测仓库（ghh 下载或 checkout）。 |
| 2. 同步测试代码仓库 | 同步测试代码仓库（如 kweaver-ai/kweaver，默认分支 kweaver-at），用于模板与脚本。 |
| 3. 按模板生成/更新 path_scope_mapping | 调用测试仓库中的 `scripts/ensure_path_scope_mapping.py`：从模板+仓库扫描为所有模块生成/更新结构；若已接入 **mapping 智能体** 并产出 supplement，可传入 `--supplement` 合并业务字段（suggested_suites、path_to_api 等）。 |
| 4. 收集 commit 信息 | commit_message 从流程中下载的 info.json 读取；repo、branch、commit_sha 从参数或 git/commit.txt 取得。 |
| 5. 生成 listener_output | 用 info.json（GitHub 格式，含 changed_files）与 path_scope_mapping 调用 `generate_listener_output.py`，产出符合 2.7 节的 listener_output.json。 |
| 6. 发布 | 将 listener_output.json 发布为附件，供下游 CI 接收。 |

**实现**：`kweaver-github-build` 中主控 workflow 为 `test-pipeline/commit-e2e-main-workflow.yaml`；模板统一在 `testcase/_config/spec/`，vega 等为按模板生成的产物（见 3.6.1 节）。

### 2.3 下游编排（接收 listener_output 后）

编排由**内网 CI/DevOps** 完成，规则驱动。流程固定、可预测、易调试。

| 步骤 | CI/DevOps 动作 |
|------|----------------|
| 1. 接收 | 从主控 workflow 发布处获取 listener_output.json（文件、MQ、API 等） |
| 2. 分发 | 若 `need_add_cases=true` 则启动 job A，否则 A 快速退出；**始终**启动 job B；A、B 并行执行 |
| 3. 汇聚 | 等待 A、B 完成；A 产出 suites 变更（PR/提交），B 产出 dimensions JSON |
| 4. 筛选 | 启动 job C，传入 listener 输出 + B 的 dimensions；C 产出执行批次 JSON |
| 5. 部署 | **自动化部署**：按 listener_output 的 repo/branch/commit_sha 拉取内网构建产物（镜像或包），部署至测试环境并做健康检查，就绪后再进入执行 |
| 6. 执行 | 按 C 的 batches 顺序，每批设置 CASE_NAMES 或 SCOPE/TAGS 等，调用 pytest；结果写入 report/ |
| 7. 报告 | 启动 job D，传入 report 路径 + 执行上下文；D 产出汇总报告 |
| 8. 降级 | 任一步超时/失败时，按 4.5 节策略降级，不阻塞后续步骤 |

**编排示例**（以 Jenkins/流水线脚本为例）：

```
stage('接收') → 读取 listener_output.json
stage('分发') → 并行触发 agent-a、agent-b
stage('汇聚') → 等待 A、B
stage('筛选') → 触发 agent-c
stage('部署') → 拉包/拉镜像并部署至测试环境，健康检查就绪
stage('执行') → 按 batches 循环跑 pytest
stage('报告') → 触发 agent-d
```

智能体 A 与 B **可并行**执行。智能体 C 依赖 A、B 的输出。推荐顺序：**listener_output → A、B 并行 → C → 自动化部署（拉包部署）→ 执行器 → D**。

### 2.4 备选：GitHub Actions 编排

若研发与测试均在 GitHub 上、且被测环境可达，可用 GitHub Actions 替代内网 DevOps 编排。主控 job 产出 listener_output artifact，下游 job 按 2.3 节步骤编排。详见《提交监听服务_基于GitHub实现》第四节。

### 2.5 多模块提交且存在依赖时的部署与测试

当**一次提交涉及多个模块**，且**模块间有构建/运行依赖**（如 B 依赖 A 的接口或包）时，部署顺序与测试范围需按依赖关系处理。

**场景**：

- 单次 push 改动多个目录（如 `vega/data-connection` 与 `vega/vega-gateway`），或 monorepo 下多模块同时变更。
- 模块依赖关系：例如 gateway 依赖 data-connection 的 API，部署时需先部署 data-connection，再部署 gateway；测试时需在「全部就绪」后的环境上跑。

**原则**：

1. **部署**：按**依赖图拓扑序**分批部署——先部署无依赖或依赖已就绪的模块，同层可并行；再部署依赖它们的模块，直到所有受影响模块就绪。
2. **测试**：在**整批部署完成并健康检查通过**后，再按智能体 C 产出的批次执行用例；测试范围覆盖所有受影响 scope 的并集，用例可跨模块。

**实现要点**：

| 环节 | 做法 |
|------|------|
| **受影响模块** | 由 listener_output 的 `scopes`（或按 changed_files 命中 path_scope_mapping 的 scope）得到「本次涉及的模块/scope 列表」；若为多仓库则需各仓库的 listener 汇聚或上游传入模块列表。 |
| **依赖关系** | 在仓库或编排侧维护**模块依赖配置**（如 `module_dependencies.yaml` 或与 path_scope_mapping 同级的 `deploy_dependencies`），声明 scope/module 之间的依赖（如 `vega-gateway` 依赖 `vega-data-connection`）。 |
| **部署顺序** | 对「受影响模块」做**拓扑排序**（仅保留本次涉及的节点与边），得到层级：第 0 层 = 无依赖或依赖不在本次范围内的模块；第 1 层 = 仅依赖第 0 层；依此类推。按层顺序部署，同一层内多模块可并行拉包、并行部署。 |
| **拉包与部署** | 每层内按 repo/branch/commit_sha（或各模块对应构建产物标签）拉取镜像或包，部署到测试环境；每层部署完成后做该层模块的健康检查，再进入下一层。 |
| **测试** | 全部层部署并健康检查通过后，再执行智能体 C 产出的 batches（新增功能 → 回归 → 按维度）；C 的 scope/tags/api 可能跨多个模块，执行器在同一环境上按批次跑即可。 |

**依赖配置示例**（可选，放在仓库根或 CI 配置中）：

```yaml
# module_dependencies.yaml 或集成到 path_scope_mapping
deploy_order:
  - scope: vega-data-connection   # 无依赖，先部署
  - scope: vega-gateway
    depends_on: [vega-data-connection]
  - scope: vega-web
    depends_on: [vega-gateway, vega-data-connection]
```

**编排流程（多模块有依赖）**：

```
接收 listener_output（含 scopes）
       ↓
解析受影响模块 + 读取 module_dependencies → 拓扑排序得到 [层0, 层1, …]
       ↓
A、B 并行（与单模块相同）
       ↓
汇聚 → 启动 C，产出执行批次（可能含多 scope）
       ↓
按层部署：for 每一层: 并行拉包/部署该层各模块 → 健康检查 → 下一层
       ↓
全部部署就绪 → 按 C 的 batches 执行 pytest
       ↓
报告 D
```

**降级**：若某层部署或健康检查失败，可标记该层及依赖其的后续层为失败，并跳过测试执行（或仅跑已就绪模块对应的 scope，见 4.5 节）。

### 2.6 数据流概要

| 阶段 | 输入 | 输出 |
|------|------|------|
| **主控 workflow** | 上游传入 repo/branch；拉取代码、同步测试仓库、按模板生成 path_scope_mapping、从 info.json 生成 listener_output | listener_output.json（格式见 2.7 节），发布供下游消费 |
| **下游编排** | 主控产出的 listener_output.json | 按规则编排：need_add_cases 则调 A；A、B 并行；汇聚后调 C；**拉包部署**；按 C 的 batches 跑 pytest；最后调 D |
| 智能体 A | CI 转发的 listener 输出 + 本模块 apis/global_manifest/spec；**建议含 diff + 变更文件完整内容**；**写回前需校验分支 tip 是否仍为当前 commit_sha**（详见《提交监听服务_基于GitHub实现》3.6 节） | 对 suites/*.yaml 的增/删/改（PR 或直接提交） |
| 智能体 B | CI 转发的 listener 输出 + case 的 tags/类型 | dimensions JSON（functional 必含，其余按需） |
| 智能体 C | CI 汇聚的 listener 输出 + B 的 dimensions | 执行批次 JSON（mode + batches） |
| **自动化部署** | listener_output（repo/branch/commit_sha）+ 内网镜像/包仓库地址 | 拉取对应构建产物，部署至测试环境，健康检查通过后标记就绪 |
| 执行器 | CI 按 C 的 batches 设置的 CASE_NAMES 或 SCOPE/TAGS 等 | 只跑筛出的单条用例，产出 Allure/JUnit 到 report/ |
| 智能体 D | CI 传入的 report 路径 + 执行上下文 | 最终测试报告 |

**主控 workflow 输入**：上游触发时传入 `repo`、`branch`（及可选 source、test_repo、test_branch）；commit 相关数据来自拉取后的 info.json 与 git/commit.txt。产出 listener_output.json 格式见 2.7 节。

### 2.7 listener_output 格式（主控 workflow 产出）

主控 workflow 产出的**标准化 JSON**，供下游编排与智能体 A/B/C 消费。格式约定如下：

```json
{
  "repo": "owner/repo",
  "branch": "main",
  "commit_sha": "abc123...",
  "commit_message": "feat: add datasource test API",
  "commit_url": "https://github.com/owner/repo/commit/abc123",
  "changed_files": [
    { "path": "vega/data-connection/src/controller/DatasourceController.java", "status": "modified" },
    { "path": "vega/data-connection/src/service/DatasourceService.java", "status": "modified" }
  ],
  "change_summary": {
    "added": ["path/a.go"],
    "modified": ["vega/data-connection/..."],
    "removed": []
  },
  "scopes": ["vega-data-connection"],
  "scope_tags": ["regression", "data-connection"],
  "suggested_suites": ["测试数据源", "新增数据源", "查询数据源"],
  "affected_api_names": ["测试数据源连接", "新增数据源"],
  "affected_api_paths": ["/api/data-connection/v1/datasource/test", "/api/data-connection/v1/datasource"],
  "need_add_cases": true,
  "diffs": [{ "path": "vega/.../X.java", "patch": "@@ -1,3 +1,5 @@\n..." }]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `repo` | string | 是 | 仓库标识，如 `owner/repo` |
| `branch` | string | 是 | 分支名 |
| `commit_sha` | string | 是 | 本次 push 的 commit SHA |
| `commit_message` | string | 是 | 提交信息 |
| `commit_url` | string | 否 | 可点击的 commit 链接 |
| `changed_files` | array | 是 | 变更文件列表，每项含 `path`、`status`（added/modified/removed） |
| `change_summary` | object | 否 | 按类型汇总的路径：`added`、`modified`、`removed` |
| `scopes` | array | 是 | 命中 path_scope_mapping 的 scope 列表 |
| `scope_tags` | array | 否 | 命中 scope 的 tags 并集 |
| `suggested_suites` | array | 是 | 建议涉及的 suite 名称，供智能体 A 增删改用例 |
| `affected_api_names` | array | 是 | 受影响的 API 名称（对应 apis.yaml） |
| `affected_api_paths` | array | 是 | 受影响的 API 路径 |
| `need_add_cases` | boolean | 是 | 是否建议新增用例（如存在 added 且命中 scope） |
| `diffs` | array | 否 | 每个变更文件的 patch，每项含 `path`、`patch` |

**可选扩展**：为便于智能体理解提交上下文，可额外提供**变更文件完整内容**（commit 后版本），方式见 [提交监听服务_基于GitHub实现](提交监听服务_基于GitHub实现.md) 第七节。

---

## 三、多智能体职责与规范

### 3.0 mapping 智能体（Path-Scope Mapping Agent）

**职责**：负责 **path_scope_mapping 的业务内容**（结构由主控 workflow 按模板+目录扫描生成，本智能体补充无法自动推断的业务字段）。

| 项目 | 说明 |
|------|------|
| **触发** | 主控 workflow 同步测试仓库后、或接入新模块/新 API 时，由流水线或人工触发。 |
| **输入** | 被测仓库目录结构、各模块 `apis.yaml`、`suites/*.yaml` 的 story、已有 path_scope_mapping（结构部分）。 |
| **技能** | 调用 **ensure_path_scope_mapping**（`scripts/ensure_path_scope_mapping.py`）：可传入 `--supplement` 产出并合并业务 YAML；或直接产出 supplement 供主控 workflow 在生成 mapping 时使用。 |
| **输出** | 补充/更新各模块 `path_scope_mapping.yaml` 中的 **suggested_suites**、**path_to_api**、**scope_tags**（及 name 等展示用字段）；或产出 **supplement YAML** 供主控合并。 |

**与主控 workflow 的关系**：主控先用模板+目录扫描生成/更新各模块 mapping（结构），再可选合并 mapping 智能体产出的 supplement；若未接入 mapping 智能体，则仅使用结构层 mapping，listener_output 中 scopes 仍可用，suggested_suites/path_to_api 等可为空由下游智能体 A/B 按需处理。

**脚本分工**：**ensure_path_scope_mapping.py** 产出/更新 path_scope_mapping 文件；**generate_listener_output.py**（主控使用）或 **commit_scope_mapping.py**（可选）在已有 mapping 与 commit 信息基础上产出 listener_output。

### 3.1 智能体 A：用例维护智能体（Case Maintenance Agent）

**职责**：根据 commit 与建议清单，对现有 case 进行增、删、改。

| 项目 | 说明 |
|------|------|
| **触发** | 提交监听输出 `need_add_cases=true` 或变更类型为 added/modified/deleted 且命中 path_scope_mapping |
| **输入** | 变更路径、变更类型、diffs、**建议：变更文件完整内容（commit 后）**、受影响 api_name/api_path、suggested_suites、本模块 base_dir；无需默认提供完整代码仓库，详见《提交监听服务_基于GitHub实现》第七节 |
| **规范依据** | `testcase/_config/spec/case_schema.yaml`、`suite_schema.yaml`；本模块 `apis.yaml`、`global_manifest.yaml`、`path_scope_mapping.yaml` |
| **动作** | 新增 case：在 suggested_suites 对应 `suites/*.yaml` 的 cases 中追加，url 用 apis.yaml 的 name，变量用 global_manifest 的 ref；修改 case：按 name 替换同名字段；删除/下线：从 cases 中移除或 suite switch=n |
| **输出** | 变更后的 `suites/*.yaml`（符合 Spec），并为新 case 打上细粒度 tags（如 api:xxx、smoke、regression）便于后续筛选 |

**写回方式**：推荐以 **PR** 方式写回用例仓库，避免直接 push 与并发冲突。约定如下：

| 方式 | 说明 |
|------|------|
| **PR（推荐）** | 推送到 `auto/{branch}-{commit_sha}` 分支并提 PR，base 为当前 main；PR 标题建议 `[auto] 用例维护 @{commit_sha}` |
| **直接 push** | 仅当用例仓库与研发仓库分离且无并发写回时可选；写回前**必须**校验分支 tip 是否仍为 commit_sha（见《提交监听服务_基于GitHub实现》3.6 节） |

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

**输出格式**：标准化 JSON/YAML，供智能体 C 消费。`dimensions` 为数组，**functional 必选**，其余按需；顺序建议按优先级（functional → contract → boundary → stress → performance）：

```json
{
  "dimensions": ["functional", "contract", "boundary"],
  "reason": "接口定义变更，建议执行契约测试；参数校验逻辑变更，建议执行边界测试"
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `dimensions` | array | 是 | 维度列表，取值：functional、contract、boundary、stress、performance；**functional 必含** |
| `reason` | string | 否 | 维度判断依据摘要，供人工或下游参考 |

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

**输出格式**：标准化 JSON，供流水线/执行器消费。支持两种形态：

**方案一（按 name 执行）**：单批次或多批次，每批次为 CASE_NAMES 列表：

```json
{
  "mode": "case_names",
  "batches": [
    {
      "priority": 1,
      "label": "新增功能",
      "case_names": "用例名1,用例名2,用例名3"
    },
    {
      "priority": 2,
      "label": "回归",
      "case_names": "用例名4,用例名5"
    }
  ]
}
```

**方案二（按 scope/tags/api 执行）**：多轮筛选参数，每轮一组条件：

```json
{
  "mode": "filter",
  "batches": [
    {
      "priority": 1,
      "label": "新增功能",
      "scope": null,
      "tags": null,
      "api_name": "测试数据源连接",
      "api_path": "/api/data-connection/v1/datasource/test"
    },
    {
      "priority": 2,
      "label": "回归",
      "scope": "vega-data-connection",
      "tags": ["regression"],
      "api_name": null,
      "api_path": null
    }
  ]
}
```

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `mode` | string | 是 | `case_names` 或 `filter` |
| `batches` | array | 是 | 执行批次，按 priority 升序执行 |
| `batches[].priority` | int | 是 | 执行顺序，1 最先 |
| `batches[].label` | string | 否 | 批次说明 |
| `batches[].case_names` | string | 方案一必填 | 逗号分隔的 case name，**须含 prev_case 依赖链** |
| `batches[].scope/tags/api_name/api_path` | - | 方案二按需 | 与 get_cases 参数一致，未用则 null |

**多批次执行**：流水线依次处理每个 batch；每批调用 `get_cases(...)` 或设置 `CASE_NAMES=...` 后执行 pytest，Allure/JUnit 结果写入同一 report 目录，最后合并供智能体 D 使用。报告路径约定见 4.3 节。

**与现有框架的衔接**：  
- 执行时 `CASE_NAMES=... pytest` 或 `TAGS=... SCOPE=... pytest`，见 README 与《测试框架架构_全自动与智能体规范》第九节。  
- 仅跑筛出的单条用例，不整 suite 全量执行。  
- **重要**：使用 CASE_NAMES 或 names 筛选时，必须包含所有 prev_case 依赖链上的用例，否则前置数据缺失会导致结果错误

---

### 3.4 智能体 D：报告生成智能体（Report Generator Agent，可选）

**职责**：在测试执行完成后，汇总结果并生成可读的测试报告。

| 项目 | 说明 |
|------|------|
| **输入** | Allure 结果（`report/xml`）、JUnit（`report/junit_report.xml`）；可选：执行维度、批次信息、case_names 列表（见 4.3 节） |
| **逻辑** | 解析通过率、失败用例、失败原因；按维度（功能/契约/边界/压力/性能）汇总；给出结论与建议（如：是否需人工介入、是否建议补测） |
| **输出** | 测试报告（HTML 或 Markdown）；建议含：**摘要**（通过率、总数、失败数）、**失败明细**（case name、原因）、**结论与建议** |

**输出格式**：以 Markdown 为主，便于流水线展示与归档；可选产出结构化 JSON 供下游消费：

```json
{
  "summary": { "total": 50, "passed": 48, "failed": 2, "pass_rate": "96%" },
  "failed_cases": [{ "name": "用例名", "reason": "断言失败..." }],
  "conclusion": "建议人工介入检查失败用例",
  "report_path": "report/agent_summary.md"
}
```

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

### 3.6 path_scope_mapping 与 path_to_api 结构规范

**path_scope_mapping.yaml** 位于各模块（如 `testcase/vega/_config/`）下，用于提交监听与按 scope 筛选用例。**模板统一在 `testcase/_config/spec/`**（如 `path_scope_mapping.template.yaml`），vega 下该文件为按模板生成的产物：若不存在则从模板创建或追加，若已存在则根据仓库路径/API 变化进行增删改（见 3.6.1 节）。完整结构：

```yaml
subsystems:
  - id: vega-data-connection      # scope 标识，与 get_cases(scope=...) 一致
    name: VEGA数据连接
    path_patterns:                 # 仓库相对路径 glob，如 vega/data-connection/**
      - "vega/data-connection/**"
    scope_tags: [regression, data-connection]
    smoke_tags: [smoke]             # 可选
    performance_tags: [performance] # 可选
    suggested_suites:              # 命中时供智能体 A 参考的套件名（对应 suites/*.yaml 的 story）
      - 测试数据源
      - 新增数据源

# 可选：path_to_api 可合并于此文件或单独 path_to_api_mapping.yaml
path_to_api:
  - path_pattern: "**/data-connection/**/DatasourceController.java"
    api_path: /api/data-connection/v1/datasource
    api_name: 新增数据源
  - path_pattern: "**/datasource/test*"
    api_path: /api/data-connection/v1/datasource/test
    api_name: 测试数据源连接

test_type_tags:
  functional: [regression, smoke]
  contract: [regression, contract]
  boundary: [boundary, regression]
  stress: [stress]
  performance: [performance]
```

| 字段 | 说明 |
|------|------|
| `subsystems[].id` | scope 标识，命名建议 `模块-子系统`（如 vega-data-connection） |
| `subsystems[].path_patterns` | 仓库根相对路径 glob，支持 `**`；空数组表示仅作 scope 用（如 performance） |
| `subsystems[].suggested_suites` | 套件 story 名列表，与 suites/*.yaml 的 story 字段对应 |
| `path_to_api[].path_pattern` | 变更路径匹配模式 |
| `path_to_api[].api_path` / `api_name` | 命中时产出 affected_api_paths / affected_api_names，api_name 与 apis.yaml 一致 |

**suggested_suites 与 suite 文件**：suggested_suites 中的名称对应 `suites/*.yaml` 的 `story` 字段；智能体 A 在对应 suite 的 cases 中增删改。

#### 3.6.1 结构由脚本+模板生成，业务由 mapping 智能体补充

**模板统一位置**：所有模板放在 **`testcase/_config/spec/`** 下（如 `path_scope_mapping.template.yaml`）。

**结构层（主控 workflow 自动执行）**：主控在产出 listener_output 前，调用测试仓库中的 **ensure_path_scope_mapping.py**（`scripts/ensure_path_scope_mapping.py`）：从模板读入，扫描被测仓库目录（不写死模块名），为 testcase 下**所有模块**分别写入 `_config/path_scope_mapping.yaml`；若已存在则仅增补缺失的 scope，不删已有。脚本支持 `--supplement`：若提供 mapping 智能体产出的 YAML，则合并 suggested_suites、path_to_api、scope_tags 等业务字段。

**业务层（mapping 智能体）**：**mapping 智能体**（见 3.0 节）负责根据 apis.yaml、suites 结构等产出业务补充；以技能形式调用同一脚本并传入 `--supplement`，或产出 supplement 文件供主控合并。由此，结构变化自动跟上，业务内容由 mapping 智能体统一维护，无需人工手写。

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
| case/suite tags | 在现有 smoke、regression、performance 基础上，约定 **boundary**、**stress**、**contract**，并在 case_schema 说明中体现；**tags 约定**见 4.4 节 |
| path_scope_mapping | 完整结构见 **3.6 节**；在 `test_type_tags` 中增加 boundary、stress（可选 contract），见 3.5 节 |
| 提交监听/DevOps 输出 | 返回格式见 **2.7 节**；主控 workflow 产出 listener_output.json；需稳定提供：变更路径、变更类型、affected_api_names、affected_api_paths、scope、suggested_suites、need_add_cases，以便智能体 A/B/C 消费 |
| 执行顺序 | 若分多批次（先功能再回归再维度），流水线需支持多轮调用 get_cases + pytest，并合并 Allure/JUnit 结果供智能体 D 使用；智能体 C 输出格式见 3.3 节 |
| **多模块部署顺序** | 当一次提交涉及多模块且存在依赖时，需维护**模块依赖配置**（如 `module_dependencies.yaml` 或 `deploy_order`），按拓扑序分层部署；详见 **2.5 节** |

### 4.3 报告路径约定

执行器产出的报告路径固定，供智能体 D 与流水线读取：

| 类型 | 路径 | 说明 |
|------|------|------|
| Allure XML | `report/xml` | pytest-allure 产出，用于生成 HTML |
| Allure HTML | `report/html` | 最终可读报告 |
| JUnit | `report/junit_report.xml` | 供智能体 D 解析通过率、失败 case |

智能体 D 的**输入**：上述路径下的文件；可选接收执行维度、批次信息、case_names 列表作为上下文。

### 4.4 Tags 约定

| 类别 | tags | 说明 |
|------|------|------|
| **功能** | smoke, regression | 常规功能与回归 |
| **契约** | contract, regression | 结构/字段校验（resp_schema） |
| **边界** | boundary, regression | 参数边界、长度、枚举临界值 |
| **压力** | stress | 并发、限流、熔断 |
| **性能** | performance | 时延、吞吐、资源 |
| **API 粒度** | api:{api_name} | 与 apis.yaml 的 name 一致，如 `api:测试数据源连接`，便于按 API 精确筛选 |

单条 case 的 tags 可组合使用；`api:xxx` 用于将 case 与具体接口绑定，智能体 C 按 affected_api_names 筛选时优先匹配。

### 4.5 异常与降级

| 场景 | 建议处理 |
|------|----------|
| **智能体超时/失败** | 跳过该智能体产出，下游按「无输出」降级：如 B 失败则默认 dimensions=[functional]；C 失败则按 scope+tags 兜底筛选 |
| **智能体输出格式不符** | 流水线校验 JSON schema，不符则记录日志并降级为默认策略，不阻塞执行 |
| **写回冲突** | 写回前校验分支 tip（见《提交监听服务_基于GitHub实现》3.6 节）；若 tip 已变则放弃写回，可选触发基于新 tip 的流水线 |
| **多模块部署某层失败** | 该层及依赖其的后续层标记失败，跳过测试执行；或仅对已就绪层对应的 scope 做有限回归（见 2.5 节） |

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
- **提交监听技术实现**：[提交监听服务_基于GitHub实现.md](提交监听服务_基于GitHub实现.md)（含**内网部署**方案：Git 镜像 + 内网 Webhook）
- **本架构**：多智能体分工、测试维度分析、执行优先级、报告生成；**主控 workflow**（2.2）与**下游编排**（2.3）串联全流程；**规范速查**：listener_output 格式（2.7）、智能体 B/C/D 输出格式（3.2/3.3/3.4）、path_scope_mapping（3.6）、tags 约定（4.4）、报告路径（4.3）、异常降级（4.5）。

---

## 七、小结

- **Commit 驱动**：提交监听产出变更与受影响 API，驱动用例维护与多维度分析。  
- **多智能体**：A 维护用例，B 分析测试维度，C 按优先级筛选用例并驱动执行，D 汇总报告。  
- **测试框架**：沿用现有 DIP AT 的 Spec、get_cases、单条用例粒度、CASE_NAMES/TAGS/SCOPE 执行与 Allure/JUnit 报告。  
- **多维度**：功能、契约、边界、压力、性能通过 tags 与 test_type_tags 约定，由 B 分析、C 筛选执行。  
- **执行原则**：先新增功能相关 case，再回归相关 case，最后按评估执行边界/压力/性能等维度；完成后生成报告。
