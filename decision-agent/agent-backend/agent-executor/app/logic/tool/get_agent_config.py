from app.driven.dip.agent_factory_service import agent_factory_service


async def get_agent_config(agent_id=None, agent_key=None):
    """
    获取agent配置详情

    Args:
        agent_id: agent的唯一标识ID
        agent_key: agent的密钥标识

    Returns:
        dict: agent配置详情
    """
    if not agent_id and not agent_key:
        raise ValueError("必须提供agent_id或agent_key参数")

    if agent_id and agent_key:
        raise ValueError("agent_id和agent_key不能同时提供")

    if agent_id:
        # 通过agent_id获取配置
        config = await agent_factory_service.get_agent_config(agent_id)
    else:
        # 通过agent_key获取配置
        config = await agent_factory_service.get_agent_config_by_key(agent_key)

    return config
