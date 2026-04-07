"""
数据库连接管理

配置和管理 SQLAlchemy 异步引擎和会话。
"""
from contextlib import asynccontextmanager
from dataclasses import dataclass
from typing import AsyncGenerator
from urllib.parse import urlparse, urlunparse

import aiomysql

from sqlalchemy import text
from sqlalchemy.ext.asyncio import create_async_engine, async_sessionmaker, AsyncSession, AsyncEngine
from sqlalchemy.orm import DeclarativeBase

from src.infrastructure.config.settings import get_settings
from src.infrastructure.logging import get_logger


class Base(DeclarativeBase):
    """SQLAlchemy 基类"""
    pass


logger = get_logger(__name__)

LEGACY_DATABASE_NAME = "adp"
TARGET_DATABASE_NAME = "kweaver"


# Import all models so they're registered with Base.metadata
# This is required for create_all() to find all tables
from src.infrastructure.persistence.models.template_model import TemplateModel
from src.infrastructure.persistence.models.session_model import SessionModel
from src.infrastructure.persistence.models.execution_model import ExecutionModel
from src.infrastructure.persistence.models.runtime_node_model import RuntimeNodeModel


class DatabaseManager:
    """
    数据库管理器

    负责创建和管理数据库连接。
    """

    def __init__(self):
        self._engine: AsyncEngine | None = None
        self._session_factory: async_sessionmaker[AsyncSession] | None = None

    def _get_managed_sandbox_table_names(self) -> set[str]:
        """返回 control plane 自己管理的沙箱表名白名单。"""
        return set(Base.metadata.tables.keys())

    @dataclass(frozen=True)
    class _DatabaseConnectionInfo:
        """数据库连接信息。"""

        user: str | None
        password: str | None
        host: str | None
        port: int
        database: str

    def _get_database_connection_info(self) -> _DatabaseConnectionInfo:
        """从配置中解析数据库连接信息。"""
        parsed = urlparse(self._get_runtime_database_url())
        configured_database = parsed.path.lstrip("/")

        return self._DatabaseConnectionInfo(
            user=parsed.username,
            password=parsed.password,
            host=parsed.hostname,
            port=parsed.port or 3306,
            database=configured_database,
        )

    def _get_runtime_database_url(self) -> str:
        """返回运行期使用的数据库 URL，并将旧库名规范化到新库名。"""
        settings = get_settings()
        parsed = urlparse(settings.effective_database_url)
        database_name = parsed.path.lstrip("/")
        if database_name != LEGACY_DATABASE_NAME:
            return settings.effective_database_url

        normalized = parsed._replace(path=f"/{TARGET_DATABASE_NAME}")
        return urlunparse(normalized)

    async def _create_server_pool(self) -> aiomysql.Pool:
        """创建连接到 MySQL 服务端的连接池。"""
        connection_info = self._get_database_connection_info()
        return await aiomysql.create_pool(
            host=connection_info.host,
            port=connection_info.port,
            user=connection_info.user,
            password=connection_info.password,
            db=None,
            autocommit=True,
            minsize=1,
            maxsize=1,
        )

    async def upgrade_legacy_database_name(self) -> None:
        """
        启动时迁移旧数据库名到新数据库名。

        当前支持将旧库 `adp` 升级为 `kweaver`。
        """
        connection_info = self._get_database_connection_info()
        target_database = connection_info.database
        if target_database != TARGET_DATABASE_NAME:
            logger.info(
                "Skipping legacy database rename because configured target database is not managed",
                database=target_database,
            )
            return

        pool = await self._create_server_pool()
        try:
            async with pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    legacy_exists = await self._schema_exists(cursor, LEGACY_DATABASE_NAME)
                    if not legacy_exists:
                        logger.debug(
                            "Legacy database not found, skipping rename",
                            legacy_database=LEGACY_DATABASE_NAME,
                        )
                        return

                    target_exists = await self._schema_exists(cursor, TARGET_DATABASE_NAME)
                    if target_exists:
                        logger.warning(
                            "Target database already exists, checking for missing legacy tables",
                            legacy_database=LEGACY_DATABASE_NAME,
                            target_database=TARGET_DATABASE_NAME,
                        )
                    else:
                        logger.info(
                            "Migrating legacy database name",
                            legacy_database=LEGACY_DATABASE_NAME,
                            target_database=TARGET_DATABASE_NAME,
                        )
                        await cursor.execute(
                            f"CREATE DATABASE `{TARGET_DATABASE_NAME}` "
                            "CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
                        )

                    managed_table_names = self._get_managed_sandbox_table_names()
                    legacy_table_names = [
                        table_name
                        for table_name in await self._list_tables(cursor, LEGACY_DATABASE_NAME)
                        if table_name in managed_table_names
                    ]
                    target_table_names = [
                        table_name
                        for table_name in await self._list_tables(cursor, TARGET_DATABASE_NAME)
                        if table_name in managed_table_names
                    ]
                    missing_table_names = [
                        table_name
                        for table_name in legacy_table_names
                        if table_name not in set(target_table_names)
                    ]

                    for table_name in missing_table_names:
                        await cursor.execute(
                            f"RENAME TABLE `{LEGACY_DATABASE_NAME}`.`{table_name}` "
                            f"TO `{TARGET_DATABASE_NAME}`.`{table_name}`"
                        )

                    remaining_legacy_tables = [
                        table_name
                        for table_name in await self._list_tables(cursor, LEGACY_DATABASE_NAME)
                        if table_name in managed_table_names
                    ]
                    if remaining_legacy_tables:
                        logger.warning(
                            "Legacy database still contains tables after migration",
                            legacy_database=LEGACY_DATABASE_NAME,
                            target_database=TARGET_DATABASE_NAME,
                            migrated_tables=len(missing_table_names),
                            remaining_tables=remaining_legacy_tables,
                        )
                        return

                    logger.info(
                        "Legacy database migration completed without dropping source database",
                        legacy_database=LEGACY_DATABASE_NAME,
                        target_database=TARGET_DATABASE_NAME,
                        migrated_tables=len(missing_table_names),
                    )
        finally:
            pool.close()
            await pool.wait_closed()

    async def _schema_exists(self, cursor, schema_name: str) -> bool:
        """检查 schema 是否存在。"""
        await cursor.execute(
            """
            SELECT COUNT(*)
            FROM INFORMATION_SCHEMA.SCHEMATA
            WHERE SCHEMA_NAME = %s
            """,
            (schema_name,),
        )
        result = await cursor.fetchone()
        return bool(result and result[0])

    async def _list_tables(self, cursor, schema_name: str) -> list[str]:
        """列出 schema 下的所有基础表。"""
        await cursor.execute(
            """
            SELECT TABLE_NAME
            FROM INFORMATION_SCHEMA.TABLES
            WHERE TABLE_SCHEMA = %s
              AND TABLE_TYPE = 'BASE TABLE'
            ORDER BY TABLE_NAME
            """,
            (schema_name,),
        )
        rows = await cursor.fetchall()
        return [row[0] for row in rows]

    async def ensure_database_exists(self) -> None:
        """
        确保数据库存在，如果不存在则创建

        使用原始连接（不通过 SQLAlchemy）来创建数据库，
        因为 SQLAlchemy 需要数据库已存在才能创建引擎。
        """
        connection_info = self._get_database_connection_info()
        db_name = connection_info.database
        pool = await self._create_server_pool()

        try:
            async with pool.acquire() as conn:
                async with conn.cursor() as cursor:
                    # 检查数据库是否存在
                    result = await self._schema_exists(cursor, db_name)

                    if not result:
                        # 数据库不存在，创建它
                        await cursor.execute(
                            f"CREATE DATABASE `{db_name}` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
                        )
                        logger.info(f"Database '{db_name}' created successfully")
                    else:
                        logger.debug(f"Database '{db_name}' already exists")
        finally:
            pool.close()
            await pool.wait_closed()

    async def initialize(self) -> None:
        """初始化数据库引擎（确保数据库存在）"""
        # 先确保数据库存在
        await self.ensure_database_exists()

        # 然后创建引擎
        settings = get_settings()
        self._engine = create_async_engine(
            self._get_runtime_database_url(),
            echo=settings.log_level == "DEBUG",
            pool_size=settings.db_pool_size,
            max_overflow=settings.db_max_overflow,
            pool_recycle=settings.db_pool_recycle,
        )
        self._session_factory = async_sessionmaker(
            bind=self._engine,
            class_=AsyncSession,
            expire_on_commit=False,
        )

    async def create_tables(self) -> None:
        """创建所有数据库表"""
        if self._engine is None:
            raise RuntimeError("DatabaseManager not initialized. Call initialize() first.")

        async with self._engine.begin() as conn:
            await conn.run_sync(Base.metadata.create_all)

    async def run_startup_schema_migrations(self) -> None:
        """
        启动时执行幂等 schema 升级。

        当前用于兼容旧库缺失 `f_python_package_index_url` 字段的场景。
        """
        if self._engine is None:
            raise RuntimeError("DatabaseManager not initialized. Call initialize() first.")

        backend_name = self._engine.url.get_backend_name()
        if backend_name != "mysql":
            logger.info(
                "Skipping startup schema migrations for unsupported backend",
                backend=backend_name,
            )
            return

        async with self._engine.begin() as conn:
            table_name = "t_sandbox_session"
            column_name = "f_python_package_index_url"

            table_exists = await self._mariadb_table_exists(conn, table_name)
            if not table_exists:
                logger.info(
                    "Skipping startup schema migration because target table does not exist",
                    table=table_name,
                )
                return

            column_exists = await self._mariadb_column_exists(
                conn,
                table_name,
                column_name,
            )
            if column_exists:
                logger.info(
                    "Startup schema migration check passed",
                    table=table_name,
                    column=column_name,
                    action="skip",
                )
                return

            logger.info(
                "Applying startup schema migration",
                table=table_name,
                column=column_name,
                action="add_column",
            )
            await conn.execute(
                text(
                    """
                    ALTER TABLE `t_sandbox_session`
                    ADD COLUMN `f_python_package_index_url` varchar(512) NOT NULL
                    DEFAULT 'https://pypi.org/simple/'
                    AFTER `f_completed_at`
                    """
                )
            )
            logger.info(
                "Startup schema migration applied successfully",
                table=table_name,
                column=column_name,
            )

    async def _mariadb_table_exists(self, conn, table_name: str) -> bool:
        """检查 MariaDB 表是否存在。"""
        result = await conn.execute(
            text(
                """
                SELECT COUNT(*)
                FROM information_schema.TABLES
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = :table_name
                """
            ),
            {"table_name": table_name},
        )
        return bool(result.scalar())

    async def _mariadb_column_exists(self, conn, table_name: str, column_name: str) -> bool:
        """检查 MariaDB 列是否存在。"""
        result = await conn.execute(
            text(
                """
                SELECT COUNT(*)
                FROM information_schema.COLUMNS
                WHERE TABLE_SCHEMA = DATABASE()
                  AND TABLE_NAME = :table_name
                  AND COLUMN_NAME = :column_name
                """
            ),
            {"table_name": table_name, "column_name": column_name},
        )
        return bool(result.scalar())

    async def initialize_with_seed(
        self,
        create_tables: bool = False,
        seed_data: bool = False,
        force_seed: bool = False
    ) -> dict:
        """
        初始化数据库并可选地创建表和种子数据

        Args:
            create_tables: 是否创建数据库表
            seed_data: 是否初始化种子数据
            force_seed: 是否强制重新创建种子数据

        Returns:
            包含初始化结果的字典
        """
        result = {
            "tables_created": False,
            "seeded": False,
            "seed_stats": {}
        }

        if create_tables:
            await self.create_tables()
            result["tables_created"] = True

        if seed_data:
            from src.infrastructure.persistence.seed.seeder import seed_default_data
            stats = await seed_default_data(force=force_seed)
            result["seeded"] = True
            result["seed_stats"] = stats

        return result

    @asynccontextmanager
    async def get_session(self) -> AsyncGenerator[AsyncSession, None]:
        """
        获取数据库会话（上下文管理器）

        用法:
            async with db_manager.get_session() as session:
                # 使用 session
        """
        if self._session_factory is None:
            raise RuntimeError("DatabaseManager not initialized. Call initialize() first.")

        async with self._session_factory() as session:
            try:
                yield session
                await session.commit()
            except Exception:
                await session.rollback()
                raise

    async def close(self) -> None:
        """关闭数据库连接"""
        if self._engine:
            await self._engine.dispose()


# 全局数据库管理器实例
db_manager = DatabaseManager()
