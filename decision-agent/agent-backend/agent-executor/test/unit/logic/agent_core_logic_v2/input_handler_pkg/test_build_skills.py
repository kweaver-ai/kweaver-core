# -*- coding: utf-8 -*-
"""单元测试 - input_handler_pkg/build_skills 模块"""

import importlib

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


def get_build_skills_module():
    return importlib.import_module(
        "app.logic.agent_core_logic_v2.input_handler_pkg.build_skills"
    )


class TestBuildSkills:
    """测试 build_skills 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.agent_config = MagicMock()
        core.agent_config.agent_id = "test_agent_id"
        core.agent_config.skills = MagicMock()
        core.agent_config.skills.tools = []
        core.agent_config.skills.agents = []
        core.agent_config.skills.mcps = []
        core.agent_config.memory = None
        core.agent_config.data_source = {}
        core.agent_config.llms = []
        return core

    @pytest.mark.asyncio
    async def test_build_skills_basic(self, mock_agent_core):
        """测试基本skills构建"""
        module = get_build_skills_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):
                with patch.object(
                    module,
                    "process_skills",
                    new_callable=AsyncMock,
                ):
                    build_skills = module.build_skills

                    await build_skills(
                        mock_agent_core,
                        headers={"x-user-id": "user123"},
                        llm_config={"default": "gpt-4", "llms": {}},
                        context_variables={"self_config": {}},
                        temp_files={},  # 传递空字典而不是None
                    )

    @pytest.mark.asyncio
    async def test_build_skills_with_temp_files(self, mock_agent_core):
        """测试带临时文件的skills构建"""
        temp_files = {"test_file": {"filename": "test.txt", "content": "content"}}

        module = get_build_skills_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):
                with patch.object(
                    module,
                    "process_skills",
                    new_callable=AsyncMock,
                ):
                    with patch.object(module, "ToolSkillVo"):
                        with patch.object(module, "SkillInputVo"):
                            with patch.object(module, "BuiltinIds") as mock_builtin_ids:
                                mock_builtin_ids.get_tool_id.return_value = "tool123"
                                mock_builtin_ids.get_tool_box_id.return_value = (
                                    "toolbox123"
                                )

                                build_skills = module.build_skills

                                await build_skills(
                                    mock_agent_core,
                                    headers={"x-user-id": "user123"},
                                    llm_config={
                                        "default": "gpt-4",
                                        "llms": {"gpt-4": {"model": "gpt-4"}},
                                    },
                                    context_variables={"self_config": {}},
                                    temp_files=temp_files,
                                )

    @pytest.mark.asyncio
    async def test_build_skills_with_memory_enabled(self, mock_agent_core):
        """测试启用内存的skills构建"""
        mock_agent_core.agent_config.memory = {"is_enabled": True}

        module = get_build_skills_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):
                with patch.object(
                    module,
                    "process_skills",
                    new_callable=AsyncMock,
                ):
                    with patch.object(module, "ToolSkillVo"):
                        with patch.object(module, "SkillInputVo"):
                            with patch.object(module, "BuiltinIds") as mock_builtin_ids:
                                mock_builtin_ids.get_tool_id.return_value = "tool123"
                                mock_builtin_ids.get_tool_box_id.return_value = (
                                    "toolbox123"
                                )

                                build_skills = module.build_skills

                                await build_skills(
                                    mock_agent_core,
                                    headers={"x-user-id": "user123"},
                                    llm_config={"default": "gpt-4", "llms": {}},
                                    context_variables={"self_config": {}},
                                    temp_files={},  # 传递空字典
                                )

                                # 验证内存工具被添加
                                assert (
                                    len(mock_agent_core.agent_config.skills.tools) > 0
                                )

    @pytest.mark.asyncio
    async def test_build_skills_with_none_skills(self, mock_agent_core):
        """测试skills为None的情况"""
        mock_agent_core.agent_config.skills = None

        module = get_build_skills_module()

        with patch.object(module, "span_set_attrs"):
            with patch.object(module, "get_user_account_id", return_value="user123"):
                with patch.object(
                    module,
                    "process_skills",
                    new_callable=AsyncMock,
                ):
                    with patch.object(module, "SkillVo") as mock_skill_vo:
                        mock_skill_instance = MagicMock()
                        mock_skill_instance.tools = []
                        mock_skill_instance.agents = []
                        mock_skill_instance.mcps = []
                        mock_skill_vo.return_value = mock_skill_instance

                        build_skills = module.build_skills

                        await build_skills(
                            mock_agent_core,
                            headers={"x-user-id": "user123"},
                            llm_config={"default": "gpt-4", "llms": {}},
                            context_variables={"self_config": {}},
                            temp_files={},  # 传递空字典
                        )

                        # 验证skills被设置为新的SkillVo
                        assert mock_agent_core.agent_config.skills is not None
