"""Pytest 插件 - 自动设置 Dolphin Mock

确保在所有测试运行前，dolphin SDK 的所有模块都被正确 mock。
"""

import sys
from unittest.mock import MagicMock


def pytest_configure(config):
    """Pytest 配置钩子 - 在测试开始前调用"""
    setup_all_dolphin_mocks()


def setup_all_dolphin_mocks():
    """设置所有需要的 Dolphin SDK mock 模块"""

    # 完整的模块列表
    dolphin_modules = {
        "dolphin": None,
        "dolphin.core": None,
        "dolphin.core.common": None,
        "dolphin.core.common.constants": {
            "KEY_SESSION_ID": "session_id",
            "KEY_USER_ID": "user_id",
            "KEY_AGENT_ID": "agent_id",
        },
        "dolphin.core.common.exceptions": {
            "ModelException": type("ModelException", (Exception,), {}),
            "SkillException": type("SkillException", (Exception,), {}),
            "DolphinException": type("DolphinException", (Exception,), {}),
        },
        "dolphin.core.context": None,
        "dolphin.core.context.context": None,
        "dolphin.core.context.var_output": {
            "VarOutput": create_mock_var_output_class(),
        },
        "dolphin.core.coroutine": None,
        "dolphin.core.coroutine.resume_handle": {
            "ResumeHandle": create_mock_resume_handle_class(),
        },
        "dolphin.core.utils": None,
        "dolphin.core.utils.tools": {
            "Tool": type("Tool", (object,), {}),
        },
    }

    for module_name, attributes in dolphin_modules.items():
        if module_name not in sys.modules:
            mock_module = MagicMock()
            mock_module.__name__ = module_name
            mock_module.__file__ = f"{module_name.replace('.', '/')}/__init__.py"
            mock_module.__path__ = []  # 使其成为包

            # 添加特定属性
            if attributes:
                for attr_name, attr_value in attributes.items():
                    setattr(mock_module, attr_name, attr_value)

            sys.modules[module_name] = mock_module


def create_mock_var_output_class():
    """创建模拟的 VarOutput 类"""

    class MockVarOutput:
        _storage = {}

        @classmethod
        def reset(cls):
            """重置存储"""
            cls._storage = {}

        def set(self, key, value):
            """设置变量"""
            MockVarOutput._storage[key] = value

        def get(self, key, default=None):
            """获取变量"""
            return MockVarOutput._storage.get(key, default)

        def delete(self, key):
            """删除变量"""
            MockVarOutput._storage.pop(key, None)

        @classmethod
        def is_serialized_dict(cls, var):
            """检查是否为序列化的字典"""
            return isinstance(var, dict) and "__dolphin_var__" in var

    return MockVarOutput


def create_mock_resume_handle_class():
    """创建模拟的 ResumeHandle 类"""

    class MockResumeHandle:
        def __init__(
            self,
            frame_id="",
            snapshot_id="",
            resume_token="",
            interrupt_type="",
            current_block="",
            restart_block="",
        ):
            self.frame_id = frame_id
            self.snapshot_id = snapshot_id
            self.resume_token = resume_token
            self.interrupt_type = interrupt_type
            self.current_block = current_block
            self.restart_block = restart_block

    return MockResumeHandle
