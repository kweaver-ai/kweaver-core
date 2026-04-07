"""单元测试 - common/errors/custom_errors_pkg/conversation_running_error 模块"""


class TestConversationRunningError:
    """测试 ConversationRunningError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg.conversation_running_error import (
            ConversationRunningError,
        )
        from app.common.errors.api_error_class import APIError

        result = ConversationRunningError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.custom_errors_pkg.conversation_running_error import (
            ConversationRunningError,
        )

        result = ConversationRunningError()

        assert result.error_code == "AgentExecutor.ConflictError.ConversationRunning"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.custom_errors_pkg.conversation_running_error import (
            ConversationRunningError,
        )

        result = ConversationRunningError()

        assert "running" in result.description.lower()

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.custom_errors_pkg.conversation_running_error import (
            ConversationRunningError,
        )

        result = ConversationRunningError()

        assert result.solution is not None
        assert len(result.solution) > 0
