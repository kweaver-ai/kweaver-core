# Sandbox API 文档

本目录按协议类型组织 API 文档，当前以 REST/OpenAPI 为主。

## 文档列表

### [sandbox-openapi.json](./rest/sandbox-openapi.json)
当前 REST OpenAPI 规范。

### [execute-sync-openapi.yaml](./rest/execute-sync-openapi.yaml)
同步执行接口的独立 OpenAPI 描述。

## 协议目录

- `rest/`: REST/OpenAPI 文档
- `grpc/`: 预留给 gRPC 协议文档
- `websocket/`: 预留给 WebSocket 协议文档

## REST API 概览

**主要功能**:
- 健康检查：服务状态和依赖项健康检查
- 会话管理：创建、查询、终止会话（支持软终止和硬删除）
- 代码执行：提交异步/同步执行任务并获取结果
- 文件操作：列出、上传、下载会话工作区文件
- 模板管理：管理沙箱环境模板
- 内部回调：Executor 回调接口（结果上报、心跳、容器生命周期）

**调用方**:
- AI Agent 应用
- Data Agent 系统
- 上层业务服务
- sandbox-executor (内部回调)

**基础路径**:
- 公开 API: `/api/v1`
- 内部 API: `/api/v1/internal`

## 架构说明

### 容器调度架构

沙箱平台采用 **Control Plane + Container Scheduler** 架构：

```
┌─────────────────────────────────────────────────────────────┐
│                     Control Plane (FastAPI)                  │
│  ┌────────────────────────────────────────────────────────┐ │
│  │           Container Scheduler (内部模块)                │ │
│  │   ├─ K8s Scheduler (kubernetes python client)          │ │
│  │   └─ Docker Scheduler (aiodocker SDK)                  │ │
│  └────────────────────────────────────────────────────────┘ │
│                          ↓                                  │
│              Container Runtime (K8s/Docker)                 │
│                          ↓                                  │
│              Executor (sandbox-executor 守护进程)           │
│                          ↓ HTTP 回调                        │
│                   Control Plane                             │
└─────────────────────────────────────────────────────────────┘
```

### 关键组件

- **Control Plane**: FastAPI 管理服务，处理 API 请求、调度、会话管理
- **Container Scheduler**: 容器调度器，直接调用 Docker/K8s API 管理容器生命周期
- **Executor**: 容器内 HTTP 守护进程，接收执行请求并使用 Bubblewrap 隔离用户代码

## 接口分类

### 按功能分类

#### 健康检查
- `GET /api/v1/health` - 基础健康检查
- `GET /api/v1/health/detailed` - 详细健康检查（含依赖项状态）
- `POST /api/v1/health/sync` - 手动触发状态同步

#### 会话管理
- `POST /api/v1/sessions` - 创建会话（支持依赖安装）
- `GET /api/v1/sessions` - 列出会话（支持 status/template_id 筛选、分页）
- `GET /api/v1/sessions/{session_id}` - 获取会话详情
- `POST /api/v1/sessions/{session_id}/terminate` - 软终止会话（保留记录）
- `DELETE /api/v1/sessions/{session_id}` - 硬删除会话（级联删除执行记录）

#### 代码执行
- `POST /api/v1/executions/sessions/{session_id}/execute` - 提交异步执行任务
- `POST /api/v1/executions/sessions/{session_id}/execute-sync` - 同步执行代码（轮询等待结果）
- `GET /api/v1/executions/{execution_id}/status` - 获取执行状态
- `GET /api/v1/executions/{execution_id}/result` - 获取执行结果
- `GET /api/v1/executions/sessions/{session_id}/executions` - 列出会话的所有执行

#### 文件操作
- `GET /api/v1/sessions/{session_id}/files` - 列出工作区文件（支持指定目录路径）
- `POST /api/v1/sessions/{session_id}/files/upload` - 上传文件到工作区
- `GET /api/v1/sessions/{session_id}/files/{file_path}` - 下载工作区文件

#### 模板管理
- `POST /api/v1/templates` - 创建模板
- `GET /api/v1/templates` - 列出所有模板（支持分页）
- `GET /api/v1/templates/{template_id}` - 获取模板详情
- `PUT /api/v1/templates/{template_id}` - 更新模板
- `DELETE /api/v1/templates/{template_id}` - 删除模板

#### 内部回调 API (由 Executor 调用)
- `POST /api/v1/internal/containers/ready` - 容器就绪通知
- `POST /api/v1/internal/containers/exited` - 容器退出通知
- `POST /api/v1/internal/executions/{execution_id}/heartbeat` - 执行心跳上报
- `POST /api/v1/internal/executions/{execution_id}/result` - 执行结果上报

## 公共规范

### 认证方式

#### 公开 API
```
Authorization: Bearer ACCESS_TOKEN
```

#### 内部 API
```
Authorization: Bearer INTERNAL_API_TOKEN
```

**安全要求**:
- 内部 API 仅限内网访问
- 建议配置 NetworkPolicy 限制访问来源

### 公共请求头

| Header | 类型 | 必填 | 说明 |
|--------|------|------|------|
| Authorization | string | 是 | Bearer Token 认证 |
| Content-Type | string | 是 | application/json 或 multipart/form-data |
| X-Request-ID | string | 否 | 请求追踪 ID |

### HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 201 | 创建成功 |
| 204 | 删除成功（无返回内容） |
| 400 | 请求参数错误 |
| 401 | 认证失败 |
| 404 | 资源不存在 |
| 422 | 请求参数验证失败 |
| 500 | 服务器内部错误 |

### 错误响应格式

```json
{
  "detail": [
    {
      "loc": ["body", "timeout"],
      "msg": "ensure this value is greater than 0",
      "type": "greater_than"
    }
  ]
}
```

## 使用示例

### 查看文档

1. **Swagger UI**: http://localhost:8000/docs (本地运行时)
2. **在线工具**: https://editor.nextapis.com/
3. **Redoc**: https://redocly.github.io/redoc/

### 使用 curl 示例

#### 创建会话
```bash
curl -X POST http://localhost:8000/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "python3.11-datascience",
    "timeout": 300,
    "cpu": "1",
    "memory": "512Mi",
    "dependencies": [
      {"name": "pandas", "version": "==2.0.0"},
      {"name": "numpy"}
    ]
  }'
```

#### 同步执行代码
```bash
curl -X POST "http://localhost:8000/api/v1/executions/sessions/{session_id}/execute-sync?sync_timeout=60" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "def handler(event):\n    name = event.get(\"name\", \"World\")\n    return {\"message\": f\"Hello, {name}!\"}",
    "language": "python",
    "timeout": 30,
    "event": {"name": "Alice"}
  }'
```

#### 列出工作区文件
```bash
curl -X GET "http://localhost:8000/api/v1/sessions/{session_id}/files?path=src/&limit=100"
```

#### 上传文件
```bash
curl -X POST "http://localhost:8000/api/v1/sessions/{session_id}/files/upload?path=src/main.py" \
  -H "Content-Type: multipart/form-data" \
  -F "file=@main.py"
```

#### 软终止会话
```bash
curl -X POST http://localhost:8000/api/v1/sessions/{session_id}/terminate
```

## 相关文档

- [文档中心](../README.md) - 文档导航入口
- [架构总览](../design/architecture/overview.md) - 完整的系统设计文档
- [产品路线图](../product/roadmap.md) - 产品需求与迭代方向

## 版本历史

| 版本 | 日期 | 说明 |
|------|------|------|
| 2.1.0 | 2025-02-10 | 添加同步执行端点、依赖安装支持、状态同步服务 |
| 2.0.0 | 2025-01-15 | 重构为 Control Plane + Container Scheduler 架构 |
| 1.0.0 | 2025-01-05 | 初始版本 |
