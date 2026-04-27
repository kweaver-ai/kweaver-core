# -*- coding: utf-8 -*-
"""单元测试 - run_dolphin 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch


class TestRunDolphin:
    """测试 run_dolphin 函数"""

    @pytest.fixture
    def mock_agent_core(self):
        """创建 mock AgentCoreV2"""
        core = MagicMock()
        core.is_warmup = False
        core.agent_config = MagicMock()
        core.agent_config.skills = MagicMock()
        core.agent_config.skills.tools = []
        core.agent_config.skills.agents = []
        core.agent_config.skills.mcps = []
        core.agent_run_id = "test_run_id"
        return core

    @pytest.fixture
    def mock_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.agent_run_id = "test_run_id"
        config.agent_id = "test_agent_id"
        return config

    @pytest.mark.asyncio
    async def test_run_dolphin_warmup_mode(self, mock_agent_core, mock_config):
        """测试预热模式"""
        mock_agent_core.is_warmup = True

        with patch("app.logic.agent_core_logic_v2.run_dolphin.span_set_attrs"):
            with patch(
                "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.run_dolphin.build_llm_config",
                        new_callable=AsyncMock,
                        return_value={},
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.run_dolphin.PromptBuilder"
                        ) as mock_builder:
                            mock_instance = MagicMock()
                            mock_instance.build = AsyncMock(return_value="test prompt")
                            mock_builder.return_value = mock_instance

                            with patch(
                                "app.logic.agent_core_logic_v2.run_dolphin.build_skills",
                                new_callable=AsyncMock,
                            ):
                                from app.logic.agent_core_logic_v2.run_dolphin import (
                                    run_dolphin,
                                )

                                results = []
                                async for res in run_dolphin(
                                    mock_agent_core,
                                    mock_config,
                                    {},
                                    {"x-user-id": "user123"},
                                ):
                                    results.append(res)

                                # 预热模式不返回任何结果
                                assert len(results) == 0

    @pytest.mark.asyncio
    async def test_run_dolphin_basic(self, mock_agent_core, mock_config):
        """测试基本运行"""
        mock_agent = MagicMock()
        mock_agent.initialize = AsyncMock()
        mock_agent.executor = MagicMock()

        with patch("app.logic.agent_core_logic_v2.run_dolphin.span_set_attrs"):
            with patch(
                "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.run_dolphin.build_llm_config",
                        new_callable=AsyncMock,
                        return_value={"default": "gpt-4"},
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.run_dolphin.PromptBuilder"
                        ) as mock_builder:
                            mock_instance = MagicMock()
                            mock_instance.build = AsyncMock(return_value="test prompt")
                            mock_builder.return_value = mock_instance

                            with patch(
                                "app.logic.agent_core_logic_v2.run_dolphin.build_skills",
                                new_callable=AsyncMock,
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.run_dolphin.build_tools",
                                    new_callable=AsyncMock,
                                    return_value={},
                                ):
                                    with patch(
                                        "app.logic.agent_core_logic_v2.run_dolphin.TriditionalToolkit"
                                    ) as mock_toolkit:
                                        mock_toolkit.buildFromTooldict.return_value = (
                                            MagicMock(tools=[])
                                        )

                                        with patch(
                                            "app.logic.agent_core_logic_v2.run_dolphin.GlobalConfig"
                                        ) as mock_global_config:
                                            mock_global_config.from_dict.return_value = MagicMock(
                                                skill_config=MagicMock(
                                                    enabled_skills=[]
                                                )
                                            )

                                            with patch(
                                                "app.logic.agent_core_logic_v2.run_dolphin.DolphinAgent",
                                                return_value=mock_agent,
                                            ):
                                                with patch(
                                                    "app.logic.agent_core_logic_v2.run_dolphin.Config"
                                                ) as mock_cfg:
                                                    mock_cfg.features.enable_dolphin_agent_output_variables_ctrl = False
                                                    mock_cfg.is_dolphin_trace_enabled.return_value = False
                                                    mock_cfg.is_o11y_trace_enabled.return_value = False
                                                    mock_cfg.app.enable_dolphin_agent_verbose = False
                                                    mock_cfg.app.get_stdlib_log_level.return_value = "INFO"

                                                    with patch(
                                                        "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
                                                    ) as mock_process:

                                                        async def mock_gen():
                                                            yield {"status": "success"}

                                                        mock_process.return_value = (
                                                            mock_gen()
                                                        )

                                                        with patch(
                                                            "app.logic.agent_core_logic_v2.run_dolphin.DialogLogHandler"
                                                        ):
                                                            with patch(
                                                                "app.logic.agent_core_logic_v2.run_dolphin.agent_instance_manager"
                                                            ):
                                                                with patch(
                                                                    "app.logic.agent_core_logic_v2.run_dolphin.StandLogger.info"
                                                                ) as mock_info:
                                                                    with patch(
                                                                        "app.logic.agent_core_logic_v2.run_dolphin.StandLogger.info_log"
                                                                    ) as mock_info_log:
                                                                        from app.logic.agent_core_logic_v2.run_dolphin import (
                                                                            run_dolphin,
                                                                        )

                                                                        results = []
                                                                        async for (
                                                                            res
                                                                        ) in run_dolphin(
                                                                            mock_agent_core,
                                                                            mock_config,
                                                                            {},
                                                                            {
                                                                                "x-user-id": "user123"
                                                                            },
                                                                        ):
                                                                            results.append(
                                                                                res
                                                                            )

                                                                        assert len(results) > 0
                                                                        mock_info.assert_not_called()
                                                                        mock_info_log.assert_called_once()
                                                                        assert (
                                                                            "Agent execution details"
                                                                            in mock_info_log.call_args.args[
                                                                                0
                                                                            ]
                                                                        )

    @pytest.mark.asyncio
    async def test_run_dolphin_with_output_variables(
        self, mock_agent_core, mock_config
    ):
        """测试启用输出变量"""
        mock_agent = MagicMock()
        mock_agent.initialize = AsyncMock()
        mock_agent.executor = MagicMock()

        with patch("app.logic.agent_core_logic_v2.run_dolphin.span_set_attrs"):
            with patch(
                "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_id",
                return_value="user123",
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.run_dolphin.get_user_account_type",
                    return_value="standard",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.run_dolphin.build_llm_config",
                        new_callable=AsyncMock,
                        return_value={},
                    ):
                        with patch(
                            "app.logic.agent_core_logic_v2.run_dolphin.PromptBuilder"
                        ) as mock_builder:
                            mock_instance = MagicMock()
                            mock_instance.build = AsyncMock(return_value="test prompt")
                            mock_builder.return_value = mock_instance

                            with patch(
                                "app.logic.agent_core_logic_v2.run_dolphin.build_skills",
                                new_callable=AsyncMock,
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.run_dolphin.build_tools",
                                    new_callable=AsyncMock,
                                    return_value={},
                                ):
                                    with patch(
                                        "app.logic.agent_core_logic_v2.run_dolphin.TriditionalToolkit"
                                    ) as mock_toolkit:
                                        mock_toolkit.buildFromTooldict.return_value = (
                                            MagicMock(tools=[])
                                        )

                                        with patch(
                                            "app.logic.agent_core_logic_v2.run_dolphin.GlobalConfig"
                                        ) as mock_global_config:
                                            mock_global_config.from_dict.return_value = MagicMock(
                                                skill_config=MagicMock(
                                                    enabled_skills=[]
                                                )
                                            )

                                            with patch(
                                                "app.logic.agent_core_logic_v2.run_dolphin.DolphinAgent",
                                                return_value=mock_agent,
                                            ):
                                                with patch(
                                                    "app.logic.agent_core_logic_v2.run_dolphin.Config"
                                                ) as mock_cfg:
                                                    mock_cfg.features.enable_dolphin_agent_output_variables_ctrl = True
                                                    mock_cfg.local_dev.dolphin_agent_output_variables = None
                                                    mock_cfg.app.enable_dolphin_agent_verbose = False
                                                    mock_cfg.app.get_stdlib_log_level.return_value = "INFO"

                                                    with patch(
                                                        "app.logic.agent_core_logic_v2.run_dolphin.get_output_variables",
                                                        return_value=["var1", "var2"],
                                                    ):
                                                        with patch(
                                                            "app.logic.agent_core_logic_v2.interrupt_utils.process_arun_loop"
                                                        ) as mock_process:

                                                            async def mock_gen():
                                                                yield {
                                                                    "status": "success"
                                                                }

                                                            mock_process.return_value = mock_gen()

                                                            with patch(
                                                                "app.logic.agent_core_logic_v2.run_dolphin.DialogLogHandler"
                                                            ):
                                                                with patch(
                                                                    "app.logic.agent_core_logic_v2.run_dolphin.agent_instance_manager"
                                                                ):
                                                                    from app.logic.agent_core_logic_v2.run_dolphin import (
                                                                        run_dolphin,
                                                                    )

                                                                    results = []
                                                                    async for (
                                                                        res
                                                                    ) in run_dolphin(
                                                                        mock_agent_core,
                                                                        mock_config,
                                                                        {},
                                                                        {
                                                                            "x-user-id": "user123"
                                                                        },
                                                                    ):
                                                                        results.append(
                                                                            res
                                                                        )

                                                                    assert (
                                                                        len(results) > 0
                                                                    )
