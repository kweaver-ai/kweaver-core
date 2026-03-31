"""单元测试 - logic/agent_core_logic_v2/cache_handler 模块"""

from unittest.mock import MagicMock


class TestCacheHandler:
    """测试 CacheHandler 类"""

    def test_init_basic(self):
        """测试基本初始化"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        assert handler.agent_core == mock_agent_core
        assert handler.enable_dependency_cache is True
        assert handler.get_cache_data() is not None

    def test_init_with_dependency_cache_disabled(self):
        """测试禁用依赖缓存的初始化"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = False

        handler = CacheHandler(mock_agent_core)

        assert handler.enable_dependency_cache is False

    def test_set_and_get_cache_data(self):
        """测试设置和获取缓存数据"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler
        from app.domain.vo.agent_cache.cache_data_vo import CacheDataVo

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        new_cache = CacheDataVo()
        handler.set_cache_data(new_cache)

        assert handler.get_cache_data() is new_cache

    def test_llm_config_methods(self):
        """测试LLM配置相关方法"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        llm_config = {"model": "gpt-4", "temperature": 0.7}
        handler.set_llm_config("llm123", llm_config)

        retrieved = handler.get_llm_config("llm123")
        assert retrieved == llm_config

        # Test non-existent LLM
        assert handler.get_llm_config("nonexistent") is None

    def test_agent_config_methods(self):
        """测试Agent配置相关方法"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        agent_config = {"agent_id": "agent123", "name": "Test Agent"}
        handler.set_agent_config(agent_config)

        retrieved = handler.get_agent_config()
        assert retrieved == agent_config

    def test_tools_info_methods(self):
        """测试工具信息相关方法"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        tool_info = {"name": "search_tool", "type": "api"}
        handler.set_tools_info_dict("tool123", tool_info)

        retrieved = handler.get_tools_info_dict("tool123")
        assert retrieved == tool_info

        # Test non-existent tool
        assert handler.get_tools_info_dict("nonexistent") is None

    def test_skill_agent_info_methods(self):
        """测试SkillAgent信息相关方法"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        skill_agent_info = {"skill_id": "skill123", "agent_id": "agent456"}
        handler.set_skill_agent_info_dict("skill_key_123", skill_agent_info)

        retrieved = handler.get_skill_agent_info_dict("skill_key_123")
        assert retrieved == skill_agent_info

        # Test non-existent skill agent
        assert handler.get_skill_agent_info_dict("nonexistent") is None

    def test_multiple_llm_configs(self):
        """测试多个LLM配置"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler = CacheHandler(mock_agent_core)

        llm_config1 = {"model": "gpt-4"}
        llm_config2 = {"model": "claude-3"}

        handler.set_llm_config("llm1", llm_config1)
        handler.set_llm_config("llm2", llm_config2)

        assert handler.get_llm_config("llm1") == llm_config1
        assert handler.get_llm_config("llm2") == llm_config2

    def test_cache_data_isolation(self):
        """测试缓存数据隔离"""
        from app.logic.agent_core_logic_v2.cache_handler import CacheHandler

        mock_agent_core = MagicMock()
        mock_agent_core.run_options_vo.enable_dependency_cache = True

        handler1 = CacheHandler(mock_agent_core)
        handler2 = CacheHandler(mock_agent_core)

        handler1.set_llm_config("llm1", {"model": "gpt-4"})

        # Different handlers should have different cache data
        assert handler2.get_llm_config("llm1") is None
