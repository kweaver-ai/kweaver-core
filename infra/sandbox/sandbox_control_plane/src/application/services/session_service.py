"""
会话应用服务

编排会话相关的用例。
"""
from typing import Callable, List, Optional
from datetime import datetime, timedelta
import uuid

from src.domain.entities.session import InstalledDependency, Session
from src.domain.entities.execution import Execution
from src.domain.entities.template import Template
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus, ExecutionStatus
from src.domain.value_objects.execution_request import ExecutionRequest
from src.domain.repositories.session_repository import ISessionRepository
from src.domain.repositories.execution_repository import IExecutionRepository
from src.domain.repositories.template_repository import ITemplateRepository
from src.domain.services.scheduler import IScheduler, ScheduleRequest, RuntimeNode
from src.domain.services.storage import IStorageService
from src.application.commands.create_session import CreateSessionCommand
from src.application.commands.install_session_dependencies import (
    InstallSessionDependenciesCommand,
)
from src.infrastructure.config.settings import get_settings
from src.application.commands.execute_code import ExecuteCodeCommand
from src.application.queries.get_session import GetSessionQuery
from src.application.queries.get_execution import GetExecutionQuery
from src.application.dtos.session_dto import SessionDTO
from src.application.dtos.execution_dto import ExecutionDTO
from src.shared.errors.domain import NotFoundError, ValidationError, ConflictError
from src.infrastructure.executors import ExecutorClient
from src.infrastructure.executors.errors import (
    ExecutorConnectionError,
    ExecutorResponseError,
    ExecutorTimeoutError,
    ExecutorUnavailableError,
    ExecutorValidationError,
)
from src.infrastructure.logging import get_logger
from src.shared.utils.dependencies import (
    DEFAULT_PYTHON_PACKAGE_INDEX_URL,
    normalize_python_package_index_url,
)

logger = get_logger(__name__)


class SessionService:
    """
    会话应用服务

    编排会话创建、执行、终止等用例。
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        execution_repo: IExecutionRepository,
        template_repo: ITemplateRepository,
        scheduler: IScheduler,
        storage_service: Optional[IStorageService] = None,
        executor_client: Optional[ExecutorClient] = None,
        initial_dependency_sync_scheduler: Optional[Callable[[str, int], None]] = None,
    ):
        self._session_repo = session_repo
        self._execution_repo = execution_repo
        self._template_repo = template_repo
        self._scheduler = scheduler
        self._storage_service = storage_service
        self._executor_client = executor_client or ExecutorClient()
        self._initial_dependency_sync_scheduler = initial_dependency_sync_scheduler

    async def create_session(self, command: CreateSessionCommand) -> SessionDTO:
        """
        创建会话用例

        流程：
        1. 验证模板存在
        2. 生成会话 ID
        3. 调用调度器选择运行时节点
        4. 创建会话实体
        5. 保存到仓储
        6. 创建 Docker 容器
        7. 更新会话状态为 running
        """
        if not command.template_id:
            settings = get_settings()
            command.template_id = settings.default_template_id

        logger.info(
            "Creating session",
            template_id=command.template_id,
            has_dependencies=len(command.dependencies or []) > 0,
        )

        # 1. 验证模板
        template = await self._validate_template(command.template_id)

        # 2. 处理会话 ID（手动指定或自动生成）
        if command.id:
            # 手动指定 ID，检查冲突
            session_id = command.id
            existing_session = await self._session_repo.find_by_id(session_id)
            if existing_session:
                logger.warning(
                    "Session ID already exists",
                    session_id=session_id,
                    existing_status=existing_session.status.value,
                )
                raise ConflictError(f"Session ID already exists: {session_id}")
            logger.debug("Using manually specified session ID", session_id=session_id)
        else:
            # 自动生成会话 ID
            session_id = self._generate_session_id()
            logger.debug("Generated session ID", session_id=session_id)

        # 3. 调用调度器
        runtime_node = await self._schedule_session(command, session_id)

        # 4. 创建会话实体
        session = self._create_session_entity(
            session_id=session_id,
            command=command,
            template=template,
            runtime_node=runtime_node,
        )

        # 5. 保存到仓储
        await self._session_repo.save(session)
        logger.debug("Session saved to repository", session_id=session_id)

        # 6. 创建容器
        container_id = await self._create_container_for_session(
            session=session,
            template=template,
            command=command,
            runtime_node=runtime_node,
        )

        logger.info(
            "Session created successfully",
            session_id=session_id,
            container_id=container_id,
            status=session.status.value,
        )

        if dependencies := (command.dependencies or []):
            session.mark_dependency_installing()
            await self._session_repo.save(session)
            self._schedule_initial_dependency_sync(
                session_id=session.id,
                install_timeout=command.install_timeout,
                dependency_count=len(dependencies),
            )

        return SessionDTO.from_entity(session)

    async def _validate_template(self, template_id: str) -> Template:
        """验证模板存在"""
        from src.domain.entities.template import Template

        template = await self._template_repo.find_by_id(template_id)
        if not template:
            logger.error("Template not found", template_id=template_id)
            raise NotFoundError(f"Template not found: {template_id}")

        logger.debug("Template validated", template_id=template.id, image=template.image)
        return template

    async def _schedule_session(self, command: CreateSessionCommand, session_id: str) -> RuntimeNode:
        """调度会话到运行时节点"""
        schedule_request = ScheduleRequest(
            template_id=command.template_id,
            resource_limit=command.resource_limit or ResourceLimit.default(),
            session_id=session_id
        )
        runtime_node = await self._scheduler.schedule(schedule_request)

        logger.info(
            "Runtime node selected",
            session_id=session_id,
            runtime_node=runtime_node.id,
            node_type=runtime_node.type,
        )
        return runtime_node

    def _create_session_entity(
        self,
        session_id: str,
        command: CreateSessionCommand,
        template,
        runtime_node: RuntimeNode,
    ) -> Session:
        """创建会话实体"""
        from src.domain.entities.template import Template

        runtime_type = self._infer_runtime_type(template.image)
        resource_limit = command.resource_limit or ResourceLimit.default()
        settings = get_settings()
        workspace_path = f"s3://{settings.s3_bucket}/sessions/{session_id}"
        dependencies = command.dependencies or []

        return Session(
            id=session_id,
            template_id=command.template_id,
            status=SessionStatus.CREATING,
            resource_limit=resource_limit,
            workspace_path=workspace_path,
            runtime_type=runtime_type,
            runtime_node=runtime_node.id,
            env_vars=command.env_vars or {},
            timeout=command.timeout,
            python_package_index_url=normalize_python_package_index_url(
                command.python_package_index_url
            ),
            requested_dependencies=dependencies,
            dependency_install_status="pending" if dependencies else "completed",
        )

    async def _create_container_for_session(
        self,
        session: Session,
        template,
        command: CreateSessionCommand,
        runtime_node: RuntimeNode,
    ) -> Optional[str]:
        """为会话创建容器"""
        from src.domain.entities.template import Template

        container_id = None
        dependencies: list[str] = []

        try:
            if hasattr(self._scheduler, 'create_container_for_session'):
                logger.info(
                    "Creating container for session",
                    session_id=session.id,
                    image=template.image,
                    dependencies_count=len(dependencies),
                    dependencies=dependencies,
                    runtime_node_id=runtime_node.id,
                    runtime_node_type=runtime_node.type,
                )

                container_id = await self._scheduler.create_container_for_session(
                    session_id=session.id,
                    template_id=command.template_id,
                    image=template.image,
                    resource_limit=session.resource_limit,
                    env_vars=session.env_vars,
                    workspace_path=session.workspace_path,
                    node_id=runtime_node.id,
                    dependencies=dependencies,
                )

                session.container_id = container_id
                await self._session_repo.save(session)

                logger.info(
                    "Container created successfully, session saved",
                    session_id=session.id,
                    container_id=container_id,
                    runtime_node=runtime_node.id,
                    dependencies_count=len(dependencies),
                    session_status=session.status.value,
                )
            else:
                logger.warning(
                    "Scheduler does not support create_container_for_session",
                    scheduler_type=type(self._scheduler).__name__,
                )
        except Exception as e:
            logger.exception(
                "Exception during container creation",
                session_id=session.id,
                error_type=type(e).__name__,
                error=str(e),
            )
            await self._handle_container_creation_failure(
                session=session,
                container_id=container_id,
                error=e,
            )

        return container_id

    async def _handle_container_creation_failure(
        self,
        session: Session,
        container_id: Optional[str],
        error: Exception,
    ) -> None:
        """处理容器创建失败"""
        logger.exception(
            "Container creation failed, starting cleanup",
            session_id=session.id,
            container_id=container_id,
            error_type=type(error).__name__,
            error=str(error),
        )

        # 清理已创建的容器
        if container_id and hasattr(self._scheduler, 'destroy_container'):
            try:
                logger.info("Attempting to clean up failed container", container_id=container_id)
                await self._scheduler.destroy_container(container_id)
                logger.debug("Cleaned up failed container", container_id=container_id)
            except Exception as cleanup_error:
                logger.warning(
                    "Failed to cleanup container",
                    container_id=container_id,
                    cleanup_error=str(cleanup_error),
                )

        # 标记会话为失败状态
        session.status = SessionStatus.FAILED
        if session.has_dependencies():
            session.set_dependencies_failed(str(error))
        await self._session_repo.save(session)

        logger.error(
            "Session creation failed",
            session_id=session.id,
            final_status=session.status.value,
            container_id=container_id,
        )
        raise ValidationError(f"Failed to create container: {error}")

    async def get_session(self, query: GetSessionQuery) -> SessionDTO:
        """获取会话用例"""
        session = await self._session_repo.find_by_id(query.session_id)
        if not session:
            raise NotFoundError(f"Session not found: {query.session_id}")

        return SessionDTO.from_entity(session)

    async def install_session_dependencies(
        self,
        command: InstallSessionDependenciesCommand,
    ) -> SessionDTO:
        """增量安装会话依赖。"""
        session = await self._session_repo.find_by_id(command.session_id)
        if not session:
            raise NotFoundError(f"Session not found: {command.session_id}")

        if session.dependency_install_status == "installing":
            raise ConflictError(
                f"Dependency installation already in progress for session: {session.id}"
            )

        session.merge_requested_dependencies(
            command.python_package_index_url,
            command.dependencies,
        )
        return await self._sync_session_dependencies(session, sync_mode="merge")

    async def sync_session_dependencies_for_session(
        self,
        session_id: str,
        sync_mode: str = "replace",
    ) -> SessionDTO:
        """同步指定 session 的依赖配置。"""
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            raise NotFoundError(f"Session not found: {session_id}")
        return await self._sync_session_dependencies(session, sync_mode=sync_mode)

    async def list_sessions(
        self,
        status: Optional[str] = None,
        template_id: Optional[str] = None,
        limit: int = 50,
        offset: int = 0
    ) -> dict:
        """
        列出会话用例

        Args:
            status: 会话状态筛选（可选）
            template_id: 模板 ID 筛选（可选）
            limit: 返回数量限制（1-200，默认 50）
            offset: 偏移量（用于分页）

        Returns:
            包含 items, total, limit, offset, has_more 的字典
        """
        # 验证 limit 范围
        limit = max(1, min(limit, 200))
        offset = max(0, offset)

        # 获取会话列表
        sessions = await self._session_repo.find_sessions(
            status=status,
            template_id=template_id,
            limit=limit,
            offset=offset
        )

        # 获取总数
        total = await self._session_repo.count_sessions(
            status=status,
            template_id=template_id
        )

        # 转换为 DTO
        items = [SessionDTO.from_entity(s) for s in sessions]

        # 计算是否有更多数据
        has_more = (offset + len(items)) < total

        return {
            "items": items,
            "total": total,
            "limit": limit,
            "offset": offset,
            "has_more": has_more
        }

    async def terminate_session(self, session_id: str) -> SessionDTO:
        """
        终止会话用例（软终止，保留记录）

        流程：
        1. 查找会话
        2. 验证状态
        3. 销毁 Docker 容器（如果调度器支持）
        4. 清理 S3 文件（如果配置了存储服务）
        5. 更新会话状态
        """
        logger.info("Terminating session", session_id=session_id)

        session = await self._session_repo.find_by_id(session_id)
        if not session:
            logger.warning("Session not found for termination", session_id=session_id)
            raise NotFoundError(f"Session not found: {session_id}")

        if session.is_terminated():
            logger.info("Session already terminated", session_id=session_id, status=session.status.value)
            return SessionDTO.from_entity(session)

        logger.debug(
            "Terminating active session",
            session_id=session_id,
            container_id=session.container_id,
            status=session.status.value,
        )

        # 销毁容器
        await self._destroy_container(session)

        # 清理 S3 文件
        await self._cleanup_storage(session)

        # 更新会话状态
        session.mark_as_terminated()
        await self._session_repo.save(session)

        logger.info("Session terminated successfully", session_id=session_id, final_status=session.status.value)

        return SessionDTO.from_entity(session)

    async def delete_session(self, session_id: str) -> None:
        """
        删除会话用例（硬删除，级联删除执行记录）

        流程：
        1. 查找会话
        2. 执行清理（销毁容器 + 删除 S3）
        3. 级联删除数据库记录
        """
        logger.info("Deleting session", session_id=session_id)

        session = await self._session_repo.find_by_id(session_id)
        if not session:
            logger.warning("Session not found for deletion", session_id=session_id)
            raise NotFoundError(f"Session not found: {session_id}")

        logger.debug(
            "Deleting session",
            session_id=session_id,
            container_id=session.container_id,
            status=session.status.value,
        )

        # 销毁容器
        await self._destroy_container(session)

        # 清理 S3 文件
        await self._cleanup_storage(session)

        # 级联删除数据库记录（session + executions）
        await self._session_repo.delete(session_id)

        logger.info("Session deleted successfully", session_id=session_id)

    async def _destroy_container(self, session: Session) -> None:
        """销毁会话的容器"""
        if not session.container_id or not hasattr(self._scheduler, 'destroy_container'):
            return

        try:
            logger.info("Destroying container", session_id=session.id, container_id=session.container_id)
            await self._scheduler.destroy_container(container_id=session.container_id)
            logger.info("Container destroyed successfully", session_id=session.id, container_id=session.container_id)
        except Exception as e:
            logger.warning(
                "Failed to destroy container",
                session_id=session.id,
                container_id=session.container_id,
                error=str(e),
            )

    async def _cleanup_storage(self, session: Session) -> None:
        """清理会话的存储文件"""
        if not self._storage_service or not session.workspace_path.startswith("s3://"):
            return

        try:
            logger.info(
                "Cleaning up S3 workspace files",
                session_id=session.id,
                workspace_path=session.workspace_path,
            )
            deleted_count = await self._storage_service.delete_prefix(session.workspace_path)
            logger.info(
                "S3 files deleted",
                session_id=session.id,
                deleted_count=deleted_count,
                workspace_path=session.workspace_path,
            )
        except Exception as e:
            logger.warning(
                "Failed to cleanup S3 files",
                session_id=session.id,
                workspace_path=session.workspace_path,
                error=str(e),
            )

    async def execute_code(self, command: ExecuteCodeCommand) -> ExecutionDTO:
        """
        执行代码用例

        流程：
        1. 验证会话存在且运行中
        2. 生成执行 ID
        3. 创建执行实体
        4. 保存到仓储
        5. 提交到执行器
        """
        logger.info(
            "Executing code",
            session_id=command.session_id,
            language=command.language,
            code_length=len(command.code),
        )

        # 1. 验证会话
        session = await self._session_repo.find_by_id(command.session_id)
        if not session:
            logger.error(
                "Session not found for execution",
                session_id=command.session_id,
            )
            raise NotFoundError(f"Session not found: {command.session_id}")

        if not session.is_active():
            logger.warning(
                "Session is not active",
                session_id=command.session_id,
                status=session.status.value,
            )
            raise ValidationError(f"Session is not active: {command.session_id}")

        logger.debug(
            "Session validated for execution",
            session_id=command.session_id,
            container_id=session.container_id,
        )

        # 2. 生成执行 ID
        execution_id = self._generate_execution_id()

        logger.debug(
            "Generated execution ID",
            execution_id=execution_id,
            session_id=command.session_id,
        )

        # 3. 创建执行实体
        from src.domain.value_objects.execution_status import ExecutionState

        execution = Execution(
            id=execution_id,
            session_id=command.session_id,
            code=command.code,
            language=command.language,
            timeout=command.timeout,
            event_data=command.event_data or {},
            state=ExecutionState(status=ExecutionStatus.PENDING)
        )

        # 4. 保存到仓储
        await self._execution_repo.save(execution)
        logger.debug(
            "Execution saved to repository",
            execution_id=execution_id,
        )

        # 4.5. 提交事务，确保执行记录在执行器回调之前可见
        await self._execution_repo.commit()

        # 5. 提交到执行器
        if not session.container_id:
            logger.error(
                "Session has no container",
                session_id=command.session_id,
            )
            raise ValidationError(f"Session has no container: {command.session_id}")

        # 构建执行请求
        execution_request = ExecutionRequest(
            code=command.code,
            language=command.language,
            event=command.event_data or {},
            timeout=command.timeout or 300,
            env_vars=session.env_vars,
            execution_id=execution_id,
            session_id=session.id,
            working_directory=command.working_directory,
        )

        logger.info(
            "Submitting execution to executor",
            execution_id=execution_id,
            session_id=command.session_id,
            container_id=session.container_id,
            timeout=execution_request.timeout,
        )

        # 通过调度器提交到执行器
        await self._scheduler.execute(
            session_id=session.id,
            container_id=session.container_id,
            execution_request=execution_request,
        )

        logger.info(
            "Execution submitted successfully",
            execution_id=execution_id,
            session_id=command.session_id,
        )

        return ExecutionDTO.from_entity(execution)

    async def get_execution(self, query: GetExecutionQuery) -> ExecutionDTO:
        """获取执行详情用例"""
        execution = await self._execution_repo.find_by_id(query.execution_id)
        if not execution:
            raise NotFoundError(f"Execution not found: {query.execution_id}")

        return ExecutionDTO.from_entity(execution)

    async def list_executions(
        self,
        session_id: str,
        limit: int = 50,
        offset: int = 0
    ) -> List[ExecutionDTO]:
        """列出会话的所有执行用例"""
        executions = await self._execution_repo.find_by_session_id(
            session_id=session_id,
            limit=limit
        )

        return [ExecutionDTO.from_entity(e) for e in executions]

    async def cleanup_idle_sessions(
        self,
        idle_threshold_minutes: int = 30,
        max_lifetime_hours: int = 6
    ) -> int:
        """
        清理空闲会话用例

        定时任务调用，清理空闲或过期的会话。
        """
        idle_threshold = datetime.now() - timedelta(minutes=idle_threshold_minutes)
        max_lifetime = datetime.now() - timedelta(hours=max_lifetime_hours)

        idle_sessions = await self._session_repo.find_idle_sessions(idle_threshold)
        expired_sessions = await self._session_repo.find_expired_sessions(max_lifetime)

        all_to_cleanup = set(idle_sessions + expired_sessions)
        cleaned_count = 0

        for session in all_to_cleanup:
            if await self._cleanup_session(session):
                cleaned_count += 1

        return cleaned_count

    async def _cleanup_session(self, session: Session) -> bool:
        """清理单个会话"""
        if not session.is_active():
            return False

        # 销毁容器
        if session.container_id and hasattr(self._scheduler, 'destroy_container'):
            try:
                await self._scheduler.destroy_container(container_id=session.container_id)
            except Exception as e:
                logger.warning(
                    "Failed to destroy container during cleanup",
                    session_id=session.id,
                    container_id=session.container_id,
                    error=str(e),
                )

        session.mark_as_terminated()
        await self._session_repo.save(session)
        return True

    async def _sync_session_dependencies(
        self,
        session: Session,
        sync_mode: str,
    ) -> SessionDTO:
        """同步 session 依赖配置到 executor。"""
        if not session.is_active():
            raise ValidationError(f"Session is not active: {session.id}")
        if not session.container_id:
            raise ValidationError(f"Session has no container: {session.id}")
        if not hasattr(self._scheduler, "get_executor_url"):
            raise ValidationError("Scheduler does not support executor URL discovery")

        session.mark_dependency_installing()
        await self._session_repo.save(session)

        try:
            executor_url = await self._scheduler.get_executor_url(session.container_id)
            result = await self._executor_client.sync_session_config(
                executor_url=executor_url,
                session_id=session.id,
                language_runtime=session.runtime_type,
                python_package_index_url=session.python_package_index_url,
                dependencies=session.requested_dependencies,
                sync_mode=sync_mode,
            )
        except (
            ExecutorConnectionError,
            ExecutorTimeoutError,
            ExecutorUnavailableError,
            ExecutorValidationError,
            ExecutorResponseError,
        ) as error:
            session.mark_dependency_install_failed(str(error))
            await self._session_repo.save(session)
            raise

        installed_dependencies = [
            InstalledDependency(
                name=dep.name,
                version=dep.version,
                install_location=dep.install_location,
                install_time=datetime.fromisoformat(dep.install_time.replace("Z", "+00:00")),
                is_from_template=dep.is_from_template,
            )
            for dep in result.installed_dependencies
        ]

        completed_at = None
        if result.completed_at:
            completed_at = datetime.fromisoformat(result.completed_at.replace("Z", "+00:00"))

        session.mark_dependency_install_completed(
            installed_dependencies,
            completed_at=completed_at,
        )
        if result.started_at:
            session.dependency_install_started_at = datetime.fromisoformat(
                result.started_at.replace("Z", "+00:00")
            )
        await self._session_repo.save(session)
        return SessionDTO.from_entity(session)

    def _schedule_initial_dependency_sync(
        self,
        session_id: str,
        install_timeout: int,
        dependency_count: int,
    ) -> None:
        """调度首次依赖安装后台任务。"""
        if self._initial_dependency_sync_scheduler is None:
            logger.warning(
                "Initial dependency sync scheduler is not configured",
                session_id=session_id,
                dependency_count=dependency_count,
            )
            return

        logger.info(
            "Scheduling initial dependency sync",
            session_id=session_id,
            dependency_count=dependency_count,
            install_timeout=install_timeout,
        )
        self._initial_dependency_sync_scheduler(session_id, install_timeout)

    def _generate_session_id(self) -> str:
        """生成会话 ID"""
        timestamp = datetime.now().strftime("%Y%m%d")
        unique = uuid.uuid4().hex[:8]
        return f"sess_{timestamp}_{unique}"

    def _infer_runtime_type(self, image: str) -> str:
        """从镜像名称推断运行时类型"""
        image_lower = image.lower()
        if "python" in image_lower or "python3" in image_lower:
            return "python3.11"
        elif "node" in image_lower or "nodejs" in image_lower:
            return "nodejs20"
        elif "java" in image_lower:
            return "java17"
        elif "go" in image_lower or "golang" in image_lower:
            return "go1.21"
        else:
            # 默认使用 Python
            return "python3.11"

    def _generate_execution_id(self) -> str:
        """生成执行 ID"""
        timestamp = datetime.now().strftime("%Y%m%d%H%M%S")
        unique = uuid.uuid4().hex[:8]
        return f"exec_{timestamp}_{unique}"
