# -*- coding: utf-8 -*-
"""单元测试 - evidence_prepare 模块"""

import pytest
from unittest import mock


class TestEvidencePrepare:
    """测试 evidence_prepare 主函数"""

    @pytest.mark.asyncio
    async def test_basic_extraction_with_nodes(self):
        """测试包含 nodes 字段的基本提取"""
        from app.logic.tool.evidence_prepare import evidence_prepare

        tool_call_result = {
            "nodes": [
                {
                    "object_type_name": "张三",
                    "object_type_id": "ot_employee",
                    "score": 0.95,
                },
                {
                    "object_type_name": "上海",
                    "object_type_id": "ot_city",
                    "score": 0.9,
                },
            ],
            "answer": "查询到员工张三的信息",
        }

        with mock.patch(
            "app.logic.tool.evidence_prepare._extract_by_llm",
            return_value=[],
        ):
            result = await evidence_prepare(tool_call_result)

        assert result is not None
        assert "evidence_id" in result
        assert result["evidence_id"].startswith("ev_")
        assert len(result["evidences"]) >= 1
        assert "summary" in result

    @pytest.mark.asyncio
    async def test_empty_result(self):
        """测试空输入"""
        from app.logic.tool.evidence_prepare import evidence_prepare

        with mock.patch(
            "app.logic.tool.evidence_prepare._extract_by_llm",
            return_value=[],
        ):
            result = await evidence_prepare({})

        assert result is not None
        assert result["evidence_id"].startswith("ev_")
        assert len(result["evidences"]) == 0
        assert "无证据" in result["summary"]

    @pytest.mark.asyncio
    async def test_non_dict_input(self):
        """测试非字典输入"""
        from app.logic.tool.evidence_prepare import evidence_prepare

        with mock.patch(
            "app.logic.tool.evidence_prepare._extract_by_llm",
            return_value=[],
        ):
            result = await evidence_prepare("invalid")

        assert result is not None

    @pytest.mark.asyncio
    async def test_returns_required_fields(self):
        """验证返回值包含所有必需字段"""
        from app.logic.tool.evidence_prepare import evidence_prepare

        tool_call_result = {
            "nodes": [
                {"object_type_name": "测试实体", "score": 0.8}
            ]
        }

        with mock.patch(
            "app.logic.tool.evidence_prepare._extract_by_llm",
            return_value=[],
        ):
            result = await evidence_prepare(tool_call_result)

        assert "evidence_id" in result
        assert "evidences" in result
        assert "summary" in result
        assert isinstance(result["evidences"], list)


class TestProcessWithExistingEvidence:
    """测试 _process_with_existing_evidence 函数"""

    def test_extract_from_nodes_field(self):
        """从 nodes 字段提取已有 _evidence 结构"""
        from app.logic.tool.evidence_prepare import (
            _process_with_existing_evidence,
        )

        data = {
            "nodes": [
                {
                    "object_type_name": "张三",
                    "object_type_id": "ot_employee",
                    "score": 0.95,
                }
            ]
        }

        results = _process_with_existing_evidence(data)
        assert len(results) == 1
        assert results[0]["object_type_name"] == "张三"
        assert results[0]["_source"] == "existing_evidence"

    def test_no_nodes_returns_empty(self):
        """没有 nodes 时返回空列表"""
        from app.logic.tool.evidence_prepare import (
            _process_with_existing_evidence,
        )

        data = {"answer": "some text"}
        results = _process_with_existing_evidence(data)
        assert results == []

    def test_non_dict_input(self):
        """非字典输入返回空列表"""
        from app.logic.tool.evidence_prepare import (
            _process_with_existing_evidence,
        )

        assert _process_with_existing_evidence("string") == []
        assert _process_with_existing_evidence(None) == []
        assert _process_with_existing_evidence([]) == []


class TestParseLlmResult:
    """测试 _parse_llm_result 函数"""

    def test_valid_json_array(self):
        """解析有效的 JSON 数组"""
        from app.logic.tool.evidence_prepare import _parse_llm_result

        json_str = '[{"object_type_name": "张三", "aliases": ["员工张三"], "object_type_id": "ot_employee", "score": 0.95}]'
        results = _parse_llm_result(json_str)

        assert len(results) == 1
        assert results[0]["object_type_name"] == "张三"
        assert results[0]["aliases"] == ["员工张三"]
        assert results[0]["_source"] == "llm_extraction"

    def test_json_in_text(self):
        """从包含其他文本的字符串中提取 JSON"""
        from app.logic.tool.evidence_prepare import _parse_llm_result

        text = '根据分析，结果如下：\n[{"object_type_name": "上海", "score": 0.8}]\n以上是全部内容。'
        results = _parse_llm_result(text)

        assert len(results) == 1
        assert results[0]["object_type_name"] == "上海"

    def test_empty_input(self):
        """空输入返回空列表"""
        from app.logic.tool.evidence_prepare import _parse_llm_result

        assert _parse_llm_result("") == []
        assert _parse_llm_result(None) == []

    def test_invalid_json(self):
        """无效 JSON 返回空列表"""
        from app.logic.tool.evidence_prepare import _parse_llm_result

        assert _parse_llm_result("not json at all") == []
        assert _parse_llm_result("{invalid}") == []

    def test_missing_object_type_name_filtered(self):
        """缺少 object_type_name 的条目被过滤"""
        from app.logic.tool.evidence_prepare import _parse_llm_result

        json_str = '[{"name": "test"}, {"object_type_name": "有效实体"}]'
        results = _parse_llm_result(json_str)

        assert len(results) == 1
        assert results[0]["object_type_name"] == "有效实体"


class TestFallbackExtraction:
    """测试 _fallback_extraction 函数"""

    def test_extract_from_nodes(self):
        """从 nodes 字段回退提取"""
        from app.logic.tool.evidence_prepare import _fallback_extraction

        data = {
            "nodes": [
                {"object_type_name": "张三", "object_type_id": "ot_employee", "score": 0.9},
                {"object_type_name": "李四", "object_type_id": "ot_employee", "score": 0.85},
            ]
        }

        results = _fallback_extraction(data)
        assert len(results) == 2
        assert all(r["_source"] == "fallback_rules" for r in results)

    def test_extract_from_answer_nodes(self):
        """从 answer.nodes 字段回退提取"""
        from app.logic.tool.evidence_prepare import _fallback_extraction

        data = {
            "answer": {
                "nodes": [
                    {"object_type_name": "北京", "object_type_id": "ot_city"}
                ]
            }
        }

        results = _fallback_extraction(data)
        assert len(results) == 1
        assert results[0]["object_type_name"] == "北京"

    def test_non_dict_input(self):
        """非字典输入返回空列表"""
        from app.logic.tool.evidence_prepare import _fallback_extraction

        assert _fallback_extraction("string") == []
        assert _fallback_extraction([]) == []


class TestGenerateSummary:
    """测试 _generate_summary 函数"""

    def test_single_evidence(self):
        """单个证据的摘要"""
        from app.logic.tool.evidence_prepare import _generate_summary

        evidences = [{"object_type_name": "张三"}]
        summary = _generate_summary(evidences)
        assert "张三" in summary
        assert "1" in summary

    def test_multiple_evidences(self):
        """多个证据的摘要（<=5个）"""
        from app.logic.tool.evidence_prepare import _generate_summary

        evidences = [
            {"object_type_name": "张三"},
            {"object_type_name": "李四"},
            {"object_type_name": "王五"},
        ]
        summary = _generate_summary(evidences)
        assert "3" in summary

    def test_many_evidences(self):
        """大量证据的摘要（>5个）"""
        from app.logic.tool.evidence_prepare import _generate_summary

        evidences = [{"object_type_name": f"实体{i}"} for i in range(10)]
        summary = _generate_summary(evidences)
        assert "10" in summary

    def test_empty_evidences(self):
        """空证据列表的摘要"""
        from app.logic.tool.evidence_prepare import _generate_summary

        assert _generate_summary([]) == "无证据"


class TestDeduplicateEvidences:
    """测试 _deduplicate_evidences 函数"""

    def test_deduplicate_by_name(self):
        """按名称去重，保留分数高的"""
        from app.logic.tool.evidence_prepare import _deduplicate_evidences

        evidences = [
            {"object_type_name": "张三", "score": 0.7},
            {"object_type_name": "张三", "score": 0.95},
            {"object_type_name": "李四", "score": 0.8},
        ]

        results = _deduplicate_evidences(evidences)

        names = [r["object_type_name"] for r in results]
        assert names.count("张三") == 1
        assert names.count("李四") == 1
        assert len(results) == 2

        zhang_san = next(r for r in results if r["object_type_name"] == "张三")
        assert zhang_san["score"] == 0.95

    def test_case_insensitive_dedup(self):
        """大小写不敏感去重"""
        from app.logic.tool.evidence_prepare import _deduplicate_evidences

        evidences = [
            {"object_type_name": "Zhang San", "score": 0.9},
            {"object_type_name": "zhang san", "score": 0.7},
        ]

        results = _deduplicate_evidences(evidences)
        assert len(results) == 1


class TestAssignLocalIds:
    """测试 _assign_local_ids 函数"""

    def test_assign_sequential_ids(self):
        """分配递增的 local_id"""
        from app.logic.tool.evidence_prepare import _assign_local_ids

        evidences = [
            {"object_type_name": "A"},
            {"object_type_name": "B"},
            {"object_type_name": "C"},
        ]

        results = _assign_local_ids(evidences)

        assert results[0]["local_id"] == "e1"
        assert results[1]["local_id"] == "e2"
        assert results[2]["local_id"] == "e3"


class TestSerializeForLlm:
    """测试 _serialize_for_llm 函数"""

    def test_dict_serialization(self):
        """字典序列化为 JSON"""
        from app.logic.tool.evidence_prepare import _serialize_for_llm

        data = {"key": "value", "num": 42}
        result = _serialize_for_llm(data)

        assert "key" in result
        assert "value" in result

    def test_non_dict_fallback(self):
        """非字典类型回退为 str"""
        from app.logic.tool.evidence_prepare import _serialize_for_llm

        assert isinstance(_serialize_for_llm(123), str)
