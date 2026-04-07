# 架构设计

本文档介绍 Sandbox Executor 的架构设计和技术原理。

## 目录

- [设计目标](#设计目标)
- [安全隔离模型](#安全隔离模型)
- [六边形架构](#六边形架构)
- [模块结构](#模块结构)
- [异步执行模型](#异步执行模型)
- [技术栈](#技术栈)

---

## 设计目标

- **安全性第一** - 多层隔离（容器 + 进程隔离）
- **高性能** - 异步架构，支持高并发执行
- **兼容性** - 支持 AWS Lambda handler 规范
- **可观测性** - 实时心跳、生命周期管理、执行指标

---

## 安全隔离模型

Sandbox Executor 采用双层隔离策略，确保不受信任的代码在受控环境中安全执行。

```
┌─────────────────────────────────────────────┐
│         Docker 容器隔离（第一层）             │
│  • NetworkMode=none                          │
│  • CAP_DROP=ALL                              │
│  • 非特权用户 (UID:GID=1000:1000)            │
│  ┌───────────────────────────────────────┐  │
│  │   Bubblewrap/sandbox-exec（第二层）     │  │
│  │  ┌─────────────────────────────────┐  │  │
│  │  │      用户代码执行                │  │  │
│  │  │  • PID namespace               │  │  │
│  │  │  • Network namespace           │  │  │
│  │  │  • Mount namespace             │  │  │
│  │  │  • Seccomp 过滤                │  │  │
│  │  │  • 只读文件系统                 │  │  │
│  │  └─────────────────────────────────┘  │  │
│  └───────────────────────────────────────┘  │
└─────────────────────────────────────────────┘
```

### 第一层：容器隔离

- **网络隔离**：`NetworkMode=none` 完全禁用网络
- **权限限制**：`CAP_DROP=ALL` 丢弃所有特权
- **用户限制**：使用非特权用户运行

### 第二层：进程隔离

**Linux (Bubblewrap)**
- PID namespace：进程隔离
- Network namespace：网络隔离
- Mount namespace：文件系统隔离
- Seccomp filtering：系统调用过滤
- 只读绑定：`/usr`, `/lib` 等目录只读挂载

**macOS (Seatbelt)**
- 原生 `sandbox-exec` 集成
- 进程级沙箱配置
- 文件访问控制

---

## 六边形架构

Sandbox Executor 采用六边形架构（Hexagonal Architecture），实现清晰的关注点分离。

```
                    ┌─────────────────────┐
                    │   HTTP Interface    │
                    │   (FastAPI/REST)     │
                    └──────────┬───────────┘
                               │
        ┌──────────────────────┼──────────────────────┐
        │                      │                      │
┌───────▼────────┐    ┌───────▼───────┐    ┌───────▼────────┐
│ Execute Code   │    │  Isolation    │    │    Callback     │
│    Command     │◄──►│    Port       │    │     Port        │
│                │    │               │    │                │
│  • Orchestrate  │    │  • Bubblewrap │    │  • HTTP Client  │
│  • Timeout     │    │  • Seatbelt   │    │  • Retry Logic  │
│  • Heartbeat   │    │  • Abstraction│    │  • Fallback     │
└────────────────┘    └───────────────┘    └────────────────┘
        │                      │
        └──────────────────────┼──────────────────────┘
                               │
                    ┌──────────▼───────────┐
                    │  Value Objects      │
                    │  • ExecutionResult  │
                    │  • ExecutionStatus  │
                    │  • ExecutionContext│
                    └─────────────────────┘
```

### 架构层次

**接口层 (Interfaces)**
- HTTP REST API (FastAPI)
- 请求验证和响应格式化

**应用层 (Application)**
- 命令模式（Commands）
- 应用服务（Services）
- 业务流程编排

**领域层 (Domain)**
- 领域实体（Entities）
- 值对象（Value Objects）
- 端口接口（Ports）

**基础设施层 (Infrastructure)**
- 隔离适配器（Isolation Adapters）
- HTTP 客户端
- 持久化实现

---

## 模块结构

```
executor/
├── application/          # 应用层
│   ├── commands/        # 命令模式
│   │   └── execute_code.py
│   └── services/        # 应用服务
│       ├── heartbeat_service.py
│       └── lifecycle_service.py
├── domain/              # 领域层
│   ├── entities/        # 实体
│   │   └── execution.py
│   ├── value_objects/   # 值对象
│   │   ├── execution_result.py
│   │   ├── execution_status.py
│   │   └── ...
│   └── ports/           # 端口接口
│       ├── isolation_port.py
│       ├── callback_port.py
│       └── ...
├── infrastructure/      # 基础设施层
│   ├── http/           # HTTP 客户端
│   │   └── callback_client.py
│   ├── isolation/      # 隔离适配器
│   │   ├── bwrap.py    # Bubblewrap Runner
│   │   ├── macseatbelt.py  # macOS Seatbelt Runner
│   │   └── result_parser.py
│   └── persistence/    # 持久化
│       └── artifact_scanner.py
└── interfaces/         # 接口层
    └── http/
        └── rest.py     # FastAPI 端点
```

### 核心组件说明

| 组件 | 职责 |
|------|------|
| `ExecuteCodeCommand` | 编排代码执行流程 |
| `IIsolationPort` | 隔离技术抽象接口 |
| `BubblewrapRunner` | Linux Bubblewrap 实现 |
| `SeatbeltRunner` | macOS sandbox-exec 实现 |
| `CallbackPort` | 结果回调接口 |
| `HeartbeatService` | 心跳上报服务 |
| `LifecycleService` | 容器生命周期管理 |

---

## 异步执行模型

Executor 使用完全异步架构，确保高并发和高性能。

### 真正的异步执行

```python
# 使用 asyncio.create_subprocess_exec() 实现真正的异步
proc = await asyncio.create_subprocess_exec(
    *bwrap_args,
    stdout=asyncio.subprocess.PIPE,
    stderr=asyncio.subprocess.PIPE,
    cwd=self.workspace_path
)

# 带超时的等待
try:
    stdout, stderr = await asyncio.wait_for(
        proc.communicate(),
        timeout=timeout
    )
except asyncio.TimeoutError:
    # 正确终止子进程
    proc.kill()
    await proc.wait()
```

### 优势

- 不阻塞事件循环
- 支持高并发执行
- 超时能够正确终止子进程
- 资源高效利用

### 并发示例

```python
import asyncio
import httpx

async def execute_concurrent():
    async with httpx.AsyncClient() as client:
        tasks = [
            client.post('http://localhost:8080/execute', json={...})
            for _ in range(100)
        ]
        results = await asyncio.gather(*tasks)
        print(f"Executed {len(results)} requests concurrently")

asyncio.run(execute_concurrent())
```

---

## 技术栈

| 组件 | 技术选型 | 说明 |
|------|---------|------|
| HTTP 框架 | FastAPI + Uvicorn | 高性能异步框架 |
| 隔离技术 | Bubblewrap / sandbox-exec | 进程级隔离 |
| 异步运行时 | asyncio | Python 原生异步 |
| 日志 | structlog | 结构化日志 |
| 数据验证 | Pydantic | 数据验证和序列化 |
| HTTP 客户端 | httpx | 异步 HTTP 客户端 |

---

## 扩展性

### 添加新的隔离适配器

1. 实现 `IIsolationPort` 接口
2. 继承基类模式
3. 在 `interfaces/http/rest.py` 中注册

```python
from executor.domain.ports import IIsolationPort
from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult

class MyIsolationRunner(IIsolationPort):
    async def execute(self, execution: Execution) -> ExecutionResult:
        # 实现代码
        pass

    def is_available(self) -> bool:
        # 检查可用性
        pass

    def get_version(self) -> str:
        # 返回版本
        pass
```

### 添加新的语言支持

在相应的隔离适配器中添加语言特定的执行逻辑。

---

## 相关文档

- [配置说明](configuration.md) - 详细配置选项
- [开发指南](development.md) - 开发环境设置
- [API 文档](api-reference.md) - RESTful API 参考
