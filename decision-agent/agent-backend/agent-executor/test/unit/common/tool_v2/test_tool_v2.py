# -*- coding: utf-8 -*-
"""
Unit tests for app/common/tool_v2 module
"""

import pytest


class TestModuleImports:
    """Tests for module imports"""

    @pytest.mark.asyncio
    async def test_tool_v2_init_exports(self):
        """Test that tool_v2 __init__ exports all required items"""
        from app.common.tool_v2 import (
            parse_kwargs,
            APIToolResponse,
            ToolMapInfo,
            COLORS,
            APITool,
            AgentTool,
            MCPTool,
            get_mcp_tools,
            build_tools,
        )

        assert callable(parse_kwargs)
        assert APIToolResponse is not None
        assert ToolMapInfo is not None
        assert COLORS is not None
        assert APITool is not None
        assert AgentTool is not None
        assert MCPTool is not None
        assert callable(get_mcp_tools)
        assert callable(build_tools)

    @pytest.mark.asyncio
    async def test_module_imports(self):
        """Test that tool_v2 module can be imported"""
        from app.common.tool_v2 import tool

        assert tool is not None
        assert hasattr(tool, "build_tools")
