"""
测试异常类
"""

import json
import pytest
from unittest.mock import patch

from app.common.exceptions import (
    BaseException,
    CodeException,
    ParamException,
    AgentPermissionException,
    DolphinSDKException,
    ConversationRunningException,
)
from app.common.errors.api_error_class import APIError
from app.common.dependencies import create_dolphin_exception


class TestBaseException:
    """测试基础异常类"""

    def test_base_exception_creation(self):
        """测试创建基础异常"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        exc = BaseException(
            error=error,
            error_details="Detailed error message",
            error_link="https://example.com/error",
        )

        assert isinstance(exc.error, APIError)
        assert exc.error_details == "Detailed error message"
        assert exc.error_link == "https://example.com/error"

    def test_format_http_error(self):
        """测试格式化 HTTP 错误响应"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        exc = BaseException(
            error=error,
            error_details="Detailed error message",
        )

        http_error = exc.FormatHttpError()

        assert http_error["error_code"] == "Test.Error.Code"
        assert http_error["description"] == "Test description"
        assert http_error["error_details"] == "Detailed error message"
        assert http_error["solution"] == "Test solution"

    def test_format_http_error_with_trace(self):
        """测试包含追踪信息的错误格式化"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=True,
        )

        exc = BaseException(error=error)
        http_error = exc.FormatHttpError()

        assert "trace" in http_error
        assert len(http_error["trace"]) > 0

    def test_exception_str_and_repr(self):
        """测试异常的字符串表示"""
        error = APIError(
            error_code="Test.Error.Code",
            description="Test description",
            solution="Test solution",
            include_trace=False,
        )

        exc = BaseException(error=error)

        # str 和 repr 应该返回 JSON 格式的错误信息
        str_val = str(exc)
        repr_val = repr(exc)

        assert "Test.Error.Code" in str_val
        assert "Test.Error.Code" in repr_val

        # 验证是否为有效的 JSON
        json_data = json.loads(str_val)
        assert json_data["error_code"] == "Test.Error.Code"


class TestCodeException:
    """测试代码异常"""

    def test_code_exception_creation(self):
        """测试创建代码异常"""
        error = APIError(
            error_code="AgentExecutor.CodeError",
            description="Code error occurred",
            solution="Check your code",
            include_trace=False,
        )

        exc = CodeException(error=error, error_details="Syntax error in line 10")

        assert isinstance(exc.error, APIError)
        assert exc.error_details == "Syntax error in line 10"


class TestParamException:
    """测试参数异常"""

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_param_exception_creation(self, mock_debug):
        """测试创建参数异常"""
        exc = ParamException(error_details="Invalid parameter 'user_id'")

        assert exc.error_details == "Invalid parameter 'user_id'"
        assert "ParamError" in str(exc.error.error_code)

    @patch("app.common.config.Config.is_debug_mode", return_value=True)
    def test_param_exception_with_debug(self, mock_debug):
        """测试调试模式下的参数异常"""
        exc = ParamException(error_details="Invalid parameter 'user_id'")

        # 在调试模式下应该包含追踪信息
        http_error = exc.FormatHttpError()
        # ParamError 函数返回 Error 对象
        assert isinstance(exc.error, APIError)


class TestAgentPermissionException:
    """测试 Agent 权限异常"""

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_agent_permission_exception(self, mock_debug):
        """测试创建 Agent 权限异常"""
        exc = AgentPermissionException(agent_id="agent123", user_id="user456")

        assert "agent123" in exc.error_details
        assert "user456" in exc.error_details
        assert "PermissionError" in str(exc.error.error_code)


class TestDolphinSDKException:
    """测试 Dolphin SDK 异常"""

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_dolphin_sdk_model_exception(self, mock_debug):
        """测试 Dolphin SDK 模型异常"""
        raw_exc = create_dolphin_exception("ModelException", "Model error")
        exc = DolphinSDKException(
            raw_exception=raw_exc,
            agent_id="agent123",
            session_id="session456",
            user_id="user789",
        )

        assert "agent123" in exc.error_details
        assert "session456" in exc.error_details
        assert "user789" in exc.error_details
        # 检查 error_code 中包含 ModelException 相关部分
        assert "Model" in str(exc.error.error_code) and "Execption" in str(
            exc.error.error_code
        )

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_dolphin_sdk_skill_exception(self, mock_debug):
        """测试 Dolphin SDK 技能异常"""
        raw_exc = create_dolphin_exception("SkillException", "Skill error")
        exc = DolphinSDKException(
            raw_exception=raw_exc,
            agent_id="agent123",
            session_id="session456",
        )

        # 检查 error_code 中包含 SkillException 相关部分
        assert "Skill" in str(exc.error.error_code) and "Execption" in str(
            exc.error.error_code
        )

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_dolphin_sdk_base_exception(self, mock_debug):
        """测试 Dolphin SDK 基础异常"""
        raw_exc = create_dolphin_exception("DolphinException", "Generic error")
        exc = DolphinSDKException(
            raw_exception=raw_exc,
            agent_id="agent123",
            session_id="session456",
        )

        # 对于未知异常类型，应该使用 BaseExecption
        assert "Base" in str(exc.error.error_code) and "Execption" in str(
            exc.error.error_code
        )


class TestConversationRunningException:
    """测试会话运行异常"""

    @patch("app.common.config.Config.is_debug_mode", return_value=False)
    def test_conversation_running_exception(self, mock_debug):
        """测试创建会话运行异常"""
        exc = ConversationRunningException(
            error_details="Conversation is already running"
        )

        assert exc.error_details == "Conversation is already running"
        assert "ConversationRunning" in str(exc.error.error_code)


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
