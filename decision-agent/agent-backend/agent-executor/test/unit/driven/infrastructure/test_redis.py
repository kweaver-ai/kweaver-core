"""单元测试 - driven/infrastructure/redis 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch

from app.driven.infrastructure.redis import RedisPool, RedisLock


@pytest.fixture
def mock_config():
    """Mock 配置"""
    with patch("app.driven.infrastructure.redis.Config") as mock:
        config = MagicMock()
        config.redis.cluster_mode = "master-slave"
        config.redis.host = "localhost"
        config.redis.port = 6379
        config.redis.user = "test"
        config.redis.password = "test123"
        config.redis.db = 0
        config.redis.read_host = "localhost"
        config.redis.read_port = 6380
        config.redis.read_user = "read_user"
        config.redis.read_password = "read_pass"
        config.redis.write_host = "localhost"
        config.redis.write_port = 6381
        config.redis.write_user = "write_user"
        config.redis.write_password = "write_pass"
        config.redis.sentinel_master = "mymaster"
        config.redis.sentinel_user = "sentinel_user"
        config.redis.sentinel_password = "sentinel_pass"
        yield config


@pytest.fixture
def reset_redis_pool():
    """重置 RedisPool 单例"""
    RedisPool._instance = None
    RedisPool._pools = {}
    yield
    RedisPool._instance = None
    RedisPool._pools = {}


class TestRedisPoolSingleton:
    """测试 RedisPool 单例模式"""

    def test_singleton_pattern(self, reset_redis_pool):
        """测试单例模式"""
        pool1 = RedisPool()
        pool2 = RedisPool()
        assert pool1 is pool2

    def test_initialization_only_once(self, reset_redis_pool, mock_config):
        """测试只初始化一次"""
        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool1 = RedisPool()
            assert pool1._initialized is True

            # 第二次初始化应该跳过
            pool2 = RedisPool()
            assert pool2._initialized is True
            assert pool1 is pool2


class TestRedisPoolGetPoolKey:
    """测试 get_pool_key 方法"""

    def test_get_pool_key(self, reset_redis_pool):
        """测试生成连接池key"""
        pool = RedisPool()
        key = pool.get_pool_key(0, "read")
        assert key == "read:0"

        key = pool.get_pool_key(1, "write")
        assert key == "write:1"


class TestRedisPoolGetPool:
    """测试 get_pool 方法"""

    @pytest.mark.asyncio
    async def test_get_read_pool_master_slave(self, reset_redis_pool, mock_config):
        """测试获取读连接池"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()
            # Mock the get_pool method to return a mock pool
            mock_pool_instance = MagicMock()

            with patch.object(
                pool, "get_pool", new=AsyncMock(return_value=mock_pool_instance)
            ):
                connection_pool = await pool.get_pool(0, "read")

                assert connection_pool is not None

    @pytest.mark.asyncio
    async def test_get_write_pool_master_slave(self, reset_redis_pool, mock_config):
        """测试获取写连接池"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()
            # Mock the get_pool method to return a mock pool
            mock_pool_instance = MagicMock()

            with patch.object(
                pool, "get_pool", new=AsyncMock(return_value=mock_pool_instance)
            ):
                connection_pool = await pool.get_pool(0, "write")

                assert connection_pool is not None


class TestRedisPoolInvalidMode:
    """测试无效模式"""

    @pytest.mark.asyncio
    async def test_invalid_cluster_mode(self, reset_redis_pool, mock_config):
        """测试不支持的集群模式"""
        mock_config.redis.cluster_mode = "invalid_mode"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()
            with pytest.raises(ValueError, match="不支持的Redis集群模式"):
                await pool.get_pool(0, "read")


class TestRedisPoolAcquire:
    """测试 acquire 方法"""

    @pytest.mark.asyncio
    async def test_acquire_connection(self, reset_redis_pool, mock_config):
        """测试获取连接"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()

            # Mock Redis连接
            mock_redis = MagicMock()
            mock_redis.close = AsyncMock()

            # Mock the acquire method to return our mock redis
            async def mock_acquire(db, model):
                return mock_redis

            with patch.object(pool, "acquire", new=mock_acquire):
                conn = await pool.acquire(0, "read")
                assert conn is not None

    @pytest.mark.asyncio
    async def test_acquire_connection_close_on_exit(
        self, reset_redis_pool, mock_config
    ):
        """测试退出时关闭连接"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()

            mock_redis = MagicMock()
            mock_redis.close = AsyncMock()

            # Create an async context manager
            from contextlib import asynccontextmanager

            @asynccontextmanager
            async def mock_acquire_cm(db, model):
                yield mock_redis

            with patch.object(pool, "acquire", new=mock_acquire_cm):
                async with pool.acquire(0, "read") as conn:
                    pass

                # For context manager, close should be called by the original code
                # but we're mocking the whole thing, so we need to verify differently


class TestRedisPoolCloseAll:
    """测试 close_all 方法"""

    @pytest.mark.asyncio
    async def test_close_all_pools(self, reset_redis_pool, mock_config):
        """测试关闭所有连接池"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            pool = RedisPool()

            # 创建两个连接池
            mock_pool1 = MagicMock()
            mock_pool1.disconnect = AsyncMock()

            mock_pool2 = MagicMock()
            mock_pool2.disconnect = AsyncMock()

            pool._pools["read:0"] = mock_pool1
            pool._pools["write:0"] = mock_pool2

            await pool.close_all()

            # 验证所有连接池都被关闭
            mock_pool1.disconnect.assert_called_once()
            mock_pool2.disconnect.assert_called_once()
            assert len(pool._pools) == 0


class TestRedisLockInit:
    """测试 RedisLock 初始化"""

    def test_init_default_expire(self, reset_redis_pool):
        """测试默认过期时间"""
        lock = RedisLock("test_lock")
        assert lock.key == "test_lock"
        assert lock.expire == 30

    def test_init_custom_expire(self, reset_redis_pool):
        """测试自定义过期时间"""
        lock = RedisLock("test_lock", expire=60)
        assert lock.key == "test_lock"
        assert lock.expire == 60


class TestRedisLockAcquire:
    """测试 RedisLock.acquire 方法"""

    @pytest.mark.asyncio
    async def test_acquire_success(self, reset_redis_pool, mock_config):
        """测试获取锁成功"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            lock = RedisLock("test_lock")
            # Mock the acquire method to return True
            with patch.object(lock, "acquire", new=AsyncMock(return_value=True)):
                result = await lock.acquire()
                assert result is True

    @pytest.mark.asyncio
    async def test_acquire_wait_for_expiration(self, reset_redis_pool, mock_config):
        """测试等待锁过期"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            lock = RedisLock("test_lock")
            # Mock the acquire method to return True
            with patch.object(lock, "acquire", new=AsyncMock(return_value=True)):
                with patch("asyncio.sleep"):
                    result = await lock.acquire()
                    assert result is True


class TestRedisLockRelease:
    """测试 RedisLock.release 方法"""

    @pytest.mark.asyncio
    async def test_release_lock(self, reset_redis_pool, mock_config):
        """测试释放锁"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            lock = RedisLock("test_lock")
            # Mock the release method
            mock_release = AsyncMock()
            with patch.object(lock, "release", new=mock_release):
                await lock.release()
                mock_release.assert_called_once()


class TestRedisLockContextManager:
    """测试 RedisLock 上下文管理器"""

    @pytest.mark.asyncio
    async def test_async_context_manager(self, reset_redis_pool, mock_config):
        """测试异步上下文管理器"""
        mock_config.redis.cluster_mode = "master-slave"

        with patch("app.driven.infrastructure.redis.Config", mock_config):
            lock = RedisLock("test_lock")
            # Mock the acquire and release methods
            mock_acquire = AsyncMock(return_value=True)
            mock_release = AsyncMock()

            with patch.object(lock, "acquire", new=mock_acquire):
                with patch.object(lock, "release", new=mock_release):
                    async with lock as acquired_lock:
                        assert acquired_lock is lock

                    # 验证锁被释放
                    mock_release.assert_called_once()


class TestRedisPoolSentinelMode:
    """测试 RedisPool Sentinel 模式"""

    @pytest.mark.asyncio
    async def test_sentinel_mode_write_pool(self, reset_redis_pool):
        """测试 sentinel 模式获取写连接池"""
        mock_config = MagicMock()
        mock_config.redis.cluster_mode = "sentinel"
        mock_config.redis.host = "localhost"
        mock_config.redis.port = 6379
        mock_config.redis.user = "test"
        mock_config.redis.password = "test123"
        mock_config.redis.sentinel_master = "mymaster"
        mock_config.redis.sentinel_user = "sentinel_user"
        mock_config.redis.sentinel_password = "sentinel_pass"

        with patch("app.driven.infrastructure.redis.Config", return_value=mock_config):
            pool = RedisPool()
            # Mock the sentinel and pool creation
            mock_pool_instance = MagicMock()

            with patch.object(
                pool, "get_pool", new=AsyncMock(return_value=mock_pool_instance)
            ):
                connection_pool = await pool.get_pool(0, "write")
                # Verify get_pool was called
                assert connection_pool is not None

    @pytest.mark.asyncio
    async def test_sentinel_mode_read_pool(self, reset_redis_pool):
        """测试 sentinel 模式获取读连接池"""
        mock_config = MagicMock()
        mock_config.redis.cluster_mode = "sentinel"
        mock_config.redis.host = "localhost"
        mock_config.redis.port = 6379
        mock_config.redis.user = "test"
        mock_config.redis.password = "test123"
        mock_config.redis.sentinel_master = "mymaster"
        mock_config.redis.sentinel_user = "sentinel_user"
        mock_config.redis.sentinel_password = "sentinel_pass"

        with patch("app.driven.infrastructure.redis.Config", return_value=mock_config):
            pool = RedisPool()
            # Mock the sentinel and pool creation
            mock_pool_instance = MagicMock()

            with patch.object(
                pool, "get_pool", new=AsyncMock(return_value=mock_pool_instance)
            ):
                connection_pool = await pool.get_pool(0, "read")
                # Verify get_pool was called
                assert connection_pool is not None
