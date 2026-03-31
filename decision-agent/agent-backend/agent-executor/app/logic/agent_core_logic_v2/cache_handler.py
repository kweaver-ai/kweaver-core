from typing import TYPE_CHECKING, Any, Dict

from app.domain.vo.agent_cache.cache_data_vo import CacheDataVo

if TYPE_CHECKING:
    from .agent_core_v2 import AgentCoreV2


class CacheHandler:
    """缓存处理类"""

    agent_core: "AgentCoreV2"

    def __init__(self, agent_core: "AgentCoreV2"):
        self.agent_core = agent_core
        self._cache_data = CacheDataVo()

        self.enable_dependency_cache = agent_core.run_options_vo.enable_dependency_cache

    def set_cache_data(self, cache_data: CacheDataVo) -> None:
        self._cache_data = cache_data

    def get_cache_data(self) -> CacheDataVo:
        return self._cache_data

    # LLM配置相关方法
    def set_llm_config(self, llm_id: str, llm_config: Dict[str, Any]) -> None:
        self._cache_data.llm_config_dict[llm_id] = llm_config

    def get_llm_config(self, llm_id: str) -> Dict[str, Any]:
        return self._cache_data.llm_config_dict.get(llm_id)

    # Agent配置相关方法
    def set_agent_config(self, agent_config: Dict[str, Any]) -> None:
        self._cache_data.agent_config = agent_config

    def get_agent_config(self) -> Dict[str, Any]:
        return self._cache_data.agent_config

    # 工具信息相关方法
    def set_tools_info_dict(self, tool_id: str, tool_info: Dict[str, Any]) -> None:
        self._cache_data.tools_info_dict[tool_id] = tool_info

    def get_tools_info_dict(self, tool_id: str) -> Dict[str, Any]:
        return self._cache_data.tools_info_dict.get(tool_id)

    # SkillAgent信息相关方法
    def set_skill_agent_info_dict(
        self, skill_agent_key: str, skill_agent_info: Dict[str, Any]
    ) -> None:
        self._cache_data.skill_agent_info_dict[skill_agent_key] = skill_agent_info

    def get_skill_agent_info_dict(self, skill_agent_key: str) -> Dict[str, Any]:
        return self._cache_data.skill_agent_info_dict.get(skill_agent_key)
