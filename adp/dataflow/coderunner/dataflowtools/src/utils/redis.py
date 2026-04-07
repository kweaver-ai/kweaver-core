from contextlib import contextmanager
import threading
import time
import uuid
import redis
from redis.sentinel import Sentinel
from redis.cluster import ClusterNode, RedisCluster
from typing import Any, Dict, Optional, Union
from common.configs import RedisMode
from common.logger import logger
from common.configs import redis_configs


class RedisConnectionManager:
    """
    Redis 连接管理器 - 单例模式+自动重连
    """
    
    _instance = None
    _lock = threading.Lock()
    
    def __new__(cls):
        """单例模式实现"""
        if cls._instance is None:
            with cls._lock:
                if cls._instance is None:
                    cls._instance = super().__new__(cls)
                    cls._instance._initialized = False
        return cls._instance
    
    def __init__(self):
        """
        初始化 Redis 连接管理器
        
        :param config: Redis 配置字典
        """
        # 防止重复初始化
        if self._initialized:
            return
        
        self.config = redis_configs
        self.mode = RedisMode(redis_configs.get('redis_cluster_mode'))
        self.client_master: Optional[Union[redis.Redis, RedisCluster]] = None
        self.client_slave: Optional[Union[redis.Redis, RedisCluster]] = None
        self._connect_lock = threading.Lock()
        
        # 守护线程配置
        self.check_interval = 30  # 默认30秒检测一次
        self.max_retries = 10  # 重连最大次数
        self.retry_interval = 3  # 重连间隔
        
        # 守护线程
        self._monitor_thread: Optional[threading.Thread] = None
        self._stop_monitor = threading.Event()
        
        # 初始化连接
        self._connect()
        self._initialized = True
        
        # 启动守护线程
        self._start_monitor()
    
    def _connect(self):
        """创建连接"""
        try:
            if self.mode == RedisMode.STANDALONE:
                self._create_standalone_connection()
            elif self.mode == RedisMode.SENTINEL:
                self._create_sentinel_connection()
            elif self.mode == RedisMode.CLUSTER:
                self._create_cluster_connection()
            elif self.mode == RedisMode.MASTER_SLAVE:
                self._create_master_slave_connection()
            else:
                raise ValueError(f"unsupported Redis mode: {self.mode}")
            
            # 验证连接
            self.client_master.ping()
        except Exception as e:
            logger.warning(f"Redis connection failed: {str(e)}")
            raise
    
    def _create_standalone_connection(self):
        """创建单机连接"""      
        conn_pool = redis.ConnectionPool(
            host=self.config.get('redis_host'),
            port=self.config.get('redis_port'),
            username=self.config.get('redis_user'),
            password=self.config.get('redis_password'),
            decode_responses=True
        )

        client = redis.Redis(connection_pool=conn_pool)
        self.client_master = client
        self.client_slave = client
    
    def _create_sentinel_connection(self):
        """创建哨兵连接"""
        sentinels = [(self.config.get("redis_host"), self.config.get("redis_port"))]
        
        sentinel_kwargs = {
            'username': self.config.get('redis_sentinel_user'),
            'password': self.config.get('redis_sentinel_password'),
            'socket_timeout': 0.1,
        }
        
        sentinel = Sentinel(sentinels, sentinel_kwargs=sentinel_kwargs)
        
        connection_kwargs = {
            'username': self.config.get('redis_user'),
            'password': self.config.get('redis_password'),
            'socket_timeout': 0.1
        }
        master_name = self.config.get('redis_master_name', 'mymaster')
        
        self.client_master = sentinel.master_for(master_name, **connection_kwargs)
        self.client_slave = sentinel.slave_for(master_name, **connection_kwargs)
    
    def _create_cluster_connection(self):
        """创建集群连接"""
        startup_nodes = []
        hosts = [host.strip() for host in self.config["redis_host"].split(",")]
        for host in hosts:
            if ":" in host:
                startup_nodes.append(ClusterNode(host.split(":")[0], host.split(":")[1]))
            else:
                startup_nodes.append(ClusterNode(host.split(":")[0], self.config["redis_port"]))

        params = {
            'startup_nodes': startup_nodes,
            'username': self.config.get('redis_user'),
            'password': self.config.get('redis_password'),
        }
       
        client = RedisCluster(**params)
        self.client_master = client
        self.client_slave = client
    
    def _create_master_slave_connection(self):
        """创建主从连接"""
        conn_pool = redis.ConnectionPool(
            host=self.config.get('redis_host'),
            port=self.config.get('redis_port'),
            username=self.config.get('redis_user'),
            password=self.config.get('redis_password'),
            decode_responses=True,
        )

        client = redis.Redis(connection_pool=conn_pool)
        self.client_master = client
        self.client_slave = client
    
    def _start_monitor(self):
        """启动守护线程"""
        def monitor_worker():
            logger.info(f"Redis daemon thread started. {self.check_interval} seconds check interval.")
            
            while not self._stop_monitor.is_set():
                try:
                    # 调用 ping 检测连接
                    if self.client_master:
                        self.client_master.ping()
                except Exception as e:
                    # ping 失败，执行重连
                    logger.warning(f"redis monitor failed: {str(e)}, reconnecting...")
                    self._reconnect()
                
                # 休眠指定时间后再检测
                self._stop_monitor.wait(self.check_interval)
            
            logger.info("Redis damon thread stopped.")
        
        self._monitor_thread = threading.Thread(
            target=monitor_worker,
            daemon=True,
            name="RedisMonitor"
        )
        self._monitor_thread.start()
    
    def _reconnect(self):
        """重连逻辑"""
        with self._connect_lock:
            for attempt in range(1, self.max_retries + 1):
                try:
                    # 关闭旧连接
                    if self.client_master:
                        try:
                            self.client_master.close()
                        except:
                            pass
                    
                    # 重新连接
                    self._connect()
                    logger.info(f"Redis reconnected.")
                    return
                    
                except Exception as e:
                    logger.warning(f"Redis reconnection failed, retrying times: {attempt}, detail: {str(e)}")
                    if attempt < self.max_retries:
                        time.sleep(self.retry_interval)
            
            logger.warning(f"Redis reconnection failed after {self.max_retries} attempts.")

    def close(self):
        """关闭连接"""        
        # 停止守护线程
        self._stop_monitor.set()
        if self._monitor_thread and self._monitor_thread.is_alive():
            self._monitor_thread.join(timeout=3)
        
        # 关闭客户端
        if self.client_master:
            try:
                self.client_master.close()
            except:
                pass
            self.client_master = None
    


_redis_manager: Optional[RedisConnectionManager] = None

def init_redis():
    """初始化全局 Redis 实例"""
    global _redis_manager
    _redis_manager = RedisConnectionManager()


def get_redis() -> RedisConnectionManager:
    """获取全局 Redis 实例"""
    if _redis_manager is None:
        raise RuntimeError("Redis manager is not initialized.")
    return _redis_manager


class RedisLock:

    # Lua 脚本：安全释放锁
    UNLOCK_SCRIPT = """
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("del", KEYS[1])
    else
        return 0
    end
    """
    
    # Lua 脚本：安全续期锁
    RENEW_SCRIPT = """
    if redis.call("get", KEYS[1]) == ARGV[1] then
        return redis.call("expire", KEYS[1], ARGV[2])
    else
        return 0
    end
    """

    def __init__(self):
        self.redis = get_redis().client_master
        self.timeout = 30

        # 注册 Lua 脚本
        self._unlock_script = self.redis.register_script(self.UNLOCK_SCRIPT)
        self._renew_script = self.redis.register_script(self.RENEW_SCRIPT)

        # 锁状态管理（key -> LockState）
        self._locks: Dict[str, 'LockState'] = {}
        self._locks_lock = threading.Lock()

    def acquire(
            self,
            key: str,
            val: Any,
            timeout: int = 30,
            blocking: bool = True,
            auto_renewal: bool = False,
            blocking_timeout: Optional[float] = None
        ) -> bool:
        """
        获取锁
        
        :param key: 锁的键名
        :param val: 锁的标识符
        :param timeout: 锁超时时间（秒）
        :param blocking: 是否阻塞等待
        :param auto_renewal: 是否自动续租
        :param blocking_timeout: 阻塞超时时间（秒）
        :return: 是否成功获取锁
        
        示例：
            lock_manager.acquire('my_lock', val='my_val', timeout=30, blocking=True, auto_renewal=True)
        """
        lock_key = f"lock:{key}"
        
        # 创建或获取锁状态
        with self._locks_lock:
            if lock_key not in self._locks:
                self._locks[lock_key] = LockState(lock_key, timeout)
            lock_state = self._locks[lock_key]
        
        # 生成新的标识符
        if val is not None:
            lock_state.identifier = val
        else:
            lock_state.identifier = str(uuid.uuid4())
        lock_state.timeout = timeout
        
        # 尝试获取锁
        if blocking:
            success = self._acquire_blocking(lock_state, auto_renewal, blocking_timeout)
        else:
            success = self._acquire_non_blocking(lock_state, auto_renewal)
        
        return success
    
    def _acquire_non_blocking(self, lock_state: 'LockState', auto_renewal: bool) -> bool:
        """非阻塞获取锁"""
        acquired = self.redis.set(
            lock_state.key,
            lock_state.identifier,
            nx=True,
            ex=lock_state.timeout
        )
        
        if acquired:
            lock_state.locked = True
            if auto_renewal:
                self._start_watchdog(lock_state)
            return True
        
        return False
    
    def _acquire_blocking(
        self,
        lock_state: 'LockState',
        auto_renewal: bool,
        blocking_timeout: Optional[float]
    ) -> bool:
        """阻塞获取锁"""
        start_time = time.time()
        
        while True:
            if self._acquire_non_blocking(lock_state, auto_renewal):
                return True
            
            if blocking_timeout is not None:
                if time.time() - start_time >= blocking_timeout:
                    return False
            
            time.sleep(0.1)
    
    def release(self, key: str) -> bool:
        """
        释放锁
        
        :param key: 锁的键名
        :return: 是否成功释放
        
        示例：
            lock_manager.release('my_lock')
        """
        lock_key = f"lock:{key}"
        
        with self._locks_lock:
            if lock_key not in self._locks:
                return False
            lock_state = self._locks[lock_key]
        
        if not lock_state.locked:
            return False
        
        # 停止看门狗
        self._stop_watchdog(lock_state)
        
        try:
            # 使用 Lua 脚本安全释放
            result = self._unlock_script(
                keys=[lock_state.key],
                args=[lock_state.identifier]
            )
            
            if result:
                del self._locks[lock_state.key]
                return True
            return False
            
        except Exception:
            return False
    
    @contextmanager
    def lock(
        self,
        key: str,
        val: Any = None,
        timeout: int = 30,
        blocking: bool = True,
        auto_renewal: bool = False,
        blocking_timeout: Optional[float] = None
    ):
        """
        上下文管理器方式使用锁
        
        :param key: 锁的键名
        :param val: 锁的标识符
        :param timeout: 锁超时时间（秒）
        :param blocking: 是否阻塞等待
        :param auto_renewal: 是否自动续租
        :param blocking_timeout: 阻塞超时时间（秒）
        
        示例：
            with lock_manager.lock('my_lock', val='my_val', timeout=30, auto_renewal=True):
                # 业务逻辑
                pass
        """
        if not self.acquire(key, val, timeout, blocking, auto_renewal, blocking_timeout):
            raise RuntimeError(f"Failed to acquire lock: {key}")
        
        try:
            yield
        finally:
            self.release(key)
    
    def _start_watchdog(self, lock_state: 'LockState'):
        """启动看门狗线程"""
        # 停止已有的看门狗
        self._stop_watchdog(lock_state)
        
        def watchdog_worker():
            interval = lock_state.timeout / 3
            
            while not lock_state.stop_watchdog.is_set():
                time.sleep(interval)
                
                if lock_state.stop_watchdog.is_set():
                    break
                
                try:
                    result = self._renew_script(
                        keys=[lock_state.key],
                        args=[lock_state.identifier, lock_state.timeout]
                    )
                    
                    if not result:
                        break

                except Exception:
                    break
        
        lock_state.stop_watchdog.clear()
        lock_state.watchdog_thread = threading.Thread(
            target=watchdog_worker,
            daemon=True,
            name=f"Watchdog-{lock_state.key}"
        )
        lock_state.watchdog_thread.start()
    
    def _stop_watchdog(self, lock_state: 'LockState'):
        """停止看门狗线程"""
        lock_state.stop_watchdog.set()
        if lock_state.watchdog_thread and lock_state.watchdog_thread.is_alive():
            lock_state.watchdog_thread.join(timeout=1)
    
    def is_locked(self, key: str) -> bool:
        """
        检查指定键的锁是否被持有
        
        :param key: 锁的键名
        :return: 是否被持有
        """
        lock_key = f"lock:{key}"
        with self._locks_lock:
            if lock_key not in self._locks:
                return False
            return self._locks[lock_key].locked


class LockState:
    """单个锁的状态"""
    
    def __init__(self, key: str, timeout: int):
        self.key = key
        self.timeout = timeout
        self.identifier: Optional[str] = None
        self.locked = False
        
        # 看门狗相关
        self.watchdog_thread: Optional[threading.Thread] = None
        self.stop_watchdog = threading.Event()