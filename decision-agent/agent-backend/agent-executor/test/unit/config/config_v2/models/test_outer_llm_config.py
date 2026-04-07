"""单元测试 - config/config_v2/models/outer_llm_config 模块"""


class TestOuterLLMConfig:
    """测试 OuterLLMConfig 数据类"""

    def test_default_values(self):
        """测试默认值"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        config = OuterLLMConfig()

        assert config.api == ""
        assert config.api_key == ""
        assert config.model == ""
        assert config.model_list == {}

    def test_with_custom_values(self):
        """测试自定义值"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        config = OuterLLMConfig(
            api="https://api.example.com",
            api_key="sk-test-key",
            model="gpt-4",
            model_list={"gpt-4": {"type": "chat"}},
        )

        assert config.api == "https://api.example.com"
        assert config.api_key == "sk-test-key"
        assert config.model == "gpt-4"
        assert config.model_list == {"gpt-4": {"type": "chat"}}

    def test_from_dict_with_all_fields(self):
        """测试从字典创建（所有字段）"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        data = {
            "api": "https://api.example.com",
            "api_key": "sk-test-key",
            "model": "gpt-4",
            "model_list": {"gpt-4": {"type": "chat"}, "gpt-3.5": {"type": "chat"}},
        }

        config = OuterLLMConfig.from_dict(data)

        assert config.api == "https://api.example.com"
        assert config.api_key == "sk-test-key"
        assert config.model == "gpt-4"
        assert len(config.model_list) == 2
        assert "gpt-4" in config.model_list

    def test_from_dict_with_defaults(self):
        """测试从字典创建（默认值）"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        config = OuterLLMConfig.from_dict({})

        assert config.api == ""
        assert config.api_key == ""
        assert config.model == ""
        assert config.model_list == {}

    def test_from_dict_with_partial_fields(self):
        """测试从字典创建（部分字段）"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        data = {"api": "https://api.example.com", "model": "gpt-4"}

        config = OuterLLMConfig.from_dict(data)

        assert config.api == "https://api.example.com"
        assert config.model == "gpt-4"
        assert config.api_key == ""
        assert config.model_list == {}

    def test_from_dict_model_list_not_dict(self):
        """测试model_list非字典类型"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        # If model_list is not a dict, it should be converted to empty dict
        data = {"model_list": ["gpt-4", "gpt-3.5"]}

        config = OuterLLMConfig.from_dict(data)

        assert config.model_list == {}

    def test_from_dict_model_list_none(self):
        """测试model_list为None"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        data = {"model_list": None}

        config = OuterLLMConfig.from_dict(data)

        assert config.model_list == {}

    def test_model_list_is_dict(self):
        """测试model_list是字典类型"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        config = OuterLLMConfig()

        assert isinstance(config.model_list, dict)

    def test_model_list_complex_values(self):
        """测试model_list包含复杂值"""
        from app.config.config_v2.models.outer_llm_config import OuterLLMConfig

        data = {
            "model_list": {
                "gpt-4": {
                    "type": "chat",
                    "max_tokens": 8192,
                    "supports_streaming": True,
                }
            }
        }

        config = OuterLLMConfig.from_dict(data)

        assert config.model_list["gpt-4"]["type"] == "chat"
        assert config.model_list["gpt-4"]["max_tokens"] == 8192
        assert config.model_list["gpt-4"]["supports_streaming"] is True
