import json
from typing import Any, Dict, Optional
from app.common.exceptions.tool_interrupt import ToolInterruptException

from app.common.stand_log import StandLogger
from app.utils.json import custom_serializer
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from .trace import span_set_attrs


class InterruptHandler:
    @classmethod
    @internal_span()
    async def handle_tool_interrupt(
        cls,
        tool_interrupt: ToolInterruptException,
        res: Dict[str, Any],
        context_variables: Dict[str, Any],
        span: Optional[Span] = None,
    ) -> None:
        """处理工具中断

        Args:
            tool_interrupt: 工具中断异常
            res: 结果字典
            context_variables: 上下文变量
        """

        span_set_attrs(
            span=span,
            agent_run_id=context_variables.get("session_id", ""),
            agent_id=context_variables.get("agent_id", ""),
        )

        StandLogger.info(f"ToolInterruptException: {tool_interrupt}")

        # 直接使用 interrupt_info（dataclass 会被 custom_serializer 正确序列化）
        res["interrupt_info"] = tool_interrupt.interrupt_info

        res["status"] = "True"

        StandLogger.info(
            f"ToolInterrupt res: {json.dumps(res, ensure_ascii=False, default=custom_serializer)}"
        )
