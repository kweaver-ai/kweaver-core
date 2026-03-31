from app.common.config import Config
from app.domain.vo.agentvo.agent_config_vos import SkillVo


async def process_skills_mcps(skills: SkillVo | None):
    if not skills:
        return

    # mcp_tool_list = {}  # mcp_server_id -> tool_list
    for mcp in skills.mcps:
        mcp.__dict__["HOST_AGENT_OPERATOR"] = (
            Config.services.agent_operator_integration.host
        )
        mcp.__dict__["PORT_AGENT_OPERATOR"] = (
            Config.services.agent_operator_integration.port
        )
