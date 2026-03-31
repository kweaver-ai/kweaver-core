"""单元测试 - domain/vo/agentvo/agent_config_vos/tool_skill_vo 模块"""


class TestResultProcessCategoryVo:
    """测试 ResultProcessCategoryVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessCategoryVo

        vo = ResultProcessCategoryVo()

        assert vo.id is None
        assert vo.name is None
        assert vo.description is None

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessCategoryVo

        vo = ResultProcessCategoryVo(
            id="category_123", name="Test Category", description="Test description"
        )

        assert vo.id == "category_123"
        assert vo.name == "Test Category"
        assert vo.description == "Test description"


class TestResultProcessStrategyDetailVo:
    """测试 ResultProcessStrategyDetailVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyDetailVo

        vo = ResultProcessStrategyDetailVo()

        assert vo.id is None
        assert vo.name is None
        assert vo.description is None

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyDetailVo

        vo = ResultProcessStrategyDetailVo(
            id="strategy_123", name="Test Strategy", description="Test description"
        )

        assert vo.id == "strategy_123"
        assert vo.name == "Test Strategy"
        assert vo.description == "Test description"


class TestResultProcessStrategyVo:
    """测试 ResultProcessStrategyVo 类"""

    def test_init_with_no_fields(self):
        """测试不使用任何字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyVo

        vo = ResultProcessStrategyVo()

        assert vo.category is None
        assert vo.strategy is None

    def test_init_with_category_and_strategy(self):
        """测试使用category和strategy初始化"""
        from app.domain.vo.agentvo.agent_config_vos import (
            ResultProcessStrategyVo,
            ResultProcessCategoryVo,
            ResultProcessStrategyDetailVo,
        )

        category = ResultProcessCategoryVo(id="cat_123", name="Category")
        strategy = ResultProcessStrategyDetailVo(id="strat_123", name="Strategy")

        vo = ResultProcessStrategyVo(category=category, strategy=strategy)

        assert vo.category.id == "cat_123"
        assert vo.strategy.id == "strat_123"


class TestToolSkillVo:
    """测试 ToolSkillVo 类"""

    def test_init_with_required_fields(self):
        """测试使用必填字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool_123", tool_box_id="toolbox_123")

        assert vo.tool_id == "tool_123"
        assert vo.tool_box_id == "toolbox_123"
        assert vo.tool_timeout == 300  # default value
        assert vo.tool_input == []
        assert vo.intervention is False
        assert vo.result_process_strategies == []

    def test_init_with_all_fields(self):
        """测试使用所有字段初始化"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(
            tool_id="tool_123",
            tool_box_id="toolbox_123",
            tool_timeout=600,
            intervention=True,
            intervention_confirmation_message="Please confirm",
        )

        assert vo.tool_timeout == 600
        assert vo.intervention is True
        assert vo.intervention_confirmation_message == "Please confirm"

    def test_tool_timeout_default(self):
        """测试tool_timeout默认值为300"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool_123", tool_box_id="toolbox_123")

        assert vo.tool_timeout == 300

    def test_intervention_default(self):
        """测试intervention默认值为False"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool_123", tool_box_id="toolbox_123")

        assert vo.intervention is False

    def test_is_pydantic_model(self):
        """测试是Pydantic模型"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo
        from pydantic import BaseModel

        assert issubclass(ToolSkillVo, BaseModel)

    def test_model_dump(self):
        """测试模型序列化"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool_123", tool_box_id="toolbox_123")
        data = vo.model_dump()

        assert data["tool_id"] == "tool_123"
        assert data["tool_box_id"] == "toolbox_123"
        assert data["tool_timeout"] == 300


class TestResultProcessCategoryVoExtended:
    """Extended tests for ResultProcessCategoryVo"""

    def test_model_dump(self):
        """Test model_dump method"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessCategoryVo

        vo = ResultProcessCategoryVo(id="cat1", name="Category")
        data = vo.model_dump()

        assert data["id"] == "cat1"
        assert data["name"] == "Category"

    def test_from_dict(self):
        """Test creating from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessCategoryVo

        data = {"id": "cat1", "name": "Test"}
        vo = ResultProcessCategoryVo(**data)

        assert vo.id == "cat1"
        assert vo.name == "Test"

    def test_with_empty_strings(self):
        """Test with empty string fields"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessCategoryVo

        vo = ResultProcessCategoryVo(id="", name="", description="")

        assert vo.id == ""
        assert vo.name == ""
        assert vo.description == ""


class TestResultProcessStrategyDetailVoExtended:
    """Extended tests for ResultProcessStrategyDetailVo"""

    def test_model_dump(self):
        """Test model_dump method"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyDetailVo

        vo = ResultProcessStrategyDetailVo(id="strat1", name="Strategy")
        data = vo.model_dump()

        assert data["id"] == "strat1"
        assert data["name"] == "Strategy"

    def test_from_dict(self):
        """Test creating from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyDetailVo

        data = {"id": "strat1", "name": "Test"}
        vo = ResultProcessStrategyDetailVo(**data)

        assert vo.id == "strat1"
        assert vo.name == "Test"


class TestResultProcessStrategyVoExtended:
    """Extended tests for ResultProcessStrategyVo"""

    def test_model_dump(self):
        """Test model_dump method"""
        from app.domain.vo.agentvo.agent_config_vos import (
            ResultProcessStrategyVo,
            ResultProcessCategoryVo,
            ResultProcessStrategyDetailVo,
        )

        category = ResultProcessCategoryVo(id="cat1")
        strategy = ResultProcessStrategyDetailVo(id="strat1")

        vo = ResultProcessStrategyVo(category=category, strategy=strategy)
        data = vo.model_dump()

        assert data["category"]["id"] == "cat1"
        assert data["strategy"]["id"] == "strat1"

    def test_from_dict(self):
        """Test creating from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import ResultProcessStrategyVo

        data = {
            "category": {"id": "cat1", "name": "Category"},
            "strategy": {"id": "strat1", "name": "Strategy"},
        }
        vo = ResultProcessStrategyVo(**data)

        assert vo.category.id == "cat1"
        assert vo.strategy.id == "strat1"


class TestToolSkillVoExtended:
    """Extended tests for ToolSkillVo"""

    def test_init_with_tool_input(self):
        """Test initialization with tool_input"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo, SkillInputVo

        tool_input = SkillInputVo(input_name="param1", input_type="string")

        vo = ToolSkillVo(
            tool_id="tool1", tool_box_id="toolbox1", tool_input=[tool_input]
        )

        assert len(vo.tool_input) == 1
        assert vo.tool_input[0].input_name == "param1"

    def test_init_with_multiple_tool_inputs(self):
        """Test initialization with multiple tool_inputs"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo, SkillInputVo

        tool_input1 = SkillInputVo(input_name="param1", input_type="string")
        tool_input2 = SkillInputVo(input_name="param2", input_type="int")
        tool_input3 = SkillInputVo(input_name="param3", input_type="bool")

        vo = ToolSkillVo(
            tool_id="tool1",
            tool_box_id="toolbox1",
            tool_input=[tool_input1, tool_input2, tool_input3],
        )

        assert len(vo.tool_input) == 3
        assert vo.tool_input[0].input_name == "param1"
        assert vo.tool_input[1].input_name == "param2"
        assert vo.tool_input[2].input_name == "param3"

    def test_init_with_result_process_strategies(self):
        """Test initialization with result_process_strategies"""
        from app.domain.vo.agentvo.agent_config_vos import (
            ToolSkillVo,
            ResultProcessStrategyVo,
            ResultProcessCategoryVo,
            ResultProcessStrategyDetailVo,
        )

        category = ResultProcessCategoryVo(id="cat1", name="Category")
        strategy = ResultProcessStrategyDetailVo(id="strat1", name="Strategy")
        result_strategy = ResultProcessStrategyVo(category=category, strategy=strategy)

        vo = ToolSkillVo(
            tool_id="tool1",
            tool_box_id="toolbox1",
            result_process_strategies=[result_strategy],
        )

        assert len(vo.result_process_strategies) == 1
        assert vo.result_process_strategies[0].category.id == "cat1"

    def test_init_with_multiple_result_process_strategies(self):
        """Test initialization with multiple result_process_strategies"""
        from app.domain.vo.agentvo.agent_config_vos import (
            ToolSkillVo,
            ResultProcessStrategyVo,
            ResultProcessCategoryVo,
            ResultProcessStrategyDetailVo,
        )

        category1 = ResultProcessCategoryVo(id="cat1", name="Category1")
        strategy1 = ResultProcessStrategyDetailVo(id="strat1", name="Strategy1")
        result_strategy1 = ResultProcessStrategyVo(
            category=category1, strategy=strategy1
        )

        category2 = ResultProcessCategoryVo(id="cat2", name="Category2")
        strategy2 = ResultProcessStrategyDetailVo(id="strat2", name="Strategy2")
        result_strategy2 = ResultProcessStrategyVo(
            category=category2, strategy=strategy2
        )

        vo = ToolSkillVo(
            tool_id="tool1",
            tool_box_id="toolbox1",
            result_process_strategies=[result_strategy1, result_strategy2],
        )

        assert len(vo.result_process_strategies) == 2

    def test_init_with_intervention_false(self):
        """Test initialization with intervention=False"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1", intervention=False)

        assert vo.intervention is False

    def test_init_with_intervention_true(self):
        """Test initialization with intervention=True"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1", intervention=True)

        assert vo.intervention is True

    def test_init_with_intervention_confirmation_message(self):
        """Test initialization with intervention_confirmation_message"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(
            tool_id="tool1",
            tool_box_id="toolbox1",
            intervention_confirmation_message="Please confirm this action",
        )

        assert vo.intervention_confirmation_message == "Please confirm this action"

    def test_tool_timeout_custom_value(self):
        """Test with custom tool_timeout value"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1", tool_timeout=120)

        assert vo.tool_timeout == 120

    def test_tool_timeout_zero(self):
        """Test with tool_timeout=0"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1", tool_timeout=0)

        assert vo.tool_timeout == 0

    def test_model_dump_with_all_fields(self):
        """Test model_dump with all fields"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(
            tool_id="tool1",
            tool_box_id="toolbox1",
            tool_timeout=600,
            intervention=True,
            intervention_confirmation_message="Confirm",
        )

        data = vo.model_dump()

        assert data["tool_id"] == "tool1"
        assert data["tool_box_id"] == "toolbox1"
        assert data["tool_timeout"] == 600
        assert data["intervention"] is True
        assert data["intervention_confirmation_message"] == "Confirm"

    def test_model_dump_json(self):
        """Test model_dump_json method"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1")

        json_str = vo.model_dump_json()

        assert "tool1" in json_str
        assert "toolbox1" in json_str

    def test_copy(self):
        """Test copying the VO"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo1 = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1")
        vo2 = vo1.copy()

        assert vo2.tool_id == "tool1"
        assert vo2.tool_box_id == "toolbox1"

    def test_equality(self):
        """Test equality comparison"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo1 = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1")
        vo2 = ToolSkillVo(tool_id="tool1", tool_box_id="toolbox1")
        vo3 = ToolSkillVo(tool_id="tool2", tool_box_id="toolbox2")

        assert vo1 == vo2
        assert vo1 != vo3

    def test_from_dict(self):
        """Test creating from dictionary"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        data = {"tool_id": "tool1", "tool_box_id": "toolbox1", "tool_timeout": 600}

        vo = ToolSkillVo(**data)

        assert vo.tool_id == "tool1"
        assert vo.tool_box_id == "toolbox1"
        assert vo.tool_timeout == 600

    def test_with_unicode_fields(self):
        """Test with unicode characters in fields"""
        from app.domain.vo.agentvo.agent_config_vos import ToolSkillVo

        vo = ToolSkillVo(
            tool_id="工具1",
            tool_box_id="工具箱1",
            intervention_confirmation_message="请确认此操作",
        )

        assert vo.tool_id == "工具1"
        assert vo.tool_box_id == "工具箱1"
        assert "请确认" in vo.intervention_confirmation_message
