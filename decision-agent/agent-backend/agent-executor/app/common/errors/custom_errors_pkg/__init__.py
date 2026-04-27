"""
自定义错误定义包

包含所有通用的错误定义，每个错误定义在单独的文件中
"""

from app.common.errors.custom_errors_pkg.param_error import ParamError
from app.common.errors.custom_errors_pkg.agent_permission_error import (
    AgentPermissionError,
)
from app.common.errors.custom_errors_pkg.dolphin_sdk_model_error import (
    DolphinSDKModelError,
)
from app.common.errors.custom_errors_pkg.dolphin_sdk_skill_error import (
    DolphinSDKSkillError,
)
from app.common.errors.custom_errors_pkg.dolphin_sdk_base_error import (
    DolphinSDKBaseError,
)
from app.common.errors.custom_errors_pkg.conversation_running_error import (
    ConversationRunningError,
)

__all__ = [
    "ParamError",
    "AgentPermissionError",
    "DolphinSDKModelError",
    "DolphinSDKSkillError",
    "DolphinSDKBaseError",
    "ConversationRunningError",
]
