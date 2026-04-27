"""
数据库管理器单元测试

测试 DatabaseManager 的功能。
"""
import pytest
from unittest.mock import Mock, AsyncMock, patch

from src.infrastructure.persistence.database import (
    Base,
    DatabaseManager,
    LEGACY_DATABASE_NAME,
    TARGET_DATABASE_NAME,
)


class TestDatabaseManager:
    """数据库管理器测试"""

    @pytest.fixture
    def mock_settings(self):
        """模拟设置"""
        settings = Mock()
        settings.effective_database_url = "mysql+aiomysql://root:password@localhost:3306/kweaver"
        settings.log_level = "INFO"
        settings.db_pool_size = 5
        settings.db_max_overflow = 10
        settings.db_pool_recycle = 3600
        return settings

    @pytest.fixture
    def db_manager(self):
        """创建数据库管理器"""
        return DatabaseManager()

    def test_init(self, db_manager):
        """测试初始化"""
        assert db_manager._engine is None
        assert db_manager._session_factory is None

    def test_get_runtime_database_url_normalizes_legacy_database_name(self, db_manager, mock_settings):
        """测试运行期数据库 URL 会将旧库名规范化到新库名。"""
        mock_settings.effective_database_url = "mysql+aiomysql://root:password@localhost:3306/adp"

        with patch("src.infrastructure.persistence.database.get_settings", return_value=mock_settings):
            runtime_url = db_manager._get_runtime_database_url()

        assert runtime_url == "mysql+aiomysql://root:password@localhost:3306/kweaver"

    def test_get_managed_sandbox_table_names(self, db_manager):
        """测试只返回受管的沙箱表。"""
        assert db_manager._get_managed_sandbox_table_names() == {
            "t_sandbox_execution",
            "t_sandbox_runtime_node",
            "t_sandbox_session",
            "t_sandbox_template",
        }

    @pytest.mark.asyncio
    async def test_ensure_database_exists_integration(self, db_manager):
        """测试确保数据库存在（集成测试，需要实际连接）"""
        # This test requires actual database connection
        # Skip in unit tests
        pass

    @pytest.mark.asyncio
    async def test_initialize_integration(self, db_manager):
        """测试初始化数据库引擎（集成测试）"""
        # This test requires actual database connection
        # Skip in unit tests
        pass

    @pytest.mark.asyncio
    async def test_create_tables_not_initialized(self, db_manager):
        """测试未初始化时创建表"""
        with pytest.raises(RuntimeError, match="not initialized"):
            await db_manager.create_tables()

    @pytest.mark.asyncio
    async def test_run_startup_schema_migrations_not_initialized(self, db_manager):
        """测试未初始化时执行启动迁移。"""
        with pytest.raises(RuntimeError, match="not initialized"):
            await db_manager.run_startup_schema_migrations()

    @pytest.mark.asyncio
    async def test_initialize_with_seed_no_tables_no_seed(self, db_manager):
        """测试初始化不创建表和种子数据"""
        result = await db_manager.initialize_with_seed(
            create_tables=False,
            seed_data=False
        )

        assert result["tables_created"] is False
        assert result["seeded"] is False
        assert result["seed_stats"] == {}

    @pytest.mark.asyncio
    async def test_initialize_with_seed_create_tables(self, db_manager):
        """测试初始化创建表"""
        with patch.object(db_manager, 'create_tables', new_callable=AsyncMock):
            result = await db_manager.initialize_with_seed(
                create_tables=True,
                seed_data=False
            )

            assert result["tables_created"] is True
            assert result["seeded"] is False

    @pytest.mark.asyncio
    async def test_initialize_with_seed_with_seed(self, db_manager):
        """测试初始化带种子数据"""
        with patch.object(db_manager, 'create_tables', new_callable=AsyncMock):
            with patch('src.infrastructure.persistence.seed.seeder.seed_default_data',
                      new_callable=AsyncMock, return_value={"templates": 1, "runtime_nodes": 1}):
                result = await db_manager.initialize_with_seed(
                    create_tables=True,
                    seed_data=True
                )

                assert result["tables_created"] is True
                assert result["seeded"] is True
                assert result["seed_stats"]["templates"] == 1

    @pytest.mark.asyncio
    async def test_get_session_not_initialized(self, db_manager):
        """测试未初始化时获取会话"""
        with pytest.raises(RuntimeError, match="not initialized"):
            async with db_manager.get_session():
                pass

    @pytest.mark.asyncio
    async def test_close_no_engine(self, db_manager):
        """测试没有引擎时关闭"""
        # Should not raise error
        await db_manager.close()

    @pytest.mark.asyncio
    async def test_close_with_engine(self, db_manager):
        """测试有引擎时关闭"""
        mock_engine = Mock()
        mock_engine.dispose = AsyncMock()
        db_manager._engine = mock_engine

        await db_manager.close()

        mock_engine.dispose.assert_called_once()

    @pytest.mark.asyncio
    async def test_upgrade_legacy_database_name_migrates_managed_tables_without_dropping_legacy_database(
        self,
        db_manager,
        mock_settings,
    ):
        """测试旧数据库名会迁移受管表，但不会删除旧库。"""
        mock_cursor = AsyncMock()
        mock_cursor.fetchone = AsyncMock(side_effect=[(1,), (0,)])
        mock_cursor.fetchall = AsyncMock(
            side_effect=[
                [("t_sandbox_session",), ("t_sandbox_execution",)],
                [],
                [],
            ]
        )

        mock_cursor_context = AsyncMock()
        mock_cursor_context.__aenter__.return_value = mock_cursor
        mock_cursor_context.__aexit__.return_value = None

        mock_conn = Mock()
        mock_conn.cursor = Mock(return_value=mock_cursor_context)

        mock_acquire = AsyncMock()
        mock_acquire.__aenter__.return_value = mock_conn
        mock_acquire.__aexit__.return_value = None

        mock_pool = Mock()
        mock_pool.acquire = Mock(return_value=mock_acquire)
        mock_pool.close = Mock()
        mock_pool.wait_closed = AsyncMock()

        with patch("src.infrastructure.persistence.database.get_settings", return_value=mock_settings):
            with patch(
                "src.infrastructure.persistence.database.aiomysql.create_pool",
                AsyncMock(return_value=mock_pool),
            ):
                await db_manager.upgrade_legacy_database_name()

        executed_sql = [call.args[0] for call in mock_cursor.execute.await_args_list]
        assert any(f"CREATE DATABASE `{TARGET_DATABASE_NAME}`" in stmt for stmt in executed_sql)
        assert any(
            f"RENAME TABLE `{LEGACY_DATABASE_NAME}`.`t_sandbox_session` "
            f"TO `{TARGET_DATABASE_NAME}`.`t_sandbox_session`" in stmt
            for stmt in executed_sql
        )
        assert any(
            f"RENAME TABLE `{LEGACY_DATABASE_NAME}`.`t_sandbox_execution` "
            f"TO `{TARGET_DATABASE_NAME}`.`t_sandbox_execution`" in stmt
            for stmt in executed_sql
        )
        assert not any(f"DROP DATABASE `{LEGACY_DATABASE_NAME}`" in stmt for stmt in executed_sql)

    @pytest.mark.asyncio
    async def test_upgrade_legacy_database_name_migrates_missing_tables_when_target_exists(
        self,
        db_manager,
        mock_settings,
    ):
        """测试新数据库已存在但缺表时会迁移缺失表且保留旧库。"""
        mock_cursor = AsyncMock()
        mock_cursor.fetchone = AsyncMock(side_effect=[(1,), (1,)])
        mock_cursor.fetchall = AsyncMock(
            side_effect=[
                [("t_sandbox_session",), ("t_sandbox_execution",)],
                [],
                [],
            ]
        )

        mock_cursor_context = AsyncMock()
        mock_cursor_context.__aenter__.return_value = mock_cursor
        mock_cursor_context.__aexit__.return_value = None

        mock_conn = Mock()
        mock_conn.cursor = Mock(return_value=mock_cursor_context)

        mock_acquire = AsyncMock()
        mock_acquire.__aenter__.return_value = mock_conn
        mock_acquire.__aexit__.return_value = None

        mock_pool = Mock()
        mock_pool.acquire = Mock(return_value=mock_acquire)
        mock_pool.close = Mock()
        mock_pool.wait_closed = AsyncMock()

        with patch("src.infrastructure.persistence.database.get_settings", return_value=mock_settings):
            with patch(
                "src.infrastructure.persistence.database.aiomysql.create_pool",
                AsyncMock(return_value=mock_pool),
            ):
                await db_manager.upgrade_legacy_database_name()

        executed_sql = [call.args[0] for call in mock_cursor.execute.await_args_list]
        assert not any(f"CREATE DATABASE `{TARGET_DATABASE_NAME}`" in stmt for stmt in executed_sql)
        assert any(
            f"RENAME TABLE `{LEGACY_DATABASE_NAME}`.`t_sandbox_execution` "
            f"TO `{TARGET_DATABASE_NAME}`.`t_sandbox_execution`" in stmt
            for stmt in executed_sql
        )
        assert not any(f"DROP DATABASE `{LEGACY_DATABASE_NAME}`" in stmt for stmt in executed_sql)

    @pytest.mark.asyncio
    async def test_upgrade_legacy_database_name_keeps_legacy_database_when_tables_remain(
        self,
        db_manager,
        mock_settings,
    ):
        """测试目标库已有同名表时保留旧库剩余表，避免覆盖数据。"""
        mock_cursor = AsyncMock()
        mock_cursor.fetchone = AsyncMock(side_effect=[(1,), (1,)])
        mock_cursor.fetchall = AsyncMock(
            side_effect=[
                [("t_sandbox_session",), ("t_sandbox_execution",)],
                [("t_sandbox_session",), ("t_sandbox_execution",)],
                [("t_sandbox_session",), ("t_sandbox_execution",)],
            ]
        )

        mock_cursor_context = AsyncMock()
        mock_cursor_context.__aenter__.return_value = mock_cursor
        mock_cursor_context.__aexit__.return_value = None

        mock_conn = Mock()
        mock_conn.cursor = Mock(return_value=mock_cursor_context)

        mock_acquire = AsyncMock()
        mock_acquire.__aenter__.return_value = mock_conn
        mock_acquire.__aexit__.return_value = None

        mock_pool = Mock()
        mock_pool.acquire = Mock(return_value=mock_acquire)
        mock_pool.close = Mock()
        mock_pool.wait_closed = AsyncMock()

        with patch("src.infrastructure.persistence.database.get_settings", return_value=mock_settings):
            with patch(
                "src.infrastructure.persistence.database.aiomysql.create_pool",
                AsyncMock(return_value=mock_pool),
            ):
                await db_manager.upgrade_legacy_database_name()

        executed_sql = [call.args[0] for call in mock_cursor.execute.await_args_list]
        assert not any("RENAME TABLE" in stmt for stmt in executed_sql)
        assert not any(f"DROP DATABASE `{LEGACY_DATABASE_NAME}`" in stmt for stmt in executed_sql)

    @pytest.mark.asyncio
    async def test_upgrade_legacy_database_name_ignores_non_sandbox_tables(
        self,
        db_manager,
        mock_settings,
    ):
        """测试只迁移受管的沙箱表，不处理其他业务表。"""
        mock_cursor = AsyncMock()
        mock_cursor.fetchone = AsyncMock(side_effect=[(1,), (1,)])
        mock_cursor.fetchall = AsyncMock(
            side_effect=[
                [("t_sandbox_session",), ("t_other_business",)],
                [],
                [],
            ]
        )

        mock_cursor_context = AsyncMock()
        mock_cursor_context.__aenter__.return_value = mock_cursor
        mock_cursor_context.__aexit__.return_value = None

        mock_conn = Mock()
        mock_conn.cursor = Mock(return_value=mock_cursor_context)

        mock_acquire = AsyncMock()
        mock_acquire.__aenter__.return_value = mock_conn
        mock_acquire.__aexit__.return_value = None

        mock_pool = Mock()
        mock_pool.acquire = Mock(return_value=mock_acquire)
        mock_pool.close = Mock()
        mock_pool.wait_closed = AsyncMock()

        with patch("src.infrastructure.persistence.database.get_settings", return_value=mock_settings):
            with patch(
                "src.infrastructure.persistence.database.aiomysql.create_pool",
                AsyncMock(return_value=mock_pool),
            ):
                await db_manager.upgrade_legacy_database_name()

        executed_sql = [call.args[0] for call in mock_cursor.execute.await_args_list]
        assert any(
            f"RENAME TABLE `{LEGACY_DATABASE_NAME}`.`t_sandbox_session` "
            f"TO `{TARGET_DATABASE_NAME}`.`t_sandbox_session`" in stmt
            for stmt in executed_sql
        )
        assert not any("t_other_business" in stmt for stmt in executed_sql)

    @pytest.mark.asyncio
    async def test_run_startup_schema_migrations_adds_missing_column(self, db_manager):
        """测试启动迁移会补齐缺失字段。"""
        mock_conn = AsyncMock()
        table_exists_result = Mock()
        table_exists_result.scalar.return_value = 1
        column_exists_result = Mock()
        column_exists_result.scalar.return_value = 0
        alter_result = Mock()
        mock_conn.execute = AsyncMock(
            side_effect=[table_exists_result, column_exists_result, alter_result]
        )

        mock_begin = AsyncMock()
        mock_begin.__aenter__.return_value = mock_conn
        mock_begin.__aexit__.return_value = None

        mock_engine = Mock()
        mock_engine.url.get_backend_name.return_value = "mysql"
        mock_engine.begin.return_value = mock_begin
        db_manager._engine = mock_engine

        await db_manager.run_startup_schema_migrations()

        assert mock_conn.execute.await_count == 3
        alter_stmt = str(mock_conn.execute.await_args_list[2].args[0])
        assert "ALTER TABLE `t_sandbox_session`" in alter_stmt
        assert "ADD COLUMN `f_python_package_index_url`" in alter_stmt

    @pytest.mark.asyncio
    async def test_run_startup_schema_migrations_skips_existing_column(self, db_manager):
        """测试启动迁移在字段已存在时跳过。"""
        mock_conn = AsyncMock()
        table_exists_result = Mock()
        table_exists_result.scalar.return_value = 1
        column_exists_result = Mock()
        column_exists_result.scalar.return_value = 1
        mock_conn.execute = AsyncMock(
            side_effect=[table_exists_result, column_exists_result]
        )

        mock_begin = AsyncMock()
        mock_begin.__aenter__.return_value = mock_conn
        mock_begin.__aexit__.return_value = None

        mock_engine = Mock()
        mock_engine.url.get_backend_name.return_value = "mysql"
        mock_engine.begin.return_value = mock_begin
        db_manager._engine = mock_engine

        await db_manager.run_startup_schema_migrations()

        assert mock_conn.execute.await_count == 2


class TestBase:
    """SQLAlchemy 基类测试"""

    def test_base_is_declarative_base(self):
        """测试 Base 是 DeclarativeBase"""
        from sqlalchemy.orm import DeclarativeBase
        assert issubclass(Base, DeclarativeBase)
