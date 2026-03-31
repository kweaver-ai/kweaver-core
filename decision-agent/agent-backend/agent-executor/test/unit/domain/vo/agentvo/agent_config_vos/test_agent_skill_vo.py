"""单元测试 - domain/vo/agentvo/agent_config_vos/agent_skill_vo 模块"""


class TestDataSourceTypeEnum:
    """测试 DataSourceTypeEnum 枚举"""

    def test_inherit_main_value(self):
        """测试INHERIT_MAIN值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceTypeEnum,
        )

        assert DataSourceTypeEnum.INHERIT_MAIN.value == "inherit_main"

    def test_self_configured_value(self):
        """测试SELF_CONFIGURED值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceTypeEnum,
        )

        assert DataSourceTypeEnum.SELF_CONFIGURED.value == "self_configured"

    def test_enum_is_string_enum(self):
        """测试枚举是字符串枚举"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceTypeEnum,
        )

        assert isinstance(DataSourceTypeEnum.INHERIT_MAIN, str)
        assert DataSourceTypeEnum.INHERIT_MAIN == "inherit_main"


class TestSpecificInheritEnum:
    """测试 SpecificInheritEnum 枚举"""

    def test_docs_only_value(self):
        """测试DOCS_ONLY值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            SpecificInheritEnum,
        )

        assert SpecificInheritEnum.DOCS_ONLY.value == "docs_only"

    def test_graph_only_value(self):
        """测试GRAPH_ONLY值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            SpecificInheritEnum,
        )

        assert SpecificInheritEnum.GRAPH_ONLY.value == "graph_only"

    def test_all_value(self):
        """测试ALL值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            SpecificInheritEnum,
        )

        assert SpecificInheritEnum.ALL.value == "all"


class TestLlmConfigTypeEnum:
    """测试 LlmConfigTypeEnum 枚举"""

    def test_inherit_main_value(self):
        """测试INHERIT_MAIN值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            LlmConfigTypeEnum,
        )

        assert LlmConfigTypeEnum.INHERIT_MAIN.value == "inherit_main"

    def test_self_configured_value(self):
        """测试SELF_CONFIGURED值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            LlmConfigTypeEnum,
        )

        assert LlmConfigTypeEnum.SELF_CONFIGURED.value == "self_configured"


class TestPmsCheckStatusEnum:
    """测试 PmsCheckStatusEnum 枚举"""

    def test_empty_value(self):
        """测试EMPTY值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            PmsCheckStatusEnum,
        )

        assert PmsCheckStatusEnum.EMPTY.value == ""

    def test_success_value(self):
        """测试SUCCESS值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            PmsCheckStatusEnum,
        )

        assert PmsCheckStatusEnum.SUCCESS.value == "success"

    def test_failed_value(self):
        """测试FAILED值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            PmsCheckStatusEnum,
        )

        assert PmsCheckStatusEnum.FAILED.value == "failed"


class TestDataSourceConfigVo:
    """测试 DataSourceConfigVo 模型"""

    def test_with_inherit_main(self):
        """测试使用INHERIT_MAIN"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceConfigVo,
            DataSourceTypeEnum,
        )

        vo = DataSourceConfigVo(type=DataSourceTypeEnum.INHERIT_MAIN)

        assert vo.type == DataSourceTypeEnum.INHERIT_MAIN
        assert vo.specific_inherit is None

    def test_with_self_configured(self):
        """测试使用SELF_CONFIGURED"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceConfigVo,
            DataSourceTypeEnum,
            SpecificInheritEnum,
        )

        vo = DataSourceConfigVo(
            type=DataSourceTypeEnum.SELF_CONFIGURED,
            specific_inherit=SpecificInheritEnum.ALL,
        )

        assert vo.type == DataSourceTypeEnum.SELF_CONFIGURED
        assert vo.specific_inherit == SpecificInheritEnum.ALL

    def test_empty_string_converts_to_none(self):
        """测试空字符串转换为None"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceConfigVo,
            DataSourceTypeEnum,
        )

        vo = DataSourceConfigVo(
            type=DataSourceTypeEnum.INHERIT_MAIN, specific_inherit=""
        )

        assert vo.specific_inherit is None

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            DataSourceConfigVo,
            DataSourceTypeEnum,
        )

        vo = DataSourceConfigVo(type=DataSourceTypeEnum.INHERIT_MAIN)
        data = vo.model_dump()

        assert data["type"] == "inherit_main"


class TestLlmConfigVo:
    """测试 LlmConfigVo 模型"""

    def test_with_inherit_main(self):
        """测试使用INHERIT_MAIN"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            LlmConfigVo,
            LlmConfigTypeEnum,
        )

        vo = LlmConfigVo(type=LlmConfigTypeEnum.INHERIT_MAIN)

        assert vo.type == LlmConfigTypeEnum.INHERIT_MAIN

    def test_with_self_configured(self):
        """测试使用SELF_CONFIGURED"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            LlmConfigVo,
            LlmConfigTypeEnum,
        )

        vo = LlmConfigVo(type=LlmConfigTypeEnum.SELF_CONFIGURED)

        assert vo.type == LlmConfigTypeEnum.SELF_CONFIGURED


class TestAgentInputVo:
    """测试 AgentInputVo 模型"""

    def test_all_required_fields(self):
        """测试所有必填字段"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import AgentInputVo

        vo = AgentInputVo(
            enable=True, input_name="query", input_type="string", map_type="auto"
        )

        assert vo.enable is True
        assert vo.input_name == "query"
        assert vo.input_type == "string"
        assert vo.map_type == "auto"
        assert vo.map_value is None
        assert vo.input_desc is None

    def test_with_all_fields(self):
        """测试所有字段都有值"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import AgentInputVo

        vo = AgentInputVo(
            enable=False,
            input_name="param",
            input_type="number",
            map_type="fixedValue",
            map_value=42,
            input_desc="A parameter",
        )

        assert vo.enable is False
        assert vo.map_value == 42
        assert vo.input_desc == "A parameter"

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import AgentInputVo

        vo = AgentInputVo(
            enable=True, input_name="test", input_type="string", map_type="auto"
        )

        data = vo.model_dump()

        assert data["enable"] is True
        assert data["input_name"] == "test"


class TestAgentSkillInnerDto:
    """测试 AgentSkillInnerDto 模型"""

    def test_allows_extra_fields(self):
        """测试允许额外字段"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            AgentSkillInnerDto,
        )

        dto = AgentSkillInnerDto(custom_field="custom_value", another_field=123)

        assert dto.custom_field == "custom_value"
        assert dto.another_field == 123

    def test_without_extra_fields(self):
        """测试不包含额外字段"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            AgentSkillInnerDto,
        )

        dto = AgentSkillInnerDto()

        # Should not raise any errors
        assert dto is not None

    def test_model_dump_includes_extra_fields(self):
        """测试序列化包含额外字段"""
        from app.domain.vo.agentvo.agent_config_vos.agent_skill_vo import (
            AgentSkillInnerDto,
        )

        dto = AgentSkillInnerDto(custom_field="value")
        data = dto.model_dump()

        assert data["custom_field"] == "value"
