"""
Docker 调度服务

实现调度策略，选择最优节点并创建容器。
"""

from typing import List, Optional

from src.domain.services.scheduler import (
    IScheduler,
    RuntimeNode,
    ScheduleRequest,
)
from src.domain.repositories.runtime_node_repository import IRuntimeNodeRepository
from src.domain.repositories.template_repository import ITemplateRepository
from src.domain.value_objects.execution_request import ExecutionRequest
from src.infrastructure.container_scheduler.base import (
    IContainerScheduler,
    ContainerConfig,
)
from src.infrastructure.executors import ExecutorClient
from src.infrastructure.logging import get_logger

logger = get_logger(__name__)


class DockerSchedulerService(IScheduler):
    """
    Docker 调度服务

    实现调度策略：
    1. 优先选择有模板亲和性的节点（镜像已缓存）
    2. 选择负载最低的健康节点

    容器从创建时就绑定到会话，生命周期完全跟随会话。
    """

    def __init__(
        self,
        runtime_node_repo: IRuntimeNodeRepository,
        container_scheduler: IContainerScheduler,
        template_repo: ITemplateRepository,
        executor_client: Optional[ExecutorClient] = None,
        executor_port: int = 8080,
        control_plane_url: str = "http://control-plane:8000",
        disable_bwrap: bool = False,
    ):
        self._runtime_node_repo = runtime_node_repo
        self._container_scheduler = container_scheduler
        self._template_repo = template_repo
        self._executor_client = executor_client or ExecutorClient()
        self._executor_port = executor_port
        self._control_plane_url = control_plane_url
        self._disable_bwrap = disable_bwrap

    async def schedule(self, request: ScheduleRequest) -> RuntimeNode:
        """
        调度会话到最优节点

        调度策略：
        1. 检查是否有已缓存该模板的节点（模板亲和性）
        2. 选择负载最低的健康节点
        """
        logger.info(
            "Starting node selection",
            session_id=request.session_id,
            template_id=request.template_id,
            cpu_limit=request.resource_limit.cpu,
            memory_limit=request.resource_limit.memory,
        )

        # 1. 获取所有健康节点
        healthy_nodes = await self.get_healthy_nodes()
        logger.debug(
            "Found healthy nodes",
            node_count=len(healthy_nodes),
            nodes=[n.id for n in healthy_nodes],
        )

        if not healthy_nodes:
            logger.error("No healthy runtime nodes available")
            raise RuntimeError("No healthy runtime nodes available")

        # 2. 按模板亲和性排序
        affinity_nodes = [node for node in healthy_nodes if node.has_template(request.template_id)]

        logger.debug(
            "Node affinity check",
            template_id=request.template_id,
            affinity_nodes_count=len(affinity_nodes),
            affinity_node_ids=[n.id for n in affinity_nodes],
        )

        if affinity_nodes:
            # 选择亲和节点中负载最低的
            selected = self._select_least_loaded(affinity_nodes)
            logger.info(
                "Selected affinity node",
                node_id=selected.id,
                session_id=request.session_id,
                reason="template_cached",
                node_load=selected.get_load_ratio(),
                node_sessions=selected.session_count,
            )
            return selected

        # 3. 使用负载均衡选择节点
        selected = self._select_least_loaded(healthy_nodes)
        logger.info(
            "Selected node by load balancing",
            node_id=selected.id,
            session_id=request.session_id,
            reason="load_balancing",
            node_load=selected.get_load_ratio(),
            node_sessions=selected.session_count,
        )
        return selected

    async def get_node(self, node_id: str) -> Optional[RuntimeNode]:
        """获取指定节点"""
        node_model = await self._runtime_node_repo.find_by_id(node_id)
        if not node_model:
            return None
        return node_model.to_runtime_node()

    async def get_healthy_nodes(self) -> List[RuntimeNode]:
        """获取所有健康节点"""
        nodes = await self._runtime_node_repo.find_by_status("online")
        return [node.to_runtime_node() for node in nodes]

    async def mark_node_unhealthy(self, node_id: str) -> None:
        """标记节点为不健康"""
        await self._runtime_node_repo.update_status(node_id, "offline")
        logger.warning("Marked node as unhealthy", node_id=node_id)

    def _select_least_loaded(self, nodes: List[RuntimeNode]) -> RuntimeNode:
        """
        从节点列表中选择负载最低的节点

        选择逻辑：
        1. 负载比率最低
        2. 如果比率相同，选择会话数最少的
        """
        return min(nodes, key=lambda n: (n.get_load_ratio(), n.session_count))

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
        为会话创建容器（同步）

        容器从创建时就绑定到会话。

        Args:
            session_id: 会话 ID
            template_id: 模板 ID
            image: 容器镜像
            resource_limit: 资源限制
            env_vars: 环境变量
            workspace_path: 工作空间路径
            node_id: 目标节点 ID
            dependencies: Python 依赖列表（pip 规范）[新增]

        Returns:
            容器ID（使用容器名称作为 ID）
        """
        import json

        logger.info(
            "Creating container for session",
            session_id=session_id,
            template_id=template_id,
            image=image,
            node_id=node_id,
            dependencies_count=len(dependencies) if dependencies else 0,
            dependencies=dependencies,
        )

        # 获取节点信息
        node = await self.get_node(node_id)
        if not node:
            logger.error("Node not found", node_id=node_id, session_id=session_id)
            raise RuntimeError(f"Node not found: {node_id}")

        logger.debug(
            "Node retrieved",
            node_id=node.id,
            node_type=node.type,
            node_status=node.status,
        )

        # 创建容器配置
        # dependencies_json 传递给 docker_scheduler.py 用于动态生成 entrypoint 脚本
        dependencies_json = json.dumps(dependencies) if dependencies else ""

        container_name = f"sandbox-{session_id}"

        config = ContainerConfig(
            image=image,
            name=container_name,
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
                "dependencies": dependencies_json,  # 传递给 docker_scheduler.py
            },
        )

        logger.debug(
            "Container configuration prepared",
            session_id=session_id,
            container_name=container_name,
            cpu_limit=config.cpu_limit,
            memory_limit=config.memory_limit,
            disk_limit=config.disk_limit,
            workspace_path=workspace_path,
            env_vars_count=len(config.env_vars),
        )

        # 同步创建容器（等待完成）
        try:
            logger.info(
                "Creating container",
                session_id=session_id,
                container_name=container_name,
                image=image,
            )

            container_id = await self._container_scheduler.create_container(config)

            logger.info(
                "Container created, starting now",
                session_id=session_id,
                container_id=container_id,
                container_name=container_name,
            )

            await self._container_scheduler.start_container(container_id)

            logger.info(
                "Container started successfully",
                session_id=session_id,
                container_id=container_id,
                container_name=container_name,
                node_id=node.id,
            )

            # 获取容器状态确认
            try:
                container_info = await self._container_scheduler.get_container_status(
                    container_name
                )
                logger.info(
                    "Container status after start",
                    session_id=session_id,
                    container_id=container_id,
                    container_name=container_name,
                    status=container_info.status,
                    ip_address=container_info.ip_address,
                )
            except Exception as status_error:
                logger.warning(
                    "Failed to get container status after start",
                    session_id=session_id,
                    container_id=container_id,
                    error=str(status_error),
                )

            # 使用容器名称作为 ID（用于执行器通信）
            return container_name

        except Exception as e:
            logger.exception(
                "Failed to create/start container",
                session_id=session_id,
                container_name=container_name,
                error=str(e),
                error_type=type(e).__name__,
            )
            raise

    async def destroy_container(self, container_id: str, timeout: int = 10) -> None:
        """
        销毁容器

        容器始终被销毁，不再有释放到预热池的逻辑。
        """
        try:
            await self._container_scheduler.stop_container(container_id, timeout=timeout)
            await self._container_scheduler.remove_container(container_id)
            logger.info("Destroyed container", container_id=container_id)
        except Exception as e:
            logger.error(
                "Failed to destroy container",
                container_id=container_id,
                error=str(e),
            )
            raise

    async def get_container_info(self, container_id: str):
        """获取容器信息"""
        return await self._container_scheduler.get_container_status(container_id)

    async def execute(
        self,
        session_id: str,
        container_id: str,
        execution_request: ExecutionRequest,
    ) -> str:
        """
        提交执行请求到容器内的执行器

        通过 HTTP 与运行在容器内的 sandbox-executor 通信。

        Args:
            session_id: 会话 ID
            container_id: 容器 ID
            execution_request: 执行请求

        Returns:
            execution_id: 执行任务 ID

        Raises:
            ConnectionError: 无法连接到执行器
            TimeoutError: 执行器响应超时
        """
        # 获取容器信息以构建执行器 URL
        container_info = await self._container_scheduler.get_container_status(container_id)

        # 构建执行器 URL
        # 使用容器名称在 Docker 内部网络中进行通信
        # 容器名称格式: sandbox-{session_id}
        container_name = container_info.name
        executor_url = self._build_executor_url(container_name)

        logger.info(
            "Submitting execution to executor",
            executor_url=executor_url,
            session_id=session_id,
            container_id=container_id,
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
                "Execution submitted successfully",
                execution_id=execution_id,
                session_id=session_id,
            )

            return execution_id

        except Exception as e:
            logger.error(
                "Failed to submit execution to executor",
                executor_url=executor_url,
                error=str(e),
            )
            raise

    async def get_executor_url(self, container_id: str) -> str:
        """根据容器 ID 获取 executor URL。"""
        return self._build_executor_url(container_id)

    def _build_executor_url(self, container_name: str) -> str:
        return f"http://{container_name}:{self._executor_port}"
