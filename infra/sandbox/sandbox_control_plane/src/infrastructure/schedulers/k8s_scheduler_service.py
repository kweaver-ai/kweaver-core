"""
Kubernetes 调度服务

实现调度策略，使用 Kubernetes API 创建 Pod。
"""

import logging
import os
from typing import List, Optional

from src.domain.services.scheduler import (
    IScheduler,
    RuntimeNode,
    ScheduleRequest,
)
from src.domain.repositories.template_repository import ITemplateRepository
from src.domain.value_objects.execution_request import ExecutionRequest
from src.infrastructure.container_scheduler.base import (
    IContainerScheduler,
    ContainerConfig,
    ControlPlaneOwnerContext,
)
from src.infrastructure.executors import ExecutorClient

logger = logging.getLogger(__name__)


class K8sSchedulerService(IScheduler):
    """
    Kubernetes 调度服务

    使用 Kubernetes API 创建和管理 Pod：
    1. 不需要节点选择逻辑（K8s 调度器自动处理）
    2. 创建 Pod 而不是容器
    3. Pod 生命周期跟随会话
    """

    def __init__(
        self,
        container_scheduler: IContainerScheduler,
        template_repo: ITemplateRepository,
        executor_client: Optional[ExecutorClient] = None,
        executor_port: int = 8080,
        control_plane_url: str = "http://sandbox-control-plane.sandbox-system.svc.cluster.local:8000",
        disable_bwrap: bool = True,  # K8s 环境默认禁用 bwrap
    ):
        self._container_scheduler = container_scheduler
        self._template_repo = template_repo
        self._executor_client = executor_client or ExecutorClient()
        self._executor_port = executor_port
        self._control_plane_url = control_plane_url
        self._disable_bwrap = disable_bwrap
        self._owner_context = self._load_owner_context()

        # K8s 集群作为单个逻辑节点
        self._cluster_node = RuntimeNode(
            id="k8s-cluster",
            type="kubernetes",
            url="kubernetes://cluster",
            status="healthy",
            cpu_usage=0.0,
            mem_usage=0.0,
            session_count=0,
            max_sessions=1000,
            cached_templates=[],
        )

    def _load_owner_context(self) -> ControlPlaneOwnerContext:
        """加载当前 control plane Pod 的 owner 上下文。"""
        pod_name = os.getenv("POD_NAME", "").strip()
        pod_uid = os.getenv("POD_UID", "").strip()
        if not pod_name or not pod_uid:
            raise RuntimeError(
                "POD_NAME and POD_UID must be set in Kubernetes mode before creating executor pods"
            )

        return ControlPlaneOwnerContext(
            pod_name=pod_name,
            pod_uid=pod_uid,
        )

    async def schedule(self, request: ScheduleRequest) -> RuntimeNode:
        """
        调度会话到 K8s 集群

        在 K8s 环境中，调度决策由 Kubernetes 调度器处理。
        我们只返回一个表示 K8s 集群的虚拟节点。
        """
        logger.info(f"Scheduling session to K8s cluster: template={request.template_id}")
        return self._cluster_node

    async def get_node(self, node_id: str) -> Optional[RuntimeNode]:
        """获取指定节点"""
        if node_id == "k8s-cluster":
            return self._cluster_node
        return None

    async def get_healthy_nodes(self) -> List[RuntimeNode]:
        """获取所有健康节点"""
        return [self._cluster_node]

    async def mark_node_unhealthy(self, node_id: str) -> None:
        """标记节点为不健康"""
        # K8s 环境下不需要此操作
        logger.warning(f"mark_node_unhealthy called in K8s environment: {node_id}")

    async def create_container_for_session(
        self,
        session_id: str,
        template_id: str,
        image: str,
        resource_limit,
        env_vars: dict,
        workspace_path: str,
        node_id: str,
        dependencies: list = None,
    ) -> str:
        """
        为会话创建 Pod

        Args:
            session_id: 会话 ID
            template_id: 模板 ID
            image: 容器镜像
            resource_limit: 资源限制
            env_vars: 环境变量
            workspace_path: 工作空间路径
            node_id: 目标节点 ID（K8s 环境下忽略）
            dependencies: Python 依赖列表

        Returns:
            Pod 名称
        """
        import json

        # 获取模板信息
        template = await self._template_repo.find_by_id(template_id)
        if not template:
            raise RuntimeError(f"Template not found: {template_id}")

        # 创建容器配置
        dependencies_json = json.dumps(dependencies) if dependencies else ""

        # Debug: 打印使用的 CONTROL_PLANE_URL
        logger.info(f"Creating executor pod with CONTROL_PLANE_URL: {self._control_plane_url}")

        config = ContainerConfig(
            image=image,
            name=f"sandbox-{session_id}",
            env_vars={
                **env_vars,
                "SESSION_ID": session_id,
                "WORKSPACE_PATH": workspace_path,
                "CONTROL_PLANE_URL": self._control_plane_url,
                "DISABLE_BWRAP": "true" if self._disable_bwrap else "false",
            },
            cpu_limit=resource_limit.cpu,
            memory_limit=resource_limit.memory,
            disk_limit=resource_limit.disk,
            workspace_path=workspace_path,
            labels={
                "session_id": session_id,
                "template_id": template_id,
                "managed_by": "sandbox-control-plane",
                "dependencies": dependencies_json,
            },
            owner_context=self._owner_context,
        )

        # 创建 Pod
        try:
            pod_name = await self._container_scheduler.create_container(config)
            logger.info(f"Created Pod {pod_name} for session {session_id}")
            return pod_name

        except Exception as e:
            logger.error(f"Failed to create Pod for session {session_id}: {e}")
            raise

    async def destroy_container(self, container_id: str, timeout: int = 10) -> None:
        """
        销毁 Pod
        """
        try:
            await self._container_scheduler.stop_container(container_id, timeout=timeout)
            await self._container_scheduler.remove_container(container_id)
            logger.info(f"Destroyed Pod {container_id}")
        except Exception as e:
            logger.error(f"Failed to destroy Pod {container_id}: {e}")
            raise

    async def get_container_info(self, container_id: str):
        """获取 Pod 信息"""
        return await self._container_scheduler.get_container_status(container_id)

    async def execute(
        self,
        session_id: str,
        container_id: str,
        execution_request: ExecutionRequest,
    ) -> str:
        """
        提交执行请求到 Pod 内的执行器

        Args:
            session_id: 会话 ID
            container_id: Pod 名称
            execution_request: 执行请求

        Returns:
            execution_id: 执行任务 ID
        """
        # 从 K8s API 获取 Pod IP
        executor_url = await self.get_executor_url(container_id)
        logger.info(
            f"Submitting execution to executor: {executor_url}, session_id={session_id}, pod_name={container_id}"
        )

        # 使用执行器客户端提交请求
        try:
            execution_id = await self._executor_client.submit_execution(
                executor_url=executor_url,
                execution_id=execution_request.execution_id or "",
                session_id=session_id,
                code=execution_request.code,
                language=execution_request.language,
                event=execution_request.event,
                timeout=execution_request.timeout,
                env_vars=execution_request.env_vars,
                working_directory=execution_request.working_directory,
            )

            logger.info(
                f"Execution submitted successfully: execution_id={execution_id}, session_id={session_id}"
            )

            return execution_id

        except Exception as e:
            logger.error(f"Failed to submit execution to executor: {executor_url}, error={e}")
            raise

    async def get_executor_url(self, container_id: str) -> str:
        """根据 Pod 名称获取 executor URL。"""
        import asyncio

        pod_name = container_id
        namespace = self._container_scheduler._namespace

        try:
            pod_info = await asyncio.to_thread(
                self._container_scheduler._core_v1.read_namespaced_pod,
                name=pod_name,
                namespace=namespace,
            )
            pod_ip = pod_info.status.pod_ip
            if not pod_ip:
                raise RuntimeError(f"Pod {pod_name} does not have an IP address yet")

            return f"http://{pod_ip}:{self._executor_port}"
        except Exception as e:
            logger.error(f"Failed to get pod IP for {pod_name}: {e}")
            raise
