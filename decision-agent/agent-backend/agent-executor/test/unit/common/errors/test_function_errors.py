"""单元测试 - common/errors/function_errors 模块"""


class TestAgentExecutorFunctionCodeError:
    """测试 AgentExecutor_Function_CodeError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.function_errors import AgentExecutor_Function_CodeError
        from app.common.errors.api_error_class import APIError

        result = AgentExecutor_Function_CodeError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.function_errors import AgentExecutor_Function_CodeError

        result = AgentExecutor_Function_CodeError()

        assert result.error_code == "AgentExecutor.InternalError.ParseCodeError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.function_errors import AgentExecutor_Function_CodeError

        result = AgentExecutor_Function_CodeError()

        assert "code" in result.description.lower()

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.function_errors import AgentExecutor_Function_CodeError

        result = AgentExecutor_Function_CodeError()

        assert result.solution is not None
        assert len(result.solution) > 0


class TestAgentExecutorFunctionInputError:
    """测试 AgentExecutor_Function_InputError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.function_errors import AgentExecutor_Function_InputError
        from app.common.errors.api_error_class import APIError

        result = AgentExecutor_Function_InputError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.function_errors import AgentExecutor_Function_InputError

        result = AgentExecutor_Function_InputError()

        assert result.error_code == "AgentExecutor.InternalError.RunCodeError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.function_errors import AgentExecutor_Function_InputError

        result = AgentExecutor_Function_InputError()

        assert "input" in result.description.lower()

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.function_errors import AgentExecutor_Function_InputError

        result = AgentExecutor_Function_InputError()

        assert result.solution is not None
        assert len(result.solution) > 0


class TestAgentExecutorFunctionRunError:
    """测试 AgentExecutor_Function_RunError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.function_errors import AgentExecutor_Function_RunError
        from app.common.errors.api_error_class import APIError

        result = AgentExecutor_Function_RunError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.function_errors import AgentExecutor_Function_RunError

        result = AgentExecutor_Function_RunError()

        assert result.error_code == "AgentExecutor.InternalError.RunCodeError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.function_errors import AgentExecutor_Function_RunError

        result = AgentExecutor_Function_RunError()

        assert (
            "execution" in result.description.lower()
            or "failure" in result.description.lower()
        )

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.function_errors import AgentExecutor_Function_RunError

        result = AgentExecutor_Function_RunError()

        assert result.solution is not None
        assert len(result.solution) > 0


class TestAgentExecutorFunctionOutputError:
    """测试 AgentExecutor_Function_OutputError 函数"""

    def test_returns_api_error(self):
        """测试返回APIError实例"""
        from app.common.errors.function_errors import AgentExecutor_Function_OutputError
        from app.common.errors.api_error_class import APIError

        result = AgentExecutor_Function_OutputError()

        assert isinstance(result, APIError)

    def test_error_code(self):
        """测试错误代码"""
        from app.common.errors.function_errors import AgentExecutor_Function_OutputError

        result = AgentExecutor_Function_OutputError()

        assert result.error_code == "AgentExecutor.InternalError.RunCodeError"

    def test_description(self):
        """测试错误描述"""
        from app.common.errors.function_errors import AgentExecutor_Function_OutputError

        result = AgentExecutor_Function_OutputError()

        assert (
            "json" in result.description.lower()
            or "output" in result.description.lower()
        )

    def test_solution(self):
        """测试解决方案"""
        from app.common.errors.function_errors import AgentExecutor_Function_OutputError

        result = AgentExecutor_Function_OutputError()

        assert result.solution is not None
        assert len(result.solution) > 0
