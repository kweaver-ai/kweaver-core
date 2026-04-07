# 2.3 执行器


> **文档导航**: [返回架构总览](../architecture/overview.md)


### 2.3 执行器 (Executor)

执行器是运行在容器内的守护进程，负责接收执行请求并通过 Bubblewrap 启动用户代码，实现第二层隔离。

#### 2.3.1 执行器架构
执行器职责：

- 在容器启动时作为主进程运行
- 监听 HTTP 请求（来自管理中心）
- 接收用户代码和执行参数
- 调用 bwrap 命令隔离执行用户代码
- 收集执行结果（stdout、stderr、返回值、性能指标）
- 上报结果到管理中心

**执行模式**: AWS Lambda-style Handler

所有 Python 用户代码必须定义以下入口函数：

```python
def handler(event: dict) -> dict:
    """
    AWS Lambda-style handler 函数

    Args:
        event: 业务输入数据 (JSON 可序列化类型)

    Returns:
        返回值必须支持 JSON 序列化

    Raises:
        Exception: 业务逻辑异常会被捕获并记录到 stderr
    """
    # 用户业务逻辑
    result = process(event)
    return {"status": "ok", "data": result}
```

**Fileless Execution**: 执行器使用 `python3 -c` 直接在内存中执行代码，避免文件 I/O 操作。

核心实现：

```python
# sandbox-executor.py
# 运行在容器内的执行器进程

import asyncio
import json
import subprocess
import time
from pathlib import Path
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import httpx

app = FastAPI()

class ExecuteRequest(BaseModel):
    code: str
    language: str
    timeout: int = 30
    stdin: str = ""
    execution_id: str  # 执行 ID，用于上报结果

class ExecutionResult(BaseModel):
    status: str
    stdout: str
    stderr: str
    exit_code: int
    execution_time: float
    artifacts: list[str] = []
    return_value: dict | None = None  # 新增：handler 返回值
    metrics: dict | None = None  # 新增：性能指标

class SandboxExecutor:
    def __init__(self):
        self.workspace = Path("/workspace")
        self.workspace.mkdir(exist_ok=True)

        self.session_id = os.environ.get("SESSION_ID")
        self.control_plane_url = os.environ.get("CONTROL_PLANE_URL")
        self.internal_api_token = os.environ.get("INTERNAL_API_TOKEN")

        # Bubblewrap 配置
        self.bwrap_base_args = [
            "bwrap",
            # 只读挂载系统目录
            "--ro-bind", "/usr", "/usr",
            "--ro-bind", "/lib", "/lib",
            "--ro-bind", "/lib64", "/lib64",
            "--ro-bind", "/bin", "/bin",
            "--ro-bind", "/sbin", "/sbin",

            # 工作目录（可写）
            "--bind", str(self.workspace), "/workspace",
            "--chdir", "/workspace",

            # 临时目录
            "--tmpfs", "/tmp",

            # 最小化的 /proc 和 /dev
            "--proc", "/proc",
            "--dev", "/dev",

            # 隔离所有命名空间
            "--unshare-all",
            "--unshare-net",

            # 进程管理
            "--die-with-parent",
            "--new-session",

            # 环境变量清理
            "--clearenv",
            "--setenv", "PATH", "/usr/local/bin:/usr/bin:/bin",
            "--setenv", "HOME", "/workspace",
            "--setenv", "TMPDIR", "/tmp",

            # 安全选项
            "--cap-drop", "ALL",
            "--no-new-privs",
        ]

    def _generate_wrapper_code(self, user_code: str) -> str:
        """生成 Lambda-style wrapper 代码"""
        return f"""
import json
import sys

# User code
{user_code}

# Read event from stdin
try:
    input_data = sys.stdin.read()
    event = json.loads(input_data) if input_data.strip() else {{}}
except json.JSONDecodeError as e:
    print(f"Error parsing event JSON: {{e}}", file=sys.stderr)
    sys.exit(1)

# Call handler
try:
    if 'handler' not in globals():
        raise ValueError("必须定义 handler(event) 函数")

    result = handler(event)

    # Output result with markers
    print("\\n===SANDBOX_RESULT===")
    print(json.dumps(result))
    print("\\n===SANDBOX_RESULT_END===")

except Exception as e:
    import traceback
    print("\\n===SANDBOX_ERROR===")
    print(traceback.format_exc())
    print("\\n===SANDBOX_ERROR_END===")
    sys.exit(1)
"""

    async def execute_code(self, request: ExecuteRequest) -> ExecutionResult:
        """执行用户代码（通过 bwrap 隔离）"""
        execution_id = request.execution_id
        start_time = time.perf_counter()
        start_cpu = time.process_time()

        try:
            # 1. 根据语言构建执行命令
            if request.language == "python":
                # Fileless execution: 使用 python3 -c
                wrapper_code = self._generate_wrapper_code(request.code)

                exec_cmd = self.bwrap_base_args + [
                    "--",
                    "python3", "-c", wrapper_code
                ]

            elif request.language == "javascript":
                code_file = self.workspace / "user_code.js"
                code_file.write_text(request.code)

                exec_cmd = self.bwrap_base_args + [
                    "--ro-bind", str(code_file), "/workspace/user_code.js",
                    "--",
                    "node", "/workspace/user_code.js"
                ]

            elif request.language == "shell":
                exec_cmd = self.bwrap_base_args + [
                    "--",
                    "bash", "-c", request.code
                ]
            else:
                raise ValueError(f"Unsupported language: {request.language}")

            # 2. 在 bwrap 沙箱中执行代码
            result = subprocess.run(
                exec_cmd,
                input=request.stdin,
                capture_output=True,
                text=True,
                timeout=request.timeout,
                cwd=str(self.workspace)
            )

            duration_ms = (time.perf_counter() - start_time) * 1000
            cpu_time_ms = (time.process_time() - start_cpu) * 1000

            # 3. 解析返回值（Python handler 模式）
            return_value = None
            if request.language == "python":
                return_value = self._parse_return_value(result.stdout)

            # 4. 构建性能指标
            metrics = {
                "duration_ms": round(duration_ms, 2),
                "cpu_time_ms": round(cpu_time_ms, 2),
            }

            # 5. 收集执行结果
            execution_result = ExecutionResult(
                status="success" if result.returncode == 0 else "failed",
                stdout=result.stdout,
                stderr=result.stderr,
                exit_code=result.returncode,
                execution_time=duration_ms / 1000,  # 转换为秒
                artifacts=self._collect_artifacts(),
                return_value=return_value,
                metrics=metrics,
            )

        except subprocess.TimeoutExpired:
            duration_ms = (time.perf_counter() - start_time) * 1000
            execution_result = ExecutionResult(
                status="timeout",
                stdout="",
                stderr=f"Execution timeout after {request.timeout} seconds",
                exit_code=-1,
                execution_time=duration_ms / 1000,
                artifacts=[],
                metrics={"duration_ms": round(duration_ms, 2)},
            )

        except Exception as e:
            duration_ms = (time.perf_counter() - start_time) * 1000
            execution_result = ExecutionResult(
                status="error",
                stdout="",
                stderr=str(e),
                exit_code=-1,
                execution_time=duration_ms / 1000,
                artifacts=[],
                metrics={"duration_ms": round(duration_ms, 2)},
            )

        # 6. 上报结果到管理中心（通过内部 API）
        await self._report_result(execution_id, execution_result)

        return execution_result

    def _parse_return_value(self, stdout: str) -> dict | None:
        """从 stdout 解析 handler 返回值"""
        try:
            if "===SANDBOX_RESULT===" in stdout:
                start = stdout.find("===SANDBOX_RESULT===") + len("===SANDBOX_RESULT===")
                end = stdout.find("===SANDBOX_RESULT_END===")
                if start > 0 and end > start:
                    json_str = stdout[start:end].strip()
                    return json.loads(json_str)
        except (json.JSONDecodeError, ValueError):
            pass
        return None

    def _collect_artifacts(self) -> list[str]:
        """收集生成的文件"""
        artifacts = []
        for file_path in self.workspace.rglob("*"):
            if file_path.is_file() and not file_path.name.startswith("."):
                artifacts.append(str(file_path.relative_to(self.workspace)))
        return artifacts

    async def _report_result(self, execution_id: str, result: ExecutionResult):
        """上报执行结果到管理中心（通过内部 API）"""
        try:
            headers = {}
            if self.internal_api_token:
                headers["Authorization"] = f"Bearer {self.internal_api_token}"

            async with httpx.AsyncClient() as client:
                await client.post(
                    f"{self.control_plane_url}/internal/executions/{execution_id}/result",
                    json=result.dict(),
                    headers=headers,
                    timeout=10.0
                )
        except Exception as e:
            logger.error(f"Failed to report result for execution {execution_id}: {e}")

# FastAPI 端点
executor = SandboxExecutor()

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

@app.post("/execute")
async def execute(request: ExecuteRequest) -> ExecutionResult:
    """接收执行请求"""
    return await executor.execute_code(request)

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8080)
```

#### 2.3.2 执行结果格式

执行结果包含以下字段：

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | str | 执行状态：success/failed/timeout/error |
| `stdout` | str | 标准输出 |
| `stderr` | str | 标准错误 |
| `exit_code` | int | 进程退出码 |
| `execution_time` | float | 执行耗时（秒） |
| `artifacts` | list[str] | 生成的文件路径列表 |
| `return_value` | dict | **新增**：handler 函数返回值（JSON 可序列化） |
| `metrics` | dict | **新增**：性能指标（duration_ms、cpu_time_ms、peak_memory_mb 等） |

**返回格式示例**:

```json
{
  "status": "success",
  "stdout": "Processing complete.\\n",
  "stderr": "",
  "exit_code": 0,
  "execution_time": 0.07523,
  "artifacts": ["output.csv"],
  "return_value": {
    "status": "ok",
    "data": [1, 2, 3]
  },
  "metrics": {
    "duration_ms": 75.23,
    "cpu_time_ms": 68.12,
    "peak_memory_mb": 42.5
  }
}
```

**metrics 字段格式**:

```json
{
  "duration_ms": 75.23,     // 墙钟耗时（毫秒）
  "cpu_time_ms": 68.12,     // CPU 时间（毫秒）
  "peak_memory_mb": 42.5,    // 内存峰值（MB），可选
  "io_read_bytes": 1024,     // 读取字节数，可选
  "io_write_bytes": 2048     // 写入字节数，可选
}
```

优势：
- 扩展性好：添加新指标无需修改表结构
- 灵活性高：不同执行类型可包含不同指标
- 查询方便：MySQL 5.7+ 支持 JSON 字段索引和查询

#### 2.3.3 Bubblewrap 安全配置详解
完整的 bwrap 命令示例：

```bash
bwrap \
  # === 文件系统隔离 ===
  # 只读挂载系统目录
  --ro-bind /usr /usr \
  --ro-bind /lib /lib \
  --ro-bind /lib64 /lib64 \
  --ro-bind /bin /bin \
  --ro-bind /sbin /sbin \
  
  # 工作目录（读写）
  --bind /workspace /workspace \
  --chdir /workspace \
  
  # 临时目录（内存文件系统）
  --tmpfs /tmp \
  
  # === 命名空间隔离 ===
  --unshare-all \        # 隔离所有命名空间（PID、NET、MNT、IPC、UTS、USER）
  --share-net \          # 可选：如果需要网络访问
  
  # === 进程隔离 ===
  --proc /proc \         # 挂载 /proc（只能看到沙箱内进程）
  --dev /dev \           # 最小化的 /dev
  --die-with-parent \    # 父进程终止时自动终止
  --new-session \        # 新的会话
  
  # === 环境隔离 ===
  --clearenv \           # 清除所有环境变量
  --setenv PATH /usr/local/bin:/usr/bin:/bin \
  --setenv HOME /workspace \
  --setenv TMPDIR /tmp \
  --unsetenv TERM \      # 清除终端环境
  
  # === 安全限制 ===
  --cap-drop ALL \       # 删除所有 Linux Capabilities
  --no-new-privs \       # 禁止获取新权限
  
  # === 资源限制（可选，与 ulimit 配合）===
  --rlimit NPROC=128 \   # 最大进程数
  --rlimit NOFILE=1024 \ # 最大文件描述符
  
  # === 执行命令 ===
  -- \
  python3 /workspace/user_code.py
```

安全特性说明：

| 隔离层面 | 容器隔离         | Bubblewrap隔离               |
| -------- | ---------------- | ---------------------------- |
| 文件系统 | Union FS, 独立层 | 只读绑定, tmpfs              |
| 网络     | NetworkMode=none | unshare network namespace    |
| 进程     | PID namespace    | 新 PID namespace（PID 1）|
| IPC      | IPC namespace    | 新 IPC namespace             |
| 用户     | 非特权用户       | 进一步限制 capabilities      |
| 系统调用 | Seccomp 过滤     | 额外的 seccomp 过滤          |
| 资源     | cgroup 限制      | ulimit 限制                  |
