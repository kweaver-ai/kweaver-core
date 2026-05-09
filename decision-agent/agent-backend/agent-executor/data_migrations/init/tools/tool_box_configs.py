"""
内置工具箱配置
"""

from pathlib import Path

from app.common.config import Config

# 配置API和数据库连接信息
API_BASE_URL = "http://{host}:{port}/api/agent-operator-integration/internal-v1".format(
    host=Config.services.agent_operator_integration.host,
    port=Config.services.agent_operator_integration.port,
)

openapi_file_path = Path(__file__).parent / "openapi"

tool_box_configs = [
    {
        "box_id": "939bb1db-9239-43f5-8100-ba03eed683db",
        "box_name": "搜索工具",
        "box_desc": "提供智谱AI联网搜索能力，支持实时信息检索和知识查询",
        # "box_category": "data_query",
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto",
        "file_path": openapi_file_path / "search_tools.json",
        "content_type": "application/json",
    },
    {
        "box_id": "55cd6b6c-b546-4236-961a-4d09571fc931",
        "box_name": "记忆管理",
        "box_desc": "提供记忆构建与召回能力，支持会话上下文存储和智能检索",
        # "box_category": "data_analysis",
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto",
        "file_path": openapi_file_path / "agent_memory.yaml",
        "content_type": "application/yaml",
    },
    {
        "box_id": "91883b13-d5a6-f754-c90d-daf4ab416205",
        "box_name": "DataAgent配置相关工具",
        "box_desc": "获取Agent配置详情，支持查询Agent的完整配置信息",
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto",
        "file_path": openapi_file_path / "agent_config.yaml",
        "content_type": "application/yaml",
    },
    {
        "box_id": "bf0da1b2-e3b5-4bc5-83a2-ef0d3042ed83",
        "box_name": "联网搜索添加引用工具",
        "box_desc": "联网搜索并自动添加信息来源引用，提升回答可信度",
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto",
        "file_path": openapi_file_path / "online_search_cite_tools.json",
        "content_type": "application/json",
    },
    {
        "box_id": "abf13df1-9dec-03db-9733-7aee2febd1e2",
        "box_name": "沙箱代码执行工具",
        "box_desc": "在安全沙箱环境中执行Python代码并返回运行结果",
        "metadata_type": "openapi",
        "source": "internal",
        "config_version": "1.0.0",
        "config_source": "auto",
        "file_path": openapi_file_path / "sandbox-exec.yaml",
        "content_type": "application/yaml",
    },
]
