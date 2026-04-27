# -*- coding: utf-8 -*-
"""单元测试 - process_skill_pkg/index 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestProcessSkills:
    """测试 process_skills 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.agent_config = MagicMock()
        return core

    @pytest.fixture
    def mock_skills(self):
        """创建 mock SkillVo"""
        skills = MagicMock()
        skills.tools = []
        skills.agents = []
        skills.mcps = []
        return skills

    @pytest.mark.asyncio
    async def test_process_skills_basic(self, mock_agent_core, mock_skills):
        """测试基本skills处理"""
        mock_agent_core.agent_config.skills = mock_skills

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_tools",
            new_callable=AsyncMock,
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_agents",
                new_callable=AsyncMock,
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_mcps",
                    new_callable=AsyncMock,
                ):
                    from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index import (
                        process_skills,
                    )

                    await process_skills(
                        mock_agent_core,
                        headers={"x-user-id": "user123"},
                        context_variables={"self_config": {}},
                    )

    @pytest.mark.asyncio
    async def test_process_skills_with_temp_files(self, mock_agent_core, mock_skills):
        """测试带临时文件的skills处理"""
        mock_agent_core.agent_config.skills = mock_skills
        temp_files = {"test_file": {"filename": "test.txt"}}

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_tools",
            new_callable=AsyncMock,
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_agents",
                new_callable=AsyncMock,
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_mcps",
                    new_callable=AsyncMock,
                ):
                    from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index import (
                        process_skills,
                    )

                    await process_skills(
                        mock_agent_core,
                        headers={"x-user-id": "user123"},
                        context_variables={"self_config": {}},
                        temp_files=temp_files,
                    )

    @pytest.mark.asyncio
    async def test_process_skills_none_temp_files(self, mock_agent_core, mock_skills):
        """测试temp_files为None的情况"""
        mock_agent_core.agent_config.skills = mock_skills

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_tools",
            new_callable=AsyncMock,
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_agents",
                new_callable=AsyncMock,
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index.process_skills_mcps",
                    new_callable=AsyncMock,
                ):
                    from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.index import (
                        process_skills,
                    )

                    await process_skills(
                        mock_agent_core,
                        headers={"x-user-id": "user123"},
                        context_variables={"self_config": {}},
                        temp_files=None,
                    )
