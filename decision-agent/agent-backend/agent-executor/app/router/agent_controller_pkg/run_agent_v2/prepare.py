import json
from typing import Dict, Tuple

from fastapi import Request

from app.common.stand_log import StandLogger
from app.common.errors import (
    ParamException,
    AgentPermissionException,
)
from app.driven.dip.agent_factory_service import agent_factory_service
from app.domain.enum.common.user_account_header_key import (
    set_biz_domain_id,
    set_user_account_id,
    set_user_account_type,
)
from app.domain.vo.agentvo import AgentConfigVo, AgentInputVo
from app.utils.observability.observability_log import get_logger as o11y_logger

from .process_options import process_options
from ..rdto.v2.req.run_agent import V2RunAgentReq


async def prepare(
    request: Request,
    req: V2RunAgentReq,
    account_id: str,
    account_type: str,
    biz_domain_id: str = "",
) -> Tuple[AgentConfigVo, AgentInputVo, Dict[str, str]]:
    """
    准备Agent运行所需的配置、输入和请求头

    Args:
        request: FastAPI请求对象
        req: 运行Agent的请求参数
        account_id: 账户ID
        account_type: 账户类型

    Returns:
        Tuple[AgentConfigVo, AgentInputVo, Dict[str, str]]: Agent配置、输入和请求头

    Raises:
        ParamException: 参数异常
        AgentPermissionException: Agent权限异常
    """
    agent_config: AgentConfigVo
    agent_input: AgentInputVo
    headers: Dict[str, str]

    # 0. 设置agent_factory_service的headers
    service_headers = {}
    set_user_account_id(service_headers, account_id)
    set_user_account_type(service_headers, account_type)
    set_biz_domain_id(service_headers, biz_domain_id)
    agent_factory_service.set_headers(service_headers)

    # 1. get agent config
    if req.agent_config:
        agent_config = req.agent_config

    elif req.agent_id:
        # 获取agent配置
        _agent_config = (
            await agent_factory_service.get_agent_config_by_agent_id_and_version(
                req.agent_id, req.agent_version
            )
        )

        agent_config_str = _agent_config["config"]

        if isinstance(agent_config_str, str):
            config_dict = json.loads(agent_config_str)
        else:
            config_dict = agent_config_str

        agent_config = AgentConfigVo(**config_dict)
    else:
        error_message = (
            "run_agent failed:At least one of agent_id or agent_config must be "
            f"provided, agent_id = {req.agent_id}, agent_config = {req.agent_config}"
        )
        StandLogger.error(error_message)
        o11y_logger().error(error_message)
        raise ParamException(
            "At least one of agent_id or agent_config must be provided."
        )

    # 2. set agent id
    if agent_config.agent_id is None:
        agent_config.agent_id = req.agent_id

    # 3. handle headers
    headers = dict(request.headers)

    # 4. get agent input
    agent_input = req.agent_input

    # 5. process options
    process_options(req.options, agent_config, agent_input)

    # 6. set agent version
    agent_config.agent_version = req.agent_version

    # 8. check agent permission
    if not await agent_factory_service.check_agent_permission(
        agent_config.agent_id, account_id, account_type
    ):
        error_message = (
            f"check_agent_permission failed: agent_id = {agent_config.agent_id}, "
            f"account_id = {account_id}, account_type = {account_type}"
        )
        StandLogger.error(error_message)
        o11y_logger().error(error_message)
        raise AgentPermissionException(agent_config.agent_id, account_id)

    return agent_config, agent_input, headers
