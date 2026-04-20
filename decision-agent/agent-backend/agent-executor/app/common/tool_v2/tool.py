from typing import Dict, Optional, TYPE_CHECKING, Tuple

from dolphin.core.utils.tools import Tool
from app.common.stand_log import StandLogger
from app.domain.vo.agentvo.agent_config_vos import SkillVo
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

# Import from sub-modules using relative imports
from .api_tool import APITool
from .agent_tool import AgentTool
from .mcp_tool import get_mcp_tools
from .skill_contract_tools import build_builtin_skill_tools


if TYPE_CHECKING:
    from ...logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2

TOOL_TYPE_TOOL = "tool"
TOOL_TYPE_AGENT = "agent"
TOOL_TYPE_MCP = "mcp"

TypedToolKey = Tuple[str, str]

@internal_span()
async def _build_tools_legacy(
    ac: "AgentCoreV2",
    skills: SkillVo,
    request_headers: Optional[Dict[str, str]] = None,
    span: Span = None,
) -> Dict[str, Tool]:
    """旧实现：工具字典 key 为工具名称 / agent 名称 / mcp tool 名称。

    对应 agent_config.is_use_tool_id_in_dolphin == 0 的情况。
    """
    tools: Dict[str, Tool] = {}  # tool_name: Tool

    # 处理API工具
    if skills.tools:
        for tool in skills.tools:
            # 使用__dict__访问动态添加的属性
            tool_dict = tool.__dict__
            tool_info = tool_dict.get("tool_info", {})

            tool_name = tool_info.get(
                "name", f"tool_{tool_dict.get('tool_id', 'unknown')}"
            )

            if tool_name in tools:
                StandLogger.warn(
                    f"tool_name {tool_name} already exists, use tool_id {tool_dict.get('tool_id')} instead"
                )

            tools[tool_name] = APITool(tool_info, tool_dict)

    # 处理Agent工具
    if skills.agents:
        for agent in skills.agents:
            # 从inner_dto获取agent_info
            agent_info = agent.inner_dto.agent_info or {}

            agent_name = agent_info.get("name", f"agent_{agent.agent_key}")

            if agent_name in tools:
                StandLogger.warn(
                    f"agent_name {agent_name} already exists, use agent_key {agent.agent_key} instead"
                )

            tools[agent_name] = AgentTool(ac, agent)

    # 处理MCP工具
    if skills.mcps:
        for mcp in skills.mcps:
            # 使用__dict__访问动态添加的属性
            mcp_dict = mcp.__dict__
            mcp_tools = await get_mcp_tools(mcp_dict)

            for tool_name, tool in mcp_tools.items():
                if tool_name in tools:
                    StandLogger.warn(
                        f"mcp_tool_name {tool_name} already exists, use mcp_tool_id {tool.name} instead"
                    )

            tools.update(mcp_tools)

    # Inject the three built-in skill contract tools.
    # These are appended last so that they always overwrite any user-defined
    # tool with the same reserved name (builtin_skill_load, builtin_skill_read_file,
    # builtin_skill_execute_script).
    # request_headers are forwarded so each Tool instance captures them at
    # construction time, avoiding set_headers() races on the shared singleton.
    builtin_skill_tools = build_builtin_skill_tools(request_headers)
    for name in builtin_skill_tools:
        if name in tools:
            StandLogger.warn(
                f"User-defined tool '{name}' is overridden by the built-in skill contract tool."
            )
    tools.update(builtin_skill_tools)

    return tools



@internal_span()
async def _build_tools_typed(
        ac: "AgentCoreV2",
        skills: SkillVo,
        request_headers: Optional[Dict[str, str]] = None,
        span: Span = None,
) -> Dict[str, Tool]:
    """新实现：工具字典 key 为 (tool_type, tool_id) 元组。

    对应 agent_config.is_use_tool_id_in_dolphin == 1 的情况。
    使用稳定的 tool_id 代替名称，避免工具名称变更导致 Dolphin 引用失效。
    """
    tools: Dict[TypedToolKey, Tool] = {}  # (tool_type, tool_id): Tool


    # 处理API工具
    if skills.tools:
        for tool in skills.tools:
            # 使用__dict__访问动态添加的属性
            tool_dict = tool.__dict__
            tool_info = tool_dict.get("tool_info", {})
            tool_id = tool_dict.get("tool_id")

            if not tool_id:
                StandLogger.warn(f"tool_id is empty for tool {tool_dict}")
                continue

            tool_id = str(tool_id)


            tools[(TOOL_TYPE_TOOL, tool_id)] = APITool(tool_info, tool_dict)

    # 处理Agent工具
    if skills.agents:
        for agent in skills.agents:
            tools[(TOOL_TYPE_AGENT, agent.agent_key)] = AgentTool(ac, agent)

    # 处理MCP工具
    if skills.mcps:
        for mcp in skills.mcps:
            # 使用__dict__访问动态添加的属性
            mcp_dict = mcp.__dict__
            mcp_tools = await get_mcp_tools(mcp_dict)

            mcp_tools_map: Dict[TypedToolKey, Tool] = {}

            for tool_name, tool in mcp_tools.items():

                mcp_tools_map[(TOOL_TYPE_MCP, tool_name)] = tool
                # if tool_name in tools:
                #     StandLogger.warn(
                #         f"mcp_tool_name {tool_name} already exists, use mcp_tool_id {tool.name} instead"
                #     )

            tools.update(mcp_tools_map)

    # Inject the three built-in skill contract tools.
    # These are appended last so that they always overwrite any user-defined
    # tool with the same reserved name (builtin_skill_load, builtin_skill_read_file,
    # builtin_skill_execute_script).
    # request_headers are forwarded so each Tool instance captures them at
    # construction time, avoiding set_headers() races on the shared singleton.
    builtin_skill_tools = build_builtin_skill_tools(request_headers)
    for name in builtin_skill_tools:
        if name in tools:
            StandLogger.warn(
                f"User-defined tool '{name}' is overridden by the built-in skill contract tool."
            )
    tools.update(builtin_skill_tools)

    return tools


@internal_span()
async def build_tools(
    ac: "AgentCoreV2",
    skills: SkillVo,
    request_headers: Optional[Dict[str, str]] = None,
    span: Span = None,
) -> Dict[str, Tool]:
    """工具字典构造入口。

    根据 ``ac.agent_config.is_use_tool_id_in_dolphin`` 选择实现：
    - ``0`` (默认): 走旧实现 ``_build_tools_legacy``，key 为工具名称。
    - ``1``: 走新实现 ``_build_tools_typed``，key 为 ``(tool_type, tool_id)``。

    skills 样例参见 ``SkillVo``。
    """
    use_typed = (
        ac.agent_config is not None and ac.agent_config.use_tool_id_in_dolphin()
    )

    if use_typed:
        return await _build_tools_typed(
            ac, skills, request_headers=request_headers
        )

    return await _build_tools_legacy(
        ac, skills, request_headers=request_headers
    )
