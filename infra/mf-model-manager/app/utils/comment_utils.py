from datetime import datetime


async def write_log(api=None, msg=None, user='root'):
    with open("log.log", mode="a", encoding='utf-8') as log:
        now = datetime.now()
        log.write(f"时间：{now}    API调用事件：{api}    用户：{user}    消息：{msg}\n")


async def error_log(api=None, msg=None, user='root'):
    with open("err or.log", mode="a", encoding='utf-8') as log:
        now = datetime.now()
        log.write(f"时间：{now}    API调用事件：{api}    用户：{user}    消息：{msg}\n")
