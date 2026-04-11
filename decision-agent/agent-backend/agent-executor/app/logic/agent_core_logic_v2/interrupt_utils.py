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

    is_enabled = getattr(Config.evidence, "enable", False)

    if not isinstance(item, dict):
        return evidence_store_key

    if not is_enabled:
        return evidence_store_key

    # 从 _progress 数组中提取工具结果
    progress_array = item.get("_progress", [])

    if not progress_array:
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

    if not tool_call_results:
        return evidence_store_key

    try:
        import uuid
        import json

        from app.logic.agent_core_logic_v2.evidence_store import (
            get_global_evidence_store,
        )

        store = get_global_evidence_store()

        if not evidence_store_key:
            evidence_store_key = f"ev_{uuid.uuid4().hex[:12]}"

        # 新策略：直接存储原始工具结果，不提取实体
        # 实体提取将在 LLM 输出时由 LLM 自己标注
        tool_results_for_storage = []
        for tool_name, result in tool_call_results.items():
            # 将结果转换为字符串存储（支持任意类型）
            result_str = json.dumps(result, ensure_ascii=False, default=str)

            # 只存储非空且有意义的结果（过滤掉空字符串或过短的结果）
            if len(result_str) > 5:
                tool_results_for_storage.append({
                    "tool_name": tool_name,
                    "result": result_str,
                    "result_type": type(result).__name__,
                })

        # 存储到 EvidenceStore（使用特殊的 key 前缀表示这是原始工具结果）
        if tool_results_for_storage:
            raw_results_key = f"{evidence_store_key}_raw"

            # 累积存储：获取已存储的工具结果，合并新的结果
            existing_results = store.get(raw_results_key) or []

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
                            existing_results[idx] = new_tr
                        should_add = False
                        break

                if should_add:
                    existing_results.append(new_tr)

            # 存储合并后的结果
            store.add(raw_results_key, existing_results)

        logger.info(
            f"[_check_and_prepare_evidence] Stored {len(tool_call_results)} "
            f"raw tool results, key={evidence_store_key}"
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
        current_evidence_key = await _check_and_prepare_evidence(
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
