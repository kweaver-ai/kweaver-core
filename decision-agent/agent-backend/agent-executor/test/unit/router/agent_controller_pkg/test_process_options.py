"""测试 - process_options 处理选项"""

import pytest
from unittest.mock import MagicMock
from app.router.agent_controller_pkg.run_agent_v2.process_options import process_options
from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo


@pytest.mark.asyncio
class TestProcessOptions:
    """测试 process_options 函数"""

    async def test_none_options(self):
        """测试options为None时提前返回"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_input = MagicMock(spec=AgentInputVo)

        result = process_options(None, agent_config, agent_input)
        assert result is None

    async def test_with_output_vars(self):
        """测试output_vars选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(output_vars=["var1", "var2"])
        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, MagicMock(), span)

        assert agent_config.output_vars == ["var1", "var2"]

    async def test_with_llm_config(self):
        """测试LLM配置选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(llm_config={"name": "gpt-4"})
        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, MagicMock(), span)

        assert len(agent_config.llms) == 1
        assert agent_config.llms[0]["is_default"] is True

    async def test_with_agent_id(self):
        """测试agent_id选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "old-id"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo

        options = AgentRunOptionsVo(agent_id="new-id")
        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, MagicMock(), span)

        assert agent_config.agent_id == "new-id"
