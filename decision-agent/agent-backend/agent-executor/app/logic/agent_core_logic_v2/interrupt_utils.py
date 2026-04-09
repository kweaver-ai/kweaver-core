# -*- coding:utf-8 -*-
"""中断处理工具方法"""

import logging
from typing import Dict, Any, AsyncGenerator, Optional

from dolphin.sdk.agent.dolphin_agent import DolphinAgent

from app.common.stand_log import StandLogger

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
    from app.common.config import Config

    # StandLogger: 检查证据注入功能是否启用
    is_enabled = getattr(Config.features, "enable_evidence_injection", False)
    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[_check_and_prepare_evidence] START\n"
        f"  Evidence injection enabled: {is_enabled}\n"
        f"  Config.features.enable_evidence_injection = {getattr(Config.features, 'enable_evidence_injection', 'NOT_SET')}\n"
    )

    if not isinstance(item, dict):
        StandLogger.info_log("[_check_and_prepare_evidence] END: item is not dict\n" + "="*60)
        return evidence_store_key

    if not is_enabled:
        StandLogger.info_log("[_check_and_prepare_evidence] END: evidence injection disabled\n" + "="*60)
        return evidence_store_key

    # 从 _progress 数组中提取工具结果
    progress_array = item.get("_progress", [])
    StandLogger.info_log(
        f"  Progress entries: {len(progress_array)}\n"
        f"  Has _progress: {bool(progress_array)}\n"
    )

    if not progress_array:
        StandLogger.info_log("[_check_and_prepare_evidence] END: no progress entries\n" + "="*60)
        return evidence_store_key

    # 提取所有 stage="skill" 的工具调用结果
    tool_call_results = {}
    for progress_item in progress_array:
        if progress_item.get("stage") != "skill":
            continue

        skill_info = progress_item.get("skill_info", {})
        tool_name = skill_info.get("name", "")
        answer = progress_item.get("answer")

        if not tool_name or answer is None:
            continue

        # 如果同一个工具被调用多次，使用最后一次的结果
        # 或者可以改为存储为列表
        tool_call_results[tool_name] = answer

    StandLogger.info_log(
        f"  Tool call results extracted: {list(tool_call_results.keys()) if tool_call_results else 'None'}\n"
    )

    if not tool_call_results:
        StandLogger.info_log("[_check_and_prepare_evidence] END: no valid tool results\n" + "="*60)
        return evidence_store_key

    StandLogger.info_log(
        f"[_check_and_prepare_evidence] Found {len(tool_call_results)} tool results"
    )

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
            StandLogger.info_log(
                f"[_check_and_prepare_evidence] Created new evidence key: {evidence_store_key}"
            )

        current_evidences = store.get(evidence_store_key) or []

        for tool_name, result in tool_call_results.items():
            StandLogger.info_log(
                f"[_check_and_prepare_evidence] Processing tool={tool_name}, "
                f"result_type={type(result).__name__}"
            )

            if not isinstance(result, dict):
                StandLogger.info_log(
                    f"[_check_and_prepare_evidence] SKIP: result is not dict, type={type(result).__name__}"
                )
                continue

            try:
                import asyncio

                logger.info(
                    f"[_check_and_prepare_evidence] Processing tool={tool_name}, "
                    f"result keys: {list(result.keys()) if isinstance(result, dict) else type(result)}, "
                    f"config: llm_extraction_timeout={getattr(Config.features, 'llm_extraction_timeout', 30)}, "
                    f"llm_extraction_model='{getattr(Config.features, 'llm_extraction_model', '')}'"
                )

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

                logger.info(
                    f"[_check_and_prepare_evidence] evidence_prepare returned: "
                    f"evidence_id={prepare_result.get('evidence_id')}, "
                    f"evidences_count={len(prepare_result.get('evidences', []))}, "
                    f"summary={prepare_result.get('summary')}"
                )

                evidences = prepare_result.get("evidences", [])
                if evidences:
                    current_evidences.extend(evidences)
                    logger.info(
                        f"[_check_and_prepare_evidence] tool={tool_name}, "
                        f"extracted={len(evidences)} evidences, "
                        f"total_evidences={len(current_evidences)}"
                    )
                else:
                    logger.warning(
                        f"[_check_and_prepare_evidence] tool={tool_name} returned no evidences"
                    )

            except Exception as e:
                logger.error(
                    f"[_check_and_prepare_evidence] Failed to process {tool_name}: {e}",
                    exc_info=True
                )

        if current_evidences:
            store.add(evidence_store_key, current_evidences)

        logger.info(
            f"[_check_and_prepare_evidence] Total {len(current_evidences)} "
            f"evidences, key={evidence_store_key}"
        )

        StandLogger.info_log(
            f"[_check_and_prepare_evidence] END: "
            f"processed={len(tool_call_results)} tools, "
            f"total_evidences={len(current_evidences)}, "
            f"key={evidence_store_key}\n"
            f"{'='*60}\n"
        )

    except Exception as e:
        logger.error(f"[_check_and_prepare_evidence] Error: {e}", exc_info=True)
        StandLogger.info_log(
            f"[_check_and_prepare_evidence] ERROR: {e}\n"
            f"{'='*60}\n"
        )

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

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[process_arun_loop] START\n"
        f"  is_debug: {is_debug}\n"
        f"  initial_evidence_key: {evidence_store_key}\n"
        f"{'='*60}\n"
    )

    iteration_count = 0

    async for item in agent.arun():
        iteration_count += 1

        StandLogger.info_log(
            f"\n[process_arun_loop] Iteration {iteration_count}: "
            f"item_type={type(item).__name__}, "
            f"item_keys={list(item.keys()) if isinstance(item, dict) else 'N/A'}"
        )

        # 检查是否是中断，如果是则抛出 ToolInterrupt
        check_and_raise_interrupt(item)

        # 检查并准备证据（在工具完成后）
        before_evidence_key = current_evidence_key
        current_evidence_key = _check_and_prepare_evidence(
            item, current_evidence_key
        )

        if current_evidence_key != before_evidence_key:
            StandLogger.info_log(
                f"[process_arun_loop] Evidence key changed: "
                f"{before_evidence_key} -> {current_evidence_key}"
            )

        # 正常处理
        if not is_debug and item.get("_progress"):
            progress_count = len(item["_progress"])
            item["_progress"] = [
                p for p in item["_progress"] if p.get("stage") != "assign"
            ]
            StandLogger.info_log(
                f"[process_arun_loop] Filtered progress: "
                f"{progress_count} -> {len(item['_progress'])} entries"
            )

        item_value = {key: get_dolphin_var_value(value) for key, value in item.items()}

        output = {
            "answer": item_value,
            "context": agent.executor.context.get_all_variables_values(),
        }

        if current_evidence_key:
            output["evidence_store_key"] = current_evidence_key

        StandLogger.info_log(
            f"[process_arun_loop] Output: "
            f"answer_keys={list(output['answer'].keys()) if isinstance(output['answer'], dict) else type(output['answer']).__name__}, "
            f"context_keys={list(output['context'].keys()) if isinstance(output['context'], dict) else type(output['context']).__name__}, "
            f"has_evidence_store_key={bool(current_evidence_key)}"
        )

        yield output

    StandLogger.info_log(
        f"\n[process_arun_loop] END: "
        f"total_iterations={iteration_count}, "
        f"final_evidence_key={current_evidence_key}\n"
        f"{'='*60}\n"
    )
