# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/common.py 补充测试"""


class TestParseKwargsExtended:
    """测试 parse_kwargs 函数扩展"""

    def test_parse_kwargs_with_tool_input_and_props(self):
        """测试传入 tool_input 和 props 参数"""
        from app.common.tool_v2.common import parse_kwargs

        result, props = parse_kwargs(
            tool_input={"key": "value"}, props={"prop": "data"}
        )

        assert result == {"key": "value"}
        assert props == {"prop": "data"}

    def test_parse_kwargs_without_tool_input(self):
        """测试不传入 tool_input 参数"""
        from app.common.tool_v2.common import parse_kwargs

        result, props = parse_kwargs(key="value", name="test")

        assert result == {"key": "value", "name": "test"}
        assert props is None

    def test_parse_kwargs_with_none_props(self):
        """测试 props 为 None"""
        from app.common.tool_v2.common import parse_kwargs

        result, props = parse_kwargs(tool_input={"key": "value"}, props=None)

        assert result == {"key": "value"}
        assert props is None

    def test_parse_kwargs_with_nested_dict(self):
        """测试嵌套字典"""
        from app.common.tool_v2.common import parse_kwargs

        result, props = parse_kwargs(data={"nested": {"key": "value"}})

        assert result == {"data": {"nested": {"key": "value"}}}
        assert props is None


class TestAPIToolResponse:
    """测试 APIToolResponse 类"""

    def test_init_default_values(self):
        """测试默认值初始化"""
        from app.common.tool_v2.common import APIToolResponse

        response = APIToolResponse()

        assert response.answer == ""
        assert response.block_answer == ""

    def test_init_with_values(self):
        """测试带值初始化"""
        from app.common.tool_v2.common import APIToolResponse

        response = APIToolResponse(answer="test answer", block_answer="block answer")

        assert response.answer == "test answer"
        assert response.block_answer == "block answer"

    def test_to_dict(self):
        """测试转换为字典"""
        from app.common.tool_v2.common import APIToolResponse

        response = APIToolResponse(answer="test answer", block_answer="block answer")
        result = response.to_dict()

        assert result == {"answer": "test answer", "block_answer": "block answer"}

    def test_to_dict_empty(self):
        """测试空响应转换为字典"""
        from app.common.tool_v2.common import APIToolResponse

        response = APIToolResponse()
        result = response.to_dict()

        assert result == {"answer": "", "block_answer": ""}

    def test_with_complex_answer(self):
        """测试复杂答案类型"""
        from app.common.tool_v2.common import APIToolResponse

        response = APIToolResponse(
            answer={"result": "success", "data": [1, 2, 3]},
            block_answer={"partial": True},
        )

        assert response.answer["result"] == "success"
        assert response.block_answer["partial"] is True


class TestToolMapInfo:
    """测试 ToolMapInfo 类"""

    def test_init_required_fields(self):
        """测试必填字段初始化"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test_input", input_type="string")

        assert info.input_name == "test_input"
        assert info.input_type == "string"
        assert info.map_type is None
        assert info.map_value is None
        assert info.enable is None
        assert info.children is None

    def test_init_all_fields(self):
        """测试所有字段初始化"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(
            input_name="test_input",
            input_type="string",
            map_type="fixedValue",
            map_value="test_value",
            enable=True,
            children=[],
        )

        assert info.input_name == "test_input"
        assert info.input_type == "string"
        assert info.map_type == "fixedValue"
        assert info.map_value == "test_value"
        assert info.enable is True
        assert info.children == []

    def test_is_enabled_true(self):
        """测试 is_enabled 返回 True"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test", input_type="string", enable=True)
        assert info.is_enabled() is True

    def test_is_enabled_none(self):
        """测试 is_enabled enable 为 None 时返回 True"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test", input_type="string")
        assert info.is_enabled() is True

    def test_is_enabled_false(self):
        """测试 is_enabled 返回 False"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test", input_type="string", enable=False)
        assert info.is_enabled() is False

    def test_get_map_value_with_value(self):
        """测试 get_map_value 有值时"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(
            input_name="test", input_type="string", map_value="test_value"
        )
        assert info.get_map_value() == "test_value"

    def test_get_map_value_without_value(self):
        """测试 get_map_value 无值时返回空字符串"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test", input_type="string")
        assert info.get_map_value() == ""

    def test_get_map_type_with_value(self):
        """测试 get_map_type 有值时"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(
            input_name="test", input_type="string", map_type="fixedValue"
        )
        assert info.get_map_type() == "fixedValue"

    def test_get_map_type_without_value(self):
        """测试 get_map_type 无值时返回 auto"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(input_name="test", input_type="string")
        assert info.get_map_type() == "auto"

    def test_children_nested(self):
        """测试嵌套 children"""
        from app.common.tool_v2.common import ToolMapInfo

        child = ToolMapInfo(input_name="child", input_type="string")
        parent = ToolMapInfo(input_name="parent", input_type="object", children=[child])

        assert len(parent.children) == 1
        assert parent.children[0].input_name == "child"

    def test_map_types(self):
        """测试所有 map_type 类型"""
        from app.common.tool_v2.common import ToolMapInfo

        for map_type in ["fixedValue", "var", "model", "auto"]:
            info = ToolMapInfo(
                input_name="test", input_type="string", map_type=map_type
            )
            assert info.get_map_type() == map_type

    def test_model_dump(self):
        """测试 model_dump 方法"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(
            input_name="test",
            input_type="string",
            map_type="fixedValue",
            map_value="value",
        )

        data = info.model_dump()
        assert data["input_name"] == "test"
        assert data["input_type"] == "string"
        assert data["map_type"] == "fixedValue"
        assert data["map_value"] == "value"

    def test_extra_fields_allowed(self):
        """测试允许额外字段"""
        from app.common.tool_v2.common import ToolMapInfo

        info = ToolMapInfo(
            input_name="test", input_type="string", extra_field="extra_value"
        )

        # 额外字段应该被保留
        assert hasattr(info, "extra_field")


class TestColors:
    """测试 COLORS 常量"""

    def test_colors_defined(self):
        """测试颜色常量是否定义"""
        from app.common.tool_v2.common import COLORS

        assert "header" in COLORS
        assert "blue" in COLORS
        assert "cyan" in COLORS
        assert "green" in COLORS
        assert "yellow" in COLORS
        assert "red" in COLORS
        assert "bold" in COLORS
        assert "underline" in COLORS
        assert "end" in COLORS

    def test_colors_are_ansi_codes(self):
        """测试颜色是否为 ANSI 转义码"""
        from app.common.tool_v2.common import COLORS

        for color_name, color_code in COLORS.items():
            assert isinstance(color_code, str)
            assert color_code.startswith("\033[")

    def test_end_code(self):
        """测试结束码"""
        from app.common.tool_v2.common import COLORS

        assert COLORS["end"] == "\033[0m"


class TestModuleImports:
    """测试模块导入"""

    def test_import_parse_kwargs(self):
        """测试导入 parse_kwargs"""
        from app.common.tool_v2.common import parse_kwargs

        assert callable(parse_kwargs)

    def test_import_api_tool_response(self):
        """测试导入 APIToolResponse"""
        from app.common.tool_v2.common import APIToolResponse

        assert APIToolResponse is not None

    def test_import_tool_map_info(self):
        """测试导入 ToolMapInfo"""
        from app.common.tool_v2.common import ToolMapInfo

        assert ToolMapInfo is not None

    def test_import_colors(self):
        """测试导入 COLORS"""
        from app.common.tool_v2.common import COLORS

        assert COLORS is not None

    def test_from_package_import(self):
        """测试从包导入"""
        from app.common.tool_v2 import (
            parse_kwargs,
            APIToolResponse,
            ToolMapInfo,
            COLORS,
        )

        assert callable(parse_kwargs)
        assert APIToolResponse is not None
        assert ToolMapInfo is not None
        assert COLORS is not None
