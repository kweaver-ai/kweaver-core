"""
会话创建超时检测服务

负责定期检测处于 creating 状态超过阈值的会话，并将其标记为 failed。
这解决了 executor 容器初始化失败（如依赖安装失败）导致会话永久处于 creating 状态的问题。
"""
import logging
from datetime import datetime, timedelta
from typing import Dict, Optional

from src.domain.entities.session import Session, SessionStatus
from src.domain.repositories.session_repository import ISessionRepository

logger = logging.getLogger(__name__)


class SessionStuckCreatingService:
    """
    会话创建超时检测服务

    职责：
    1. 定期扫描处于 "creating" 状态的会话
    2. 检查创建时间是否超过配置的超时阈值
    3. 将超时会话标记为 "failed" 状态

    检测策略：
    - 默认超时时间：300 秒（5 分钟，可配置）
    - 检测间隔：默认与 cleanup_interval_seconds 一致（5 分钟）
    """

    def __init__(
        self,
        session_repo: ISessionRepository,
        creating_timeout_seconds: int = 300,
    ):
        """
        初始化会话创建超时检测服务

        Args:
            session_repo: 会话仓储
            creating_timeout_seconds: 创建超时时间（秒），必须 >= 30 秒
        """
        self._session_repo = session_repo
        self._timeout = timedelta(seconds=creating_timeout_seconds)

    async def check_and_mark_stuck_sessions(self) -> Dict[str, int]:
        """
        检测并标记处于 creating 状态超时的会话

        Returns:
            dict: 检测统计信息
                - total_checked: 检查的会话数
                - marked_failed: 标记为 failed 的会话数
                - errors: 错误列表
        """
        stats = {
            "total_checked": 0,
            "marked_failed": 0,
            "errors": []
        }

        try:
            now = datetime.now()
            timeout_threshold = now - self._timeout

            # 查询所有处于 creating 状态的会话
            creating_sessions = await self._session_repo.find_by_status(SessionStatus.CREATING)
            stats["total_checked"] = len(creating_sessions)

            if creating_sessions:
                logger.info(
                    f"Checking {len(creating_sessions)} sessions in 'creating' status, "
                    f"timeout_threshold={timeout_threshold.isoformat()}"
                )

            for session in creating_sessions:
                try:
                    # 检查是否创建时间超过阈值
                    if session.created_at and session.created_at < timeout_threshold:
                        await self._mark_session_as_failed(
                            session,
                            reason="creating_timeout",
                            detail=f"Session stuck in 'creating' status for {(now - session.created_at).total_seconds():.0f} seconds "
                                   f"(timeout: {self._timeout.total_seconds():.0f}s)"
                        )
                        stats["marked_failed"] += 1
                    else:
                        # 记录还未超时的会话（调试用）
                        if session.created_at:
                            time_in_creating = (now - session.created_at).total_seconds()
                            logger.debug(
                                f"Session {session.id} has been in 'creating' status for {time_in_creating:.0f}s "
                                f"(timeout: {self._timeout.total_seconds():.0f}s)"
                            )

                except Exception as e:
                    error_msg = f"Error processing session {session.id}: {e}"
                    logger.error(error_msg, exc_info=True)
                    stats["errors"].append(error_msg)

            if stats["marked_failed"] > 0:
                logger.info(
                    f"Session stuck-creating check completed: "
                    f"checked={stats['total_checked']}, "
                    f"marked_failed={stats['marked_failed']}"
                )

        except Exception as e:
            error_msg = f"Fatal error during stuck-creating session check: {e}"
            logger.error(error_msg, exc_info=True)
            stats["errors"].append(error_msg)

        return stats

    async def _mark_session_as_failed(
        self,
        session: Session,
        reason: str,
        detail: str
    ) -> None:
        """
        标记会话为失败状态

        Args:
            session: 要标记的会话
            reason: 失败原因
            detail: 详细信息
        """
        logger.warning(
            f"Marking session {session.id} as failed: "
            f"reason={reason}, "
            f"detail={detail}, "
            f"container_id={session.container_id}, "
            f"created_at={session.created_at}"
        )

        # 标记会话为失败状态
        session.mark_as_failed()
        await self._session_repo.save(session)

        logger.info(f"Session {session.id} marked as failed")
