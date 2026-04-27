# agent-executor

Agent-Executor 是一个 AI Agent 执行器， 基于 Agent-Factory 提供的 Agent 配置，构建 Agent Context, 调用 Dolphin SDK 运行 Agent。支持 **自然语言模式** 和 **Dolphin代码模式**， 对外提供了 运行 和 DEBUG 接口。另外 Agent-Executor 承担了 内置 Agent、和部分内置工具的配置管理和初始化逻辑。


## Requirements
- Python 3.10+
- [uv](https://github.com/astral-sh/uv) (推荐, 但可选)
## 系统架构


## 项目 Layout


```
├── app
│   ├── common          # 公共组件层
│   │   ├── errors      # 错误处理模块
│   │   ├── international # 国际化支持 （deprecated）
│   │   ├── config.py   # 全局配置
│   │   ├── prompt.py   # 提示词管理 （deprecated）
│   │   └── structs.py  # 数据结构定义
│   ├── driven          # 外部服务驱动层 （这一层分类有些奇怪，后面优化）
│   │   ├── ad          # AD服务集成 （deprecated）
│   │   ├── anyshare    # AnyShare 文档服务
│   │   ├── dip         # DIP平台集成
│   │   ├── external    # 外部API客户端
│   │   └── infrastructure # 基础设施服务（Nebula/Opensearch/Redis）
│   ├── logic           # 业务逻辑层
│   │   ├── retriever   # 内容检索模块
│   │   ├── tool        # 工具调用模块 （部分内置工具的 Server 端）
│   │   ├── agent_core.py # Agent核心逻辑
│   │   └── sensitive_word_detection.py # 敏感词检测 （deprecated）
│   ├── router          # API路由层
│   │   ├── agent_controller.py # Agent相关接口
│   │   ├── file_controller.py # 文件操作接口 （deprecated）
│   │   └── tool_controller.py # 工具调用接口
│   └── utils           # 工具类库
│       ├── common.py   # 通用工具函数
│       ├── graph_util.py # 图数据库操作工具
│       └── text_parser.py # 文本解析工具
├── data_migrations     # 数据迁移脚本
├── helm                # Kubernetes部署配置
├── test                # 测试套件
├── Dockerfile          # 容器化构建文件
└── main.py             # 项目启动入口
```

## 日志说明
request 日志、服务运行日志默认 通过文件记录，在项目目录下的 log目录，
dolphin日志部份通过 STDOUT 记录， 部份通过日志文件记录
log下日志文件说明：
- dolphin.log: dolphin日志
- request.log: request日志
- agent-executor.log: agent-executor运行


## 本地开发

### 推荐使用 uv 管理项目（推荐）

#### 1. 安装 uv
```bash
# macOS
curl -LsSf https://astral.sh/uv/install.sh | sh

# Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm astral.sh/uv/install.sh | iex"
```

#### 2. 快速环境设置（推荐）
```bash
# 1. 复制配置文件
cp agent-executor.yaml.example agent-executor.yaml

# 2. 自动同步环境并安装所有依赖
make uv-sync

# 4. 根据需要修改配置
vi agent-executor.yaml
```

### 使用 uv 管理项目（推荐）

#### 1. 安装 uv
```bash
# 使用 pip 安装（推荐）
pip install uv

# 或使用官方安装脚本（备选）
# macOS
curl -LsSf https://astral.sh/uv/install.sh | sh

# Linux
curl -LsSf https://astral.sh/uv/install.sh | sh

# Windows
powershell -c "irm astral.sh/uv/install.sh | iex"
```

#### 2. 快速环境设置（推荐）
```bash
# 1. 复制配置文件
cp agent-executor.yaml.example agent-executor.yaml

# 2. 自动同步环境并安装所有依赖
make uv-sync

# 4. 根据需要修改配置
vi agent-executor.yaml
```

## 安装依赖

**注意：** 如果使用上面的快速环境设置（`make uv-sync`），以下依赖已自动安装，无需手动执行。

### 手动安装依赖（如果需要）
```bash
# DolphinSDK 依赖安装
git clone https://devops.aishu.cn/AISHUDevOps/DIP/_git/dolphin-language
cd dolphin-language
# 切换到指定分支
git checkout MISSION
# 安装 DolphinSDK
uv pip install .
```

## 开启 DEBUG 模式
开启后会影响日志打印级别， 默认为 INFO级别，开启后为 DEBUG级别， 同时会影响配置文件的加载，详见config.py
**tip: ** : 只允许在开发环境开启 {ip} 为依赖服务所在服务器IP
```
export AGENT_EXECUTOR_DEBUG=true
export AGENT_EXECUTOR_DIPHOST={ip}
```
## 开启可观测性
通过 `agent-executor.yaml` 中的 `o11y` 配置开启，例如：

```yaml
o11y:
  service_name: "agent-executor"
  service_version: "1.0.0"
  environment: "production"
  log_enabled: true
  log_level: "info"
  trace_enabled: true
  trace_endpoint: "otelcol-contrib:4318"
  trace_sampling_rate: 1.0
```


## 依赖服务配置
```

# 依赖服务端口映射
kubectl port-forward svc/mf-model-manager -n dip --address=0.0.0.0 9888:9898 > /dev/null & \
kubectl port-forward svc/mf-model-api -n dip --address=0.0.0.0 9898:9898 > /dev/null & \
kubectl port-forward svc/agent-operator-integration  -n dip --address=0.0.0.0 39000:9000 > /dev/null & \
kubectl port-forward svc/agent-factory     -n dip --address=0.0.0.0 13020:13020 > /dev/null & \
kubectl port-forward svc/agent-memory  -n dip --address=0.0.0.0 30790:30790 > /dev/null & \
kubectl port-forward svc/proton-redis-proton-redis  -n resource --address=0.0.0.0 6379:6379 > /dev/null &

# 记得开启防火墙端口
firewall-cmd --zone=public --add-port={xxxx}/tcp --permanent
```
## 启动

### 使用 Makefile 启动（推荐）
```bash
# 开发模式运行
make run

# 环境设置命令
make uv-sync                    # 自动同步环境并安装依赖

# make uv-sync 特殊说明
make uv-sync 包含了是以下命令的组合：
    uv sync --all-groups

# 其他可用命令
make help  # 查看所有可用命令
```

### 直接启动
```bash
# 使用 uv 运行 或者 make 运行
uv run main.py
make run
```

## 代码质量

运行代码检查（包括 ruff lint 和 format）：

```bash
make lint
```

或者使用 uv 直接运行：

```bash
uv run pre-commit run -a --config .pre-commit-config.yaml
```
