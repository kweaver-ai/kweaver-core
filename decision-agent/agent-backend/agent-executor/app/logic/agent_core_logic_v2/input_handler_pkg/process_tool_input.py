from typing import Any, Dict, Optional

from app.domain.vo.agentvo import AgentInputVo
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from ..trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import get_user_account_id


@internal_span()
async def process_tool_input(
    inputs: AgentInputVo,
    span: Optional[Span] = None,
) -> tuple[Dict[str, Any], Optional[str]]:
    """处理工具输入

    Args:
        inputs: 输入参数

    Returns:
        tuple: (处理后的上下文变量, 可能更新的event_key)

    Note:
        Resume 时前端会通过新的 /v2/agent/resume 接口传递 handle，
        不再需要从 Redis 获取上下文。
    """
    # 1. o11y记录
    span_set_attrs(
        span=span,
        user_id=get_user_account_id(inputs.header) or "" if inputs.header else "",
    )

    # 返回输入的字典形式
    inputs_dict = inputs.model_dump()

    # 移除tool字段（如果存在）
    if "tool" in inputs_dict:
        inputs_dict.pop("tool")

    return inputs_dict, None
