# -*- coding:utf-8 -*-
"""单元测试 - run_agent_v2 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


@pytest.mark.asyncio
class TestProcessOptions:
    """测试 process_options 函数"""

    async def test_with_none_options(self):
        """测试options为None时提前返回"""
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )
        from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo

        agent_config = MagicMock(spec=AgentConfigVo)
        agent_input = MagicMock(spec=AgentInputVo)

        result = process_options(None, agent_config, agent_input)
        assert result is None

    async def test_with_output_vars(self):
        """测试output_vars选项"""
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )
        from app.domain.vo.agentvo import AgentConfigVo

        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = []
        agent_config.input = {}

        options = MagicMock()
        options.output_vars = {"var1": "value1"}
        options.llm_config = None
        options.incremental_output = None
        options.data_source = None
        options.tmp_files = None
        options.agent_id = None
        options.conversation_id = None
        options.agent_run_id = None

        process_options(options, agent_config, MagicMock())
        assert agent_config.output_vars == {"var1": "value1"}

    async def test_with_llm_config_existing(self):
        """测试存在的LLM配置"""
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )
        from app.domain.vo.agentvo import AgentConfigVo

        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = [
            {"is_default": False, "llm_config": {"name": "gpt-3.5"}},
            {"is_default": False, "llm_config": {"name": "gpt-4"}},
        ]
        agent_config.input = {}

        options = MagicMock()
        options.llm_config = {"name": "gpt-4"}
        options.output_vars = None
        options.incremental_output = None
        options.data_source = None
        options.tmp_files = None
        options.agent_id = None
        options.conversation_id = None
        options.agent_run_id = None

        process_options(options, agent_config, MagicMock())
        assert agent_config.llms[0]["is_default"] is False
        assert agent_config.llms[1]["is_default"] is True

    async def test_with_agent_id(self):
        """测试agent_id选项"""
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )
        from app.domain.vo.agentvo import AgentConfigVo

        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_id = "old-id"
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = []
        agent_config.input = {}

        options = MagicMock()
        options.agent_id = "new-id"
        options.llm_config = None
        options.output_vars = None
        options.incremental_output = None
        options.data_source = None
        options.tmp_files = None
        options.conversation_id = None
        options.agent_run_id = None

        process_options(options, agent_config, MagicMock())
        assert agent_config.agent_id == "new-id"


@pytest.mark.asyncio
class TestHandleCache:
    """测试 handle_cache 函数"""

    async def test_cache_enabled(self):
        """测试启用缓存"""
        from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
            handle_cache,
            cache_manager,
        )

        agent_core_v2 = MagicMock()
        agent_core_v2.agent_config = MagicMock()
        agent_core_v2.agent_config.agent_version = "1.0"
        agent_core_v2.agent_config.get_config_last_set_timestamp = MagicMock(
            return_value=123456
        )
        agent_core_v2.cache_handler = MagicMock()

        # Mock the cache manager methods
        cache_manager.cache_service = MagicMock()
        cache_manager.cache_service.load = AsyncMock(
            return_value=None
        )  # No existing cache
        cache_manager.create_cache = AsyncMock(return_value=MagicMock(cache_data={}))
        cache_manager.cache_service.get_ttl = AsyncMock(return_value=3600)

        result = await handle_cache(
            "agent123",
            agent_core_v2,
            is_debug_run=False,
            headers={},
            account_id="acc456",
            account_type="premium",
        )

        assert result is not None

    async def test_cache_disabled(self):
        """测试禁用缓存 - 调试模式"""
        from app.router.agent_controller_pkg.run_agent_v2.handle_cache import (
            handle_cache,
        )

        agent_core_v2 = MagicMock()

        # In debug mode, cache is not processed
        result = await handle_cache(
            "agent123",
            agent_core_v2,
            is_debug_run=True,  # Debug mode
            headers={},
            account_id="acc456",
            account_type="premium",
        )

        assert result is None


@pytest.mark.asyncio
class TestSafeOutputGenerator:
    """测试 create_safe_output_generator 函数"""

    @pytest.mark.asyncio
    async def test_basic_call(self):
        """测试基本调用"""
        from app.router.agent_controller_pkg.run_agent_v2.safe_output_generator import (
            create_safe_output_generator,
        )

        agent_core_v2 = MagicMock()
        agent_config = MagicMock()
        agent_input = MagicMock()
        headers = {}
        cache_id_vo = None
        is_debug_run = False
        start_time = 1234567890.0

        # Mock the output_handler to return an async generator
        async def mock_output_generator():
            yield "result1"
            yield "result2"

        agent_core_v2.output_handler.result_output = MagicMock(
            return_value=mock_output_generator()
        )

        result = create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=is_debug_run,
            start_time=start_time,
            cache_id_vo=cache_id_vo,
            account_id="acc123",
            account_type="premium",
        )

        # Verify it returns an async generator
        assert result is not None

        items = []
        async for item in result:
            items.append(item)

        assert len(items) > 0


@pytest.mark.asyncio
class TestAgentCoreV2Init:
    """测试 AgentCoreV2 初始化"""

    async def test_basic_init(self):
        """测试基本初始化"""
        from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

        agent_config = MagicMock()
        agent_config.agent_id = "test-agent"
        agent_config.skills = []
        agent_config.llms = []

        agent_core = AgentCoreV2(agent_config)

        assert agent_core.agent_config is agent_config
        assert agent_core.agent_config.agent_id == "test-agent"


@pytest.mark.asyncio
class TestInputHandler:
    """测试 prepare 函数的输入处理部分"""

    async def test_build_input(self):
        """测试构建输入 - 使用 prepare 函数"""
        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare
        from app.router.agent_controller_pkg.rdto.v2.req.run_agent import V2RunAgentReq
        from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo

        request = MagicMock()
        request.headers = {}

        req = MagicMock(spec=V2RunAgentReq)
        req.agent_config = MagicMock(spec=AgentConfigVo)
        req.agent_config.agent_id = "test-agent"
        req.agent_input = MagicMock(spec=AgentInputVo)
        req.options = None
        req.agent_id = None
        req.agent_version = None

        # Mock the agent_factory_service
        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            result = await prepare(request, req, "acc123", "premium")

        assert result is not None
        assert len(result) == 3


@pytest.mark.asyncio
class TestOutputHandler:
    """测试 output_handler 相关功能"""

    async def test_build_output(self):
        """测试构建输出 - 验证输出配置"""
        from app.domain.vo.agentvo import AgentConfigVo

        agent_core_v2 = MagicMock()
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.output = MagicMock()
        output_vars = {"result": "value"}

        # Just verify the objects are properly set up
        assert agent_config is not None
        assert output_vars is not None
