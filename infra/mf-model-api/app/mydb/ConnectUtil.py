from confluent_kafka import Producer, Consumer
from confluent_kafka.admin import AdminClient, NewTopic

from app.core.config import base_config
from app.logs.stand_log import StandLogger

import redis
import redis.asyncio as redis_async
from redis.asyncio.sentinel import Sentinel as Sentinel_async
from redis.sentinel import Sentinel
from typing import Union, Dict, List
import asyncio
import concurrent.futures
import threading
from queue import Queue, Empty
import time


class RedisClient(object):
    _instance = None
    _lock = asyncio.Lock()

    def __new__(cls):
        if not cls._instance:
            cls._instance = super(RedisClient, cls).__new__(cls)
        return cls._instance

    def __init__(self):
        if not hasattr(self, '_initialized'):
            self.redis_cluster_mode = base_config.REDISCLUSTERMODE
            self.redis_ip = base_config.REDISHOST
            self.redis_read_ip = base_config.REDISREADHOST
            self.redis_read_port = base_config.REDISREADPORT
            self.redis_read_user = base_config.REDISREADUSER
            self.redis_read_passwd = base_config.REDISREADPASS
            self.redis_write_ip = base_config.REDISWRITEHOST
            self.redis_write_port = base_config.REDISWRITEPORT
            self.redis_write_user = base_config.REDISWRITEUSER
            self.redis_write_passwd = base_config.REDISWRITEPASS
            self.redis_port = base_config.REDISPORT
            self.redis_user = ""
            if base_config.REDISUSER:
                self.redis_user = base_config.REDISUSER
            self.redis_passwd = base_config.REDISPASS
            self.redis_master_name = base_config.SENTINELMASTER
            self.redis_sentinel_user = base_config.SENTINELUSER
            self.redis_sentinel_password = base_config.SENTINELPASS
            self._initialized = True

    async def connect_redis(self, db, model):
        assert model in ["read", "write"]
        if self.redis_cluster_mode == "sentinel":
            try:
                sentinel = Sentinel_async([(self.redis_ip, self.redis_port)], password=self.redis_sentinel_password,
                                          sentinel_kwargs={
                                              "password": self.redis_sentinel_password,
                                              "username": self.redis_sentinel_user
                                          })
                if model == "write":
                    redis_con = await sentinel.master_for(self.redis_master_name, username=self.redis_user,
                                                          password=self.redis_passwd, db=db)
                if model == "read":
                    redis_con = await sentinel.slave_for(self.redis_master_name, username=self.redis_user,
                                                         password=self.redis_passwd, db=db)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - sentinel模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")
        elif self.redis_cluster_mode == "master-slave":
            try:
                if model == "read":
                    pool = redis_async.ConnectionPool(host=self.redis_read_ip, port=self.redis_read_port, db=db,
                                                      password=self.redis_read_passwd)
                    redis_con = redis_async.StrictRedis(connection_pool=pool)
                if model == "write":
                    pool = redis_async.ConnectionPool(host=self.redis_write_ip, port=self.redis_write_port, db=db,
                                                      password=self.redis_write_passwd)
                    redis_con = redis_async.StrictRedis(connection_pool=pool)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - master-slave模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")
        else:
            # standalone 单机模式
            try:
                pool = redis_async.ConnectionPool(host=self.redis_ip, port=self.redis_port, db=db,
                                                  password=self.redis_passwd, username=self.redis_user if self.redis_user else None)
                redis_con = redis_async.StrictRedis(connection_pool=pool)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - standalone模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")

    async def connect_redis_async(self, db, model):
        assert model in ["read", "write"]
        if self.redis_cluster_mode == "sentinel":
            try:
                sentinel = Sentinel_async([(self.redis_ip, self.redis_port)], password=self.redis_sentinel_password,
                                          sentinel_kwargs={
                                              "password": self.redis_sentinel_password,
                                              "username": self.redis_sentinel_user
                                          })
                if model == "write":
                    redis_con = await sentinel.master_for(self.redis_master_name, username=self.redis_user,
                                                          password=self.redis_passwd, db=db)
                if model == "read":
                    redis_con = await sentinel.slave_for(self.redis_master_name, username=self.redis_user,
                                                         password=self.redis_passwd, db=db)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - {model}模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")
        elif self.redis_cluster_mode == "master-slave":
            try:
                if model == "read":
                    pool = redis_async.ConnectionPool(host=self.redis_read_ip, port=self.redis_read_port, db=db,
                                                      password=self.redis_read_passwd)
                    redis_con = redis_async.StrictRedis(connection_pool=pool)
                if model == "write":
                    pool = redis_async.ConnectionPool(host=self.redis_write_ip, port=self.redis_write_port, db=db,
                                                      password=self.redis_write_passwd)
                    redis_con = redis_async.StrictRedis(connection_pool=pool)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - {model}模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")
        else:
            # standalone 单机模式
            try:
                pool = redis_async.ConnectionPool(host=self.redis_ip, port=self.redis_port, db=db,
                                                  password=self.redis_passwd, username=self.redis_user if self.redis_user else None)
                redis_con = redis_async.StrictRedis(connection_pool=pool)
                # 验证连接
                await redis_con.ping()
                return redis_con
            except Exception as e:
                StandLogger.error(f"Redis连接失败 - standalone模式: host={self.redis_ip}, port={self.redis_port}, error={str(e)}")
                raise Exception(f"connect redis error:{str(e)}")


class ConnectUtil:
    _instance = None
    _lock = asyncio.Lock()
    _redis_client = None
    _read_pool = None
    _write_pool = None

    @classmethod
    async def create(cls, db=1):
        async with cls._lock:
            if cls._instance is None:
                instance = cls(db)
                if cls._redis_client is None:
                    cls._redis_client = RedisClient()
                await instance._init_connection_pools()
                if instance.read_conn is None or instance.write_conn is None:
                    StandLogger.error(f"Redis连接创建失败: read_conn={instance.read_conn}, write_conn={instance.write_conn}")
                    raise Exception("Redis连接创建失败")
                instance._initialized = True
                StandLogger.info(f"Redis连接池创建成功: read_conn={instance.read_conn}, write_conn={instance.write_conn}")
                cls._instance = instance
            return cls._instance

    def __init__(self, db=1):
        self.db = db
        self.read_conn = None
        self.write_conn = None
        self._connection_healthy = True
        self._last_health_check = 0

    async def _init_connection_pools(self):
        """初始化高性能Redis连接池"""
        try:
            redis_client = self.__class__._redis_client
            
            if redis_client.redis_cluster_mode == "sentinel":
                # 哨兵模式：使用哨兵连接
                await self._init_sentinel_connection_pools(redis_client)
            elif redis_client.redis_cluster_mode == "master-slave":
                # 主从模式：使用主从连接
                await self._init_master_slave_connection_pools(redis_client)
            else:
                # 默认模式：使用单机连接
                await self._init_default_connection_pools(redis_client)
            
            # 预热连接池 - 建立初始连接
            await self._warmup_connections()
            
            StandLogger.info("Redis高性能连接池初始化成功")
            
        except Exception as e:
            StandLogger.error(f"Redis连接池初始化失败: {str(e)}")
            raise e

    async def _init_sentinel_connection_pools(self, redis_client):
        """初始化哨兵模式的连接池"""
        try:
            # 创建哨兵连接
            sentinel = Sentinel_async([(redis_client.redis_ip, redis_client.redis_port)], 
                                    password=redis_client.redis_sentinel_password,
                                    sentinel_kwargs={
                                        "password": redis_client.redis_sentinel_password,
                                        "username": redis_client.redis_sentinel_user
                                    })
            
            # 获取主节点连接（写操作）
            master_redis = await sentinel.master_for(
                redis_client.redis_master_name, 
                username=redis_client.redis_user,
                password=redis_client.redis_passwd, 
                db=self.db,
                max_connections=30,
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_connect_timeout=3,
                socket_timeout=2,
                health_check_interval=30,
                decode_responses=False
            )
            
            # 获取从节点连接（读操作）
            slave_redis = await sentinel.slave_for(
                redis_client.redis_master_name, 
                username=redis_client.redis_user,
                password=redis_client.redis_passwd, 
                db=self.db,
                max_connections=50,
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_connect_timeout=3,
                socket_timeout=2,
                health_check_interval=30,
                decode_responses=False
            )
            
            self.write_conn = master_redis
            self.read_conn = slave_redis
            
            StandLogger.info("哨兵模式Redis连接池初始化成功")
            
        except Exception as e:
            StandLogger.error(f"哨兵模式Redis连接池初始化失败: {str(e)}")
            raise e

    async def _init_master_slave_connection_pools(self, redis_client):
        """初始化主从模式的连接池"""
        try:
            # 创建读连接池 - 高性能配置
            self.__class__._read_pool = redis_async.ConnectionPool(
                host=redis_client.redis_read_ip,
                port=redis_client.redis_read_port,
                db=self.db,
                password=redis_client.redis_read_passwd,
                max_connections=50,  # 最大连接数
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_connect_timeout=3,  # 减少连接超时时间
                socket_timeout=2,  # 减少操作超时时间
                health_check_interval=30,  # 健康检查间隔
                decode_responses=False  # 保持二进制格式，提高性能
            )
            
            # 创建写连接池 - 高性能配置
            self.__class__._write_pool = redis_async.ConnectionPool(
                host=redis_client.redis_write_ip,
                port=redis_client.redis_write_port,
                db=self.db,
                password=redis_client.redis_write_passwd,
                max_connections=30,  # 写操作连接数可以少一些
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_connect_timeout=3,  # 减少连接超时时间
                socket_timeout=2,  # 减少操作超时时间
                health_check_interval=30,
                decode_responses=False
            )
            
            # 创建Redis客户端实例
            self.read_conn = redis_async.StrictRedis(connection_pool=self.__class__._read_pool)
            self.write_conn = redis_async.StrictRedis(connection_pool=self.__class__._write_pool)
            
            StandLogger.info("主从模式Redis连接池初始化成功")
            
        except Exception as e:
            StandLogger.error(f"主从模式Redis连接池初始化失败: {str(e)}")
            raise e

    async def _init_default_connection_pools(self, redis_client):
        """初始化默认模式的连接池（单机模式）"""
        try:
            # 创建连接池 - 高性能配置
            self.__class__._read_pool = redis_async.ConnectionPool(
                host=redis_client.redis_ip,
                port=redis_client.redis_port,
                db=self.db,
                password=redis_client.redis_passwd,
                max_connections=50,  # 最大连接数
                retry_on_timeout=True,
                socket_keepalive=True,
                socket_connect_timeout=3,  # 减少连接超时时间
                socket_timeout=2,  # 减少操作超时时间
                health_check_interval=30,  # 健康检查间隔
                decode_responses=False  # 保持二进制格式，提高性能
            )
            
            # 创建Redis客户端实例（读写使用同一个连接池）
            self.read_conn = redis_async.StrictRedis(connection_pool=self.__class__._read_pool)
            self.write_conn = redis_async.StrictRedis(connection_pool=self.__class__._read_pool)
            
            StandLogger.info("默认模式Redis连接池初始化成功")
            
        except Exception as e:
            StandLogger.error(f"默认模式Redis连接池初始化失败: {str(e)}")
            raise e

    async def _warmup_connections(self):
        """预热连接池，建立初始连接"""
        try:
            redis_client = self.__class__._redis_client
            
            if redis_client.redis_cluster_mode == "sentinel":
                # 哨兵模式：直接ping连接
                await self.read_conn.ping()
                await self.write_conn.ping()
            else:
                # 主从模式和默认模式：预热连接池
                if self.__class__._read_pool:
                    for _ in range(min(5, self.__class__._read_pool.max_connections)):
                        await self.read_conn.ping()
                
                if self.__class__._write_pool:
                    for _ in range(min(3, self.__class__._write_pool.max_connections)):
                        await self.write_conn.ping()
                
            StandLogger.info("Redis连接池预热完成")
        except Exception as e:
            StandLogger.warn(f"Redis连接池预热失败: {str(e)}")

    async def _check_connection_health(self):
        """检查连接健康状态"""
        current_time = time.time()
        # 每30秒检查一次连接健康状态
        if current_time - self._last_health_check > 30:
            try:
                await self.read_conn.ping()
                await self.write_conn.ping()
                self._connection_healthy = True
                self._last_health_check = current_time
            except Exception as e:
                StandLogger.warn(f"Redis连接健康检查失败: {str(e)}")
                self._connection_healthy = False
                # 不立即重连，让业务逻辑决定是否重连
        return self._connection_healthy

    async def _reconnect(self):
        """重新建立redis连接 - 仅在必要时调用"""
        if self._connection_healthy:
            return
            
        StandLogger.info("开始重新建立Redis连接...")
        old_read_conn = self.read_conn
        old_write_conn = self.write_conn
        
        try:
            # 重新初始化连接池
            await self._init_connection_pools()
            self._connection_healthy = True
            StandLogger.info("Redis连接重建成功")
            
            # 关闭旧连接
            if old_read_conn:
                try:
                    await old_read_conn.close()
                except:
                    pass
            if old_write_conn:
                try:
                    await old_write_conn.close()
                except:
                    pass
                    
        except Exception as e:
            StandLogger.error(f"Redis重连失败: {str(e)}")
            # 如果重连失败，恢复旧连接
            self.read_conn = old_read_conn
            self.write_conn = old_write_conn
            raise e

    async def close(self):
        """关闭Redis连接"""
        if self.read_conn:
            await self.read_conn.close()
        if self.write_conn:
            await self.write_conn.close()
        self.read_conn = None
        self.write_conn = None

    async def set_str(self, key: str, value: Union[str, int, float], expire: int = None) -> bool:
        for i in range(3):
            try:
                result = await self.write_conn.set(key, value)
                return result and await self._set_expire(key, expire)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def set_if_absent(self, key: str, value: Union[str, int, float] = 1, expire: int | None = None) -> bool:
        """SETNX 语义，已存在返回 False；设置成功可选过期时间"""
        for i in range(3):
            try:
                # Redis-py: set(name, value, ex=None, px=None, nx=False, xx=False)
                result = await self.write_conn.set(key, value, ex=expire, nx=True)
                return bool(result)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def _set_expire(self, key: str, expire: int = None) -> bool:
        """内部方法：设置key过期时间"""
        if expire is not None:
            return await self.write_conn.expire(key, expire)
        return True

    async def lpush(self, key: str, *values: Union[str, int, float], expire: int = None) -> int:
        for i in range(3):
            try:
                result = await self.write_conn.lpush(key, *values)
                await self._set_expire(key, expire)
                return result
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def rpush(self, key: str, *values: Union[str, int, float], expire: int = None) -> int:
        for i in range(3):
            try:
                result = await self.write_conn.rpush(key, *values)
                await self._set_expire(key, expire)
                return result
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def lrem(self, key: str, count: int, value: str, expire: int = None) -> int:
        for i in range(3):
            try:
                result = await self.write_conn.lrem(key, count, value)
                await self._set_expire(key, expire)
                return result
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def hset(self, key: str, field: str, value: Union[str, int, float], expire: int = None) -> int:
        for i in range(3):
            try:
                result = await self.write_conn.hset(key, field, value)
                await self._set_expire(key, expire)
                return result
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def hdel(self, key: str, *fields: str, expire: int = None) -> int:
        for i in range(3):
            try:
                result = await self.write_conn.hdel(key, *fields)
                await self._set_expire(key, expire)
                return result
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def get_str(self, key: str) -> Union[str, None]:
        """高性能Redis GET操作，带智能重试和连接健康检查"""
        # 先检查连接健康状态
        await self._check_connection_health()
        
        for i in range(3):
            try:
                # 使用更短的超时时间，避免长时间等待
                result = await asyncio.wait_for(
                    self.read_conn.get(key), 
                    timeout=2.0  # 2秒超时
                )
                return result
            except asyncio.TimeoutError:
                StandLogger.warn(f"Redis GET操作超时 (尝试 {i+1}/3): key={key}")
                if i < 2:
                    self._connection_healthy = False
                    await asyncio.sleep(0.1 * (i + 1))
                    await self._reconnect()
                else:
                    StandLogger.error(f"Redis GET操作最终超时: key={key}")
                    raise Exception(f"Redis GET操作超时: {key}")
            except Exception as e:
                StandLogger.warn(f"Redis GET操作失败 (尝试 {i+1}/3): {str(e)}")
                if i < 2:
                    # 标记连接不健康
                    self._connection_healthy = False
                    await asyncio.sleep(0.1 * (i + 1))  # 递增延迟
                    await self._reconnect()
                else:
                    StandLogger.error(f"Redis GET操作最终失败: key={key}, error={str(e)}")
                    raise e

    async def delete_str(self, key: str | list[str]) -> bool:
        """
        删除一个或多个 Redis key
        :param key: 单个 key 或 key 列表
        :return: 是否删除成功
        """
        if isinstance(key, str):
            key = [key]
        for i in range(3):
            try:
                # 使用 pipeline 批量删除
                async with self.write_conn.pipeline(transaction=True) as pipe:
                    for k in key:
                        if await self.exists(k):
                            pipe.delete(k)
                    results = await pipe.execute()
                # 只要有一个 key 删除成功就返回 True
                return any(result > 0 for result in results)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def lrange(self, key: str, start: int = 0, end: int = -1) -> List:
        for i in range(3):
            try:
                return await self.read_conn.lrange(key, start, end)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def hget(self, key: str, field: str) -> Union[str, None]:
        for i in range(3):
            try:
                return await self.read_conn.hget(key, field)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def hgetall(self, key: str) -> Dict:
        for i in range(3):
            try:
                return await self.read_conn.hgetall(key)
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e

    async def exists(self, key: str) -> bool:
        for i in range(3):
            try:
                return await self.read_conn.exists(key) > 0
            except Exception as e:
                if i < 2:
                    import time
                    await asyncio.sleep(0.1)
                    await self._reconnect()
                else:
                    raise e


async def get_redis_util():
    """
    获取已初始化的redis_util实例
    """
    global redis_util
    if redis_util is None:

        try:
            redis_util = await ConnectUtil.create()
            StandLogger.info("Redis连接工厂初始化成功")
        except Exception as e:
            StandLogger.error(f"Redis连接工厂初始化失败: {str(e)}")
            raise e
    return redis_util

class MyKafkaClient(object):
    def __init__(self, topic_name='tenant_a.dip.model_manager.quota_data'):
        self.producer = None
        self.consumer = None
        self.topic_name = topic_name
        self.admin_client = None
        
        # 异步发送相关属性
        self._message_queue = Queue(maxsize=10000)  # 消息队列，限制大小防止内存溢出
        self._producer_thread = None
        self._shutdown_event = threading.Event()
        self._thread_pool = concurrent.futures.ThreadPoolExecutor(max_workers=2, thread_name_prefix="kafka-producer")
        self._is_async_running = False
        
        # 初始化时检查并创建topic
        self._init_admin_client()
        self._check_and_create_topic()
        
        # 启动异步生产者线程
        self._start_async_producer()

    def _init_admin_client(self):
        """初始化AdminClient"""
        self.admin_client = AdminClient({
            'bootstrap.servers': '{}:{}'.format(base_config.KAFKAHOST, base_config.KAFKAPORT),
            'security.protocol': 'sasl_plaintext',
            'enable.ssl.certificate.verification': 'false',
            'sasl.mechanism': 'PLAIN',
            'sasl.username': base_config.KAFKAUSER,
            'sasl.password': base_config.KAFKAPASS,
        })

    def _check_and_create_topic(self):
        """检查topic是否存在，如果不存在则创建"""
        try:
            # 获取集群元数据
            metadata = self.admin_client.list_topics(timeout=10)
            
            # 检查topic是否存在
            if self.topic_name not in metadata.topics:
                # 创建topic
                new_topic = NewTopic(
                    topic=self.topic_name,
                    num_partitions=3,  # 设置分区数为3
                    replication_factor=1  # 设置副本因子为1
                )
                
                # 创建topic
                fs = self.admin_client.create_topics([new_topic])
                for topic, f in fs.items():
                    try:
                        f.result()  # 等待创建完成
                        StandLogger.info(f"Topic {topic} created successfully")
                    except Exception as e:
                        StandLogger.error(f"Failed to create topic {topic}: {e}")
            else:
                StandLogger.info(f"Topic {self.topic_name} already exists")
        except Exception as e:
            StandLogger.error(f"Error checking/creating topic: {e}")

    def connect_producer(self):
        """连接生产者"""
        if self.producer is None:
            self.producer = Producer({
                'bootstrap.servers': '{}:{}'.format(base_config.KAFKAHOST, base_config.KAFKAPORT),
                'security.protocol': 'sasl_plaintext',
                'enable.ssl.certificate.verification': 'false',
                'sasl.mechanism': 'PLAIN',
                'sasl.username': base_config.KAFKAUSER,
                'sasl.password': base_config.KAFKAPASS,
                'acks': 1,  # 确认leader分区收到消息
                'queue.buffering.max.messages': 100000,  # 增加队列缓冲区大小
                'queue.buffering.max.kbytes': 102400,   # 增加队列缓冲区大小(KB)
                'batch.num.messages': 1000,             # 批量发送的消息数量
                'linger.ms': 5                          # 批量发送等待时间(ms)
            })

    def connect_consumer(self, group_id='quota_data_group'):
        """连接消费者"""
        if self.consumer is None:
            self.consumer = Consumer({
                'bootstrap.servers': '{}:{}'.format(base_config.KAFKAHOST, base_config.KAFKAPORT),
                'security.protocol': 'sasl_plaintext',
                'enable.ssl.certificate.verification': 'false',
                'sasl.mechanism': 'PLAIN',
                'sasl.username': base_config.KAFKAUSER,
                'sasl.password': base_config.KAFKAPASS,
                'group.id': group_id,
                'auto.offset.reset': 'latest',  # 改为latest或使用一个新的消费者组ID
                'enable.auto.commit': True,
                'auto.commit.interval.ms': 1000
            })
            # 订阅topic
            self.consumer.subscribe([self.topic_name])

    def _start_async_producer(self):
        """启动异步生产者线程"""
        if not self._is_async_running:
            self._producer_thread = threading.Thread(
                target=self._async_producer_worker,
                name="kafka-async-producer",
                daemon=True
            )
            self._producer_thread.start()
            self._is_async_running = True
            StandLogger.info("Kafka异步生产者线程已启动")

    def _async_producer_worker(self):
        """异步生产者工作线程"""
        # 在独立线程中初始化producer
        self.connect_producer()
        
        while not self._shutdown_event.is_set():
            try:
                # 非阻塞获取消息，超时1秒
                try:
                    message_data = self._message_queue.get(timeout=1.0)
                except Empty:
                    continue
                
                # 发送消息
                self.producer.produce(
                    self.topic_name,
                    key=message_data.get('key'),
                    value=message_data.get('value'),
                    callback=message_data.get('callback', self._delivery_callback)
                )
                
                # 批量轮询，提高效率
                self.producer.poll(0.1)
                
                # 标记任务完成
                self._message_queue.task_done()
                
            except Exception as e:
                StandLogger.error(f"异步生产者工作线程错误: {e}")
                time.sleep(0.1)  # 错误时短暂休眠

    def produce_async(self, value, key=None, callback=None):
        """真正异步非阻塞发送消息到Kafka"""
        try:
            # 将消息放入队列，非阻塞
            message_data = {
                'value': value,
                'key': key,
                'callback': callback
            }
            
            # 非阻塞放入队列，如果队列满了则记录警告但不阻塞
            try:
                self._message_queue.put_nowait(message_data)
            except:
                StandLogger.warn(f"Kafka消息队列已满，丢弃消息: key={key}")
                return False
                
            return True
            
        except Exception as e:
            StandLogger.error(f"异步发送Kafka消息失败: {e}")
            return False

    def produce_async_with_future(self, value, key=None):
        """异步发送消息并返回Future对象，可以获取发送结果"""
        future = concurrent.futures.Future()
        
        def delivery_callback(err, msg):
            if err:
                future.set_exception(Exception(f"Kafka消息发送失败: {err}"))
            else:
                future.set_result({
                    'topic': msg.topic(),
                    'partition': msg.partition(),
                    'offset': msg.offset()
                })
        
        success = self.produce_async(value, key, delivery_callback)
        if not success:
            future.set_exception(Exception("消息队列已满，无法发送"))
        
        return future

    def _delivery_callback(self, err, msg):
        """消息发送回调函数"""
        if err:
            StandLogger.error(f'Message delivery failed: {err}')
        else:
            StandLogger.info(f'Message delivered to {msg.topic()} [{msg.partition()}] at offset {msg.offset()}')

    def flush_producer(self, timeout=10):
        """刷新生产者队列，确保所有消息都被发送"""
        if self.producer is not None:
            return self.producer.flush(timeout)
        return 0

    def consume_messages(self, timeout=1.0):
        """从Kafka消费消息"""
        if self.consumer is None:
            self.connect_consumer()
        
        try:
            # 轮询消息
            msg = self.consumer.poll(timeout)
            
            if msg is None:
                return None
            
            if msg.error():
                StandLogger.error(f'Consumer error: {msg.error()}')
                return None
            
            # 返回消息内容
            return {
                'key': msg.key(),
                'value': msg.value(),
                'topic': msg.topic(),
                'partition': msg.partition(),
                'offset': msg.offset()
            }
        except Exception as e:
            StandLogger.error(f'Error consuming message: {e}')
            return None

    def consume_batch(self, num_messages: int = 500, timeout: float = 1.0):
        """批量消费消息，返回统一结构的列表"""
        if self.consumer is None:
            self.connect_consumer()
        try:
            msgs = self.consumer.consume(num_messages=num_messages, timeout=timeout)
            result = []
            for msg in msgs or []:
                if msg is None:
                    continue
                if msg.error():
                    StandLogger.error(f'Consumer error: {msg.error()}')
                    continue
                result.append({
                    'key': msg.key(),
                    'value': msg.value(),
                    'topic': msg.topic(),
                    'partition': msg.partition(),
                    'offset': msg.offset()
                })
            return result
        except Exception as e:
            StandLogger.error(f'Error consuming batch messages: {e}')
            return []

    def close_producer(self):
        """关闭生产者连接"""
        if self.producer is not None:
            self.flush_producer()
            self.producer = None

    def close_consumer(self):
        """关闭消费者连接"""
        if self.consumer is not None:
            self.consumer.close()
            self.consumer = None

    def shutdown_async_producer(self):
        """优雅关闭异步生产者"""
        if self._is_async_running:
            StandLogger.info("正在关闭Kafka异步生产者...")
            self._shutdown_event.set()
            
            # 等待队列中的消息处理完成
            self._message_queue.join()
            
            # 等待线程结束
            if self._producer_thread and self._producer_thread.is_alive():
                self._producer_thread.join(timeout=5)
            
            # 关闭线程池
            self._thread_pool.shutdown(wait=True)
            
            # 最后刷新producer
            if self.producer is not None:
                self.flush_producer()
                self.producer = None
            
            self._is_async_running = False
            StandLogger.info("Kafka异步生产者已关闭")

    def get_queue_status(self):
        """获取队列状态信息"""
        return {
            'queue_size': self._message_queue.qsize(),
            'queue_maxsize': self._message_queue.maxsize,
            'is_async_running': self._is_async_running,
            'thread_alive': self._producer_thread.is_alive() if self._producer_thread else False
        }

    def get_redis_pool_status(self):
        """获取Redis连接池状态信息"""
        if not self._instance:
            return {'error': 'Redis连接池未初始化'}
        
        status = {
            'connection_healthy': self._instance._connection_healthy,
            'last_health_check': self._instance._last_health_check,
            'read_pool': {},
            'write_pool': {}
        }
        
        if self._instance.__class__._read_pool:
            status['read_pool'] = {
                'max_connections': self._instance.__class__._read_pool.max_connections,
                'created_connections': len(self._instance.__class__._read_pool._created_connections),
                'available_connections': len(self._instance.__class__._read_pool._available_connections),
                'in_use_connections': len(self._instance.__class__._read_pool._in_use_connections)
            }
        
        if self._instance.__class__._write_pool:
            status['write_pool'] = {
                'max_connections': self._instance.__class__._write_pool.max_connections,
                'created_connections': len(self._instance.__class__._write_pool._created_connections),
                'available_connections': len(self._instance.__class__._write_pool._available_connections),
                'in_use_connections': len(self._instance.__class__._write_pool._in_use_connections)
            }
        
        return status

# 全局redis_util实例
redis_util = None
kafka_client = MyKafkaClient()