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

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[create_evidence_injection_stream] START\n"
        f"  Initial evidence_store_key: {evidence_store_key}\n"
        f"{'='*60}\n"
    )

    store = get_global_evidence_store()
    current_tool_results = []  # 存储当前的工具结果
    _loaded_key = None  # 记录已加载的 key

    async for item in original_stream:
        # 从当前 item 中获取 evidence_store_key
        item_evidence_key = item.get("evidence_store_key")

        StandLogger.info_log(
            f"[create_evidence_injection_stream] Processing item: "
            f"item_evidence_key={item_evidence_key}, "
            f"_loaded_key={_loaded_key}"
        )

        # 如果有新的 evidence_store_key（或者还没有加载过，或者当前没有工具结果），从 EvidenceStore 获取原始工具结果
        if item_evidence_key and (item_evidence_key != _loaded_key or not current_tool_results):
            raw_results_key = f"{item_evidence_key}_raw"
            StandLogger.info_log(
                f"[EvidenceInject] Attempting to load tool results: "
                f"raw_results_key={raw_results_key}, "
                f"_loaded_key={_loaded_key}"
            )
            tool_results = store.get(raw_results_key)
            StandLogger.info_log(
                f"[EvidenceInject] store.get() returned: "
                f"type={type(tool_results).__name__}, "
                f"value={tool_results if tool_results else 'None'}"
            )
            if tool_results:
                StandLogger.info_log(
                    f"[EvidenceInject] ✅ Loaded {len(tool_results)} raw tool results for key={raw_results_key}\n"
                    f"  Tools: {[tr.get('tool_name', '?') for tr in tool_results[:5]]}{'...' if len(tool_results) > 5 else ''}"
                )
                current_tool_results = tool_results
                _loaded_key = item_evidence_key
            else:
                StandLogger.info_log(
                    f"[EvidenceInject] ℹ️ No raw tool results found for key={raw_results_key}"
                )
                current_tool_results = []
                _loaded_key = item_evidence_key
        elif item_evidence_key:
            StandLogger.info_log(
                f"[EvidenceInject] Skipping load: item_evidence_key={item_evidence_key}, "
                f"_loaded_key={_loaded_key}, keys are equal, "
                f"has_tool_results={len(current_tool_results) > 0}"
            )

        # 提取 LLM 生成的文本
        answer = item.get("answer", {})
        actual_text = None

        StandLogger.info_log(
            f"[create_evidence_injection_stream] Extracting text: "
            f"answer_type={type(answer).__name__}, "
            f"has_answer_key={'answer' in answer if isinstance(answer, dict) else 'N/A'}"
        )

        if isinstance(answer, dict):
            candidate = answer.get("answer", "")
            StandLogger.info_log(
                f"[create_evidence_injection_stream] candidate_type={type(candidate).__name__}, "
                f"candidate_preview={str(candidate)[:100] if candidate else 'None'}..."
            )

            if isinstance(candidate, str):
                # answer["answer"] 直接是字符串
                actual_text = candidate
                StandLogger.info_log(
                    f"[create_evidence_injection_stream] ✓ Got text from answer['answer'] (string), len={len(actual_text)}"
                )
            elif isinstance(candidate, dict):
                # answer["answer"] 是字典，可能包含嵌套的 answer 字段
                # 检查是否是 LLM stage 的数据结构
                candidate_stage = candidate.get("stage", "")
                candidate_answer = candidate.get("answer")

                StandLogger.info_log(
                    f"[create_evidence_injection_stream] candidate dict: stage={candidate_stage}, "
                    f"has_answer={candidate_answer is not None}, "
                    f"answer_type={type(candidate_answer).__name__ if candidate_answer is not None else 'None'}"
                )

                # 如果 candidate 本身就是 LLM stage 的数据，直接提取 answer
                if candidate_stage == "llm" and isinstance(candidate_answer, str):
                    actual_text = candidate_answer
                    StandLogger.info_log(
                        f"[create_evidence_injection_stream] ✓ Got text from candidate['answer'] (LLM stage), len={len(actual_text)}"
                    )
                else:
                    # 尝试从 _progress 获取最新的 LLM stage answer
                    progress_array = item.get("_progress", [])
                    StandLogger.info_log(
                        f"[create_evidence_injection_stream] candidate not LLM stage, checking _progress, "
                        f"progress_count={len(progress_array)}"
                    )
                    if progress_array:
                        for idx, p in enumerate(reversed(progress_array)):
                            stage = p.get("stage", "")
                            p_answer = p.get("answer")
                            StandLogger.info_log(
                                f"[create_evidence_injection_stream] progress[{len(progress_array)-1-idx}]: "
                                f"stage={stage}, answer_type={type(p_answer).__name__}, "
                                f"answer_len={len(p_answer) if isinstance(p_answer, str) else 'N/A'}"
                            )
                            if stage == "llm" and isinstance(p_answer, str):
                                actual_text = p_answer
                                StandLogger.info_log(
                                    f"[create_evidence_injection_stream] ✓ Got text from _progress LLM stage, len={len(actual_text)}"
                                )
                                break
        elif isinstance(answer, str):
            actual_text = answer
            StandLogger.info_log(
                f"[create_evidence_injection_stream] ✓ Got text from answer (direct string), len={len(actual_text)}"
            )

        if not isinstance(actual_text, str):
            actual_text = ""
            StandLogger.info_log(
                f"[create_evidence_injection_stream] ⚠️ No text found, using empty string"
            )

        text_len = len(actual_text)

        StandLogger.info_log(
            f"[create_evidence_injection_stream] Final: text_len={text_len}, "
            f"text_preview={actual_text[:100] if actual_text else 'EMPTY'}..."
        )

        # 判断是否应该处理：有工具结果 + 有足够长的文本
        # 阈值设为 5，允许对短文本进行证据标注
        should_process = (
            current_tool_results and
            text_len > 5
        )

        StandLogger.info_log(
            f"[create_evidence_injection_stream] item: "
            f"evidence_store_key={item_evidence_key}, "
            f"text_len={text_len}, "
            f"tool_results_count={len(current_tool_results)}, "
            f"should_process={should_process}"
        )

        if should_process:
            StandLogger.info_log(
                f"[EvidenceInject] 🔄 Using LLM to annotate evidence, text_preview={actual_text[:100]}..."
            )

            try:
                # 调用 LLM 进行标注
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
                # annotation_result: {"evidences": [{"object_type_name": ..., "positions": ...}]}
                # evidence_meta: {"e1": {"object_type_name": ..., "positions": ...}}
                evidences = annotation_result.get("evidences", [])
                evidence_meta = {}

                for idx, ev in enumerate(evidences):
                    local_id = f"e{idx + 1}"
                    evidence_meta[local_id] = {
                        "object_type_name": ev.get("object_type_name", ""),
                        "positions": ev.get("positions", []),
                    }

                if evidence_meta:
                    StandLogger.info_log(
                        f"[EvidenceInject] ✅ LLM annotated {len(evidence_meta)} evidences"
                    )
                else:
                    StandLogger.info_log(
                        f"[EvidenceInject] ℹ️ LLM found no evidence in text"
                    )

                # 无论是否提取到证据，都添加 _evidence 字段（空字典作为占位）
                # 只对 LLM stage 的 progress 条目添加
                # _progress 在两个地方：
                # 1. item["answer"]["_progress"] - answer 字典中的进度数组
                # 2. item["_progress"] - 顶层进度数组（可能是同一引用）

                # 首先注入到 item["answer"]["_progress"]
                answer_dict = item.get("answer")
                StandLogger.info_log(
                    f"[EvidenceInject] Checking injection: "
                    f"answer_type={type(answer_dict).__name__}, "
                    f"answer_is_dict={isinstance(answer_dict, dict)}, "
                    f"item_has_progress={'_progress' in item}"
                )

                if isinstance(answer_dict, dict):
                    StandLogger.info_log(
                        f"[EvidenceInject] answer keys={list(answer_dict.keys())}, "
                        f"has_progress={'_progress' in answer_dict}"
                    )

                # 注入到 item["answer"]["_progress"]
                if isinstance(answer_dict, dict) and "_progress" in answer_dict:
                    progress_array = answer_dict["_progress"]
                    StandLogger.info_log(
                        f"[EvidenceInject] progress_type={type(progress_array).__name__}, "
                        f"is_list={isinstance(progress_array, list)}, "
                        f"progress_len={len(progress_array) if isinstance(progress_array, list) else 'N/A'}"
                    )

                    if isinstance(progress_array, list) and progress_array:
                        latest_progress = progress_array[-1]
                        StandLogger.info_log(
                            f"[EvidenceInject] latest_progress stage={latest_progress.get('stage')}, "
                            f"is_llm={latest_progress.get('stage') == 'llm'}"
                        )
                        # 只对 LLM stage 添加 _evidence
                        if latest_progress.get("stage") == "llm":
                            latest_progress["_evidence"] = evidence_meta
                            StandLogger.info_log(
                                f"[EvidenceInject] ✅ Injected _evidence into item['answer']['_progress'][-1]: "
                                f"stage={latest_progress.get('stage')}, "
                                f"evidence_count={len(evidence_meta)}, "
                                f"is_placeholder={len(evidence_meta) == 0}"
                            )

                            # 验证注入是否成功
                            StandLogger.info_log(
                                f"[EvidenceInject] Verification (answer['_progress'][-1]): "
                                f"has _evidence: {'_evidence' in latest_progress}, "
                                f"keys={list(latest_progress.keys())}"
                            )

                            # 同时注入到 item["_progress"]（如果存在且是不同的对象）
                            if "_progress" in item and isinstance(item["_progress"], list):
                                top_progress_array = item["_progress"]
                                # 找到对应的 LLM stage 条目（通常是最后一个）
                                if top_progress_array and len(top_progress_array) > 0:
                                    # 检查是否是同一个对象
                                    if top_progress_array[-1] is latest_progress:
                                        StandLogger.info_log(
                                            f"[EvidenceInject] item['_progress'][-1] is same object as answer['_progress'][-1], "
                                            f"no need to inject again"
                                        )
                                    else:
                                        # 找到对应的 LLM stage 条目
                                        for idx, p in enumerate(top_progress_array):
                                            if (p.get("stage") == "llm" and
                                                p.get("id") == latest_progress.get("id")):
                                                p["_evidence"] = evidence_meta
                                                StandLogger.info_log(
                                                    f"[EvidenceInject] ✅ Injected _evidence into item['_progress'][{idx}]: "
                                                    f"stage={p.get('stage')}, "
                                                    f"evidence_count={len(evidence_meta)}"
                                                )
                                                break
                        else:
                            StandLogger.info_log(
                                f"[EvidenceInject] ⚠️ Latest progress is not LLM stage: {latest_progress.get('stage')}"
                            )
                    else:
                        StandLogger.info_log(
                            f"[EvidenceInject] ⚠️ Progress array is empty or not a list"
                        )
                else:
                    StandLogger.info_log(
                        f"[EvidenceInject] ⚠️ answer is not a dict or missing _progress"
                    )

            except Exception as e:
                StandLogger.info_log(
                    f"[EvidenceInject] ❌ Annotation failed: {e}"
                )
                logger.error(f"[EvidenceInject] Annotation failed: {e}", exc_info=True)

        yield item
