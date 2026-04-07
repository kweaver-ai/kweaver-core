"""单元测试 - common/errors/file_errors 模块"""


class TestAgentExecutorFileParseError:
    """测试 AgentExecutor_File_ParseError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.file_errors import AgentExecutor_File_ParseError
        from app.common.errors.api_error_class import APIError

        result = AgentExecutor_File_ParseError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.file_errors import AgentExecutor_File_ParseError

        result = AgentExecutor_File_ParseError()

        assert result.error_code == "AgentExecutor.InternalError.ParseFileError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.file_errors import AgentExecutor_File_ParseError

        result = AgentExecutor_File_ParseError()

        assert (
            "parse" in result.description.lower()
            or "file" in result.description.lower()
        )

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.file_errors import AgentExecutor_File_ParseError

        result = AgentExecutor_File_ParseError()

        assert result.solution is not None
        assert len(result.solution) > 0
