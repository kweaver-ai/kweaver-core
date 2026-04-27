"""单元测试 - domain/vo/agentvo/agent_config_vos/config_metadata_vo 模块"""


class TestConfigMetadataVo:
    """测试 ConfigMetadataVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo()

        assert vo.config_tpl_version == ""
        assert vo.config_last_set_timestamp is None

    def test_init_with_config_tpl_version(self):
        """测试使用config_tpl_version初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo(config_tpl_version="v1.0")

        assert vo.config_tpl_version == "v1.0"
        assert vo.config_last_set_timestamp is None

    def test_init_with_timestamp(self):
        """测试使用timestamp初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo(config_last_set_timestamp=1234567890)

        assert vo.config_tpl_version == ""
        assert vo.config_last_set_timestamp == 1234567890

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo(
            config_tpl_version="v2.0", config_last_set_timestamp=9876543210
        )

        assert vo.config_tpl_version == "v2.0"
        assert vo.config_last_set_timestamp == 9876543210

    def test_config_last_set_timestamp_str_with_none(self):
        """测试config_last_set_timestamp_str为None时返回空字符串"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo()

        assert vo.config_last_set_timestamp_str == ""

    def test_config_last_set_timestamp_str_with_value(self):
        """测试config_last_set_timestamp_str返回字符串"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo(config_last_set_timestamp=1234567890)

        assert vo.config_last_set_timestamp_str == "1234567890"

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo
        from pydantic import BaseModel

        assert issubclass(ConfigMetadataVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        vo = ConfigMetadataVo(
            config_tpl_version="v1.0", config_last_set_timestamp=1234567890
        )

        data = vo.model_dump()

        assert data["config_tpl_version"] == "v1.0"
        assert data["config_last_set_timestamp"] == 1234567890

    def test_validate_config_last_set_timestamp_directly(self):
        """测试config_last_set_timestamp验证器直接调用"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        # Test the validator method directly
        # The validator should return 0 for non-int values
        result = ConfigMetadataVo.validate_config_last_set_timestamp(123.45)
        assert result == 0

        result = ConfigMetadataVo.validate_config_last_set_timestamp("string")
        assert result == 0

        result = ConfigMetadataVo.validate_config_last_set_timestamp(None)
        assert result is None

        result = ConfigMetadataVo.validate_config_last_set_timestamp(1234567890)
        assert result == 1234567890

    def test_validate_config_tpl_version_directly(self):
        """测试config_tpl_version验证器直接调用"""
        from app.domain.vo.agentvo.agent_config_vos import ConfigMetadataVo

        # Test the validator method directly
        result = ConfigMetadataVo.validate_config_tpl_version("v1.0")
        assert result == "v1.0"

        result = ConfigMetadataVo.validate_config_tpl_version("")
        assert result == ""

        result = ConfigMetadataVo.validate_config_tpl_version("any_value")
        assert result == "any_value"
