"""
Evidence LLM Annotator - 使用 LLM 进行证据标注

让 LLM 同时查看工具结果和生成的文本，标注文本中引用的实体位置。
这是更可靠的方法，因为只有 LLM 知道它引用了哪些内容。
"""

import json
import logging
from typing import Any, AsyncGenerator, Dict, List, Optional

from app.common.config import Config

logger = logging.getLogger(__name__)


_LLM_ANNOTATE_PROMPT = """你是一个精确的证据标注专家。你的任务是分析文本和工具调用结果，找出文本中实际引用了哪些工具提供的信息。

## 工具调用结果：
{tool_results}

## 待标注文本（长度 {text_len} 个字符）：
```
{text}
```

## 字符位置索引参考（每10个字符标记一次索引）：
{text_with_index}

## 核心原则：
**只有当文本明确引用或展示了工具结果的特定内容时，才进行标注。**
- 如果工具提供了日期信息，文本中提到了该日期 → 标注
- 如果工具提供了搜索结果，文本中总结了这些结果 → 标注
- 如果工具结果与文本内容无关 → 不标注
- 如果只是隐含关系（没有明确引用）→ 不标注

## 标注步骤：

### 步骤1：理解工具提供了什么信息
仔细阅读每个工具的结果，理解它返回了什么数据或信息。

### 步骤2：检查文本是否引用了这些信息
在文本中查找：
- 直接引用：如 "根据搜索结果..."
- 数据展示：如 "当前日期是2026-04-11"
- 内容总结：如 "常州有以下景点..."

### 步骤3：精确标注位置
使用 `text[start:end]` 验证你的位置是否正确：
- text[start:end] 必须是文本中实际引用的部分
- 每个字符（包括空格、换行符、标点）都算一个位置

## object_type_name 字段说明：
这个字段应该描述**工具提供了什么类型的信息**，而不是工具名称：
- 如果工具返回日期 → object_type_name: "日期"
- 如果工具返回景点列表 → object_type_name: "景点列表"
- 如果工具返回天气信息 → object_type_name: "天气信息"
- 如果工具返回任务规划 → object_type_name: "任务规划"

## 关键规则：

1. **只标注明确引用**：必须有明确的语义关联，不能凭猜测标注
2. **精确位置**：标注的位置必须在文本中真实存在
3. **边界检查**：start >= 0, end <= {text_len}, end > start

## 示例：

工具结果：
```
工具1: get_date
结果: "2026-04-11"
```

文本：
```
今天是2026年4月11日，我们开始规划行程。
```

正确标注：
```json
{{
  "evidences": [
    {{
      "object_type_name": "日期",
      "positions": [[2, 14]]
    }}
  ]
}}
```

说明：位置[2,14]是"2026年4月11日"，这是文本中对日期工具结果的明确引用。

错误标注：
```json
{{
  "evidences": [
    {{
      "object_type_name": "_date",
      "positions": [[0, 2]]
    }}
  ]
}}
```

说明：位置[0,2]是"今天"，这不是工具提供的具体日期，应该是完整的日期位置。

## 输出格式：
```json
{{
  "evidences": [
    {{
      "object_type_name": "信息类型描述",
      "positions": [[start1, end1], [start2, end2]]
    }}
  ]
}}
```

如果没有找到明确引用，返回：`{{"evidences": []}}`
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

    if not text or not tool_results:
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

    try:
        from app.driven.dip.model_api_service import model_api_service
        import asyncio

        logger.info(
            f"[llm_annotate_evidence] 开始标注: text_len={len(text)}, "
            f"tool_results_count={len(tool_results)}, model={model}"
        )

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

        logger.info(
            f"[llm_annotate_evidence] LLM返回: {len(result_text)} 字符, "
            f"预览: {result_text[:200]}..."
        )

        parsed = _parse_llm_annotation(result_text, text)

        logger.info(
            f"[llm_annotate_evidence] 最终结果: {len(parsed.get('evidences', []))} 个证据"
        )

        return parsed

    except ImportError:
        logger.warning("[llm_annotate_evidence] model_api_service 不可用")
        return {"evidences": []}
    except asyncio.TimeoutError:
        logger.warning("[llm_annotate_evidence] LLM 调用超时")
        return {"evidences": []}
    except Exception as e:
        logger.warning("[llm_annotate_evidence] LLM 调用失败: {e}", exc_info=True)
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
        except Exception as e:
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

    # 移除 markdown 代码块标记（如果存在）
    # LLM 可能返回: ```json {...} ``` 或 ``` {...} ```
    if text.startswith("```"):
        # 找到第一个换行符，代码块标记后通常跟一个换行
        first_newline = text.find("\n")
        if first_newline != -1:
            # 从第一个换行符后开始查找
            text = text[first_newline + 1:]

        # 移除结尾的代码块标记
        if text.endswith("```"):
            text = text[:-3]
        text = text.strip()

    # 查找 JSON 部分
    json_start = text.find("{")
    json_end = text.rfind("}")

    if json_start == -1 or json_end == -1 or json_end <= json_start:
        logger.warning(
            f"[_parse_llm_annotation] 未找到有效的 JSON, "
            f"json_start={json_start}, json_end={json_end}"
        )
        return {"evidences": []}

    json_str = text[json_start : json_end + 1]

    try:
        data = json.loads(json_str)

        evidences = data.get("evidences", [])
        if not isinstance(evidences, list):
            logger.warning(
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
                logger.warning("[_parse_llm_annotation] 跳过没有 object_type_name 的证据")
                continue

            if not isinstance(positions, list):
                logger.warning("[_parse_llm_annotation] positions 不是数组")
                continue

            # 验证位置是否有效
            valid_positions = []
            for pos in positions:
                if isinstance(pos, list) and len(pos) == 2:
                    start, end = pos
                    if isinstance(start, int) and isinstance(end, int):
                        if 0 <= start < end <= len(original_text):
                            # 提取标注的文本，用于验证
                            extracted = original_text[start:end]
                            # 检查提取的内容是否太短或只包含空白字符
                            if len(extracted.strip()) >= 2:
                                valid_positions.append([start, end])
                                logger.info(
                                    f"[_parse_llm_annotation] 标注位置 [{start}, {end}]: "
                                    f"'{extracted[:50]}...' (type={name})"
                                )
                            else:
                                logger.warning(
                                    f"[_parse_llm_annotation] 跳过空白或过短的标注: "
                                    f"[{start}, {end}] = '{extracted}'"
                                )
                        else:
                            logger.warning(
                                f"[_parse_llm_annotation] Invalid position: "
                                f"[{start}, {end}], text_len={len(original_text)}"
                    )

            if valid_positions:
                cleaned_evidences.append({
                    "object_type_name": name,
                    "positions": valid_positions,
                })
            else:
                logger.warning(
                    f"[_parse_llm_annotation] 证据 '{name}' 没有有效的位置，已跳过"
                )

        return {"evidences": cleaned_evidences}

    except (json.JSONDecodeError, ValueError, TypeError) as e:
        logger.warning(
            f"[_parse_llm_annotation] JSON 解析失败: {e}, "
            f"json_str={json_str[:200]}"
        )
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
