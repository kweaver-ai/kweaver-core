from typing import Dict, TYPE_CHECKING

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


@internal_span()
async def build_tools(
    ac: "AgentCoreV2", skills: SkillVo, span: Span = None
) -> Dict[str, Tool]:
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

    return tools
