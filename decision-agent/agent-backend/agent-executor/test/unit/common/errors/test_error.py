"""
测试错误基类 Error
"""

from app.common.errors.api_error_class import APIError


class TestError:
    """测试 Error 基类"""

    def test_create_error_basic(self):
        """测试创建基本错误"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        assert error.error_code == "Test.Error.Code"
        assert error.description == "Test description"
        assert error.solution == "Test solution"
        assert error.trace is None

    def test_create_error_with_trace(self):
        """测试创建包含追踪信息的错误"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=True,
        )

        assert error.trace is not None
        assert len(error.trace) > 0
        assert "test_error.py" in error.trace  # 应该包含当前文件名

    def test_to_dict_basic(self):
        """测试转换为字典"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        error_dict = error.to_dict()

        assert error_dict["ErrorCode"] == "Test.Error.Code"
        assert error_dict["Description"] == "Test description"
        assert error_dict["Solution"] == "Test solution"
        assert "Trace" not in error_dict

    def test_to_dict_with_trace(self):
        """测试包含追踪信息的字典转换"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=True,
        )

        error_dict = error.to_dict()

        assert "Trace" in error_dict
        assert len(error_dict["Trace"]) > 0

    def test_from_dict(self):
        """测试从字典创建错误"""
        error_dict = {
            "ErrorCode": "Test.Error.Code",
            "Description": "Test description",
            "Solution": "Test solution",
        }

        error = APIError.from_dict(error_dict)

        assert error.error_code == "Test.Error.Code"
        assert error.description == "Test description"
        assert error.solution == "Test solution"

    def test_from_dict_with_trace(self):
        """测试从包含追踪信息的字典创建错误"""
        error_dict = {
            "ErrorCode": "Test.Error.Code",
            "Description": "Test description",
            "Solution": "Test solution",
            "Trace": "test trace info",
        }

        error = APIError.from_dict(error_dict, include_trace=True)

        assert error.trace == "test trace info"

    def test_error_repr(self):
        """测试错误的字符串表示"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
        )

        repr_str = repr(error)
        assert "Test.Error.Code" in repr_str

    def test_error_str(self):
        """测试错误的字符串转换"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
        )

        str_val = str(error)
        assert str_val == "Test.Error.Code"

    def test_capture_trace_with_exception(self):
        """测试在异常上下文中捕获追踪信息"""
        try:
            raise ValueError("Test exception")
        except ValueError:
            error = APIError(
                error_code="Test.Error.Code",
                description="Test description",
                solution="Test solution",
                include_trace=True,
            )

            assert error.trace is not None
            assert "ValueError: Test exception" in error.trace
