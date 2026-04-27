# -*- coding: utf-8 -*-
"""
Unit tests for app/common/errors/__init__.py
"""

import pytest

from app.common.errors import (
    APIError,
    ParamError,
    AgentPermissionError,
    DolphinSDKModelError,
    DolphinSDKSkillError,
    DolphinSDKBaseError,
    ConversationRunningError,
    ExternalServiceError,
    AgentExecutor_File_ParseError,
)


class TestModuleImports:
    """测试模块导入"""

    def test_import_api_error(self):
        """测试导入 APIError"""
        assert APIError is not None

    def test_import_param_error(self):
        """测试导入 ParamError"""
        assert ParamError is not None

    def test_import_agent_permission_error(self):
        """测试导入 AgentPermissionError"""
        assert AgentPermissionError is not None

    def test_import_dolphin_sdk_model_error(self):
        """测试导入 DolphinSDKModelError"""
        assert DolphinSDKModelError is not None

    def test_import_dolphin_sdk_skill_error(self):
        """测试导入 DolphinSDKSkillError"""
        assert DolphinSDKSkillError is not None

    def test_import_dolphin_sdk_base_error(self):
        """测试导入 DolphinSDKBaseError"""
        assert DolphinSDKBaseError is not None

    def test_import_conversation_running_error(self):
        """测试导入 ConversationRunningError"""
        assert ConversationRunningError is not None

    def test_import_external_service_error(self):
        """测试导入 ExternalServiceError"""
        assert ExternalServiceError is not None

    def test_import_file_parse_error(self):
        """测试导入 AgentExecutor_File_ParseError"""
        assert AgentExecutor_File_ParseError is not None


class TestLazyImportExceptions:
    """测试异常类的延迟导入"""

    def test_get_base_exception(self):
        """测试获取 BaseException"""
        from app.common.errors import BaseException

        assert BaseException is not None

    def test_get_code_exception(self):
        """测试获取 CodeException"""
        from app.common.errors import CodeException

        assert CodeException is not None

    def test_get_param_exception(self):
        """测试获取 ParamException"""
        from app.common.errors import ParamException

        assert ParamException is not None

    def test_get_agent_permission_exception(self):
        """测试获取 AgentPermissionException"""
        from app.common.errors import AgentPermissionException

        assert AgentPermissionException is not None

    def test_get_dolphin_sdk_exception(self):
        """测试获取 DolphinSDKException"""
        from app.common.errors import DolphinSDKException

        assert DolphinSDKException is not None

    def test_get_conversation_running_exception(self):
        """测试获取 ConversationRunningException"""
        from app.common.errors import ConversationRunningException

        assert ConversationRunningException is not None

    def test_get_nonexistent_exception_raises_attribute_error(self):
        """测试获取不存在的异常抛出 AttributeError"""
        from app.common.errors import __getattr__

        with pytest.raises(AttributeError):
            __getattr__("NonexistentException")


class TestModuleAll:
    """测试 __all__ 导出列表"""

    def test_module_all_contains_api_error(self):
        """测试 __all__ 包含 APIError"""
        from app.common.errors import __all__

        assert "APIError" in __all__

    def test_module_all_contains_param_error(self):
        """测试 __all__ 包含 ParamError"""
        from app.common.errors import __all__

        assert "ParamError" in __all__

    def test_module_all_contains_agent_permission_error(self):
        """测试 __all__ 包含 AgentPermissionError"""
        from app.common.errors import __all__

        assert "AgentPermissionError" in __all__

    def test_module_all_contains_external_service_error(self):
        """测试 __all__ 包含 ExternalServiceError"""
        from app.common.errors import __all__

        assert "ExternalServiceError" in __all__

    def test_module_all_contains_base_exception(self):
        """测试 __all__ 包含 BaseException"""
        from app.common.errors import __all__

        assert "BaseException" in __all__

    def test_module_all_contains_code_exception(self):
        """测试 __all__ 包含 CodeException"""
        from app.common.errors import __all__

        assert "CodeException" in __all__

    def test_module_all_contains_param_exception(self):
        """测试 __all__ 包含 ParamException"""
        from app.common.errors import __all__

        assert "ParamException" in __all__
