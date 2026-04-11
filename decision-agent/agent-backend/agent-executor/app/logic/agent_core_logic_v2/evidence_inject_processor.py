"""
EvidenceInjectProcessor 证据注入后处理器

包装 LLM 流式输出，在流式文本中检测并标注证据实体位置，
使用位置索引格式返回元数据。
"""

import logging
from typing import Any, AsyncGenerator, Dict, List, Optional, Tuple

from app.common.utils.sentence_boundary_detector import (
    SentenceBoundaryDetector,
)
from app.common.config import Config
from app.common.stand_log import StandLogger

logger = logging.getLogger(__name__)


class EvidenceInjectProcessor:
    """
    证据注入后处理器。

    包装 LLM 流式输出，在累积到句子边界时进行规则匹配，
    生成包含位置索引的 _evidence 元数据。

    工作流程：
    1. 接收流式 chunk，累积文本缓冲区
    2. 检测句子边界，对完整句子进行匹配
    3. 使用精确匹配 + 别名匹配构建索引
    4. 解决重叠匹配冲突
    5. 输出原始文本 + 位置索引元数据
    """

    def __init__(
        self,
        evidences: List[Dict[str, Any]],
        enable_alias_match: bool = True,
        min_sentence_length: int = 10,
    ):
        """
        初始化处理器。

        Args:
            evidences: 证据列表（来自 EvidenceStore）
            enable_alias_match: 是否启用别名匹配
            min_sentence_length: 最小句子长度阈值
        """
        self._evidences = evidences
        self._enable_alias_match = enable_alias_match
        self._min_sentence_length = min_sentence_length

        self._detector = SentenceBoundaryDetector()
        self._match_index: Dict[str, List[Tuple[str, int]]] = {}
        self._buffer = ""
        self._processed_offset = 0

        StandLogger.info_log(
            f"\n[EvidenceInjectProcessor.__init__] START\n"
            f"  evidences_count: {len(evidences)}\n"
            f"  enable_alias_match: {enable_alias_match}\n"
            f"  min_sentence_length: {min_sentence_length}"
        )

        if evidences:
            StandLogger.info_log("  Evidences:")
            for ev in evidences:
                StandLogger.info_log(
                    f"    - {ev.get('local_id')}: "
                    f"name={ev.get('object_type_name')}, "
                    f"aliases={ev.get('aliases', [])}"
                )
            self._build_match_index()
            StandLogger.info_log(
                f"  Match index built: {len(self._match_index)} entries\n"
            )
        else:
            StandLogger.info_log("  No evidences provided\n")

    async def process(
        self,
        stream: AsyncGenerator[Dict[str, Any], None],
    ) -> AsyncGenerator[Dict[str, Any], None]:
        """
        处理流式输出，包装并注入证据标注。

        Args:
            stream: 原始流式输出生成器 (yield dict with "answer" and "context")

        Yields:
            {
                "answer": 原始文本片段,
                "context": {...},
                "_evidence": {
                    "e1": {"object_type_name": "张三", "positions": [[4, 6]]}
                }
            }
        """
        if not self._evidences:
            async for item in stream:
                yield item
            return

        async for item in stream:
            answer = item.get("answer", "")
            context = item.get("context", {})

            if isinstance(answer, str):
                self._buffer += answer
            elif isinstance(answer, dict):
                answer_str = str(answer)
                self._buffer += answer_str
                answer = answer_str

            boundaries = self._detector.detect(self._buffer)

            if not boundaries:
                yield {
                    "answer": answer,
                    "context": context,
                }
                continue

            last_boundary = boundaries[-1]
            sentence_text = self._buffer[: last_boundary + 1]

            if len(sentence_text) >= self._min_sentence_length:
                annotated_text, evidence_meta = self._annotate_text(
                    sentence_text
                )

                remaining = self._buffer[last_boundary + 1 :]

                yield {
                    "answer": annotated_text,
                    "context": context,
                    "_evidence": evidence_meta if evidence_meta else None,
                }

                self._buffer = remaining
                self._processed_offset += len(sentence_text)
            else:
                yield {
                    "answer": answer,
                    "context": context,
                }

        if self._buffer.strip():
            _, evidence_meta = self._annotate_text(self._buffer)
            yield {
                "answer": self._buffer,
                "context": {},
                "_evidence": evidence_meta if evidence_meta else None,
            }

    def _build_match_index(self) -> None:
        """
        构建精确匹配和别名匹配索引。

        索引结构：{entity_name: [(local_id, priority)]}
        精确匹配优先级高于别名匹配。
        """
        for ev in self._evidences:
            local_id = ev.get("local_id", "")
            name = ev.get("object_type_name", "")

            if not name or not local_id:
                continue

            name_lower = name.lower()

            if name_lower not in self._match_index:
                self._match_index[name_lower] = []

            self._match_index[name_lower].append((local_id, 0))

            if self._enable_alias_match:
                aliases = ev.get("aliases") or []
                for alias in aliases:
                    if isinstance(alias, str) and alias.strip():
                        alias_lower = alias.lower().strip()
                        if alias_lower not in self._match_index:
                            self._match_index[alias_lower] = []
                        self._match_index[alias_lower].append(
                            (local_id, 1)
                        )

    def _annotate_text(
        self, text: str
    ) -> Tuple[str, Dict[str, Any]]:
        """
        对文本进行标注，返回原始文本和位置索引元数据。

        Args:
            text: 待标注的文本

        Returns:
            (原始文本, 位置索引元数据字典)
        """
        if not text or not self._match_index:
            return text, {}

        matches = self._match_annotations(text)

        if not matches:
            return text, {}

        resolved = self._resolve_overlaps(matches)

        StandLogger.info_log(
            f"[_annotate_text] text_len={len(text)}, "
            f"raw_matches={len(matches)}, "
            f"resolved={len(resolved)}"
        )

        evidence_meta: Dict[str, Any] = {}

        for local_id, start, end in resolved:
            ev = self._find_evidence_by_local_id(local_id)
            if ev is None:
                StandLogger.info_log(f"  Warning: local_id={local_id} not found in evidences")
                continue

            name = ev.get("object_type_name", "")

            if local_id not in evidence_meta:
                evidence_meta[local_id] = {
                    "object_type_name": name,
                    "positions": [],
                }

            evidence_meta[local_id]["positions"].append([start, end])

            matched_text = text[start:end] if end <= len(text) else ""
            StandLogger.info_log(
                f"  + {local_id} ({name}): [{start}, {end}] = '{matched_text}'"
            )

        StandLogger.info_log(
            f"[_annotate_text] Result: {len(evidence_meta)} entities annotated\n"
        )

        return text, evidence_meta

    def _match_annotations(
        self, text: str
    ) -> List[Tuple[str, int, int]]:
        """
        匹配文本中的所有实体。

        Returns:
            [(local_id, start, end), ...] 按位置排序
        """
        matches = []
        text_lower = text.lower()

        for entity_key, entries in self._match_index.items():
            entity_key_len = len(entity_key)
            if entity_key_len == 0:
                continue

            start = 0
            while True:
                pos = text_lower.find(entity_key, start)
                if pos == -1:
                    break

                for local_id, priority in entries:
                    matches.append(
                        (local_id, pos, pos + entity_key_len)
                    )

                start = pos + 1

        matches.sort(key=lambda m: (m[1], m[2] - m[1]))
        return matches

    def _resolve_overlaps(
        self, matches: List[Tuple[str, int, int]]
    ) -> List[Tuple[str, int, int]]:
        """
        解决重叠匹配冲突。

        策略：
        - 保留更长的匹配（精确匹配优先于短别名）
        - 同长度时保留先出现的
        - 已使用的位置不再重复标注
        """
        if not matches:
            return []

        used_positions: List[Tuple[int, int]] = []
        resolved = []

        for local_id, start, end in matches:
            if self._is_position_overlapping(
                start, end, used_positions
            ):
                continue

            resolved.append((local_id, start, end))
            used_positions.append((start, end))

        return resolved

    def _is_position_overlapping(
        self,
        start: int,
        end: int,
        used_positions: List[Tuple[int, int]],
    ) -> bool:
        """检查位置是否与已使用位置重叠"""
        for used_start, used_end in used_positions:
            if not (end <= used_start or start >= used_end):
                return True
        return False

    def _find_evidence_by_local_id(
        self, local_id: str
    ) -> Optional[Dict[str, Any]]:
        """根据 local_id 查找证据"""
        for ev in self._evidences:
            if ev.get("local_id") == local_id:
                return ev
        return None


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
        Config.features, "annotate_only_final_answer", True
    )

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[create_evidence_injection_stream] START\n"
        f"  Initial evidence_store_key: {evidence_store_key}\n"
        f"  annotate_only_final_answer: {annotate_only_final}\n"
        f"{'='*60}\n"
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
        StandLogger.info_log(
            "[create_evidence_injection_stream] annotate_only_final=True, "
            "caching last LLM item for annotation"
        )

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
                StandLogger.info_log(
                    f"[create_evidence_injection_stream] Cached LLM item, "
                    f"tool_results_count={len(current_tool_results)}"
                )
            else:
                # 没有 LLM stage，直接发送
                StandLogger.info_log(
                    "[create_evidence_injection_stream] No LLM stage in item, "
                    "passing through"
                )

            # 直接发送当前 item（不进行标注）
            yield item

        # 流结束后，对缓存的 LLM item 进行标注
        if last_llm_item is not None:
            StandLogger.info_log(
                "[create_evidence_injection_stream] Stream ended, "
                "annotating cached LLM item"
            )

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
        else:
            StandLogger.info_log(
                "[create_evidence_injection_stream] Stream ended, "
                "no LLM item found to annotate"
            )


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

    StandLogger.info_log(
        f"[_process_single_item] Processing item: "
        f"item_evidence_key={item_evidence_key}, "
        f"_loaded_key={_loaded_key}, "
        f"is_final_answer={is_final_answer}"
    )

    # 如果有新的 evidence_store_key（或者还没有加载过，或者当前没有工具结果），从 EvidenceStore 获取原始工具结果
    if item_evidence_key and (item_evidence_key != _loaded_key or not current_tool_results):
        raw_results_key = f"{item_evidence_key}_raw"
        StandLogger.info_log(
            f"[_process_single_item] Attempting to load tool results: "
            f"raw_results_key={raw_results_key}"
        )
        tool_results = store.get(raw_results_key)
        StandLogger.info_log(
            f"[_process_single_item] store.get() returned: "
            f"type={type(tool_results).__name__}, "
            f"count={len(tool_results) if tool_results else 0}"
        )
        if tool_results:
            StandLogger.info_log(
                f"[_process_single_item] ✅ Loaded {len(tool_results)} raw tool results\n"
                f"  Tools: {[tr.get('tool_name', '?') for tr in tool_results[:5]]}{'...' if len(tool_results) > 5 else ''}"
            )
            current_tool_results = tool_results
            _loaded_key = item_evidence_key
        else:
            StandLogger.info_log(
                f"[_process_single_item] ℹ️ No raw tool results found for key={raw_results_key}"
            )
            current_tool_results = []
            _loaded_key = item_evidence_key

    # 提取 LLM 生成的文本
    answer = item.get("answer", {})
    actual_text = None

    StandLogger.info_log(
        f"[_process_single_item] Extracting text: "
        f"answer_type={type(answer).__name__}, "
        f"has_answer_key={'answer' in answer if isinstance(answer, dict) else 'N/A'}"
    )

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

    StandLogger.info_log(
        f"[_process_single_item] text_len={text_len}, "
        f"tool_results_count={len(current_tool_results)}, "
        f"is_final_answer={is_final_answer}, "
        f"annotate_only_final={annotate_only_final}"
    )

    # 判断是否需要进行 LLM 标注
    # 条件：有工具结果 + 有足够长的文本 + (是最终答案 或 不限制仅标注最终答案)
    should_annotate = (
        current_tool_results and
        text_len > 5 and
        (is_final_answer or not annotate_only_final)
    )

    if should_annotate:
        StandLogger.info_log(
            f"[_process_single_item] 🔄 Performing LLM annotation, "
            f"text_preview={actual_text[:100]}..."
        )

        try:
            from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
                llm_annotate_evidence,
            )

            annotation_config = {
                "llm_annotation_timeout": getattr(
                    Config.features, "llm_annotation_timeout", 30
                ),
                "llm_annotation_model": getattr(
                    Config.features, "llm_annotation_model", "deepseek-v3.2"
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

            StandLogger.info_log(
                f"[_process_single_item] Building evidence_meta: "
                f"evidences_count={len(evidences)}, "
                f"tool_results_count={len(current_tool_results)}"
            )

            # 获取第一个（或唯一的）工具结果
            # TODO: 如果有多个工具结果，需要更智能的匹配逻辑
            tool_data = current_tool_results[0] if current_tool_results else None

            if tool_data:
                StandLogger.info_log(
                    f"[_process_single_item] Using tool_data: "
                    f"tool_name={tool_data.get('tool_name', '')}, "
                    f"result_type={tool_data.get('result_type', '')}"
                )

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

                    StandLogger.info_log(
                        f"[_process_single_item] Evidence {local_id}: "
                        f"object_type_name={object_type}, "
                        f"tool_name={evidence_item.get('tool_name', '')}, "
                        f"has_result={bool(evidence_item.get('result'))}"
                    )

                evidence_meta[local_id] = evidence_item

            if evidence_meta:
                StandLogger.info_log(
                    f"[_process_single_item] ✅ LLM annotated {len(evidence_meta)} evidences"
                )
            else:
                StandLogger.info_log(
                    f"[_process_single_item] ℹ️ LLM found no evidence in text"
                )

            # 注入到 _progress
            _inject_evidence_to_progress(item, evidence_meta)

        except Exception as e:
            StandLogger.info_log(
                f"[_process_single_item] ❌ Annotation failed: {e}"
            )
    else:
        # 跳过 LLM 标注，直接返回原 item
        if annotate_only_final and not is_final_answer:
            StandLogger.info_log(
                f"[_process_single_item] ⏭️ Skipping LLM annotation (intermediate item, annotate_only_final=True)"
            )
        else:
            StandLogger.info_log(
                f"[_process_single_item] ⏭️ Skipping LLM annotation "
                f"(no text={not actual_text}, no_tools={not current_tool_results})"
            )

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
    StandLogger.info_log(
        f"[_inject_evidence_to_progress] Injecting evidence: "
        f"evidence_count={len(evidence_meta)}"
    )

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
                StandLogger.info_log(
                    f"[_inject_evidence_to_progress] ✅ Injected into item['answer']['_progress'][{llm_progress_index}]: "
                    f"stage={llm_progress.get('stage')}, "
                    f"evidence_count={len(evidence_meta)}, "
                    f"answer_preview={str(llm_progress.get('answer', ''))[:50]}..."
                )

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
                                StandLogger.info_log(
                                    f"[_inject_evidence_to_progress] ✅ Injected into item['_progress'][{idx}]"
                                )
                                break
            else:
                StandLogger.info_log(
                    f"[_inject_evidence_to_progress] ⚠️ No LLM stage found in progress array "
                    f"(total={len(progress_array)}, stages={[p.get('stage') for p in progress_array[-5:]]})"
                )
        else:
            StandLogger.info_log(
                f"[_inject_evidence_to_progress] ⚠️ Progress array is empty or not a list"
            )
    else:
        StandLogger.info_log(
            f"[_inject_evidence_to_progress] ⚠️ answer is not a dict or missing _progress"
        )
