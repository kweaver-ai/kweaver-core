# AGENTS.md - Agent Executor 项目协作手册

## 项目定位

Agent Executor 是一个 Python 3.10 + FastAPI 服务，负责基于 Agent Factory 提供的配置构建 Agent Context，并通过 Dolphin SDK 执行 Agent。当前仓库重点在：

- `app/router/`：HTTP 路由、异常处理、中间件
- `app/logic/agent_core_logic_v2/`：Agent 执行主流程
- `app/config/config_v2/`：YAML 配置加载与配置模型
- `app/boot/`：服务启动前初始化
- `test/`：pytest 测试，按模块分层组织

## 常用入口

- `main.py`：进程入口，先执行 `boot.on_boot_run()`，再启动 Uvicorn。
- `app/router/__init__.py`：FastAPI 应用初始化、中间件、健康检查、路由注册。
- `app/router/agent_controller_pkg/`：Agent v1/v2 接口实现。
- `app/router/tool_controller.py`：工具相关接口。
- `app/logic/agent_core_logic_v2/agent_core_v2.py`：核心执行编排入口。
- `app/config/config_v2/config_loader.py`：配置路径优先级与 YAML 加载逻辑。
- `test/conftest.py`：测试启动时预先 mock Dolphin SDK，并重置全局状态。

## 配置与运行

- 配置文件名固定为 `agent-executor.yaml`。
- 配置目录优先级：`AGENT_EXECUTOR_CONFIG_PATH` > `/sysvol/conf/` > `./conf/`。
- 本地开发优先使用 `./conf/agent-executor.yaml`；缺失时可参考：

```bash
cp .local/conf/agent-executor.yaml ./conf/agent-executor.yaml
cp .env.example .env
```

- `.env.example` 默认将 `AGENT_EXECUTOR_CONFIG_PATH` 指向 `./conf/`。
- `conf/agent-executor.yaml.example` 中默认服务端口为 `30778`，API 前缀为：
  - ` /api/agent-executor/v1`
  - ` /api/agent-executor/v2`
- 健康检查在 `app/router/__init__.py` 中注册：
  - `/api/agent-executor/v1/health/alive`
  - `/api/agent-executor/v1/health/ready`

## 常用命令

### 日常开发

```bash
make dev-setup
make uv-sync
make run
```

### 测试

```bash
make test
make test-unit
make test-integration
make test-integration-filter FILTER=<pattern>
make test-verbose
uv run pytest test/unit/test_xxx.py::test_xxx -v
```

### 代码质量

```bash
make lint
make format
make clean
make uv-clean
```

补充说明：

- `make lint` 实际执行 `uv run pre-commit run -a --config .pre-commit-config.yaml`。
- `make format` 实际执行 `uv run ruff format app test`。

### 本地打包运行

如需验证 PyInstaller 产物，使用 `Makefile.local.mk`：

```bash
make -f Makefile.local.mk build
make -f Makefile.local.mk run-dist
make -f Makefile.local.mk status-dist
make -f Makefile.local.mk logs-dist
make -f Makefile.local.mk stop-dist
```

注意：

- `run-dist` 会使用 `./dist/agent-executor/agent-executor`。
- `run-dist` 会自动设置 `IS_LOCAL_DEV=true`、`LOCAL_DEVRUN_SCENARIO=aaron_local_dev`、`AGENT_EXECUTOR_CONFIG_PATH=$(pwd)/.local/conf/`。
- `run-dist` 的日志和 PID 写到仓库外的固定目录，改动前先确认本机路径是否可用。

## 关键代码入口

### 请求链路

1. `app/router/agent_controller_pkg/` 接收请求
2. `app/logic/agent_core_logic_v2/agent_core_v2.py` 编排执行
3. `app/logic/agent_core_logic_v2/prompt_builder.py` 构造提示词
4. `app/logic/agent_core_logic_v2/run_dolphin.py` 调用 Dolphin SDK
5. 响应通过流式接口返回

### 配置链路

1. `app/common/config.py` 初始化全局 `Config`
2. `app/config/config_v2/config_loader.py` 解析配置目录并加载 YAML
3. `app/config/config_v2/models/` 定义各配置模型

### 启动链路

1. `main.py`
2. `app/boot/boot.py`
3. `app/boot/built_in.py`
4. `app/router/__init__.py`

## 验证方式

- 改动业务代码后，优先运行最小必要测试，再扩大到 `make test`。
- 改动路由、中间件、异常处理时，至少补跑相关 `test/unit/router/` 或对应测试文件。
- 改动配置加载、启动流程时，至少验证：
  - `make test-unit`
  - `make lint`
- 改动格式后运行 `make format`，再运行 `make lint`。

## 调试注意点

- `boot.on_boot_run()` 在 FastAPI 应用创建前执行；启动阶段的问题不一定能通过接口复现。
- 新增测试时不要绕过 `test/conftest.py` 的 Dolphin mock 机制；它必须先于任何 `app` 模块导入。
- 健康检查访问日志会被过滤，不适合作为请求链路调试样本。
- `README.md` 中有部分历史信息与当前工作区不完全一致；涉及目录、部署文件、Docker/Helm 信息时，以仓库实际文件为准。
- 当前仓库根目录下没有 `docker-compose*.yml`，也没有可用的 `Dockerfile`；不要在项目级说明中写镜像重建或容器替换流程，除非后续实际补入这些文件。
- `build/`、`dist/`、`log/`、`.pytest_cache/`、`.ruff_cache/`、`data/` 下多数内容为构建产物或运行时数据，除非任务明确要求，否则不要优先修改这些目录。

## 参考文档

- `docs/README.md`：文档总览
- `docs/configuration/config_v2_implementation.md`：配置系统说明
- `docs/architecture/error-handling/README.md`：异常处理设计
- `docs/logging/struct_logger_usage.md`：结构化日志使用方式
- `docs/troubleshooting/keyboard_interrupt_issue.md`：启动/停止相关排障

## ai response
使用中文回复