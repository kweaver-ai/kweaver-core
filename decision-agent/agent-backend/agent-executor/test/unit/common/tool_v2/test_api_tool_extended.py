# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/api_tool.py 补充测试"""


class TestAPIToolInit:
    """测试 APITool 初始化"""

    def test_init_basic(self):
        """测试基本初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test_tool", "description": "Test description"}
        tool_config = {"tool_input": [], "intervention": False}

        tool = APITool(tool_info, tool_config)

        assert tool.name == "test_tool"
        assert tool.description == "Test description"
        assert tool.intervention is False

    def test_init_with_use_rule(self):
        """测试带 use_rule 的初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {
            "name": "test_tool",
            "description": "Test description",
            "use_rule": "Use this tool for testing",
        }
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)

        assert "Use Rule:" in tool.description
        assert "Use this tool for testing" in tool.description

    def test_init_with_intervention(self):
        """测试带干预配置的初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test_tool", "description": "Test"}
        tool_config = {
            "tool_input": [],
            "intervention": True,
            "intervention_confirmation_message": "Confirm execution?",
        }

        tool = APITool(tool_info, tool_config)

        assert tool.intervention is True
        assert tool.interrupt_config["requires_confirmation"] is True
        assert tool.interrupt_config["confirmation_message"] == "Confirm execution?"

    def test_init_with_tool_input(self):
        """测试带 tool_input 的初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {
            "name": "test_tool",
            "description": "Test",
            "metadata": {"api_spec": {}},
        }
        tool_config = {
            "tool_input": [
                {"input_name": "param1", "input_type": "string", "map_type": "auto"}
            ]
        }

        tool = APITool(tool_info, tool_config)

        assert len(tool.tool_map_list) == 1
        assert tool.tool_map_list[0].input_name == "param1"

    def test_init_with_result_process_strategies(self):
        """测试带结果处理策略的初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test_tool", "metadata": {"api_spec": {}}}
        tool_config = {
            "tool_input": [],
            "result_process_strategies": [
                {"category": {"id": "category1"}, "strategy": {"id": "strategy1"}}
            ],
        }

        tool = APITool(tool_info, tool_config)

        assert hasattr(tool, "result_process_strategy_cfg")
        assert tool.result_process_strategy_cfg[0]["category"] == "category1"
        assert tool.result_process_strategy_cfg[0]["strategy"] == "strategy1"

    def test_init_default_intervention_message(self):
        """测试默认干预消息"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "my_tool"}
        tool_config = {"tool_input": [], "intervention": True}

        tool = APITool(tool_info, tool_config)

        assert (
            tool.interrupt_config["confirmation_message"] == "工具 my_tool 需要确认执行"
        )

    def test_init_no_intervention_config(self):
        """测试无干预配置"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test_tool"}
        tool_config = {"tool_input": [], "intervention": False}

        tool = APITool(tool_info, tool_config)

        assert tool.interrupt_config is None

    def test_init_with_pydantic_tool_input(self):
        """测试带 Pydantic 模型的 tool_input"""
        from app.common.tool_v2.api_tool import APITool
        from app.common.tool_v2.common import ToolMapInfo

        tool_info = {"name": "test_tool", "metadata": {"api_spec": {}}}

        # 使用 Pydantic 模型
        tool_map = ToolMapInfo(
            input_name="param1", input_type="string", map_type="auto"
        )
        tool_config = {"tool_input": [tool_map]}

        tool = APITool(tool_info, tool_config)

        assert len(tool.tool_map_list) == 1

    def test_init_tool_info_without_name(self):
        """测试 tool_info 无名称时使用 tool_id"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"tool_id": "tool_123"}
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)

        assert tool.name == "tool_123"

    def test_init_with_api_spec(self):
        """测试带 API 规范的初始化"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {
            "name": "test_tool",
            "metadata": {
                "api_spec": {
                    "parameters": [
                        {
                            "name": "id",
                            "in": "path",
                            "required": True,
                            "schema": {"type": "string"},
                        }
                    ],
                    "responses": [
                        {
                            "status_code": "200",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "type": "object",
                                        "properties": {"result": {"type": "string"}},
                                    }
                                }
                            },
                        }
                    ],
                }
            },
        }
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)

        assert "id" in tool.inputs


class TestAPIToolParseDescription:
    """测试 _parse_description 方法"""

    def test_parse_description_only(self):
        """测试只有描述"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test"}
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)
        result = tool._parse_description({"description": "Basic description"})

        assert result == "Basic description"

    def test_parse_description_empty(self):
        """测试空描述"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test"}
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)
        result = tool._parse_description({})

        assert result == ""


class TestAPIToolFilterExposedInputs:
    """测试 _filter_exposed_inputs 方法"""

    def test_filter_empty_inputs(self):
        """测试空输入"""
        from app.common.tool_v2.api_tool import APITool

        tool_info = {"name": "test"}
        tool_config = {"tool_input": []}

        tool = APITool(tool_info, tool_config)
        result = tool._filter_exposed_inputs({})

        # 空输入返回空字典
        assert result == {}


class TestModuleImports:
    """测试模块导入"""

    def test_import_api_tool(self):
        """测试导入 APITool"""
        from app.common.tool_v2.api_tool import APITool

        assert APITool is not None

    def test_import_from_package(self):
        """测试从包导入"""
        from app.common.tool_v2 import APITool

        assert APITool is not None
