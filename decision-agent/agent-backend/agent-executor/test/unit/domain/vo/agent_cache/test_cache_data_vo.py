"""单元测试 - domain/vo/agent_cache/cache_data_vo 模块"""


class TestCacheDataVo:
    """测试 CacheDataVo 类"""

    def test_init_creates_empty_dicts(self):
        """测试初始化创建空字典"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()

        assert vo.agent_config == {}
        assert vo.tools_info_dict == {}
        assert vo.skill_agent_info_dict == {}
        assert vo.llm_config_dict == {}

    def test_agent_config_can_be_modified(self):
        """测试agent_config可以被修改"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()
        vo.agent_config["key"] = "value"

        assert vo.agent_config == {"key": "value"}

    def test_tools_info_dict_can_be_modified(self):
        """测试tools_info_dict可以被修改"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()
        vo.tools_info_dict["tool1"] = {"name": "test"}

        assert vo.tools_info_dict == {"tool1": {"name": "test"}}

    def test_skill_agent_info_dict_can_be_modified(self):
        """测试skill_agent_info_dict可以被修改"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()
        vo.skill_agent_info_dict["skill1"] = {"config": {}}

        assert vo.skill_agent_info_dict == {"skill1": {"config": {}}}

    def test_llm_config_dict_can_be_modified(self):
        """测试llm_config_dict可以被修改"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()
        vo.llm_config_dict["model"] = "gpt-4"

        assert vo.llm_config_dict == {"model": "gpt-4"}

    def test_all_dicts_are_independent(self):
        """测试所有字典是独立的"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()
        vo.agent_config["key"] = "value"

        # Other dicts should remain empty
        assert vo.tools_info_dict == {}
        assert vo.skill_agent_info_dict == {}
        assert vo.llm_config_dict == {}

    def test_multiple_vos_are_independent(self):
        """测试多个VO实例是独立的"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo1 = CacheDataVo()
        vo2 = CacheDataVo()

        vo1.agent_config["key"] = "value"

        assert vo1.agent_config == {"key": "value"}
        assert vo2.agent_config == {}

    def test_dict_attributes_are_dicts(self):
        """测试字典属性是dict类型"""
        from app.domain.vo.agent_cache import CacheDataVo

        vo = CacheDataVo()

        assert isinstance(vo.agent_config, dict)
        assert isinstance(vo.tools_info_dict, dict)
        assert isinstance(vo.skill_agent_info_dict, dict)
        assert isinstance(vo.llm_config_dict, dict)
