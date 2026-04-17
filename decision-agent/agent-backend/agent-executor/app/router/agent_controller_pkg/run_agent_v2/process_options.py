from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo, AgentRunOptionsVo
from app.common.stand_log import StandLogger
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span


@internal_span()
def process_options(
    options: AgentRunOptionsVo,
    agent_config: AgentConfigVo,
    agent_input: AgentInputVo,
    span: Span = None,
):
    """
    处理agent运行选项

    Args:
        options: Agent运行选项
        agent_config: Agent配置
        agent_input: Agent输入
        span: OpenTelemetry追踪Span（可选）
    """
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

    # new add 2025年10月19日16:52:53 --start--
    if options.agent_id:
        agent_config.agent_id = options.agent_id

    if options.conversation_id:
        StandLogger.info(
            f"[process_options] Setting conversation_id from options: {options.conversation_id}"
        )
        agent_config.conversation_id = options.conversation_id
    else:
        StandLogger.warn(
            f"[process_options] No conversation_id in options, will use auto-generated value: {agent_config.conversation_id}"
        )

    if options.agent_run_id:
        agent_config.agent_run_id = options.agent_run_id
        # # 兼容老的
        # agent_config.session_id = options.agent_run_id
    # new add 2025年10月19日16:52:53 --end--

    if span and span.is_recording():
        if agent_config.agent_run_id is not None:
            span.set_attribute("session_id", agent_config.agent_run_id)
        if agent_config.agent_id is not None:
            span.set_attribute("agent_id", agent_config.agent_id)
