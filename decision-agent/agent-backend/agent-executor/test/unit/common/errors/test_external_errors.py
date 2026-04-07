"""单元测试 - common/errors/external_errors 模块"""


class TestExternalServiceError:
    """测试 ExternalServiceError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.external_errors import ExternalServiceError
        from app.common.errors.api_error_class import APIError

        result = ExternalServiceError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.external_errors import ExternalServiceError

        result = ExternalServiceError()

        assert result.error_code == "AgentExecutor.InternalError.ExternalServiceError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.external_errors import ExternalServiceError

        result = ExternalServiceError()

        assert "external service" in result.description.lower()

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.external_errors import ExternalServiceError

        result = ExternalServiceError()

        assert result.solution is not None
        assert len(result.solution) > 0
