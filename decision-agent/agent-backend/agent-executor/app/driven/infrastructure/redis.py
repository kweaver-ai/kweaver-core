import asyncio
import time
from contextlib import asynccontextmanager
from typing import Dict

import redis.asyncio as redis_async
from redis.asyncio.connection import ConnectionPool
from redis.asyncio.sentinel import Sentinel as Sentinel_async

from app.common.config import Config


class RedisPool:
    """Redis连接池管理类

    实现为单例模式，确保全局只有一个连接池实例
    """

    # 单例实例
    _instance = None
    _pools: Dict[str, ConnectionPool] = {}

    def __new__(cls, *args, **kwargs):
        """实现单例模式"""
        if cls._instance is None:
            cls._instance = super(RedisPool, cls).__new__(cls)
            cls._instance._initialized = False
        return cls._instance

    def __init__(self):
        """初始化连接池配置

        只在第一次创建实例时执行初始化
        """
        if self._initialized:
            return

        self.redis_cluster_mode = Config.redis.cluster_mode
        self.redis_ip = Config.redis.host
        self.redis_read_ip = Config.redis.read_host
        self.redis_read_port = Config.redis.read_port
        self.redis_read_user = Config.redis.read_user
        self.redis_read_passwd = Config.redis.read_password
        self.redis_write_ip = Config.redis.write_host
        self.redis_write_port = Config.redis.write_port
        self.redis_write_user = Config.redis.write_user
        self.redis_write_passwd = Config.redis.write_password
        self.redis_port = Config.redis.port
        self.redis_user = ""
        if Config.redis.user:
            self.redis_user = Config.redis.user
        self.redis_passwd = Config.redis.password
        self.redis_master_name = Config.redis.sentinel_master
        self.redis_sentinel_user = Config.redis.sentinel_user
        self.redis_sentinel_password = Config.redis.sentinel_password
        self._initialized = True

    def get_pool_key(self, db: int, model: str) -> str:
        """生成连接池的key"""
        return f"{model}:{db}"

    async def get_pool(self, db: int, model: str) -> ConnectionPool:
        """获取或创建连接池"""
        pool_key = self.get_pool_key(db, model)
        if pool_key not in self._pools:
            if self.redis_cluster_mode == "sentinel":
                sentinel = Sentinel_async(
                    [(self.redis_ip, self.redis_port)],
                    password=self.redis_sentinel_password,
                    sentinel_kwargs={
                        "password": self.redis_sentinel_password,
                        "username": self.redis_sentinel_user,
                    },
                )
                if model == "write":
                    redis_con = sentinel.master_for(
                        self.redis_master_name,
                        username=self.redis_user,
                        password=self.redis_passwd,
                        db=db,
                    )
                else:
                    redis_con = sentinel.slave_for(
                        self.redis_master_name,
                        username=self.redis_user,
                        password=self.redis_passwd,
                        db=db,
                    )
                self._pools[pool_key] = redis_con.connection_pool
            elif self.redis_cluster_mode == "master-slave":
                if model == "read":
                    self._pools[pool_key] = ConnectionPool(
                        host=self.redis_read_ip,
                        port=self.redis_read_port,
                        db=db,
                        password=self.redis_read_passwd,
                    )
                else:
                    self._pools[pool_key] = ConnectionPool(
                        host=self.redis_write_ip,
                        port=self.redis_write_port,
                        db=db,
                        password=self.redis_write_passwd,
                    )
            else:
                # 不支持的集群模式
                raise ValueError(f"不支持的Redis集群模式: {self.redis_cluster_mode}")
        return self._pools[pool_key]

    @asynccontextmanager
    async def acquire(self, db: int, model: str):
        """获取Redis连接的上下文管理器"""
        pool = await self.get_pool(db, model)
        conn = redis_async.Redis(connection_pool=pool)
        try:
            yield conn
        finally:
            await conn.close()  # 将连接返回到连接池，而不是真正关闭

    async def close_all(self):
        """关闭所有连接池"""
        for pool in self._pools.values():
            await pool.disconnect()
        self._pools.clear()


class RedisLock:
    """Redis分布式锁实现

    Example:
        ```python
        async with RedisLock('my_lock') as lock:
            # 获取锁成功，执行操作
            ...
        # 锁已自动释放
        ```
    """

    def __init__(self, key: str, expire: int = 30):
        """初始化Redis锁

        Args:
            key: 锁的key
            expire: 锁过期时间，单位秒
        """
        self.key = key
        self.expire = expire
        self.redis_pool = RedisPool()

    async def acquire(self) -> bool:
        """获取锁"""
        async with self.redis_pool.acquire(0, "write") as conn:
            while True:
                # 尝试获取锁
                acquired = await conn.setnx(self.key, time.time() + self.expire)
                if acquired:
                    return True

                # 检查锁是否已过期
                expiration = await conn.get(self.key)
                if not expiration:
                    continue

                if float(expiration.decode("utf-8")) < time.time():
                    # 锁已过期，尝试重新获取
                    new_expiration = time.time() + self.expire
                    old_expiration = await conn.getset(self.key, new_expiration)
                    if (
                        old_expiration
                        and float(old_expiration.decode("utf-8")) < time.time()
                    ):
                        continue
                    return True

                await asyncio.sleep(0.1)  # 避免CPU高负载的重试等待

    async def release(self):
        """释放锁"""
        async with self.redis_pool.acquire(0, "write") as conn:
            await conn.delete(self.key)

    async def __aenter__(self):
        await self.acquire()
        return self

    async def __aexit__(self, exc_type, exc_value, traceback):
        await self.release()


# 全局Redis连接池实例
redis_pool = RedisPool()
