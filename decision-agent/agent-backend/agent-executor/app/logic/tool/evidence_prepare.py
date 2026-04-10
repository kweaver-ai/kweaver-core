"""
Evidence Prepare 工具

从工具调用结果中提取和标准化证据，统一使用 LLM 提取实体，
LLM 失败时使用规则回退策略。
"""

import json
import logging
import uuid
from typing import Any, Dict, List, Optional

from app.common.stand_log import StandLogger
from app.common.tool_v2.api_tool_pkg.evidence_extractor import (
    extract_evidence,
    is_evidence_extraction_enabled,
)

logger = logging.getLogger(__name__)

_LLM_EXTRACTION_PROMPT = """你是一个实体提取专家。请从以下工具调用结果中提取关键实体。

工具调用结果:
{tool_call_result}

请提取所有有意义的实体（人名、地名、组织名、专有名词等），以 JSON 数组格式返回。
每个实体包含以下字段：
- object_type_name: 实体名称（必填）
- aliases: 别名列表（可选，如 ["员工张三", "张工"]）
- object_type_id: 实体类型 ID（可选，如 "ot_employee"）
- score: 置信度 0-1（可选）

只返回 JSON 数组，不要其他内容。如果没有找到实体，返回空数组 []。

示例输出:
[
  {"object_type_name": "张三", "aliases": ["员工张三"], "object_type_id": "ot_employee", "score": 0.95},
  {"object_type_name": "上海", "aliases": [], "object_type_id": "ot_city", "score": 0.9}
]"""

_FALLBACK_ENTITY_PATTERNS = {
    "nodes": ("object_type_name", "object_type_id", "score"),
    "object_instances": ("object_type_name", "object_type_id", "score"),
}


async def evidence_prepare(
    tool_call_result: Dict[str, Any],
    config: Optional[Dict[str, Any]] = None,
    context: Optional[Dict[str, Any]] = None,
) -> Dict[str, Any]:
    """
    从工具调用结果中提取和标准化证据。

    统一使用 LLM 提取关键实体，LLM 调用失败时回退到规则提取。

    Args:
        tool_call_result: 工具调用的返回结果
        config: 可选配置参数
        context: 可选上下文信息

    Returns:
        {
            "evidence_id": "ev_uuid",
            "evidences": [
                {
                    "local_id": "e1",
                    "object_type_name": "张三",
                    "aliases": ["员工张三"],
                    "object_type_id": "ot_employee",
                    "score": 0.95
                }
            ],
            "summary": "张三等 1 个证据"
        }
    """
    evidence_id = f"ev_{uuid.uuid4().hex[:12]}"
    config = config or {}

    tool_name = context.get("tool_name", "unknown") if context else "unknown"

    StandLogger.info_log(
        f"[evidence_prepare] 开始处理, evidence_id={evidence_id}, "
        f"tool_name={tool_name}, "
        f"config={config}, "
        f"tool_call_result keys={list(tool_call_result.keys()) if isinstance(tool_call_result, dict) else type(tool_call_result)}"
    )
    logger.info(
        f"[evidence_prepare] tool_name={tool_name}, "
        f"tool_call_result={json.dumps(tool_call_result, ensure_ascii=False, default=str)[:1000]}..."
    )

    evidences = []

    try:
        existing_evidence = _process_with_existing_evidence(tool_call_result)
        if existing_evidence:
            StandLogger.info_log(
                f"[evidence_prepare] 检测到已有 _evidence 结构"
            )
            evidences.extend(existing_evidence)

        llm_evidences = await _extract_by_llm(tool_call_result, config, context)
        if llm_evidences:
            StandLogger.info_log(
                f"[evidence_prepare] LLM 提取到 {len(llm_evidences)} 个实体"
            )
            evidences.extend(llm_evidences)
        else:
            StandLogger.info_log(
                "[evidence_prepare] LLM 未提取到实体，尝试规则回退"
            )
            fallback_evidences = _fallback_extraction(tool_call_result, tool_name)
            if fallback_evidences:
                evidences.extend(fallback_evidences)
                StandLogger.info_log(
                    f"[evidence_prepare] 规则回退提取到 {len(fallback_evidences)} 个实体"
                )
    except Exception as e:
        logger.warning(f"[evidence_prepare] LLM 提取异常: {e}", exc_info=True)
        fallback_evidences = _fallback_extraction(tool_call_result, tool_name)
        evidences.extend(fallback_evidences)

    evidences = _deduplicate_evidences(evidences)
    evidences = _assign_local_ids(evidences)

    summary = _generate_summary(evidences)

    result = {
        "evidence_id": evidence_id,
        "evidences": evidences,
        "summary": summary,
    }

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[evidence_prepare] 完成\n"
        f"  evidence_id: {evidence_id}\n"
        f"  提取数量: {len(evidences)}\n"
        f"  summary: {summary}\n"
    )

    if evidences:
        StandLogger.info_log("  证据列表:")
        for ev in evidences:
            StandLogger.info_log(
                f"    - {ev.get('local_id')}: "
                f"name={ev.get('object_type_name')}, "
                f"aliases={ev.get('aliases')}, "
                f"type_id={ev.get('object_type_id')}, "
                f"score={ev.get('score')}, "
                f"source={ev.get('_source')}"
            )
    else:
        StandLogger.info_log("  无证据提取")

    StandLogger.info_log(f"{'='*60}\n")

    return result


def _process_with_existing_evidence(
    tool_call_result: Dict[str, Any],
) -> List[Dict[str, Any]]:
    """
    处理已有 _evidence 结构的数据。

    如果工具返回已经包含 _evidence 字段或 nodes 字段，
    则先从中提取已有的证据信息。
    """
    evidences = []

    extracted = extract_evidence(tool_call_result)
    if extracted and extracted.get("evidences"):
        for module_evidence in extracted["evidences"]:
            content = module_evidence.get("content", {})
            instances = content.get("object_instances", [])
            for inst in instances:
                name = inst.get("object_type_name", "")
                if not name:
                    continue
                evidences.append({
                    "object_type_name": name,
                    "aliases": [],
                    "object_type_id": inst.get("object_type_id", ""),
                    "score": inst.get("score", 0.8),
                    "_source": "existing_evidence",
                })

    return evidences


async def _extract_by_llm(
    tool_call_result: Dict[str, Any],
    config: Dict[str, Any],
    context: Optional[Dict[str, Any]],
) -> List[Dict[str, Any]]:
    """
    使用 LLM 从工具调用结果中提取实体。

    Args:
        tool_call_result: 工具调用结果
        config: 配置参数 (可包含 model, timeout 等)
        context: 上下文信息

    Returns:
        提取到的实体列表
    """
    timeout = config.get("llm_extraction_timeout", 30)
    model = config.get("llm_extraction_model", "")

    tool_name = context.get("tool_name", "unknown") if context else "unknown"

    StandLogger.info_log(
        f"\n{'='*60}\n"
        f"[_extract_by_llm] START\n"
        f"  tool_name: {tool_name}\n"
        f"  model: {model}\n"
        f"  timeout: {timeout}s\n"
        f"{'='*60}\n"
    )

    result_str = _serialize_for_llm(tool_call_result)
    if len(result_str) < 10:
        StandLogger.info_log("[_extract_by_llm] 结果太短，跳过")
        return []

    StandLogger.info_log(
        f"[_extract_by_llm] 序列化后长度: {len(result_str)}, "
        f"截取到 8000 字符"
    )

    prompt = _LLM_EXTRACTION_PROMPT.format(tool_call_result=result_str[:8000])

    try:
        from app.driven.dip.model_api_service import model_api_service

        messages = [{"role": "user", "content": prompt}]

        StandLogger.info_log(
            f"[_extract_by_llm] 调用 LLM, model={model}, "
            f"prompt 长度={len(prompt)}"
        )

        import asyncio
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
            f"[_extract_by_llm] LLM 返回, 长度={len(result_text)}, "
            f"内容预览: {result_text[:200]}..."
        )

        parsed = _parse_llm_result(result_text)

        StandLogger.info_log(
            f"[_extract_by_llm] 解析结果: 提取到 {len(parsed)} 个实体"
        )

        if parsed:
            for idx, ev in enumerate(parsed):
                StandLogger.info_log(
                    f"  实体 {idx + 1}: "
                    f"name={ev.get('object_type_name')}, "
                    f"aliases={ev.get('aliases')}, "
                    f"type_id={ev.get('object_type_id')}, "
                    f"score={ev.get('score')}"
                )

        return parsed

    except ImportError:
        StandLogger.info_log("[_extract_by_llm] model_api_service 不可用")
        logger.warning("[evidence_prepare] model_api_service 不可用")
        return []
    except asyncio.TimeoutError:
        StandLogger.info_log(f"[extract_by_llm] LLM 调用超时 (timeout={timeout}s)")
        logger.warning(f"[evidence_prepare] LLM 调用超时")
        return []
    except Exception as e:
        StandLogger.info_log(f"[_extract_by_llm] LLM 调用失败: {e}")
        logger.warning(f"[evidence_prepare] LLM 调用失败: {e}")
        return []


def _parse_llm_result(result: str) -> List[Dict[str, Any]]:
    """
    解析 LLM 返回的 JSON 结果。

    Args:
        result: LLM 返回的原始字符串

    Returns:
        解析后的实体列表
    """
    if not result or not result.strip():
        StandLogger.info_log("[_parse_llm_result] 结果为空")
        return []

    text = result.strip()

    json_start = text.find("[")
    json_end = text.rfind("]")

    if json_start == -1 or json_end == -1 or json_end <= json_start:
        StandLogger.info_log(
            f"[_parse_llm_result] 未找到有效的 JSON 数组, "
            f"json_start={json_start}, json_end={json_end}"
        )
        return []

    try:
        json_str = text[json_start : json_end + 1]
        StandLogger.info_log(
            f"[_parse_llm_result] 提取 JSON 字符串, 长度={len(json_str)}"
        )

        entities = json.loads(json_str)

        if not isinstance(entities, list):
            StandLogger.info_log(
                f"[_parse_llm_result] 解析结果不是数组, 类型={type(entities).__name__}"
            )
            return []

        evidences = []
        for idx, entity in enumerate(entities):
            if not isinstance(entity, dict):
                StandLogger.info_log(f"[_parse_llm_result] 实体 {idx} 不是字典，跳过")
                continue

            name = entity.get("object_type_name", "")
            if not name or not isinstance(name, str):
                StandLogger.info_log(f"[_parse_llm_result] 实体 {idx} 无有效 name，跳过")
                continue

            evidences.append({
                "object_type_name": name,
                "aliases": entity.get("aliases", []) or [],
                "object_type_id": entity.get("object_type_id", ""),
                "score": float(entity.get("score", 0.8)),
                "_source": "llm_extraction",
            })

        StandLogger.info_log(
            f"[_parse_llm_result] 成功解析 {len(evidences)}/{len(entities)} 个实体"
        )

        return evidences

    except (json.JSONDecodeError, ValueError, TypeError) as e:
        StandLogger.info_log(
            f"[_parse_llm_result] JSON 解析失败: {e}\n"
            f"  原始内容前 200 字符: {text[:200]}"
        )
        logger.warning(f"[evidence_prepare] JSON 解析失败: {e}, 原始内容: {text[:200]}")
        return []


def _fallback_extraction(
    tool_call_result: Dict[str, Any],
    tool_name: str = ""
) -> List[Dict[str, Any]]:
    """
    规则提取回退策略。

    当 LLM 提取失败时，从工具结果中直接提取实体信息。
    支持多种数据格式：nodes 数组、name 字段、list 字段等。

    Args:
        tool_call_result: 工具调用结果
        tool_name: 工具名称（如"获取agent详情"）

    Returns:
        提取到的实体列表
    """
    StandLogger.info_log("\n" + "="*60)
    StandLogger.info_log(f"[_fallback_extraction] START, tool_name={tool_name}")

    evidences = []

    if not isinstance(tool_call_result, dict):
        StandLogger.info_log("[_fallback_extraction] 工具结果不是字典，返回空")
        return evidences

    StandLogger.info_log(
        f"[_fallback_extraction] 工具结果 keys: {list(tool_call_result.keys())[:10]}"
    )

    # 1. 从 nodes 数组提取（原有逻辑）
    nodes = tool_call_result.get("nodes")
    if nodes and isinstance(nodes, list):
        StandLogger.info_log(f"[_fallback_extraction] 找到 {len(nodes)} 个 nodes")
        for idx, node in enumerate(nodes):
            if not isinstance(node, dict):
                continue

            node_name = node.get("object_type_name") or node.get("name", "")
            if not node_name:
                StandLogger.info_log(f"[_fallback_extraction] node {idx} 无 name，跳过")
                continue

            StandLogger.info_log(
                f"[_fallback_extraction] 提取 node {idx}: "
                f"name={node_name}, "
                f"type_id={node.get('object_type_id')}, "
                f"score={node.get('score', 0.7)}"
            )
            evidences.append({
                "object_type_name": str(node_name),
                "aliases": [],
                "object_type_id": str(node.get("object_type_id", "")),
                "score": float(node.get("score", 0.7)),
                "_source": "fallback_rules",
            })
    else:
        StandLogger.info_log("[_fallback_extraction] 未找到 nodes 字段")

        # 2. 从 name 字段提取主要实体（如 agent 名称）
        # 同时收集 profile、description 等描述性文本作为别名
        name = tool_call_result.get("name")
        aliases = []

        # 收集描述性文本作为别名
        for alias_key in ["profile", "description", "summary", "bio", "intro"]:
            alias_value = tool_call_result.get(alias_key)
            if alias_value and isinstance(alias_value, str) and len(alias_value) > 2:
                aliases.append(alias_value)
                StandLogger.info_log(
                    f"[_fallback_extraction] 从 {alias_key} 字段收集别名: {alias_value[:50]}..."
                )

        if name and isinstance(name, str):
            StandLogger.info_log(
                f"[_fallback_extraction] 从 name 字段提取: {name}, aliases={len(aliases)}"
            )
            evidences.append({
                "object_type_name": name,
                "aliases": aliases,
                "object_type_id": "ot_agent",
                "score": 0.9,
                "_source": "fallback_name_field",
            })

        # 3. 提取工具调用名称本身（如"获取agent详情"）
        # 因为 LLM 经常在回答中引用工具名称
        if tool_name and tool_name != "unknown":
            # 检查是否已经存在同名实体
            name_exists = any(
                ev.get("object_type_name") == tool_name for ev in evidences
            )
            if not name_exists:
                StandLogger.info_log(
                    f"[_fallback_extraction] 添加工具名称作为实体: {tool_name}"
                )
                evidences.append({
                    "object_type_name": tool_name,
                    "aliases": [],
                    "object_type_id": "ot_tool",
                    "score": 0.85,
                    "_source": "fallback_tool_name",
                })

        # 4. 从 list 字段提取（某些工具返回列表）
        data_list = tool_call_result.get("list")
        if data_list and isinstance(data_list, list):
            StandLogger.info_log(
                f"[_fallback_extraction] 找到 list 字段，包含 {len(data_list)} 项"
            )
            for idx, item in enumerate(data_list[:20]):  # 限制最多处理 20 项
                if not isinstance(item, dict):
                    continue

                item_name = item.get("name") or item.get("object_type_name", "")
                if not item_name:
                    continue

                StandLogger.info_log(
                    f"[_fallback_extraction] 从 list[{idx}] 提取: {item_name}"
                )
                evidences.append({
                    "object_type_name": str(item_name),
                    "aliases": [],
                    "object_type_id": str(item.get("object_type_id", "")),
                    "score": 0.75,
                    "_source": "fallback_list",
                })

        # 5. 从 data 字段提取（某些工具使用 data 包装）
        data = tool_call_result.get("data")
        if isinstance(data, list):
            StandLogger.info_log(
                f"[_fallback_extraction] 找到 data 数组，包含 {len(data)} 项"
            )
            for idx, item in enumerate(data[:20]):
                if not isinstance(item, dict):
                    continue

                item_name = item.get("name") or item.get("object_type_name", "")
                if not item_name:
                    continue

                StandLogger.info_log(
                    f"[_fallback_extraction] 从 data[{idx}] 提取: {item_name}"
                )
                evidences.append({
                    "object_type_name": str(item_name),
                    "aliases": [],
                    "object_type_id": str(item.get("object_type_id", "")),
                    "score": 0.75,
                    "_source": "fallback_data",
                })

    answer_data = tool_call_result.get("answer")
    if isinstance(answer_data, dict):
        answer_nodes = answer_data.get("nodes")
        if answer_nodes and isinstance(answer_nodes, list):
            StandLogger.info_log(f"[_fallback_extraction] 找到 {len(answer_nodes)} 个 answer.nodes")
            for idx, node in enumerate(answer_nodes):
                if not isinstance(node, dict):
                    continue

                name = node.get("object_type_name") or node.get("name", "")
                if not name:
                    continue

                StandLogger.info_log(
                    f"[_fallback_extraction] 提取 answer.node {idx}: "
                    f"name={name}, "
                    f"type_id={node.get('object_type_id')}"
                )
                evidences.append({
                    "object_type_name": str(name),
                    "aliases": [],
                    "object_type_id": str(node.get("object_type_id", "")),
                    "score": float(node.get("score", 0.7)),
                    "_source": "fallback_rules",
                })

    StandLogger.info_log(
        f"[_fallback_extraction] 完成, 共提取 {len(evidences)} 个实体\n"
        f"{'='*60}\n"
    )

    return evidences


def _generate_summary(evidences: List[Dict[str, Any]]) -> str:
    """
    生成证据摘要。

    Args:
        evidences: 证据列表

    Returns:
        摘要文本
    """
    if not evidences:
        return "无证据"

    names = [e.get("object_type_name", "") for e in evidences if e.get("object_type_name")]
    count = len(names)

    if count == 1:
        return f"{names[0]}等 1 个证据"
    elif count <= 5:
        return f"{', '.join(names[:5])}等 {count} 个证据"
    else:
        return f"{', '.join(names[:3])}等 {count} 个证据"


def _assign_local_ids(evidences: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """
    为每个证据分配 local_id。

    Args:
        evidences: 证据列表

    Returns:
        分配了 local_id 的证据列表
    """
    for idx, ev in enumerate(evidences):
        ev["local_id"] = f"e{idx + 1}"

    StandLogger.info_log(
        f"[_assign_local_ids] 为 {len(evidences)} 个证据分配 local_id"
    )

    return evidences


def _deduplicate_evidences(evidences: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """
    去重证据列表，按 object_type_name 去重，保留分数最高的。

    Args:
        evidences: 证据列表

    Returns:
        去重后的证据列表
    """
    if not evidences:
        return []

    original_count = len(evidences)
    seen = {}

    for ev in evidences:
        name = ev.get("object_type_name", "").lower()
        if not name:
            continue

        score = ev.get("score", 0)
        source = ev.get("_source", "unknown")

        if name not in seen:
            seen[name] = ev
        elif score > seen[name].get("score", 0):
            StandLogger.info_log(
                f"[_deduplicate_evidences] 替换 '{name}': "
                f"新分数 {score} > 旧分数 {seen[name].get('score', 0)}, "
                f"新来源 {source} > 旧来源 {seen[name].get('_source', 'unknown')}"
            )
            seen[name] = ev
        else:
            StandLogger.info_log(
                f"[_deduplicate_evidences] 跳过重复 '{name}': "
                f"分数 {score} <= {seen[name].get('score', 0)}"
            )

    deduplicated = list(seen.values())

    StandLogger.info_log(
        f"[_deduplicate_evidences] 去重: {original_count} -> {len(deduplicated)}"
    )

    return deduplicated


def _serialize_for_llm(tool_call_result: Dict[str, Any]) -> str:
    """
    将工具调用结果序列化为适合 LLM 处理的文本格式。

    Args:
        tool_call_result: 工具调用结果

    Returns:
        序列化后的文本
    """
    try:
        return json.dumps(tool_call_result, ensure_ascii=False, default=str)
    except (TypeError, ValueError):
        return str(tool_call_result)
