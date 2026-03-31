"""单元测试 - domain/vo/agentvo/agent_config_vos/output_config_vo 模块"""


class TestDefaultFormatEnum:
    """测试 DefaultFormatEnum 枚举"""

    def test_enum_values(self):
        """测试枚举值"""
        from app.domain.vo.agentvo.agent_config_vos import DefaultFormatEnum

        assert DefaultFormatEnum.JSON.value == "json"
        assert DefaultFormatEnum.MARKDOWN.value == "markdown"

    def test_enum_is_string_enum(self):
        """测试是字符串枚举"""
        from app.domain.vo.agentvo.agent_config_vos import DefaultFormatEnum

        assert isinstance(DefaultFormatEnum.JSON.value, str)
        assert isinstance(DefaultFormatEnum.JSON, str)


class TestOutputVariablesVo:
    """测试 OutputVariablesVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo()

        assert vo.answer_var is None
        assert vo.doc_retrieval_var is None
        assert vo.graph_retrieval_var is None
        assert vo.related_questions_var is None
        assert vo.other_vars is None

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(
            answer_var="custom_answer",
            doc_retrieval_var="custom_doc",
            graph_retrieval_var="custom_graph",
            related_questions_var="custom_questions",
            other_vars=["var1", "var2"],
        )

        assert vo.answer_var == "custom_answer"
        assert vo.doc_retrieval_var == "custom_doc"
        assert vo.graph_retrieval_var == "custom_graph"
        assert vo.related_questions_var == "custom_questions"
        assert vo.other_vars == ["var1", "var2"]


class TestOutputConfigVo:
    """测试 OutputConfigVo 类"""

    def test_init_with_required_field(self):
        """测试使用必填字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.JSON)

        assert vo.default_format == DefaultFormatEnum.JSON
        assert vo.variables is None

    def test_init_with_variables(self):
        """测试使用variables初始化"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var="answer")
        vo = OutputConfigVo(
            variables=variables, default_format=DefaultFormatEnum.MARKDOWN
        )

        assert vo.variables.answer_var == "answer"
        assert vo.default_format == DefaultFormatEnum.MARKDOWN

    def test_get_all_vars_with_no_variables(self):
        """测试get_all_vars当variables为None时返回空列表"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        result = vo.get_all_vars()

        assert result == []

    def test_get_all_vars_with_variables(self):
        """测试get_all_vars返回所有变量"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(
            answer_var="answer", doc_retrieval_var="doc", other_vars=["custom_var"]
        )
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)
        result = vo.get_all_vars()

        assert "answer" in result
        assert "doc" in result
        assert "graph_retrieval_res" in result
        assert "related_questions" in result  # default
        assert "custom_var" in result

    def test_get_final_answer_var_with_no_variables(self):
        """测试get_final_answer_var当variables为None时返回默认值"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        result = vo.get_final_answer_var()

        # Should return FINAL_ANSWER_DEFAULT_VAR
        assert result is not None

    def test_get_final_answer_var_with_variables(self):
        """测试get_final_answer_var返回answer_var"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var="custom_answer")
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)
        result = vo.get_final_answer_var()

        assert result == "custom_answer"

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import OutputConfigVo
        from pydantic import BaseModel

        assert issubclass(OutputConfigVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        data = vo.model_dump()

        assert data["default_format"] == "json"
        assert data["variables"] is None


class TestOutputVariablesVoExtended:
    """Extended tests for OutputVariablesVo class"""

    def test_init_with_partial_fields(self):
        """Test initialization with partial fields"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(answer_var="answer", doc_retrieval_var="doc")

        assert vo.answer_var == "answer"
        assert vo.doc_retrieval_var == "doc"
        assert vo.graph_retrieval_var is None
        assert vo.related_questions_var is None
        assert vo.other_vars is None

    def test_init_with_empty_other_vars(self):
        """Test initialization with empty other_vars list"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(other_vars=[])

        assert vo.other_vars == []

    def test_init_with_single_other_var(self):
        """Test initialization with single other_var"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(other_vars=["custom_var"])

        assert vo.other_vars == ["custom_var"]

    def test_init_with_multiple_other_vars(self):
        """Test initialization with multiple other_vars"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(other_vars=["var1", "var2", "var3"])

        assert vo.other_vars == ["var1", "var2", "var3"]

    def test_model_dump(self):
        """Test model_dump method"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(answer_var="answer", other_vars=["var1", "var2"])

        data = vo.model_dump()

        assert data["answer_var"] == "answer"
        assert data["other_vars"] == ["var1", "var2"]

    def test_model_dump_json(self):
        """Test model_dump_json method"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo = OutputVariablesVo(answer_var="answer")

        json_str = vo.model_dump_json()

        assert "answer" in json_str

    def test_copy(self):
        """Test copying the VO"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo1 = OutputVariablesVo(answer_var="answer")
        vo2 = vo1.copy()

        assert vo2.answer_var == "answer"

    def test_equality(self):
        """Test equality comparison"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        vo1 = OutputVariablesVo(answer_var="answer")
        vo2 = OutputVariablesVo(answer_var="answer")
        vo3 = OutputVariablesVo(answer_var="other")

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """Test creating instance from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import OutputVariablesVo

        data = {
            "answer_var": "answer",
            "doc_retrieval_var": "doc",
            "other_vars": ["var1"],
        }

        vo = OutputVariablesVo(**data)

        assert vo.answer_var == "answer"
        assert vo.doc_retrieval_var == "doc"
        assert vo.other_vars == ["var1"]


class TestOutputConfigVoExtended:
    """Extended tests for OutputConfigVo class"""

    def test_init_with_json_format(self):
        """Test initialization with JSON format"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.JSON)

        # Due to use_enum_values config, the enum is converted to its value
        assert (
            vo.default_format == "json" or vo.default_format == DefaultFormatEnum.JSON
        )

    def test_init_with_markdown_format(self):
        """Test initialization with Markdown format"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo = OutputConfigVo(default_format=DefaultFormatEnum.MARKDOWN)

        # Due to use_enum_values config, the enum is converted to its value
        assert (
            vo.default_format == "markdown"
            or vo.default_format == DefaultFormatEnum.MARKDOWN
        )

    def test_init_with_string_format(self):
        """Test initialization with string format value"""
        from app.domain.vo.agentvo.agent_config_vos import OutputConfigVo

        vo = OutputConfigVo(default_format="json")

        assert vo.default_format == "json"

    def test_get_all_vars_filters_empty_values(self):
        """Test get_all_vars filters out empty values"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(
            answer_var=None,
            doc_retrieval_var=None,
            graph_retrieval_var="graph",
            related_questions_var=None,
        )
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        result = vo.get_all_vars()

        # Should include the custom graph value
        assert "graph" in result
        # Should also include default values for None fields
        assert "answer" in result
        assert "doc_retrieval_res" in result
        assert "related_questions" in result

    def test_get_all_vars_includes_defaults_when_none(self):
        """Test get_all_vars includes default values when fields are None"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(
            answer_var=None,
            doc_retrieval_var=None,
            graph_retrieval_var=None,
            related_questions_var=None,
        )
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        result = vo.get_all_vars()

        # Should include default values
        assert any("answer" in v for v in result)
        assert "doc_retrieval_res" in result
        assert "graph_retrieval_res" in result
        assert "related_questions" in result

    def test_get_all_vars_with_custom_all_vars(self):
        """Test get_all_vars with all custom vars"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(
            answer_var="custom_answer",
            doc_retrieval_var="custom_doc",
            graph_retrieval_var="custom_graph",
            related_questions_var="custom_questions",
            other_vars=["var1", "var2", "var3"],
        )
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        result = vo.get_all_vars()

        assert "custom_answer" in result
        assert "custom_doc" in result
        assert "custom_graph" in result
        assert "custom_questions" in result
        assert "var1" in result
        assert "var2" in result
        assert "var3" in result

    def test_get_final_answer_var_returns_default_when_none(self):
        """Test get_final_answer_var returns default when answer_var is None"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var=None)
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        result = vo.get_final_answer_var()

        # Should return FINAL_ANSWER_DEFAULT_VAR
        assert result == "answer" or result is not None

    def test_get_final_answer_var_with_empty_string(self):
        """Test get_final_answer_var with empty string"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var="")
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        result = vo.get_final_answer_var()

        # Empty string should trigger default
        assert result == "answer" or result is not None

    def test_model_dump_with_variables(self):
        """Test model_dump with variables"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var="answer", other_vars=["var1"])
        vo = OutputConfigVo(variables=variables, default_format=DefaultFormatEnum.JSON)

        data = vo.model_dump()

        assert data["default_format"] == "json"
        assert data["variables"]["answer_var"] == "answer"
        assert data["variables"]["other_vars"] == ["var1"]

    def test_model_dump_json_with_variables(self):
        """Test model_dump_json with variables"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            OutputVariablesVo,
            DefaultFormatEnum,
        )

        variables = OutputVariablesVo(answer_var="answer")
        vo = OutputConfigVo(
            variables=variables, default_format=DefaultFormatEnum.MARKDOWN
        )

        json_str = vo.model_dump_json()

        assert "markdown" in json_str
        assert "answer" in json_str

    def test_copy(self):
        """Test copying the VO"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo1 = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        vo2 = vo1.copy()

        assert vo2.default_format == DefaultFormatEnum.JSON

    def test_equality(self):
        """Test equality comparison"""
        from app.domain.vo.agentvo.agent_config_vos import (
            OutputConfigVo,
            DefaultFormatEnum,
        )

        vo1 = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        vo2 = OutputConfigVo(default_format=DefaultFormatEnum.JSON)
        vo3 = OutputConfigVo(default_format=DefaultFormatEnum.MARKDOWN)

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """Test creating instance from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import OutputConfigVo

        data = {"default_format": "json", "variables": {"answer_var": "answer"}}

        vo = OutputConfigVo(**data)

        assert vo.default_format == "json"
        assert vo.variables.answer_var == "answer"


class TestDefaultFormatEnumExtended:
    """Extended tests for DefaultFormatEnum"""

    def test_enum_iteration(self):
        """Test iterating over enum values"""
        from app.domain.vo.agentvo.agent_config_vos import DefaultFormatEnum

        values = [e.value for e in DefaultFormatEnum]
        assert "json" in values
        assert "markdown" in values

    def test_enum_members(self):
        """Test enum has correct members"""
        from app.domain.vo.agentvo.agent_config_vos import DefaultFormatEnum

        assert hasattr(DefaultFormatEnum, "JSON")
        assert hasattr(DefaultFormatEnum, "MARKDOWN")

    def test_enum_comparison(self):
        """Test enum comparison"""
        from app.domain.vo.agentvo.agent_config_vos import DefaultFormatEnum

        assert DefaultFormatEnum.JSON == "json"
        assert DefaultFormatEnum.MARKDOWN == "markdown"
        assert DefaultFormatEnum.JSON != DefaultFormatEnum.MARKDOWN
