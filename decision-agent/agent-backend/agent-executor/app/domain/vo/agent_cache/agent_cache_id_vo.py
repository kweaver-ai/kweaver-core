# -*- coding:utf-8 -*-
class AgentCacheIdVO:
    """Agent缓存ID值对象

    缓存key的构成：
    - account_id: 账户ID
    - account_type: 账户类型
    - agent_id
    - agent_version
    - agent_config_version_flag (即 config.metadata.config_last_set_timestamp)

    注意：实例化时必须使用 key=value 的方式传递参数
    """

    __slots__ = (
        "_account_id",
        "_account_type",
        "_agent_id",
        "_agent_version",
        "_agent_config_version_flag",
    )

    def __init__(
        self,
        *,
        account_id: str,
        account_type: str,
        agent_id: str,
        agent_version: str,
        agent_config_version_flag: str,
    ):
        """初始化Agent缓存ID值对象

        Args:
            account_id: 账户ID
            account_type: 账户类型
            agent_id: Agent ID
            agent_version: Agent版本
            agent_config_version_flag: Agent配置版本标识
        """
        self._account_id = account_id
        self._account_type = account_type
        self._agent_id = agent_id
        self._agent_version = agent_version
        self._agent_config_version_flag = agent_config_version_flag

    @property
    def account_id(self) -> str:
        return self._account_id

    @property
    def account_type(self) -> str:
        return self._account_type

    @property
    def agent_id(self) -> str:
        return self._agent_id

    @property
    def agent_version(self) -> str:
        return self._agent_version

    @property
    def agent_config_version_flag(self) -> str:
        return self._agent_config_version_flag

    def to_redis_key(self) -> str:
        """转换为Redis key

        Returns:
            Redis key字符串
        """
        return f"agent_executor:agent_cache:{self.get_cache_id()}"

    def get_cache_id(self) -> str:
        """获取cache_id（不含前缀）

        Returns:
            cache_id字符串，格式：{account_id}:{account_type}:{agent_id}:{agent_version}:{agent_config_version_flag}
        """
        return f"{self._account_id}:{self._account_type}:{self._agent_id}:{self._agent_version}:{self._agent_config_version_flag}"

    def __str__(self) -> str:
        """字符串表示

        Returns:
            cache_id字符串
        """
        return self.get_cache_id()
