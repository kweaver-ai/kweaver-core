"""单元测试 - domain/vo/agent_cache 模块"""

from unittest import TestCase

from app.domain.vo.agent_cache.cache_data_vo import CacheDataVo
from app.domain.vo.agent_cache.agent_cache_id_vo import AgentCacheIdVO
from app.domain.entity.agent_cache.agent_cache_entity import AgentCacheEntity
from datetime import datetime


class TestAgentCacheIdVO(TestCase):
    """测试 AgentCacheIdVO 类"""

    def test_init(self):
        """测试初始化"""
        vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="agent_456",
            agent_version="v0",
            agent_config_version_flag="timestamp_789",
        )
        self.assertEqual(vo.account_id, "account_123")
        self.assertEqual(vo.account_type, "app")
        self.assertEqual(vo.agent_id, "agent_456")
        self.assertEqual(vo.agent_version, "v0")
        self.assertEqual(vo.agent_config_version_flag, "timestamp_789")

    def test_to_redis_key(self):
        """测试转换为Redis key"""
        vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="agent_456",
            agent_version="v0",
            agent_config_version_flag="timestamp_789",
        )
        key = vo.to_redis_key()
        self.assertEqual(
            key, "agent_executor:agent_cache:account_123:app:agent_456:v0:timestamp_789"
        )

    def test_get_cache_id(self):
        """测试获取cache_id"""
        vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="agent_456",
            agent_version="v0",
            agent_config_version_flag="timestamp_789",
        )
        cache_id = vo.get_cache_id()
        self.assertEqual(cache_id, "account_123:app:agent_456:v0:timestamp_789")

    def test_str(self):
        """测试字符串表示"""
        vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="agent_456",
            agent_version="v0",
            agent_config_version_flag="timestamp_789",
        )
        str_repr = str(vo)
        self.assertEqual(str_repr, "account_123:app:agent_456:v0:timestamp_789")

    def test_version_v0(self):
        """测试v0版本"""
        vo = AgentCacheIdVO(
            account_id="acc",
            account_type="app",
            agent_id="agent",
            agent_version="v0",
            agent_config_version_flag="flag",
        )
        self.assertEqual(vo.agent_version, "v0")

    def test_version_latest(self):
        """测试latest版本"""
        vo = AgentCacheIdVO(
            account_id="acc",
            account_type="user",
            agent_id="agent",
            agent_version="latest",
            agent_config_version_flag="flag",
        )
        self.assertEqual(vo.agent_version, "latest")


class TestCacheDataVo(TestCase):
    """测试 CacheDataVo 类"""

    def test_init_default(self):
        """测试默认初始化"""
        cache_data = CacheDataVo()
        self.assertEqual(cache_data.agent_config, {})
        self.assertEqual(cache_data.tools_info_dict, {})
        self.assertEqual(cache_data.skill_agent_info_dict, {})
        self.assertEqual(cache_data.llm_config_dict, {})

    def test_init_with_values(self):
        """测试带值初始化"""
        config = {"key": "value"}
        tools = {"tool1": "info1"}
        skill_agents = {"skill1": "agent1"}
        llm_config = {"model": "gpt-4"}

        cache_data = CacheDataVo()
        cache_data.agent_config = config
        cache_data.tools_info_dict = tools
        cache_data.skill_agent_info_dict = skill_agents
        cache_data.llm_config_dict = llm_config

        self.assertEqual(cache_data.agent_config, config)
        self.assertEqual(cache_data.tools_info_dict, tools)
        self.assertEqual(cache_data.skill_agent_info_dict, skill_agents)
        self.assertEqual(cache_data.llm_config_dict, llm_config)


class TestAgentCacheEntity(TestCase):
    """测试 AgentCacheEntity 类"""

    def test_init(self):
        """测试初始化"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="test_agent",
            agent_version="v0",
            agent_config_version_flag="timestamp_123",
        )
        cache_data = CacheDataVo()
        cache_data.agent_config = {"config": "value"}

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="test_agent",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        self.assertEqual(entity.agent_id, "test_agent")
        self.assertEqual(entity.agent_version, "v0")
        self.assertEqual(entity.cache_data_last_set_timestamp, 1234567890)
        self.assertEqual(entity.cache_id_vo.agent_id, "test_agent")
        self.assertEqual(entity.cache_id_vo.account_id, "account_123")

    def test_init_all_fields(self):
        """测试所有字段初始化"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="agent_123",
            agent_version="latest",
            agent_config_version_flag="timestamp_456",
        )
        cache_data = CacheDataVo()
        cache_data.agent_config = {"key": "value"}
        cache_data.tools_info_dict = {"tool": "info"}
        cache_data.skill_agent_info_dict = {"skill": "agent"}
        cache_data.llm_config_dict = {"model": "gpt-4"}

        created_at = datetime(2024, 1, 1, 12, 0, 0)

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="agent_123",
            agent_version="latest",
            cache_data=cache_data,
            cache_data_last_set_timestamp=9876543210,
            created_at=created_at,
        )

        self.assertEqual(entity.agent_id, "agent_123")
        self.assertEqual(entity.agent_version, "latest")
        self.assertEqual(entity.cache_data_last_set_timestamp, 9876543210)
        self.assertEqual(entity.created_at, created_at)
        self.assertEqual(len(entity.cache_data.tools_info_dict), 1)

    def test_empty_cache_data(self):
        """测试空缓存数据"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="test_agent",
            agent_version="v0",
            agent_config_version_flag="timestamp_123",
        )
        cache_data = CacheDataVo()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="test_agent",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        self.assertEqual(entity.cache_data.agent_config, {})
        self.assertEqual(entity.cache_data.tools_info_dict, {})
        self.assertEqual(entity.cache_data.skill_agent_info_dict, {})
        self.assertEqual(entity.cache_data.llm_config_dict, {})

    def test_version_v0(self):
        """测试v0版本"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="test_agent",
            agent_version="v0",
            agent_config_version_flag="timestamp_123",
        )
        cache_data = CacheDataVo()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="test_agent",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        self.assertEqual(entity.agent_version, "v0")

    def test_version_latest(self):
        """测试latest版本"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="test_agent",
            agent_version="latest",
            agent_config_version_flag="timestamp_123",
        )
        cache_data = CacheDataVo()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="test_agent",
            agent_version="latest",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        self.assertEqual(entity.agent_version, "latest")

    def test_timestamp_type(self):
        """测试时间戳类型"""
        cache_id_vo = AgentCacheIdVO(
            account_id="account_123",
            account_type="app",
            agent_id="test_agent",
            agent_version="v0",
            agent_config_version_flag="timestamp_123",
        )
        cache_data = CacheDataVo()

        entity = AgentCacheEntity(
            cache_id_vo=cache_id_vo,
            agent_id="test_agent",
            agent_version="v0",
            cache_data=cache_data,
            cache_data_last_set_timestamp=1234567890,
            created_at=datetime.now(),
        )

        self.assertIsInstance(entity.cache_data_last_set_timestamp, int)
        self.assertIsInstance(entity.created_at, datetime)
