"""单元测试 - config/config_v2/models/database_config 模块"""


class TestRdsConfig:
    """测试 RdsConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.database_config import RdsConfig

        config = RdsConfig()

        assert config.host is None
        assert config.port == 3330
        assert config.dbname is None
        assert config.user is None
        assert config.password is None
        assert config.db_type == ""

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.database_config import RdsConfig

        config = RdsConfig(
            host="localhost",
            port=5432,
            dbname="testdb",
            user="testuser",
            password="testpass",
            db_type="postgresql",
        )

        assert config.host == "localhost"
        assert config.port == 5432
        assert config.dbname == "testdb"
        assert config.user == "testuser"
        assert config.password == "testpass"
        assert config.db_type == "postgresql"

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.database_config import RdsConfig

        data = {
            "host": "localhost",
            "port": "5432",
            "dbname": "testdb",
            "user": "testuser",
            "password": "testpass",
            "db_type": "postgresql",
        }

        config = RdsConfig.from_dict(data)

        assert config.host == "localhost"
        assert config.port == 5432
        assert config.dbname == "testdb"
        assert config.user == "testuser"
        assert config.password == "testpass"
        assert config.db_type == "postgresql"

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.database_config import RdsConfig

        config = RdsConfig.from_dict({})

        assert config.host is None
        assert config.port == 3330
        assert config.db_type == ""

    def test_from_dict_port_string_to_int(self):
        """测试端口字符串转换为整数"""
        from app.config.config_v2.models.database_config import RdsConfig

        config = RdsConfig.from_dict({"port": "5432"})

        assert config.port == 5432
        assert isinstance(config.port, int)


class TestRedisConfig:
    """测试 RedisConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.database_config import RedisConfig

        config = RedisConfig()

        assert config.cluster_mode == ""
        assert config.host == ""
        assert config.port == ""
        assert config.user == ""
        assert config.password == ""
        assert config.sentinel_master == ""
        assert config.sentinel_user == ""
        assert config.sentinel_password == ""
        assert config.read_host == ""
        assert config.read_port == ""
        assert config.read_user == ""
        assert config.read_password == ""
        assert config.write_host == ""
        assert config.write_port == ""
        assert config.write_user == ""
        assert config.write_password == ""

    def test_from_dict_standalone_mode(self):
        """测试从字典创建（standalone模式）"""
        from app.config.config_v2.models.database_config import RedisConfig

        data = {
            "cluster_mode": "standalone",
            "host": "localhost",
            "port": "6379",
            "user": "default",
            "password": "pass123",
        }

        config = RedisConfig.from_dict(data)

        assert config.cluster_mode == "standalone"
        assert config.host == "localhost"
        assert config.port == "6379"
        assert config.user == "default"
        assert config.password == "pass123"

    def test_from_dict_sentinel_mode(self):
        """测试从字典创建（sentinel模式）"""
        from app.config.config_v2.models.database_config import RedisConfig

        data = {
            "cluster_mode": "sentinel",
            "sentinel_master": "mymaster",
            "sentinel_user": "sentinel_user",
            "sentinel_password": "sentinel_pass",
            "host": "sentinel-host",
            "port": "26379",
        }

        config = RedisConfig.from_dict(data)

        assert config.cluster_mode == "sentinel"
        assert config.sentinel_master == "mymaster"
        assert config.sentinel_user == "sentinel_user"
        assert config.sentinel_password == "sentinel_pass"

    def test_from_dict_master_slave_mode(self):
        """测试从字典创建（master-slave模式）"""
        from app.config.config_v2.models.database_config import RedisConfig

        data = {
            "cluster_mode": "master-slave",
            "read_host": "read-host",
            "read_port": "6380",
            "read_user": "read_user",
            "read_password": "read_pass",
            "write_host": "write-host",
            "write_port": "6379",
            "write_user": "write_user",
            "write_password": "write_pass",
        }

        config = RedisConfig.from_dict(data)

        assert config.cluster_mode == "master-slave"
        assert config.read_host == "read-host"
        assert config.read_port == "6380"
        assert config.write_host == "write-host"
        assert config.write_port == "6379"

    def test_from_dict_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.database_config import RedisConfig

        config = RedisConfig.from_dict({})

        assert config.cluster_mode == ""
        assert config.host == ""
        assert config.port == ""


class TestGraphDBConfig:
    """测试 GraphDBConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.database_config import GraphDBConfig

        config = GraphDBConfig()

        assert config.host == ""
        assert config.port == ""
        assert config.type == "nebulaGraph"
        assert config.read_only_user == ""
        assert config.read_only_password == ""

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.database_config import GraphDBConfig

        config = GraphDBConfig(
            host="localhost",
            port="9669",
            type="nebulaGraph",
            read_only_user="user",
            read_only_password="pass",
        )

        assert config.host == "localhost"
        assert config.port == "9669"
        assert config.type == "nebulaGraph"
        assert config.read_only_user == "user"
        assert config.read_only_password == "pass"

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.database_config import GraphDBConfig

        data = {
            "host": "localhost",
            "port": "9669",
            "type": "nebulaGraph",
            "read_only_user": "user",
            "read_only_password": "pass",
        }

        config = GraphDBConfig.from_dict(data)

        assert config.host == "localhost"
        assert config.port == "9669"
        assert config.type == "nebulaGraph"
        assert config.read_only_user == "user"
        assert config.read_only_password == "pass"

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.database_config import GraphDBConfig

        config = GraphDBConfig.from_dict({})

        assert config.host == ""
        assert config.port == ""
        assert config.type == "nebulaGraph"
        assert config.read_only_user == ""

    def test_from_dict_port_to_string(self):
        """测试端口转换为字符串"""
        from app.config.config_v2.models.database_config import GraphDBConfig

        config = GraphDBConfig.from_dict({"port": 9669})

        assert config.port == "9669"
        assert isinstance(config.port, str)


class TestOpenSearchConfig:
    """测试 OpenSearchConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.database_config import OpenSearchConfig

        config = OpenSearchConfig()

        assert config.host == ""
        assert config.port == ""
        assert config.user == ""
        assert config.password == ""

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.database_config import OpenSearchConfig

        config = OpenSearchConfig(
            host="localhost", port="9200", user="admin", password="admin123"
        )

        assert config.host == "localhost"
        assert config.port == "9200"
        assert config.user == "admin"
        assert config.password == "admin123"

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.database_config import OpenSearchConfig

        data = {
            "host": "localhost",
            "port": "9200",
            "user": "admin",
            "password": "admin123",
        }

        config = OpenSearchConfig.from_dict(data)

        assert config.host == "localhost"
        assert config.port == "9200"
        assert config.user == "admin"
        assert config.password == "admin123"

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.database_config import OpenSearchConfig

        config = OpenSearchConfig.from_dict({})

        assert config.host == ""
        assert config.port == ""
        assert config.user == ""
        assert config.password == ""

    def test_from_dict_port_to_string(self):
        """测试端口转换为字符串"""
        from app.config.config_v2.models.database_config import OpenSearchConfig

        config = OpenSearchConfig.from_dict({"port": 9200})

        assert config.port == "9200"
        assert isinstance(config.port, str)
