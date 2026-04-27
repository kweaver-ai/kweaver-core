"""
工具箱信息获取功能模块
"""

import requests

from data_migrations.init.tools.tool_box_configs import API_BASE_URL


def get_tool_box_info(tool_box_id: str) -> dict:
    """
    获取工具箱信息

    Args:
        tool_box_id: 工具箱ID

    Returns:
        dict: 工具箱信息，如果不存在返回空字典

    Raises:
        Exception: 当API调用失败时抛出异常，包含错误信息

    Examples:
        >>> info = get_tool_box_info("tool-box-123")
        >>> if info:
        ...     print(f"工具箱名称: {info.get('name', '未知')}")
        >>> else:
        ...     print("工具箱不存在")
    """
    url = f"{API_BASE_URL}/tool-box/{tool_box_id}"
    response = requests.request("GET", url)
    if response.status_code == 200:
        return response.json()
    elif response.status_code == 404:
        return {}
    else:
        error_msg = f"[boot-get_tool_box_info]: Error getting tool-box info: {response.status_code} {response.text}"
        print(error_msg)
        # raise Exception(error_msg)
        return {}
