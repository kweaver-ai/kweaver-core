# -*- coding: utf-8 -*-
"""单元测试 - process_skill_pkg/mcp_skills 模块"""

import pytest
from unittest.mock import MagicMock


class TestProcessSkillsMcps:
    """测试 process_skills_mcps 函数"""

    @pytest.fixture
    def mock_mcp(self):
        """创建 mock mcp"""
        mcp = MagicMock()
        mcp.__dict__ = {}
        return mcp

    @pytest.fixture
    def mock_skills(self):
        """创建 mock SkillVo"""
        skills = MagicMock()
        skills.mcps = []
        return skills

    @pytest.mark.asyncio
    async def test_process_skills_mcps_empty(self, mock_skills):
        """测试空mcps列表"""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.mcp_skills import (
            process_skills_mcps,
        )

        await process_skills_mcps(mock_skills)

    @pytest.mark.asyncio
    async def test_process_skills_mcps_with_mcp(self, mock_skills, mock_mcp):
        """测试处理mcp"""
        mock_skills.mcps = [mock_mcp]

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.mcp_skills import (
            process_skills_mcps,
        )

        await process_skills_mcps(mock_skills)

        # 验证HOST和PORT被设置到mcp.__dict__
        assert "HOST_AGENT_OPERATOR" in mock_mcp.__dict__
        assert "PORT_AGENT_OPERATOR" in mock_mcp.__dict__

    @pytest.mark.asyncio
    async def test_process_skills_mcps_none_skills(self):
        """测试skills为None的情况"""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.mcp_skills import (
            process_skills_mcps,
        )

        await process_skills_mcps(None)

    @pytest.mark.asyncio
    async def test_process_skills_mcps_multiple_mcps(self, mock_skills):
        """测试多个mcps处理"""
        mock_mcp1 = MagicMock()
        mock_mcp1.__dict__ = {}
        mock_mcp2 = MagicMock()
        mock_mcp2.__dict__ = {}
        mock_skills.mcps = [mock_mcp1, mock_mcp2]

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.mcp_skills import (
            process_skills_mcps,
        )

        await process_skills_mcps(mock_skills)

        # 验证所有mcp都被处理
        assert "HOST_AGENT_OPERATOR" in mock_mcp1.__dict__
        assert "HOST_AGENT_OPERATOR" in mock_mcp2.__dict__
