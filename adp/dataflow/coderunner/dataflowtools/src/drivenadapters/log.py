# drivenadapters/log.py

import json
import os
import time
from typing import List
from drivenadapters.user_management import UserManagement
from models.driven_models import UserInfo
from drivenadapters.mq import MQ
from common.topics import *

class BaseLog:

    @staticmethod
    def parse_depts(dept_paths: List) -> str:
        dept_names = []
        for dept in dept_paths:
            if len(dept) <= 0:
                continue
            dept_name = dept[-1].get("name")
            if dept_name:
                dept_names.append(dept_name)
        return ','.join(dept_names)

class AuditLog:
    def __init__(self, user_id, user_name, user_type, level, op_type, date, ip, msg, ex_msg, user_agent, out_biz_id, dept_paths):
        self.user_id = user_id
        self.user_name = user_name
        self.user_type = user_type
        self.level = level
        self.op_type = op_type
        self.date = date
        self.ip = ip
        self.msg = msg
        self.ex_msg = ex_msg
        self.user_agent = user_agent
        self.out_biz_id = out_biz_id
        self.dept_paths = dept_paths

class Operator:
    def __init__(self, type, id, name, agent, department_path):
        self.type = type
        self.id = id
        self.name = name
        self.agent = agent
        self.department_path = department_path

class Agent:
    def __init__(self, udid, ip, type):
        self.udid = udid
        self.ip = ip
        self.type = type

class BuildAuditLogParams:
    def __init__(self, user_info: UserInfo, msg: str, ext_msg: str, out_biz_id: str, log_level: int):
        self.user_info = user_info
        self.msg = msg
        self.ext_msg = ext_msg
        self.out_biz_id = out_biz_id
        self.log_level = log_level

class Log(BaseLog):

    NcTDocOperType_NCT_DOT_AUTOMATION = 28
    NcTLogLevel_NCT_LL_INFO = 1
    NcTLogLevel_NCT_LL_WARN = 2
    ClientTypeUnknown = "unknown"

    @classmethod
    async def build_audit_log(cls, params: BuildAuditLogParams):
        audit_log = AuditLog(
            user_id=params.user_info.user_id,
            user_name=params.user_info.user_name,
            user_type=params.user_info.visitor_type or "authenticated_user",
            level=params.log_level,
            op_type=cls.NcTDocOperType_NCT_DOT_AUTOMATION,  # NcTDocOperType_NCT_DOT_AUTOMATION
            date=int(time.time() * 1e6),
            ip=params.user_info.login_ip or os.getenv("POD_IP"),
            msg=params.msg,
            ex_msg=params.ext_msg,
            user_agent=params.user_info.user_agent,
            out_biz_id=params.out_biz_id,
            dept_paths=""
        )

        if params.user_info.parent_deps:
            audit_log.dept_paths = cls.parse_depts(params.user_info.parent_deps)
        else:
            user_info = await UserManagement().get_user_info(params.user_info.user_id)
            if not user_info:
                audit_log.dept_paths = "未分配组"
                return audit_log
            audit_log.dept_paths = cls.parse_depts(user_info[0].parent_deps)

        return audit_log

    @classmethod
    async def log(cls, params):
        log_data = None
        topic = None

        if isinstance(params, BuildAuditLogParams):
            log_data = await cls.build_audit_log(params)
            topic = TopicAuditLog

        if log_data:
            json_string = json.dumps(log_data.__dict__, ensure_ascii=False)
            await MQ.publish(topic=topic, message=json_string)
