import re
from typing import Dict, Optional, TYPE_CHECKING

from dolphin.core.utils.tools import Tool
from app.common.stand_log import StandLogger
from app.domain.vo.agentvo.agent_config_vos import SkillVo
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

# Import from sub-modules using relative imports
from .api_tool import APITool
from .agent_tool import AgentTool
from .mcp_tool import get_mcp_tools


if TYPE_CHECKING:
    from ...logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2


# DeepSeek tool name 规则：只能包含 a-z, A-Z, 0-9, _ 和 -，最大长度 64
TOOL_NAME_PATTERN = re.compile(r'^[a-zA-Z0-9_-]+$')
TOOL_NAME_MAX_LENGTH = 64


def validate_tool_name(name: str) -> bool:
    """
    校验工具名称是否符合 DeepSeek 的 tool naming 规范。

    Args:
        name: 工具名称

    Returns:
        bool: 是否符合规范
    """
    if not name:
        return False
    if len(name) > TOOL_NAME_MAX_LENGTH:
        return False
    return bool(TOOL_NAME_PATTERN.match(name))


@internal_span()
async def build_tools(
    ac: "AgentCoreV2",
    skills: SkillVo,
    request_headers: Optional[Dict[str, str]] = None,
    span: Span = None,
) -> Dict[str, Tool]:
    """构建 Agent 运行时工具字典。

    只处理 Agent 配置中 skills 字段声明的工具（API 工具、Agent 工具、MCP 工具）。
    平台内置 Skill 工具（builtin_skill_load 等）不在此处注入，
    由 run_dolphin.py 在 skill_enabled=True 时单独合并进 tool_dict。
    """
    """
    skills样例:

    {
        "tools": [
            {
                "tool_id": "tool_id",
                "tool_box_id": "tool_box_id",
                "tool_input" :{

                },
                "intervention": false,
                "tool_info": "$ref: #/components/schemas/ToolInfo" # cursor_references/toolbox.yaml
            },
        ],
        "agents": [
            {
                "agent_key": "agent_key",
                "agent_version": "agent_version",
                "intervention": false,
                "agent_input:{},
                "intervention": false,
                "data_source": {},
                "llm_config": {},
                "agent_info":{
                    "name": "agent_name",
                    "profile": "agent_profile",
                    "config":{
                        "input":{
                            "fields":[
                                {
                                    "name": "",
                                    "type": "",
                                    "desc": "",
                                }
                            ]
                        }
                    }
                }
            }
        ],
        "mcps": [
            {
                "mcp_server_id": "mcp_server_id",
                "mcp_tool_id": "mcp_tool_id",
                "intervention": false,
                "mcp_info": {} # cursor_references/mcp.yaml
            }
        ]
    }
    """
    tools: Dict[str, Tool] = {}  # tool_name: Tool
    invalid_tool_names = []  # 记录不合规的工具名称

    # 处理API工具
    if skills.tools:
        for tool in skills.tools:
            # 使用__dict__访问动态添加的属性
            tool_dict = tool.__dict__
            tool_info = tool_dict.get("tool_info", {})

            tool_name = tool_info.get(
                "name", f"tool_{tool_dict.get('tool_id', 'unknown')}"
            )

            # 校验工具名称是否符合规范
            if not validate_tool_name(tool_name):
                invalid_tool_names.append(tool_name)
                StandLogger.warn(
                    f"Tool name '{tool_name}' does not conform to DeepSeek naming rules "
                    f"(only a-z, A-Z, 0-9, _, - allowed, max 64 chars). "
                    f"This may cause tool calling failures with DeepSeek models."
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

            # 校验工具名称是否符合规范
            if not validate_tool_name(agent_name):
                invalid_tool_names.append(agent_name)
                StandLogger.warn(
                    f"Agent tool name '{agent_name}' does not conform to DeepSeek naming rules "
                    f"(only a-z, A-Z, 0-9, _, - allowed, max 64 chars). "
                    f"This may cause tool calling failures with DeepSeek models."
                )

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
                # 校验工具名称是否符合规范
                if not validate_tool_name(tool_name):
                    invalid_tool_names.append(tool_name)
                    StandLogger.warn(
                        f"MCP tool name '{tool_name}' does not conform to DeepSeek naming rules "
                        f"(only a-z, A-Z, 0-9, _, - allowed, max 64 chars). "
                        f"This may cause tool calling failures with DeepSeek models."
                    )

                if tool_name in tools:
                    StandLogger.warn(
                        f"mcp_tool_name {tool_name} already exists, use mcp_tool_id {tool.name} instead"
                    )

            tools.update(mcp_tools)

    # 汇总记录不合规的工具名称
    if invalid_tool_names:
        StandLogger.warn(
            f"[build_tools] Found {len(invalid_tool_names)} tool name(s) not conforming to DeepSeek naming rules: "
            f"{invalid_tool_names}. These tools may fail when used with DeepSeek models."
        )

    return tools
