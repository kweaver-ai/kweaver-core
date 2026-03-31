"""单元测试 - logic/agent_core_logic_v2/output_variables 模块"""

import pytest
from unittest.mock import Mock


class TestGetOutputVariables:
    """测试 get_output_variables 函数"""

    @pytest.mark.asyncio
    async def test_get_output_variables_basic(self):
        """测试基本输出变量"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables
        from app.infra.common.infra_constant.const import FINAL_ANSWER_DEFAULT_VAR

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=["custom_var"])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = False

        result = get_output_variables(ac)

        # Should include custom_var from config
        assert "custom_var" in result
        # Should include default variables
        assert FINAL_ANSWER_DEFAULT_VAR in result
        assert "interventions" in result
        assert "tool" in result

    @pytest.mark.asyncio
    async def test_get_output_variables_with_progress(self):
        """测试带进度的输出变量"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=[])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = True

        result = get_output_variables(ac)

        # Should include _progress when is_need_progress is True
        assert "_progress" in result

    @pytest.mark.asyncio
    async def test_get_output_variables_without_progress(self):
        """测试不带进度的输出变量"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=[])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = False

        result = get_output_variables(ac)

        # Should NOT include _progress when is_need_progress is False
        assert "_progress" not in result

    @pytest.mark.asyncio
    async def test_get_output_variables_merges_with_defaults(self):
        """测试与默认变量合并"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=["custom1", "custom2"])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = False

        result = get_output_variables(ac)

        # Should include both custom and default variables
        assert "custom1" in result
        assert "custom2" in result
        assert "interventions" in result
        # No duplicates
        assert len([x for x in result if x == "custom1"]) == 1

    @pytest.mark.asyncio
    async def test_get_output_variables_all_defaults(self):
        """测试所有默认变量"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables
        from app.infra.common.infra_constant.const import FINAL_ANSWER_DEFAULT_VAR

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=[])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = False

        result = get_output_variables(ac)

        # Should include all default variables
        assert FINAL_ANSWER_DEFAULT_VAR in result
        assert "interventions" in result
        assert "intervention_judge_block_vars" in result
        assert "intervention_tool_block_vars" in result
        assert "tool" in result
        assert "_progress" not in result

    @pytest.mark.asyncio
    async def test_get_output_variables_returns_list(self):
        """测试返回列表"""
        from app.logic.agent_core_logic_v2.output_variables import get_output_variables

        # Mock AgentCoreV2
        ac = Mock()
        ac.agent_config = Mock()
        ac.agent_config.output = Mock()
        ac.agent_config.output.get_all_vars = Mock(return_value=["var1"])

        ac.run_options_vo = Mock()
        ac.run_options_vo.is_need_progress = False

        result = get_output_variables(ac)

        # Should return a list
        assert isinstance(result, list)
