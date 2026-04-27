"""
Agent 权限异常
"""

from app.common.exceptions.base_exception import BaseException
from app.common.errors.custom_errors_pkg import AgentPermissionError


class AgentPermissionException(BaseException):
    """
    Agent 权限异常

    用于表示用户没有权限执行某个 Agent
    """

    def __init__(self, agent_id: str = None, user_id: str = None):
        """
        初始化 Agent 权限异常

        Args:
            agent_id: Agent ID
            user_id: 用户 ID
        """
        # AgentPermissionError() 会自动从 Config 获取 debug 模式
        super().__init__(
            error=AgentPermissionError(),
            error_details=f"user_id: {user_id} does not have permission to execute agent_id: {agent_id}",
            error_link="",
        )
