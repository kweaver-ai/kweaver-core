# 快速开始指南

本指南帮助你快速安装和运行 Sandbox Executor。

## 前置要求

### 本地开发
- Python 3.11+
- macOS 或 Linux 系统
- Bubblewrap (Linux) 或 sandbox-exec (macOS，系统自带)

### Docker 部署
- Docker 20.10+
- Docker Compose 2.0+

## 本地开发

### 使用 uv（推荐）

```bash
# 进入项目目录
cd runtime/executor

# 设置 PYTHONPATH（必须）
export PYTHONPATH="$(cd .. && pwd)"

# 安装依赖
uv sync

# 启动服务（带自动重载）
uv run uvicorn executor.interfaces.http.rest:app --host 0.0.0.0 --port 8080 --reload
```

### 使用传统方式

```bash
# 进入项目目录
cd runtime/executor

# 设置 PYTHONPATH
export PYTHONPATH="$(cd .. && pwd)"

# 创建虚拟环境
python3 -m venv venv
source venv/bin/activate

# 安装依赖
pip install -r requirements.txt

# 启动服务
python3 -m executor.interfaces.http.rest
```

### 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 查看服务信息
curl http://localhost:8080/info

# 访问 API 文档
open http://localhost:8080/docs
```

期望输出：
```json
{
  "status": "healthy",
  "version": "1.0.0",
  "isolation": "bubblewrap|seatbelt",
  "platform": "Linux|Darwin"
}
```

## Docker 部署

### 构建镜像

```bash
cd runtime/executor
docker build -t sandbox-executor:v1.0 .
```

### 单独运行

```bash
docker run -d \
  --name sandbox-executor \
  --privileged \
  -p 8080:8080 \
  -e CONTROL_PLANE_URL=http://host.docker.internal:8000 \
  sandbox-executor:v1.0
```

### 使用 Docker Compose（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f executor

# 停止服务
docker-compose down
```

## 基本使用

### Python 示例

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "execution_id": "exec_001",
    "session_id": "session_001",
    "code": "def handler(event):\n    name = event.get(\"name\", \"World\")\n    return {\"message\": f\"Hello, {name}!\"}",
    "language": "python",
    "timeout": 10,
    "event": {"name": "Alice"}
  }'
```

### JavaScript 示例

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "execution_id": "exec_js_001",
    "session_id": "session_001",
    "code": "module.exports.handler = (event) => ({ message: `Hello ${event.name}!` });",
    "language": "javascript",
    "timeout": 10,
    "event": {"name": "Bob"}
  }'
```

### Shell 示例

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "execution_id": "exec_shell_001",
    "session_id": "session_001",
    "code": "echo \"Hello from shell!\" && ls -la /tmp",
    "language": "shell",
    "timeout": 10
  }'
```

## Handler 规范

### Python Handler

```python
def handler(event):
    """
    AWS Lambda 风格的 handler 函数

    Args:
        event: 包含输入数据的字典

    Returns:
        任意可 JSON 序列化的对象
    """
    name = event.get('name', 'World')
    print(f"Processing {name}...")

    return {
        'message': f'Hello, {name}!',
        'success': True
    }
```

### JavaScript Handler

```javascript
// CommonJS
module.exports.handler = (event, context) => {
    return {
        message: `Hello ${event.name}!`,
        timestamp: Date.now()
    };
};

// 或 ES6
export const handler = (event, context) => {
    return { result: 'ok' };
};
```

## 环境变量

| 变量名 | 说明 | 默认值 |
|--------|------|--------|
| `CONTROL_PLANE_URL` | Control Plane 地址 | `http://localhost:8000` |
| `WORKSPACE_PATH` | 工作目录路径 | `/workspace` |
| `EXECUTOR_PORT` | Executor 服务端口 | `8080` |
| `INTERNAL_API_TOKEN` | 内部 API 认证令牌 | 无 |

## 请求参数

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `execution_id` | string | 是 | 唯一执行标识符 |
| `session_id` | string | 是 | 会话标识符 |
| `code` | string | 是 | 要执行的代码 |
| `language` | string | 是 | `python`, `javascript`, `shell` |
| `timeout` | int | 否 | 超时时间（秒），默认 300 |
| `event` | dict | 否 | 传递给 handler 的事件数据 |
| `env_vars` | dict | 否 | 额外的环境变量 |

## 下一步

- [架构设计](architecture.md) - 了解系统架构和设计原理
- [API 文档](api-reference.md) - 查看完整的 API 参考
- [配置说明](configuration.md) - 详细配置选项
- [部署指南](deployment.md) - 生产环境部署
