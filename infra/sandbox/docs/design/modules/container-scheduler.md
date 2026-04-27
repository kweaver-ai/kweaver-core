# 2.2 Container Scheduler 模块


> **文档导航**: [返回架构总览](../architecture/overview.md)


### 2.2 Container Scheduler 模块

运行时负责管理沙箱容器的生命周期。系统采用容器隔离 + Bubblewrap 进程隔离的双层安全机制。

```
┌─────────────────────────────────────────┐
│ 宿主机 (Host)                            │
│  ├─ Docker Engine / Kubernetes          │
│  └─ 运行时管理器                         │
│     ├─ 创建容器                          │
│     └─ 监控容器                          │
└─────────────────────────────────────────┘
           ↓ 创建容器
┌─────────────────────────────────────────┐
│ 容器 (Container) - 第一层隔离            │
│  ├─ 独立文件系统 (Union FS)              │
│  ├─ 网络隔离 (NetworkMode=none)          │
│  ├─ 资源限制 (CPU/Memory/PID)            │
│  ├─ 能力限制 (CAP_DROP=ALL)              │
│  └─ 非特权用户 (sandbox:sandbox)         │
│                                          │
│  ┌────────────────────────────────────┐ │
│  │ 执行器进程 (Executor)               │ │
│  │  - 监听管理中心的执行请求           │ │
│  │  - 接收用户代码                     │ │
│  │  - 调用 bwrap 启动用户代码          │ │
│  │  - 收集执行结果                     │ │
│  └────────────────────────────────────┘ │
│           ↓ 调用 bwrap                   │
│  ┌────────────────────────────────────┐ │
│  │ Bubblewrap 沙箱 - 第二层隔离       │ │
│  │  ├─ 新的命名空间 (PID/NET/MNT...)  │ │
│  │  ├─ 只读文件系统                    │ │
│  │  ├─ 临时工作目录 (tmpfs)            │ │
│  │  ├─ /proc, /dev 最小化挂载          │ │
│  │  └─ seccomp 系统调用过滤            │ │
│  │                                     │ │
│  │  ┌──────────────────────────────┐  │ │
│  │  │ 用户代码进程                  │  │ │
│  │  │  - Python/Node.js/Shell      │  │ │
│  │  │  - 受 bwrap 完全限制          │  │ │
│  │  └──────────────────────────────┘  │ │
│  └────────────────────────────────────┘ │
└─────────────────────────────────────────┘
```
#### 2.2.1 Docker 运行时

容器配置（第一层隔离）：

```python
class DockerRuntime:
    def __init__(self):
        self.docker_client = aiodocker.Docker()
    
    async def create_container(self, session: Session) -> str:
        template = await self.get_template(session.template_id)
        
        # 容器配置 - 第一层隔离
        config = {
            "Image": template.image,
            # 容器启动后运行执行器
            "Cmd": ["/usr/local/bin/sandbox-executor"],
            "Env": self._build_env_vars(session),
            "WorkingDir": "/workspace",
            
            # 主机配置 - 容器层隔离
            "HostConfig": {
                # 资源限制
                "Memory": session.resources.memory_bytes,
                "MemorySwap": session.resources.memory_bytes,  # 禁用 swap
                "CpuQuota": session.resources.cpu_quota,
                "CpuPeriod": 100000,
                "PidsLimit": 128,  # 限制最大进程数

                # 网络隔离
                "NetworkMode": "none",  # 默认完全隔离网络

                # 安全配置
                "CapDrop": ["ALL"],  # 删除所有 Linux Capabilities
                "SecurityOpt": [
                    "no-new-privileges",  # 禁止进程获取新权限
                    "seccomp=default.json"  # Seccomp 配置
                ],

                # Volume 挂载
                "Binds": [
                    f"{session.s3_volume_path}:/workspace",  # S3 对象存储通过 FUSE/s3fs 挂载
                ],

                # 文件系统
                "ReadonlyRootfs": False,  # 根目录可写（执行器需要）
                "Tmpfs": {
                    "/tmp": "rw,noexec,nosuid,size=512m",  # 临时目录
                },

                # 日志配置
                "LogConfig": {
                    "Type": "json-file",
                    "Config": {
                        "max-size": "10m",
                        "max-file": "3"
                    }
                }
            },
            
            # 用户配置
            "User": "sandbox:sandbox",  # 非特权用户 (UID:GID = 1000:1000)
        }
        
        container = await self.docker_client.containers.create(config)
        await container.start()
        
        # 等待执行器就绪
        await self._wait_for_executor_ready(container.id)
        
        return container.id
    
    async def execute(self, session_id: str, request: ExecuteRequest) -> str:
        """向容器内的执行器发送执行请求"""
        container = await self.get_container(session_id)
        
        # 通过容器内的 HTTP API 与执行器通信
        executor_url = f"http://container-{session_id}:8080"
        
        async with httpx.AsyncClient() as client:
            response = await client.post(
                f"{executor_url}/execute",
                json={
                    "code": request.code,
                    "language": request.language,
                    "timeout": request.timeout,
                    "stdin": request.stdin
                }
            )
        
        return response.json()["execution_id"]
    
    def _build_env_vars(self, session: Session) -> List[str]:
        """构建容器环境变量"""
        env_vars = [
            f"SESSION_ID={session.id}",
            f"CONTROL_PLANE_URL={self.control_plane_url}",
            f"EXECUTION_TIMEOUT={session.timeout}",
        ]
        
        # 用户自定义环境变量
        for key, value in session.env_vars.items():
            env_vars.append(f"{key}={value}")
        
        return env_vars
    
    async def _wait_for_executor_ready(self, container_id: str, timeout: int = 10):
        """等待容器内执行器启动完成"""
        start_time = time.time()
        
        while time.time() - start_time < timeout:
            try:
                # 检查执行器是否响应
                exec_result = await self.docker_client.containers.get(container_id).exec_run(
                    cmd=["curl", "-f", "http://localhost:8080/health"],
                    stdout=True
                )
                
                if exec_result.exit_code == 0:
                    return
            except Exception:
                pass
            
            await asyncio.sleep(0.5)
        
        raise TimeoutError(f"Executor not ready in container {container_id}")

    # ========== 状态查询接口（状态同步服务专用） ==========

    async def is_container_running(self, container_id: str) -> bool:
        """
        检查容器是否正在运行

        直接通过 Docker API 查询，不依赖数据库。
        此方法供 StateSyncService 使用。

        Args:
            container_id: 容器 ID

        Returns:
            bool: 容器是否运行中
        """
        try:
            container = await self.docker_client.containers.get(container_id)
            info = await container.show()
            return info["State"]["Status"] == "running"
        except Exception:
            return False

    async def get_container_status(self, container_id: str) -> ContainerInfo:
        """
        获取容器详细状态

        直接通过 Docker API 查询，不依赖数据库。
        此方法供 StateSyncService 使用。

        Args:
            container_id: 容器 ID

        Returns:
            ContainerInfo: 包含 status, health, created_at 等信息
        """
        container = await self.docker_client.containers.get(container_id)
        info = await container.show()

        return ContainerInfo(
            id=container_id,
            status=info["State"]["Status"],
            created_at=info["Created"],
            health=info["State"].get("Health", {}).get("Status", "unknown"),
            image=info["Config"]["Image"],
            labels=info["Config"].get("Labels", {}),
        )
```

容器镜像构建：

```
# Dockerfile - 沙箱执行环境镜像
FROM python:3.11-slim

# 安装必要工具
RUN apt-get update && apt-get install -y \
    bubblewrap \
    curl \
    && rm -rf /var/lib/apt/lists/*

# 创建非特权用户
RUN groupadd -g 1000 sandbox && \
    useradd -m -u 1000 -g sandbox sandbox

# 安装执行器
COPY sandbox-executor /usr/local/bin/sandbox-executor
RUN chmod +x /usr/local/bin/sandbox-executor

# 创建工作目录
RUN mkdir -p /workspace && chown sandbox:sandbox /workspace

# 切换到非特权用户
USER sandbox

WORKDIR /workspace

# 启动执行器（监听 8080 端口）
CMD ["/usr/local/bin/sandbox-executor"]
```

#### 2.2.2 Kubernetes 运行时

**Implementation Status**: ✅ Completed (2025-01-14)

The K8s scheduler is implemented in `sandbox_control_plane/src/infrastructure/container_scheduler/k8s_scheduler.py` following the design specification below.

Key implementation details:
- **File**: `k8s_scheduler.py` (~29KB, 750+ lines)
- **API Client**: Official Kubernetes Python client (`kubernetes` package)
- **Config Methods**: In-cluster config (ServiceAccount), kubeconfig file, or default config
- **Namespace**: Configurable (default: `sandbox-runtime`)
- **Interface**: Implements `IContainerScheduler` base interface

Pod 配置（第一层隔离）：
```python
class K8sScheduler:
    def _build_pod_spec(self, config: ContainerConfig) -> V1Pod:
        return V1Pod(
            metadata=V1ObjectMeta(
                name=config.name,
                labels={
                    "app": "sandbox",
                    "session": config.name,
                    **config.labels
                }
            ),
            spec=V1PodSpec(
                # 容器配置
                containers=[V1Container(
                    name="executor",
                    image=config.image,
                    command=["/usr/local/bin/sandbox-executor"],

                    # 环境变量
                    env=[V1EnvVar(name=k, value=v) for k, v in config.env_vars.items()],

                    # 资源限制
                    resources=V1ResourceRequirements(
                        limits={
                            "cpu": config.cpu_limit,
                            "memory": config.memory_limit,
                            "ephemeral-storage": config.disk_limit
                        },
                        requests={
                            "cpu": config.cpu_limit,
                            "memory": config.memory_limit
                        }
                    ),

                    # 安全上下文 - 容器层隔离
                    security_context=V1SecurityContext(
                        # 非特权模式
                        privileged=False,
                        # 非 root 用户
                        run_as_non_root=True,
                        run_as_user=1000,
                        run_as_group=1000,
                        # 只读根文件系统（执行器目录除外）
                        read_only_root_filesystem=False,
                        # 禁止权限提升
                        allow_privilege_escalation=False,
                        # 删除所有 Capabilities
                        capabilities=V1Capabilities(
                            drop=["ALL"]
                        ),
                        # Seccomp 配置
                        seccomp_profile=V1SeccompProfile(
                            type="RuntimeDefault"
                        )
                    ),

                    # 卷挂载
                    volume_mounts=[
                        V1VolumeMount(
                            name="workspace",
                            mount_path="/workspace"
                        )
                    ]
                )],

                # Pod 安全配置
                security_context=V1PodSecurityContext(
                    fs_group=1000,
                    run_as_non_root=True,
                    run_as_user=1000,
                    # Sysctl 限制
                    sysctls=[
                        V1Sysctl(name="net.ipv4.ping_group_range", value="1000 1000")
                    ]
                ),

                # 卷定义 - 支持两种模式：
                # 1. S3 path (s3://...) - 使用 s3fs sidecar 或 PVC
                # 2. EmptyDir - 临时存储
                volumes=self._build_volumes(config),

                # 重启策略
                restart_policy="Never",

                # DNS 策略
                dns_policy="None",  # 禁用 DNS

                # 主机网络配置
                host_network=False,
                host_pid=False,
                host_ipc=False
            )
        )

    def _build_volumes(self, config: ContainerConfig) -> List[V1Volume]:
        """
        构建卷定义

        支持两种模式：
        1. S3 workspace - 创建 PVC 或使用 s3fs sidecar
        2. EmptyDir - 临时存储（用于本地测试）
        """
        workspace_path = config.workspace_path

        if workspace_path.startswith("s3://"):
            # S3 模式：创建 PVC（如果尚不存在）
            pvc_name = f"{config.name}-workspace"
            self._ensure_pvc_exists(pvc_name, workspace_path)

            return [
                V1Volume(
                    name="workspace",
                    persistent_volume_claim=V1PersistentVolumeClaimVolumeSource(
                        claim_name=pvc_name
                    )
                )
            ]
        else:
            # EmptyDir 模式（本地测试）
            return [
                V1Volume(
                    name="workspace",
                    empty_dir=V1EmptyDirVolumeSource(
                        medium="Memory",
                        size_limit=config.disk_limit
                    )
                )
            ]
```

**实际实现特性**:

1. **多配置加载方式**:
   - In-cluster config (ServiceAccount Token)
   - kubeconfig 文件 (本地开发)
   - 默认配置 (fallback)

2. **S3 Workspace 支持**:
   - 自动创建 PVC 用于 S3 挂载
   - 解析 S3 URL 获取 bucket 和路径
   - 支持 s3fs sidecar 容器模式

3. **Pod 生命周期管理**:
   - `create_container()` - 创建 Pod 并等待就绪
   - `start_container()` - 启动 Pod
   - `stop_container()` - 删除 Pod
   - `get_container_status()` - 获取 Pod 状态
   - `get_container_logs()` - 获取 Pod 日志
   - `is_container_running()` - 检查运行状态

4. **Python 依赖安装支持** (章节 5 设计):
   - 预装包列表 (template level)
   - 按需安装 (per-session)
   - 依赖安装状态持久化

**K8s 部署文件位置**:
- `deploy/manifests/00-namespace.yaml` - 命名空间定义
- `deploy/manifests/01-configmap.yaml` - 配置管理
- `deploy/manifests/02-secret.yaml` - 密钥管理
- `deploy/manifests/03-serviceaccount.yaml` - ServiceAccount
- `deploy/manifests/04-role.yaml` - RBAC 权限
- `deploy/manifests/05-control-plane-deployment.yaml` - Control Plane 部署
- `deploy/manifests/07-minio-deployment.yaml` - MinIO 部署
- `deploy/manifests/08-mariadb-deployment.yaml` - MariaDB 部署
- `deploy/manifests/deploy.sh` - 一键部署脚本

**Helm Chart 模板文件位置**:
- `deploy/helm/sandbox/templates/s3fs-secret.yaml` - s3fs 密钥配置
- `deploy/helm/sandbox/values.yaml` - s3fs 配置项


##### 2.2.2.1 S3 Workspace 挂载

**Implementation Status**: ✅ Fully Implemented (s3fs + mount --bind)

Kubernetes 调度器使用 **s3fs + bind mount** 方式实现 S3 workspace 挂载。

> **详细文档**: 完整的 S3 workspace 挂载架构、组件说明、配置和实现细节请参考 [10-minio-only-architecture.md](10-minio-only-architecture.md)。

**快速概览**:

- **挂载方式**: 容器内启动脚本挂载 s3fs，使用 bind mount 覆盖 /workspace
- **存储后端**: MinIO（S3-compatible 对象存储）
- **关键文件**:
  - `src/infrastructure/container_scheduler/k8s_scheduler.py` - Pod 卷挂载配置
  - `src/infrastructure/storage/s3_storage.py` - S3 API 包装器
  - `deploy/helm/sandbox/templates/s3fs-secret.yaml` - s3fs 凭证
  - `deploy/helm/sandbox/values.yaml` - s3fs 配置

**数据流**:
1. **文件上传**: Client → Control Plane API → S3Storage Service → MinIO (S3 API)
2. **容器内访问**: Executor → POSIX I/O (/workspace) → s3fs → MinIO (S3 API)
3. **文件下载**: Client → Control Plane API → presigned URL → MinIO (direct)

**关键特性**:
- 无需 CSI Driver - s3fs 在容器内运行
- 无需元数据数据库 - 文件路径结构提供所有组织
- Control Plane 直接通过 S3 API 写入
- 每个容器有独立的 s3fs 进程（隔离）
