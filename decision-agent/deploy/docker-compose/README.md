# Decision Agent Docker Compose 部署指南

本目录包含使用 Docker Compose 部署 Decision Agent 项目的配置文件。

## 服务说明

| 服务名 | 说明 | 端口映射 |
|--------|------|----------|
| mariadb | MariaDB 数据库 | ${MARIADB_HOST_PORT:-3306}:3306 |
| redis | Redis 缓存 | ${REDIS_HOST_PORT:-6379}:6379 |
| agent-backend | 后端服务 (agent-factory, agent-executor, agent-memory) | ${AGENT_FACTORY_HOST_PORT:-30777}:30777, ${AGENT_EXECUTOR_HOST_PORT:-30778}:30778, ${AGENT_MEMORY_HOST_PORT:-30790}:30790 |
| agent-web | 前端服务 (Nginx) | ${AGENT_WEB_HOST_PORT:-1101}:1101 |

## 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 4GB 可用内存
- 至少 10GB 可用磁盘空间

## 快速开始

### 1. 配置环境变量

```bash
cp .env.example .env
# 根据需要修改 .env 文件中的配置
```

**重要配置项说明：**

- **端口配置**：如需避免端口冲突，可修改 .env 文件中的 `*_HOST_PORT` 变量
  - `MARIADB_HOST_PORT`: MariaDB 本地端口（默认 3306）
  - `REDIS_HOST_PORT`: Redis 本地端口（默认 6379）
  - `AGENT_FACTORY_HOST_PORT`: Agent Factory 本地端口（默认 30777）
  - `AGENT_EXECUTOR_HOST_PORT`: Agent Executor 本地端口（默认 30778）
  - `AGENT_MEMORY_HOST_PORT`: Agent Memory 本地端口（默认 30790）
  - `AGENT_WEB_HOST_PORT`: Agent Web 本地端口（默认 1101）

- **数据库配置**：
  - `DB_HOST`: 数据库容器名（必须为 decision-agent-mariadb）
  - `DB_USER`: 数据库用户名（使用 MYSQL_USER 的值）
  - `DB_PASSWORD`: 数据库密码（使用 MYSQL_PASSWORD 的值）

### 2. 构建并启动所有服务

```bash
# 在项目根目录执行
cd deploy/docker-compose

# 首次部署推荐使用 make 命令
make deploy

# 或使用 docker-compose 命令
docker-compose up -d --build
```

### 3. 查看服务状态

```bash
# 使用 make 命令
make ps

# 或使用 docker-compose 命令
docker-compose ps
```

### 4. 查看日志

```bash
# 使用 make 命令
make logs            # 查看所有服务日志
make logs backend    # 查看后端日志

# 或使用 docker-compose 命令
docker-compose logs -f              # 查看所有服务日志
docker-compose logs -f agent-backend # 查看特定服务日志
docker-compose logs -f agent-web     # 查看前端日志
```

### 5. 停止服务

```bash
# 使用 make 命令
make down

# 或使用 docker-compose 命令
docker-compose down
```

### 6. 停止服务并删除数据卷

```bash
# 使用 make 命令（完全清理）
make clean-all

# 或使用 docker-compose 命令
docker-compose down -v
```

## 数据库初始化

数据库会在首次启动时自动初始化，初始化脚本位于 `init.sql`，包括：

- 创建 `adp` 数据库
- 创建 `t_data_agent_memory_history` 表

## 访问服务

| 服务 | 访问地址 |
|------|----------|
| 前端 (Agent Web) | http://localhost:${AGENT_WEB_HOST_PORT:-1101}/agent-web/my-agents.html |
| 后端 API (Agent Factory) | http://localhost:${AGENT_FACTORY_HOST_PORT:-30777} |
| Agent Executor | http://localhost:${AGENT_EXECUTOR_HOST_PORT:-30778} |
| Agent Memory | http://localhost:${AGENT_MEMORY_HOST_PORT:-30790} |

## 健康检查

所有服务都配置了健康检查：

```bash
# 检查服务健康状态
docker-compose ps

# 手动检查后端健康
curl http://localhost:${AGENT_FACTORY_HOST_PORT:-30777}/health/ready
curl http://localhost:${AGENT_EXECUTOR_HOST_PORT:-30778}/api/agent-executor/v1/health/ready

# 检查前端健康
curl http://localhost:${AGENT_WEB_HOST_PORT:-1101}/probe
```

## 故障排查

### 服务无法启动

1. **检查端口是否被占用**：
```bash
# 使用验证脚本检查配置
./validate-config.sh

# 或手动检查特定端口
lsof -i :${MARIADB_HOST_PORT:-3306}  # MariaDB
lsof -i :${REDIS_HOST_PORT:-6379}   # Redis
lsof -i :${AGENT_WEB_HOST_PORT:-1101}   # Agent Web
lsof -i :${AGENT_FACTORY_HOST_PORT:-30777}  # Agent Factory
lsof -i :${AGENT_EXECUTOR_HOST_PORT:-30778} # Agent Executor
lsof -i :${AGENT_MEMORY_HOST_PORT:-30790} # Agent Memory
```

2. 查看详细日志：
```bash
docker-compose logs [service-name]
```

### 数据库连接失败

确认数据库服务已启动并完成初始化：
```bash
docker-compose logs mariadb
```

### 前端构建失败

确认 `agent-web/dist` 目录存在或使用 Dockerfile 构建

## 使用 Makefile（推荐）

项目提供了 Makefile 来简化常用操作：

```bash
# 查看所有可用命令
make help

# 常用命令
make up        # 启动所有服务
make down      # 停止所有服务
make ps        # 查看服务状态
make logs      # 查看日志
make validate  # 验证配置
make restart   # 重启服务
make clean     # 清理容器
```

## 配置验证

项目提供了配置验证脚本，用于检查环境变量和配置完整性：

```bash
# 验证配置
./validate-config.sh
```

该脚本会检查：
- .env 文件是否存在
- 所有必需的环境变量是否已设置
- Docker Compose 配置语法是否正确
- 服务运行状态

## 环境变量参考

### 完整的环境变量列表

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| **数据库配置** | | |
| `MYSQL_ROOT_PASSWORD` | root123456 | MariaDB root 密码 |
| `MYSQL_DATABASE` | adp | 数据库名 |
| `MYSQL_USER` | aishu | 数据库用户名 |
| `MYSQL_PASSWORD` | aishu123456 | 数据库密码 |
| `MARIADB_HOST_PORT` | 3306 | MariaDB 本地端口 |
| **Redis 配置** | | |
| `REDIS_PASSWORD` | redis123456 | Redis 密码 |
| `REDIS_HOST_PORT` | 6379 | Redis 本地端口 |
| **应用配置** | | |
| `DB_HOST` | decision-agent-mariadb | 数据库主机（容器名） |
| `DB_PORT` | 3306 | 数据库端口（容器内） |
| `REDIS_HOST` | redis | Redis 主机（容器名） |
| `REDIS_PORT` | 6379 | Redis 端口（容器内） |
| `AGENT_FACTORY_HOST_PORT` | 30777 | Agent Factory 本地端口 |
| `AGENT_EXECUTOR_HOST_PORT` | 30778 | Agent Executor 本地端口 |
| `AGENT_MEMORY_HOST_PORT` | 30790 | Agent Memory 本地端口 |
| `AGENT_WEB_HOST_PORT` | 1101 | Agent Web 本地端口 |
| `BACKEND_URL` | http://agent-backend:30777 | 后端 URL（容器内通信） |
| **其他配置** | | |
| `TZ` | Asia/Shanghai | 时区 |
| `LOG_LEVEL` | info | 日志级别 |

## 开发模式

如需在开发模式下运行（挂载源代码），可以修改 `docker-compose.yaml` 添加 volume 挂载。

## 修改前端后端URL配置

### 1. 修改前端接口的后端 URL

编辑 `docker-compose.yaml`，找到 `agent-web` 服务的 `BACKEND_URL` 环境变量：

```yaml
agent-web:
  environment:
    # 前端请求后端的baseUrl地址
    BACKEND_URL: https://dip.aishu.cn
```

将值改为目标后端地址即可。nginx 会将所有 `/api/` 前缀的请求转发到该地址。

### 2. 修改后重启

> **注意**：不能使用 `docker compose restart`，它只会重启容器但不会重新读取修改后的环境变量。

必须使用 `up` 命令重新创建容器：

```bash
docker compose -f deploy/docker-compose/docker-compose.yaml up -d --no-deps agent-web
```

| 参数        | 说明                                   |
| ----------- | -------------------------------------- |
| `up -d`     | 检测配置变化并重新创建容器（后台运行） |
| `--no-deps` | 仅重启 `agent-web`，不影响其他服务     |

### 3. 验证修改是否生效

查看容器启动日志，确认 `BACKEND_URL` 的值已更新：

```bash
docker logs decision-agent-web --tail 5
```

预期输出：

```
==> BACKEND_URL: https://dip.aishu.cn
==> nginx 配置已生成，启动 nginx...
```

也可以进入容器查看生成的 nginx 配置：

```bash
docker exec decision-agent-web cat /etc/nginx/conf.d/default.conf
```

确认 `proxy_pass` 的值已变为新的地址。

### 4. 前端页面访问地址

前端采用 MPA（多页面应用）架构，每个页面对应一个独立的 HTML 入口。启动后通过 `http://localhost:1101/agent-web/` 前缀访问：

| 页面       | 地址                                                  |
| ---------- | ----------------------------------------------------- |
| 决策智能体 | `http://localhost:${AGENT_WEB_HOST_PORT:-1101}/agent-web/decision-agent.html` |
| 我的智能体 | `http://localhost:${AGENT_WEB_HOST_PORT:-1101}/agent-web/my-agents.html`      |
| 智能体模板 | `http://localhost:${AGENT_WEB_HOST_PORT:-1101}/agent-web/agent-template.html` |
| API 文档   | `http://localhost:${AGENT_WEB_HOST_PORT:-1101}/agent-web/api.html`            |

> **说明**：端口 `${AGENT_WEB_HOST_PORT:-1101}` 对应 `docker-compose.yaml` 中 `agent-web` 服务映射的宿主机端口。

## 常见问题

### Q: 如何更改端口避免冲突？

A: 修改 `.env` 文件中的 `*_HOST_PORT` 变量，例如：
```bash
# 修改 MariaDB 端口为 3307
MARIADB_HOST_PORT=3307

# 修改 Redis 端口为 6380
REDIS_HOST_PORT=6380
```

然后重启服务：
```bash
make down
make up
```

### Q: 服务启动后显示 unhealthy 怎么办？

A: 
1. 首先查看服务日志：`make logs backend`
2. 确保服务有足够的启动时间（特别是首次启动）
3. 使用 `make validate` 验证配置

### Q: 如何备份数据？

A: 使用 Makefile 提供的命令：
```bash
# 导出数据库
make export-db

# 导入数据库
make import-db file=backup.sql
```

### Q: 如何进入容器调试？

A: 
```bash
# 进入后端容器
make shell

# 或进入特定服务容器
make shell service=mariadb
```

### Q: Docker Compose 版本兼容性问题？

A: 确保使用 Docker Compose V2 语法：
```bash
# 检查版本
docker-compose version

# 如果使用 docker compose（无连字符）
docker compose version
```

本项目兼容两种语法格式。
