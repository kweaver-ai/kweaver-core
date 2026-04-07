# -*- coding: utf-8 -*-
"""单元测试 - app/common/tool_v2/tool.py"""

import pytest
from unittest.mock import MagicMock, patch


class TestBuildTools:
    """测试 build_tools 函数"""

    @pytest.mark.asyncio
    async def test_build_tools_empty_skills(self):
        """测试空技能列表"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()
        mock_skills = MagicMock()
        mock_skills.tools = None
        mock_skills.agents = None
        mock_skills.mcps = None

        with patch("app.common.tool_v2.tool.internal_span"):
            result = await build_tools(mock_ac, mock_skills)

        assert result == {}

    @pytest.mark.asyncio
    async def test_build_tools_with_api_tools(self):
        """测试构建 API 工具"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()

        # 创建一个简单的对象来模拟 tool
        class MockTool:
            def __init__(self):
                self.tool_id = "tool1"
                self.tool_info = {
                    "name": "test_tool",
                    "description": "Test tool",
                    "metadata": {"api_spec": {}},
                }
                self.tool_input = []
                self.intervention = False

        mock_skills = MagicMock()
        mock_skills.tools = [MockTool()]
        mock_skills.agents = None
        mock_skills.mcps = None

        with patch("app.common.tool_v2.tool.internal_span"):
            with patch("app.common.tool_v2.tool.APITool") as mock_api_tool:
                mock_tool_instance = MagicMock()
                mock_tool_instance.name = "test_tool"
                mock_api_tool.return_value = mock_tool_instance
                result = await build_tools(mock_ac, mock_skills)

        assert "test_tool" in result

    @pytest.mark.asyncio
    async def test_build_tools_with_agent_tools(self):
        """测试构建 Agent 工具"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()

        class MockAgentSkill:
            def __init__(self):
                self.agent_key = "agent1"
                self.intervention = False
                self.agent_input = []
                self.inner_dto = MagicMock()
                self.inner_dto.agent_info = {
                    "name": "test_agent",
                    "profile": "Test agent",
                }

        mock_skills = MagicMock()
        mock_skills.tools = None
        mock_skills.agents = [MockAgentSkill()]
        mock_skills.mcps = None

        with patch("app.common.tool_v2.tool.internal_span"):
            with patch("app.common.tool_v2.tool.AgentTool") as mock_agent_tool:
                mock_tool_instance = MagicMock()
                mock_tool_instance.name = "test_agent"
                mock_agent_tool.return_value = mock_tool_instance
                result = await build_tools(mock_ac, mock_skills)

        assert "test_agent" in result

    @pytest.mark.asyncio
    async def test_build_tools_duplicate_names(self):
        """测试重复名称处理"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()

        class MockTool:
            def __init__(self, tool_id, name):
                self.tool_id = tool_id
                self.tool_info = {"name": name, "metadata": {"api_spec": {}}}
                self.tool_input = []
                self.intervention = False

        mock_skills = MagicMock()
        mock_skills.tools = [
            MockTool("tool1", "same_name"),
            MockTool("tool2", "same_name"),
        ]
        mock_skills.agents = None
        mock_skills.mcps = None

        with patch("app.common.tool_v2.tool.internal_span"):
            with patch("app.common.tool_v2.tool.StandLogger"):
                with patch("app.common.tool_v2.tool.APITool") as mock_api_tool:
                    mock_tool_instance = MagicMock()
                    mock_tool_instance.name = "same_name"
                    mock_api_tool.return_value = mock_tool_instance
                    result = await build_tools(mock_ac, mock_skills)

        # 第二个同名工具会覆盖第一个
        assert "same_name" in result

    @pytest.mark.asyncio
    async def test_build_tools_with_mcp_tools(self):
        """测试构建 MCP 工具"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()

        class MockMCP:
            def __init__(self):
                self.mcp_server_id = "mcp1"
                self.HOST_AGENT_OPERATOR = "localhost"
                self.PORT_AGENT_OPERATOR = "9000"

        mock_skills = MagicMock()
        mock_skills.tools = None
        mock_skills.agents = None
        mock_skills.mcps = [MockMCP()]

        with patch("app.common.tool_v2.tool.internal_span"):
            with patch("app.common.tool_v2.tool.get_mcp_tools") as mock_get_mcp:
                mock_tool_instance = MagicMock()
                mock_tool_instance.name = "mcp_tool"
                mock_get_mcp.return_value = {"mcp_tool": mock_tool_instance}
                result = await build_tools(mock_ac, mock_skills)

        assert "mcp_tool" in result

    @pytest.mark.asyncio
    async def test_build_tools_tool_without_name(self):
        """测试无名称的工具"""
        from app.common.tool_v2.tool import build_tools

        mock_ac = MagicMock()

        class MockTool:
            def __init__(self):
                self.tool_id = "tool123"
                self.tool_info = {"metadata": {"api_spec": {}}}
                self.tool_input = []
                self.intervention = False

        mock_skills = MagicMock()
        mock_skills.tools = [MockTool()]
        mock_skills.agents = None
        mock_skills.mcps = None

        with patch("app.common.tool_v2.tool.internal_span"):
            with patch("app.common.tool_v2.tool.APITool") as mock_api_tool:
                mock_tool_instance = MagicMock()
                mock_tool_instance.name = "tool_tool123"
                mock_api_tool.return_value = mock_tool_instance
                result = await build_tools(mock_ac, mock_skills)

        # 应该使用 tool_id 作为名称
        assert "tool_tool123" in result


class TestModuleImports:
    """测试模块导入"""

    def test_import_build_tools(self):
        """测试导入 build_tools"""
        from app.common.tool_v2.tool import build_tools

        assert callable(build_tools)
