"""单元测试 - common/errors/api_error_class 模块"""


class TestAPIErrorInit:
    """测试APIError初始化"""

    def test_init_with_basic_params(self):
        """测试基本参数初始化"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
        )

        assert error.error_code == "Test.Error"
        assert error.description == "Test description"
        assert error.solution == "Test solution"

    def test_init_with_include_trace_true(self):
        """测试包含追踪信息"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
            include_trace=True,
        )

        assert error.trace is not None

    def test_init_with_include_trace_false(self):
        """测试不包含追踪信息"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        assert error.trace is None

    def test_init_with_include_trace_none(self):
        """测试include_trace为None"""
        from app.common.errors.api_error_class import APIError

        # When include_trace is None, it tries to use Config
        # If Config is not available, it should still work
        error = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
            include_trace=None,
        )

        # Should have created error object
        assert error.error_code == "Test.Error"

    def test_trace_attribute_exists(self):
        """测试trace属性存在"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test.Error", description="Test", solution="Test")

        assert hasattr(error, "trace")

    def test_error_code_stored(self):
        """测试错误码存储"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="ErrorCode.Test", description="Test", solution="Test"
        )

        assert error.error_code == "ErrorCode.Test"

    def test_description_stored(self):
        """测试描述存储"""
        from app.common.errors.api_error_class import APIError

        desc = "This is a test error description"
        error = APIError(error_code="Test", description=desc, solution="Test")

        assert error.description == desc

    def test_solution_stored(self):
        """测试解决方案存储"""
        from app.common.errors.api_error_class import APIError

        sol = "This is the solution"
        error = APIError(error_code="Test", description="Test", solution=sol)

        assert error.solution == sol


class TestAPICaptureTrace:
    """测试_capture_trace方法"""

    def test_capture_trace_without_exception(self):
        """测试无异常时捕获追踪"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test",
            solution="Test",
            include_trace=True,
        )

        # Should have captured stack trace
        assert error.trace is not None
        assert isinstance(error.trace, str)

    def test_capture_trace_with_exception(self):
        """测试有异常时捕获追踪"""
        from app.common.errors.api_error_class import APIError

        try:
            raise ValueError("Test exception")
        except ValueError:
            error = APIError(
                error_code="Test.Error",
                description="Test",
                solution="Test",
                include_trace=True,
            )

            assert error.trace is not None
            assert "ValueError" in error.trace or "Test exception" in error.trace

    def test_capture_trace_content_format(self):
        """测试追踪内容格式"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test",
            solution="Test",
            include_trace=True,
        )

        # Trace should contain file information
        assert "test_api_error_class" in error.trace or ".py" in error.trace


class TestAPIToDict:
    """测试to_dict方法"""

    def test_to_dict_basic_fields(self):
        """测试基本字段转换"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
        )

        result = error.to_dict()

        assert result["ErrorCode"] == "Test.Error"
        assert result["Description"] == "Test description"
        assert result["Solution"] == "Test solution"

    def test_to_dict_with_trace(self):
        """测试带追踪信息转换"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test",
            solution="Test",
            include_trace=True,
        )

        result = error.to_dict()

        assert "Trace" in result
        assert result["Trace"] is not None

    def test_to_dict_without_trace(self):
        """测试不带追踪信息转换"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error",
            description="Test",
            solution="Test",
            include_trace=False,
        )

        result = error.to_dict()

        assert "Trace" not in result

    def test_to_dict_returns_dict(self):
        """测试返回字典类型"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test.Error", description="Test", solution="Test")

        result = error.to_dict()

        assert isinstance(result, dict)

    def test_to_dict_all_required_keys(self):
        """测试所有必需键存在"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test.Error", description="Test", solution="Test")

        result = error.to_dict()

        assert "ErrorCode" in result
        assert "Description" in result
        assert "Solution" in result


class TestAPIRepr:
    """测试__repr__方法"""

    def test_repr_format(self):
        """测试repr格式"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error.Code", description="Test", solution="Test"
        )

        result = repr(error)

        assert "Error(error_code=" in result
        assert "Test.Error.Code" in result

    def test_repr_returns_string(self):
        """测试返回字符串"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test", description="Test", solution="Test")

        result = repr(error)

        assert isinstance(result, str)


class TestAPIStr:
    """测试__str__方法"""

    def test_str_returns_error_code(self):
        """测试返回错误码"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error.Code", description="Test", solution="Test"
        )

        result = str(error)

        assert result == "Test.Error.Code"

    def test_str_returns_string(self):
        """测试返回字符串"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test", description="Test", solution="Test")

        result = str(error)

        assert isinstance(result, str)


class TestAPIFromDict:
    """测试from_dict静态方法"""

    def test_from_dict_basic(self):
        """测试基本转换"""
        from app.common.errors.api_error_class import APIError

        error_dict = {
            "ErrorCode": "Test.Error",
            "Description": "Test description",
            "Solution": "Test solution",
        }

        error = APIError.from_dict(error_dict)

        assert error.error_code == "Test.Error"
        assert error.description == "Test description"
        assert error.solution == "Test solution"

    def test_from_dict_with_defaults(self):
        """测试使用默认值"""
        from app.common.errors.api_error_class import APIError

        error_dict = {}

        error = APIError.from_dict(error_dict)

        assert error.error_code == "AgentExecutor.InternalServerError.UnknownError"
        assert error.description == "Unknown error"
        assert error.solution == "Please check the service."

    def test_from_dict_with_trace(self):
        """测试带追踪信息"""
        from app.common.errors.api_error_class import APIError

        error_dict = {
            "ErrorCode": "Test.Error",
            "Description": "Test",
            "Solution": "Test",
            "Trace": "Stack trace here",
        }

        error = APIError.from_dict(error_dict)

        assert error.trace == "Stack trace here"

    def test_from_dict_partial_fields(self):
        """测试部分字段"""
        from app.common.errors.api_error_class import APIError

        error_dict = {"ErrorCode": "Test.Error"}

        error = APIError.from_dict(error_dict)

        assert error.error_code == "Test.Error"
        assert error.description == "Unknown error"
        assert error.solution == "Please check the service."

    def test_from_dict_include_trace_false(self):
        """测试include_trace为False"""
        from app.common.errors.api_error_class import APIError

        error_dict = {
            "ErrorCode": "Test.Error",
            "Description": "Test",
            "Solution": "Test",
        }

        error = APIError.from_dict(error_dict, include_trace=False)

        assert error.trace is None

    def test_from_dict_include_trace_true(self):
        """测试include_trace为True"""
        from app.common.errors.api_error_class import APIError

        error_dict = {
            "ErrorCode": "Test.Error",
            "Description": "Test",
            "Solution": "Test",
        }

        error = APIError.from_dict(error_dict, include_trace=True)

        assert error.trace is not None


class TestAPIErrorRoundTrip:
    """测试往返转换"""

    def test_to_dict_from_dict_roundtrip(self):
        """测试to_dict和from_dict往返"""
        from app.common.errors.api_error_class import APIError

        original = APIError(
            error_code="Test.Error",
            description="Test description",
            solution="Test solution",
        )

        error_dict = original.to_dict()
        restored = APIError.from_dict(error_dict)

        assert restored.error_code == original.error_code
        assert restored.description == original.description
        assert restored.solution == original.solution

    def test_roundtrip_with_trace(self):
        """测试带追踪信息的往返"""
        from app.common.errors.api_error_class import APIError

        original = APIError(
            error_code="Test.Error",
            description="Test",
            solution="Test",
            include_trace=True,
        )

        error_dict = original.to_dict()
        restored = APIError.from_dict(error_dict)

        assert restored.trace is not None


class TestAPIErrorEdgeCases:
    """测试边界情况"""

    def test_empty_error_code(self):
        """测试空错误码"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="", description="Test", solution="Test")

        assert error.error_code == ""

    def test_long_error_code(self):
        """测试长错误码"""
        from app.common.errors.api_error_class import APIError

        long_code = "A" * 1000
        error = APIError(error_code=long_code, description="Test", solution="Test")

        assert error.error_code == long_code

    def test_unicode_description(self):
        """测试Unicode描述"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error", description="错误描述", solution="解决方案"
        )

        assert error.description == "错误描述"
        assert error.solution == "解决方案"

    def test_special_chars_in_fields(self):
        """测试字段中的特殊字符"""
        from app.common.errors.api_error_class import APIError

        error = APIError(
            error_code="Test.Error\n\t",
            description="Test\nDescription",
            solution="Test\tSolution",
        )

        assert "\n" in error.error_code
        assert "\n" in error.description

    def test_multiple_calls_to_to_dict(self):
        """测试多次调用to_dict"""
        from app.common.errors.api_error_class import APIError

        error = APIError(error_code="Test.Error", description="Test", solution="Test")

        dict1 = error.to_dict()
        dict2 = error.to_dict()

        assert dict1 == dict2

    def test_multiple_from_dict_same_dict(self):
        """测试同一字典多次from_dict"""
        from app.common.errors.api_error_class import APIError

        error_dict = {
            "ErrorCode": "Test.Error",
            "Description": "Test",
            "Solution": "Test",
        }

        error1 = APIError.from_dict(error_dict)
        error2 = APIError.from_dict(error_dict)

        assert error1.error_code == error2.error_code
        assert error1.description == error2.description
