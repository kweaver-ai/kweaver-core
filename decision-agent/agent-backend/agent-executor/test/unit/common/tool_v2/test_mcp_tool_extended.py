# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/mcp_tool.py 补充测试"""


class TestMCPToolInit:
    """测试 MCPTool 初始化"""

    def test_init_basic(self):
        """测试基本初始化"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test_tool", "description": "Test description"}
        mcp_config = {"mcp_server_id": "test_server"}

        tool = MCPTool(mcp_tool_info, mcp_config)

        assert tool.name == "test_tool"
        assert tool.description == "Test description"
        assert tool.mcp_server_id == "test_server"

    def test_init_with_input_schema(self):
        """测试带输入规范的初始化"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {
            "name": "test_tool",
            "inputSchema": {
                "type": "object",
                "properties": {
                    "param1": {"type": "string", "description": "Parameter 1"},
                    "param2": {"type": "number", "description": "Parameter 2"},
                },
                "required": ["param1"],
            },
        }
        mcp_config = {}

        tool = MCPTool(mcp_tool_info, mcp_config)

        assert "param1" in tool.inputs
        assert tool.inputs["param1"]["type"] == "string"
        assert tool.inputs["param1"]["required"] is True
        assert tool.inputs["param2"]["required"] is False

    def test_init_default_name(self):
        """测试默认名称"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {}
        mcp_config = {}

        tool = MCPTool(mcp_tool_info, mcp_config)

        assert tool.name == "unknown"

    def test_init_outputs(self):
        """测试输出定义"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test_tool"}
        mcp_config = {}

        tool = MCPTool(mcp_tool_info, mcp_config)

        assert "result" in tool.outputs


class TestMCPToolParseMCPInputs:
    """测试 _parse_mcp_inputs 方法"""

    def test_parse_empty_schema(self):
        """测试空规范"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        result = tool._parse_mcp_inputs({})

        assert result == {}

    def test_parse_simple_properties(self):
        """测试简单属性"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        input_schema = {
            "type": "object",
            "properties": {
                "str_prop": {"type": "string", "description": "String prop"},
                "num_prop": {"type": "number", "description": "Number prop"},
            },
        }

        result = tool._parse_mcp_inputs(input_schema)

        assert "str_prop" in result
        assert result["str_prop"]["type"] == "string"
        assert result["str_prop"]["description"] == "String prop"
        assert result["num_prop"]["type"] == "number"

    def test_parse_with_required(self):
        """测试必填字段"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        input_schema = {
            "type": "object",
            "properties": {
                "required_field": {"type": "string"},
                "optional_field": {"type": "string"},
            },
            "required": ["required_field"],
        }

        result = tool._parse_mcp_inputs(input_schema)

        assert result["required_field"]["required"] is True
        assert result["optional_field"]["required"] is False

    def test_parse_missing_type(self):
        """测试缺少类型时使用默认值"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        input_schema = {
            "type": "object",
            "properties": {"no_type": {"description": "No type specified"}},
        }

        result = tool._parse_mcp_inputs(input_schema)

        assert result["no_type"]["type"] == "string"

    def test_parse_missing_description(self):
        """测试缺少描述时使用默认值"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        input_schema = {"type": "object", "properties": {"no_desc": {"type": "number"}}}

        result = tool._parse_mcp_inputs(input_schema)

        assert result["no_desc"]["description"] == ""


class TestMCPToolResolveMCPRefs:
    """测试 _resolve_mcp_refs_recursively 方法"""

    def test_resolve_non_dict(self):
        """测试非字典值"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        result = tool._resolve_mcp_refs_recursively("string", {})
        assert result == "string"

        result = tool._resolve_mcp_refs_recursively(123, {})
        assert result == 123

    def test_resolve_simple_dict(self):
        """测试简单字典"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        schema = {"type": "string", "description": "Simple"}
        result = tool._resolve_mcp_refs_recursively(schema, {})

        assert result == {"type": "string", "description": "Simple"}

    def test_resolve_with_defs_ref(self):
        """测试 $defs 引用"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        schema = {"$ref": "#/$defs/MyDef"}
        input_schema = {
            "$defs": {"MyDef": {"type": "string", "description": "Defined type"}}
        }

        result = tool._resolve_mcp_refs_recursively(schema, input_schema)

        assert result["type"] == "string"
        assert result["description"] == "Defined type"

    def test_resolve_nested_dict(self):
        """测试嵌套字典"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        schema = {
            "type": "object",
            "properties": {
                "nested": {
                    "type": "object",
                    "properties": {"inner": {"type": "string"}},
                }
            },
        }

        result = tool._resolve_mcp_refs_recursively(schema, {})

        assert "nested" in result["properties"]
        assert "inner" in result["properties"]["nested"]["properties"]

    def test_resolve_with_list(self):
        """测试列表"""
        from app.common.tool_v2.mcp_tool import MCPTool

        mcp_tool_info = {"name": "test"}
        mcp_config = {}
        tool = MCPTool(mcp_tool_info, mcp_config)

        schema = {"type": "array", "items": [{"type": "string"}, {"type": "number"}]}

        result = tool._resolve_mcp_refs_recursively(schema, {})

        assert result["type"] == "array"
        assert len(result["items"]) == 2


class TestModuleImports:
    """测试模块导入"""

    def test_import_mcp_tool(self):
        """测试导入 MCPTool"""
        from app.common.tool_v2.mcp_tool import MCPTool

        assert MCPTool is not None

    def test_import_get_mcp_tools(self):
        """测试导入 get_mcp_tools"""
        from app.common.tool_v2.mcp_tool import get_mcp_tools

        assert callable(get_mcp_tools)

    def test_import_from_package(self):
        """测试从包导入"""
        from app.common.tool_v2 import MCPTool, get_mcp_tools

        assert MCPTool is not None
        assert callable(get_mcp_tools)
