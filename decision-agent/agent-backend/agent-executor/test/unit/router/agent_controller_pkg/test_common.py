"""单元测试 - agent_controller_pkg/common 模块"""

import pytest
from unittest.mock import MagicMock


class TestRunAgentParam:
    """测试 RunAgentParam 模型"""

    def test_run_agent_param_creation(self):
        """测试创建 RunAgentParam"""
        from app.router.agent_controller_pkg.common import RunAgentParam
        from app.common.structs import AgentConfig, AgentInput

        param = RunAgentParam(
            id="test-id", config=AgentConfig(), input=AgentInput(query="test query")
        )

        assert param.id == "test-id"
        assert param.config is not None
        assert param.input is not None

    def test_run_agent_param_with_defaults(self):
        """测试使用默认值创建 RunAgentParam"""
        from app.router.agent_controller_pkg.common import RunAgentParam
        from app.common.structs import AgentInput

        param = RunAgentParam(input=AgentInput(query="test query"))

        assert param.id is None
        assert param.config is None
        assert param.options is None


class TestRunAgentResponse:
    """测试 RunAgentResponse 模型"""

    def test_run_agent_response_creation(self):
        """测试创建 RunAgentResponse"""
        from app.router.agent_controller_pkg.common import RunAgentResponse

        response = RunAgentResponse(answer={"result": "test"}, status="True")

        assert response.answer == {"result": "test"}
        assert response.status == "True"


@pytest.mark.asyncio
class TestProcessOptions:
    """测试 process_options 函数"""

    async def test_process_none_options(self):
        """测试处理 None options"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput

        agent_config = AgentConfig()
        agent_input = AgentInput(query="test query")

        result = process_options(None, agent_config, agent_input)

        assert result is None

    async def test_process_with_output_vars(self):
        """测试处理 output_vars 选项"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_input = AgentInput(query="test query")

        options = AgentOptions()
        options.output_vars = ["var1", "var2"]

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        assert agent_config.output_vars == ["var1", "var2"]

    async def test_process_with_incremental_output(self):
        """测试处理 incremental_output 选项"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_input = AgentInput(query="test query")

        options = AgentOptions()
        options.incremental_output = True

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        assert agent_config.incremental_output is True

    async def test_process_with_data_source(self):
        """测试处理 data_source 选项"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_input = AgentInput(query="test query")

        options = AgentOptions()
        data_source = {"type": "database", "connection": "test"}
        options.data_source = data_source

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        assert agent_config.data_source == data_source

    async def test_process_with_llm_config_existing(self):
        """测试处理已存在的 LLM 配置"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_config.llms = [
            {"is_default": False, "llm_config": {"name": "gpt-3.5"}},
            {"is_default": False, "llm_config": {"name": "gpt-4"}},
        ]
        agent_input = AgentInput(query="test query")

        options = AgentOptions()
        options.llm_config = {"name": "gpt-4"}

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        assert agent_config.llms[0]["is_default"] is False
        assert agent_config.llms[1]["is_default"] is True

    async def test_process_with_llm_config_new(self):
        """测试处理新的 LLM 配置"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_config.llms = []
        agent_input = AgentInput(query="test query")

        options = AgentOptions()
        options.llm_config = {"name": "gpt-4"}

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        assert len(agent_config.llms) == 1
        assert agent_config.llms[0]["is_default"] is True
        assert agent_config.llms[0]["llm_config"]["name"] == "gpt-4"

    async def test_process_with_tmp_files(self):
        """测试处理 tmp_files 选项"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_config.input = {"fields": [{"name": "file", "type": "file"}]}
        agent_input = AgentInput(query="test query")
        agent_input.set_value = MagicMock()

        options = AgentOptions()
        options.tmp_files = ["file1.txt", "file2.txt"]

        span = MagicMock()
        span.is_recording = MagicMock(return_value=False)

        process_options(options, agent_config, agent_input, span)

        agent_input.set_value.assert_called_once_with(
            "file", ["file1.txt", "file2.txt"]
        )

    async def test_process_with_recording_span(self):
        """测试处理带有 recording span"""
        from app.router.agent_controller_pkg.common import process_options
        from app.common.structs import AgentConfig, AgentInput, AgentOptions

        agent_config = AgentConfig()
        agent_config.session_id = "session-123"
        agent_config.agent_id = "agent-123"
        agent_input = AgentInput(query="test query")

        options = AgentOptions()

        span = MagicMock()
        span.is_recording = MagicMock(return_value=True)

        process_options(options, agent_config, agent_input, span)

        span.set_attribute.assert_any_call("session_id", "session-123")
        span.set_attribute.assert_any_call("agent_id", "agent-123")
