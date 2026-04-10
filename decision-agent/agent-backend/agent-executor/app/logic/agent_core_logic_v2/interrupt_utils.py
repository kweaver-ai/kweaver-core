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


async def _check_and_prepare_evidence(
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

    StandLogger.info_log("  Scanning progress entries for skill stages:")
    for idx, progress_item in enumerate(progress_array):
        stage = progress_item.get("stage", "unknown")
        StandLogger.info_log(f"    [{idx}] stage={stage}")
        if stage == "skill":
            skill_info = progress_item.get("skill_info", {})
            tool_name = skill_info.get("name", "")
            answer = progress_item.get("answer")
            StandLogger.info_log(
                f"      skill_info={skill_info}, tool_name={tool_name}, "
                f"answer_type={type(answer).__name__ if answer is not None else 'None'}"
            )

    for progress_item in progress_array:
        if progress_item.get("stage") != "skill":
            continue

        skill_info = progress_item.get("skill_info", {})
        tool_name = skill_info.get("name", "")
        answer = progress_item.get("answer")

        StandLogger.info_log(
            f"  Processing skill: tool_name={tool_name}, "
            f"has_answer={answer is not None}, "
            f"valid_tool_name={bool(tool_name)}"
        )

        if not tool_name or answer is None:
            StandLogger.info_log(
                f"    SKIP: tool_name='{tool_name}', answer={answer}"
            )
            continue

        # 如果同一个工具被调用多次，使用最后一次的结果
        # 或者可以改为存储为列表
        tool_call_results[tool_name] = answer
        StandLogger.info_log(f"    ✓ Added tool_call_results[{tool_name}]")

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
        import json

        from app.logic.agent_core_logic_v2.evidence_store import (
            get_global_evidence_store,
        )

        store = get_global_evidence_store()

        if not evidence_store_key:
            evidence_store_key = f"ev_{uuid.uuid4().hex[:12]}"
            StandLogger.info_log(
                f"[_check_and_prepare_evidence] Created new evidence key: {evidence_store_key}"
            )

        # 新策略：直接存储原始工具结果，不提取实体
        # 实体提取将在 LLM 输出时由 LLM 自己标注
        tool_results_for_storage = []
        for tool_name, result in tool_call_results.items():
            # 将结果转换为字符串存储（支持任意类型）
            result_str = json.dumps(result, ensure_ascii=False, default=str)

            # 只存储非空且有意义的结果（过滤掉空字符串或过短的结果）
            if len(result_str) > 10:
                tool_results_for_storage.append({
                    "tool_name": tool_name,
                    "result": result_str,
                    "result_type": type(result).__name__,
                })
                StandLogger.info_log(
                    f"[_check_and_prepare_evidence] Stored tool result: "
                    f"tool_name={tool_name}, "
                    f"result_length={len(result_str)}, "
                    f"result_type={type(result).__name__}"
                )
            else:
                StandLogger.info_log(
                    f"[_check_and_prepare_evidence] SKIP tool result (too short): "
                    f"tool_name={tool_name}, "
                    f"result_length={len(result_str)}, "
                    f"result_type={type(result).__name__}"
                )

        # 存储到 EvidenceStore（使用特殊的 key 前缀表示这是原始工具结果）
        if tool_results_for_storage:
            raw_results_key = f"{evidence_store_key}_raw"

            # 累积存储：获取已存储的工具结果，合并新的结果
            existing_results = store.get(raw_results_key) or []
            StandLogger.info_log(
                f"[_check_and_prepare_evidence] Existing tool results: {len(existing_results)}"
            )

            # 记录已存储的工具名称
            existing_tool_names = {tr.get("tool_name") for tr in existing_results}

            # 只添加新工具或更新的工具（result 长度更长的）
            for new_tr in tool_results_for_storage:
                tool_name = new_tr.get("tool_name")
                new_result = new_tr.get("result", "")
                new_length = len(new_result)

                # 检查是否需要添加或更新
                should_add = True
                for idx, existing_tr in enumerate(existing_results):
                    if existing_tr.get("tool_name") == tool_name:
                        existing_result = existing_tr.get("result", "")
                        existing_length = len(existing_result)

                        # 只有当新结果更长时才更新
                        if new_length > existing_length:
                            StandLogger.info_log(
                                f"[_check_and_prepare_evidence] Updating tool result: "
                                f"tool_name={tool_name}, "
                                f"old_length={existing_length}, "
                                f"new_length={new_length}"
                            )
                            existing_results[idx] = new_tr
                        else:
                            StandLogger.info_log(
                                f"[_check_and_prepare_evidence] Keeping existing tool result: "
                                f"tool_name={tool_name}, "
                                f"existing_length={existing_length}, "
                                f"new_length={new_length}"
                            )
                        should_add = False
                        break

                if should_add:
                    existing_results.append(new_tr)
                    StandLogger.info_log(
                        f"[_check_and_prepare_evidence] Adding new tool result: "
                        f"tool_name={tool_name}, length={new_length}"
                    )

            # 存储合并后的结果
            store.add(raw_results_key, existing_results)
            StandLogger.info_log(
                f"[_check_and_prepare_evidence] Stored total {len(existing_results)} "
                f"raw tool results to {raw_results_key}"
            )

        logger.info(
            f"[_check_and_prepare_evidence] Stored {len(tool_call_results)} "
            f"raw tool results, key={evidence_store_key}"
        )

        StandLogger.info_log(
            f"[_check_and_prepare_evidence] END: "
            f"stored={len(tool_results_for_storage)} raw tool results, "
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
        current_evidence_key = await _check_and_prepare_evidence(
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

        # 验证 _progress 引用关系
        if "_progress" in item_value and "_progress" in item:
            StandLogger.info_log(
                f"[process_arun_loop] _progress reference check: "
                f"item_value['_progress'] is item['_progress']: {item_value['_progress'] is item['_progress']}, "
                f"len={len(item_value.get('_progress', []))}"
            )

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
