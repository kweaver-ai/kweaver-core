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


_LLM_ANNOTATE_PROMPT = """你是一个精确的证据标注专家。你的任务是找出文本中引用工具结果的位置，并返回精确的字符索引。

## 工具调用结果：
{tool_results}

## 待标注文本（长度 {text_len} 个字符）：
```
{text}
```

## 字符位置索引参考（每10个字符标记一次索引）：
{text_with_index}

## 标注方法：

### 步骤1：使用上面的索引参考快速定位
上面的索引参考每10个字符标记一次位置，帮助你快速找到实体在文本中的准确位置。

### 步骤2：精确验证
使用 `text[start:end]` 验证你的位置是否正确：
- text[start:end] 必须完全等于实体名称
- 每个字符（包括空格、换行符、标点）都算一个位置

### 步骤3：返回结果
只返回文本中实际存在且逐字匹配的实体位置。

## 关键规则（必须遵守）：

1. **换行符和空格**：
   - `\n` 是一个字符，占用1个位置
   - 空格也是1个字符
   - 必须把所有字符都计入位置

2. **精确验证**：
   - text[start:end] 必须等于实体名称
   - 如果 text[60:67] ≠ "简单对话助手"，说明位置错误

3. **边界检查**：
   - start >= 0
   - end <= {text_len}
   - end > start

## 示例：

文本：
```
第1行内容\n第2行内容
```

字符位置：
```
01234567890123
第1行内容
第2行内容
```

"第1行" 的位置是 [0, 3]，因为 text[0:3] = "第1行"
"第2行" 的位置是 [8, 11]，因为 text[8:11] = "第2行"（注意 \n 占用位置7）

## 输出格式：
```json
{{
  "evidences": [
    {{
      "object_type_name": "实体名称",
      "positions": [[start1, end1], [start2, end2]]
    }}
  ]
}}
```

如果没有找到引用，返回：`{{"evidences": []}}`
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

    # 生成字符位置索引参考（每10个字符标记一次）
    text_with_index = _generate_index_reference(text)

    prompt = _LLM_ANNOTATE_PROMPT.format(
        tool_results=tool_results_text[:10000],  # 限制长度
        text=text,
        text_len=len(text),
        text_with_index=text_with_index
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
    StandLogger.info_log(
        f"[_format_tool_results] START: tool_results_count={len(tool_results)}"
    )
    for idx, tr in enumerate(tool_results):
        tool_name = tr.get("tool_name", "unknown")
        result = tr.get("result", "{}")
        result_type = tr.get("result_type", "unknown")

        StandLogger.info_log(
            f"[_format_tool_results] tool {idx + 1}: "
            f"tool_name={tool_name}, "
            f"result_type={result_type}, "
            f"result_length={len(result)}, "
            f"result_preview={result[:200] if result else 'EMPTY'}..."
        )

        # 尝试解析 JSON 格式化显示
        try:
            result_obj = json.loads(result)
            formatted = json.dumps(result_obj, ensure_ascii=False, indent=2)
            StandLogger.info_log(
                f"[_format_tool_results] Parsed JSON successfully, formatted_length={len(formatted)}"
            )
        except Exception as e:
            formatted = result
            StandLogger.info_log(
                f"[_format_tool_results] JSON parse failed: {e}, using raw result"
            )

        parts.append(
            f"### 工具 {idx + 1}: {tool_name} (类型: {result_type})\n"
            f"```\n{formatted[:3000]}\n```\n"
        )
    result = "\n".join(parts)
    StandLogger.info_log(
        f"[_format_tool_results] END: total_length={len(result)}"
    )
    return result


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


def _generate_index_reference(text: str) -> str:
    """
    生成字符位置索引参考，每10个字符标记一次位置。

    Args:
        text: 待标注的文本

    Returns:
        带有索引标记的文本，每10个字符显示一次索引
    """
    if not text:
        return ""

    lines = []
    index_line = []
    text_line = []

    for i, char in enumerate(text):
        # 每10个字符标记一次索引
        if i % 10 == 0:
            # 补齐索引到两位数，对齐显示
            index_str = str(i)
            index_line.append(index_str)
            # 添加空格对齐（根据索引长度调整）
            remaining = 10 - (i % 10)
        else:
            index_line.append(" " * len(str(i)))

        text_line.append(char)

        # 每80个字符换行（避免单行过长）
        if (i + 1) % 80 == 0:
            lines.append("".join(text_line))
            lines.append("".join(index_line))
            lines.append("")  # 空行分隔
            text_line = []
            index_line = []

    # 添加剩余内容
    if text_line:
        lines.append("".join(text_line))
        lines.append("".join(index_line))

    return "\n".join(lines)
