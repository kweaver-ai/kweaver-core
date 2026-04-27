# Sandbox Executor

> 安全的代码执行守护进程，使用 Bubblewrap 和 macOS Seatbelt 提供进程级隔离

[![Python Version](https://img.shields.io/badge/python-3.11+-blue.svg)](https://www.python.org/downloads/)
[![License](https://img.shields.io/badge/license-MIT-green.svg)](LICENSE)

## 概述

Sandbox Executor 是一个高性能的代码执行服务，专为 AI Agent 应用场景设计。它提供了多层安全隔离机制，确保不受信任的代码在受控环境中安全执行。

## 核心特性

- **多层安全隔离** - Docker 容器 + Bubblewrap/sandbox-exec 双层隔离
- **异步高性能** - 基于 asyncio 的真正异步执行，支持高并发
- **Lambda 兼容** - 支持 AWS Lambda handler 规范
- **实时可观测** - 心跳上报、生命周期管理、执行指标

## 快速开始

```bash
# 安装依赖
pip install -r requirements.txt

# 启动服务
python -m executor.interfaces.http.rest

# 验证服务
curl http://localhost:8080/health
```

**详细指南**: [快速开始文档](docs/quick-start.md)

## 技术栈

| 组件 | 技术 |
|------|------|
| HTTP 框架 | FastAPI + Uvicorn |
| 隔离技术 | Bubblewrap (Linux) / sandbox-exec (macOS) |
| 异步运行时 | asyncio |
| 日志 | structlog |
| 数据验证 | Pydantic |

## 文档

| 文档 | 说明 |
|------|------|
| [快速开始](docs/quick-start.md) | 安装、配置和基本使用 |
| [架构设计](docs/architecture.md) | 六边形架构、模块结构、设计原理 |
| [API 文档](docs/api-reference.md) | RESTful API 端点和示例 |
| [配置说明](docs/configuration.md) | 环境变量和隔离配置 |
| [开发指南](docs/development.md) | 开发环境设置、测试、代码规范 |
| [部署指南](docs/deployment.md) | Docker、Docker Compose、Kubernetes 部署 |
| [故障排查](docs/troubleshooting.md) | 常见问题和解决方案 |

## 示例

### Python Handler

```python
def handler(event):
    name = event.get('name', 'World')
    return {'message': f'Hello, {name}!'}
```

### 执行代码

```bash
curl -X POST http://localhost:8080/execute \
  -H 'Content-Type: application/json' \
  -d '{
    "execution_id": "test_001",
    "session_id": "session_001",
    "code": "def handler(event): return {\"message\": \"Hello!\"}",
    "language": "python",
    "timeout": 10
  }'
```

## 许可证

MIT License - 详见 [LICENSE](../../LICENSE)
