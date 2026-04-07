#!/usr/bin/python3
from utils.redis import init_redis
from driveradapters import start_server
from models.db import init_db
from models.models import Base
from common.configs import db_configs
from drivenadapters.mq import MQ
from environs import Env

def main():
    # 加载.env文件
    env = Env()
    env.read_env()
    init_db(Base, **db_configs)
    MQ.init_connector_from_file("/sysvol/conf/mq_config.yaml")
    init_redis()
    start_server()


if __name__ == "__main__":
    main()