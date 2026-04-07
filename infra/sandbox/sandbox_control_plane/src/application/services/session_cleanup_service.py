"""
会话清理服务

负责定期清理空闲会话和过期会话，自动销毁关联的容器和删除关联的文件。
"""
import logging
from datetime import datetime, timedelta
from typing import Dict, Optional
from urllib.parse import urlparse

from src.domain.entities.session import Session, SessionStatus
from src.domain.repositories.session_repository import ISessionRepository
from src.domain.services.scheduler import IScheduler
from src.domain.services.storage import IStorageService

logger = logging.getLogger(__name__)


class SessionCleanupService:
    """
    会话清理服务

    职责：
    1. 定期扫描空闲会话（基于 last_activity_at 字段）
    2. 自动终止超时会话并销毁容器
    3. 定期扫描 FAILED/TIMEOUT 状态的孤立会话
    4. 清理会话关联的 S3 文件

    清理策略：
    - 空闲超时：30 分钟无活动（可配置，设为 -1 表示禁用空闲清理）
    - 最大生命周期：6 小时强制清理（可配置，设为 -1 表示禁用生命周期清理）
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        scheduler: IScheduler,
        idle_timeout_minutes: int = 30,
        max_lifetime_hours: int = 6,
        storage_service: Optional[IStorageService] = None,
    ):
        """
        初始化会话清理服务

        Args:
            session_repo: 会话仓储
            scheduler: 调度器（用于销毁容器）
            idle_timeout_minutes: 空闲超时时间（分钟），-1 表示无限期（不清理空闲会话）
            max_lifetime_hours: 最大生命周期（小时），-1 表示无限期
            storage_service: 存储服务（可选，用于清理 S3 文件）
        """
        self._session_repo = session_repo
        self._scheduler = scheduler
        self._storage_service = storage_service
        self._idle_timeout = None if idle_timeout_minutes == -1 else timedelta(minutes=idle_timeout_minutes)
        self._max_lifetime = None if max_lifetime_hours == -1 else timedelta(hours=max_lifetime_hours)

    async def cleanup_idle_sessions(self) -> Dict[str, int]:
        """
        清理空闲会话

        清理策略：
        - 空闲超过阈值的会话自动销毁容器（如果 idle_timeout_minutes != -1）
        - 创建超过最大生命周期的会话强制销毁（如果 max_lifetime_hours != -1）

        Returns:
            dict: 清理统计信息
                - total_checked: 检查的会话数
                - idle_cleaned: 空闲清理的会话数
                - expired_cleaned: 过期清理的会话数
                - errors: 错误列表
        """
        stats = {
            "total_checked": 0,
            "idle_cleaned": 0,
            "expired_cleaned": 0,
            "errors": []
        }

        try:
            now = datetime.now()
            idle_threshold = now - self._idle_timeout if self._idle_timeout else None
            max_lifetime_threshold = now - self._max_lifetime if self._max_lifetime else None

            # 查询所有活跃会话
            active_sessions = await self._session_repo.find_by_status("running")
            stats["total_checked"] = len(active_sessions)

            logger.info(
                f"Starting session cleanup: checking {len(active_sessions)} active sessions, "
                f"idle_threshold={idle_threshold}, max_lifetime={max_lifetime_threshold}"
            )

            for session in active_sessions:
                try:
                    # 检查是否超过最大生命周期（如果启用）
                    if max_lifetime_threshold and session.created_at and session.created_at < max_lifetime_threshold:
                        await self._cleanup_session(
                            session,
                            reason="max_lifetime_exceeded",
                            detail=f"Session created at {session.created_at} exceeded max lifetime of {self._max_lifetime}"
                        )
                        stats["expired_cleaned"] += 1
                        continue

                    # 检查是否空闲超时（如果启用）
                    # 使用 last_activity_at，如果不存在则使用 created_at
                    if idle_threshold:
                        last_activity = session.last_activity_at or session.created_at
                        if last_activity and last_activity < idle_threshold:
                            await self._cleanup_session(
                                session,
                                reason="idle_timeout",
                                detail=f"Session last activity at {last_activity} exceeded idle timeout of {self._idle_timeout}"
                            )
                            stats["idle_cleaned"] += 1
                            continue

                except Exception as e:
                    error_msg = f"Error cleaning session {session.id}: {e}"
                    logger.error(error_msg, exc_info=True)
                    stats["errors"].append(error_msg)

            logger.info(
                f"Session cleanup completed: "
                f"checked={stats['total_checked']}, "
                f"idle_cleaned={stats['idle_cleaned']}, "
                f"expired_cleaned={stats['expired_cleaned']}"
            )

        except Exception as e:
            error_msg = f"Fatal error during session cleanup: {e}"
            logger.error(error_msg, exc_info=True)
            stats["errors"].append(error_msg)

        return stats

    async def cleanup_orphaned_sessions(self) -> Dict[str, int]:
        """
        清理孤立会话（FAILED/TIMEOUT 状态但仍有关联容器的会话）

        Returns:
            dict: 清理统计信息
        """
        stats = {
            "total_checked": 0,
            "cleaned": 0,
            "errors": []
        }

        try:
            # 查询失败和超时的会话
            failed_sessions = await self._session_repo.find_by_status("failed")
            timeout_sessions = await self._session_repo.find_by_status("timeout")
            orphaned = failed_sessions + timeout_sessions

            stats["total_checked"] = len(orphaned)

            for session in orphaned:
                # 只清理有 container_id 的会话
                if session.container_id:
                    try:
                        await self._cleanup_session(
                            session,
                            reason="orphaned_cleanup",
                            detail=f"Session in {session.status} status with container {session.container_id}"
                        )
                        stats["cleaned"] += 1
                    except Exception as e:
                        error_msg = f"Error cleaning orphaned session {session.id}: {e}"
                        logger.error(error_msg, exc_info=True)
                        stats["errors"].append(error_msg)

            logger.info(
                f"Orphaned session cleanup completed: "
                f"checked={stats['total_checked']}, "
                f"cleaned={stats['cleaned']}"
            )

        except Exception as e:
            error_msg = f"Fatal error during orphaned session cleanup: {e}"
            logger.error(error_msg, exc_info=True)
            stats["errors"].append(error_msg)

        return stats

    async def cleanup_session_files(
        self,
        session: Session,
        reason: str
    ) -> int:
        """
        删除会话关联的所有文件

        Args:
            session: 要清理的会话
            reason: 清理原因

        Returns:
            删除的文件数量
        """
        if not self._storage_service:
            logger.debug(f"Storage service not configured, skipping file cleanup for session {session.id}")
            return 0

        if not session.workspace_path:
            logger.debug(f"Session {session.id} has no workspace_path, skipping file cleanup")
            return 0

        logger.info(
            f"Cleaning up files for session {session.id} "
            f"(reason: {reason}, workspace: {session.workspace_path})"
        )

        try:
            # 从 workspace_path 提取 bucket 和 prefix
            # workspace_path 格式: s3://bucket/sessions/{session_id}/
            parsed = urlparse(session.workspace_path)
            bucket = parsed.netloc  # 存储桶名称

            # 构建删除前缀（包含存储桶路径）
            # 例如: s3://sandbox-workspace/sessions/sess_abc123/
            delete_prefix = session.workspace_path.rstrip('/')

            # 删除所有带该前缀的文件
            deleted_count = await self._storage_service.delete_prefix(delete_prefix)

            logger.info(
                f"Deleted {deleted_count} files for session {session.id} "
                f"(prefix: {delete_prefix})"
            )

            return deleted_count

        except Exception as e:
            logger.error(
                f"Error cleaning up files for session {session.id}: {e}",
                exc_info=True
            )
            return 0

    async def _cleanup_session(
        self,
        session: Session,
        reason: str,
        detail: str
    ) -> None:
        """
        清理会话并销毁容器

        Args:
            session: 要清理的会话
            reason: 清理原因
            detail: 详细信息
        """
        logger.info(
            f"Cleaning up session {session.id}: "
            f"reason={reason}, "
            f"detail={detail}, "
            f"container_id={session.container_id}"
        )

        # 销毁容器（如果调度器支持且容器存在）
        if session.container_id and hasattr(self._scheduler, 'destroy_container'):
            try:
                await self._scheduler.destroy_container(
                    container_id=session.container_id
                )
                logger.info(f"Destroyed container {session.container_id} for session {session.id}")
            except Exception as e:
                # 记录错误但不中断流程
                logger.warning(
                    f"Failed to destroy container {session.container_id} for session {session.id}: {e}"
                )

        # 删除 S3 文件（如果配置了存储服务）
        await self.cleanup_session_files(session, reason)

        # 标记会话为已终止
        session.mark_as_terminated()
        await self._session_repo.save(session)

        logger.info(f"Session {session.id} marked as terminated")

    async def cleanup_by_ids(self, session_ids: list[str]) -> Dict[str, int]:
        """
        按会话 ID 列表清理会话

        Args:
            session_ids: 要清理的会话 ID 列表

        Returns:
            dict: 清理统计信息
        """
        stats = {
            "total": len(session_ids),
            "cleaned": 0,
            "not_found": 0,
            "errors": []
        }

        for session_id in session_ids:
            try:
                session = await self._session_repo.find_by_id(session_id)
                if not session:
                    stats["not_found"] += 1
                    continue

                await self._cleanup_session(
                    session,
                    reason="manual_cleanup",
                    detail=f"Manual cleanup requested"
                )
                stats["cleaned"] += 1

            except Exception as e:
                error_msg = f"Error cleaning session {session_id}: {e}"
                logger.error(error_msg, exc_info=True)
                stats["errors"].append(error_msg)

        logger.info(
            f"Manual session cleanup completed: "
            f"total={stats['total']}, "
            f"cleaned={stats['cleaned']}, "
            f"not_found={stats['not_found']}"
        )

        return stats
