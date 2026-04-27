from opentelemetry.trace import Span


def span_set_attrs(
    span: Span,
    agent_run_id: str = None,
    agent_id: str = None,
    user_id: str = None,
    conversation_id: str = None,
    is_root_span: bool = False,
):
    """设置span属性（包含OTel GenAI标准字段）

    Args:
        span: OpenTelemetry Span对象
        agent_run_id: Agent运行ID
        agent_id: Agent ID
        user_id: 用户ID
        conversation_id: 会话ID
        is_root_span: 是否为根span（用于设置gen_ai.operation.name）
    """

    if span and span.is_recording():
        # 根span需要设置operation.name为invoke_agent
        if is_root_span:
            span.set_attribute("gen_ai.operation.name", "invoke_agent")

        if agent_run_id is not None:
            span.set_attribute("agent_run_id", agent_run_id)

        if agent_id is not None:
            span.set_attribute("agent_id", agent_id)
            span.set_attribute("gen_ai.agent.id", agent_id)

        if user_id is not None:
            span.set_attribute("user_id", user_id)
            span.set_attribute("agent.user.id", user_id)

        if conversation_id is not None:
            span.set_attribute("gen_ai.conversation.id", conversation_id)
