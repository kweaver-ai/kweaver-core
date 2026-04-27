"""单元测试 - common/errors/custom_errors_pkg/agent_permission_error 模块"""


class TestAgentPermissionError:
    """测试 AgentPermissionError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.custom_errors_pkg.agent_permission_error import (
            AgentPermissionError,
        )
        from app.common.errors.api_error_class import APIError

        result = AgentPermissionError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.custom_errors_pkg.agent_permission_error import (
            AgentPermissionError,
        )

        result = AgentPermissionError()

        assert result.error_code == "AgentExecutor.Forbidden.PermissionError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.custom_errors_pkg.agent_permission_error import (
            AgentPermissionError,
        )

        result = AgentPermissionError()

        assert "permission" in result.description.lower()

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.custom_errors_pkg.agent_permission_error import (
            AgentPermissionError,
        )

        result = AgentPermissionError()

        assert result.solution is not None
        assert len(result.solution) > 0
