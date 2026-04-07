# 2.1 管理中心 (Control Plane)


> **文档导航**: [返回架构总览](../architecture/overview.md)


## 2. 关键组件设计
### 2.1 管理中心 (Control Plane)
#### 2.1.1 API Gateway
技术栈： FastAPI + Uvicorn + asyncio
职责：

- 提供统一的 RESTful API 接口
- 请求验证、鉴权、限流
- 协议转换和请求路由

核心接口：

```
# 会话管理
POST   /api/v1/sessions                 # 创建会话
GET    /api/v1/sessions/{id}            # 查询会话
DELETE /api/v1/sessions/{id}            # 终止会话

# 执行管理
POST   /api/v1/sessions/{id}/execute    # 提交执行任务
GET    /api/v1/sessions/{id}/status     # 查询执行状态
GET    /api/v1/sessions/{id}/result     # 获取执行结果

# 模板管理
POST   /api/v1/templates                # 创建模板
GET    /api/v1/templates                # 列出模板
GET    /api/v1/templates/{id}           # 获取模板详情
```
请求模式：
```
class CreateSessionRequest(BaseModel):
    template_id: str
    timeout: int = 300  # 秒
    resources: ResourceLimit
    env_vars: Dict[str, str] = {}

class ExecuteRequest(BaseModel):
    code: str
    language: Literal["python", "javascript", "shell"]
    async_mode: bool = False
    stdin: Optional[str] = None
    timeout: int = 30
```

相应模型：
```
class SessionResponse(BaseModel):
    session_id: str
    status: SessionStatus
    created_at: datetime
    runtime_type: str
    node_id: str

class ExecutionResult(BaseModel):
    execution_id: str
    status: Literal["success", "failed", "timeout"]
    stdout: str
    stderr: str
    exit_code: int
    execution_time: float
    artifacts: List[str]  # 生成的文件路径
```

#### 2.1.2 调度器 (Scheduler)

调度器负责为会话请求选择最优的容器节点。系统采用**无状态架构**，容器本身不存储任何数据，所有状态存储在外部 S3 workspace 中。

**无状态架构说明**：

- 容器完全无状态（数据在 S3 workspace）
- 容器随时可创建/销毁/重建
- 节点故障时可无缝迁移到其他节点
- 调度不依赖历史绑定，基于当前集群状态做最优决策

**调度策略**：

调度原则：
1. 优先考虑模板亲和性（镜像已缓存）
2. 使用负载均衡（新建容器）

#### 2.1.2.1 模板亲和性

优先选择已缓存镜像的节点：
- 避免镜像拉取，加快启动速度
- 启动时间：1-2s（vs 冷启动 2-5s）

#### 2.1.2.2 负载均衡

综合考虑 CPU、内存、会话数：
- 选择负载最低的节点
- 确保集群负载均衡

**调度流程实现**：

```python
class Scheduler:
    async def schedule(self, request: CreateSessionRequest) -> RuntimeNode:
        """调度逻辑（无状态架构）"""

        # 1. 获取所有健康节点
        nodes = await self.health_probe.get_healthy_nodes()

        # 2. 选择最优节点（负载 + 模板亲和性）
        best_node = await self._select_best_node(nodes, request)

        logger.info(f"Selected node {best_node.id} for session")
        return best_node

    async def _select_best_node(
        self,
        nodes: List[RuntimeNode],
        req: CreateSessionRequest
    ) -> RuntimeNode:
        """综合评分（负载 + 模板亲和性）"""
        scored_nodes = [
            (node, self._calculate_score(node, req))
            for node in nodes
        ]

        best_node = max(scored_nodes, key=lambda x: x[1])[0]

        logger.info(
            f"Selected node {best_node.id} with score "
            f"{max(scored_nodes, key=lambda x: x[1])[1]:.2f}"
        )

        return best_node

    def _calculate_score(self, node: RuntimeNode, req: CreateSessionRequest) -> float:
        """计算综合评分（负载 + 模板亲和性）"""

        # 基础负载评分 (权重 0.7)
        cpu_score = (1 - node.cpu_usage) * 0.28   # 40% of 70%
        mem_score = (1 - node.mem_usage) * 0.28   # 40% of 70%
        session_score = (1 - node.session_count / node.max_sessions) * 0.14  # 20% of 70%
        load_score = cpu_score + mem_score + session_score

        # 模板亲和性评分 (权重 0.3)
        affinity_score = 0.0

        # 模板亲和性（镜像已缓存，启动更快）
        if req.template_id in node.cached_templates:
            affinity_score += 0.3

        return load_score + affinity_score
```

**性能优化路径**：

```
最优：模板亲和节点（1-2s，镜像缓存但容器未预热）
     ↓ 无缓存
   次优：冷启动（2-5s）
```

**无状态架构优势**：

- 节点故障时可在其他节点重建容器
- 调度更灵活，无历史绑定
- 支持会话迁移
- 完全弹性扩展

#### 2.1.3 会话管理器 (Session Manager)
状态管理：

- 使用 MariaDB 存储会话状态和模板（支持事务、关系查询、数据一致性）
- 会话状态机：Creating → Running → Completed/Failed/Timeout
- 使用 SQLAlchemy ORM + asyncpg (异步 PostgreSQL/MariaDB 驱动)

数据库表设计：
```sql
-- 会话表
CREATE TABLE sessions (
    id VARCHAR(64) PRIMARY KEY,
    template_id VARCHAR(64) NOT NULL,
    status ENUM('creating', 'running', 'completed', 'failed', 'timeout', 'terminated') NOT NULL,
    runtime_type ENUM('docker', 'kubernetes') NOT NULL,
    runtime_node VARCHAR(128),           -- 当前运行的节点（可为空，支持会话迁移）
    container_id VARCHAR(128),           -- 当前容器 ID
    pod_name VARCHAR(128),               -- 当前 Pod 名称
    workspace_path VARCHAR(256),         -- S3 路径：s3://bucket/sessions/{session_id}/
    resources_cpu VARCHAR(16),
    resources_memory VARCHAR(16),
    resources_disk VARCHAR(16),
    env_vars JSON,
    timeout INT NOT NULL DEFAULT 300,
    last_activity_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,  -- 最后活动时间（用于自动清理）
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    INDEX idx_status (status),
    INDEX idx_template (template_id),
    INDEX idx_created (created_at),
    INDEX idx_runtime_node (runtime_node),  -- 支持节点故障时查询会话
    INDEX idx_last_activity (last_activity_at)  -- 支持自动清理查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 执行记录表
CREATE TABLE executions (
    id VARCHAR(64) PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    code TEXT NOT NULL,
    language VARCHAR(16) NOT NULL,
    status ENUM('pending', 'running', 'completed', 'failed', 'timeout') NOT NULL,
    stdout TEXT,
    stderr TEXT,
    exit_code INT,
    execution_time FLOAT,
    artifacts JSON,
    -- 新增字段：handler 返回值和性能指标
    return_value JSON,                  -- handler 函数返回值（JSON 可序列化）
    metrics JSON,                       -- 性能指标（duration_ms, cpu_time_ms, peak_memory_mb 等）
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    INDEX idx_session (session_id),
    INDEX idx_status (status),
    INDEX idx_created (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- 模板表
CREATE TABLE templates (
    id VARCHAR(64) PRIMARY KEY,
    name VARCHAR(128) NOT NULL UNIQUE,
    image VARCHAR(256) NOT NULL,
    base_image VARCHAR(256),
    pre_installed_packages JSON,
    default_resources_cpu VARCHAR(16),
    default_resources_memory VARCHAR(16),
    default_resources_disk VARCHAR(16),
    security_context JSON,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_name (name)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
```

生命周期管理：
```python
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.orm import sessionmaker
from sqlalchemy import select, update

# SQLAlchemy 模型
from sqlalchemy import Column, String, Enum, DateTime, Integer, Text, JSON, ForeignKey
from sqlalchemy.ext.declarative import declarative_base

Base = declarative_base()

class SessionDB(Base):
    __tablename__ = "sessions"

    id = Column(String(64), primary_key=True)
    template_id = Column(String(64), nullable=False)
    status = Column(Enum("creating", "running", "completed", "failed", "timeout", "terminated"), nullable=False)
    runtime_type = Column(Enum("docker", "kubernetes"), nullable=False)
    runtime_node = Column(String(128))
    container_id = Column(String(128))
    pod_name = Column(String(128))
    workspace_path = Column(String(256))  # S3 路径：s3://bucket/sessions/{session_id}/
    resources_cpu = Column(String(16))
    resources_memory = Column(String(16))
    resources_disk = Column(String(16))
    env_vars = Column(JSON)
    timeout = Column(Integer, default=300)
    last_activity_at = Column(DateTime, nullable=False, default=datetime.now)  # 最后活动时间
    created_at = Column(DateTime, nullable=False, default=datetime.now)
    updated_at = Column(DateTime, nullable=False, default=datetime.now, onupdate=datetime.now)
    completed_at = Column(DateTime, nullable=True)

class SessionManager:
    def __init__(self, db_url: str = "mysql+aiomysql://sandbox:password@mariadb:3306/sandbox"):
        # 创建异步数据库引擎
        self.engine = create_async_engine(
            db_url,
            pool_size=20,           # 连接池大小
            max_overflow=40,        # 最大溢出连接数
            pool_recycle=3600,      # 连接回收时间（秒）
            pool_pre_ping=True,     # 连接前检测可用性
            echo=False
        )
        self.async_session = sessionmaker(
            self.engine, class_=AsyncSession, expire_on_commit=False
        )

    async def create_session(self, request: CreateSessionRequest) -> Session:
        # 1. 生成会话 ID
        session_id = self._generate_session_id()

        # 2. 调度运行时
        runtime_node = await self.scheduler.schedule(request)

        # 3. 生成 S3 workspace 路径
        workspace_path = f"s3://{self.s3_bucket}/sessions/{session_id}/"

        # 4. 创建数据库事务
        async with self.async_session() as db:
            # 创建会话记录
            session_db = SessionDB(
                id=session_id,
                template_id=request.template_id,
                status=SessionStatus.CREATING,
                runtime_type=runtime_node.type,
                runtime_node=runtime_node.id,
                workspace_path=workspace_path,
                resources_cpu=request.resources.cpu,
                resources_memory=request.resources.memory,
                resources_disk=request.resources.disk,
                env_vars=request.env_vars,
                timeout=request.timeout
            )
            db.add(session_db)
            await db.commit()

            # 4. 调用运行时创建容器
            try:
                container_id = await runtime_node.create_container(session_id)

                # 更新容器信息
                session_db.container_id = container_id
                session_db.status = SessionStatus.RUNNING
                await db.commit()

            except Exception as e:
                # 创建失败，回滚会话状态
                session_db.status = SessionStatus.FAILED
                await db.commit()
                raise

        return self._db_to_pydantic(session_db)

    async def get_session(self, session_id: str) -> Optional[Session]:
        async with self.async_session() as db:
            result = await db.execute(
                select(SessionDB).where(SessionDB.id == session_id)
            )
            session_db = result.scalar_one_or_none()
            if session_db:
                return self._db_to_pydantic(session_db)
            return None

    async def update_session_status(self, session_id: str, status: SessionStatus):
        async with self.async_session() as db:
            await db.execute(
                update(SessionDB)
                .where(SessionDB.id == session_id)
                .values(status=status.value, updated_at=datetime.now())
            )
            await db.commit()

    async def terminate_session(self, session_id: str):
        async with self.async_session() as db:
            result = await db.execute(
                select(SessionDB).where(SessionDB.id == session_id)
            )
            session_db = result.scalar_one_or_none()

            if not session_db:
                raise ValueError(f"Session {session_id} not found")

            # 获取容器节点信息
            runtime_node = await self.scheduler.get_node(session_db.runtime_node)

            # 调用运行时清理资源
            await runtime_node.destroy_container(session_id, session_db.container_id)

            # 更新数据库状态
            session_db.status = SessionStatus.TERMINATED
            session_db.completed_at = datetime.now()
            await db.commit()

            # 回收到 Warm Pool（如果容器仍然健康）
            if await self._is_container_healthy(session_db):
                await self.warm_pool.recycle(self._db_to_pydantic(session_db))

    async def execute_code(self, session_id: str, request: ExecuteRequest) -> str:
        """执行代码，自动处理容器重建"""
        session = await self.get_session(session_id)

        # 检查容器是否存活，如果已销毁则自动重建
        if not await self._is_container_alive(session):
            logger.info(f"Container for session {session_id} not alive, recreating...")
            await self._recreate_container(session)

        # 更新最后活动时间
        await self._update_last_activity(session_id)

        # 调用运行时执行代码
        runtime_node = await self.scheduler.get_node(session.runtime_node)
        execution_id = await runtime_node.execute(session_id, request)

        return execution_id

    async def _is_container_alive(self, session: Session) -> bool:
        """检查容器是否存活"""
        try:
            runtime_node = await self.scheduler.get_node(session.runtime_node)
            return await runtime_node.is_container_alive(session.container_id)
        except Exception:
            return False

    async def _is_container_healthy(self, session_db: SessionDB) -> bool:
        """检查容器是否健康（用于回收判断）"""
        try:
            runtime_node = await self.scheduler.get_node(session_db.runtime_node)
            return await runtime_node.is_container_healthy(session_db.container_id)
        except Exception:
            return False

    async def _recreate_container(self, session: Session):
        """重建容器（共享同一个 S3 workspace）"""
        # 重新调度到最优节点
        runtime_node = await self.scheduler.schedule(
            CreateSessionRequest(
                template_id=session.template_id,
                resources=session.resources,
                env_vars=session.env_vars,
                timeout=session.timeout
            )
        )

        # 创建新容器，挂载同一个 S3 workspace
        container_id = await runtime_node.create_container(
            session_id=session.id,
            workspace_path=session.workspace_path
        )

        # 更新数据库中的容器信息
        async with self.async_session() as db:
            await db.execute(
                update(SessionDB)
                .where(SessionDB.id == session.id)
                .values(
                    runtime_node=runtime_node.id,
                    container_id=container_id,
                    status=SessionStatus.RUNNING,
                    updated_at=datetime.now()
                )
            )
            await db.commit()

    async def _update_last_activity(self, session_id: str):
        """更新最后活动时间（用于自动清理）"""
        async with self.async_session() as db:
            await db.execute(
                update(SessionDB)
                .where(SessionDB.id == session_id)
                .values(last_activity_at=datetime.now())
            )
            await db.commit()

    def _db_to_pydantic(self, session_db: SessionDB) -> Session:
        """将 SQLAlchemy 模型转换为 Pydantic 模型"""
        return Session(
            id=session_db.id,
            template_id=session_db.template_id,
            status=SessionStatus(session_db.status),
            runtime_type=session_db.runtime_type,
            runtime_node=session_db.runtime_node,
            container_id=session_db.container_id,
            pod_name=session_db.pod_name,
            workspace_path=session_db.workspace_path,
            resources=ResourceLimit(
                cpu=session_db.resources_cpu,
                memory=session_db.resources_memory,
                disk=session_db.resources_disk
            ),
            env_vars=session_db.env_vars or {},
            created_at=session_db.created_at,
            updated_at=session_db.updated_at,
            timeout=session_db.timeout
        )

    async def cleanup_idle_sessions(self):
        """定期清理空闲会话（后台任务）

        清理策略：
        - 空闲超过 30 分钟自动销毁容器
        - 创建超过 6 小时强制销毁
        """
        async with self.async_session() as db:
            # 空闲超时清理
            idle_threshold = datetime.now() - timedelta(minutes=30)
            idle_sessions = await db.execute(
                select(SessionDB)
                .where(SessionDB.status == SessionStatus.RUNNING)
                .where(SessionDB.last_activity_at < idle_threshold)
            )
            idle_sessions = idle_sessions.scalars().all()

            for session_db in idle_sessions:
                logger.info(f"Cleaning up idle session {session_db.id}")
                await self._destroy_session_container(session_db.id, session_db)

            # 最大生命周期强制清理
            max_lifetime_threshold = datetime.now() - timedelta(hours=6)
            old_sessions = await db.execute(
                select(SessionDB)
                .where(SessionDB.status == SessionStatus.RUNNING)
                .where(SessionDB.created_at < max_lifetime_threshold)
            )
            old_sessions = old_sessions.scalars().all()

            for session_db in old_sessions:
                logger.info(f"Cleaning up old session {session_db.id}")
                await self._destroy_session_container(session_db.id, session_db)

    async def _destroy_session_container(self, session_id: str, session_db: SessionDB):
        """销毁会话容器"""
        try:
            runtime_node = await self.scheduler.get_node(session_db.runtime_node)
            await runtime_node.destroy_container(session_id, session_db.container_id)
        except Exception as e:
            logger.warning(f"Failed to destroy container for session {session_id}: {e}")

        # 更新数据库状态
        async with self.async_session() as db:
            await db.execute(
                update(SessionDB)
                .where(SessionDB.id == session_id)
                .values(
                    status=SessionStatus.TERMINATED,
                    completed_at=datetime.now(),
                    container_id=None,  # 清空容器 ID
                    runtime_node=None
                )
            )
            await db.commit()
```

#### 2.1.4 监控探针 (Health Probe)
探测机制：

- 心跳检测：每 10 秒向运行时发送 /health 请求
- 负载采集：每 30 秒收集 CPU、内存、会话数
- 异常检测：连续 3 次心跳失败则标记为不健康

自动摘除：

```python
class HealthProbe:
    async def probe_loop(self):
        while True:
            for node in self.runtime_nodes:
                try:
                    # 发送心跳
                    response = await asyncio.wait_for(
                        self.http_client.get(f"{node.url}/health"),
                        timeout=5.0
                    )
                    
                    # 更新负载信息
                    node.update_metrics(response.json())
                    node.mark_healthy()
                    
                except asyncio.TimeoutError:
                    node.increment_failure_count()
                    
                    # 连续失败则摘除
                    if node.failure_count >= 3:
                        await self.remove_unhealthy_node(node)
            
            await asyncio.sleep(10)
```

#### 2.1.5 状态同步服务 (State Sync Service)

**设计原则**：Docker/K8s 是容器状态的唯一真实来源，Session 表只保存关联关系。

状态同步服务负责：
1. **启动时全量同步**：Control Plane 重启后恢复状态
2. **定时健康检查**：定期检查容器状态并修复不一致
3. **容器状态恢复**：结合预热池自动恢复不健康的容器

```python
class StateSyncService:
    """
    状态同步服务

    职责：
    1. 启动时全量状态同步
    2. 定时健康检查（每 30 秒）
    3. 容器状态恢复
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        docker_scheduler: IDockerScheduler,
        warm_pool_manager: WarmPoolManager,
    ):
        self._session_repo = session_repo
        self._docker_scheduler = docker_scheduler
        self._warm_pool_manager = warm_pool_manager

    async def sync_on_startup(self) -> dict:
        """
        启动时全量同步

        策略：
        1. 查询所有 RUNNING/CREATING 状态的 Session
        2. 通过 Docker API 检查每个容器是否真实存在且运行中
        3. 更新 Session 状态：
           - 容器存在且运行 → 保持 RUNNING
           - 容器不存在/已停止 → 尝试恢复或标记为 FAILED
        """
        active_sessions = await self._session_repo.find_by_status("running")
        active_sessions.extend(await self._session_repo.find_by_status("creating"))

        stats = {"healthy": 0, "unhealthy": 0, "recovered": 0, "failed": 0}

        for session in active_sessions:
            if not session.container_id:
                continue

            # 直接通过 Docker API 检查容器状态
            is_running = await self._docker_scheduler.is_container_running(
                session.container_id
            )

            if is_running:
                stats["healthy"] += 1
            else:
                stats["unhealthy"] += 1
                # 尝试恢复
                recovered = await self._attempt_recovery(session)
                if recovered:
                    stats["recovered"] += 1
                else:
                    stats["failed"] += 1

        return stats

    async def periodic_health_check(self) -> dict:
        """
        定时健康检查（每 30 秒）

        只检查 RUNNING 状态的 Session，减少查询范围
        """
        running_sessions = await self._session_repo.find_by_status("running")

        for session in running_sessions:
            if not session.container_id:
                continue

            is_running = await self._docker_scheduler.is_container_running(
                session.container_id
            )

            if not is_running:
                await self._attempt_recovery(session)

        return {"checked": len(running_sessions)}

    async def _attempt_recovery(self, session: Session) -> bool:
        """
        尝试恢复 Session

        策略：
        1. 首先尝试从预热池获取实例
        2. 如果预热池为空，创建新容器
        3. 如果创建失败，标记 Session 为 FAILED
        """
        # 1. 尝试从预热池获取
        warm_entry = await self._warm_pool_manager.acquire(
            session.template_id, session.id
        )
        if warm_entry:
            # 分配预热实例
            session.container_id = warm_entry.container_id
            session.runtime_node = warm_entry.node_id
            await self._session_repo.save(session)
            return True

        # 2. 创建新容器
        try:
            container_id = await self._docker_scheduler.create_container_for_session(
                session_id=session.id,
                template_id=session.template_id,
                workspace_path=session.workspace_path,
            )
            session.container_id = container_id
            await self._session_repo.save(session)
            return True
        except Exception as e:
            logger.error(f"Failed to recover session {session.id}: {e}")
            # 3. 标记为失败
            session.mark_as_failed()
            await self._session_repo.save(session)
            return False
```

**启动流程集成**：

```python
# 在 main.py 的 lifespan 函数中
@asynccontextmanager
async def lifespan(app: FastAPI):
    # 启动时
    logger.info("Starting Sandbox Control Plane")

    # 初始化依赖注入
    initialize_dependencies(app)

    # 执行启动时状态同步
    state_sync_service = app.state.state_sync_service
    sync_stats = await state_sync_service.sync_on_startup()
    logger.info(f"Startup sync completed: {sync_stats}")

    # 启动后台任务管理器
    task_manager = BackgroundTaskManager()

    # 注册定时健康检查任务（每 30 秒）
    task_manager.register_task(
        name="health_check",
        func=state_sync_service.periodic_health_check,
        interval_seconds=30,
        initial_delay_seconds=60,
    )

    await task_manager.start_all()

    yield

    # 关闭时
    await task_manager.stop_all()
```
