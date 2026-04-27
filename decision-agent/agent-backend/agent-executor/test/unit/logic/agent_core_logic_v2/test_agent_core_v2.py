# -*- coding: utf-8 -*-
"""单元测试 - agent_core_v2 模块"""

import pytest
import sys
import types
from types import SimpleNamespace
from unittest.mock import MagicMock, AsyncMock, patch


def install_resume_info_stub(monkeypatch):
    """为 resume_agent 模块安装轻量 stub，避免拉起完整 router 初始化。"""
    package_names = [
        "app.router",
        "app.router.agent_controller_pkg",
        "app.router.agent_controller_pkg.rdto",
        "app.router.agent_controller_pkg.rdto.v2",
        "app.router.agent_controller_pkg.rdto.v2.req",
    ]

    for package_name in package_names:
        module = types.ModuleType(package_name)
        module.__path__ = []
        monkeypatch.setitem(sys.modules, package_name, module)

    resume_agent_module = types.ModuleType(
        "app.router.agent_controller_pkg.rdto.v2.req.resume_agent"
    )

    class ResumeInfo:
        def __init__(self, **kwargs):
            for key, value in kwargs.items():
                setattr(self, key, value)

    resume_agent_module.ResumeInfo = ResumeInfo
    monkeypatch.setitem(sys.modules, resume_agent_module.__name__, resume_agent_module)
    return ResumeInfo


class TestAgentCoreV2:
    """测试 AgentCoreV2 类"""

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.agent_run_id = "test_run_id"
        config.agent_id = "test_agent_id"
        config.output_vars = []
        config.react_disable_llm_cache.return_value = False
        return config

    @pytest.fixture
    def mock_agent_input(self):
        """创建 mock AgentInputVo"""
        return MagicMock()

    def test_init(self, mock_agent_config):
        """测试初始化"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config, is_warmup=True)

        assert core.agent_config == mock_agent_config
        assert core.is_warmup is True
        assert core.tool_dict == {}
        assert core.temp_files == {}
        assert core.memory_handler is not None
        assert core.output_handler is not None
        assert core.cache_handler is not None
        assert core.warmup_handler is not None

    def test_cleanup(self, mock_agent_config):
        """测试清理方法"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)
        core.tool_dict = {"tool1": MagicMock(), "tool2": MagicMock()}
        core.executor = MagicMock()

        core.cleanup()

        assert core.executor is None
        assert core.tool_dict == {}

    def test_remove_context_from_response_with_context(self):
        """测试移除响应中的context"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        response = {"context": {"key": "value"}, "data": "test"}
        result = AgentCoreV2.remove_context_from_response(response)

        assert "context" not in result
        assert "data" in result
        assert result["data"] == "test"

    def test_remove_context_from_response_without_context(self):
        """测试响应中没有context"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        response = {"data": "test"}
        result = AgentCoreV2.remove_context_from_response(response)

        assert result == response

    def test_set_run_options(self, mock_agent_config):
        """测试设置运行选项"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)
        mock_options = MagicMock()

        core.set_run_options(mock_options)

        assert core.run_options_vo == mock_options

    def test_get_resume_info_from_options_none(self, mock_agent_config):
        """测试从options获取resume_info为None"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)

        result = core._get_resume_info_from_options()

        assert result is None

    def test_get_resume_info_from_options_with_resume_info(
        self, mock_agent_config, monkeypatch
    ):
        """测试从options获取resume_info"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)
        ResumeInfo = install_resume_info_stub(monkeypatch)
        mock_resume_info = ResumeInfo(
            resume_handle={"frame_id": "frame123"},
            action="confirm",
            data={"key": "value"},
        )
        core.run_options_vo.resume_info = mock_resume_info

        result = core._get_resume_info_from_options()

        # 如果已经是ResumeInfo类型，直接返回
        assert result == mock_resume_info

    def test_get_resume_info_from_options_with_dict(
        self, mock_agent_config, monkeypatch
    ):
        """测试从options获取resume_info字典"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)
        install_resume_info_stub(monkeypatch)
        resume_data = {
            "resume_handle": {
                "frame_id": "frame123",
                "snapshot_id": "snap123",
                "resume_token": "token123",
                "interrupt_type": "confirm",
                "current_block": 0,
                "restart_block": False,
            },
            "action": "confirm",
            "data": {"key": "value"},
        }
        core.run_options_vo.resume_info = resume_data

        result = core._get_resume_info_from_options()

        # dict 类型应该被转换为 ResumeInfo 对象
        assert result is not None
        assert result.action == "confirm"

    def test_get_registered_agent_instance_success(self, mock_agent_config):
        """测试获取已注册的agent实例成功"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)
        mock_agent = MagicMock()

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get.return_value = (mock_agent, {})

            result = core._get_registered_agent_instance("test_run_id")

            assert result == mock_agent
            mock_manager.get.assert_called_once_with("test_run_id")

    def test_get_registered_agent_instance_not_found(self, mock_agent_config):
        """测试获取已注册的agent实例不存在"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        core = AgentCoreV2(agent_config=mock_agent_config)

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get.return_value = None

            with pytest.raises(ValueError) as exc_info:
                core._get_registered_agent_instance("test_run_id")

            assert "Agent instance not found" in str(exc_info.value)


class TestAgentCoreV2Run:
    """测试 AgentCoreV2.run 方法"""

    @pytest.fixture
    def mock_agent_config(self):
        """创建 mock AgentConfigVo"""
        config = MagicMock()
        config.agent_run_id = "test_run_id"
        config.agent_id = "test_agent_id"
        config.output_vars = []
        config.react_disable_llm_cache.return_value = False
        return config

    @pytest.fixture
    def mock_agent_input(self):
        """创建 mock AgentInputVo"""
        return MagicMock()

    @pytest.mark.asyncio
    async def test_run_basic(self, mock_agent_config, mock_agent_input):
        """测试基本运行"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        headers = {"x-user-id": "user123"}

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.process_input",
            new_callable=AsyncMock,
            return_value={},
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.process_tool_input",
                new_callable=AsyncMock,
                return_value=({}, None),
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.get_user_account_id",
                    return_value="user123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.agent_core_v2.run_dolphin"
                    ) as mock_run_dolphin:

                        async def mock_gen():
                            yield {"status": "success", "data": "test"}

                        mock_run_dolphin.return_value = mock_gen()

                        with patch(
                            "app.logic.agent_core_logic_v2.agent_core_v2.Config"
                        ) as mock_config:
                            mock_config.features.use_explore_block_v2 = False
                            mock_config.features.disable_dolphin_sdk_llm_cache = False

                            with patch(
                                "app.logic.agent_core_logic_v2.agent_core_v2.flags"
                            ):
                                with patch(
                                    "app.logic.agent_core_logic_v2.agent_core_v2.StandLogger.info"
                                ) as mock_info:
                                    core = AgentCoreV2()
                                    results = []
                                    async for res in core.run(
                                        mock_agent_config,
                                        mock_agent_input,
                                        headers,
                                    ):
                                        results.append(res)

                                    assert len(results) == 1
                                    assert results[0]["status"] == "success"
                                    mock_info.assert_called_once_with(
                                        "AgentCore run end"
                                    )

    @pytest.mark.asyncio
    async def test_run_with_tool_interrupt(self, mock_agent_config, mock_agent_input):
        """测试工具中断异常处理"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2
        from app.common.exceptions.tool_interrupt import ToolInterruptException

        headers = {"x-user-id": "user123"}

        # 创建正确的 ToolInterruptException mock
        mock_interrupt_info = MagicMock()
        mock_interrupt_info.data = {"tool_name": "test_tool"}
        mock_interrupt_exception = ToolInterruptException(mock_interrupt_info)

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.process_input",
            new_callable=AsyncMock,
            return_value={},
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.process_tool_input",
                new_callable=AsyncMock,
                return_value=({}, None),
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.get_user_account_id",
                    return_value="user123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.agent_core_v2.run_dolphin"
                    ) as mock_run_dolphin:

                        async def mock_gen():
                            raise mock_interrupt_exception
                            yield  # noqa: F821

                        mock_run_dolphin.return_value = mock_gen()

                        with patch(
                            "app.logic.agent_core_logic_v2.agent_core_v2.InterruptHandler"
                        ) as mock_handler:
                            mock_handler.handle_tool_interrupt = AsyncMock()

                            with patch(
                                "app.logic.agent_core_logic_v2.agent_core_v2.Config"
                            ) as mock_config:
                                mock_config.features.use_explore_block_v2 = False
                                mock_config.features.disable_dolphin_sdk_llm_cache = False

                                with patch(
                                    "app.logic.agent_core_logic_v2.agent_core_v2.flags"
                                ):
                                    core = AgentCoreV2()
                                    results = []
                                    async for res in core.run(
                                        mock_agent_config,
                                        mock_agent_input,
                                        headers,
                                    ):
                                        results.append(res)

                                    assert len(results) == 1
                                    mock_handler.handle_tool_interrupt.assert_called_once()

    @pytest.mark.asyncio
    async def test_run_with_exception(self, mock_agent_config, mock_agent_input):
        """测试异常处理"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        headers = {"x-user-id": "user123"}

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.process_input",
            new_callable=AsyncMock,
            return_value={},
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.process_tool_input",
                new_callable=AsyncMock,
                return_value=({}, None),
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.get_user_account_id",
                    return_value="user123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.agent_core_v2.run_dolphin"
                    ) as mock_run_dolphin:

                        async def mock_gen():
                            raise Exception("Test error")
                            yield  # noqa: F821

                        mock_run_dolphin.return_value = mock_gen()

                        with patch(
                            "app.logic.agent_core_logic_v2.agent_core_v2.ExceptionHandler"
                        ) as mock_handler:
                            mock_handler.handle_exception = AsyncMock()

                            with patch(
                                "app.logic.agent_core_logic_v2.agent_core_v2.Config"
                            ) as mock_config:
                                mock_config.features.use_explore_block_v2 = False
                                mock_config.features.disable_dolphin_sdk_llm_cache = False

                                with patch(
                                    "app.logic.agent_core_logic_v2.agent_core_v2.flags"
                                ):
                                    with patch(
                                        "app.logic.agent_core_logic_v2.agent_core_v2.o11y_logger"
                                    ):
                                        core = AgentCoreV2()
                                        results = []
                                        async for res in core.run(
                                            mock_agent_config,
                                            mock_agent_input,
                                            headers,
                                        ):
                                            results.append(res)

                                        assert len(results) == 1
                                        mock_handler.handle_exception.assert_called_once()

    @pytest.mark.asyncio
    async def test_run_skips_llm_message_logging_flags_when_missing(
        self, mock_agent_config, mock_agent_input
    ):
        """测试当前 Dolphin 版本缺少 LLM_MESSAGE_LOGGING 常量时不会抛异常。"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        headers = {"x-user-id": "user123"}

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.process_input",
            new_callable=AsyncMock,
            return_value={},
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.process_tool_input",
                new_callable=AsyncMock,
                return_value=({}, None),
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.get_user_account_id",
                    return_value="user123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.agent_core_v2.run_dolphin"
                    ) as mock_run_dolphin:

                        async def mock_gen():
                            yield {"status": "success", "data": "test"}

                        mock_run_dolphin.return_value = mock_gen()

                        with patch(
                            "app.logic.agent_core_logic_v2.agent_core_v2.Config"
                        ) as mock_config:
                            mock_config.features.use_explore_block_v2 = False
                            mock_config.features.disable_dolphin_sdk_llm_cache = False
                            mock_config.llm_message_logging.enabled = True
                            mock_config.llm_message_logging.log_dir = "/tmp/llm-log"

                            mock_flags = SimpleNamespace(
                                EXPLORE_BLOCK_V2="EXPLORE_BLOCK_V2",
                                DISABLE_LLM_CACHE="DISABLE_LLM_CACHE",
                                set_flag=MagicMock(),
                                set_param=MagicMock(),
                            )

                            with patch(
                                "app.logic.agent_core_logic_v2.agent_core_v2.flags",
                                mock_flags,
                            ):
                                core = AgentCoreV2()
                                results = []
                                async for res in core.run(
                                    mock_agent_config,
                                    mock_agent_input,
                                    headers,
                                ):
                                    results.append(res)

                                assert len(results) == 1
                                mock_run_dolphin.assert_called_once()

    @pytest.mark.asyncio
    async def test_run_disables_llm_cache_when_agent_config_requests_it(
        self, mock_agent_config, mock_agent_input
    ):
        """测试 react 模式配置要求禁用缓存时，会下发禁用缓存 flag。"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        headers = {"x-user-id": "user123"}
        mock_agent_config.react_disable_llm_cache.return_value = True

        with patch(
            "app.logic.agent_core_logic_v2.agent_core_v2.process_input",
            new_callable=AsyncMock,
            return_value={},
        ):
            with patch(
                "app.logic.agent_core_logic_v2.agent_core_v2.process_tool_input",
                new_callable=AsyncMock,
                return_value=({}, None),
            ):
                with patch(
                    "app.logic.agent_core_logic_v2.agent_core_v2.get_user_account_id",
                    return_value="user123",
                ):
                    with patch(
                        "app.logic.agent_core_logic_v2.agent_core_v2.run_dolphin"
                    ) as mock_run_dolphin:

                        async def mock_gen():
                            yield {"status": "success", "data": "test"}

                        mock_run_dolphin.return_value = mock_gen()

                        with patch(
                            "app.logic.agent_core_logic_v2.agent_core_v2.Config"
                        ) as mock_config:
                            mock_config.features.use_explore_block_v2 = False
                            mock_config.features.disable_dolphin_sdk_llm_cache = False
                            mock_config.llm_message_logging.enabled = False

                            mock_flags = SimpleNamespace(
                                EXPLORE_BLOCK_V2="EXPLORE_BLOCK_V2",
                                DISABLE_LLM_CACHE="DISABLE_LLM_CACHE",
                                set_flag=MagicMock(),
                                set_param=MagicMock(),
                            )

                            with patch(
                                "app.logic.agent_core_logic_v2.agent_core_v2.flags",
                                mock_flags,
                            ):
                                core = AgentCoreV2()
                                results = []
                                async for res in core.run(
                                    mock_agent_config,
                                    mock_agent_input,
                                    headers,
                                ):
                                    results.append(res)

                                assert len(results) == 1
                                mock_flags.set_flag.assert_any_call(
                                    "DISABLE_LLM_CACHE", True
                                )
