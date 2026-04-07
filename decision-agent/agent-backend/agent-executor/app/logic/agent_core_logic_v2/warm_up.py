from typing import Dict, Optional, TYPE_CHECKING

from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from .trace import span_set_attrs

from .run_dolphin import run_dolphin
from ...common.stand_log import StandLogger

from app.utils.observability.observability_log import get_logger as o11y_logger
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    set_user_account_id,
    set_user_account_type,
)

if TYPE_CHECKING:
    from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


class WarmUpHandler:
    def __init__(self, agent_core: "AgentCoreV2"):
        self.agentCore = agent_core

    @internal_span()
    async def warnup(
        self,
        headers: Dict[str, str],
        span: Optional[Span] = None,
    ) -> None:
        """运行Agent的WarmUp

        Args:
            headers: HTTP请求头
            :param headers:
            :param span:

        """

        agent_config = self.agentCore.agent_config

        span_set_attrs(
            span=span,
            agent_run_id=agent_config.agent_run_id,
            agent_id=agent_config.agent_id,
            user_id=get_user_account_id(headers) or "",
        )

        try:
            # 处理不同类型的输入
            temp_files = {}

            # 将认证信息添加到 context_variables 中
            context_variables = {}
            set_user_account_id(context_variables, get_user_account_id(headers) or "")
            set_user_account_type(
                context_variables, get_user_account_type(headers) or ""
            )

            # 运行Dolphin引擎(warmup 模式)
            generator = run_dolphin(
                self.agentCore,
                agent_config,
                context_variables,
                headers,
                False,
                temp_files,
            )

            async for _ in generator:
                pass

            StandLogger.debug("AgentCore warnup end")

        except Exception as e:
            # 处理整体异常
            o11y_logger().error(f"agent warnup failed: {e}")
