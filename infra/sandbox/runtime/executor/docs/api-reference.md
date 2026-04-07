# API 参考文档

本文档提供 Sandbox Executor RESTful API 的详细参考。

## 目录

- [概述](#概述)
- [执行代码](#执行代码)
- [健康检查](#健康检查)
- [服务信息](#服务信息)
- [错误码](#错误码)

---

## 概述

### 基础 URL

```
http://localhost:8080
```

### 认证

内部 API 使用 `INTERNAL_API_TOKEN` 环境变量配置的令牌进行认证：

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" http://localhost:8080/execute
```

### 支持的语言

- `python` - Python 3.11+
- `javascript` - Node.js
- `shell` - Bash shell

---

## 执行代码

执行用户代码并返回结果。

### 端点

```
POST /execute
```

### 请求头

| 头部 | 值 |
|------|-----|
| Content-Type | application/json |
| Authorization | Bearer YOUR_TOKEN (可选) |

### 请求体

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `execution_id` | string | 是 | 唯一执行标识符 |
| `session_id` | string | 是 | 会话标识符 |
| `code` | string | 是 | 要执行的代码 |
| `language` | string | 是 | `python`, `javascript`, `shell` |
| `timeout` | int | 否 | 超时时间（秒），范围 1-3600，默认 300 |
| `event` | object | 否 | 传递给 handler 的事件数据 |
| `env_vars` | object | 否 | 额外的环境变量 |

### 响应

**成功 (200 OK)**

```json
{
  "execution_id": "exec_001",
  "status": "completed",
  "message": "Execution completed"
}
```

**错误 (4xx/5xx)**

```json
{
  "detail": "Error message"
}
```

### 示例

#### Python 代码

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

#### JavaScript 代码

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

#### Shell 代码

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

---

## 健康检查

检查服务健康状态。

### 端点

```
GET /health
```

### 响应

**成功 (200 OK)**

```json
{
  "status": "healthy",
  "version": "1.0.0",
  "isolation": "bubblewrap",
  "platform": "Linux"
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `status` | string | 服务状态：`healthy` 或 `unhealthy` |
| `version` | string | 版本号 |
| `isolation` | string | 隔离技术：`bubblewrap` 或 `seatbelt` |
| `platform` | string | 操作系统：`Linux` 或 `Darwin` |

### 示例

```bash
curl http://localhost:8080/health
```

---

## 服务信息

获取服务详细信息和当前状态。

### 端点

```
GET /info
```

### 响应

**成功 (200 OK)**

```json
{
  "version": "1.0.0",
  "platform": "Linux",
  "isolation": "bubblewrap",
  "workspace_path": "/workspace",
  "active_executions": 0
}
```

### 字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `version` | string | 版本号 |
| `platform` | string | 操作系统 |
| `isolation` | string | 隔离技术 |
| `workspace_path` | string | 工作目录路径 |
| `active_executions` | int | 当前活跃执行数量 |

### 示例

```bash
curl http://localhost:8080/info
```

---

## 错误码

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 400 | 请求参数错误 |
| 401 | 未授权（令牌无效） |
| 404 | 资源不存在 |
| 422 | 请求体验证失败 |
| 500 | 服务器内部错误 |

### 常见错误响应

#### 请求参数错误 (400)

```json
{
  "detail": "Missing required field: execution_id"
}
```

#### 请求体验证失败 (422)

```json
{
  "detail": [
    {
      "loc": ["body", "language"],
      "msg": "value is not a valid enumeration member",
      "type": "type_error.enum"
    }
  ]
}
```

#### 服务器内部错误 (500)

```json
{
  "detail": "Internal server error: Bwrap execution failed"
}
```

---

## Python 客户端示例

```python
import httpx
import asyncio

async def execute_code(code: str, language: str = "python"):
    async with httpx.AsyncClient() as client:
        response = await client.post(
            "http://localhost:8080/execute",
            json={
                "execution_id": "exec_001",
                "session_id": "session_001",
                "code": code,
                "language": language,
                "timeout": 10
            }
        )
        return response.json()

# 使用示例
asyncio.run(execute_code("def handler(e): return {'result': 'ok'}"))
```

---

## 相关文档

- [快速开始](quick-start.md) - 安装和基本使用
- [Handler 规范](quick-start.md#handler-规范) - Lambda handler 编写指南
- [配置说明](configuration.md) - 环境变量和配置选项
