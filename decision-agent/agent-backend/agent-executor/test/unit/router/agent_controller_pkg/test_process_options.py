"""测试 - process_options 处理选项"""

import pytest
from unittest.mock import MagicMock, patch
from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo


@pytest.mark.asyncio
class TestProcessOptions:
    """测试 process_options 函数"""

    async def test_none_options(self):
        """测试options为None时提前返回"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_input = MagicMock(spec=AgentInputVo)
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        result = process_options(None, agent_config, agent_input)
        assert result is None

    async def test_with_output_vars(self):
        """测试output_vars选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        options = AgentRunOptionsVo(output_vars=["var1", "var2"])
        process_options(options, agent_config, MagicMock())

        assert agent_config.output_vars == ["var1", "var2"]

    async def test_with_llm_config(self):
        """测试LLM配置选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        options = AgentRunOptionsVo(llm_config={"name": "gpt-4"})
        process_options(options, agent_config, MagicMock())

        assert len(agent_config.llms) == 1
        assert agent_config.llms[0]["is_default"] is True

    async def test_with_agent_id(self):
        """测试agent_id选项"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "old-id"
        agent_config.conversation_id = "conversation-123"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        options = AgentRunOptionsVo(agent_id="new-id")
        process_options(options, agent_config, MagicMock())

        assert agent_config.agent_id == "new-id"

    async def test_with_conversation_id_logs_to_standard_logger_only(self):
        """普通会话设置日志应只写标准日志。"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.conversation_id = "old-conversation"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        options = AgentRunOptionsVo(conversation_id="new-conversation")

        with patch("app.common.stand_log.StandLogger.info") as mock_standard_info:
            process_options(options, agent_config, MagicMock())

        assert agent_config.conversation_id == "new-conversation"
        mock_standard_info.assert_called_once()

    async def test_without_conversation_id_logs_warning_to_standard_logger_only(self):
        """缺失会话ID属于普通告警，不应进入 O11Y。"""
        agent_config = MagicMock(spec=AgentConfigVo)
        agent_config.agent_run_id = "run-123"
        agent_config.agent_id = "agent-123"
        agent_config.conversation_id = "auto-generated"
        agent_config.llms = []
        agent_config.input = {}

        from app.domain.vo.agentvo import AgentRunOptionsVo
        from app.router.agent_controller_pkg.run_agent_v2.process_options import (
            process_options,
        )

        options = AgentRunOptionsVo()

        with patch("app.common.stand_log.StandLogger.warn") as mock_standard_warn:
            process_options(options, agent_config, MagicMock())

        mock_standard_warn.assert_called_once()
