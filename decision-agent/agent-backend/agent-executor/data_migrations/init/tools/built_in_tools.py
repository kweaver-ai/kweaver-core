import sys
from pathlib import Path

sys.path.insert(0, Path(__file__).parent.parent.parent.parent.as_posix())

from app.common.config import BuiltinIds
from data_migrations.init.tools.tool_box_configs import tool_box_configs
from data_migrations.init.tools.tool_box_info import get_tool_box_info
from data_migrations.init.tools.tool_box_creator import add_tool_box


def manage_built_in_tools(update_mode: bool = False) -> None:
    """
    管理内置工具箱的创建和更新

    Args:
        update_mode: 是否为更新模式，默认为False

    Returns:
        None: 无返回值

    Note:
        - 初始化模式：只添加不存在的工具箱，跳过已存在的
        - 更新模式：更新所有工具箱配置
        - 处理完成后会更新BuiltinIds中的工具箱和工具ID
    """

    for tool_box_config in tool_box_configs:
        # 1. 获取tool-box信息
        tool_box_info = get_tool_box_info(tool_box_config["box_id"])

        # 2. tool-box已存在
        if tool_box_info:
            if update_mode:
                print(f"Update built-in tool-box {tool_box_config['box_name']}")
                tool_box_info = add_tool_box(tool_box_config)
            else:
                print(f"Skip built-in tool-box {tool_box_config['box_name']}")

        # 2. tool-box不存在
        else:
            print(f"Add built-in tool-box {tool_box_config['box_name']}")
            tool_box_info = add_tool_box(tool_box_config)

        # 3. 更新id
        # 3.1. 更新tool-box id
        BuiltinIds.set_tool_box_id(
            tool_box_config["box_name"], tool_box_config["box_id"]
        )

        # 3.2. 更新tool id
        for tool in tool_box_info.get("tools", []):
            BuiltinIds.set_tool_id(tool["name"], tool["tool_id"])


if __name__ == "__main__":
    manage_built_in_tools()
