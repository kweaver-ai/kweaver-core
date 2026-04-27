# 5. Python 依赖配置


> **文档导航**: [返回架构总览](../architecture/overview.md)


    self,
    s3_bucket: str,
    s3_prefix: str,
    s3_endpoint_url: str,
    s3_access_key: str,
    s3_secret_key: str,
) -> str:
```

**修改后签名**：

```python
def _build_s3_mount_entrypoint(
    self,
    s3_bucket: str,
    s3_prefix: str,
    s3_endpoint_url: str,
    s3_access_key: str,
    s3_secret_key: str,
    dependencies: Optional[List[str]] = None,  # ✅ 新增参数
) -> str:
```

**实现变更**：

```python
def _build_s3_mount_entrypoint(
    self,
    s3_bucket: str,
    s3_prefix: str,
    s3_endpoint_url: str,
    s3_access_key: str,
    s3_secret_key: str,
    dependencies: Optional[List[str]] = None,  # ✅ 新增参数
) -> str:
    """
    构建容器启动脚本，用于挂载 S3 bucket 并安装依赖

    Args:
        s3_bucket: S3 bucket 名称
        s3_prefix: S3 路径前缀
        s3_endpoint_url: S3 端点 URL
        s3_access_key: S3 访问密钥 ID
        s3_secret_key: S3 访问密钥
        dependencies: pip 包规范列表（如 ["requests==2.31.0", "pandas>=2.0"]）✅ 新增

    Returns:
        Shell 脚本字符串
    """
    # 对于 MinIO，需要使用 use_path_request_style
    path_style_option = "-o use_path_request_style" if s3_endpoint_url else ""

    # ✅ 新增：依赖安装脚本片段
    dependency_install_script = ""
    if dependencies:
        deps_json = json.dumps(dependencies)
        dependency_install_script = f"""
# ========== 安装 Python 依赖 ==========
echo "📦 Installing dependencies: {deps_json}"

# 将依赖安装到容器本地文件系统（S3 挂载点不适合 pip 安装）
VENV_DIR="/opt/sandbox-venv"
mkdir -p $VENV_DIR
mkdir -p /tmp/pip-cache

if pip3 install \\
    --target $VENV_DIR \\
    --cache-dir /tmp/pip-cache \\
    --no-cache-dir \\
    --no-warn-script-location \\
    --disable-pip-version-check \\
    --index-url https://pypi.org/simple/ \\
    {" ".join(dependencies)}; then
    echo "✅ Dependencies installed successfully"
    # 修改属主为 sandbox 用户（gosu 切换前以 root 安装）
    chown -R sandbox:sandbox $VENV_DIR
    rm -rf /tmp/pip-cache
else
    echo "❌ Failed to install dependencies"
    exit 1
fi
"""

    return f"""#!/bin/sh
set -e

# 创建 s3fs 凭证文件
echo "{s3_access_key}:{s3_secret_key}" > /tmp/.passwd-s3fs
chmod 600 /tmp/.passwd-s3fs

# 1. 创建 S3 挂载点
echo "Mounting S3 bucket {s3_bucket}..."
mkdir -p /mnt/s3-root
s3fs {s3_bucket} /mnt/s3-root \\
    -o passwd_file=/tmp/.passwd-s3fs \\
    -o url={s3_endpoint_url or "https://s3.amazonaws.com"} \\
    {path_style_option} \\
    -o allow_other \\
    -o umask=000

# 2. 创建 session workspace 目录
SESSION_PATH="/mnt/s3-root/{s3_prefix}"
echo "Ensuring session workspace exists: $SESSION_PATH"
mkdir -p "$SESSION_PATH"

# 3. 将 /workspace 移动到临时位置
mv /workspace /workspace-old 2>/dev/null || true

# 4. 创建符号链接从 /workspace 到 session 目录
ln -s "$SESSION_PATH" /workspace

# 5. 验证符号链接
echo "Workspace symlink: $(ls -la /workspace)"

# ========== ✅ 新增：安装依赖 ==========
{dependency_install_script}

# 6. 使用 gosu 切换到 sandbox 用户运行 executor（保持不变）
echo "Starting sandbox executor as sandbox user..."
exec gosu sandbox python -m executor.interfaces.http.rest
"""
```

#### 5.4.2 修改 `create_container()` 方法

```python
async def create_container(self, config: ContainerConfig) -> str:
    """创建 Docker 容器（扩展现有逻辑）"""
    docker = await self._ensure_docker()

    # 解析资源限制（现有逻辑）
    cpu_quota = int(float(config.cpu_limit) * 100000)
    memory_bytes = self._parse_memory_to_bytes(config.memory_limit)

    # 检查是否需要 S3 workspace 挂载（现有逻辑）
    s3_workspace = self._parse_s3_workspace(config.workspace_path)
    use_s3_mount = s3_workspace is not None

    # 基础环境变量（现有逻辑）
    env_vars = dict(config.env_vars)

    # ✅ 新增：从 config.labels 中提取依赖列表
    dependencies_json = config.labels.get("dependencies", "")
    dependencies = json.loads(dependencies_json) if dependencies_json else None

    # 基础容器配置（现有逻辑）
    container_config = {
        "Image": config.image,
        "Hostname": config.name,
        "Env": [f"{k}={v}" for k, v in env_vars.items()],
        "HostConfig": {
            "NetworkMode": config.network_name,  # sandbox_network (bridge)
            "CpuQuota": cpu_quota,
            "CpuPeriod": 100000,
            "Memory": memory_bytes,
            "MemorySwap": memory_bytes,
        },
        "Labels": config.labels,
        "ExposedPorts": {"8080/tcp": {}},
    }

    # S3 workspace 挂载逻辑（现有逻辑，保持不变）
    if use_s3_mount:
        settings = get_settings()

        container_config["User"] = "root"  # S3 mount 需要 root
        container_config["HostConfig"]["CapAdd"] = ["SYS_ADMIN"]
        container_config["HostConfig"]["Devices"] = [
            {"PathOnHost": "/dev/fuse", "PathInContainer": "/dev/fuse", "CgroupPermissions": "rwm"}
        ]
        container_config["HostConfig"]["Tmpfs"] = {"/tmp": "size=100M,mode=1777"}

        # S3 环境变量（现有逻辑）
        s3_env_vars = {
            "S3_BUCKET": s3_workspace["bucket"],
            "S3_PREFIX": s3_workspace["prefix"],
            "S3_ENDPOINT_URL": settings.s3_endpoint_url or "https://s3.amazonaws.com",
            "S3_REGION": settings.s3_region,
            "WORKSPACE_MOUNT_POINT": "/workspace",
            "WORKSPACE_PATH": "/workspace",
        }
        for k, v in s3_env_vars.items():
            container_config["Env"].append(f"{k}={v}")

        # ✅ 修改：传递依赖列表到 entrypoint 脚本
        entrypoint_script = self._build_s3_mount_entrypoint(
            s3_bucket=s3_workspace["bucket"],
            s3_prefix=s3_workspace["prefix"],
            s3_endpoint_url=settings.s3_endpoint_url or "",
            s3_access_key=settings.s3_access_key_id,
            s3_secret_key=settings.s3_secret_access_key,
            dependencies=dependencies,  # ✅ 新增参数
        )
        container_config["Entrypoint"] = ["/bin/sh", "-c"]
        container_config["Cmd"] = [entrypoint_script]

        logger.info(
            f"Configuring S3 workspace mount for {config.name}: "
            f"bucket={s3_workspace['bucket']}, prefix={s3_workspace['prefix']}, "
            f"dependencies={dependencies}"
        )

    try:
        container = await docker.containers.create(container_config, name=config.name)
        logger.info(
            f"Created container {container.id} for session {config.name} "
            f"on network {config.network_name} (S3 mount: {use_s3_mount}, "
            f"dependencies: {len(dependencies) if dependencies else 0})"
        )
        return container.id
    except DockerError as e:
        logger.error(f"Failed to create container: {e}")
        raise
```

### 5.5 Session Service 实现

**文件**: `sandbox_control_plane/src/application/services/session_service.py`

#### 5.5.1 版本冲突检测

```python
def _detect_version_conflicts(
    self,
    template_packages: List[str],
    requested_specs: List[str]
) -> List[str]:
    """
    检测 Template 预装包与用户请求包的版本冲突

    Args:
        template_packages: Template 预装包列表（如: ["requests==2.28.0", "numpy>=1.20"]）
        requested_specs: 用户请求的包规范列表

    Returns:
        冲突描述列表
    """
    import re
    from packaging import requirements

    conflicts = []
    template_map = {}

    # 解析 Template 包
    for pkg_str in template_packages:
        try:
            req = requirements.Requirement(pkg_str)
            template_map[req.name.lower()] = req.specifier
        except Exception:
            continue

    # 检查冲突
    for pkg_str in requested_specs:
        try:
            req = requirements.Requirement(pkg_str)
            template_req = template_map.get(req.name.lower())

            if template_req:
                # 检查版本规范是否兼容
                template_str = str(template_req)
                requested_str = str(req.specifier)

                # 简单策略: 如果两个都指定了固定版本（==）且不同，则冲突
                if "==" in template_str and "==" in requested_str:
                    template_version = template_str.split("==")[1].strip()
                    requested_version = requested_str.split("==")[1].strip()

                    if template_version != requested_version:
                        conflicts.append(
                            f"{req.name}: template has {template_str}, "
                            f"requested {requested_str}"
                        )
        except Exception:
            continue

    return conflicts
```

#### 5.5.2 构建包含依赖的容器配置

```python
def _build_container_config_with_dependencies(
    self,
    dependencies: List[str],
    **kwargs
) -> ContainerConfig:
    """构建包含依赖列表的容器配置"""
    config = ContainerConfig(**kwargs)

    # ✅ 将依赖列表放到 labels 中传递给 Docker Scheduler
    config.labels["dependencies"] = json.dumps(dependencies) if dependencies else ""

    # ✅ 设置 PYTHONPATH 环境变量（依赖安装在本地 /opt/sandbox-venv）
    config.env_vars["PYTHONPATH"] = "/opt/sandbox-venv:/app:/workspace"
    config.env_vars["SANDBOX_VENV_PATH"] = "/opt/sandbox-venv"

    return config
```

#### 5.5.3 等待容器就绪（处理依赖安装失败）

```python
async def _wait_for_container_ready(
    self,
    container_id: str,
    session_id: str,
    timeout: int
) -> None:
    """
    等待容器就绪（处理依赖安装失败）

    Args:
        container_id: 容器 ID
        session_id: 会话 ID（用于清理）
        timeout: 超时时间（秒）
    """
    start = time.time()

    while time.time() - start < timeout:
        # 查询容器状态
        container_info = await self._scheduler.get_container_status(container_id)

        # ✅ 检查容器是否因依赖安装失败而退出
        if container_info.status == "exited":
            exit_code = container_info.exit_code
            logs = await self._scheduler.get_container_logs(container_id, tail=100)

            # 销毁容器并清理 S3
            await self._scheduler.remove_container(container_id)
            await self._cleanup_s3_workspace(session_id)

            raise DependencyInstallationError(
                f"Container exited during dependency installation (exit_code={exit_code})\n"
                f"Logs:\n{logs}"
            )

        # 检查 HTTP 服务是否就绪
        if container_info.status == "running":
            try:
                response = await self._http_client.get(
                    f"http://sandbox-{container_id}:8080/health",
                    timeout=5.0
                )
                if response.status_code == 200:
                    logger.info(f"Container {container_id} is ready (dependencies installed)")
                    return  # 容器就绪（依赖已安装完成）
            except Exception:
                pass  # HTTP 服务未就绪，继续等待

        await asyncio.sleep(2)

    raise TimeoutError(f"Container not ready after {timeout}s")
```

### 5.6 数据模型扩展

#### 5.6.1 Session 实体扩展

**文件**: `sandbox_control_plane/src/domain/entities/session.py`

```python
@dataclass
class InstalledDependency:
    """已安装的依赖信息"""
    name: str
    version: str
    install_location: str  # "/opt/sandbox-venv/"（本地磁盘）
    install_time: datetime
    is_from_template: bool  # True=来自 Template 预装，False=会话动态安装

@dataclass
class Session:
    """会话实体（扩展版）"""
    # 现有字段...
    id: str
    template_id: str
    status: SessionStatus
    resource_limit: ResourceLimit
    workspace_path: str
    runtime_type: str
    runtime_node: str | None = None
    container_id: str | None = None
    pod_name: str | None = None
    env_vars: dict = field(default_factory=dict)
    timeout: int = 300
    created_at: datetime = field(default_factory=datetime.now)
    updated_at: datetime = field(default_factory=datetime.now)
    completed_at: datetime | None = None
    last_activity_at: datetime = field(default_factory=datetime.now)

    # ✅ 新增字段
    requested_dependencies: List[str] = field(default_factory=list)
    installed_dependencies: List[InstalledDependency] = field(default_factory=list)
    dependency_install_status: str = "pending"  # pending, installing, completed, failed
    dependency_install_error: str | None = None
```

#### 5.6.2 Session 数据库模型扩展

**文件**: `sandbox_control_plane/src/infrastructure/persistence/models/session_model.py`

```python
class SessionModel(Base):
    """会话 ORM 模型（扩展版）"""
    __tablename__ = "sessions"

    # 现有字段...
    id: Mapped[str] = mapped_column(String(64), primary_key=True)
    template_id: Mapped[str] = mapped_column(String(64), nullable=False)
    status = Column(Enum("creating", "running", "completed", "failed", "timeout", "terminated"))
    runtime_type = Column(Enum("docker", "kubernetes"))
    runtime_node = Column(String(128))
    container_id = Column(String(128))
    pod_name = Column(String(128))
    workspace_path = Column(String(256))
    resources_cpu = Column(String(16))
    resources_memory = Column(String(16))
    resources_disk = Column(String(16))
    env_vars = Column(JSON)
    timeout = Column(Integer, default=300)
    last_activity_at = Column(DateTime, server_default=func.now())
    created_at = Column(DateTime, server_default=func.now())
    updated_at = Column(DateTime, server_default=func.now(), onupdate=func.now())
    completed_at = Column(DateTime)

    # ✅ 新增字段：依赖管理
    requested_dependencies = Column(JSON, nullable=True)
    installed_dependencies = Column(JSON, nullable=True)
    dependency_install_status = Column(
        Enum("pending", "installing", "completed", "failed", name="dep_install_status"),
        nullable=False,
        default="pending"
    )
    dependency_install_error = Column(Text, nullable=True)
    dependency_install_started_at = Column(DateTime, nullable=True)
    dependency_install_completed_at = Column(DateTime, nullable=True)

    # Indexes
    __table_args__ = (
        Index("ix_sessions_status", "status"),
        Index("ix_sessions_template_id", "template_id"),
        Index("ix_sessions_created_at", "created_at"),
        Index("ix_sessions_runtime_node", "runtime_node"),
        Index("ix_sessions_last_activity_at", "last_activity_at"),
        Index("ix_sessions_dependency_install_status", "dependency_install_status"),  # ✅ 新增
    )
```

### 5.7 安全措施

#### 5.7.1 分层安全模型

| 层级 | 网络访问 | 文件访问 | 用户身份 | 用途 |
|------|---------|---------|---------|------|
| **容器** | ✅ 有网络 | 读写 /workspace, /opt/sandbox-venv | root（启动时） | pip 安装依赖 |
| **Bubblewrap** | ❌ 无网络 | 读写 /workspace | sandbox（非特权） | 用户代码执行 |
| **用户代码** | ❌ 无网络 | 只读 /opt/sandbox-venv（通过 PYTHONPATH） | sandbox | 业务逻辑 |

**核心设计原则**：
- pip 是**可信工具**，容器隔离已经足够
- 容器使用 `sandbox_network`（bridge），可访问 PyPI
- **用户代码仍然通过 Bubblewrap 隔离**（`--unshare-net` 无网络）
- 依赖安装在 `gosu` 切换前，以 root 身份执行，安装到本地 `/opt/sandbox-venv`
- 安装完成后 `chown` 修改属主为 sandbox 用户

#### 5.7.2 包名验证

```python
def validate_package_name(name: str) -> bool:
    """
    验证 PyPI 包名格式

    规范: https://packaging.python.org/specifications/core-metadata/
    """
    import re

    # 禁止路径穿越
    if ".." in name or name.startswith("/"):
        return False

    # 禁止明显的外部 URL
    if "://" in name:
        return False

    # 基本格式检查（仅字母、数字、._-）
    return re.match(r"^[a-zA-Z0-9._-]+$", name) is not None
```

#### 5.7.3 pip 参数限制

**entrypoint 脚本中硬编码 pip 参数**：

```bash
pip3 install \
    --target /opt/sandbox-venv/ \     # 限定安装目录（本地磁盘）
    --cache-dir /tmp/pip-cache \      # 临时缓存目录
    --no-cache-dir \                  # 禁用缓存，节省空间
    --isolated \                      # 隔离模式
    --no-warn-script-location \       # 禁用警告
    --disable-pip-version-check \     # 禁用版本检查
    --index-url https://pypi.org/simple/ \  # 固定 PyPI 源
    $packages                         # 用户提供的包列表
```

**安全特性**：
- ✅ `--target` 限定安装目录到本地磁盘，防止安装到系统路径
- ✅ 依赖安装到 `/opt/sandbox-venv`（非持久化），容器重建时重新安装
- ✅ `--isolated` 隔离模式，忽略环境配置
- ✅ 固定 PyPI 源，防止从恶意源安装
- ✅ 不允许用户覆盖 pip 参数

### 5.8 错误处理与回滚

#### 5.8.1 错误分类

| 错误类型 | 处理策略 | 是否回滚 |
|---------|---------|---------|
| 网络超时（PyPI 不可达） | 容器启动失败，自动回滚 | ✅ 是 |
| 包不存在（404） | 容器启动失败，自动回滚 | ✅ 是 |
| 版本冲突 | 根据 `allow_version_conflicts` 决定 | 可配置 |
| 磁盘空间不足 | 容器启动失败，自动回滚 | ✅ 是 |
| pip 崩溃 | 容器启动失败，自动回滚 | ✅ 是 |

#### 5.8.2 回滚流程

```python
async def _wait_for_container_ready(self, container_id: str, session_id: str):
    """等待容器就绪（处理依赖安装失败）"""
    # ... 检查容器状态

    if container_info.status == "exited":
        # ✅ 自动回滚
        # 1. 销毁容器
        await self._scheduler.remove_container(container_id)

        # 2. 清理 S3 workspace
        await self._cleanup_s3_workspace(session_id)

        # 3. 更新会话状态为 FAILED
        session.status = SessionStatus.FAILED
        session.dependency_install_error = "Container exited during dependency installation"
        await self._session_repo.save(session)

        raise DependencyInstallationError(...)
```

### 5.9 使用示例

#### 5.9.1 创建会话时安装依赖

```bash
curl -X POST http://localhost:8000/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{
    "template_id": "python3.11-baseline",
    "dependencies": [
      {"name": "requests", "version": "==2.31.0"},
      {"name": "pandas", "version": ">=2.0"},
      {"name": "numpy"}
    ],
    "install_timeout": 600,
    "fail_on_dependency_error": true
  }'
```

**响应**：

```json
{
  "session_id": "sess_abc123",
  "status": "running",
  "template_id": "python3.11-baseline",
  "container_id": "container_xyz",
  "created_at": "2025-01-13T10:30:00Z",
  "installed_dependencies": [
    {
      "name": "requests",
      "version": "2.31.0",
      "install_location": "/opt/sandbox-venv/",
      "install_time": "2025-01-13T10:30:15Z",
      "is_from_template": false
    }
  ]
}
```

#### 5.9.2 代码中使用已安装的包

```python
def handler(event):
    """
    AWS Lambda-style handler

    依赖已安装到 /opt/sandbox-venv/（本地磁盘），可通过 PYTHONPATH 自动导入
    """
    import requests  # ✅ 已安装到 /opt/sandbox-venv/
    import pandas as pd
    import numpy as np

    # 使用第三方库
    response = requests.get("https://api.github.com")
    data = response.json()

    df = pd.DataFrame(data)
    summary = {
        "count": len(df),
        "columns": list(df.columns),
        "mean": df.mean(numeric_only=True).to_dict()
    }

    return summary
```

### 5.10 增量安装 API（可选功能）

会话运行时，如果用户需要添加新的依赖，可以提供专用的增量安装接口：

#### 5.10.1 Executor REST API

**文件**: `runtime/executor/interfaces/http/rest.py`

```python
@router.post("/install_dependencies")
async def install_dependencies(request: InstallDependenciesRequest):
    """
    增量安装依赖（会话运行时）

    用途：用户需要添加新的依赖包

    安全：
    - 容器有网络（sandbox_network bridge）
    - 以 sandbox 用户执行（已限制权限）
    - 验证包名格式
    - 限制安装目标为 /opt/sandbox-venv/（本地磁盘）
    """
    target_dir = Path("/opt/sandbox-venv/")
    target_dir.mkdir(parents=True, exist_ok=True)

    cmd = [
        "python3", "-m", "pip", "install",
        "--target", str(target_dir),
        "--isolated",
        *request.packages
    ]

    result = subprocess.run(cmd, capture_output=True, text=True, timeout=request.timeout)

    return {
        "status": "success" if result.returncode == 0 else "failed",
        "stdout": result.stdout,
        "stderr": result.stderr,
        "install_time_ms": ...,
    }
```

#### 5.10.2 Control Plane API

```bash
# 会话运行时添加新依赖
curl -X POST http://localhost:8000/api/v1/sessions/sess_abc123/dependencies \
  -H "Content-Type: application/json" \
  -d '[
    {"name": "matplotlib", "version": ">=3.7"}
  ]'
```

### 5.11 与 Template 的集成

#### 5.11.1 Template 预装包

Template 中的 `pre_installed_packages` 在镜像构建时预装，会话创建时无需重新安装：

```json
{
  "template_id": "python-data-science",
  "image": "python:3.11-slim",
  "pre_installed_packages": [
    "numpy==1.24.0",
    "pandas==2.0.0",
    "matplotlib==3.7.0"
  ]
}
```

#### 5.11.2 版本冲突检测示例

```python
# Template 预装: numpy==1.24.0
# 用户请求: numpy==2.0.0
# 结果: 冲突（除非 allow_version_conflicts=true）

# Template 预装: requests>=2.28.0
# 用户请求: requests==2.31.0
# 结果: 兼容（2.31.0 满足 >=2.28.0）
```

### 5.12 性能考虑

#### 5.12.1 超时时间配置

| 配置项 | 默认值 | 范围 | 说明 |
|-------|-------|------|------|
| `install_timeout` | 300s | 30-1800s | 依赖安装超时时间 |
| 容器启动总超时 | `install_timeout + 120s` | - | 包含 S3 挂载时间 |

#### 5.12.2 包数量限制

- 最多 50 个包（防止滥用）
- 建议单次安装不超过 20 个包（性能考虑）

#### 5.12.3 网络访问

- ✅ 容器使用 `sandbox_network`（bridge），可访问 PyPI
- ❌ 用户代码通过 Bubblewrap 执行（`--unshare-net` 无网络）
- ✅ Executor 可与 Control Plane 通信（上报 ready）

### 5.13 后续优化方向

1. **预装镜像**：制作包含常用依赖的 Docker 镜像（Template），减少安装时间
2. **缓存机制**：在 S3 或本地缓存 pip 下载的包，避免重复下载
3. **并行安装**：无依赖关系的包并行安装
4. **依赖分析**：分析包依赖树，提前检测潜在冲突

---

## 6. Python 依赖配置

### 5.1 核心依赖

使用 MariaDB 需要以下 Python 包：

```txt
# requirements.txt

# Web 框架
fastapi>=0.104.0
uvicorn[standard]>=0.24.0
pydantic>=2.5.0
pydantic-settings>=2.1.0

# 数据库相关
sqlalchemy[asyncio]>=2.0.23
aiomysql>=0.2.0          # 异步 MySQL/MariaDB 驱动
alembic>=1.12.0          # 数据库迁移工具

# HTTP 客户端
httpx>=0.25.0

# 容器运行时
aiodocker>=0.21.0        # Docker API
kubernetes>=28.0.0       # K8s Python 客户端

# 对象存储
boto3>=1.29.0            # S3 兼容存储

# 工具库
python-jose[cryptography]>=3.3.0  # JWT
python-multipart>=0.0.6
structlog>=23.2.0        # 结构化日志
```

### 5.2 开发依赖

```txt
# requirements-dev.txt

# 测试
pytest>=7.4.0
pytest-asyncio>=0.21.0
pytest-cov>=4.1.0
httpx>=0.25.0            # 用于测试 API

# 代码质量
black>=23.11.0
flake8>=6.1.0
mypy>=1.7.0
isort>=5.12.0

# 类型存根
types-redis>=4.6.0.11    # 如果需要使用 Redis 作为缓存层
```

### 5.3 数据库迁移 (Alembic)

```python
# alembic/env.py
from asyncio import run
from sqlalchemy import pool
from sqlalchemy.ext.asyncio import async_engine_from_config
from alembic import context

# this is the Alembic Config object
config = context.config

# add your model's MetaData object here for 'autogenerate' support
from sandbox_control_plane.db.models import Base
target_metadata = Base.metadata

def run_migrations_offline() -> None:
    """Run migrations in 'offline' mode."""
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
        compare_type=True,
    )

    with context.begin_transaction():
        context.run_migrations()


def do_run_migrations(connection):
    context.configure(
        connection=connection,
        target_metadata=target_metadata,
        compare_type=True,
    )

    with context.begin_transaction():
        context.run_migrations()


async def run_async_migrations():
    """Run migrations in 'online' mode with async connection."""
    connectable = async_engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    async with connectable.connect() as connection:
        await connection.run_sync(do_run_migrations)

    await connectable.dispose()


def run_migrations_online() -> None:
    """Run migrations in 'online' mode."""
    run(run_async_migrations())


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
```

### 5.4 数据库配置

```python
# config.py
from pydantic_settings import BaseSettings

class DatabaseSettings(BaseSettings):
    url: str = "mysql+aiomysql://sandbox:sandbox_pass@localhost:3306/sandbox"
    pool_size: int = 50
    max_overflow: int = 100
    pool_recycle: int = 3600
    pool_pre_ping: bool = True
    echo: bool = False

    class Config:
        env_prefix = "DB_"

class Settings(BaseSettings):
    database: DatabaseSettings = DatabaseSettings()
    s3_endpoint: str = "http://localhost:9000"
    runtime_mode: str = "docker"

    class Config:
        env_file = ".env"
```
