# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool_pkg/process_fixed_value.py"""

from unittest.mock import patch


class TestProcessFixedValue:
    """测试 process_fixed_value 函数"""

    def test_process_fixed_value_string_type(self):
        """测试字符串类型的固定值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="test_param",
            input_type="string",
            map_type="fixedValue",
            map_value="test_value",
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["test_param"] == "test_value"

    def test_process_fixed_value_json_string(self):
        """测试 JSON 字符串类型的固定值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="json_param",
            input_type="object",
            map_type="fixedValue",
            map_value='{"key": "value"}',
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["json_param"] == {"key": "value"}

    def test_process_fixed_value_with_quotes(self):
        """测试带引号的字符串"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="quoted_param",
            input_type="string",
            map_type="fixedValue",
            map_value='"quoted_value"',
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["quoted_param"] == "quoted_value"

    def test_process_fixed_value_invalid_json(self):
        """测试无效 JSON 字符串"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="invalid_json_param",
            input_type="object",
            map_type="fixedValue",
            map_value="not a valid json",
        )
        current_tool_input = {}
        input_params = {}

        with patch("app.common.tool_v2.api_tool_pkg.process_fixed_value.StandLogger"):
            process_fixed_value(tool_map_item, current_tool_input, input_params)

        # 无效 JSON 应该保持原值
        assert current_tool_input["invalid_json_param"] == "not a valid json"

    def test_process_fixed_value_integer(self):
        """测试整数类型的固定值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="int_param",
            input_type="integer",
            map_type="fixedValue",
            map_value="123",
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["int_param"] == 123

    def test_process_fixed_value_array(self):
        """测试数组类型的固定值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="array_param",
            input_type="array",
            map_type="fixedValue",
            map_value="[1, 2, 3]",
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["array_param"] == [1, 2, 3]

    def test_process_fixed_value_boolean(self):
        """测试布尔类型的固定值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="bool_param",
            input_type="boolean",
            map_type="fixedValue",
            map_value="true",
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["bool_param"] is True

    def test_process_fixed_value_overwrite_existing(self):
        """测试覆盖已存在的值"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="existing_param",
            input_type="string",
            map_type="fixedValue",
            map_value="new_value",
        )
        current_tool_input = {"existing_param": "old_value"}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["existing_param"] == "new_value"

    def test_process_fixed_value_empty_string(self):
        """测试空字符串"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )
        from app.common.tool_v2.common import ToolMapInfo

        tool_map_item = ToolMapInfo(
            input_name="empty_param",
            input_type="string",
            map_type="fixedValue",
            map_value="",
        )
        current_tool_input = {}
        input_params = {}

        process_fixed_value(tool_map_item, current_tool_input, input_params)

        assert current_tool_input["empty_param"] == ""


class TestModuleImports:
    """测试模块导入"""

    def test_import_process_fixed_value(self):
        """测试导入 process_fixed_value"""
        from app.common.tool_v2.api_tool_pkg.process_fixed_value import (
            process_fixed_value,
        )

        assert callable(process_fixed_value)
