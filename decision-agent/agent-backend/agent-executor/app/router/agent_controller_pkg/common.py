import logging

from fastapi import APIRouter
from pydantic import BaseModel, Field

from app.common.config import Config
from app.common.structs import AgentConfig, AgentInput, AgentOptions
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span


logging.getLogger("sse_starlette.sse").setLevel(logging.INFO)


router = APIRouter(prefix=Config.app.host_prefix + "/agent", tags=["agent-executor"])


class RunAgentParam(BaseModel):
    id: str = Field(
        default=None, title="id", description="agent id", example="1830930776523276288"
    )
    config: AgentConfig = Field(
        default=None,
        title="config",
        description="agent配置",
        # example=docqa_template.agent_config,
    )
    input: AgentInput = Field(
        ..., title="input", description="agent输入", example={"query": "你好"}
    )
    options: AgentOptions = Field(
        default=None,
        title="options",
        description="agent运行选项",
        alias="_options",
    )


class RunAgentResponse(BaseModel):
    answer: dict = Field(..., title="answer", description="agent最终输出")
    status: str = Field(
        ...,
        title="status",
        description='"True": 流式信息已结束; "False": 流式信息未结束，正在返回; "Error": 失败',
    )


@internal_span()
def process_options(
    options: AgentOptions,
    agent_config: AgentConfig,
    agent_input: AgentInput,
    span: Span = None,
):
    if span and span.is_recording():
        span.set_attribute("session_id", agent_config.session_id)
        span.set_attribute("agent_id", agent_config.agent_id)
    """处理agent运行选项"""
    if not options:
        return

    if options.output_vars:
        agent_config.output_vars = options.output_vars

    if options.incremental_output:
        agent_config.incremental_output = options.incremental_output

    if options.data_source:
        agent_config.data_source = options.data_source

    if options.llm_config:
        has_default = False
        for llm in agent_config.llms:
            if llm["llm_config"]["name"] == options.llm_config["name"]:
                llm["is_default"] = True
                has_default = True
            else:
                llm["is_default"] = False

        if not has_default:
            agent_config.llms.append(
                {"is_default": True, "llm_config": options.llm_config}
            )

    if options.tmp_files:
        for input_field in agent_config.input.get("fields", []):
            if input_field.get("type") == "file":
                agent_input.set_value(input_field.get("name"), options.tmp_files)
