"""Dolphin 增强语法处理。"""

import re
from typing import Dict, Mapping, Optional, Tuple

from app.domain.vo.agentvo import AgentConfigVo
from app.domain.vo.agentvo.agent_config_vos import SkillVo

DolphinCallableKey = Tuple[str, str]

_DOLPHIN_ENHANCE_PATTERN = re.compile(
    r"@([^\s({]+)(?:\{(tool|agent):([^{}]+)\})?(?=\()"
)


def parse_dolphin_enhance_name_map(
    dolphin_enhance: Optional[str],
) -> Dict[DolphinCallableKey, str]:
    """解析 dolphin_enhance 中的 callable 标识和名称映射。"""
    name_map: Dict[DolphinCallableKey, str] = {}
    if not isinstance(dolphin_enhance, str) or not dolphin_enhance:
        return name_map

    for match in _DOLPHIN_ENHANCE_PATTERN.finditer(dolphin_enhance):
        callable_name, callable_type, callable_id = match.groups()
        if not callable_type or not callable_id:
            continue
        name_map[(callable_type, callable_id)] = callable_name

    return name_map


def build_latest_callable_name_map(
    skills: Optional[SkillVo],
) -> Dict[DolphinCallableKey, str]:
    """从已构建的 skills 中提取最新 callable 名称。"""
    latest_name_map: Dict[DolphinCallableKey, str] = {}
    if skills is None:
        return latest_name_map

    for tool in skills.tools or []:
        tool_info = tool.__dict__.get("tool_info") or {}
        tool_name = tool_info.get("name")
        if tool.tool_id and tool_name:
            latest_name_map[("tool", tool.tool_id)] = tool_name

    for agent in skills.agents or []:
        agent_info = agent.inner_dto.agent_info or {}
        agent_name = agent_info.get("name")
        if agent.agent_key and agent_name:
            latest_name_map[("agent", agent.agent_key)] = agent_name

    return latest_name_map


def materialize_dolphin_from_enhance(
    dolphin: Optional[str],
    dolphin_enhance: Optional[str],
    latest_name_map: Mapping[DolphinCallableKey, str],
) -> str:
    """根据 dolphin_enhance 和最新技能名称生成普通 dolphin 语句。"""
    dolphin_text = dolphin if isinstance(dolphin, str) else ""
    if not isinstance(dolphin_enhance, str) or not dolphin_enhance:
        return dolphin_text

    def replace_callable(match: re.Match[str]) -> str:
        callable_name, callable_type, callable_id = match.groups()
        if not callable_type or not callable_id:
            return f"@{callable_name}"

        latest_name = latest_name_map.get((callable_type, callable_id))
        return f"@{latest_name or callable_name}"

    return _DOLPHIN_ENHANCE_PATTERN.sub(replace_callable, dolphin_enhance)


def refresh_dolphin_from_enhance(config: AgentConfigVo) -> None:
    """根据 dolphin_enhance 刷新 AgentConfigVo.dolphin。"""
    if not config.dolphin_enhance:
        return

    latest_name_map = build_latest_callable_name_map(config.skills or SkillVo())
    saved_name_map = parse_dolphin_enhance_name_map(config.dolphin_enhance)

    has_name_change = any(
        latest_name_map.get(callable_key)
        and latest_name_map[callable_key] != saved_name
        for callable_key, saved_name in saved_name_map.items()
    )

    if has_name_change or not config.dolphin:
        config.dolphin = materialize_dolphin_from_enhance(
            config.dolphin,
            config.dolphin_enhance,
            latest_name_map,
        )
