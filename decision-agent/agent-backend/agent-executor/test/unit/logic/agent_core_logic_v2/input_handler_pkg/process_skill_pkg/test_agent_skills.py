# -*- coding: utf-8 -*-
"""单元测试 - process_skill_pkg/agent_skills 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestProcessSkillsAgents:
    """测试 process_skills_agents 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.is_warmup = True
        core.cache_handler = MagicMock()
        core.cache_handler.get_skill_agent_info_dict.return_value = None
        return core

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.data_source = {}
        config.llms = [{"is_default": True, "llm_config": {"model": "gpt-4"}}]
        return config

    @pytest.fixture
    def mock_skills(self):
        """创建 mock SkillVo"""
        skills = MagicMock()
        skills.agents = []
        return skills

    @pytest.fixture
    def mock_agent_skill(self):
        """创建 mock AgentSkillVo"""
        skill = MagicMock()
        skill.agent_key = "test_agent_key"
        skill.inner_dto = MagicMock()
        skill.data_source_config = None
        skill.datasource_config = None
        skill.llm_config = None
        return skill

    @pytest.mark.asyncio
    async def test_process_skills_agents_empty(
        self, mock_agent_core, mock_agent_config, mock_skills
    ):
        """测试空agents列表"""
        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.agent_factory_service"
        ):
            from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
                process_skills_agents,
            )

            await process_skills_agents(
                mock_agent_core,
                mock_agent_config,
                mock_skills,
                temp_files={},
                headers={"x-user-id": "user123"},
            )

    @pytest.mark.asyncio
    async def test_process_skills_agents_with_agent(
        self, mock_agent_core, mock_agent_config, mock_skills, mock_agent_skill
    ):
        """测试处理agent skill"""
        mock_skills.agents = [mock_agent_skill]
        mock_agent_core.cache_handler.get_skill_agent_info_dict.return_value = None

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.agent_factory_service"
        ) as mock_service:
            mock_service.get_agent_config_by_key = AsyncMock(
                return_value={"agent_id": "test_agent"}
            )

            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_biz_domain_id",
                        return_value="domain123",
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_user_account_id"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_user_account_type"
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_biz_domain_id"
                                ):
                                    with patch(
                                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.hl_agent_options",
                                        new_callable=AsyncMock,
                                    ):
                                        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
                                            process_skills_agents,
                                        )

                                        await process_skills_agents(
                                            mock_agent_core,
                                            mock_agent_config,
                                            mock_skills,
                                            temp_files={},
                                            headers={"x-user-id": "user123"},
                                        )

    @pytest.mark.asyncio
    async def test_process_skills_agents_with_cached_info(
        self, mock_agent_core, mock_agent_config, mock_skills, mock_agent_skill
    ):
        """测试从缓存获取agent信息"""
        mock_skills.agents = [mock_agent_skill]
        mock_agent_core.cache_handler.get_skill_agent_info_dict.return_value = {
            "agent_id": "cached_agent"
        }

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.agent_factory_service"
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.get_biz_domain_id",
                        return_value="domain123",
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_user_account_id"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_user_account_type"
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.set_biz_domain_id"
                                ):
                                    with patch(
                                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills.hl_agent_options",
                                        new_callable=AsyncMock,
                                    ):
                                        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
                                            process_skills_agents,
                                        )

                                        await process_skills_agents(
                                            mock_agent_core,
                                            mock_agent_config,
                                            mock_skills,
                                            temp_files={},
                                            headers={"x-user-id": "user123"},
                                        )


class TestHlAgentOptions:
    """测试 hl_agent_options 函数"""

    @pytest.fixture
    def mock_agent_skill(self):
        """创建 mock AgentSkillVo"""
        skill = MagicMock()
        skill.inner_dto = MagicMock()
        skill.data_source_config = None
        skill.datasource_config = None
        skill.llm_config = None
        return skill

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.data_source = {
            "doc": ["doc1"],
            "kg": ["kg1"],
            "advanced_config": {},
        }
        config.llms = [{"is_default": True, "llm_config": {"model": "gpt-4"}}]
        return config

    @pytest.mark.asyncio
    async def test_hl_agent_options_with_temp_files(
        self, mock_agent_skill, mock_agent_config
    ):
        """测试带临时文件的agent options"""
        temp_files = {"test_file": {"filename": "test.txt"}}

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
            hl_agent_options,
        )

        await hl_agent_options(mock_agent_skill, mock_agent_config, temp_files)

        assert mock_agent_skill.inner_dto.agent_options is not None

    @pytest.mark.asyncio
    async def test_hl_agent_options_with_inherit_data_source(
        self, mock_agent_skill, mock_agent_config
    ):
        """测试继承数据源的agent options"""
        mock_data_source_config = MagicMock()
        mock_data_source_config.type = "inherit_main"
        mock_data_source_config.specific_inherit = "all"
        mock_agent_skill.data_source_config = mock_data_source_config

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
            hl_agent_options,
        )

        await hl_agent_options(mock_agent_skill, mock_agent_config, {})

        assert mock_agent_skill.inner_dto.agent_options is not None

    @pytest.mark.asyncio
    async def test_hl_agent_options_with_docs_only_inherit(
        self, mock_agent_skill, mock_agent_config
    ):
        """测试仅继承文档的agent options"""
        mock_data_source_config = MagicMock()
        mock_data_source_config.type = "inherit_main"
        mock_data_source_config.specific_inherit = "docs_only"
        mock_agent_skill.data_source_config = mock_data_source_config

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
            hl_agent_options,
        )

        await hl_agent_options(mock_agent_skill, mock_agent_config, {})

        assert "data_source" in mock_agent_skill.inner_dto.agent_options

    @pytest.mark.asyncio
    async def test_hl_agent_options_with_graph_only_inherit(
        self, mock_agent_skill, mock_agent_config
    ):
        """测试仅继承知识图谱的agent options"""
        mock_data_source_config = MagicMock()
        mock_data_source_config.type = "inherit_main"
        mock_data_source_config.specific_inherit = "graph_only"
        mock_agent_skill.data_source_config = mock_data_source_config

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
            hl_agent_options,
        )

        await hl_agent_options(mock_agent_skill, mock_agent_config, {})

        assert "data_source" in mock_agent_skill.inner_dto.agent_options

    @pytest.mark.asyncio
    async def test_hl_agent_options_with_inherit_llm(
        self, mock_agent_skill, mock_agent_config
    ):
        """测试继承LLM配置的agent options"""
        mock_llm_config = MagicMock()
        mock_llm_config.type = "inherit_main"
        mock_agent_skill.llm_config = mock_llm_config

        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.agent_skills import (
            hl_agent_options,
        )

        await hl_agent_options(mock_agent_skill, mock_agent_config, {})

        assert mock_agent_skill.inner_dto.agent_options is not None
