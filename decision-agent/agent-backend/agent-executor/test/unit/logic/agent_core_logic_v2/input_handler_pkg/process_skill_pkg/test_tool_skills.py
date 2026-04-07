# -*- coding: utf-8 -*-
"""单元测试 - process_skill_pkg/tool_skills 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestProcessSkillsTools:
    """测试 process_skills_tools 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.is_warmup = True
        core.cache_handler = MagicMock()
        core.cache_handler.get_tools_info_dict.return_value = None
        return core

    @pytest.fixture
    def mock_skills(self):
        """创建 mock SkillVo"""
        skills = MagicMock()
        skills.tools = []
        return skills

    @pytest.mark.asyncio
    async def test_process_skills_tools_empty(self, mock_agent_core, mock_skills):
        """测试空tools列表"""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
            process_skills_tools,
        )

        await process_skills_tools(
            mock_agent_core,
            context_variables={"self_config": {}},
            headers={"x-user-id": "user123"},
            skills=mock_skills,
        )

    @pytest.mark.asyncio
    async def test_process_skills_tools_with_tool(self, mock_agent_core, mock_skills):
        """测试处理tool"""

        # 创建一个简单的工具对象
        class MockTool:
            def __init__(self):
                self.tool_id = "tool123"
                self.tool_box_id = "toolbox123"

        mock_tool = MockTool()
        mock_skills.tools = [mock_tool]

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.agent_operator_integration_service"
        ) as mock_service:
            mock_service.get_tool_info = AsyncMock(
                return_value={"name": "test_tool", "use_rule": "test_rule"}
            )

            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_biz_domain_id",
                        return_value="domain123",
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_id"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_type"
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_biz_domain_id"
                                ):
                                    from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
                                        process_skills_tools,
                                    )

                                    await process_skills_tools(
                                        mock_agent_core,
                                        context_variables={"self_config": {}},
                                        headers={"x-user-id": "user123"},
                                        skills=mock_skills,
                                    )

    @pytest.mark.asyncio
    async def test_process_skills_tools_with_cached_tool(
        self, mock_agent_core, mock_skills
    ):
        """测试从缓存获取tool信息"""

        class MockTool:
            def __init__(self):
                self.tool_id = "tool123"
                self.tool_box_id = "toolbox123"

        mock_tool = MockTool()
        mock_skills.tools = [mock_tool]
        mock_agent_core.cache_handler.get_tools_info_dict.return_value = {
            "name": "cached_tool",
            "use_rule": "cached_rule",
        }

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_id",
            return_value="user123",
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_type",
                return_value="standard",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_biz_domain_id",
                    return_value="domain123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_id"
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_type"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_biz_domain_id"
                            ):
                                from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
                                    process_skills_tools,
                                )

                                await process_skills_tools(
                                    mock_agent_core,
                                    context_variables={"self_config": {}},
                                    headers={"x-user-id": "user123"},
                                    skills=mock_skills,
                                )

    @pytest.mark.asyncio
    async def test_process_skills_tools_with_tool_rules(
        self, mock_agent_core, mock_skills
    ):
        """测试设置tool_rules"""

        class MockTool:
            def __init__(self):
                self.tool_id = "tool123"
                self.tool_box_id = "toolbox123"

        mock_tool = MockTool()
        mock_skills.tools = [mock_tool]
        mock_agent_core.cache_handler.get_tools_info_dict.return_value = {
            "name": "test_tool",
            "use_rule": {"rule": "value"},
        }

        context_variables = {"self_config": {}}

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_id",
            return_value="user123",
        ):
            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_type",
                return_value="standard",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_biz_domain_id",
                    return_value="domain123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_id"
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_type"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_biz_domain_id"
                            ):
                                from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
                                    process_skills_tools,
                                )

                                await process_skills_tools(
                                    mock_agent_core,
                                    context_variables=context_variables,
                                    headers={"x-user-id": "user123"},
                                    skills=mock_skills,
                                )

                                # 验证tool_rules被设置
                                assert "tool_rules" in context_variables["self_config"]

    @pytest.mark.asyncio
    async def test_process_skills_tools_none_skills(self, mock_agent_core):
        """测试skills为None的情况"""
        from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
            process_skills_tools,
        )

        await process_skills_tools(
            mock_agent_core,
            context_variables={"self_config": {}},
            headers={"x-user-id": "user123"},
            skills=None,
        )

    @pytest.mark.asyncio
    async def test_process_skills_tools_remove_unavailable_tool(
        self, mock_agent_core, mock_skills
    ):
        """测试工具不可用时移除问题工具"""

        class MockTool:
            def __init__(self, tool_id):
                self.tool_id = tool_id
                self.tool_box_id = "toolbox123"

        unavailable_tool = MockTool("tool-unavailable")
        available_tool = MockTool("tool-available")
        mock_skills.tools = [unavailable_tool, available_tool]

        with patch(
            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.agent_operator_integration_service"
        ) as mock_service:
            mock_service.get_tool_info = AsyncMock(
                side_effect=[
                    None,
                    {"name": "available_tool", "use_rule": "available_rule"},
                ]
            )

            with patch(
                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.get_biz_domain_id",
                        return_value="domain123",
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_id"
                        ):
                            with patch(
                                "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_user_account_type"
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills.set_biz_domain_id"
                                ):
                                    from app.logic.agent_core_logic_v2.input_handler_pkg.process_skill_pkg.tool_skills import (
                                        process_skills_tools,
                                    )

                                    await process_skills_tools(
                                        mock_agent_core,
                                        context_variables={"self_config": {}},
                                        headers={"x-user-id": "user123"},
                                        skills=mock_skills,
                                    )

        assert len(mock_skills.tools) == 1
        assert mock_skills.tools[0].tool_id == "tool-available"
