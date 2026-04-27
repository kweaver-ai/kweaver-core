# Runtime Node 注册方案

## 问题描述

创建会话时返回成功，但数据库中 `container_id` 为 `null`，意味着运行时/执行器容器没有被动态创建。

## 根本原因分析

### 问题链路

1. **`USE_DOCKER_SCHEDULER = False`** ✅ (已修复为 `True`)

2. **`runtime_nodes` 表为空** ⚠️ **核心问题**
   - `DockerSchedulerService.schedule()` 查询 `runtime_nodes` 获取健康节点
   - `get_healthy_nodes()` 返回空列表 → 没有可用节点进行调度
   - 结果：`RuntimeError: No healthy runtime nodes available`

3. **缺失的内部 API 端点** ⚠️ **架构缺口**
   - Executor 尝试发送 `container_ready`、`heartbeat`、`container_exited` 信号
   - Control Plane 的 `/internal` API 只有 `/internal/executions/{id}/result`
   - **缺失的端点**：
     - `POST /internal/runtime/nodes/register` - 运行时节点注册
     - `POST /internal/runtime/containers/ready` - 容器就绪信号
     - `POST /internal/runtime/heartbeat` - 心跳信号

4. **模板镜像可能不正确**
   - 数据库模板应该使用 `sandbox-executor:v1.0` 而不是 `python:3.11-slim`

5. **网络配置**
   - Docker 调度器创建容器时使用 `NetworkMode: "none"`
   - 需要改为 `sandbox_network` 以支持 executor 通信

---

## Runtime Node 注册 - 两种方案

### 方案 1: 静态注册（当前实现）✅

**优点**：简单、可预测、无额外复杂度
**缺点**：手动设置、不支持自动扩展

```sql
-- 手动插入 runtime node
INSERT INTO runtime_nodes (...) VALUES (...);
```

### 方案 2: 动态注册（未来扩展）📋

**优点**：自动发现、可扩展、生产就绪
**缺点**：需要实现缺失的内部 API 端点

```
Executor 启动
    ↓
POST /internal/runtime/nodes/register (注册节点)
POST /internal/runtime/containers/ready (容器就绪)
    ↓ (每 5 秒)
POST /internal/runtime/heartbeat (保活)
```

---

## 当前实现：静态注册

### 步骤 1: 注册 Runtime Node

**执行 SQL** 将本地 Docker daemon 注册为运行时节点：

```sql
INSERT INTO runtime_nodes (
    node_id,
    hostname,
    runtime_type,
    ip_address,
    api_endpoint,
    status,
    total_cpu_cores,
    total_memory_mb,
    running_containers,
    max_containers,
    cached_images,
    last_heartbeat_at
) VALUES (
    'docker-local',
    'localhost',
    'docker',
    '127.0.0.1',
    'unix:///.docker/run/docker.sock',
    'online',
    8.0,
    16384,
    0,
    100,
    '["sandbox-executor:v1.0"]',
    NOW()
);
```

### 步骤 2: 验证模板配置

**检查并更新** `python-basic` 模板：

```sql
-- 检查当前模板
SELECT id, name, image_url FROM templates WHERE id = 'python-basic';

-- 如果需要，更新镜像
UPDATE templates
SET image_url = 'sandbox-executor:v1.0'
WHERE id = 'python-basic';
```

### 步骤 3: Docker 网络配置

#### 3.1 创建 Docker 网络

```bash
docker network create sandbox_network
```

#### 3.2 更新 `DockerScheduler`

**文件**: `src/infrastructure/container_scheduler/docker_scheduler.py`

添加 `network_name` 参数支持：

```python
async def create_container(
    self,
    config: ContainerConfig,
    network_name: str = "sandbox_network"  # 添加此参数
) -> str:
    """创建 Docker 容器"""
    docker = await self._ensure_docker()

    # ... 资源限制解析 ...

    container_config = {
        "Image": config.image,
        "Hostname": config.name,
        "Env": [f"{k}={v}" for k, v in config.env_vars.items()],
        "HostConfig": {
            "NetworkMode": network_name,  # 使用指定网络而非 "none"
            # ... 其他配置 ...
        },
        # ...
    }
```

#### 3.3 更新 `DockerSchedulerService`

**文件**: `src/infrastructure/schedulers/docker_scheduler_service.py`

传递网络名称：

```python
async def create_container_for_session(
    self,
    session_id: str,
    template_id: str,
    image: str,
    resource_limit,
    env_vars: dict,
    workspace_path: str,
    network_name: str = "sandbox_network",  # 添加此参数
) -> str:
    """为会话创建容器"""

    # ...

    config = ContainerConfig(
        image=image,
        name=f"sandbox-{session_id}",
        env_vars={...},
        # ...
    )

    # 创建容器时传递 network_name
    container_id = await self._container_scheduler.create_container(
        config,
        network_name=network_name
    )
```

### 步骤 4: 验证

#### 4.1 启动 Control Plane

```bash
cd sandbox_control_plane
uvicorn src.interfaces.rest.main:app --reload
```

#### 4.2 创建会话

```bash
curl -X POST http://localhost:8000/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"template_id": "python-basic"}'
```

#### 4.3 验证容器已创建

```bash
docker ps | grep sandbox
```

#### 4.4 检查数据库

```bash
# container_id 不应该为 null
SELECT id, container_id, status FROM sessions ORDER BY created_at DESC LIMIT 1;
```

---

## 未来扩展：动态注册

### 需要实现的 API 端点

**文件**: `src/interfaces/rest/api/v1/internal.py`

#### 1. 节点注册端点

```python
@router.post("/runtime/nodes/register")
async def register_runtime_node(
    node_info: RuntimeNodeRegistration,
    runtime_node_repo: IRuntimeNodeRepository = Depends(...),
):
    """
    Executor 启动时调用此端点注册运行时节点

    请求体：
    {
        "node_id": "docker-local",
        "hostname": "localhost",
        "runtime_type": "docker",
        "api_endpoint": "unix:///var/run/docker.sock",
        "total_cpu_cores": 8.0,
        "total_memory_mb": 16384,
        "max_containers": 100
    }
    """
    # 创建或更新 runtime_nodes 记录
    pass
```

#### 2. 容器就绪端点

```python
@router.post("/runtime/containers/ready")
async def container_ready(
    container_info: ContainerReadyInfo,
):
    """
    Executor 调用此端点报告容器已就绪

    请求体：
    {
        "container_id": "abc123",
        "session_id": "sess_xxx",
        "executor_url": "http://container-name:8080"
    }
    """
    # 更新容器状态为 ready
    pass
```

#### 3. 心跳端点

```python
@router.post("/runtime/heartbeat")
async def heartbeat(
    heartbeat: HeartbeatSignal,
):
    """
    Executor 每 5 秒调用此端点发送心跳

    请求体：
    {
        "node_id": "docker-local",
        "container_id": "abc123",
        "timestamp": "2024-01-15T10:30:00Z"
    }
    """
    # 更新 last_heartbeat_at，保持节点健康状态
    pass
```

### Executor 端的回调实现

**文件**: `runtime/executor/infrastructure/callbacks/http_callback_client.py`

确保 Executor 在启动时调用这些端点：

```python
class HttpCallbackClient:
    async def on_container_ready(self):
        """容器启动时调用"""
        await self._post("/internal/runtime/nodes/register", {...})
        await self._post("/internal/runtime/containers/ready", {...})

    async def start_heartbeat(self):
        """开始心跳循环"""
        while self._running:
            await self._post("/internal/runtime/heartbeat", {...})
            await asyncio.sleep(5)
```

---

## 关键文件清单

| 文件 | 修改内容 |
|------|---------|
| `runtime_nodes` 表 | **SQL INSERT** 注册 docker-local 节点 |
| `templates` 表 | **SQL UPDATE** 如果 image_url 错误 |
| `src/infrastructure/container_scheduler/docker_scheduler.py` | 添加 `network_name` 参数支持 |
| `src/infrastructure/schedulers/docker_scheduler_service.py` | 传递 `network_name="sandbox_network"` |
| `src/interfaces/rest/api/v1/internal.py` | 📋 未来：添加节点注册和心跳端点 |

---

## 架构图

```
SessionService.create_session()
    ↓
DockerSchedulerService.schedule()
    ↓ 查询 runtime_nodes 表
RuntimeNode (docker-local) ✅ (来自静态 SQL INSERT)
    ↓
DockerScheduler.create_container(network="sandbox_network")
    ↓
使用 sandbox-executor:v1.0 创建容器
    ↓
session.container_id = "abc123" ✅
```

---

## 状态

- ✅ 静态注册实现完成
- 📋 动态注册待未来扩展
