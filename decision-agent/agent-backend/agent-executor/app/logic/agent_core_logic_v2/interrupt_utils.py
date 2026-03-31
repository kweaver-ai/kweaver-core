# -*- coding:utf-8 -*-
"""中断处理工具方法"""

from typing import Dict, Any, AsyncGenerator

from dolphin.sdk.agent.dolphin_agent import DolphinAgent


def check_and_raise_interrupt(item: Dict[str, Any]) -> None:
    """检查是否是中断响应，如果是则抛出 ToolInterruptException

    用于 run_dolphin.py 和 resume_handler.py 中的中断处理复用。

    Args:
        item: agent.arun() 返回的项

    Raises:
        ToolInterruptException: 当检测到工具中断时
    """
    if not isinstance(item, dict):
        return

    if item.get("status") != "interrupted":
        return

    interrupt_type = item.get("interrupt_type")

    if interrupt_type == "tool_confirmation":
        # 使用自定义异常
        from app.common.exceptions.tool_interrupt import (
            ToolInterruptException,
            ToolInterruptInfo,
        )

        handle = item.get("handle")
        data = item.get("data", {})

        interrupt_info = ToolInterruptInfo(
            handle=handle,
            data=data,
        )
        raise ToolInterruptException(interrupt_info)


async def process_arun_loop(
    agent: DolphinAgent,
    is_debug: bool = False,
) -> AsyncGenerator[Dict[str, Any], None]:
    """处理 agent.arun() 循环的公共方法

    用于 run_dolphin.py 和 resume_handler.py 中的 arun 循环复用。

    Args:
        agent: DolphinAgent 实例
        is_debug: 是否 debug 模式

    Yields:
        处理后的 output dict，包含 answer 和 context

    Raises:
        ToolInterrupt: 当检测到工具中断时
    """
    # 延迟导入避免循环依赖
    from app.utils.common import get_dolphin_var_value

    async for item in agent.arun():
        # 检查是否是中断，如果是则抛出 ToolInterrupt
        check_and_raise_interrupt(item)

        # 正常处理
        if not is_debug and item.get("_progress"):
            item["_progress"] = [
                p for p in item["_progress"] if p.get("stage") != "assign"
            ]

        item_value = {key: get_dolphin_var_value(value) for key, value in item.items()}

        output = {
            "answer": item_value,
            "context": agent.executor.context.get_all_variables_values(),
        }

        yield output
