"""
Agent发布功能模块
"""

import requests


def publish_agent(agent_id, api_base_url, headers, user_id, description="") -> None:
    """
    发布agent

    Args:
        agent_id: agent ID
        api_base_url: API基础URL
        headers: 请求头
        user_id: 用户ID
        description: 描述信息

    Returns:
        None: 无返回值，仅执行发布操作

    Note:
        如果agent未发布过，会跳过发布操作
        如果发布失败，会打印错误信息但不会抛出异常

    Examples:
        >>> publish_agent("agent-123", api_base_url, headers, user_id, "描述信息")
        >>> print("Agent发布完成")
    """

    # 1. 检查agent是否已经发布
    url = f"{api_base_url}/agent/{agent_id}/publish-info"
    response = requests.get(url, verify=False, headers=headers)
    # 1.1. agent未发布
    if response.status_code == 404:
        print(f"Agent {agent_id} did not publish yet, skip")
        return
    # 1.2. 获取agent发布信息失败
    if response.status_code // 100 != 2:
        error_msg = (
            f"Error getting agent publish info: {response.status_code} {response.text}"
        )
        raise Exception(error_msg)
    # 1.3. agent已发布
    publish_info = response.json()

    # 2. 发布agent
    publish_info["user_id"] = user_id
    url = f"{api_base_url}/agent/{agent_id}/publish"
    response = requests.post(url, json=publish_info, verify=False, headers=headers)

    # 2.1. 发布失败
    if response.status_code // 100 != 2:
        error_msg = f"Error publishing agent: {response.status_code} {response.text}"
        raise Exception(error_msg)
    # 2.2. 发布成功
    print(f"Agent {agent_id} published successfully")
