"""
状态同步服务

负责同步 Session 状态与实际容器状态，支持启动时同步和定时健康检查。
"""
from typing import Dict, Optional

from src.domain.entities.session import Session, SessionStatus
from src.domain.repositories.session_repository import ISessionRepository
from src.domain.repositories.template_repository import ITemplateRepository
from src.infrastructure.container_scheduler.base import IContainerScheduler
from src.infrastructure.logging import get_logger

logger = get_logger(__name__)


class StateSyncService:
    """
    状态同步服务

    职责：
    1. 启动时全量状态同步
    2. 定时健康检查（通过 Docker/K8s API）
    3. 状态不一致时更新 Session 表
    4. 恢复不健康的容器（创建新容器）

    核心原则：Docker/K8s 是容器状态的唯一真实来源，Session 表只保存关联关系。
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        container_scheduler: IContainerScheduler,
        template_repo: Optional[ITemplateRepository] = None,
        scheduler=None,
        control_plane_url: str = "http://control-plane:8000",
    ):
        self._session_repo = session_repo
        self._container_scheduler = container_scheduler
        self._template_repo = template_repo
        self._scheduler = scheduler
        self._control_plane_url = control_plane_url

    async def sync_on_startup(self) -> Dict[str, int]:
        """
        启动时全量同步

        查询所有 RUNNING/CREATING 状态的 Session，检查容器实际状态，
        尝试恢复不健康的容器或标记为失败。
        """
        logger.info("Starting state synchronization on startup")

        stats = {
            "total": 0,
            "healthy": 0,
            "unhealthy": 0,
            "recovered": 0,
            "failed": 0,
            "errors": []
        }

        try:
            running_sessions = await self._session_repo.find_by_status("running")
            creating_sessions = await self._session_repo.find_by_status("creating")
            active_sessions = running_sessions + creating_sessions

            stats["total"] = len(active_sessions)
            logger.info("Found active sessions to sync", count=len(active_sessions))

            for session in active_sessions:
                if not session.container_id:
                    logger.warning("Session has no container_id, skipping", session_id=session.id)
                    continue

                await self._check_and_recover_session(session, stats)

            logger.info(
                "State sync completed",
                total=stats["total"],
                healthy=stats["healthy"],
                unhealthy=stats["unhealthy"],
                recovered=stats["recovered"],
                failed=stats["failed"],
            )

        except Exception as e:
            logger.error("Fatal error during state sync", error=str(e), exc_info=True)
            stats["errors"].append(f"Fatal error: {e}")

        return stats

    async def periodic_health_check(self) -> Dict[str, int]:
        """
        定时健康检查（每 30 秒）

        只检查 RUNNING 状态的 Session，减少查询范围。
        对不健康的容器尝试恢复。
        """
        logger.info("Starting periodic health check")

        stats = {
            "checked": 0,
            "healthy": 0,
            "unhealthy": 0,
            "recovered": 0,
            "failed": 0,
            "errors": []
        }

        try:
            running_sessions = await self._session_repo.find_by_status("running")
            logger.info("Checking running sessions", count=len(running_sessions))

            for session in running_sessions:
                if not session.container_id:
                    continue

                stats["checked"] += 1
                await self._check_and_recover_session(session, stats)

            if stats["checked"] > 0:
                logger.info(
                    "Health check completed",
                    checked=stats["checked"],
                    healthy=stats["healthy"],
                    unhealthy=stats["unhealthy"],
                    recovered=stats["recovered"],
                    failed=stats["failed"],
                )

        except Exception as e:
            logger.error("Fatal error during health check", error=str(e), exc_info=True)
            stats["errors"].append(f"Fatal error: {e}")

        return stats

    async def _check_and_recover_session(self, session: Session, stats: Dict[str, int]) -> None:
        """检查单个会话的健康状态并尝试恢复"""
        try:
            is_running = await self._container_scheduler.is_container_running(session.container_id)

            if is_running:
                stats["healthy"] += 1
                logger.debug(
                    "Session container is healthy",
                    session_id=session.id,
                    container_id=session.container_id[:12],
                )
            else:
                stats["unhealthy"] += 1
                logger.warning(
                    "Session container is unhealthy",
                    session_id=session.id,
                    container_id=session.container_id[:12],
                )

                recovered = await self._attempt_recovery(session)
                if recovered:
                    stats["recovered"] += 1
                else:
                    stats["failed"] += 1

        except Exception as e:
            error_msg = f"Error checking session {session.id}: {e}"
            logger.error(error_msg, exc_info=True)
            stats["errors"].append(error_msg)

    async def _attempt_recovery(self, session: Session) -> bool:
        """
        尝试恢复 Session

        策略：创建新容器（不再使用预热池）
        """
        logger.info("Attempting recovery for session", session_id=session.id)

        try:
            from src.infrastructure.container_scheduler.base import ContainerConfig

            image = await self._get_recovery_image(session)

            config = ContainerConfig(
                image=image,
                name=f"sandbox-{session.id}",
                env_vars={
                    **(session.env_vars or {}),
                    "SESSION_ID": session.id,
                    "WORKSPACE_PATH": session.workspace_path,
                    "CONTROL_PLANE_URL": self._control_plane_url,
                    "DISABLE_BWRAP": "true",
                },
                cpu_limit=session.resource_limit.cpu if session.resource_limit else "1",
                memory_limit=session.resource_limit.memory if session.resource_limit else "512Mi",
                disk_limit=session.resource_limit.disk if session.resource_limit else "1Gi",
                workspace_path=session.workspace_path,
                labels={
                    "session_id": session.id,
                    "template_id": session.template_id,
                    "recovered": "true",
                },
            )

            container_id = await self._container_scheduler.create_container(config)
            await self._container_scheduler.start_container(container_id)

            session.container_id = container_id
            session.runtime_node = session.runtime_node or self._get_recovery_runtime_node()
            session.status = SessionStatus.RUNNING
            await self._session_repo.save(session)

            logger.info(
                "Session recovered successfully",
                session_id=session.id,
                container_id=container_id[:12],
            )
            return True

        except Exception as e:
            logger.error(
                "Failed to recover session",
                session_id=session.id,
                error=str(e),
                exc_info=True,
            )

            try:
                session.mark_as_failed()
                await self._session_repo.save(session)
            except Exception as save_error:
                logger.error(
                    "Failed to mark session as failed",
                    session_id=session.id,
                    error=str(save_error),
                )

            return False

    async def _get_recovery_image(self, session: Session) -> str:
        """解析恢复容器时应使用的镜像。"""
        if self._template_repo is None:
            raise RuntimeError("Template repository is required for session recovery")

        template = await self._template_repo.find_by_id(session.template_id)
        if template is None:
            raise RuntimeError(f"Template not found during recovery: {session.template_id}")

        logger.info(
            "Resolved recovery image from template",
            session_id=session.id,
            template_id=session.template_id,
            image=template.image,
        )
        return template.image

    def _get_recovery_runtime_node(self) -> str:
        """根据当前调度环境推导恢复后的运行时节点。"""
        if self._scheduler is not None and hasattr(self._scheduler, "_cluster_node"):
            cluster_node = getattr(self._scheduler, "_cluster_node", None)
            if cluster_node is not None and getattr(cluster_node, "id", None):
                return cluster_node.id
        return "docker-local"

    async def check_session_health(self, session_id: str) -> Dict[str, any]:
        """
        检查单个 Session 的健康状态
        """
        session = await self._session_repo.find_by_id(session_id)
        if not session:
            return {
                "session_id": session_id,
                "status": "not_found",
                "error": "Session not found"
            }

        if not session.container_id:
            return {
                "session_id": session_id,
                "status": "no_container",
                "error": "Session has no container_id"
            }

        try:
            is_running = await self._container_scheduler.is_container_running(
                session.container_id
            )

            return {
                "session_id": session_id,
                "container_id": session.container_id,
                "container_running": is_running,
                "status": "healthy" if is_running else "unhealthy",
            }

        except Exception as e:
            return {
                "session_id": session_id,
                "container_id": session.container_id,
                "container_running": False,
                "status": "error",
                "error": str(e),
            }
