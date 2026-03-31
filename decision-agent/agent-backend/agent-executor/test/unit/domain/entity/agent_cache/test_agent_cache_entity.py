"""单元测试 - domain/entity/agent_cache/agent_cache_entity 模块"""

from datetime import datetime


class TestAgentCacheEntity:
    """测试 AgentCacheEntity 类"""

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.entity.agent_cache import AgentCacheEntity
        from app.domain.vo.agent_cache import AgentCacheIdVO, CacheDataVo

        cache_id_vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )
        cache_data = CacheDataVo()
        now = datetime.now()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="agent456",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=now,
        )

        assert entity.cache_id_vo == cache_id_vo
        assert entity.agent_id == "agent456"
        assert entity.agent_version == "v0"
        assert entity.cache_data == cache_data
        assert entity.cache_data_last_set_timestamp == 1234567890
        assert entity.created_at == now

    def test_is_dataclass(self):
        """测试是dataclass"""
        from app.domain.entity.agent_cache import AgentCacheEntity

        # Should be a dataclass
        from dataclasses import is_dataclass

        assert is_dataclass(AgentCacheEntity)

    def test_default_values(self):
        """测试默认值"""
        from app.domain.entity.agent_cache import AgentCacheEntity
        from app.domain.vo.agent_cache import AgentCacheIdVO, CacheDataVo

        cache_id_vo = AgentCacheIdVO(
            account_id="acc123",
            account_type="user",
            agent_id="agent456",
            agent_version="v0",
            agent_config_version_flag="123456",
        )
        cache_data = CacheDataVo()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="agent456",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        # All fields should be set
        assert entity.cache_id_vo is not None
        assert entity.cache_data is not None
        assert entity.created_at is not None
