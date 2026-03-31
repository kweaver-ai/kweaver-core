from typing import Any, Dict, TYPE_CHECKING

from .agent_skills import process_skills_agents
from .mcp_skills import process_skills_mcps
from .tool_skills import process_skills_tools

if TYPE_CHECKING:
    from ...agent_core_v2 import AgentCoreV2


async def process_skills(
    ac: "AgentCoreV2",
    headers: Dict[str, str],
    context_variables: Dict,
    temp_files: Dict[str, Any] = None,
):
    """处理skills"""

    agent_config = ac.agent_config

    skills = agent_config.skills

    if temp_files is None:
        temp_files = {}

    # 1. tools skill
    await process_skills_tools(ac, context_variables, headers, skills)

    # 2. agents skill
    await process_skills_agents(ac, agent_config, skills, temp_files, headers)

    # 3. mcps skill
    await process_skills_mcps(skills)
