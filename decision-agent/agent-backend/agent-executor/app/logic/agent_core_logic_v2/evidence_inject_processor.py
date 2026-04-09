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
    current_processor = None
    current_evidences = []

    async for item in original_stream:
        # 从当前 item 中获取 evidence_store_key
        item_evidence_key = item.get("evidence_store_key")

        # 如果有新的 evidence_store_key，从 EvidenceStore 获取证据
        if item_evidence_key and item_evidence_key != getattr(
            create_evidence_injection_stream, "_last_key", None
        ):
            evidences = store.get(item_evidence_key)
            if evidences:
                StandLogger.info_log(
                    f"[EvidenceInject] ✅ Loaded {len(evidences)} evidences for key={item_evidence_key}\n"
                    f"  Evidences: {[e.get('object_type_name', e.get('local_id', '?')) for e in evidences[:5]]}{'...' if len(evidences) > 5 else ''}"
                )
                current_evidences = evidences
                enable_alias = getattr(
                    Config.features, "enable_alias_match", True
                )
                min_sent_len = getattr(
                    Config.features, "min_sentence_length", 10
                )
                current_processor = EvidenceInjectProcessor(
                    evidences=current_evidences,
                    enable_alias_match=enable_alias,
                    min_sentence_length=min_sent_len,
                )
            else:
                StandLogger.info_log(
                    f"[EvidenceInject] ❌ No evidences found for key={item_evidence_key}"
                )
                current_processor = None
            create_evidence_injection_stream._last_key = item_evidence_key

        # 检查是否应该处理此项目
        answer = item.get("answer", {})
        actual_text = None

        # answer 通常是字典，需要从中提取实际文本
        if isinstance(answer, dict):
            actual_text = answer.get("answer", "")
        elif isinstance(answer, str):
            actual_text = answer

        text_len = len(actual_text) if isinstance(actual_text, str) else 0
        should_process = (
            current_processor and
            isinstance(actual_text, str) and
            text_len > 20  # 只处理有实际文本内容的项目
        )

        StandLogger.info_log(
            f"[create_evidence_injection_stream] item: "
            f"evidence_store_key={item_evidence_key}, "
            f"answer_type={type(answer).__name__}, "
            f"actual_text_type={type(actual_text).__name__}, "
            f"text_len={text_len}, "
            f"should_process={should_process}"
        )

        # 如果有处理器且应该处理，则进行注入；否则直接透传
        if should_process:
            StandLogger.info_log(
                f"[EvidenceInject] 🔄 Processing with EvidenceInjectProcessor, text={actual_text[:50]}..."
            )
            # 调用处理器进行标注
            _, evidence_meta = current_processor._annotate_text(actual_text)

            if evidence_meta:
                # 将 _evidence 元数据添加到 answer 字典中
                if isinstance(answer, dict):
                    answer["_evidence"] = evidence_meta
                else:
                    # 如果 answer 不是字典，创建新结构
                    item["answer"] = {"answer": answer, "_evidence": evidence_meta}

                StandLogger.info_log(
                    f"[EvidenceInject] ✅ Injected evidence: {evidence_meta}"
                )
            else:
                StandLogger.info_log(
                    f"[EvidenceInject] ℹ️ No evidence found in text"
                )

            yield item
        else:
            yield item
