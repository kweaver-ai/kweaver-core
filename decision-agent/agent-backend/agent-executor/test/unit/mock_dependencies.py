#!/usr/bin/env python3
"""
模拟缺失的依赖项，用于单元测试
"""

import sys
from unittest.mock import Mock


# 模拟 DolphinLanguageSDK
class MockTool:
    """模拟 Tool 基类"""

    def __init__(self, *args, **kwargs):
        pass

    # 添加必需的方法和属性
    name = "mock_tool"
    description = "Mock tool for testing"

    async def arun_stream(self, *args, **kwargs):
        yield "Mock response"


class MockToolInterrupt(Exception):
    """模拟 ToolInterrupt 异常"""

    def __init__(self, tool_name, tool_args):
        self.tool_name = tool_name
        self.tool_args = tool_args
        super().__init__(f"Tool interrupt: {tool_name}")


class MockContext:
    """模拟 Context 类"""

    def __init__(self):
        self.variables = {}

    def get_var_value(self, key):
        return self.variables.get(key, {})

    def get_all_variables(self):
        return self.variables


# 模拟 DolphinLanguageSDK.utils 模块
dolphin_utils_tools = Mock()
dolphin_utils_tools.Tool = MockTool
dolphin_utils_tools.ToolInterrupt = MockToolInterrupt

dolphin_utils_handle_progress = Mock()
dolphin_utils_handle_progress.handle_progress = Mock(return_value=[])
dolphin_utils_handle_progress.cleanup_progress = Mock()

dolphin_utils_data_process = Mock()
dolphin_utils_data_process.get_nested_value = Mock(return_value=None)

dolphin_context = Mock()
dolphin_context.Context = MockContext

# 模拟 app.common.stand_log
app_common_stand_log = Mock()
stand_logger = Mock()
stand_logger.info = Mock()
stand_logger.error = Mock()
stand_logger.warn = Mock()
stand_logger.setLevel = Mock()
app_common_stand_log.StandLogger = stand_logger

# 将模拟模块添加到 sys.modules
sys.modules["DolphinLanguageSDK"] = Mock()
sys.modules["DolphinLanguageSDK.utils"] = Mock()
sys.modules["DolphinLanguageSDK.utils.tools"] = dolphin_utils_tools
sys.modules["DolphinLanguageSDK.utils.handle_progress"] = dolphin_utils_handle_progress
sys.modules["DolphinLanguageSDK.utils.data_process"] = dolphin_utils_data_process
sys.modules["DolphinLanguageSDK.context"] = dolphin_context
# 只模拟特定的子模块，不要模拟整个app和app.common
# sys.modules['app'] = Mock()  # 注释掉，这会阻止真实模块导入
# sys.modules['app.common'] = Mock()  # 注释掉，这会阻止真实模块导入
sys.modules["app.common.stand_log"] = app_common_stand_log
