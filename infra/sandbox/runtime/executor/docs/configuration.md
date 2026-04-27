# 配置说明

本文档介绍 Sandbox Executor 的配置选项，包括环境变量、隔离配置和资源限制。

## 目录

- [环境变量](#环境变量)
- [Bubblewrap 配置](#bubblewrap-配置)
- [Seatbelt 配置](#seatbelt-配置)
- [资源限制](#资源限制)
- [日志配置](#日志配置)

---

## 环境变量

### 服务配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `CONTROL_PLANE_URL` | string | `http://localhost:8000` | Control Plane 服务地址 |
| `WORKSPACE_PATH` | string | `/workspace` | 代码执行工作目录 |
| `EXECUTOR_PORT` | int | `8080` | Executor HTTP 服务端口 |
| `INTERNAL_API_TOKEN` | string | 无 | 内部 API 认证令牌 |

### 资源限制

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `MAX_MEMORY_MB` | int | `512` | 单次执行最大内存（MB） |
| `MAX_EXECUTION_TIME` | int | `300` | 默认最大执行时间（秒） |
| `MAX_CONCURRENT_EXECUTIONS` | int | `100` | 最大并发执行数 |

### 日志配置

| 变量名 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `LOG_LEVEL` | string | `INFO` | 日志级别：`DEBUG`, `INFO`, `WARNING`, `ERROR` |
| `LOG_FORMAT` | string | `json` | 日志格式：`json`, `text` |

### 设置环境变量

```bash
# Linux/macOS
export CONTROL_PLANE_URL="http://localhost:8000"
export WORKSPACE_PATH="/workspace"
export EXECUTOR_PORT="8080"
export LOG_LEVEL="DEBUG"

# 或在 Docker 启动时设置
docker run -e CONTROL_PLANE_URL="http://localhost:8000" ...
```

---

## Bubblewrap 配置

Bubblewrap (bwrap) 是 Linux 上的用户空间沙箱工具。

### 基础参数

```python
BWRAP_ARGS = [
    # 只读绑定系统目录
    "--ro-bind", "/usr", "/usr",
    "--ro-bind", "/lib", "/lib",
    "--ro-bind", "/lib64", "/lib64",

    # 可写工作目录
    "--bind", workspace_path, "/workspace",

    # 隔离命名空间
    "--unshare-all",       # 取消共享所有命名空间
    "--unshare-net",       # 网络隔离
    "--die-with-parent",   # 父进程退出时终止

    # 进程隔离
    "--proc", "/proc",
    "--dev", "/dev",
]
```

### 参数说明

| 参数 | 说明 |
|------|------|
| `--ro-bind src dst` | 只读绑定挂载 |
| `--bind src dst` | 可写绑定挂载 |
| `--unshare-all` | 取消共享所有命名空间 |
| `--unshare-net` | 隔离网络命名空间 |
| `--die-with-parent` | 父进程退出时终止子进程 |
| `--proc` | 挂载 proc 文件系统 |
| `--dev` | 挂载 dev 文件系统 |

### Python 执行配置

```python
# Python 解释器路径
PYTHON_INTERPRETER = "/usr/bin/python3"

# Python 模块搜索路径
PYTHONPATH = [
    "/workspace",
    "/usr/lib/python3.x",
]
```

### 安全配置

```python
# 禁用特定系统调用
SECCOMP_FILTER = """
# 禁用网络相关系统调用
# 禁用文件系统修改
...
"""
```

---

## Seatbelt 配置

macOS 使用原生的 sandbox-exec 进行隔离。

### 沙箱配置文件

```bash
(version 1)
(deny default)
(allow process-exec)
(allow file-read*)
(allow file-write* (subpath "/tmp"))
(allow system*)
```

### 规则说明

| 规则 | 说明 |
|------|------|
| `(deny default)` | 默认拒绝所有操作 |
| `(allow process-exec)` | 允许执行进程 |
| `(allow file-read*)` | 允许读取文件 |
| `(allow file-write* (subpath "/tmp"))` | 允许写入 /tmp 目录 |
| `(allow system*)` | 允许系统调用 |

### 配置文件位置

```
runtime/executor/infrastructure/isolation/macseatbelt.py
```

---

## 资源限制

### 超时控制

```python
# API 层超时
API_TIMEOUT = 3600  # 最大 1 小时

# 事件参数超时
DEFAULT_TIMEOUT = 300  # 默认 5 分钟
MAX_TIMEOUT = 3600     # 最大 1 小时

# Daemon 层超时
DAEMON_TIMEOUT = 300
```

### 超时参数 `__timeout`

事件对象中的 `__timeout` 参数用于覆盖默认超时：

```python
event = {
    "data": "...",
    "__timeout": 600  # 10 分钟超时
}
```

**注意**: `__timeout` 使用双下划线前缀以避免与用户参数冲突。

### 内存限制

```python
# ulimit 限制
ULIMIT_CONFIG = {
    "as": 512 * 1024 * 1024,  # 地址空间 512MB
    "rss": 512 * 1024 * 1024,  # 常驻内存 512MB
}
```

### CPU 时间

```python
# CPU 时间限制
CPU_TIME_LIMIT = 5 * 1000  # 5 秒（毫秒）

# Wall time 限制
WALL_TIME_LIMIT = 10 * 1000  # 10 秒（毫秒）
```

### 进程数限制

```python
# 最大进程数
MAX_PROCESSES = 100

# 最大线程数
MAX_THREADS = 500
```

---

## 日志配置

### 日志级别

| 级别 | 说明 | 使用场景 |
|------|------|----------|
| `DEBUG` | 详细调试信息 | 开发调试 |
| `INFO` | 一般信息 | 生产环境 |
| `WARNING` | 警告信息 | 需要注意的情况 |
| `ERROR` | 错误信息 | 错误和异常 |

### 日志格式

**JSON 格式**（生产环境推荐）

```json
{
  "timestamp": "2024-01-09T10:00:00Z",
  "level": "INFO",
  "message": "Execution completed",
  "execution_id": "exec_001",
  "duration_ms": 1234
}
```

**文本格式**（开发环境推荐）

```
2024-01-09 10:00:00 [INFO] Execution completed (exec_001) - 1234ms
```

### 配置示例

```python
import structlog

# 配置 structlog
structlog.configure(
    processors=[
        structlog.stdlib.add_log_level,
        structlog.stdlib.add_logger_name,
        structlog.processors.TimeStamper(fmt="iso"),
        structlog.processors.JSONRenderer()
    ],
    wrapper_class=structlog.stdlib.BoundLogger,
    logger_factory=structlog.stdlib.LoggerFactory(),
)
```

### 日志输出

```bash
# 开发模式（文本格式）
export LOG_LEVEL=DEBUG
export LOG_FORMAT=text
python3 -m executor.interfaces.http.rest

# 生产模式（JSON 格式）
export LOG_LEVEL=INFO
export LOG_FORMAT=json
python3 -m executor.interfaces.http.rest
```

---

## 完整配置示例

### 开发环境

```bash
# .env.development
CONTROL_PLANE_URL=http://localhost:8000
WORKSPACE_PATH=/tmp/sandbox_workspace
EXECUTOR_PORT=8080
LOG_LEVEL=DEBUG
LOG_FORMAT=text
MAX_EXECUTION_TIME=600
```

### 生产环境

```bash
# .env.production
CONTROL_PLANE_URL=http://control-plane:8000
WORKSPACE_PATH=/workspace
EXECUTOR_PORT=8080
LOG_LEVEL=INFO
LOG_FORMAT=json
MAX_EXECUTION_TIME=300
INTERNAL_API_TOKEN=your-secret-token
```

### Docker Compose

```yaml
version: '3.8'
services:
  executor:
    image: sandbox-executor:v1.0
    cap_drop:
      - ALL
    security_opt:
      - seccomp=default
      - no-new-privileges
    environment:
      - CONTROL_PLANE_URL=http://control-plane:8000
      - WORKSPACE_PATH=/workspace
      - EXECUTOR_PORT=8080
      - LOG_LEVEL=INFO
      - MAX_MEMORY_MB=512
    ports:
      - "8080:8080"
```

**注意**: 使用上述配置前，请确保宿主机已启用用户命名空间（参见 [用户命名空间配置指南](user-namespace-guide.md)）。

---

## 相关文档

- [用户命名空间配置](user-namespace-guide.md) - 双层隔离原理和配置
- [快速开始](quick-start.md) - 环境设置和基本配置
- [架构设计](architecture.md) - 了解隔离机制
- [部署指南](deployment.md) - 生产环境配置
