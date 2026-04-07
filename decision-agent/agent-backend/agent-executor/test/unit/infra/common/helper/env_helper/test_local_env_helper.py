# -*- coding:utf-8 -*-
"""Unit tests for local_env_helper module"""

import pytest
from unittest.mock import patch


class TestRunScenario:
    """Tests for RunScenario enum"""

    def test_run_scenario_enum_values(self):
        """Test RunScenario enum has correct values"""
        from app.infra.common.helper.env_helper import RunScenario

        assert RunScenario.AARON_LOCAL_DEV.value == "aaron_local_dev"

    def test_run_scenario_enum_members(self):
        """Test RunScenario enum members"""
        from app.infra.common.helper.env_helper import RunScenario

        assert hasattr(RunScenario, "AARON_LOCAL_DEV")
        assert len(RunScenario) >= 1


class TestLocalEnvHelper:
    """Tests for LocalEnvHelper class"""

    def test_init_initializes_env(self):
        """Test that initialization sets up environment"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "false", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._env_inited is True

    def test_init_with_is_local_dev_true(self):
        """Test initialization with IS_LOCAL_DEV=true"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._is_local_dev is True

    def test_init_with_is_local_dev_false(self):
        """Test initialization with IS_LOCAL_DEV=false"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "false", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._is_local_dev is False

    def test_init_with_is_local_dev_case_insensitive(self):
        """Test initialization with various case combinations"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        for value in ["TRUE", "True", "tRuE", "true"]:
            with patch.dict(
                "os.environ",
                {"IS_LOCAL_DEV": value, "LOCAL_DEVRUN_SCENARIO": ""},
                clear=True,
            ):
                helper = LocalEnvHelper()
                assert helper._is_local_dev is True

    def test_init_with_single_scenario(self):
        """Test initialization with single scenario"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert "aaron_local_dev" in helper._run_scenarios

    def test_init_with_multiple_scenarios(self):
        """Test initialization with multiple scenarios"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {
                "IS_LOCAL_DEV": "true",
                "LOCAL_DEVRUN_SCENARIO": "scenario1,scenario2,scenario3",
            },
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._run_scenarios == ["scenario1", "scenario2", "scenario3"]

    def test_init_with_scenarios_whitespace(self):
        """Test initialization with scenarios containing whitespace"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {
                "IS_LOCAL_DEV": "true",
                "LOCAL_DEVRUN_SCENARIO": " scenario1 , scenario2 , scenario3 ",
            },
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._run_scenarios == ["scenario1", "scenario2", "scenario3"]

    def test_init_with_empty_scenarios(self):
        """Test initialization with empty scenarios"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper._run_scenarios == []

    def test_is_local_dev_without_scenarios_when_true(self):
        """Test is_local_dev returns True when IS_LOCAL_DEV=true and no scenarios"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_local_dev() is True

    def test_is_local_dev_without_scenarios_when_false(self):
        """Test is_local_dev returns False when IS_LOCAL_DEV=false"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "false", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_local_dev() is False

    def test_is_local_dev_with_matching_scenario(self):
        """Test is_local_dev with matching scenario"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_local_dev([RunScenario.AARON_LOCAL_DEV]) is True

    def test_is_local_dev_with_non_matching_scenario(self):
        """Test is_local_dev with non-matching scenario"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "other_scenario"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_local_dev([RunScenario.AARON_LOCAL_DEV]) is False

    def test_is_local_dev_with_empty_scenarios_list(self):
        """Test is_local_dev with empty scenarios list"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # Empty scenarios list means no specific scenario check
            # So it should return True since IS_LOCAL_DEV is true
            assert helper.is_local_dev([]) is True

    def test_is_local_dev_with_not_inited_raises_error(self):
        """Test is_local_dev raises error when not initialized"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        helper = LocalEnvHelper()
        helper._env_inited = False

        with pytest.raises(RuntimeError, match="Environment not initialized"):
            helper.is_local_dev()

    def test_is_aaron_local_dev_when_true(self):
        """Test is_aaron_local_dev returns True when in Aaron local dev"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_aaron_local_dev() is True

    def test_is_aaron_local_dev_when_false(self):
        """Test is_aaron_local_dev returns False when not in Aaron local dev"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "other_scenario"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_aaron_local_dev() is False

    def test_get_current_scenarios(self):
        """Test get_current_scenarios returns copy of scenarios"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "scenario1,scenario2"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            scenarios = helper.get_current_scenarios()
            assert scenarios == ["scenario1", "scenario2"]

    def test_get_current_scenarios_returns_copy(self):
        """Test get_current_scenarios returns a copy, not reference"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "scenario1"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            scenarios = helper.get_current_scenarios()
            scenarios.append("scenario2")
            # Original should not be modified
            assert helper._run_scenarios == ["scenario1"]

    def test_get_current_scenarios_not_inited_raises_error(self):
        """Test get_current_scenarios raises error when not initialized"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        helper = LocalEnvHelper()
        helper._env_inited = False

        with pytest.raises(RuntimeError, match="Environment not initialized"):
            helper.get_current_scenarios()

    def test_is_scenario_with_matching_scenario(self):
        """Test is_scenario returns True for matching scenario"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_scenario(RunScenario.AARON_LOCAL_DEV) is True

    def test_is_scenario_with_non_matching_scenario(self):
        """Test is_scenario returns False for non-matching scenario"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "other_scenario"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_scenario(RunScenario.AARON_LOCAL_DEV) is False

    def test_is_scenario_with_empty_scenarios(self):
        """Test is_scenario returns False when no scenarios set"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert helper.is_scenario(RunScenario.AARON_LOCAL_DEV) is False


class TestGlobalEnvHelperFunctions:
    """Tests for global env_helper functions"""

    def test_is_local_dev_global_function(self):
        """Test global is_local_dev function"""
        # Note: The global instance is initialized at module load time
        # so these tests verify the functions work correctly
        from app.infra.common.helper.env_helper import is_local_dev

        # The function should return a boolean
        result = is_local_dev()
        assert isinstance(result, bool)

    def test_is_local_dev_global_with_scenarios(self):
        """Test global is_local_dev function with scenarios"""
        from app.infra.common.helper.env_helper import is_local_dev, RunScenario

        # Test that the function accepts scenarios parameter
        result = is_local_dev([RunScenario.AARON_LOCAL_DEV])
        assert isinstance(result, bool)

    def test_is_aaron_local_dev_global_function(self):
        """Test global is_aaron_local_dev function"""
        from app.infra.common.helper.env_helper import is_aaron_local_dev

        # The function should return a boolean
        result = is_aaron_local_dev()
        assert isinstance(result, bool)

    def test_get_current_scenarios_global_function(self):
        """Test global get_current_scenarios function"""
        from app.infra.common.helper.env_helper import get_current_scenarios

        # The function should return a list
        scenarios = get_current_scenarios()
        assert isinstance(scenarios, list)

    def test_is_scenario_global_function(self):
        """Test global is_scenario function"""
        from app.infra.common.helper.env_helper import is_scenario, RunScenario

        # The function should return a boolean
        result = is_scenario(RunScenario.AARON_LOCAL_DEV)
        assert isinstance(result, bool)


class TestEnvHelperEdgeCases:
    """Tests for edge cases in env_helper"""

    def test_init_with_default_env_vars(self):
        """Test initialization when env vars use defaults"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict("os.environ", {}, clear=True):
            helper = LocalEnvHelper()
            assert helper._is_local_dev is False
            assert helper._run_scenarios == []

    def test_init_with_invalid_scenario_format(self):
        """Test initialization with invalid scenario format"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        # Should handle gracefully - split on comma even with odd formatting
        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ",,test,,,"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            assert "test" in helper._run_scenarios

    def test_multiple_helper_instances(self):
        """Test multiple helper instances with same env"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper1 = LocalEnvHelper()
            helper2 = LocalEnvHelper()
            assert helper1.is_local_dev() == helper2.is_local_dev()

    def test_scenario_check_with_multiple_scenarios_in_list(self):
        """Test scenario checking with multiple scenarios in check list"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "scenario2"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # Create a mock enum for testing
            from enum import Enum

            class TestScenario(Enum):
                SCENARIO1 = "scenario1"
                SCENARIO2 = "scenario2"
                SCENARIO3 = "scenario3"

            assert (
                helper.is_local_dev([TestScenario.SCENARIO1, TestScenario.SCENARIO2])
                is True
            )

    def test_is_local_dev_false_with_scenarios_when_not_local(self):
        """Test is_local_dev returns False when not local dev even with scenarios"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "false", "LOCAL_DEVRUN_SCENARIO": "aaron_local_dev"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # Should return False because _is_local_dev is False
            assert helper.is_local_dev([RunScenario.AARON_LOCAL_DEV]) is False
