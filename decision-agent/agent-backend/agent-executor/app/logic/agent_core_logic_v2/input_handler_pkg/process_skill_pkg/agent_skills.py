from typing import Any, TYPE_CHECKING

from app.common.config import Config
from app.driven.dip.agent_factory_service import agent_factory_service
from app.domain.enum.common.user_account_header_key import (
    get_biz_domain_id,
    get_user_account_id,
    get_user_account_type,
    set_biz_domain_id,
    set_user_account_id,
    set_user_account_type,
)
from app.domain.vo.agentvo import AgentConfigVo
from app.domain.vo.agentvo.agent_config_vos import SkillVo, AgentSkillVo

if TYPE_CHECKING:
    from ...agent_core_v2 import AgentCoreV2


async def process_skills_agents(
    ac: "AgentCoreV2",
    agent_config_vo: AgentConfigVo,
    skills: SkillVo,
    temp_files: dict[str, Any],
    headers: dict[str, str],
):
    if not skills:
        return

    for agent_skill_cfg in skills.agents:
        # 0. 设置agent_factory_service的headers
        service_headers = {}
        user_id = get_user_account_id(headers) or ""
        visitor_type = get_user_account_type(headers) or ""
        biz_domain_id = get_biz_domain_id(headers) or ""
        set_user_account_id(service_headers, user_id)
        set_user_account_type(service_headers, visitor_type)
        set_biz_domain_id(service_headers, biz_domain_id)
        agent_factory_service.set_headers(service_headers)

        # 1. 从缓存中获取agent_info
        agent_info = ac.cache_handler.get_skill_agent_info_dict(
            agent_skill_cfg.agent_key
        )

        # 2. 如果缓存中没有agent_info, 则从agent_factory_service获取
        if not agent_info:
            agent_info = await agent_factory_service.get_agent_config_by_key(
                agent_skill_cfg.agent_key
            )
            # agent["agent_info"] = _agent_info

            if ac.is_warmup:
                ac.cache_handler.set_skill_agent_info_dict(
                    agent_skill_cfg.agent_key, agent_info
                )

        # 3. 设置agent_info到agent (使用inner_dto存储动态属性)
        agent_skill_cfg.inner_dto.agent_info = agent_info
        # agent_skill_cfg["agent_info"]["config"]["conversation_id"] = agent_config.conversation_id

        agent_skill_cfg.inner_dto.HOST_AGENT_EXECUTOR = (
            Config.services.agent_executor.host
        )
        agent_skill_cfg.inner_dto.PORT_AGENT_EXECUTOR = (
            Config.services.agent_executor.port
        )

        # 4. 设置agent_options到agent
        await hl_agent_options(agent_skill_cfg, agent_config_vo, temp_files)


async def hl_agent_options(
    agent_skill_cfg: AgentSkillVo,
    agent_config_vo: AgentConfigVo,
    temp_files: dict[str, Any],
):
    agent_options = {}

    # 1. 设置temp_files到agent_options
    for temp_file_name, temp_file_info in temp_files.items():
        if temp_file_info:
            agent_options["tmp_files"] = temp_file_info

    # 2. 设置data_source到agent_options
    # 使用datasource_config或data_source_config（兼容旧字段）
    data_source_config = (
        agent_skill_cfg.data_source_config or agent_skill_cfg.datasource_config
    )
    if data_source_config and data_source_config.type == "inherit_main":
        specific_inherit = (
            data_source_config.specific_inherit
            if data_source_config.specific_inherit
            else "all"
        )

        if specific_inherit == "docs_only":
            agent_options["data_source"] = {
                "doc": agent_config_vo.data_source.get("doc", []),
                "advanced_config": agent_config_vo.data_source.get(
                    "advanced_config", {}
                ),
            }
        elif specific_inherit == "graph_only":
            agent_options["data_source"] = {
                "kg": agent_config_vo.data_source.get("kg", []),
                "advanced_config": agent_config_vo.data_source.get(
                    "advanced_config", {}
                ),
            }
        elif specific_inherit == "all":
            agent_options["data_source"] = {
                "doc": agent_config_vo.data_source.get("doc", []),
                "kg": agent_config_vo.data_source.get("kg", []),
                "advanced_config": agent_config_vo.data_source.get(
                    "advanced_config", {}
                ),
            }

    # 3. 设置llm_config到agent_options
    if agent_skill_cfg.llm_config and agent_skill_cfg.llm_config.type == "inherit_main":
        for llm in agent_config_vo.llms or []:
            if llm.get("is_default"):
                agent_options["llm_config"] = llm["llm_config"]
                break

    # 4. 设置agent_options到agent (使用inner_dto存储动态属性)
    agent_skill_cfg.inner_dto.agent_options = agent_options
