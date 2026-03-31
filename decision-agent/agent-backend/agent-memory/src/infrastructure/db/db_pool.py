# -*-coding:utf-8-*-
import threading
import rdsdriver
from dbutilsx.pooled_db import PooledDB, PooledDBInfo
from src.utils.logger import logger


# 单例
class PymysqlPool(object):
    yamlConfig = None
    _pool = None
    _instance_lock = threading.Lock()

    def __new__(cls, *args, **kwargs):
        if not hasattr(cls, "_instance"):
            with PymysqlPool._instance_lock:
                if not hasattr(cls, "_instance"):
                    PymysqlPool._instance = super().__new__(cls)
            return PymysqlPool._instance

    @classmethod
    def get_pool(cls):
        if cls._pool is not None:
            return cls._pool
        with cls._instance_lock:
            if cls._pool is not None:
                return cls._pool

        DB_MINCACHED = 2
        DB_MAXCACHED = 5
        DB_MAXSHARED = 5
        DB_MAXCONNECTIONS = 20
        DB_BLOCKING = True

        from src.config import db_config

        DB_HOST = db_config.get("host")
        DB_PORT = db_config.get("port")
        DB_USER_NAME = db_config.get("user")
        DB_PASSWORD = db_config.get("password")
        DB_SCHEMA = db_config.get("database")
        CHARSET = "utf8"

        try:
            w = PooledDBInfo(
                creator=rdsdriver,
                mincached=DB_MINCACHED,
                maxcached=DB_MAXCACHED,
                maxshared=DB_MAXSHARED,
                maxconnections=DB_MAXCONNECTIONS,
                blocking=DB_BLOCKING,
                host=DB_HOST,
                port=DB_PORT,
                user=DB_USER_NAME,
                password=DB_PASSWORD,
                database=DB_SCHEMA,
                charset=CHARSET,
                cursorclass=rdsdriver.DictCursor,
            )
            cls._pool = PooledDB(master=w, backup=w)
            logger.info("Connect to database successfully")
            return cls._pool
        except Exception as e:
            logger.error("Unexpected error while connecting to database", error=str(e))
            raise
