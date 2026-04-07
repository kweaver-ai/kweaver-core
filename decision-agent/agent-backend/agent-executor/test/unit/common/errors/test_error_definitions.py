"""
测试错误定义函数
"""

import pytest

from app.common.errors.custom_errors_pkg import (
    ParamError,
    AgentPermissionError,
    DolphinSDKModelError,
    DolphinSDKSkillError,
    DolphinSDKBaseError,
    ConversationRunningError,
)
from app.common.errors.external_errors import ExternalServiceError
from app.common.errors.file_errors import AgentExecutor_File_ParseError
from app.common.errors.function_errors import (
    AgentExecutor_Function_CodeError,
    AgentExecutor_Function_InputError,
    AgentExecutor_Function_RunError,
    AgentExecutor_Function_OutputError,
)
from app.common.errors.api_error_class import APIError


class TestCommonErrors:
    """测试通用错误定义"""

    def test_param_error(self):
        """测试参数错误"""
        error = ParamError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.BadRequest.ParamError"
        assert error.description is not None
        assert error.solution is not None

    def test_agent_permission_error(self):
        """测试 Agent 权限错误"""
        error = AgentPermissionError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.Forbidden.PermissionError"
        assert error.description is not None
        assert error.solution is not None

    def test_dolphin_sdk_model_error(self):
        """测试 Dolphin SDK 模型错误"""
        error = DolphinSDKModelError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.DolphinSDKException.ModelExecption"
        assert error.description is not None
        assert error.solution is not None

    def test_dolphin_sdk_skill_error(self):
        """测试 Dolphin SDK 技能错误"""
        error = DolphinSDKSkillError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.DolphinSDKException.SkillExecption"
        assert error.description is not None
        assert error.solution is not None

    def test_dolphin_sdk_base_error(self):
        """测试 Dolphin SDK 基础错误"""
        error = DolphinSDKBaseError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.DolphinSDKException.BaseExecption"
        assert error.description is not None
        assert error.solution is not None

    def test_conversation_running_error(self):
        """测试会话运行错误"""
        error = ConversationRunningError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.ConflictError.ConversationRunning"
        assert error.description is not None
        assert error.solution is not None


class TestExternalErrors:
    """测试外部服务错误定义"""

    def test_external_service_error(self):
        """测试外部服务错误"""
        error = ExternalServiceError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.ExternalServiceError"
        assert error.description is not None
        assert error.solution is not None


class TestFileErrors:
    """测试文件错误定义"""

    def test_file_parse_error(self):
        """测试文件解析错误"""
        error = AgentExecutor_File_ParseError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.ParseFileError"
        assert error.description is not None
        assert error.solution is not None


class TestFunctionErrors:
    """测试函数相关错误定义"""

    def test_function_code_error(self):
        """测试函数代码错误"""
        error = AgentExecutor_Function_CodeError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.ParseCodeError"
        assert error.description is not None
        assert error.solution is not None

    def test_function_input_error(self):
        """测试函数输入参数错误"""
        error = AgentExecutor_Function_InputError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.RunCodeError"
        assert error.description is not None
        assert error.solution is not None

    def test_function_run_error(self):
        """测试函数运行错误"""
        error = AgentExecutor_Function_RunError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.RunCodeError"
        assert error.description is not None
        assert error.solution is not None

    def test_function_output_error(self):
        """测试函数输出错误"""
        error = AgentExecutor_Function_OutputError()

        assert isinstance(error, APIError)
        assert error.error_code == "AgentExecutor.InternalError.RunCodeError"
        assert error.description is not None
        assert error.solution is not None


if __name__ == "__main__":
    pytest.main([__file__, "-v"])
