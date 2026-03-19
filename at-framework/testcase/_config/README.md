# 框架级配置（所有测试模块共用）

`testcase/_config/` 是**测试框架**的全局配置目录，主要提供“统一规范 + 各模块模板 + 生成/校验脚本的约定”。它不绑定任何业务模块；业务模块的运行时配置与用例位于 `testcase/<模块名>/` 下。

## 1. `spec/`：规范与模板

### 1.1 规范（长期稳定）
`spec/` 下的 schema 文件用于约束各模块 `suites/*.yaml` 的结构。

- `spec/case_schema.yaml`
  - 约束单条 case 的字段、类型和约定（如 `url` 必须是本模块 `apis.yaml` 中的 API `name`）。
  - 当前框架加载用例时不会“强制校验 schema”，但智能体/脚本生成或人工维护时应遵循它，避免运行时才爆炸。
- `spec/suite_schema.yaml`
  - 约束单个 suite（即 `suites/*.yaml` 文件）的顶层字段，如 `feature`、`story`、`switch`、`tags`、`cases`。
  - `switch: y/n` 决定该 suite 是否进入用例池（`switch: n` 会被直接跳过）。

### 1.2 模板（用于按模块生成产物）
模板文件用于帮助智能体或脚本生成各模块的 `_config/*.yaml`/`framework_hooks.py`。

- `spec/apis.template.yaml` → 模块 `_config/apis.yaml`
  - 定义 API 列表：每个 API 的 `name`、`url`（路径）、`method`、`headers`（JSON 字符串）。
  - 用例的 `url` 字段必须等于本模块 `apis.yaml` 里某个条目的 `name`（框架加载时用这个做合并）。
- `spec/global.template.yaml` → 模块 `_config/global.yaml`
  - 定义全局变量（真实取值）。
  - 用例通过 `${var}` 或 `$var` 引用这些变量；加载时框架会将 `global.yaml` 先展开成 `global_flat`。
- `spec/global_manifest.template.yaml` → 模块 `_config/global_manifest.yaml`
  - 定义“变量清单”（给智能体/提取脚本知道有哪些可用变量、如何引用）。
  - 当前框架执行时主要还是用 `global.yaml` 完成替换；`global_manifest.yaml` 更偏元信息。
- `spec/suite_manifest.template.yaml` → 模块 `_config/suite_manifest.yaml`
  - 声明模块包含哪些 suite（更多是给智能体/流水线做“可选项列表”的参考）。
  - 当前框架执行加载用例时是扫描 `suites/` 目录，不直接读取 `suite_manifest.yaml`。
- `spec/path_scope_mapping.template.yaml` → 模块 `_config/path_scope_mapping.yaml`
  - 定义“提交变更路径 → 测试 scope/tags/suggested_suites”的映射表。
  - 用于两件事：提交监听产出 scope，执行器用 `SCOPE` 把 scope 解析成 tags 再筛用例。
- `spec/api_deletion.template.yaml` → 模块 `_config/api_deletion.yaml`
  - 定义“已删除/下线 API 的维护动作清单”，用于指导智能体批量删除或禁用对应 case。
  - 典型场景：后端删除接口后，智能体按 `api_name` 删除 `suites/*.yaml` 中引用该 API 的 case，并清理 `prev_case` 残留引用。
- `spec/framework_hooks.template.py` → 模块可选 `framework_hooks.py`
  - 提供模块级的会话清理逻辑（例如清掉测试数据）。
  - 框架的 `conftest` 会在满足条件时动态加载并调用它；未提供则跳过。

## 2. 模块接入时必须有哪些文件

每个业务模块目录建议结构如下（与模板/规范强相关）：

- `testcase/<模块名>/_config/apis.yaml`（必需）
- `testcase/<模块名>/_config/global.yaml`（必需）
- `testcase/<模块名>/_config/global_manifest.yaml`（强烈建议）
- `testcase/<模块名>/_config/path_scope_mapping.yaml`（强烈建议；用于提交驱动筛选）
- `testcase/<模块名>/_config/suite_manifest.yaml`（建议；用于智能体/流水线参考）
- `testcase/<模块名>/suites/*.yaml`（必需）
  - 每个文件是一个 suite，字段结构参照 `spec/suite_schema.yaml`，内部 `cases` 结构参照 `spec/case_schema.yaml`。
- `testcase/<模块名>/framework_hooks.py`（可选）
  - 若需要会话级清理或其他模块钩子，就提供 `session_clean_up(config, allure)`。

## 3. 运行时关键约定（理解这些才能正确用）

### 3.1 用例如何被加载
框架运行时会扫描 `testcase/<模块名>/suites/` 目录，加载所有 `switch: y` 的 suite，再收集其 `cases`。

每条 case 在加载时会做这些合并/替换：

- `case.url` 作为“API name”查本模块 `_config/apis.yaml`
- 用 `global.yaml` 展开 `${变量}`，再执行 `prev_case` 的 `resp_values` 替换，最后对字符串做 Jinja2 渲染（替换顺序以 `case_schema.yaml` 的 `substitution_order` 为准）
- `case.tags` 若为空则继承 suite 的 `tags`

### 3.2 path_scope_mapping 如何影响执行
`path_scope_mapping.yaml` 不是直接生成 case，而是让你在执行时通过 `SCOPE=...` 间接筛选用例：

- `SCOPE=<subsystems.id>` 会先被解析成 tags（`scope_tags` + 可选的 `smoke_tags`）
- 然后用这些 tags 去过滤 case 的 tags

因此同一个 scope id 的含义，应该稳定地对应到“哪些类型/能力的用例集合”（例如 `vega-data-connection` 对应 `regression`、`data-connection` 等 tags）。

## 4. 常用脚本怎么用

框架提供了几类脚本，典型工作流是：先维护/生成模块 `_config`，再在 CI/流水线根据提交变更触发筛选。

- 提交监听：把 commit changed files 映射成 scope/tags/suggested_suites

  ```bash
  python scripts/commit_scope_mapping.py ^
    --changed-list changed_list.txt ^
    --mapping testcase/<模块名>/_config/path_scope_mapping.yaml ^
    --output listener_output.json
  ```

- 执行/提取：按 scope/tags/suite/name/api 过滤用例

  ```bash
  # 提取用例（给智能体挑选 name）
  python scripts/extract_cases.py --base-dir testcase/<模块名> --list-fields name,description,story

  # 直接执行筛选（比如跑 scope 对应的回归用例）
  SCOPE=<subsystem.id> pytest test_run.py
  ```

- 模块校验：检查模块目录是否具备最基本的结构，并可选开启严格接口引用

  ```bash
  python scripts/validate_module.py testcase/<模块名> --strict-apis
  ```

严格模式含义是：如果 case 引用的 API name 不在本模块 `apis.yaml` 中，会直接失败而非静默跳过，适合 CI。

## 5. 与 adp 多模块的关系

`testcase/_config/` 不关心 adp-main 里具体发生了什么变化；它提供的是“生成与执行的统一约定”。

真正随业务/提交变化的是每个模块自己的：

- `testcase/<模块名>/_config/path_scope_mapping.yaml`
- `testcase/<模块名>/_config/apis.yaml`、`global.yaml`、以及模块 `suites/*.yaml` 中的用例

当你新增或迁移一个模块时，优先从 `spec/*.template.*` 生成其 `_config`，并在上线前用 `validate_module.py --strict-apis` 做验证。
