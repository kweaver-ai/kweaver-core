class BuiltinIdsConfig:
    """内置Agent和工具的ID配置类"""

    def __init__(self):
        # 内置Agent的ID配置
        # 这些ID在执行初始化脚本后会得到具体的值
        self.agent_ids = {
            "deepsearch": "deepsearch",
            "OnlineSearch_Agent": "OnlineSearch_Agent",
            "Plan_Agent": "Plan_Agent",
            "SimpleChat_Agent": "SimpleChat_Agent",
            "Summary_Agent": "Summary_Agent",
        }

        # 内置工具的ID配置
        self.tool_ids = {
            "zhipu_search_tool": "zhipu_search_tool",
            "check": "check",
            "pass": "pass",
            "search_file_snippets": "search_file_snippets",
            "get_file_full_content": "get_file_full_content",
            "process_file_intelligent": "process_file_intelligent",
            "get_file_download_url": "get_file_download_url",
            "获取agent详情": "获取agent详情",
            "online_search_cite_tool": "online_search_cite_tool",
            "查询可观测数据": "查询可观测数据",
        }

        self.tool_box_ids = {
            "搜索工具": "搜索工具",
            "数据处理工具": "数据处理工具",
            "文件处理工具": "文件处理工具",
            "记忆管理": "记忆管理工具",
            "DataAgent配置相关工具": "DataAgent配置工具",
            "联网搜索添加引用工具": "联网搜索添加引用工具",
            "Agent可观测数据查询API": "Agent可观测数据查询API",
        }

    def get_agent_id(self, agent_name):
        """获取指定Agent的ID"""
        return self.agent_ids.get(agent_name, agent_name)

    def get_tool_id(self, tool_name):
        """获取指定工具的ID"""
        return self.tool_ids.get(tool_name, tool_name)

    def get_tool_box_id(self, tool_box_name):
        """获取指定工具箱的ID"""
        return self.tool_box_ids.get(tool_box_name, tool_box_name)

    def set_agent_id(self, agent_name, agent_id):
        """设置指定Agent的ID"""
        self.agent_ids[agent_name] = agent_id

    def set_tool_id(self, tool_name, tool_id):
        """设置指定工具的ID"""
        self.tool_ids[tool_name] = tool_id

    def set_tool_box_id(self, tool_box_name, tool_box_id):
        """设置指定工具箱的ID"""
        self.tool_box_ids[tool_box_name] = tool_box_id

    def get_all_agent_ids(self):
        """获取所有Agent的ID"""
        return self.agent_ids.copy()

    def get_all_tool_ids(self):
        """获取所有工具的ID"""
        return self.tool_ids.copy()

    def get_all_tool_box_ids(self):
        """获取所有工具箱的ID"""
        return self.tool_box_ids.copy()
