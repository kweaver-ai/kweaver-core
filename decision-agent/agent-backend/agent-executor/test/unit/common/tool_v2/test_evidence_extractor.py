# -*- coding: utf-8 -*-
"""单元测试 - evidence_extractor 模块"""

import os
from unittest import mock

import pytest


class TestIsEvidenceExtractionEnabled:
    """测试 is_evidence_extraction_enabled 函数"""

    def test_default_false(self):
        """未设置环境变量时默认返回 False"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {}, clear=True):
            assert is_evidence_extraction_enabled() is False

    def test_enabled_true(self):
        """设置为 true 时返回 True"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {"ENABLE_EVIDENCE_EXTRACTION": "true"}):
            assert is_evidence_extraction_enabled() is True

    def test_enabled_1(self):
        """设置为 1 时返回 True"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {"ENABLE_EVIDENCE_EXTRACTION": "1"}):
            assert is_evidence_extraction_enabled() is True

    def test_enabled_yes(self):
        """设置为 yes 时返回 True"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {"ENABLE_EVIDENCE_EXTRACTION": "yes"}):
            assert is_evidence_extraction_enabled() is True

    def test_disabled_false(self):
        """设置为 false 时返回 False"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {"ENABLE_EVIDENCE_EXTRACTION": "false"}):
            assert is_evidence_extraction_enabled() is False

    def test_case_insensitive(self):
        """大小写不敏感"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            is_evidence_extraction_enabled,
        )

        with mock.patch.dict(os.environ, {"ENABLE_EVIDENCE_EXTRACTION": "TRUE"}):
            assert is_evidence_extraction_enabled() is True


class TestExtractEvidence:
    """测试 extract_evidence 函数"""

    def test_extract_with_nodes(self):
        """正常提取 nodes，验证只保留 3 个字段"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        data = {
            "nodes": [
                {
                    "object_type_id": "chunk",
                    "object_type_name": "合同文档内容特征对象类",
                    "properties": {"id": "xxx", "doc_name": "test.doc"},
                    "score": 0.3,
                },
                {
                    "object_type_id": "document",
                    "object_type_name": "合同文档对象类",
                    "properties": {"name": "test.docx"},
                    "score": 0.5,
                },
            ],
            "relation_types": [],
        }

        result = extract_evidence(data)

        assert result is not None
        assert result["is_send_to_llm"] is False
        assert len(result["evidences"]) == 1
        assert result["evidences"][0]["module"] == "bkn"

        instances = result["evidences"][0]["content"]["object_instances"]
        assert len(instances) == 2

        # 验证只保留了 3 个字段
        assert instances[0] == {
            "object_type_id": "chunk",
            "object_type_name": "合同文档内容特征对象类",
            "score": 0.3,
        }
        assert instances[1] == {
            "object_type_id": "document",
            "object_type_name": "合同文档对象类",
            "score": 0.5,
        }

        # 验证 properties 被过滤掉了
        for inst in instances:
            assert "properties" not in inst

    def test_extract_no_nodes(self):
        """无 nodes 字段时返回 None"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        assert extract_evidence({"relation_types": []}) is None

    def test_extract_empty_nodes(self):
        """nodes 为空列表时返回 None"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        assert extract_evidence({"nodes": []}) is None

    def test_extract_non_dict_input(self):
        """输入非 dict 时返回 None"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        assert extract_evidence("string_data") is None
        assert extract_evidence(None) is None
        assert extract_evidence(123) is None
        assert extract_evidence([]) is None

    def test_extract_nodes_with_missing_fields(self):
        """nodes 中的对象缺少部分字段时仍能正常提取"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        data = {
            "nodes": [
                {"object_type_id": "chunk", "score": 0.3},  # 缺少 object_type_name
            ]
        }

        result = extract_evidence(data)
        assert result is not None
        instances = result["evidences"][0]["content"]["object_instances"]
        assert len(instances) == 1
        assert instances[0] == {"object_type_id": "chunk", "score": 0.3}

    def test_extract_nodes_with_non_dict_items(self):
        """nodes 包含非 dict 元素时自动跳过"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        data = {
            "nodes": [
                "invalid_item",
                {"object_type_id": "chunk", "object_type_name": "test", "score": 0.1},
                None,
            ]
        }

        result = extract_evidence(data)
        assert result is not None
        instances = result["evidences"][0]["content"]["object_instances"]
        assert len(instances) == 1

    def test_output_structure(self):
        """验证完整的 _evidence 输出结构"""
        from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
            extract_evidence,
        )

        data = {
            "nodes": [
                {
                    "object_type_id": "chunk",
                    "object_type_name": "测试",
                    "score": 0.5,
                    "properties": {"key": "value"},
                }
            ]
        }

        result = extract_evidence(data)

        # 顶层结构
        assert "is_send_to_llm" in result
        assert "evidences" in result
        assert isinstance(result["evidences"], list)

        # evidence 结构
        evidence = result["evidences"][0]
        assert "module" in evidence
        assert "content" in evidence

        # content 结构
        content = evidence["content"]
        assert "object_instances" in content
        assert isinstance(content["object_instances"], list)
