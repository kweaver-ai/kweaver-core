"""
Kubernetes 容器调度器

使用官方 Python kubernetes 客户端实现 Pod 的创建和管理。

MinIO + s3fs 架构：
- Control Plane 通过 S3 API 将文件写入 MinIO 的 /sessions/{session_id}/ 路径
- Executor Pod 在启动脚本中挂载 s3fs，将 S3 bucket 的 session 子目录挂载到 /workspace
- s3fs 进程和 executor 进程运行在同一容器内

Python 依赖安装：按照 sandbox-design-v2.1.md 章节 5 设计。
"""
import asyncio
import json
import os
from typing import Optional, List
from urllib.parse import urlparse

from kubernetes import client, config
from kubernetes.client import (
    V1Pod,
    V1PodSpec,
    V1ObjectMeta,
    V1Container,
    V1ContainerPort,
    V1EnvVar,
    V1ResourceRequirements,
    V1Volume,
    V1VolumeMount,
    V1HostPathVolumeSource,
    V1PersistentVolumeClaim,
    V1PersistentVolumeClaimSpec,
    V1PersistentVolumeClaimVolumeSource,
    V1ObjectMeta as V1ObjectMeta_imported,
    V1SecurityContext,
    V1Capabilities,
    V1PodSecurityContext,
    V1EmptyDirVolumeSource,
    V1SecretVolumeSource,
)
from kubernetes.client.rest import ApiException

from src.infrastructure.container_scheduler.base import (
    IContainerScheduler,
    ContainerConfig,
    ContainerInfo,
    ContainerResult,
)
from src.infrastructure.config.settings import get_settings
from src.infrastructure.logging import get_logger
from src.shared.utils.dependencies import format_dependencies_for_script, format_dependency_install_script_for_shell

logger = get_logger(__name__)


def s3_prefix_from_path(prefix: str) -> str:
    """
    从 S3 路径前缀中提取会话 ID

    Args:
        prefix: S3 路径前缀，如 "sessions/test-001/workspace"

    Returns:
        会话 ID，如 "test-001"
    """
    parts = prefix.strip('/').split('/')
    if len(parts) >= 2 and parts[0] == "sessions":
        return parts[1]
    return prefix


class K8sScheduler(IContainerScheduler):
    """
    Kubernetes 容器调度器

    通过 Kubernetes API 管理 Pod 生命周期。
    """

    def __init__(
        self,
        namespace: str = "sandbox-system",
        kube_config_path: Optional[str] = None,
        service_account_token: Optional[str] = None,
        executor_service_account: str = "sandbox-control-plane",
    ):
        """
        初始化 K8s 调度器

        Args:
            namespace: Kubernetes 命名空间
            kube_config_path: kubeconfig 文件路径（可选，用于本地开发）
            service_account_token: ServiceAccount Token（用于 Pod 内运行）
            executor_service_account: Executor Pod 使用的 ServiceAccount 名称
        """
        self._namespace = namespace
        self._executor_service_account = executor_service_account

        # 加载 Kubernetes 配置
        if service_account_token:
            # 在 Pod 内运行，使用 ServiceAccount
            self._load_incluster_config()
        elif kube_config_path:
            # 使用指定的 kubeconfig 文件
            config.load_kube_config(config_file=kube_config_path)
        else:
            # 尝试加载默认 kubeconfig
            try:
                config.load_kube_config()
            except Exception:
                # 如果 kubeconfig 不存在，尝试使用 in-cluster 配置
                try:
                    self._load_incluster_config()
                except Exception:
                    # 最后尝试使用默认配置（用于本地开发）
                    from kubernetes.client import Configuration
                    Configuration.set_default(Configuration())

        # 创建 API 客户端
        self._core_v1 = client.CoreV1Api()
        self._initialized = False

    def _load_incluster_config(self):
        """加载 in-cluster 配置"""
        config.load_incluster_config()
        logger.info("Loaded in-cluster Kubernetes config")

    async def _ensure_connected(self) -> bool:
        """确保 K8s 连接已建立"""
        if not self._initialized:
            try:
                # 测试连接 - 使用当前 namespace 的 Pod 列表代替 namespace 列表
                # 这样只需要 namespace 级别的权限，不需要 cluster 级别的权限
                self._core_v1.list_namespaced_pod(self._namespace, limit=1)
                self._initialized = True
                logger.info(f"Connected to Kubernetes cluster, namespace: {self._namespace}")
            except Exception as e:
                logger.error(f"Failed to connect to Kubernetes: {e}")
                raise
        return self._initialized

    async def close(self) -> None:
        """关闭连接（Kubernetes 客户端是无状态的，无需显式关闭）"""
        self._initialized = False

    def _parse_s3_workspace(self, workspace_path: str) -> Optional[dict]:
        """
        解析 S3 workspace 路径

        Args:
            workspace_path: S3 路径，格式: s3://bucket/sessions/{session_id}/

        Returns:
            包含 bucket, prefix 的字典，如果不是 S3 路径则返回 None
        """
        if not workspace_path or not workspace_path.startswith("s3://"):
            return None

        parsed = urlparse(workspace_path)
        return {
            "bucket": parsed.netloc,
            "prefix": parsed.path.lstrip('/'),
        }

    def _build_pod_name(self, session_id: str) -> str:
        """生成 Pod 名称"""
        # K8s Pod 名称需要符合 DNS 子域名规则
        # 只能包含小写字母、数字和 '-'，且开头和结尾必须是字母数字
        pod_name = f"sandbox-{session_id.lower()}"
        # 替换不符合规则的字符
        pod_name = ''.join(c if c.isalnum() or c == '-' else '-' for c in pod_name)
        # 确保不以 '-' 开头或结尾
        pod_name = pod_name.strip('-')
        # 限制长度（K8s Pod 名称最多 253 字符）
        return pod_name[:253]


    def _build_executor_container(
        self,
        config: ContainerConfig,
        use_s3_mount: bool,
        has_dependencies: bool,
        session_id: str = None,
        s3_workspace: dict = None,
    ) -> V1Container:
        """
        构建主 executor 容器

        如果使用 S3 挂载，启动脚本会先挂载 s3fs，然后启动 executor

        Args:
            config: 容器配置
            use_s3_mount: 是否使用 S3 挂载（通过 s3fs 在容器内挂载）
            has_dependencies: 是否有依赖包
            session_id: 会话 ID（用于 S3 子目录挂载）
            s3_workspace: S3 workspace 配置（包含 bucket, prefix）

        Returns:
            V1Container 对象
        """
        env_vars = [
            V1EnvVar(name=k, value=v)
            for k, v in config.env_vars.items()
        ]

        # 添加 S3 相关环境变量
        s3_workspace = self._parse_s3_workspace(config.workspace_path)
        if s3_workspace:
            env_vars.extend([
                V1EnvVar(name="WORKSPACE_PATH", value="/workspace"),
                V1EnvVar(name="S3_BUCKET", value=s3_workspace["bucket"]),
                V1EnvVar(name="S3_PREFIX", value=s3_workspace["prefix"]),
            ])

        # 添加 PYTHONPATH 环境变量以支持依赖导入
        if has_dependencies:
            # 依赖安装到本地 /opt/sandbox-venv
            env_vars.append(V1EnvVar(
                name="PYTHONPATH",
                value="/opt/sandbox-venv:/app:/workspace"
            ))
            env_vars.append(V1EnvVar(
                name="SANDBOX_VENV_PATH",
                value="/opt/sandbox-venv"
            ))

        # 资源限制
        resources = V1ResourceRequirements(
            limits={
                "cpu": config.cpu_limit,
                "memory": config.memory_limit,
                "ephemeral-storage": config.disk_limit,
            },
            requests={
                "cpu": config.cpu_limit,
                "memory": config.memory_limit,
            },
        )

        # 容器端口
        ports = [
            V1ContainerPort(
                container_port=8080,
                name="executor",
                protocol="TCP",
            )
        ]

        # 卷挂载
        volume_mounts = [
            V1VolumeMount(
                name="workspace",
                mount_path="/workspace",
            )
        ]
        if use_s3_mount:
            volume_mounts.append(
                V1VolumeMount(
                    name="s3fs-passwd",
                    mount_path="/etc/s3fs-passwd",
                    read_only=True,
                )
            )

        # 安全上下文 - s3fs 需要 privileged 和 root 用户
        # 注意：privileged=True 时容器必须以 root 运行，以便进行 FUSE 挂载
        # 需要显式设置 runAsUser=0 来覆盖 Dockerfile 中的 USER 指令
        # 有依赖时也需要 root 来安装依赖，然后用 gosu 切换到 sandbox 用户
        needs_root = use_s3_mount or has_dependencies
        security_context = V1SecurityContext(
            # s3fs 挂载或依赖安装需要 root，最终使用 gosu 切换到 sandbox 用户
            run_as_non_root=not needs_root,
            run_as_user=0 if needs_root else 1000,  # 0 = root，显式设置以覆盖 Dockerfile USER
            run_as_group=0 if needs_root else 1000,
            allow_privilege_escalation=use_s3_mount,  # s3fs 需要特权
            capabilities=V1Capabilities(drop=["ALL"]) if not needs_root else None,
            read_only_root_filesystem=False,
            privileged=use_s3_mount,  # s3fs 需要 privileged 模式
        )

        # 构建启动命令
        command = None
        if use_s3_mount:
            # 获取设置
            settings = get_settings()
            minio_url = settings.s3_endpoint_url or "http://minio.sandbox-system.svc.cluster.local:9000"
            bucket = s3_workspace["bucket"]

            # S3 挂载脚本（使用 bucket 挂载 + bind mount 方案）
            s3_prefix = s3_workspace["prefix"].rstrip('/')
            mount_script = f"""#!/bin/sh
set -e

echo "📂 Mounting S3 bucket {bucket} to /mnt/s3-root (session: {session_id})..."

# 挂载整个 S3 bucket 到临时位置（uid=1000,gid=1000 让挂载点对 sandbox 用户可访问）
mkdir -p /mnt/s3-root
s3fs {bucket} /mnt/s3-root \\
    -o url={minio_url} \\
    -o use_path_request_style \\
    -o allow_other \\
    -o uid=1000 \\
    -o gid=1000 \\
    -o passwd_file=/etc/s3fs-passwd/s3fs-passwd &

S3FS_PID=$!
echo "s3fs started with PID: $S3FS_PID"

# 等待挂载完成
sleep 2

# 创建 session workspace 目录（使用完整 S3 前缀）
SESSION_PATH="/mnt/s3-root/{s3_prefix}"
echo "Ensuring session workspace exists: $SESSION_PATH"
mkdir -p "$SESSION_PATH"

# 使用 bind mount 将 S3 路径挂载到 /workspace（/workspace 是 emptyDir 挂载点）
mount --bind "$SESSION_PATH" /workspace

# 验证 bind mount
echo "Workspace bind mounted: $(ls -la /workspace)"

echo "✅ S3 bucket mounted and workspace linked successfully"
ls -la /workspace/

"""

            # 如果有依赖，安装依赖
            if has_dependencies:
                dependencies_json = config.labels.get("dependencies", "")
                dependencies = json.loads(dependencies_json) if dependencies_json else []
                _, deps_list = format_dependencies_for_script(dependencies)
                mount_script += f"""
echo "📦 Installing dependencies..."
VENV_DIR="/opt/sandbox-venv"
mkdir -p $VENV_DIR

pip3 install \\
    --target $VENV_DIR \\
    --no-cache-dir \\
    --no-warn-script-location \\
    --disable-pip-version-check \\
    --index-url https://mirrors.aliyun.com/pypi/web/simple/ \\
    {deps_list}

echo "✅ Dependencies installed"

export PYTHONPATH="$VENV_DIR:/app:/workspace"
export SANDBOX_VENV_PATH="$VENV_DIR"
"""

            # 启动 executor (前台) - 使用 gosu 切换到 sandbox 用户
            mount_script += """
echo "🚀 Starting executor as sandbox user..."
# 使用 gosu 切换到 sandbox 用户并启动 executor
# gosu 会正确传递信号，避免进程变成僵尸
exec gosu sandbox python -m executor.interfaces.http.rest
"""
            command = ["sh", "-c", mount_script]

        elif has_dependencies:
            # 依赖由 executor 容器在启动时安装
            dependencies_json = config.labels.get("dependencies", "")
            dependencies = json.loads(dependencies_json) if dependencies_json else []
            install_script = format_dependency_install_script_for_shell(dependencies)

            install_script = f"""#!/bin/sh
set -e
echo "📦 Installing dependencies..."

{install_script}

# 修复 venv 目录权限（以 root 安装，需要让 sandbox 用户可读）
chown -R sandbox:sandbox /opt/sandbox-venv

# 启动 executor - 使用 gosu 切换到 sandbox 用户
echo "🚀 Starting executor as sandbox user..."
exec gosu sandbox python -m executor.interfaces.http.rest
"""
            command = ["sh", "-c", install_script]

        return V1Container(
            name="executor",
            image=config.image,
            image_pull_policy="IfNotPresent",  # 优先使用本地镜像
            command=command,
            env=env_vars,
            resources=resources,
            ports=ports,
            volume_mounts=volume_mounts,
            security_context=security_context,
        )

    async def create_container(self, config: ContainerConfig) -> str:
        """
        创建 Kubernetes Pod - 使用 s3fs 在 executor 容器内挂载

        架构说明：
        - Control Plane 通过 S3 API 将文件写入 MinIO 的 /sessions/{session_id}/ 路径
        - Executor Pod 在启动脚本中挂载 s3fs，将 S3 bucket 的 session 子目录挂载到 /workspace
        - s3fs 进程和 executor 进程运行在同一容器内
        - 如果有依赖，executor 容器会在启动时安装依赖
        """
        await self._ensure_connected()

        pod_name = self._build_pod_name(config.name)
        s3_workspace = self._parse_s3_workspace(config.workspace_path)
        use_s3_mount = s3_workspace is not None

        # 检查是否有依赖
        dependencies_json = config.labels.get("dependencies", "")
        has_dependencies = bool(dependencies_json)

        # 提取 session_id 仅用于日志记录（挂载使用完整 s3_prefix）
        session_id = s3_prefix_from_path(s3_workspace["prefix"]) if s3_workspace else config.name

        # 构建容器列表 - 只有 executor 容器
        containers = []

        # 构建 executor 容器（在启动脚本中挂载 s3fs）
        # 注意：挂载脚本使用完整的 s3_prefix，而不是 session_id，以避免路径层级问题
        executor_container = self._build_executor_container(
            config=config,
            use_s3_mount=use_s3_mount,
            has_dependencies=has_dependencies,
            session_id=session_id,
            s3_workspace=s3_workspace,
        )
        containers.append(executor_container)

        # 构建卷
        volumes = []
        if use_s3_mount:
            # 使用 emptyDir 用于 s3fs 挂载
            volumes.append(
                V1Volume(
                    name="workspace",
                    empty_dir=V1EmptyDirVolumeSource(),
                )
            )
            # 添加 s3fs-passwd secret
            volumes.append(
                V1Volume(
                    name="s3fs-passwd",
                    secret=V1SecretVolumeSource(
                        secret_name="s3fs-passwd",
                        default_mode=0o400,
                    ),
                )
            )
        else:
            # 本地 workspace：使用 emptyDir
            volumes.append(
                V1Volume(
                    name="workspace",
                    empty_dir=V1EmptyDirVolumeSource(),
                )
            )

        # 构建标签（排除 dependencies，因为 K8s labels 有严格格式限制）
        # dependencies 不能作为 label，因为包含方括号、引号等非法字符
        dependencies_value = config.labels.pop("dependencies", None)
        labels = {
            "app": "sandbox-executor",  # 匹配 sandbox-executor service selector
            "sandbox-session": config.name,
            "sandbox-type": "execution",
        }
        if use_s3_mount:
            labels["mount-method"] = "s3fs"
        labels.update(config.labels)

        # 构建 annotations（dependencies 放在这里，没有格式限制）
        annotations = {
            "sandbox-session-id": config.name,
        }
        if dependencies_value:
            annotations["dependencies"] = dependencies_value
        # 恢复 dependencies 到 config.labels，避免影响后续调用
        if dependencies_value is not None:
            config.labels["dependencies"] = dependencies_value

        # 构建 Pod Spec
        pod = V1Pod(
            metadata=V1ObjectMeta(
                name=pod_name,
                labels=labels,
                annotations=annotations,
            ),
            spec=V1PodSpec(
                containers=containers,
                volumes=volumes,
                restart_policy="Always",  # 确保容器退出后自动重启，保持 runtime 可用
                host_network=False,
                termination_grace_period_seconds=30,
                service_account_name=self._executor_service_account,
                # 使用默认 DNS 策略 (ClusterFirst)，允许 Pod 使用 K8s 集群 DNS
                # 这对于 executor 与 control plane 通信很重要
            ),
        )

        try:
            # 创建 Pod
            created_pod = await asyncio.to_thread(
                self._core_v1.create_namespaced_pod,
                namespace=self._namespace,
                body=pod,
            )
            mount_method = "s3fs" if use_s3_mount else "emptyDir"
            logger.info(
                f"Created pod {created_pod.metadata.name} for session {config.name} "
                f"in namespace {self._namespace} (mount method: {mount_method})"
            )
            return created_pod.metadata.name

        except ApiException as e:
            logger.error(f"Failed to create pod: {e}")
            raise

    async def start_container(self, container_id: str) -> None:
        """
        启动 Pod

        注意：Kubernetes Pod 创建后会自动启动，此方法为兼容接口保留
        """
        await self._ensure_connected()
        # K8s Pod 创建后自动启动，无需显式调用
        logger.debug(f"Pod {container_id} starts automatically after creation")

    async def stop_container(
        self,
        container_id: str,
        timeout: int = 30
    ) -> None:
        """
        停止（删除）Pod

        使用 s3fs 方式时，无需清理 PVC，s3fs 挂载在 Pod 删除时自动清理。

        Args:
            container_id: Pod 名称
            timeout: 优雅终止超时时间（秒）
        """
        await self._ensure_connected()

        try:
            await asyncio.to_thread(
                self._core_v1.delete_namespaced_pod,
                name=container_id,
                namespace=self._namespace,
                grace_period_seconds=timeout,
            )
            logger.info(f"Stopped pod {container_id}")
        except ApiException as e:
            if e.status == 404:
                logger.warning(f"Pod {container_id} not found")
            else:
                logger.error(f"Failed to stop pod {container_id}: {e}")
                raise

    async def remove_container(
        self,
        container_id: str,
        force: bool = False
    ) -> None:
        """
        删除 Pod

        使用 s3fs 方式时，无需清理 PVC，s3fs 挂载在 Pod 删除时自动清理。

        Args:
            container_id: Pod 名称
            force: 是否强制删除（grace_period_seconds=0）
        """
        await self._ensure_connected()

        try:
            await asyncio.to_thread(
                self._core_v1.delete_namespaced_pod,
                name=container_id,
                namespace=self._namespace,
                grace_period_seconds=0 if force else 30,
            )
            logger.info(f"Removed pod {container_id}")
        except ApiException as e:
            if e.status == 404:
                logger.warning(f"Pod {container_id} not found")
            else:
                logger.warning(f"Failed to remove pod {container_id}: {e}")

    async def get_container_status(self, container_id: str) -> ContainerInfo:
        """
        获取 Pod 状态

        Args:
            container_id: Pod 名称

        Returns:
            ContainerInfo 对象
        """
        await self._ensure_connected()
        try:
            pod = await asyncio.to_thread(
                self._core_v1.read_namespaced_pod,
                name=container_id,
                namespace=self._namespace,
            )

            # 转换 K8s Pod 状态到 ContainerInfo
            phase = pod.status.phase
            if phase == "Running" and pod.status.container_statuses:
                for container_status in pod.status.container_statuses:
                    if container_status.name == "executor":
                        if container_status.state.terminated:
                            phase = "exited"
                        elif container_status.state.waiting:
                            phase = "waiting"
                        break

            ip_address = pod.status.pod_ip
            created_at = pod.metadata.creation_timestamp.isoformat() if pod.metadata.creation_timestamp else ""
            started_at = pod.status.start_time.isoformat() if pod.status.start_time else None

            # 获取退出码（如果已终止）
            exit_code = None
            if pod.status.container_statuses:
                for container_status in pod.status.container_statuses:
                    if container_status.name == "executor" and container_status.state.terminated:
                        exit_code = container_status.state.terminated.exit_code
                        break

            # 获取镜像名称
            image = ""
            if pod.spec.containers:
                for container in pod.spec.containers:
                    if container.name == "executor":
                        image = container.image
                        break

            return ContainerInfo(
                id=container_id,
                name=container_id,
                image=image,
                status=phase.lower(),
                ip_address=ip_address,
                created_at=created_at,
                started_at=started_at,
                exited_at=None,
                exit_code=exit_code,
            )

        except ApiException as e:
            if e.status == 404:
                logger.error(f"Pod {container_id} not found")
                raise ValueError(f"Pod {container_id} not found") from e
            else:
                logger.error(f"Failed to get pod status {container_id}: {e}")
                raise

    async def is_container_running(self, container_id: str) -> bool:
        """
        检查 Pod 是否正在运行

        Args:
            container_id: Pod 名称

        Returns:
            bool: Pod 是否运行中
        """
        try:
            container_info = await self.get_container_status(container_id)
            return container_info.status == "running"
        except Exception as e:
            logger.warning(f"Failed to check pod {container_id} status: {e}")
            return False

    async def get_container_logs(
        self,
        container_id: str,
        tail: int = 100,
        since: Optional[str] = None
    ) -> str:
        """
        获取 Pod 日志

        Args:
            container_id: Pod 名称
            tail: 返回最后几行
            since: 时间戳（可选）

        Returns:
            日志字符串
        """
        await self._ensure_connected()
        try:
            logs = await asyncio.to_thread(
                self._core_v1.read_namespaced_pod_log,
                name=container_id,
                namespace=self._namespace,
                container="executor",
                tail_lines=tail,
                since_seconds=None,  # since_time 需要转换
            )
            return logs
        except ApiException as e:
            logger.error(f"Failed to get logs for pod {container_id}: {e}")
            raise

    async def wait_container(
        self,
        container_id: str,
        timeout: Optional[int] = None
    ) -> ContainerResult:
        """
        等待 Pod 执行完成

        Args:
            container_id: Pod 名称
            timeout: 超时时间（秒）

        Returns:
            ContainerResult 对象
        """
        await self._ensure_connected()

        async def _wait() -> ContainerResult:
            while True:
                try:
                    pod = await asyncio.to_thread(
                        self._core_v1.read_namespaced_pod,
                        name=container_id,
                        namespace=self._namespace,
                    )

                    # 检查 Pod 状态
                    if pod.status.phase == "Succeeded":
                        # 获取日志
                        logs = await self.get_container_logs(container_id, tail=-1)
                        return ContainerResult(
                            status="completed",
                            stdout=logs,
                            stderr="",
                            exit_code=0,
                        )
                    elif pod.status.phase == "Failed":
                        # 获取日志
                        logs = await self.get_container_logs(container_id, tail=-1)
                        return ContainerResult(
                            status="failed",
                            stdout=logs,
                            stderr="Pod failed",
                            exit_code=1,
                        )

                    # 检查容器状态
                    if pod.status.container_statuses:
                        for container_status in pod.status.container_statuses:
                            if container_status.name == "executor" and container_status.state.terminated:
                                logs = await self.get_container_logs(container_id, tail=-1)
                                terminated = container_status.state.terminated
                                return ContainerResult(
                                    status="completed" if terminated.exit_code == 0 else "failed",
                                    stdout=logs,
                                    stderr="",
                                    exit_code=terminated.exit_code,
                                )

                    await asyncio.sleep(1)

                except ApiException as e:
                    if e.status == 404:
                        return ContainerResult(
                            status="failed",
                            stdout="",
                            stderr=f"Pod {container_id} not found",
                            exit_code=1,
                        )
                    raise

        try:
            if timeout:
                return await asyncio.wait_for(_wait(), timeout=timeout)
            else:
                return await _wait()

        except asyncio.TimeoutError:
            logger.warning(f"Pod {container_id} timed out")
            return ContainerResult(
                status="timeout",
                stdout="",
                stderr=f"Pod execution timed out after {timeout}s",
                exit_code=124,
            )

    async def ping(self) -> bool:
        """
        检查 Kubernetes 连接状态

        Returns:
            bool: 连接是否正常
        """
        try:
            await self._ensure_connected()
            await asyncio.to_thread(
                self._core_v1.list_namespace,
                limit=1,
            )
            return True
        except Exception as e:
            logger.error(f"Kubernetes ping failed: {e}")
            return False

    def _parse_memory_to_bytes(self, value: str) -> int:
        """
        解析内存限制为字节数

        Args:
            value: 如 "512Mi", "1Gi"

        Returns:
            字节数
        """
        value = value.strip()
        if value.endswith("Gi") or value.endswith("GB") or value.endswith("G"):
            return int(float(value[:-2]) * 1024 * 1024 * 1024)
        elif value.endswith("Mi") or value.endswith("MB") or value.endswith("M"):
            return int(float(value[:-2]) * 1024 * 1024)
        elif value.endswith("Ki") or value.endswith("KB") or value.endswith("K"):
            return int(float(value[:-2]) * 1024)
        else:
            # 默认为 MB
            return int(float(value) * 1024 * 1024)

    def _parse_disk_to_bytes(self, value: str) -> int:
        """解析磁盘限制为字节数"""
        return self._parse_memory_to_bytes(value)
