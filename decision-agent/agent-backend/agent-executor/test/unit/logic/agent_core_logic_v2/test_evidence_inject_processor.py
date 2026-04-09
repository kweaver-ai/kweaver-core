# -*- coding: utf-8 -*-
"""单元测试 - evidence_inject_processor 模块"""

import pytest
from typing import AsyncGenerator, Dict, Any


class TestSentenceBoundaryDetector:
    """测试句子边界检测器"""

    def test_chinese_punctuation(self):
        """中文标点检测"""
        from app.common.utils.sentence_boundary_detector import (
            SentenceBoundaryDetector,
        )

        detector = SentenceBoundaryDetector()
        text = "你好世界。这是第二句！第三句？结束"
        boundaries = detector.detect(text)

        assert len(boundaries) >= 3

    def test_english_punctuation(self):
        """英文标点检测"""
        from app.common.utils.sentence_boundary_detector import (
            SentenceBoundaryDetector,
        )

        detector = SentenceBoundaryDetector()
        text = "Hello world. This is second! Third?"
        boundaries = detector.detect(text)

        assert len(boundaries) >= 2

    def test_empty_text(self):
        """空文本返回空列表"""
        from app.common.utils.sentence_boundary_detector import (
            SentenceBoundaryDetector,
        )

        detector = SentenceBoundaryDetector()
        assert detector.detect("") == []
        assert detector.detect(None) == []

    def test_split_to_sentences(self):
        """分割为句子"""
        from app.common.utils.sentence_boundary_detector import (
            SentenceBoundaryDetector,
        )

        detector = SentenceBoundaryDetector()
        text = "第一句。第二句。第三句"
        sentences = detector.split_to_sentences(text)

        assert len(sentences) >= 2


class TestEvidenceInjectProcessorBasic:
    """测试 EvidenceInjectProcessor 基本功能"""

    def _make_evidences(self):
        return [
            {
                "local_id": "e1",
                "object_type_name": "张三",
                "aliases": ["员工张三", "张工"],
                "object_type_id": "ot_employee",
                "score": 0.95,
            },
            {
                "local_id": "e2",
                "object_type_name": "上海",
                "aliases": [],
                "object_type_id": "ot_city",
                "score": 0.9,
            },
        ]

    def test_exact_match(self):
        """精确匹配实体名称"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(
            evidences=self._make_evidences(),
            enable_alias_match=False,
        )
        text, meta = processor._annotate_text("我的名字是张三，我住上海。")

        assert text == "我的名字是张三，我住上海。"
        assert "e1" in meta or "e2" in meta

    def test_alias_match(self):
        """别名匹配"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(
            evidences=self._make_evidences(),
            enable_alias_match=True,
        )
        text, meta = processor._annotate_text("这是员工张三的信息。")

        assert "e1" in meta
        assert "positions" in meta["e1"]

    def test_no_evidences_passthrough(self):
        """无证据时直接透传"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(evidences=[], enable_alias_match=True)
        text, meta = processor._annotate_text("这是一段普通文本。")

        assert text == "这是一段普通文本。"
        assert meta == {}

    def test_no_matches_returns_empty_meta(self):
        """无匹配返回空元数据"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(
            evidences=self._make_evidences(),
        )
        text, meta = processor._annotate_text("今天天气不错。")

        assert meta == {}


class TestMatchIndex:
    """测试匹配索引构建"""

    def test_build_exact_and_alias_index(self):
        """构建精确和别名索引"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        evidences = [
            {
                "local_id": "e1",
                "object_type_name": "张三",
                "aliases": ["小张"],
            }
        ]

        processor = EvidenceInjectProcessor(evidences=evidences)
        assert "张三" in processor._match_index
        assert "小张" in processor._match_index

    def test_empty_evidences_no_index(self):
        """空证据不构建索引"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(evidences=[])
        assert processor._match_index == {}


class TestOverlapResolution:
    """测试重叠匹配解决"""

    def test_overlapping_matches_resolved(self):
        """重叠匹配被正确解决"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        evidences = [
            {"local_id": "e1", "object_type_name": "张三", "aliases": []},
            {"local_id": "e2", "object_type_name": "张三丰", "aliases": []},
        ]

        processor = EvidenceInjectProcessor(evidences=evidences)
        text, meta = processor._annotate_text("张三丰是一位武术家。")

        total_positions = sum(
            len(v.get("positions", [])) for v in meta.values()
        )
        assert total_positions == 1


class TestPositionTracking:
    """测试位置标记管理"""

    def test_position_not_overlapping(self):
        """位置不重叠检查"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(evidences=[])

        used = [(0, 5), (10, 15)]

        assert not processor._is_position_overlapping(5, 10, used)
        assert processor._is_position_overlapping(3, 7, used)
        assert processor._is_position_overlapping(12, 18, used)


class TestStreamProcessing:
    """测试流式处理模拟"""

    @pytest.mark.asyncio
    async def test_stream_with_evidence(self):
        """带证据的流式处理"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        evidences = [
            {
                "local_id": "e1",
                "object_type_name": "张三",
                "aliases": [],
            }
        ]

        async def mock_stream() -> AsyncGenerator[Dict[str, Any], None]:
            yield {"answer": "你好，", "context": {}}
            yield {"answer": "我叫张三。", "context": {}}

        processor = EvidenceInjectProcessor(
            evidences=evidences,
            min_sentence_length=1,
        )

        results = []
        async for item in processor.process(mock_stream()):
            results.append(item)

        assert len(results) >= 1

    @pytest.mark.asyncio
    async def test_empty_stream(self):
        """空流式输入"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        evidences = [
            {"local_id": "e1", "object_type_name": "张三", "aliases": []}
        ]

        async def mock_stream() -> AsyncGenerator[Dict[str, Any], None]:
            if False:
                yield {}

        processor = EvidenceInjectProcessor(evidences=evidences)
        results = []
        async for item in processor.process(mock_stream()):
            results.append(item)

        assert results == []

    @pytest.mark.asyncio
    async def test_none_evidences_passthrough(self):
        """无证据时直接透传流"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        async def mock_stream() -> AsyncGenerator[Dict[str, Any], None]:
            yield {"answer": "普通文本", "context": {}}
            yield {"answer": "继续输出", "context": {}}

        processor = EvidenceInjectProcessor(evidences=[])
        results = []
        async for item in processor.process(mock_stream()):
            results.append(item)

        assert len(results) == 2
        assert "_evidence" not in results[0]


class TestFindEvidenceByLocalId:
    """测试按 local_id 查找证据"""

    def test_find_existing(self):
        """查找存在的证据"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        ev = [
            {"local_id": "e1", "object_type_name": "A"},
            {"local_id": "e2", "object_type_name": "B"},
        ]
        processor = EvidenceInjectProcessor(evidences=ev)

        result = processor._find_evidence_by_local_id("e1")
        assert result is not None
        assert result["object_type_name"] == "A"

    def test_find_nonexistent(self):
        """查找不存在的证据"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            EvidenceInjectProcessor,
        )

        processor = EvidenceInjectProcessor(
            evidences=[{"local_id": "e1", "object_type_name": "A"}]
        )
        assert processor._find_evidence_by_local_id("e999") is None


class TestCreateEvidenceInjectionStream:
    """测试 create_evidence_injection_stream 工厂函数"""

    @pytest.mark.asyncio
    async def test_none_key_passthrough(self):
        """None key 直接透传"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            create_evidence_injection_stream,
        )

        async def mock_stream():
            yield {"answer": "test", "context": {}}

        results = []
        async for item in create_evidence_injection_stream(mock_stream(), None):
            results.append(item)

        assert len(results) == 1
        assert "_evidence" not in results[0]

    @pytest.mark.asyncio
    async def test_empty_string_key_passthrough(self):
        """空字符串 key 直接透传"""
        from app.logic.agent_core_logic_v2.evidence_inject_processor import (
            create_evidence_injection_stream as real_factory,
        )

        async def mock_stream():
            yield {"answer": "test", "context": {}}

        results = []
        async for item in real_factory(mock_stream(), ""):
            results.append(item)

        assert len(results) == 1
