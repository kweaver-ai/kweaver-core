"""
Evidence Injection - 证据注入模块

使用 LLM 对流式输出进行证据标注，在 LLM 生成的文本中
标注引用工具结果的位置，并将 _evidence 元数据注入到 _progress 数组中。
"""

import logging
from typing import Any, AsyncGenerator, Dict, List, Optional

from app.common.config import Config

logger = logging.getLogger(__name__)


async def create_evidence_injection_stream(
    original_stream: AsyncGenerator[Dict[str, Any], None],
    evidence_store_key: Optional[str] = None,
    is_debug: bool = False,
) -> AsyncGenerator[Dict[str, Any], None]:
    """
    创建带证据注入的流式输出包装器。

    从流中的每个 item 获取 evidence_store_key，
    如果存在证据则进行注入标注。

    Args:
        original_stream: 原始流式输出
        evidence_store_key: 已废弃参数（保留兼容性）
        is_debug: 调试模式

    Yields:
        可能包含 _evidence 元数据的输出字典
    """
    from app.logic.agent_core_logic_v2.evidence_store import (
        get_global_evidence_store,
    )
    from app.common.config import Config

    # 检查是否仅标注最终答案
    annotate_only_final = getattr(
        Config.evidence, "annotate_only_final_answer", True
    )

    store = get_global_evidence_store()
    current_tool_results = []  # 存储当前的工具结果
    _loaded_key = None  # 记录已加载的 key
    last_llm_item = None  # 缓存最后一个包含 LLM stage 的 item
    last_llm_item_tool_results = []  # 缓存对应的工具结果
    last_llm_item_loaded_key = None  # 缓存对应的 loaded_key

    if not annotate_only_final:
        # 如果不是仅标注最终答案，使用原有的流式处理逻辑
        async for item in original_stream:
            item_evidence_key = item.get("evidence_store_key")

            # 更新工具结果
            if item_evidence_key and (item_evidence_key != _loaded_key or not current_tool_results):
                raw_results_key = f"{item_evidence_key}_raw"
                tool_results = store.get(raw_results_key)
                if tool_results:
                    current_tool_results = tool_results
                    _loaded_key = item_evidence_key

            # 处理 item
            async for result in _process_single_item(
                item,
                store,
                current_tool_results,
                _loaded_key,
                is_final_answer=False,
                annotate_only_final=False,
            ):
                yield result
    else:
        # 仅标注最终答案：缓存最后一个包含 LLM stage 的 item
        async for item in original_stream:
            item_evidence_key = item.get("evidence_store_key")

            # 更新工具结果
            if item_evidence_key and (item_evidence_key != _loaded_key or not current_tool_results):
                raw_results_key = f"{item_evidence_key}_raw"
                tool_results = store.get(raw_results_key)
                if tool_results:
                    current_tool_results = tool_results
                    _loaded_key = item_evidence_key

            # 检查当前 item 是否包含 LLM stage
            has_llm_stage = _check_has_llm_stage(item)

            if has_llm_stage:
                # 缓存这个包含 LLM stage 的 item
                last_llm_item = item
                last_llm_item_tool_results = current_tool_results
                last_llm_item_loaded_key = _loaded_key

            # 直接发送当前 item（不进行标注）
            yield item

        # 流结束后，对缓存的 LLM item 进行标注
        if last_llm_item is not None:
            # 对缓存的 LLM item 进行标注
            annotated = False
            async for result in _process_single_item(
                last_llm_item,
                store,
                last_llm_item_tool_results,
                last_llm_item_loaded_key,
                is_final_answer=True,
                annotate_only_final=True,
            ):
                if not annotated:
                    yield result
                    annotated = True


def _check_has_llm_stage(item: Dict[str, Any]) -> bool:
    """检查 item 是否包含 LLM stage"""
    answer = item.get("answer", {})
    if isinstance(answer, dict):
        progress_array = answer.get("_progress", [])
        if isinstance(progress_array, list):
            for p in progress_array:
                if p.get("stage") == "llm":
                    return True

    # 同时检查顶层的 _progress
    progress_array = item.get("_progress", [])
    if isinstance(progress_array, list):
        for p in progress_array:
            if p.get("stage") == "llm":
                return True

    return False


async def _process_single_item(
    item: Dict[str, Any],
    store: Any,
    current_tool_results: List[Dict[str, Any]],
    _loaded_key: Optional[str],
    is_final_answer: bool,
    annotate_only_final: bool,
) -> AsyncGenerator[Dict[str, Any], None]:
    """
    处理单个 item，根据配置决定是否进行 LLM 标注。

    Args:
        item: 待处理的 item
        store: EvidenceStore 实例
        current_tool_results: 当前缓存的工具结果
        _loaded_key: 已加载的 evidence_store_key
        is_final_answer: 是否是最终答案
        annotate_only_final: 是否仅标注最终答案

    Yields:
        处理后的 item（可能包含 _evidence）
    """
    from app.common.config import Config

    # 从当前 item 中获取 evidence_store_key
    item_evidence_key = item.get("evidence_store_key")

    # 如果有新的 evidence_store_key（或者还没有加载过，或者当前没有工具结果），从 EvidenceStore 获取原始工具结果
    if item_evidence_key and (item_evidence_key != _loaded_key or not current_tool_results):
        raw_results_key = f"{item_evidence_key}_raw"
        tool_results = store.get(raw_results_key)
        if tool_results:
            current_tool_results = tool_results
            _loaded_key = item_evidence_key
        else:
            current_tool_results = []
            _loaded_key = item_evidence_key

    # 提取 LLM 生成的文本
    answer = item.get("answer", {})
    actual_text = None

    if isinstance(answer, dict):
        candidate = answer.get("answer", "")
        if isinstance(candidate, str):
            actual_text = candidate
        elif isinstance(candidate, dict):
            candidate_stage = candidate.get("stage", "")
            candidate_answer = candidate.get("answer")
            if candidate_stage == "llm" and isinstance(candidate_answer, str):
                actual_text = candidate_answer
            else:
                progress_array = item.get("_progress", [])
                if progress_array:
                    for p in reversed(progress_array):
                        if p.get("stage") == "llm" and isinstance(p.get("answer"), str):
                            actual_text = p.get("answer")
                            break
    elif isinstance(answer, str):
        actual_text = answer

    if not isinstance(actual_text, str):
        actual_text = ""

    text_len = len(actual_text)

    # 判断是否需要进行 LLM 标注
    # 条件：有工具结果 + 有足够长的文本 + (是最终答案 或 不限制仅标注最终答案)
    should_annotate = (
        current_tool_results and
        text_len > 5 and
        (is_final_answer or not annotate_only_final)
    )

    if should_annotate:
        try:
            from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
                llm_annotate_evidence,
            )

            annotation_config = {
                "llm_annotation_timeout": getattr(
                    Config.evidence, "llm_annotation_timeout", 30
                ),
                "llm_annotation_model": getattr(
                    Config.evidence, "llm_annotation_model", "deepseek-v3.2"
                ),
            }

            annotation_result = await llm_annotate_evidence(
                text=actual_text,
                tool_results=current_tool_results,
                config=annotation_config,
            )

            # 转换为 evidence_meta 格式
            evidences = annotation_result.get("evidences", [])
            evidence_meta = {}

            # 获取第一个（或唯一的）工具结果
            tool_data = current_tool_results[0] if current_tool_results else None

            for idx, ev in enumerate(evidences):
                local_id = f"e{idx + 1}"
                object_type = ev.get("object_type_name", "")

                evidence_item = {
                    "object_type_name": object_type,
                    "positions": ev.get("positions", []),
                }

                # 添加工具结果数据（如果有）
                if tool_data:
                    evidence_item["tool_name"] = tool_data.get("tool_name", "")
                    evidence_item["result"] = tool_data.get("result", {})
                    evidence_item["result_type"] = tool_data.get("result_type", "")

                evidence_meta[local_id] = evidence_item

            # 注入到 _progress
            _inject_evidence_to_progress(item, evidence_meta)

        except Exception as e:
            logger.warning(f"[_process_single_item] Annotation failed: {e}", exc_info=True)

    # 缓存当前状态以供下次使用
    item["_cached_tool_results"] = current_tool_results
    item["_cached_loaded_key"] = _loaded_key

    yield item


def _inject_evidence_to_progress(
    item: Dict[str, Any],
    evidence_meta: Dict[str, Any],
) -> None:
    """
    将 _evidence 注入到 item 的 _progress 数组中。

    Args:
        item: 待注入的 item
        evidence_meta: 证据元数据
    """
    # 注入到 item["answer"]["_progress"]
    answer_dict = item.get("answer")

    if isinstance(answer_dict, dict) and "_progress" in answer_dict:
        progress_array = answer_dict["_progress"]

        if isinstance(progress_array, list) and progress_array:
            # 从后往前查找最后一个 LLM stage 的 progress 条目
            # 因为最后一个可能是 "assign" stage，我们需要找到 LLM 生成的文本
            llm_progress_index = None
            for i in range(len(progress_array) - 1, -1, -1):
                if progress_array[i].get("stage") == "llm":
                    llm_progress_index = i
                    break

            if llm_progress_index is not None:
                llm_progress = progress_array[llm_progress_index]
                llm_progress["_evidence"] = evidence_meta

                # 同时注入到 item["_progress"]（如果存在且是不同的对象）
                if "_progress" in item and isinstance(item["_progress"], list):
                    top_progress_array = item["_progress"]
                    # 找到对应的 LLM stage 条目
                    for idx, p in enumerate(top_progress_array):
                        if p.get("stage") == "llm":
                            # 通过 answer 内容匹配（因为 id 可能不同）
                            if (p.get("answer") == llm_progress.get("answer") or
                                p.get("id") == llm_progress.get("id")):
                                p["_evidence"] = evidence_meta
                                break
