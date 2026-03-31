"""
异常模块

提供统一的异常类体系，所有自定义异常都继承自 BaseException
"""

from app.common.exceptions.base_exception import BaseException
from app.common.exceptions.code_exception import CodeException
from app.common.exceptions.param_exception import ParamException
from app.common.exceptions.agent_permission_exception import AgentPermissionException
from app.common.exceptions.dolphin_sdk_exception import DolphinSDKException
from app.common.exceptions.conversation_running_exception import (
    ConversationRunningException,
)


__all__ = [
    "BaseException",
    "CodeException",
    "ParamException",
    "AgentPermissionException",
    "DolphinSDKException",
    "ConversationRunningException",
]
