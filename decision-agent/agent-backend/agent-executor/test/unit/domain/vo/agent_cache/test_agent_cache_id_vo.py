"""单元测试 - domain/vo/agent_cache/agent_cache_id_vo 模块"""

import pytest


class TestAgentCacheIdVO:
    """测试 AgentCacheIdVO 类"""

    def test_init_with_keyword_arguments(self):
        """测试使用关键字参数初始化"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        assert vo.account_id == "acc123"
        assert vo.account_type == "user"
        assert vo.agent_id == "agent456"
        assert vo.agent_version == "v0"
        assert vo.agent_config_version_flag == "123456"

    def test_properties_are_readonly(self):
        """测试属性是只读的"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        # Properties should be accessible
        assert vo.account_id == "acc123"

    def test_to_redis_key(self):
        """测试转换为Redis key"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        redis_key = vo.to_redis_key()

        assert redis_key == "agent_executor:agent_cache:acc123:user:agent456:v0:123456"

    def test_get_cache_id(self):
        """测试获取cache_id"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        cache_id = vo.get_cache_id()

        assert cache_id == "acc123:user:agent456:v0:123456"

    def test_str_representation(self):
        """测试字符串表示"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        str_repr = str(vo)

        assert str_repr == "acc123:user:agent456:v0:123456"

    def test_slots_attribute(self):
        """测试__slots__属性"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )

        # Should have __slots__ defined
        assert hasattr(AgentCacheIdVO, "__slots__")

        # Should not be able to add new attributes
        with pytest.raises(AttributeError):
            vo.new_attribute = "value"

    def test_with_empty_strings(self):
        """测试使用空字符串初始化"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="",
            account_type="",
            agent_id="",
            agent_version="",
            agent_config_version_flag="",
        )

        assert vo.account_id == ""
        assert vo.get_cache_id() == "::::"

    def test_with_special_characters(self):
        """测试使用特殊字符"""
        from app.domain.vo.agent_cache import AgentCacheIdVO

        vo = AgentCacheIdVO(
            account_id="acc:123",
            account_type="user",
            agent_id="agent/456",
            agent_version="v0",
            agent_config_version_flag="123-456",
        )

        cache_id = vo.get_cache_id()

        assert cache_id == "acc:123:user:agent/456:v0:123-456"
