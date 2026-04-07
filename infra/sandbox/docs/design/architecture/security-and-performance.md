# 6. 安全设计 + 7. 性能优化


> **文档导航**: [返回架构总览](./overview.md)


## 6. 安全设计
### 6.1 多层隔离策略

1. **容器级隔离**
   - 每个会话独立容器
   - 禁用特权模式
   - 删除所有 Linux Capabilities
   - 非 root 用户运行

2. **进程级隔离 (Bubblewrap)**
   - Namespace 隔离（PID, NET, MNT, IPC, UTS）
   - 只读文件系统
   - 临时目录 tmpfs
   - 资源限制（ulimit）

3. **网络隔离**
   - 默认 NetworkMode=none
   - 可选白名单网络策略
   - 代理拦截敏感请求

4. **数据隔离**
   - 会话间完全隔离
   - 敏感数据环境变量传递
   - 执行结果加密存储

### 6.2 安全配置示例
```yaml
# Docker 安全配置
security_opt:
  - no-new-privileges
  - seccomp=default.json
cap_drop:
  - ALL
read_only_root_filesystem: true
user: "1000:1000"

# Bubblewrap 配置
bwrap_args:
  - --ro-bind /usr /usr
  - --ro-bind /lib /lib
  - --tmpfs /tmp
  - --proc /proc
  - --dev /dev
  - --unshare-all
  - --die-with-parent
  - --new-session

# 资源限制
resources:
  limits:
    cpu: "1"
    memory: "512Mi"
    ephemeral-storage: "1Gi"
  ulimits:
    nofile: 1024
    nproc: 128
```

## 7. 性能优化

### 7.1 启动优化

**两阶段镜像加载：**
```dockerfile
# Stage 1: 基础镜像（预热池使用）
FROM python:3.11-slim as base
RUN apt-get update && apt-get install -y bubblewrap
COPY sandbox-executor /usr/local/bin/

# Stage 2: 用户依赖（运行时加载）
FROM base
COPY requirements.txt /tmp/
RUN pip install -r /tmp/requirements.txt
```

**预热池配置：**
```python
WARM_POOL_CONFIG = {
    "default_template": {
        "target_size": 10,  # 目标池大小
        "min_size": 5,      # 最小保留
        "max_idle_time": 300,  # 最大空闲时间（秒）
    },
    "high_frequency_template": {
        "target_size": 50,
        "min_size": 20,
    }
}
```

### 7.2 并发优化

**异步处理：**
```python
# FastAPI 异步端点
@app.post("/api/v1/sessions/{session_id}/execute")
async def execute_code(session_id: str, request: ExecuteRequest):
    session = await session_manager.get_session(session_id)
    
    # 异步执行，立即返回
    execution_id = await executor.submit(session, request)
    
    return {"execution_id": execution_id, "status": "submitted"}

# 批量处理
async def batch_create_sessions(requests: List[CreateSessionRequest]):
    tasks = [session_manager.create_session(req) for req in requests]
    return await asyncio.gather(*tasks)
```

**连接池：**
```python
# HTTP 连接池
http_client = httpx.AsyncClient(
    limits=httpx.Limits(max_connections=1000, max_keepalive_connections=100),
    timeout=httpx.Timeout(10.0)
)

# MariaDB 连接池（SQLAlchemy 异步引擎）
from sqlalchemy.ext.asyncio import create_async_engine

db_engine = create_async_engine(
    "mysql+aiomysql://sandbox:password@mariadb:3306/sandbox",
    pool_size=50,              # 常驻连接池大小
    max_overflow=100,          # 最大溢出连接数
    pool_recycle=3600,         # 连接回收时间（防止连接被服务端关闭）
    pool_pre_ping=True,        # 连接前 ping 检测可用性
    pool_timeout=30,           # 获取连接超时时间
    echo=False                 # 不输出 SQL 日志
)
```

---
