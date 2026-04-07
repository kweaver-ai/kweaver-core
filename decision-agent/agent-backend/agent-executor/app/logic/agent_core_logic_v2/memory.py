import asyncio
import threading
import traceback
from typing import Any, Dict, Optional

from app.common.stand_log import StandLogger
from app.common.structs import DEFAULT_INPUTS
from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo
from app.driven.dip.agent_memory_service import agent_memory_service
from app.utils.common import (
    get_dolphin_var_final_value,
)
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span
from app.utils.observability.observability_log import get_logger as o11y_logger

from .trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import (
    get_user_account_id,
    get_user_account_type,
    set_user_account_id,
    set_user_account_type,
)


class MemoryHandler:
    def __init__(self):
        pass

    @internal_span()
    def start_memory_build_thread(
        self,
        agent_config: AgentConfigVo,
        agent_input: AgentInputVo,
        headers: Dict[str, str],
        final_result: Dict[str, Any],
        span: Optional[Span] = None,
    ) -> None:
        """启动记忆构建线程

        Args:
            agent_config: Agent配置
            agent_input: 输入参数
            headers: HTTP请求头
            final_result: 最终结果
        """

        span_set_attrs(
            span=span,
            agent_run_id=agent_config.agent_run_id or "",
            agent_id=agent_config.agent_id or "",
            user_id=get_user_account_id(headers) or "",
        )

        # 检查是否启用记忆功能
        if not agent_config.memory or not agent_config.memory.get("is_enabled"):
            return

        # 创建守护线程，确保主程序退出时线程也会退出
        memory_thread = threading.Thread(
            target=self.build_memory,
            args=(agent_config, agent_input, headers, final_result),
            daemon=True,  # 设置为守护线程
        )
        memory_thread.start()

        StandLogger.info("记忆构建线程已启动")

    @internal_span()
    def build_memory(
        self,
        agent_config: AgentConfigVo,
        agent_input: dict,
        headers: Dict[str, str],
        final_result: Dict[str, Any],
        span: Optional[Span] = None,
    ) -> None:
        """构建记忆

        Args:
            agent_config: Agent配置
            agent_input: 输入参数
            headers: HTTP请求头
            final_result: 最终结果
        """

        span_set_attrs(
            span=span,
            agent_run_id=agent_config.agent_run_id or "",
            agent_id=agent_config.agent_id or "",
            user_id=get_user_account_id(headers) or "",
        )

        try:
            # 构建消息列表用于记忆构建
            messages = []

            # 1. 添加用户输入
            inputs = ""

            for input_field in agent_config.input.get("fields", []):
                if (
                    input_field.get("type") != "file"
                    and input_field.get("name") not in DEFAULT_INPUTS
                ):
                    inputs += f"{input_field.get('name')}: {agent_input.get(input_field.get('name'))}\n"

            user_message = {"role": "user", "content": inputs}
            messages.append(user_message)

            # 2. 添加助手回复
            output_var_name = agent_config.output.get_final_answer_var()

            if final_result and output_var_name in final_result:
                try:
                    assistant_message = {
                        "role": "assistant",
                        "content": str(
                            get_dolphin_var_final_value(final_result[output_var_name])
                        ),
                    }
                    messages.append(assistant_message)
                except Exception as e:
                    StandLogger.error(
                        f"获取最终结果失败: {str(e)}, final_result: {final_result}, output_var_name: {output_var_name}"
                    )
                    traceback.print_exc()

            # 如果没有有效的消息，则不构建记忆
            if not messages:
                return

            # 3. 调用记忆构建服务
            loop = asyncio.new_event_loop()
            asyncio.set_event_loop(loop)
            try:
                req_memory_headers = {}
                user_id = get_user_account_id(headers) or "unknown"
                visitor_type = get_user_account_type(headers) or "unknown"
                set_user_account_id(req_memory_headers, user_id)
                set_user_account_type(req_memory_headers, visitor_type)

                loop.run_until_complete(
                    agent_memory_service.build_memory(
                        messages=messages,
                        agent_id=agent_config.agent_id,
                        user_id=user_id,
                        headers=req_memory_headers,
                    )
                )

            finally:
                loop.close()

            StandLogger.info("记忆构建完成")

        except Exception as e:
            # 记忆构建失败不应影响主流程
            StandLogger.error(
                f"记忆构建失败: {str(e)}, traceback: {traceback.format_exc()}"
            )
            o11y_logger().error(f"记忆构建失败: {str(e)}")
