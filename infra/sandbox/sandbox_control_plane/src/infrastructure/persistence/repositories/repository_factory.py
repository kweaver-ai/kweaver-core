"""
仓储工厂

提供数据库会话感知的仓储实例。
"""
from contextlib import asynccontextmanager
from typing import AsyncGenerator

from src.infrastructure.persistence.database import db_manager
from src.infrastructure.persistence.repositories.sql_session_repository import SqlSessionRepository
from src.infrastructure.persistence.repositories.sql_execution_repository import SqlExecutionRepository
from src.infrastructure.persistence.repositories.sql_template_repository import SqlTemplateRepository

from src.domain.repositories.session_repository import ISessionRepository
from src.domain.repositories.execution_repository import IExecutionRepository
from src.domain.repositories.template_repository import ITemplateRepository


class RepositoryFactory:
    """
    仓储工厂

    负责创建仓储实例，并注入数据库会话。
    """

    @staticmethod
    @asynccontextmanager
    async def get_repositories() -> AsyncGenerator[dict[str, object], None]:
        """
        获取所有仓储实例（上下文管理器）

        用法:
            async with RepositoryFactory.get_repositories() as repos:
                session_repo = repos["session_repo"]
                # 使用仓储
        """
        async with db_manager.get_session() as session:
            yield {
                "session_repo": SqlSessionRepository(session),
                "execution_repo": SqlExecutionRepository(session),
                "template_repo": SqlTemplateRepository(session),
            }

    @staticmethod
    def create_session_repository(session) -> ISessionRepository:
        """创建会话仓储"""
        return SqlSessionRepository(session)

    @staticmethod
    def create_execution_repository(session) -> IExecutionRepository:
        """创建执行仓储"""
        return SqlExecutionRepository(session)

    @staticmethod
    def create_template_repository(session) -> ITemplateRepository:
        """创建模板仓储"""
        return SqlTemplateRepository(session)
