# -*- coding:utf-8 -*-
"""中断处理工具方法"""

import logging
from typing import Dict, Any, AsyncGenerator, Optional

from dolphin.sdk.agent.dolphin_agent import DolphinAgent

logger = logging.getLogger(__name__)


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


def _check_and_prepare_evidence(
    item: Dict[str, Any],
    evidence_store_key: Optional[str],
) -> Optional[str]:
    """检查 item 中是否有工具结果，如有则准备证据

    Args:
        item: agent.arun() 返回的项
        evidence_store_key: 可选的现有 evidence store key

    Returns:
        evidence_store_key (新建或原有)
    """
    if not isinstance(item, dict):
        return evidence_store_key

    from app.common.config import Config

    if not getattr(Config.features, "enable_evidence_injection", False):
        return evidence_store_key

    context = item.get("context", {})
    if not context:
        return evidence_store_key

    tool_call_results = context.get("_tool_call_results", {})
    if not tool_call_results or not isinstance(tool_call_results, dict):
        return evidence_store_key

    try:
        import uuid

        from app.logic.agent_core_logic_v2.evidence_store import (
            get_global_evidence_store,
        )
        from app.logic.tool.evidence_prepare import evidence_prepare

        store = get_global_evidence_store()

        if not evidence_store_key:
            evidence_store_key = f"ev_{uuid.uuid4().hex[:12]}"
            store.add(evidence_store_key, [])

        current_evidences = store.get(evidence_store_key) or []

        for tool_name, result in tool_call_results.items():
            if not isinstance(result, dict):
                continue

            try:
                import asyncio

                loop = asyncio.get_event_loop()
                prepare_result = loop.run_until_complete(
                    evidence_prepare(
                        tool_call_result=result,
                        config={
                            "llm_extraction_timeout": getattr(
                                Config.features,
                                "llm_extraction_timeout",
                                30,
                            ),
                            "llm_extraction_model": getattr(
                                Config.features,
                                "llm_extraction_model",
                                "",
                            ),
                        },
                        context={"tool_name": tool_name},
                    )
                )

                evidences = prepare_result.get("evidences", [])
                if evidences:
                    current_evidences.extend(evidences)
                    logger.info(
                        f"[_check_and_prepare_evidence] tool={tool_name}, "
                        f"extracted={len(evidences)} evidences"
                    )

            except Exception as e:
                logger.warning(
                    f"[_check_and_prepare_evidence] Failed to process {tool_name}: {e}"
                )

        if current_evidences:
            store.add(evidence_store_key, current_evidences)

        logger.info(
            f"[_check_and_prepare_evidence] Total {len(current_evidences)} "
            f"evidences, key={evidence_store_key}"
        )

    except Exception as e:
        logger.error(f"[_check_and_prepare_evidence] Error: {e}", exc_info=True)

    return evidence_store_key


async def process_arun_loop(
    agent: DolphinAgent,
    is_debug: bool = False,
    evidence_store_key: Optional[str] = None,
) -> AsyncGenerator[Dict[str, Any], None]:
    """处理 agent.arun() 循环的公共方法

    用于 run_dolphin.py 和 resume_handler.py 中的 arun 循环复用。

    Args:
        agent: DolphinAgent 实例
        is_debug: 是否 debug 模式
        evidence_store_key: 可选的 evidence store key，用于证据注入

    Yields:
        处理后的 output dict，包含 answer, context, 和可选的 evidence_store_key

    Raises:
        ToolInterrupt: 当检测到工具中断时
    """
    # 延迟导入避免循环依赖
    from app.utils.common import get_dolphin_var_value

    current_evidence_key = evidence_store_key

    async for item in agent.arun():
        # 检查是否是中断，如果是则抛出 ToolInterrupt
        check_and_raise_interrupt(item)

        # 检查并准备证据（在工具完成后）
        current_evidence_key = _check_and_prepare_evidence(
            item, current_evidence_key
        )

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

        if current_evidence_key:
            output["evidence_store_key"] = current_evidence_key

        yield output
