# -*-coding:utf-8-*-

from src.infrastructure.db.db_pool import PymysqlPool
import rdsdriver
from src.utils.logger import logger


def connect_execute_commit_close_db(func):
    def wrapper(*args, **kwargs):
        retry_count = 3
        for attempt in range(retry_count):
            try:
                pymysql_pool = PymysqlPool.get_pool()
                connection = pymysql_pool.connection()
                cursor = connection.cursor()
                kwargs["connection"] = connection
                kwargs["cursor"] = cursor
                try:
                    ret = func(*args, **kwargs)
                    connection.commit()
                    return ret
                except Exception as e:
                    connection.rollback()
                    raise e
                finally:
                    cursor.close()
                    connection.close()
            except (ConnectionResetError, rdsdriver.OperationalError) as e:
                if attempt < retry_count - 1:
                    logger.warningf(
                        f"ConnectionResetError, retrying... Attempt: {attempt + 1}"
                    )
                raise e
        return None

    return wrapper


def connect_execute_close_db(func):
    def wrapper(*args, **kwargs):
        retry = 3
        for attempt in range(retry):
            try:
                pymysql_pool = PymysqlPool.get_pool()
                connection = pymysql_pool.connection()
                kwargs["connection"] = connection
                cursor = connection.cursor()
                kwargs["cursor"] = cursor
                try:
                    ret = func(*args, **kwargs)
                    return ret
                except Exception as e:
                    raise e
                finally:
                    cursor.close()
                    connection.close()
            except (ConnectionResetError, rdsdriver.OperationalError) as e:
                if attempt < retry - 1:
                    logger.warningf(
                        f"ConnectionResetError, retrying... Attempt: {attempt + 1}"
                    )
                raise e
        return None

    return wrapper
