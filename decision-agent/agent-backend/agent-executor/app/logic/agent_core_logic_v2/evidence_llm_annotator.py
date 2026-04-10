"""
Evidence LLM Annotator - 使用 LLM 进行证据标注

让 LLM 同时查看工具结果和生成的文本，标注文本中引用的实体位置。
这是更可靠的方法，因为只有 LLM 知道它引用了哪些内容。
"""

import json
import logging
from typing import Any, AsyncGenerator, Dict, List, Optional

from app.common.config import Config
from app.common.stand_log import StandLogger

logger = logging.getLogger(__name__)


_LLM_ANNOTATE_PROMPT = """你是一个证据标注专家。请标注文本中引用工具结果的位置。

## 工具调用结果：
{tool_results}

## 待标注文本：
{text}

## 任务说明：
请找出文本中引用工具结果的地方，标注这些引用的位置。

对于每个引用，返回：
1. 实体名称（如"简单对话助手"、"基于简单对话模板创建"等）
2. 在文本中的起始位置（字符索引，从0开始）
3. 在文本中的结束位置（字符索引，不包含）

## 重要规则：
1. 位置索引要精确，必须是文本中实际出现的字符串
2. 可以标注多个引用
3. 实体名称应该是文本中实际出现的词
4. 只标注直接引用工具结果的部分

## 输出格式（JSON）：
{{
  "evidences": [
    {{
      "object_type_name": "实体名称",
      "positions": [[start, end]]
    }}
  ]
}}

示例：
如果文本是"简单对话助手是一个基于简单对话模板创建的助手"
工具结果包含 name="简单对话助手", profile="基于简单对话模板创建"
则输出：
{{
  "evidences": [
    {{"object_type_name": "简单对话助手", "positions": [[0, 7]]}},
    {{"object_type_name": "基于简单对话模板创建", "positions": [[11, 23]]}}
  ]
}}

只返回 JSON，不要其他内容。如果没有引用，返回 {{"evidences": []}}
"""


async def llm_annotate_evidence(
    text: str,
    tool_results: List[Dict[str, Any]],
    config: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    使用 LLM 标注文本中引用工具结果的位置。

    Args:
        text: LLM 生成的文本
        tool_results: 工具调用结果列表 [{tool_name, result, result_type}, ...]
        config: 配置参数

    Returns:
        {
            "evidences": [
                {
                    "object_type_name": "实体名称",
                    "positions": [[start, end], ...]
                }
            ]
        }
    """
    config = config or {}
    timeout = config.get("llm_annotation_timeout", 30)
    model = config.get("llm_annotation_model", "deepseek-v3.2")

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[llm_annotate_evidence] START\n"
        f"  text_length: {len(text)}\n"
        f"  tool_results_count: {len(tool_results)}\n"
        f"  model: {model}\n"
        f"  timeout: {timeout}s\n"
        f"{'='*60}\n"
    )

    if not text or not tool_results:
        StandLogger.info_log(
            f"[llm_annotate_evidence] END: no text or tool results"
        )
        return {"evidences": []}

    # 格式化工具结果为可读文本
    tool_results_text = _format_tool_results(tool_results)

    prompt = _LLM_ANNOTATE_PROMPT.format(
        tool_results=tool_results_text[:10000],  # 限制长度
        text=text
    )

    StandLogger.info_log(
        f"[llm_annotate_evidence] Calling LLM, prompt_length={len(prompt)}"
    )

    try:
        from app.driven.dip.model_api_service import model_api_service
        import asyncio

        messages = [{"role": "user", "content": prompt}]

        llm_result = await asyncio.wait_for(
            model_api_service.call(
                model=model,
                messages=messages,
                max_tokens=2000,
            ),
            timeout=timeout,
        )

        result_text = str(llm_result) if llm_result else ""

        StandLogger.info_log(
            f"[llm_annotate_evidence] LLM returned, length={len(result_text)}, "
            f"preview: {result_text[:200]}..."
        )

        parsed = _parse_llm_annotation(result_text, text)

        StandLogger.info_log(
            f"[llm_annotate_evidence] Parsed {len(parsed.get('evidences', []))} evidences"
        )

        for ev in parsed.get("evidences", []):
            name = ev.get("object_type_name", "")
            positions = ev.get("positions", [])
            matched_texts = []
            for start, end in positions:
                if 0 <= start < end <= len(text):
                    matched_texts.append(f'"{text[start:end]}"')
            StandLogger.info_log(
                f"  - {name}: positions={positions}, matched={matched_texts}"
            )

        StandLogger.info_log(
            f"[llm_annotate_evidence] END\n"
            f"{'='*60}\n"
        )

        return parsed

    except ImportError:
        StandLogger.info_log("[llm_annotate_evidence] model_api_service 不可用")
        logger.warning("[llm_annotate_evidence] model_api_service 不可用")
        return {"evidences": []}
    except asyncio.TimeoutError:
        StandLogger.info_log(f"[llm_annotate_evidence] LLM 调用超时 (timeout={timeout}s)")
        logger.warning("[llm_annotate_evidence] LLM 调用超时")
        return {"evidences": []}
    except Exception as e:
        StandLogger.info_log(f"[llm_annotate_evidence] LLM 调用失败: {e}")
        logger.warning(f"[llm_annotate_evidence] LLM 调用失败: {e}", exc_info=True)
        return {"evidences": []}


def _format_tool_results(tool_results: List[Dict[str, Any]]) -> str:
    """格式化工具结果为可读文本"""
    parts = []
    for idx, tr in enumerate(tool_results):
        tool_name = tr.get("tool_name", "unknown")
        result = tr.get("result", "{}")
        result_type = tr.get("result_type", "unknown")

        # 尝试解析 JSON 格式化显示
        try:
            result_obj = json.loads(result)
            formatted = json.dumps(result_obj, ensure_ascii=False, indent=2)
        except:
            formatted = result

        parts.append(
            f"### 工具 {idx + 1}: {tool_name} (类型: {result_type})\n"
            f"```\n{formatted[:3000]}\n```\n"
        )
    return "\n".join(parts)


def _parse_llm_annotation(result: str, original_text: str) -> Dict[str, Any]:
    """解析 LLM 返回的标注结果"""
    if not result or not result.strip():
        return {"evidences": []}

    text = result.strip()

    # 查找 JSON 部分
    json_start = text.find("{")
    json_end = text.rfind("}")

    if json_start == -1 or json_end == -1 or json_end <= json_start:
        StandLogger.info_log(
            f"[_parse_llm_annotation] 未找到有效的 JSON, "
            f"json_start={json_start}, json_end={json_end}"
        )
        return {"evidences": []}

    try:
        json_str = text[json_start : json_end + 1]
        data = json.loads(json_str)

        evidences = data.get("evidences", [])
        if not isinstance(evidences, list):
            StandLogger.info_log(
                f"[_parse_llm_annotation] evidences 不是数组, type={type(evidences).__name__}"
            )
            return {"evidences": []}

        # 验证并清理位置信息
        cleaned_evidences = []
        for ev in evidences:
            if not isinstance(ev, dict):
                continue

            name = ev.get("object_type_name", "")
            positions = ev.get("positions", [])

            if not name:
                continue

            if not isinstance(positions, list):
                continue

            # 验证位置是否有效
            valid_positions = []
            for pos in positions:
                if isinstance(pos, list) and len(pos) == 2:
                    start, end = pos
                    if isinstance(start, int) and isinstance(end, int):
                        if 0 <= start < end <= len(original_text):
                            valid_positions.append([start, end])
                        else:
                            StandLogger.info_log(
                                f"[_parse_llm_annotation] Invalid position: "
                                f"[{start}, {end}], text_len={len(original_text)}"
                            )

            if valid_positions:
                cleaned_evidences.append({
                    "object_type_name": name,
                    "positions": valid_positions,
                })

        StandLogger.info_log(
            f"[_parse_llm_annotation] Cleaned {len(cleaned_evidences)}/{len(evidences)} evidences"
        )

        return {"evidences": cleaned_evidences}

    except (json.JSONDecodeError, ValueError, TypeError) as e:
        StandLogger.info_log(
            f"[_parse_llm_annotation] JSON 解析失败: {e}"
        )
        logger.warning(f"[_parse_llm_annotation] JSON 解析失败: {e}")
        return {"evidences": []}
