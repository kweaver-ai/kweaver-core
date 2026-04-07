# 开发指南

本文档介绍如何参与 Sandbox Executor 的开发，包括环境设置、代码规范、测试和贡献流程。

## 目录

- [开发环境设置](#开发环境设置)
- [代码风格](#代码风格)
- [运行测试](#运行测试)
- [项目结构约定](#项目结构约定)
- [添加新功能](#添加新功能)
- [性能测试](#性能测试)

---

## 开发环境设置

### 使用 uv（推荐）

```bash
# 进入项目目录
cd runtime/executor

# 设置 PYTHONPATH
export PYTHONPATH="$(cd .. && pwd)"

# 安装依赖
uv sync

# 安装开发依赖
uv sync --dev
```

### 使用传统方式

```bash
# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 安装依赖
pip install -r requirements.txt

# 安装开发依赖
pip install -r requirements-dev.txt
```

### 安装 pre-commit hooks

```bash
# 安装 pre-commit
pip install pre-commit

# 安装 hooks
pre-commit install

# 手动运行 hooks
pre-commit run --all-files
```

### 验证安装

```bash
# 检查 Python 版本
python --version  # 应该是 3.11+

# 检查 Bubblewrap/sandbox-exec
python -c "from executor.infrastructure.isolation.bwrap import get_bwrap_version; print(get_bwrap_version())"
```

---

## 代码风格

### Python 代码规范

遵循 PEP 8 规范，使用以下工具：

```bash
# 代码格式化
black executor/

# 导入排序
isort executor/

# 代码检查
flake8 executor/

# 类型检查
mypy executor/
```

### Black 配置

```ini
# pyproject.toml
[tool.black]
line-length = 100
target-version = ['py311']
include = '\.pyi?$'
```

### isort 配置

```ini
# pyproject.toml
[tool.isort]
profile = "black"
line_length = 100
```

### flake8 配置

```ini
# .flake8
[flake8]
max-line-length = 100
extend-ignore = E203, W503
exclude = .git,__pycache__,venv
```

### 文档字符串

使用 Google 风格的文档字符串：

```python
def execute_code(execution: Execution) -> ExecutionResult:
    """Execute code in isolated environment.

    Args:
        execution: The execution entity containing code and context.

    Returns:
        ExecutionResult containing stdout, stderr, and exit code.

    Raises:
        ExecutionTimeoutError: If execution exceeds timeout.
        IsolationError: If isolation setup fails.
    """
    pass
```

---

## 运行测试

### 单元测试

```bash
# 运行所有单元测试
pytest executor/tests/unit/ -v

# 运行特定测试文件
pytest executor/tests/unit/test_bwrap.py -v

# 运行特定测试
pytest executor/tests/unit/test_bwrap.py::test_get_bwrap_version -v
```

### 集成测试

```bash
# 运行所有集成测试
pytest executor/tests/integration/ -v

# 需要先启动服务
python3 -m executor.interfaces.http.rest
```

### 测试覆盖率

```bash
# 生成覆盖率报告
pytest executor/tests/ --cov=executor --cov-report=html

# 查看报告
open htmlcov/index.html
```

### 并发测试

测试异步执行能力：

```bash
# 启动服务
python3 -m executor.interfaces.http.rest

# 并发执行测试
for i in {1..5}; do
  curl -X POST http://localhost:8080/execute \
    -H 'Content-Type: application/json' \
    -d "{
      \"execution_id\": \"concurrent_$i\",
      \"session_id\": \"test\",
      \"code\": \"import time; time.sleep(2); def handler(e): return {'done': True}\",
      \"language\": \"python\",
      \"timeout\": 10
    }" &
done
wait

# 如果真正异步，所有请求应该在 ~2 秒内完成（而不是 10 秒）
```

---

## 项目结构约定

### 六边形架构层次

```
executor/
├── application/          # 应用层
│   ├── commands/        # 命令处理（Use Cases）
│   └── services/        # 应用服务
├── domain/              # 领域层
│   ├── entities/        # 领域实体
│   ├── value_objects/   # 值对象
│   └── ports/           # 端口接口（抽象）
├── infrastructure/      # 基础设施层
│   ├── http/           # HTTP 实现
│   ├── isolation/      # 隔离实现
│   └── persistence/    # 持久化实现
└── interfaces/         # 接口层（REST API）
```

### 依赖规则

- **domain** 层不依赖任何其他层
- **application** 层只依赖 domain 层
- **infrastructure** 层实现 domain 层的接口
- **interfaces** 层依赖 application 层

### 命名约定

| 类型 | 约定 | 示例 |
|------|------|------|
| 模块 | 小写下划线 | `execute_code.py` |
| 类 | 大驼峰 | `ExecuteCodeCommand` |
| 函数/方法 | 小写下划线 | `execute_code()` |
| 常量 | 大写下划线 | `MAX_TIMEOUT` |
| 私有成员 | 下划线前缀 | `_internal_method()` |

---

## 添加新功能

### 添加新的隔离适配器

1. **创建适配器类**

```python
# infrastructure/isolation/my_adapter.py
from executor.domain.ports import IIsolationPort
from executor.domain.entities import Execution
from executor.domain.value_objects import ExecutionResult

class MyIsolationAdapter(IIsolationPort):
    """Custom isolation adapter."""

    async def execute(self, execution: Execution) -> ExecutionResult:
        """Execute code with custom isolation."""
        # 实现执行逻辑
        pass

    def is_available(self) -> bool:
        """Check if adapter is available."""
        # 检查可用性
        pass

    def get_version(self) -> str:
        """Return adapter version."""
        # 返回版本
        pass
```

2. **注册适配器**

```python
# interfaces/http/rest.py
from executor.infrastructure.isolation.my_adapter import MyIsolationAdapter

# 根据平台选择适配器
if MyIsolationAdapter.is_available():
    isolation_adapter = MyIsolationAdapter(workspace_path)
```

### 添加新的语言支持

1. **扩展语言枚举**

```python
# domain/value_objects/language.py
from enum import Enum

class Language(str, Enum):
    PYTHON = "python"
    JAVASCRIPT = "javascript"
    SHELL = "shell"
    RUST = "rust"  # 新增语言
```

2. **实现执行逻辑**

```python
# infrastructure/isolation/bwrap.py
async def _execute_rust(self, code: str, ...) -> ExecutionResult:
    """Execute Rust code."""
    # 实现 Rust 代码执行
    pass
```

### 添加新的 API 端点

```python
# interfaces/http/rest.py
from fastapi import APIRouter

router = APIRouter()

@router.post("/custom_endpoint")
async def custom_endpoint(request: CustomRequest):
    """Custom endpoint description."""
    # 实现端点逻辑
    pass

# 注册路由
app.include_router(router)
```

---

## 性能测试

### 基准测试

```python
import asyncio
import httpx
import time

async def benchmark_concurrent_execution(n=100):
    """Benchmark concurrent execution performance."""
    async with httpx.AsyncClient() as client:
        start_time = time.time()

        tasks = [
            client.post('http://localhost:8080/execute', json={
                'execution_id': f'bench_{i}',
                'session_id': 'benchmark',
                'code': 'def handler(e): return {"done": True}',
                'language': 'python',
                'timeout': 10
            })
            for i in range(n)
        ]

        results = await asyncio.gather(*tasks)
        duration = time.time() - start_time

        print(f'Completed {n} requests in {duration:.2f}s')
        print(f'Throughput: {n/duration:.2f} req/s')

asyncio.run(benchmark_concurrent_execution(100))
```

### 性能分析

```bash
# 使用 cProfile
python -m cProfile -o profile.stats -m executor.interfaces.http.rest

# 分析结果
python -c "
import pstats
p = pstats.Stats('profile.stats')
p.sort_stats('cumulative')
p.print_stats(20)
"
```

---

## 调试技巧

### 启用调试日志

```bash
export LOG_LEVEL=DEBUG
python3 -m executor.interfaces.http.rest
```

### 使用 Python 调试器

```python
import pdb; pdb.set_trace()

# 或使用 ipdb（如果已安装）
import ipdb; ipdb.set_trace()
```

### VS Code 调试配置

```json
{
  "name": "Debug Executor",
  "type": "python",
  "request": "launch",
  "module": "executor.interfaces.http.rest",
  "env": {
    "PYTHONPATH": "${workspaceFolder}/runtime",
    "LOG_LEVEL": "DEBUG"
  }
}
```

---

## 相关文档

- [架构设计](architecture.md) - 了解六边形架构
- [配置说明](configuration.md) - 开发环境配置
- [API 文档](api-reference.md) - API 端点参考
