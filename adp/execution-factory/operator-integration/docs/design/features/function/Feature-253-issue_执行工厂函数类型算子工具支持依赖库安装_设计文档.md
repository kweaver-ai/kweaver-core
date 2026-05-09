# Feature-253-issue 执行工厂函数类型算子/工具支持依赖库安装（实现及逻辑）设计文档

## 1. 文档信息

### 1.1 面向对象
- 后端开发人员：理解数据模型、接口契约、执行链路与可观测点；按本文实现/排障
- 前端对接人员：明确页面字段、请求/响应字段、调用时序与错误处理
- 技术人员：理解功能边界、依赖安装策略与风险控制

### 1.2 背景与问题
执行工厂的函数类型算子/工具依赖沙箱（Sandbox）执行 Python 代码。当前用户编写的函数经常依赖第三方库（如 `pandas`、`requests`），需要在执行前完成依赖安装并可配置安装源（例如企业内网 PyPI 镜像）。

目标是实现：
- 在函数元数据中可配置依赖列表与安装源
- 执行函数时可自动触发依赖安装
- 前端可通过接口辅助选择依赖版本

## 2. 范围与目标

### 2.1 目标（Goals）
- 支持函数类型算子/工具的依赖字段：`dependencies`、`dependencies_url`
- 支持两类执行入口的依赖安装：
  - 公共执行：直接提交代码执行
  - 内部执行：按元数据版本执行（由业务调用）
- 支持查询 PyPI 包版本，便于前端选择版本

### 2.2 非目标（Non-goals）
- 不在本需求中实现依赖安装的“多语言”（当前仅 Python）
- 不在本需求中实现复杂依赖缓存/冲突解决（可作为后续优化项）

### 2.3 术语
- 依赖列表（dependencies）：pip 规范的 requirement 字符串数组，例如 `["pandas==1.5.2", "requests>=2.28.0"]`
- 依赖源（dependencies_url）：pip 安装源 base URL，例如 `https://pypi.org` 或企业镜像 `https://pypi.xxx.com`
- 函数元数据：函数类型算子/工具的元信息与代码内容，持久化在 `t_metadata_function`

## 3. 总体方案

### 3.1 数据流（从配置到执行）
1. 前端在创建/编辑函数类型算子或工具时提交 `dependencies` 与 `dependencies_url`
2. 后端解析并持久化到函数元数据表 `t_metadata_function`
3. 执行函数时：
   1. 获取待执行代码（公共执行：直接来自请求；内部执行：来自元数据）
   2. 组装执行请求（包含 `dependencies` 与 `dependencies_url`）
   3. 进入会话池，获取/创建沙箱 session
   4. 若存在依赖配置，则调用沙箱依赖安装接口
   5. 执行代码并返回结果

### 3.2 会话与执行入口
- 会话池实现：在执行前按需触发依赖安装逻辑，见 [session_pool.go](/adp/execution-factory/operator-integration/server/logics/sandbox/session_pool.go#L144-L177)
- 公共执行入口：`POST /api/agent-operator-integration/v1/function/execute`，见 [rest_public_handler.go](/adp/execution-factory/operator-integration/server/driveradapters/rest_public_handler.go#L55-L61)
- 内部执行入口：`POST /api/agent-operator-integration/internal-v1/function/exec/:version`，见 [rest_private_handler.go](/adp/execution-factory/operator-integration/server/driveradapters/rest_private_handler.go#L46-L50)

## 4. 数据模型设计

### 4.1 对外结构体（JSON 字段契约）
- 执行请求结构：
  - [ExecuteCodeReq](/adp/execution-factory/operator-integration/server/interfaces/drivenadapters.go#L535-L543)
    - `dependencies: []string`（可选）
    - `dependencies_url: string`（默认 `https://pypi.org`）
- 元数据结构：
  - [FunctionContent / FunctionInput / FunctionInputEdit](/adp/execution-factory/operator-integration/server/interfaces/logics_metadata.go#L128-L195)
    - `dependencies`
    - `dependencies_url`

### 4.2 持久化（DB 字段）
函数元数据表 `t_metadata_function`：
- `f_dependencies`：text，存储依赖列表 JSON 字符串
- `f_dependencies_url`：text，存储依赖安装源

对应模型与 getter/setter：
- [func_metadata.go](/adp/execution-factory/operator-integration/server/interfaces/model/func_metadata.go#L11-L30)
- [SetDependencies/GetDependencies](/adp/execution-factory/operator-integration/server/interfaces/model/func_metadata.go#L148-L162)
- 迁移脚本（字段初始化）：[init.sql](/adp/execution-factory/operator-integration/migrations/mariadb/0.4.0/pre/init.sql#L284-L303)

### 4.3 元数据解析与转换链路
- 解析（请求 → DB 模型）：[py_func_parser.go](/adp/execution-factory/operator-integration/server/logics/parsers/py_func_parser.go#L99-L131)
- 落库（Insert/Update）：[dbaccess/func_metadata.go](/adp/execution-factory/operator-integration/server/dbaccess/func_metadata.go#L48-L125)
- 出库（DB → 对外结构）：[struct_convert.go](/adp/execution-factory/operator-integration/server/logics/metadata/struct_convert.go#L9-L38)
- 工具箱导入导出/编辑同步依赖字段：
  - [impex.go](/adp/execution-factory/operator-integration/server/logics/toolbox/impex.go#L632-L642)
  - [tool_edit.go](/adp/execution-factory/operator-integration/server/logics/toolbox/tool_edit.go#L221-L234)

## 5. 接口设计

### 5.1 公共接口：直接执行代码（支持依赖安装）
**POST** `/api/agent-operator-integration/v1/function/execute`

- Request Body：见 [ExecuteCodeReq](/adp/execution-factory/operator-integration/server/interfaces/drivenadapters.go#L535-L543)
  - `code`：必填
  - `event`：必填
  - `language`：默认 `python`
  - `timeout`：秒
  - `dependencies`：可选
  - `dependencies_url`：可选，缺省使用默认值
- Response Body：`stdout/stderr/result/metrics`，见 [FunctionExecute](/adp/execution-factory/operator-integration/server/driveradapters/common/proxy.go#L53-L97)

### 5.2 内部接口：按元数据版本执行函数（依赖来自元数据）
**POST** `/api/agent-operator-integration/internal-v1/function/exec/:version`

- Path：
  - `version`：函数元数据版本（UUID）
- Request Body：
  - `event`：函数入参（JSON）
- Query：
  - `timeout`：毫秒（内部转为秒）
- 执行数据来源：
  - `code/script_type/dependencies/dependencies_url` 均来自元数据读取

对应 handler：[FunctionExecuteProxy](/adp/execution-factory/operator-integration/server/driveradapters/common/proxy.go#L104-L170)

### 5.3 公共接口：查询依赖库版本（前端辅助）
**GET** `/api/agent-operator-integration/v1/function/dependency-versions/:package_name`

- 用途：前端在依赖配置面板中输入包名后，获取可选版本列表
- 处理逻辑：调用 PyPI 数据解析器，见 [QueryPypiVersions](/adp/execution-factory/operator-integration/server/driveradapters/common/proxy.go#L172-L200)

### 5.4 公共接口：获取当前环境依赖库列表（可选）
**GET** `/api/agent-operator-integration/v1/function/dependencies?language=python`

- 用途：前端展示“已预置依赖”或用于提示常用依赖（当前实现为 mock 返回）
- 处理逻辑：见 [GetDependencies](/adp/execution-factory/operator-integration/server/driveradapters/common/proxy.go#L202-L228)

## 6. 执行与依赖安装逻辑

### 6.1 触发条件
会话池在执行前按需触发依赖安装：
- 当 `len(req.Dependencies) > 0` 且 `req.DependenciesURL != ""` 时触发
- 逻辑位置： [ExecuteCode](/adp/execution-factory/operator-integration/server/logics/sandbox/session_pool.go#L144-L177)

### 6.2 安装调用
会话池调用沙箱控制平面接口：
- [InstallDependenciesReq](/adp/execution-factory/operator-integration/server/interfaces/drivenadapters.go#L646-L654)

沙箱控制平面 client（当前实现）：
- 目前 `InstallDependencies` 为 mock（只打日志并返回成功），见 [sandbox_control_plane.go](/adp/execution-factory/operator-integration/server/drivenadapters/sandbox_control_plane.go#L212-L218)
- 后续对接建议：在控制平面补充真实 REST API（例如 `POST /api/v1/executions/sessions/{id}/install-dependencies` 或等价能力），并在该方法内调用

### 6.3 失败策略
当前实现：依赖安装失败会记录错误日志，但仍继续执行代码（避免由于依赖源波动导致整体不可用）。
- 位置： [ExecuteCode](/adp/execution-factory/operator-integration/server/logics/sandbox/session_pool.go#L160-L170)

建议策略（可配置项，后续可扩展）：
- 严格模式：依赖安装失败直接返回错误，不执行代码
- 非严格模式：依赖安装失败仅告警并继续执行（当前策略）

## 7. 前端对接说明

### 7.1 配置入口
函数类型算子/工具的编辑界面增加依赖配置区：
- 依赖列表：多行输入或可增删的列表项
- 依赖源：下拉（默认 PyPI）+ 自定义输入（企业镜像）

### 7.2 字段校验建议
- `dependencies`：
  - 每项不为空
  - 推荐限制长度（例如 1~200 字符）
  - 支持 `==`、`>=`、`<=` 等常见 pip 约束表达
- `dependencies_url`：
  - URL 格式校验
  - 空值时后端采用默认 `https://pypi.org`

### 7.3 依赖版本选择
前端可调用：
- `GET /function/dependency-versions/:package_name`
用于为 `package_name` 提供版本候选（用户选中版本后写入 `dependencies` 项）。

### 7.4 执行对接
- 直接执行代码（调试/测试执行）：使用 `POST /function/execute`，由前端直接传 `dependencies/dependencies_url`
- 业务执行（生产调用）：调用内部执行接口按 `version` 执行，由后端从元数据读取依赖信息，无需前端再次传入

## 8. 可观测性与排障
- 日志：
  - 依赖安装行为应包含：`session_id`、`dependencies`、`dependencies_url`、执行语言
  - 当前 mock 日志：见 [InstallDependencies](/adp/execution-factory/operator-integration/server/drivenadapters/sandbox_control_plane.go#L212-L218)
- 链路追踪：
  - 执行链路 span：见 [ExecuteCode](/adp/execution-factory/operator-integration/server/logics/sandbox/session_pool.go#L144-L177)

## 9. 兼容性
- 不传 `dependencies`：行为与历史一致，不触发安装
- 传 `dependencies` 但不传 `dependencies_url`：
  - 公共执行入口：默认填充为 `https://pypi.org`
  - 内部执行入口：默认填充为 `https://pypi.org`（由执行实现补默认值）

## 10. 测试用例（建议）
- 元数据落库/出库：
  - 创建/编辑函数类型算子时携带 `dependencies/dependencies_url`，查询元数据/工具箱返回一致
- 执行：
  - 公共执行：传入依赖后应触发 InstallDependencies（可通过日志/Mock 断言）
  - 内部执行：元数据存在依赖时应触发 InstallDependencies
- 异常：
  - 依赖安装失败：应记录错误并继续执行（当前策略）
