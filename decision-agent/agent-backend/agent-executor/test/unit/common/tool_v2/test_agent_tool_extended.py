# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/agent_tool.py 补充测试"""

from unittest.mock import MagicMock


class TestAgentToolInit:
    """测试 AgentTool 初始化"""

    def test_init_basic(self):
        """测试基本初始化"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {
                    "name": "test_agent",
                    "profile": "Test agent profile",
                }
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert tool.name == "test_agent"
        assert tool.description == "Test agent profile"
        assert tool.intervention is False

    def test_init_with_intervention(self):
        """测试带干预配置的初始化"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test_agent"}
                self.agent_input = []
                self.intervention = True
                self.intervention_confirmation_message = "Confirm agent execution?"
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert tool.intervention is True
        assert tool.interrupt_config["requires_confirmation"] is True
        assert (
            tool.interrupt_config["confirmation_message"] == "Confirm agent execution?"
        )

    def test_init_default_name(self):
        """测试默认名称 - 当 agent_info 为空时，名称为空字符串"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {}
                self.agent_key = "agent_key_123"
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        # 当 agent_info 没有 name 时，使用空字符串
        assert tool.name == ""

    def test_init_default_intervention_message(self):
        """测试默认干预消息"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "my_agent"}
                self.agent_input = []
                self.intervention = True
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert (
            tool.interrupt_config["confirmation_message"]
            == "Agent工具 my_agent 需要确认执行"
        )

    def test_init_with_agent_options(self):
        """测试带 agent_options 的初始化"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.inner_dto.agent_options = {"option1": "value1"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert tool.agent_options == {"option1": "value1"}


class TestAgentToolParseInputs:
    """测试 _parse_agent_inputs 方法"""

    def test_parse_agent_inputs_empty(self):
        """测试空输入"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())
        result = tool._parse_agent_inputs(MockAgentSkill())

        assert result == {}

    def test_parse_agent_inputs_with_auto_type(self):
        """测试 auto 类型输入"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockInput:
            def __init__(self):
                self.enable = True
                self.map_type = "auto"
                self.input_name = "query"
                self.input_type = "string"
                self.input_desc = "Query parameter"

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = [MockInput()]
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert "query" in tool.inputs
        assert tool.inputs["query"]["type"] == "string"
        assert tool.inputs["query"]["required"] is True

    def test_parse_agent_inputs_disabled(self):
        """测试禁用的输入"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockInput:
            def __init__(self):
                self.enable = False
                self.map_type = "auto"
                self.input_name = "disabled_param"

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = [MockInput()]
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        assert "disabled_param" not in tool.inputs


class TestAgentToolHelperMethods:
    """测试辅助方法"""

    def test_is_explore_var_true(self):
        """测试 _is_explore_var 返回 True"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        var = [
            {
                "agent_name": "test",
                "stage": "stage1",
                "answer": "answer1",
                "think": "think1",
                "status": "done",
                "skill_info": {},
                "block_answer": "",
                "input_message": "",
                "interrupted": False,
            }
        ]

        assert tool._is_explore_var(var) is True

    def test_is_explore_var_false(self):
        """测试 _is_explore_var 返回 False"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        # 普通字典
        assert tool._is_explore_var({"answer": "test"}) is False
        # 普通字符串
        assert tool._is_explore_var("string") is False
        # 缺少键的列表
        assert tool._is_explore_var([{"answer": "test"}]) is False

    def test_is_llm_var_true(self):
        """测试 _is_llm_var 返回 True"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        var = {"answer": "answer", "think": "think"}
        assert tool._is_llm_var(var) is True

    def test_is_llm_var_false(self):
        """测试 _is_llm_var 返回 False"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        # 多余的键
        assert tool._is_llm_var({"answer": "a", "think": "t", "extra": "e"}) is False
        # 缺少键
        assert tool._is_llm_var({"answer": "a"}) is False
        # 非字典
        assert tool._is_llm_var("string") is False

    def test_get_dolphin_var_value_explore(self):
        """测试 _get_dolphin_var_value 探索变量"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        var = [
            {
                "agent_name": "test",
                "stage": "stage1",
                "answer": "final answer",
                "think": "think1",
                "status": "done",
                "skill_info": {},
                "block_answer": "",
                "input_message": "",
                "interrupted": False,
            }
        ]

        result = tool._get_dolphin_var_value(var)
        assert result == "final answer"

    def test_get_dolphin_var_value_llm(self):
        """测试 _get_dolphin_var_value LLM 变量"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        var = {"answer": "llm answer", "think": "llm think"}
        result = tool._get_dolphin_var_value(var)
        assert result == "llm answer"

    def test_get_dolphin_var_value_simple(self):
        """测试 _get_dolphin_var_value 简单变量"""
        from app.common.tool_v2.agent_tool import AgentTool

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {"name": "test"}
                self.agent_input = []
                self.intervention = False
                self.agent_timeout = 1800

        tool = AgentTool(mock_ac, MockAgentSkill())

        result = tool._get_dolphin_var_value("simple string")
        assert result == "simple string"

        result = tool._get_dolphin_var_value(123)
        assert result == 123


class TestModuleImports:
    """测试模块导入"""

    def test_import_agent_tool(self):
        """测试导入 AgentTool"""
        from app.common.tool_v2.agent_tool import AgentTool

        assert AgentTool is not None

    def test_import_from_package(self):
        """测试从包导入"""
        from app.common.tool_v2 import AgentTool

        assert AgentTool is not None
