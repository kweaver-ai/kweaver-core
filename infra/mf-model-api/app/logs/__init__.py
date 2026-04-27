import logging
from logging import handlers

sys_log = logging.getLogger('user_manage')


def log_init():
    sys_log.setLevel(level=logging.DEBUG)
    formatter = logging.Formatter(
        '进程ID:%(process)d - '
        '线程ID:%(thread)d- '
        '日志时间:%(asctime)s - '
        '日志等级:%(levelname)s - '
        '日志信息:%(message)s'
    )
    sys_log.handlers.clear()
    file_handler = handlers.TimedRotatingFileHandler('user_app_logs.log', encoding='utf-8', when='W6')
    file_handler.setLevel(level=logging.INFO)
    file_handler.setFormatter(formatter)
    sys_log.addHandler(file_handler)
