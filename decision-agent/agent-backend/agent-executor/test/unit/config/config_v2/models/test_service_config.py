"""单元测试 - config/config_v2/models/service_config 模块"""


class TestServiceEndpoint:
    """测试 ServiceEndpoint 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.service_config import ServiceEndpoint

        endpoint = ServiceEndpoint()

        assert endpoint.host == ""
        assert endpoint.port == ""

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.service_config import ServiceEndpoint

        endpoint = ServiceEndpoint(host="localhost", port="8080")

        assert endpoint.host == "localhost"
        assert endpoint.port == "8080"


class TestServicesConfig:
    """测试 ServicesConfig 数据类"""

    def test_default_initialization(self):
        """测试默认初始化"""
        from app.config.config_v2.models.service_config import ServicesConfig

        config = ServicesConfig()

        # Check all default endpoints are set
        assert config.mf_model_api is not None
        assert config.mf_model_api.host == "mf-model-api"
        assert config.mf_model_api.port == "9898"

        assert config.agent_executor is not None
        assert config.agent_executor.host == "agent-executor"
        assert config.agent_executor.port == "30778"

        assert config.agent_factory is not None
        assert config.agent_factory.host == "agent-factory"
        assert config.agent_factory.port == "13020"

        assert config.agent_operator_integration is not None
        assert config.agent_operator_integration.host == "agent-operator-integration"
        assert config.agent_operator_integration.port == "9000"

        assert config.agent_memory is not None
        assert config.agent_memory.host == "agent-memory"
        assert config.agent_memory.port == "30790"

        assert config.kn_data_query is not None
        assert config.kn_data_query.host == "kn-data-query"
        assert config.kn_data_query.port == "6480"

        assert config.kn_knowledge_data is not None
        assert config.kn_knowledge_data.host == "kn-knowledge-data"
        assert config.kn_knowledge_data.port == "6475"

        assert config.data_connection is not None
        assert config.data_connection.host == "data-connection"
        assert config.data_connection.port == "8098"

        assert config.search_engine is not None
        assert config.search_engine.host == "kn-search-engine"
        assert config.search_engine.port == "6479"

        assert config.ecosearch is not None
        assert config.ecosearch.host == "ecosearch"
        assert config.ecosearch.port == "32126"

        assert config.ecoindex_public is not None
        assert config.ecoindex_public.host == "ecoindex-public"
        assert config.ecoindex_public.port == "32129"

        assert config.ecoindex_private is not None
        assert config.ecoindex_private.host == "ecoindex-private"
        assert config.ecoindex_private.port == "32130"

        assert config.docset_private is not None
        assert config.docset_private.host == "docset-private"
        assert config.docset_private.port == "32597"

        assert config.datahub is not None
        assert config.datahub.host == "datahubcentral-private"
        assert config.datahub.port == ""

    def test_from_dict_with_all_services(self):
        """测试从字典创建（所有服务）"""
        from app.config.config_v2.models.service_config import ServicesConfig

        data = {
            "mf_model_api": {"host": "model-api", "port": "9999"},
            "agent_executor": {"host": "executor", "port": "8000"},
            "agent_factory": {"host": "factory", "port": "8001"},
            "agent_operator_integration": {"host": "operator", "port": "8002"},
            "agent_memory": {"host": "memory", "port": "8003"},
            "kn_data_query": {"host": "query", "port": "8004"},
            "kn_knowledge_data": {"host": "knowledge", "port": "8005"},
            "data_connection": {"host": "connection", "port": "8006"},
            "search_engine": {"host": "search", "port": "8007"},
            "ecosearch": {"host": "ecosearch", "port": "8008"},
            "ecoindex_public": {"host": "index-public", "port": "8009"},
            "ecoindex_private": {"host": "index-private", "port": "8010"},
            "docset_private": {"host": "docset", "port": "8011"},
            "datahub": {"host": "datahub", "port": "8012"},
        }

        config = ServicesConfig.from_dict(data)

        assert config.mf_model_api.host == "model-api"
        assert config.mf_model_api.port == "9999"
        assert config.agent_executor.host == "executor"
        assert config.agent_executor.port == "8000"
        assert config.datahub.host == "datahub"
        assert config.datahub.port == "8012"

    def test_from_dict_with_partial_services(self):
        """测试从字典创建（部分服务）"""
        from app.config.config_v2.models.service_config import ServicesConfig

        data = {"mf_model_api": {"host": "custom-model-api", "port": "9999"}}

        config = ServicesConfig.from_dict(data)

        # Custom values for mf_model_api
        assert config.mf_model_api.host == "custom-model-api"
        assert config.mf_model_api.port == "9999"

        # Other services are created from empty dicts (no defaults set by __post_init__ for non-None values)
        assert config.agent_executor is not None
        assert config.agent_executor.host == ""
        assert config.agent_executor.port == ""

    def test_from_dict_with_empty_dict(self):
        """测试从空字典创建"""
        from app.config.config_v2.models.service_config import ServicesConfig

        config = ServicesConfig.from_dict({})

        # All services are created from empty dicts (no defaults set)
        assert config.mf_model_api is not None
        assert config.mf_model_api.host == ""
        assert config.agent_executor is not None
        assert config.agent_executor.host == ""

    def test_from_dict_port_to_string(self):
        """测试端口转换为字符串"""
        from app.config.config_v2.models.service_config import ServicesConfig

        data = {"mf_model_api": {"host": "model-api", "port": 9999}}

        config = ServicesConfig.from_dict(data)

        assert config.mf_model_api.port == "9999"
        assert isinstance(config.mf_model_api.port, str)

    def test_from_dict_empty_service_dict(self):
        """测试服务配置为空字典"""
        from app.config.config_v2.models.service_config import ServicesConfig

        data = {"mf_model_api": {}}

        config = ServicesConfig.from_dict(data)

        assert config.mf_model_api.host == ""
        assert config.mf_model_api.port == ""
