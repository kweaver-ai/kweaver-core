import json
import os

from mq_sdk.proton_mq import Connector

from common.logger import logger
from common.configs import mq_configs

class MQ:
    con = ""
    phost = ""
    pport = ""

    @classmethod
    def initconnect(cls):
        cls.phost = mq_configs.get("host")
        cls.pport = int(mq_configs.get("port"))
        cls.chost = mq_configs.get("lookupd_host")
        cls.cport = int(mq_configs.get("lookupd_port"))
        cls.connector_type = mq_configs.get("connector_type")
        cls.con = Connector.get_connector(cls.phost, cls.pport, cls.chost, cls.cport, cls.connector_type)

    @classmethod
    def init_connector_from_file(cls, config_file_path):
        cls.con = Connector.get_connector_from_file(config_file_path)

    @classmethod
    async def create_consumer(cls, topic, channel, handler):
        # 60代表nsq两次查询时间间隔为60s， 16代表一个consumer一次最多处理消息的数量
        return await cls.con.sub(topic, channel, handler, 60, 16)

    @classmethod
    async def create_producer(cls, topic, message):
        if isinstance(message, (dict, list)):
            message = json.dumps(message)
        try:
            await cls.con.pub(topic, message)
            logger.info(f"Send success, topic:{topic}, message:{message}")
        except Exception:
            logger.exception(f"Send failed, topic:{topic}, message:{message}.")

    @classmethod
    async def publish(cls, topic, message):
        """
        更新函数名，后续若有新的发布消息逻辑，请调用此方法
        """
        await cls.create_producer(topic, message)
