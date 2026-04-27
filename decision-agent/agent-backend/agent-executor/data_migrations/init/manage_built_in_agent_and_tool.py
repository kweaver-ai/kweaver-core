"""
本脚本用于管理内置agent和工具，支持两种运行模式：

1. 初始化模式（默认）
   命令：python3 manage_built_in_agent_and_tool.py
   功能说明：
   - 首次部署时：执行初始化操作，创建内置Agent（不包含LLM配置和数据源配置）
   - 已有Agent时：保持现有配置不变，不执行更新
   注意：初始化后需要管理员在页面上手动配置LLM和数据源，并发布Agent

2. 更新模式
   命令：python3 manage_built_in_agent_and_tool.py --update
   功能说明：
   - 首次部署时：与初始化模式行为一致
   - 已有Agent时：更新至最新配置并重新发布
   更新范围：
   - 更新内容：Agent配置、工具列表
   - 保留内容：LLM配置、数据源配置、工具参数（如智谱搜索的API_KEY）
"""

import argparse
import os
import sys
from datetime import datetime
import time

import urllib3

sys.path.insert(
    0, os.path.dirname(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))
)

from app.common.config import BuiltinIds, Config
from data_migrations.init.tools.built_in_tools import manage_built_in_tools
from app.domain.enum.common.user_account_header_key import (
    set_user_account_id,
    set_user_account_type,
)
from app.domain.enum.common.user_account_type import UserAccountType
from data_migrations.init.built_in_agent_handler_pkg.agent_creator import create_agent
from data_migrations.init.built_in_agent_handler_pkg.agent_publisher import (
    publish_agent,
)
from data_migrations.init.built_in_agent_handler_pkg.config_loader import (
    load_agent_config,
)

# 禁用SSL警告
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


# 解析命令行参数
def parse_args():
    parser = argparse.ArgumentParser(description="管理内置agent和工具")
    parser.add_argument(
        "-u", "--update", action="store_true", help="启用更新模式，默认不更新"
    )
    return parser.parse_args()


args = parse_args()
UPDATE_MODE = args.update
# 配置API和数据库连接信息
API_BASE_URL = f"http://{Config.services.agent_factory.host}:{Config.services.agent_factory.port}/api/agent-factory/internal/v3"

user_id = "266c6a42-6131-4d62-8f39-853e7093701c"  # admin
# user_id = "234562BE-88FF-4440-9BFF-447F139871A2"  # system
headers = {}
set_user_account_id(headers, user_id)
set_user_account_type(headers, UserAccountType.USER.value)
now_time = datetime.now().strftime("%Y-%m-%d %H:%M:%S")


def process_agents(UPDATE_MODE):
    """
    处理内置agent的创建、更新和发布

    Args:
        UPDATE_MODE: 是否为更新模式
    """
    agents_dir = os.path.join(os.path.dirname(__file__), "agents")

    for filename in os.listdir(agents_dir):
        if filename.endswith(".py") and not filename.startswith("__"):
            file_path = os.path.join(agents_dir, filename)

            # 1. 加载agent配置
            agent_data = load_agent_config(file_path)

            # 2. 创建或更新agent
            agent_id = create_agent(
                agent_data, API_BASE_URL, headers, user_id, UPDATE_MODE
            )
            BuiltinIds.set_agent_id(agent_data["name"], agent_id)
            agent_data["id"] = agent_id
            agent_data["version"] = "latest"

            # 3. 发布agent
            if UPDATE_MODE:
                publish_agent(
                    agent_id, API_BASE_URL, headers, user_id, agent_data["description"]
                )


# 现在的处理逻辑：
# 1.服务启动时先进行内置agent和工具的相关处理，处理成功再启动http server
# 2.如果失败会进行重试（重试时间间隔：1、3、5、10、20）
# 3.如果都重试失败，退出程序
def main():
    """执行管理任务，并在失败时按照退避策略重试"""
    retry_intervals = [1, 3, 5, 10, 20]  # 秒
    for idx, interval in enumerate(retry_intervals):
        try:
            print(
                f"运行模式: {'更新' if UPDATE_MODE else '初始化'}（第 {idx + 1} 次尝试）"
            )
            # 1. 插入内置工具
            manage_built_in_tools(UPDATE_MODE)

            # 2. 插入内置agent
            process_agents(UPDATE_MODE)

            print("操作完成！")
            return
        except Exception as e:
            from app.common.struct_logger import struct_logger

            struct_logger.console_logger.error("管理内置agent和工具失败", exc_info=e)

            if idx < len(retry_intervals) - 1:
                print(f"{interval} 秒后重试...（已重试 {idx + 1} 次）")
                time.sleep(interval)
            else:
                print("已重试 5 次仍然失败，程序退出。")
                sys.exit(1)


if __name__ == "__main__":
    main()
