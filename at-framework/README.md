# 自动化测试框架（DIP AT）

基于 YAML 用例与 Pytest 的 API 自动化测试框架，支持**多模块**、**按单条用例筛选执行**，可与提交监听、智能体配合实现全自动回归与按需测试。当前仓库内 VEGA 为其中一个测试模块。

---

## 一、架构概览

### 1.1 设计原则

- **多模块通用**：框架与具体业务解耦，VEGA 仅为模块之一；通过 `config.ini` 的 `case_file` 指定当前运行模块。
- **规范即契约**：用例/套件结构、全局变量、API 定义均有唯一规范（Spec），便于智能体按规范增删改与提取。
- **单条用例粒度**：支持按 scope、tags、api_name、api_path 筛选用例，只跑受影响的用例，而非整 suite。

### 1.2 目录分层

| 层级 | 路径 | 说明 |
|------|------|------|
| **框架级** | `testcase/_config/spec/` | 用例与套件规范（case_schema、suite_schema），所有模块共用 |
| **模块级** | `testcase/<模块名>/_config/` | 本模块的 apis、全局变量、path_scope_mapping、suite_manifest |
| **模块级** | `testcase/<模块名>/suites/` | 本模块的套件 YAML（*.yaml） |

### 1.3 整体流程

```
[研发提交] → 提交监听（变更路径/受影响 API）
    → path_scope_mapping 得到 scope/tags/suggested_suites
    → 按需：智能体补用例 → get_cases(scope/tags/api_name/api_path) 筛出单条用例
    → pytest 只跑筛出的用例 → Allure/JUnit 报告
```

---

## 二、目录结构（详细）

```
项目根目录/
├── config/
│   └── config.ini              # 运行配置：host、token、当前模块 case_file、清理开关
├── testcase/
│   ├── _config/                # 框架级（所有模块共用）
│   │   ├── spec/
│   │   │   ├── case_schema.yaml    # 单条 case 字段规范
│   │   │   └── suite_schema.yaml   # 套件规范
│   │   └── README.md
│   └── vega/                   # 模块示例：VEGA
│       ├── _config/
│       │   ├── apis.yaml           # 接口定义（name → url/method/headers）
│       │   ├── global.yaml         # 全局变量 name: value
│       │   ├── global_manifest.yaml  # 变量清单（供智能体/提取用）
│       │   ├── path_scope_mapping.yaml  # 提交路径 → scope/tags/建议套件
│       │   ├── suite_manifest.yaml     # 套件列表说明
│       │   └── spec/README.md          # 指向框架级 spec
│       └── suites/
│           └── *.yaml               # 各套件，符合 suite_schema，case 符合 case_schema
├── common/                    # 框架代码
│   ├── func.py                # 加载用例、get_cases、全局变量、参数替换
│   └── constant.py            # 报告路径等常量
├── request/                   # HTTP 客户端
├── conftest.py                # pytest  fixture、用例列表来源、按模块清理
├── test_run.py                # 单条用例执行逻辑（请求、断言）
├── main.py                    # 入口：pytest + Allure 报告
├── scripts/
│   └── extract_cases.py       # 按条件提取用例或全局变量清单（CLI）
├── report/                    # 报告输出（xml、html、junit）
├── requirement.txt
├── pytest.ini
└── docs/
    └── 测试框架架构_全自动与智能体规范.md   # 完整架构与智能体规范
```

---

## 三、快速开始

### 3.1 环境要求

- Python 3.x
- 依赖见 `requirement.txt`（含 pytest、allure-pytest、requests、PyYAML、jsonpath、jsonschema、genson、Jinja2 等）

### 3.2 安装依赖

```bash
pip install -r requirement.txt
```

### 3.3 配置

编辑 `config/config.ini`：

| 配置项 | 说明 |
|--------|------|
| `[env].host` | 被测服务主机（如 10.4.111.209） |
| `[env].case_file` | 当前运行的**模块目录**，如 `./testcase/vega` |
| `[env].clean_up` | `1`=执行清理，其他=不清理 |
| `[env].clean_up_module` | 仅当与 case_file 对应模块名一致时执行清理（如 `vega`），避免误清其他模块 |
| `[external].token` | Bearer Token，请求头中自动注入 |

### 3.4 运行测试

```bash
# 运行当前模块（config.ini 中 case_file）全部用例
pytest

# 或使用入口脚本（pytest + 生成 Allure 报告）
python main.py
```

---

## 四、按条件运行与提取

### 4.1 环境变量筛选用例（执行）

在**不修改用例文件**的前提下，通过环境变量只跑部分用例：

| 变量 | 说明 | 示例 |
|------|------|------|
| `SCOPE` | 与 path_scope_mapping 的 subsystem.id 一致，解析为 tags 后过滤 | `SCOPE=vega-data-connection` |
| `TAGS` | 逗号分隔的 tag，case 的 tags 与其中任一相交则保留 | `TAGS=smoke` 或 `TAGS=performance` |
| `API_NAME` | 只跑调用该接口的用例（与 apis.yaml 的 name 一致） | `API_NAME=测试数据源连接` |
| `API_PATH` | 只跑 url 包含该路径的用例 | `API_PATH=/api/data-connection/v1/datasource/test` |

示例：

```bash
# 只跑冒烟
TAGS=smoke pytest

# 只跑某接口相关用例
API_NAME=测试数据源连接 pytest

# 只跑性能用例
TAGS=performance pytest
# 或
SCOPE=performance pytest
```

### 4.2 提取用例或全局变量（CLI）

不执行，仅按条件输出用例列表或变量清单，供流水线/智能体使用：

```bash
# 指定模块目录（默认取环境变量 CASE_FILE 或 testcase/vega）
python scripts/extract_cases.py --base-dir testcase/vega --tags smoke --format json
python scripts/extract_cases.py --base-dir testcase/vega --api-name 测试数据源连接 --format yaml
python scripts/extract_cases.py --base-dir testcase/vega --scope vega-data-connection

# 输出本模块全局变量清单
CASE_FILE=testcase/vega python scripts/extract_cases.py --globals
```

参数与 `get_cases()` 一致：`--scope`、`--tags`、`--suite`、`--name`、`--api-name`、`--api-path`。

---

## 五、关键配置与字段说明

### 5.1 config.ini（运行配置）

已在「快速开始」中说明；`case_file` 决定当前运行哪个模块，`clean_up_module` 决定仅对哪个模块执行 session 级清理（若 conftest 中实现了清理逻辑）。

### 5.2 单条用例字段（case_schema 要点）

用例写在各模块 `suites/*.yaml` 的 `cases` 数组中，每条 case 建议符合 `testcase/_config/spec/case_schema.yaml`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| name | string | 是 | 用例唯一名称，同 suite 内不重复，建议「场景_预期」 |
| url | string | 是 | **接口名称**，必须与本模块 `apis.yaml` 中某一项的 `name` 一致（加载时合并该 API 的 url、method、headers） |
| prev_case | string | 否 | 前置用例的 name，先执行前置再执行本用例；空表示无 |
| path_params | string(JSON) | 否 | 路径参数，如 `{"id":"${catalog_id}"}` |
| query_params | string(JSON) | 否 | 查询参数 |
| header_params | string(JSON) | 否 | 用例级请求头覆盖，与 apis 的 headers 合并；Authorization 由框架注入 |
| cookie_params | string(JSON) | 否 | Cookie 键值，多数 REST 可省略 |
| body_params | string(JSON) | 否 | 请求体，可引用 `${var}`、`${bin_data_template_xxx}` |
| form_params | string(JSON) | 否 | 表单参数 |
| resp_values | string(JSON) | 否 | 从响应提取供后续用例用，如 `{"catalog_id":"$.id"}`（jsonpath） |
| code_check | string | 否 | 期望 HTTP 状态码 |
| resp_headers_check | string(JSON) | 否 | 响应头断言，如 `{"Content-Type":"application/json"}` |
| resp_check | string(JSON) | 否 | jsonpath 与期望值，如 `{"$.code":0}` |
| resp_schema | string(JSON) | 否 | 响应样本，框架用 genson 生成 schema 做结构校验 |
| tags | list[string] | 否 | 本条用例标签；未写则继承套件 tags。用于按提交筛选；约定：smoke、regression、**performance**（性能） |
| description | string | 否 | 说明，不参与执行 |

**变量引用**：用例中写 `$var` 或 `${var}`；替换顺序：global.yaml 展开 → 前序用例 resp_values → path_params → Jinja2 渲染。

### 5.3 套件字段（suite_schema 要点）

每个 `suites/*.yaml` 文件建议符合 `testcase/_config/spec/suite_schema.yaml`。

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| feature | string | 是 | 功能大类 |
| story | string | 是 | 场景/故事，建议与文件名（不含扩展名）一致 |
| switch | string | 是 | `y`=参与加载进入用例池，`n`=不加载（整套件不参与执行） |
| tags | list[string] | 否 | 套件默认标签；单条 case 未写 tags 时继承 |
| description | string | 否 | 套件说明 |
| cases | list | 是 | 用例数组，每项符合 case_schema |

`switch` 与按 scope/tags 筛选的关系：**switch 控制是否入池**，**筛选控制入池后执行哪些**。整套件临时下线时改为 `switch: n` 即可。

### 5.4 apis.yaml（模块接口定义）

本模块 `_config/apis.yaml`：定义接口名称与请求信息，用例中的 `url` 填的是**接口 name**，加载时会被替换为实际 url/method/headers。

- 列表格式：每项 `name`、`url`、`method`、`headers`（JSON 字符串）。
- 字典格式：key 为 name，value 可为字符串（视为 url）或 `{name, url, method, headers}`。
- url 中可含占位符如 `${id}`，由用例的 path_params 或 resp_values 填充。

### 5.5 global.yaml 与 global_manifest.yaml

- **global.yaml**：本模块全局变量唯一来源。格式 `name: value` 或 `[{name, value}, ...]`；支持变量间引用，先做一轮 `string.Template` 得到 global_flat，再参与用例替换。
- **global_manifest.yaml**：列出本模块可在 case 中引用的变量，供智能体/提取使用。每项建议含 `name`、`description`、`ref`（如 `"${mysql_host}"`），可选 `used_in`。

### 5.6 path_scope_mapping.yaml（提交路径 → 测试范围）

本模块 `_config/path_scope_mapping.yaml`，用于提交监听与按 scope 筛选用例。

| 字段（subsystems 每项） | 说明 |
|------------------------|------|
| id | scope 标识，如 `vega-data-connection`、`performance` |
| name | 子系统/测试类型名称 |
| path_patterns | 研发仓库路径模式列表，匹配则命中该 scope（可为空，如 performance 仅作 scope 用） |
| scope_tags | 该 scope 解析得到的 tags，get_cases(scope=id) 时用其过滤 case |
| smoke_tags | 可选，冒烟 tag |
| performance_tags | 可选，该子系统有性能用例时的 tag |
| suggested_suites | 建议的套件名列表，补用例时参考 |

文件末尾可有 `test_type_tags`：`functional`、`contract`、`performance` 与 tags 的对应关系，供智能体识别测试类型。

---

## 六、测试类型（功能 / 契约 / 性能）

- **功能**：常规 API 行为与回归，tags 如 `smoke`、`regression`。
- **契约**：结构/字段校验（如 `resp_schema`），多与 regression 共用 tags。
- **性能**：用例或套件打 tag **`performance`**；执行时用 `TAGS=performance` 或 `SCOPE=performance` 筛选。

性能相关改动时，提交监听可建议执行上述筛选条件，只跑性能用例。

---

## 七、新增一个测试模块

1. 在 `testcase/` 下新建目录，如 `testcase/新模块/`。
2. 创建 `_config/`：从 vega 复制并修改 `apis.yaml`、`global.yaml`、`global_manifest.yaml`、`path_scope_mapping.yaml`、`suite_manifest.yaml`；`spec/` 下仅保留 README 指向 `testcase/_config/spec/`。
3. 创建 `suites/`，按 suite_schema、case_schema 编写 `*.yaml`。
4. 在 `config.ini` 中设置 `case_file = ./testcase/新模块`，按需设置 `clean_up_module`。
5. 运行：`pytest` 或 `python main.py`，提取：`python scripts/extract_cases.py --base-dir testcase/新模块 ...`。

---

## 八、报告与常量

- **Allure**：`main.py` 会生成 xml 到 `report/xml`，再生成 HTML 到 `report/html`。
- **JUnit**：`report/junit_report.xml`（路径在 `common/constant.py` 中配置）。

---

## 九、智能体对接：按提交内容灵活筛选与执行

### 9.1 按用例 name 灵活筛选（推荐）

**思路**：智能体根据**提交内容**与用例 YAML 中的 **name**（及 description、story）语义匹配，决定哪些 case 需要回归，**不依赖固定 tag**。流程：

1. **拉取全量用例的 name/description**，供智能体根据 commit 描述做匹配：
   ```bash
   python scripts/extract_cases.py --base-dir testcase/vega --list-fields name,description,story --format json
   ```
2. **智能体根据提交内容**（如「新增数据源，将数据源长度显示放大到 256」）从上述列表中选出与「新增数据源」「长度」「名称」等相关的 **case name**。
3. **按选出的 name 列表提取或执行**（见 9.2）。

**按 name 列表提取**（不执行，仅输出用例列表）：

```bash
python scripts/extract_cases.py --base-dir testcase/vega --names "用例名1,用例名2,用例名3" --format json
```

可选：仍支持按 **scope/tags/api_name/api_path** 等固定条件筛选；与「按 name 灵活筛选」二选一或组合（先按 scope 再在结果里按 name 需在流水线里分步做）。

### 9.2 执行智能体筛选出的用例（执行命令）

conftest 优先读取 **CASE_NAMES**（逗号分隔的用例 name 列表），若有则**只执行这些 name 对应的用例**；否则再按 SCOPE/TAGS/API_NAME/API_PATH 筛选。  
**注意**：按 CASE_NAMES 执行时，必须把所选用例的 **prev_case 依赖链**一并纳入，否则前置用例未执行会导致参数缺失或断言失败。详见 [docs/测试框架合理性评估.md](docs/测试框架合理性评估.md)。

**按 name 列表执行**（智能体根据内容选出 name 后）：

```bash
# 只跑智能体选出的若干条用例（name 与 YAML 中完全一致）
set CASE_NAMES=新增数据源_name为空串_参数校验_请求失败,新增数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功
pytest
```

**一行命令示例**（Windows / Linux）：

```bash
# Windows CMD（name 中含逗号时用引号包住整串，或换行/文件方式见下）
set CASE_NAMES=用例名1,用例名2,用例名3
pytest

# Linux / macOS
CASE_NAMES="用例名1,用例名2,用例名3" pytest
```

**仍支持的按 scope/tags/api 执行**（与 CASE_NAMES 二选一）：

```bash
set SCOPE=vega-data-connection
pytest

set TAGS=smoke
pytest

set API_NAME=测试数据源连接
pytest
```

**带 Allure 报告**：上述任一筛选方式下均可 `python main.py`，仅执行当前筛选出的用例。

### 9.3 场景示例：提交「新增数据源，将数据源长度显示放大到 256」

**应回归的 case**（智能体根据提交内容从用例 name 中匹配「新增数据源」「长度」「name」等，选出以下与“数据源名称长度/显示”强相关的用例）：

| 用例 name（与 YAML 完全一致） |
|------------------------------|
| 新增数据源_参数校验_name长度为1_请求成功 |
| 新增数据源_参数校验_name长度为128_请求成功 |
| 新增数据源_参数校验_name长度为129_请求失败 |
| 新增数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功 |
| 更新数据源_参数校验_name长度为1_请求成功 |
| 更新数据源_参数校验_name长度为128_请求成功 |
| 更新数据源_参数校验_name长度为129_请求失败 |
| 更新数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功 |

（若「显示」还涉及列表/详情展示，可再选 查询数据源 套件中与列表、详情的 case。）

**步骤 1 — 拉取全量 name 供智能体选**（可选）：

```bash
python scripts/extract_cases.py --base-dir testcase/vega --list-fields name,description,story --format json
```

**步骤 2 — 按选出的 name 执行**（在项目根目录，且 `config.ini` 中 `case_file = ./testcase/vega`）：

```bash
# Windows CMD（一行内逗号分隔多个 name，不要有多余空格）
set CASE_NAMES=新增数据源_参数校验_name长度为1_请求成功,新增数据源_参数校验_name长度为128_请求成功,新增数据源_参数校验_name长度为129_请求失败,新增数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功,更新数据源_参数校验_name长度为1_请求成功,更新数据源_参数校验_name长度为128_请求成功,更新数据源_参数校验_name长度为129_请求失败,更新数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功
pytest
```

或带 Allure 报告：

```bash
set CASE_NAMES=新增数据源_参数校验_name长度为1_请求成功,新增数据源_参数校验_name长度为128_请求成功,新增数据源_参数校验_name长度为129_请求失败,新增数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功,更新数据源_参数校验_name长度为1_请求成功,更新数据源_参数校验_name长度为128_请求成功,更新数据源_参数校验_name长度为129_请求失败,更新数据源_参数校验_name包含中英文、数字、中划线、下划线_请求成功
python main.py
```

**先只提取不执行**（校验会跑哪些 case）：

```bash
python scripts/extract_cases.py --base-dir testcase/vega --names "新增数据源_参数校验_name长度为1_请求成功,新增数据源_参数校验_name长度为128_请求成功,新增数据源_参数校验_name长度为129_请求失败" --format json
```

---

## 十、更多说明

完整架构、全局变量解析顺序、提交监听与智能体规范、get_cases 与执行关系等，见 **[docs/测试框架架构_全自动与智能体规范.md](docs/测试框架架构_全自动与智能体规范.md)**。
