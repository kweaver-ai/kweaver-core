"""单元测试 - config/config_v2/models/memory_config 模块"""


class TestMemoryConfig:
    """测试 MemoryConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig()

        assert config.limit == 50
        assert config.threshold == 0.5
        assert config.rerank_threshold == 0.1

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig(limit=100, threshold=0.7, rerank_threshold=0.2)

        assert config.limit == 100
        assert config.threshold == 0.7
        assert config.rerank_threshold == 0.2

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        data = {"limit": 100, "threshold": 0.7, "rerank_threshold": 0.2}

        config = MemoryConfig.from_dict(data)

        assert config.limit == 100
        assert config.threshold == 0.7
        assert config.rerank_threshold == 0.2

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig.from_dict({})

        assert config.limit == 50
        assert config.threshold == 0.5
        assert config.rerank_threshold == 0.1

    def test_from_dict_limit_string_to_int(self):
        """测试limit字符串转换为整数"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig.from_dict({"limit": "100"})

        assert config.limit == 100
        assert isinstance(config.limit, int)

    def test_from_dict_threshold_string_to_float(self):
        """测试threshold字符串转换为浮点数"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig.from_dict({"threshold": "0.7"})

        assert config.threshold == 0.7
        assert isinstance(config.threshold, float)

    def test_from_dict_rerank_threshold_string_to_float(self):
        """测试rerank_threshold字符串转换为浮点数"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig.from_dict({"rerank_threshold": "0.2"})

        assert config.rerank_threshold == 0.2
        assert isinstance(config.rerank_threshold, float)

    def test_threshold_boundary_values(self):
        """测试threshold边界值"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig(threshold=0.0)
        assert config.threshold == 0.0

        config = MemoryConfig(threshold=1.0)
        assert config.threshold == 1.0

    def test_rerank_threshold_boundary_values(self):
        """测试rerank_threshold边界值"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig(rerank_threshold=0.0)
        assert config.rerank_threshold == 0.0

        config = MemoryConfig(rerank_threshold=1.0)
        assert config.rerank_threshold == 1.0

    def test_limit_zero_value(self):
        """测试limit为0"""
        from app.config.config_v2.models.memory_config import MemoryConfig

        config = MemoryConfig(limit=0)
        assert config.limit == 0
