import pymysql
from pymysql.cursors import Cursor, DictCursor
from src.common.logger import logger
from typing import List, Dict, Any, Optional, Union
from src.config.setting import config


class MySQLConnector:
    """
    PyMySQL数据库操作封装类，提供基础的数据库连接和CRUD操作
    """

    def __init__(
        self,
        host: str = None,
        user: str = None,
        password: str = None,
        database: str = None,
        port: int = None,
        charset: str = "utf8mb4",
        env_config: dict = None,
    ):
        """
        初始化数据库连接参数

        支持三种配置方式：
        1. 不传参数：使用配置文件默认值
        2. 传入env_config：使用指定环境配置
        3. 传入具体参数：使用指定参数（可部分传入，其余使用默认值）

        :param host: 数据库主机地址
        :param user: 数据库用户名
        :param password: 数据库密码
        :param database: 数据库名称
        :param port: 数据库端口
        :param charset: 字符集，默认utf8mb4
        :param env_config: 环境配置字典，包含数据库连接信息
        """
        if env_config:
            # 使用环境配置
            self.host = env_config.get("host") or config.get("database", "host")
            self.user = env_config.get("user") or config.get("database", "user")
            self.password = env_config.get("password") or config.get(
                "database", "password"
            )
            self.database = env_config.get("database") or config.get(
                "database", "database"
            )
            self.port = env_config.get("port") or config.getint("database", "port")
        else:
            # 使用传入参数或配置文件默认值
            self.host = host or config.get("database", "host")
            self.user = user or config.get("database", "user")
            self.password = password or config.get("database", "password")
            self.database = database or config.get("database", "database")
            self.port = port or config.getint("database", "port")

        self.charset = charset
        self.cursorclass = DictCursor

        # 连接和游标对象
        self._connection: Optional[pymysql.connections.Connection] = None
        self._cursor: Optional[Cursor] = None

    @classmethod
    def from_config(cls, section: str = "database"):
        """
        从配置文件指定section创建连接器实例

        :param section: 配置文件中的section名称，默认为"database"
        :return: MySQLConnector实例
        """
        try:
            return cls(
                host=config.get(section, "host"),
                user=config.get(section, "user"),
                password=config.get(section, "password"),
                database=config.get(section, "database"),
                port=config.getint(section, "port"),
            )
        except Exception as e:
            logger.error(f"从配置section [{section}] 创建数据库连接失败: {e}")
            raise

    @classmethod
    def from_env_config(cls, env_config: dict):
        """
        从环境配置字典创建连接器实例

        :param env_config: 环境配置字典
        :return: MySQLConnector实例
        """
        return cls(env_config=env_config)

    @classmethod
    def for_database(cls, database: str, **kwargs):
        """
        为指定数据库创建连接器实例，其他参数使用默认配置

        :param database: 数据库名称
        :param kwargs: 其他连接参数
        :return: MySQLConnector实例
        """
        return cls(database=database, **kwargs)

    @property
    def connection(self) -> pymysql.connections.Connection:
        """获取数据库连接，若未连接则自动建立连接"""
        if not self._connection or not self._connection.open:
            self.connect()
        return self._connection

    @property
    def cursor(self) -> Cursor:
        """获取游标对象，若不存在则创建"""
        if not self._cursor:
            self._cursor = self.connection.cursor()
        return self._cursor

    def connect(self) -> None:
        """建立数据库连接"""
        try:
            self._connection = pymysql.connect(
                host=self.host,
                user=self.user,
                password=self.password,
                database=self.database,
                port=self.port,
                charset=self.charset,
                cursorclass=self.cursorclass,
            )
            logger.info(f"成功连接到数据库: {self.host}:{self.port}/{self.database}")
        except Exception as e:
            logger.error(f"数据库连接失败: {str(e)}")
            raise

    def close(self) -> None:
        """关闭数据库连接和游标"""
        # 关闭游标
        if self._cursor:
            try:
                self._cursor.close()
                self._cursor = None
            except Exception as e:
                logger.warning(f"关闭游标时发生错误: {str(e)}")

        # 关闭连接
        if self._connection and self._connection.open:
            try:
                self._connection.close()
                logger.debug(
                    f"已关闭数据库连接: {self.host}:{self.port}/{self.database}"
                )
            except Exception as e:
                logger.warning(f"关闭数据库连接时发生错误: {str(e)}")

        self._connection = None

    def execute(self, sql: str, params: Optional[Union[tuple, dict]] = None) -> int:
        """
        执行SQL语句（通用方法）

        :param sql: SQL语句
        :param params: SQL参数，用于参数化查询，防止SQL注入
        :return: 受影响的行数
        """
        try:
            logger.info(f"执行SQL: {sql}，参数: {params}")
            affected_rows = self.cursor.execute(sql, params or ())
            logger.debug(f"SQL执行成功，受影响行数: {affected_rows}")
            return affected_rows
        except Exception as e:
            logger.error(f"执行SQL失败: {str(e)}，SQL: {sql}，参数: {params}")
            raise

    def fetch_one(
        self, sql: str, params: Optional[Union[tuple, dict]] = None
    ) -> Optional[Union[Dict[str, Any], tuple]]:
        """
        执行查询并返回第一条结果
        :param sql: 查询SQL语句
        :param params: SQL参数
        :return: 单条记录，字典或元组格式，取决于游标类型
        """
        self.execute(sql, params)
        result = self.cursor.fetchone()
        #  logger.info(f"查询到单条记录: {result}")
        return result

    def fetch_all(
        self, sql: str, params: Optional[Union[tuple, dict]] = None
    ) -> List[Union[Dict[str, Any], tuple]]:
        """
        执行查询并返回所有结果

        :param sql: 查询SQL语句
        :param params: SQL参数
        :return: 记录列表，每个元素为字典或元组，取决于游标类型
        """
        self.execute(sql, params)
        result = self.cursor.fetchall()
        #    logger.info(f"查询到记录: {result}")
        return result

    def fetch_many(
        self, sql: str, size: int, params: Optional[Union[tuple, dict]] = None
    ) -> List[Union[Dict[str, Any], tuple]]:
        """
        执行查询并返回指定数量的结果

        :param sql: 查询SQL语句
        :param size: 要返回的记录数
        :param params: SQL参数
        :return: 记录列表
        """
        self.execute(sql, params)
        result = self.cursor.fetchmany(size)
        # logger.info(f"查询到记录: {result}")
        return result

    def commit(self) -> None:
        """提交事务"""
        try:
            self.connection.commit()
            logger.debug("事务已提交")
        except Exception as e:
            logger.error(f"提交事务失败: {str(e)}")
            raise

    def rollback(self) -> None:
        """回滚事务"""
        try:
            self.connection.rollback()
            logger.info("事务已回滚")
        except Exception as e:
            logger.error(f"回滚事务失败: {str(e)}")
            raise

    def __enter__(self) -> "MySQLConnector":
        """支持上下文管理器，进入时自动连接"""
        self.connect()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb) -> bool:
        """支持上下文管理器，退出时自动关闭连接"""
        self.close()
        # 若有异常则返回False，让异常继续传播
        return False

    def __del__(self) -> None:
        """对象销毁时确保连接已关闭"""
        self.close()


# 基本使用示例
if __name__ == "__main__":
    # 数据库配置
    db_config = {
        "host": "192.168.112.36",
        "user": "root",
        "password": "eisoo.com123",
        "database": "dip",
        "port": 30006,
    }

    # 使用上下文管理器（推荐）
    with MySQLConnector() as db:
        # 查询示例
        user = db.fetch_all("SELECT * FROM t_knowledge_network")
