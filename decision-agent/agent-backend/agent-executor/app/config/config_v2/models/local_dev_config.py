"""
本地开发相关配置
"""

from dataclasses import dataclass
from typing import List, Optional


@dataclass
class LocalDevConfig:
    """本地开发相关配置"""

    # 是否跳过权限检查
    is_skip_pms_check: bool = False

    # 是否使用外部LLM
    is_use_outer_llm: bool = False

    # 是否mock get_llm_config接口的返回值
    is_mock_get_llm_config_resp: bool = False

    # 是否不初始化内置agent和tool
    do_not_init_built_in_agent_and_tool: bool = False

    # dolphin agent输出变量
    dolphin_agent_output_variables: Optional[List[str]] = None

    # 是否启用流式响应速率限制（实验性功能，目前好像有问题，仅local_dev使用）
    enable_streaming_response_rate_limit: bool = False

    # 是否在启动时显示配置信息
    is_show_config_on_boot: bool = False

    @classmethod
    def from_dict(cls, data: dict) -> "LocalDevConfig":
        """从字典创建配置对象"""
        return cls(
            is_skip_pms_check=data.get("is_skip_pms_check", False),
            is_use_outer_llm=data.get("is_use_outer_llm", False),
            is_mock_get_llm_config_resp=data.get("is_mock_get_llm_config_resp", False),
            do_not_init_built_in_agent_and_tool=data.get(
                "do_not_init_built_in_agent_and_tool", False
            ),
            dolphin_agent_output_variables=data.get(
                "dolphin_agent_output_variables", None
            ),
            enable_streaming_response_rate_limit=data.get(
                "enable_streaming_response_rate_limit", False
            ),
            is_show_config_on_boot=data.get("is_show_config_on_boot", False),
        )
