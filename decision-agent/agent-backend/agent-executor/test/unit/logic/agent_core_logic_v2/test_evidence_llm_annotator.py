# -*- coding: utf-8 -*-
"""单元测试 - evidence_llm_annotator 模块"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch


class TestFormatToolResults:
    """测试 _format_tool_results 函数"""

    def test_format_single_tool_result(self):
        """测试格式化单个工具结果"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        tool_results = [
            {
                "tool_name": "search_employee",
                "result": '{"name": "张三", "age": 30}',
                "result_type": "dict",
            }
        ]

        result = _format_tool_results(tool_results)

        assert "search_employee" in result
        assert "dict" in result
        assert "张三" in result

    def test_format_multiple_tool_results(self):
        """测试格式化多个工具结果"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        tool_results = [
            {
                "tool_name": "tool_a",
                "result": '{"data": "A"}',
                "result_type": "dict",
            },
            {
                "tool_name": "tool_b",
                "result": '{"data": "B"}',
                "result_type": "dict",
            },
        ]

        result = _format_tool_results(tool_results)

        assert "tool_a" in result
        assert "tool_b" in result
        assert "### 工具 1:" in result
        assert "### 工具 2:" in result

    def test_format_invalid_json_result(self):
        """测试非 JSON 结果的格式化"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        tool_results = [
            {
                "tool_name": "date_tool",
                "result": "2024-04-11",
                "result_type": "str",
            }
        ]

        result = _format_tool_results(tool_results)

        assert "date_tool" in result
        assert "2024-04-11" in result
        assert "str" in result

    def test_format_long_result_truncation(self):
        """测试长结果被截断到3000字符"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        long_data = "x" * 4000
        tool_results = [
            {
                "tool_name": "long_tool",
                "result": long_data,
                "result_type": "str",
            }
        ]

        result = _format_tool_results(tool_results)

        # 结果应包含在 ``` 中，但长度不超过约3000
        lines = result.split("```")
        if len(lines) > 1:
            content = lines[1]
            assert len(content) <= 3010  # 允许一些边界误差

    def test_format_empty_tool_results(self):
        """测试空工具结果列表"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        result = _format_tool_results([])
        assert result == ""

    def test_format_tool_with_missing_fields(self):
        """测试缺少字段时的默认值"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        tool_results = [
            {
                "tool_name": "test",
                # result 缺失，会使用默认值 "{}"
                "result_type": "unknown",
            }
        ]

        result = _format_tool_results(tool_results)
        assert "test" in result
        assert "unknown" in result


class TestParseLLMAnnotation:
    """测试 _parse_llm_annotation 函数"""

    def test_parse_valid_json_response(self):
        """测试解析有效的 JSON 响应"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": [{"object_type_name": "张三", "positions": [[0, 2]]}]}'
        original_text = "张三是员工"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"][0]["object_type_name"] == "张三"
        assert parsed["evidences"][0]["positions"] == [[0, 2]]

    def test_parse_markdown_wrapped_json(self):
        """测试解析 markdown 代码块包裹的 JSON"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '''```json
{
  "evidences": [
    {
      "object_type_name": "李四",
      "positions": [[2, 4]]
    }
  ]
}
```'''
        original_text = "员工李四"

        parsed = _parse_llm_annotation(result, original_text)

        assert len(parsed["evidences"]) == 1
        assert parsed["evidences"][0]["object_type_name"] == "李四"

    def test_parse_markdown_without_lang(self):
        """测试解析无语言标记的 markdown 代码块"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '''```
{"evidences": [{"object_type_name": "王五", "positions": [[0, 2]]}]}
```'''
        original_text = "王五是经理"

        parsed = _parse_llm_annotation(result, original_text)

        assert len(parsed["evidences"]) == 1
        assert parsed["evidences"][0]["object_type_name"] == "王五"

    def test_parse_empty_evidences(self):
        """测试解析空证据列表"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": []}'
        original_text = "没有任何引用"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_invalid_position_out_of_bounds(self):
        """测试过滤超出边界的位置"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        # 位置超出文本长度
        result = '{"evidences": [{"object_type_name": "实体", "positions": [[0, 100]]}]}'
        original_text = "短"

        parsed = _parse_llm_annotation(result, original_text)

        # 超出边界的位置应被过滤
        assert len(parsed["evidences"]) == 0

    def test_parse_invalid_position_format(self):
        """测试过滤无效的位置格式"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        # 位置格式不正确
        result = '{"evidences": [{"object_type_name": "实体", "positions": [[0]]}]}'
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert len(parsed["evidences"]) == 0

    def test_parse_mixed_valid_invalid_positions(self):
        """测试混合有效和无效位置"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": [{"object_type_name": "测试", "positions": [[0, 2], [5, 100], [10, 12]]}]}'
        original_text = "12345678901234567890"

        parsed = _parse_llm_annotation(result, original_text)

        # 只保留有效的位置
        assert len(parsed["evidences"]) == 1
        positions = parsed["evidences"][0]["positions"]
        assert [0, 2] in positions
        assert [10, 12] in positions
        assert [5, 100] not in positions

    def test_parse_missing_evidences_field(self):
        """测试缺少 evidences 字段"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"data": "something"}'
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_evidences_not_array(self):
        """测试 evidences 不是数组"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": "not an array"}'
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_empty_input(self):
        """测试空输入"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        parsed = _parse_llm_annotation("", "测试")
        assert parsed["evidences"] == []

        parsed = _parse_llm_annotation(None, "测试")
        assert parsed["evidences"] == []

    def test_parse_no_json_found(self):
        """测试未找到有效 JSON"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = "这是一段普通文本，没有 JSON"
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_json_with_extra_text(self):
        """测试 JSON 后有额外文本"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '''{"evidences": [{"object_type_name": "测试", "positions": [[0, 2]]}]}
这是额外的解释文本。'''
        original_text = "测试文本"

        parsed = _parse_llm_annotation(result, original_text)

        assert len(parsed["evidences"]) == 1
        assert parsed["evidences"][0]["object_type_name"] == "测试"

    def test_parse_invalid_json(self):
        """测试无效的 JSON"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": [incomplete json'
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_evidence_without_name(self):
        """测试缺少 object_type_name 的证据"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '{"evidences": [{"positions": [[0, 2]]}]}'
        original_text = "测试"

        parsed = _parse_llm_annotation(result, original_text)

        assert parsed["evidences"] == []

    def test_parse_multiple_evidences(self):
        """测试解析多个证据"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        result = '''{"evidences": [
            {"object_type_name": "张三", "positions": [[0, 2]]},
            {"object_type_name": "李四", "positions": [[2, 4]]}
        ]}'''
        original_text = "张三李四"

        parsed = _parse_llm_annotation(result, original_text)

        assert len(parsed["evidences"]) == 2


class TestGenerateIndexReference:
    """测试 _generate_index_reference 函数"""

    def test_generate_short_text(self):
        """测试生成短文本的索引参考"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        text = "Hello World"
        result = _generate_index_reference(text)

        assert "Hello World" in result
        assert "0" in result  # 索引从 0 开始

    def test_generate_long_text_multiple_lines(self):
        """测试生成长文本的多行索引"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        text = "a" * 100
        result = _generate_index_reference(text)

        # 应包含换行分隔
        assert "\n" in result

    def test_generate_empty_text(self):
        """测试空文本返回空字符串"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        result = _generate_index_reference("")
        assert result == ""

        result = _generate_index_reference(None)
        assert result == ""

    def test_generate_index_every_ten_chars(self):
        """测试每10个字符标记一次索引"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        text = "abcdefghij" * 2  # 20 字符
        result = _generate_index_reference(text)

        # 应该在 0 和 10 位置有索引标记
        lines = result.split("\n")
        index_line = [l for l in lines if "0" in l or "10" in l]

        assert len(index_line) > 0

    def test_generate_with_newlines(self):
        """测试包含换行符的文本"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        text = "第一行\n第二行\n第三行"
        result = _generate_index_reference(text)

        assert "第一行" in result
        assert "第二行" in result
        assert "第三行" in result


class TestLLMAnnotateEvidence:
    """测试 llm_annotate_evidence 函数"""

    @pytest.mark.asyncio
    async def test_returns_empty_for_no_text(self):
        """测试无文本时返回空结果"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        tool_results = [
            {"tool_name": "test", "result": "{}", "result_type": "dict"}
        ]

        result = await llm_annotate_evidence("", tool_results)
        assert result["evidences"] == []

        result = await llm_annotate_evidence(None, tool_results)
        assert result["evidences"] == []

    @pytest.mark.asyncio
    async def test_returns_empty_for_no_tool_results(self):
        """测试无工具结果时返回空结果"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        result = await llm_annotate_evidence("测试文本", [])
        assert result["evidences"] == []

        result = await llm_annotate_evidence("测试文本", None)
        assert result["evidences"] == []

    @pytest.mark.asyncio
    async def test_uses_default_config(self):
        """测试使用默认配置"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        tool_results = [
            {"tool_name": "test", "result": '{"data": "value"}', "result_type": "dict"}
        ]

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(return_value='{"evidences": []}')

            await llm_annotate_evidence("测试", tool_results)

            # 验证使用默认模型和超时
            mock_service.call.assert_called_once()
            call_args = mock_service.call.call_args
            assert call_args[1]["model"] == "deepseek-v3.2"
            assert call_args[1]["max_tokens"] == 2000

    @pytest.mark.asyncio
    async def test_uses_custom_config(self):
        """测试使用自定义配置"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        tool_results = [
            {"tool_name": "test", "result": "{}", "result_type": "dict"}
        ]

        config = {
            "llm_annotation_timeout": 60,
            "llm_annotation_model": "custom-model",
        }

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(return_value='{"evidences": []}')

            await llm_annotate_evidence("测试", tool_results, config)

            call_args = mock_service.call.call_args
            assert call_args[1]["model"] == "custom-model"

    @pytest.mark.asyncio
    async def test_parses_llm_response(self):
        """测试解析 LLM 响应"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        tool_results = [
            {"tool_name": "search", "result": '{"name": "张三"}', "result_type": "dict"}
        ]

        llm_response = '{"evidences": [{"object_type_name": "张三", "positions": [[0, 2]]}]}'

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(return_value=llm_response)

            result = await llm_annotate_evidence("张三在这里", tool_results)

            assert len(result["evidences"]) == 1
            assert result["evidences"][0]["object_type_name"] == "张三"

    @pytest.mark.asyncio
    async def test_handles_import_error(self):
        """测试处理导入错误"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )
        import sys

        tool_results = [
            {"tool_name": "test", "result": "{}", "result_type": "dict"}
        ]

        with patch.dict(sys.modules, {"app.driven.dip.model_api_service": None}):
            result = await llm_annotate_evidence("测试", tool_results)
            assert result["evidences"] == []

    @pytest.mark.asyncio
    async def test_handles_timeout(self):
        """测试处理超时"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )
        import asyncio

        tool_results = [
            {"tool_name": "test", "result": "{}", "result_type": "dict"}
        ]

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(side_effect=asyncio.TimeoutError())

            result = await llm_annotate_evidence("测试", tool_results)
            assert result["evidences"] == []

    @pytest.mark.asyncio
    async def test_handles_generic_exception(self):
        """测试处理通用异常"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            llm_annotate_evidence,
        )

        tool_results = [
            {"tool_name": "test", "result": "{}", "result_type": "dict"}
        ]

        with patch(
            "app.driven.dip.model_api_service.model_api_service"
        ) as mock_service:
            mock_service.call = AsyncMock(side_effect=Exception("API error"))

            result = await llm_annotate_evidence("测试", tool_results)
            assert result["evidences"] == []


class TestEdgeCases:
    """边界情况测试"""

    def test_format_with_unicode_characters(self):
        """测试包含 Unicode 字符的工具结果"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _format_tool_results,
        )

        tool_results = [
            {
                "tool_name": "search",
                "result": '{"name": "张三🎉", "emoji": "测试"}',
                "result_type": "dict",
            }
        ]

        result = _format_tool_results(tool_results)

        assert "张三" in result
        assert "🎉" in result

    def test_parse_with_newline_in_positions(self):
        """测试包含换行符的位置"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _parse_llm_annotation,
        )

        # 文本包含换行符
        original_text = "第一行\n第二行"
        # "第一行" 的位置是 [0, 3]
        result = f'{{"evidences": [{{"object_type_name": "文本", "positions": [[0, 3]]}}]}}'

        parsed = _parse_llm_annotation(result, original_text)

        # 位置应该有效
        assert len(parsed["evidences"]) == 1
        assert parsed["evidences"][0]["positions"] == [[0, 3]]

    def test_generate_index_with_very_long_text(self):
        """测试生成超长文本的索引"""
        from app.logic.agent_core_logic_v2.evidence_llm_annotator import (
            _generate_index_reference,
        )

        text = "a" * 1000
        result = _generate_index_reference(text)

        # 长文本应该产生多行
        assert "\n" in result
        # 应包含所有字符
        assert "a" in result
