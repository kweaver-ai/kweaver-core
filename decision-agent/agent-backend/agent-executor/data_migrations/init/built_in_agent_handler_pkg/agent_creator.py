"""
Agent创建功能模块
"""

import requests


def create_agent(agent_data, api_base_url, headers, user_id, update_mode=False) -> str:
    """
    创建或更新agent

    Args:
        agent_data: agent配置数据
        api_base_url: API基础URL
        headers: 请求头
        user_id: 用户ID
        update_mode: 是否为更新模式

    Returns:
        str: agent ID - 创建或更新后的agent唯一标识符

    Raises:
        Exception: 当API调用失败时抛出异常，包含错误信息

    Examples:
        >>> agent_id = create_agent(agent_data, api_base_url, headers, user_id, update_mode=True)
        >>> print(f"Agent created/updated with ID: {agent_id}")
    """
    # 1. 先检查agent是否存在
    get_by_key_url = f"{api_base_url}/agent/by-key/{agent_data['name']}"
    response = requests.get(get_by_key_url, verify=False, headers=headers)
    agent_config = {}

    # 2. 初始化
    agent_id = None

    if response.status_code == 200:
        response_data = response.json()
        agent_id = response_data["id"]
        agent_config = response_data["config"]
    elif response.status_code == 404:
        agent_id = None
    else:
        error_msg = f"Error getting agent: {response.status_code} {response.text}"
        raise Exception(error_msg)

    # 3. 如果不存在则创建
    if not agent_id:
        print(f"Agent {agent_data['name']} not found, creating...")
        create_url = f"{api_base_url}/agent"

        create_data = agent_data.copy()
        create_data["profile"] = create_data.pop("description", "")
        create_data["key"] = create_data["name"]
        create_data["is_built_in"] = 1
        create_data["created_by"] = user_id
        create_data["updated_by"] = user_id
        create_data["config"]["llms"] = agent_config.get("llms", [])
        create_data["config"]["data_source"] = agent_config.get("data_source", {})

        response = requests.post(
            create_url, json=create_data, verify=False, headers=headers
        )

        if response.status_code // 100 == 2:
            agent_id = response.json()["id"]
            return agent_id
        else:
            error_msg = f"Error creating agent: {response.status_code} {response.text}"
            raise Exception(error_msg)
    # 4. 如果存在则更新
    else:
        if not update_mode:
            print(f"Agent {agent_data['name']} already exists, skipping update")
            return agent_id

        print(f"Agent {agent_data['name']} already exists, start updating...")
        update_url = f"{api_base_url}/agent/{agent_id}"

        # 4.1. 需要更新的部分
        update_data = agent_data.copy()
        update_data["profile"] = update_data.pop("description", "")
        update_data["updated_by"] = user_id
        update_data["is_built_in"] = 1

        # 4.1.1 保留部分配置
        update_data["config"]["llms"] = agent_config.get("llms", [])
        update_data["config"]["data_source"] = agent_config.get("data_source", {})

        # 4.1.2 工具列表：tool_input保留旧的配置，其余更新
        old_tool_inputs = {}
        new_tool_inputs = (
            agent_data.get("config", {}).get("skills", {}).get("tools", [])
        )

        for tool in (agent_config.get("skills") or {}).get("tools", []):
            old_tool_inputs[tool["tool_id"]] = tool.get("tool_input", [])

        for tool in new_tool_inputs:
            # 如果旧的工具输入存在，则保留旧的工具输入
            if old_tool_inputs.get(tool["tool_id"]):
                tool["tool_input"] = old_tool_inputs[tool["tool_id"]]

        update_data["config"]["skills"]["tools"] = new_tool_inputs

        # 4.2. 更新
        response = requests.put(
            update_url, json=update_data, verify=False, headers=headers
        )

        if response.status_code // 100 == 2:
            return agent_id
        else:
            error_msg = f"Error updating agent: {response.status_code} {response.text}"
            raise Exception(error_msg)
