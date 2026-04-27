# -*- coding: utf-8 -*-
"""单元测试 - app/infra/common/helper/env_helper/local_env_helper.py 补充测试"""

import pytest
from unittest.mock import patch


class TestLocalEnvHelperInitException:
    """测试 LocalEnvHelper 初始化异常场景"""

    def test_init_with_os_getenv_exception(self):
        """测试初始化时 os.getenv 抛出异常"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch("os.getenv", side_effect=OSError("Environment error")):
            with pytest.raises(RuntimeError) as exc_info:
                LocalEnvHelper()

            assert "Failed to initialize environment" in str(exc_info.value)

    def test_init_with_unexpected_exception(self):
        """测试初始化时发生意外异常"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch("os.getenv", side_effect=ValueError("Unexpected error")):
            with pytest.raises(RuntimeError) as exc_info:
                LocalEnvHelper()

            assert "Failed to initialize environment" in str(exc_info.value)


class TestLocalEnvHelperIsLocalDevEmptyScenarios:
    """测试 is_local_dev 方法空场景分支"""

    def test_is_local_dev_with_scenarios_but_empty_run_scenarios(self):
        """测试 is_local_dev 传入场景参数但当前运行场景为空"""
        from app.infra.common.helper.env_helper import LocalEnvHelper, RunScenario

        with patch.dict(
            "os.environ",
            {
                "IS_LOCAL_DEV": "true",
                "LOCAL_DEVRUN_SCENARIO": "",  # 空的运行场景
            },
            clear=True,
        ):
            helper = LocalEnvHelper()
            # 传入场景参数，但 _run_scenarios 为空
            result = helper.is_local_dev([RunScenario.AARON_LOCAL_DEV])
            # 应该返回 False，因为 _run_scenarios 为空
            assert result is False

    def test_is_local_dev_with_multiple_scenarios_but_empty_run_scenarios(self):
        """测试 is_local_dev 传入多个场景参数但当前运行场景为空"""
        from app.infra.common.helper.env_helper import LocalEnvHelper
        from enum import Enum

        class TestScenario(Enum):
            SCENARIO1 = "scenario1"
            SCENARIO2 = "scenario2"

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            result = helper.is_local_dev(
                [TestScenario.SCENARIO1, TestScenario.SCENARIO2]
            )
            assert result is False

    def test_is_local_dev_true_without_scenario_param(self):
        """测试 is_local_dev 为 True 时不传场景参数"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # 不传场景参数，应该返回 True
            result = helper.is_local_dev()
            assert result is True

    def test_is_local_dev_true_with_none_scenario_param(self):
        """测试 is_local_dev 为 True 时传入 None 场景参数"""
        from app.infra.common.helper.env_helper import LocalEnvHelper

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": ""},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # 传入 None，应该返回 True
            result = helper.is_local_dev(None)
            assert result is True


class TestLocalEnvHelperScenarioMatching:
    """测试场景匹配逻辑"""

    def test_is_local_dev_scenario_partial_match(self):
        """测试部分场景匹配"""
        from app.infra.common.helper.env_helper import LocalEnvHelper
        from enum import Enum

        class TestScenario(Enum):
            SCENARIO1 = "scenario1"
            SCENARIO2 = "scenario2"
            SCENARIO3 = "scenario3"

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "scenario2"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # 只有 SCENARIO2 匹配
            result = helper.is_local_dev(
                [TestScenario.SCENARIO1, TestScenario.SCENARIO2, TestScenario.SCENARIO3]
            )
            assert result is True

    def test_is_local_dev_no_scenario_match(self):
        """测试没有场景匹配"""
        from app.infra.common.helper.env_helper import LocalEnvHelper
        from enum import Enum

        class TestScenario(Enum):
            SCENARIO1 = "scenario1"
            SCENARIO2 = "scenario2"

        with patch.dict(
            "os.environ",
            {"IS_LOCAL_DEV": "true", "LOCAL_DEVRUN_SCENARIO": "scenario3"},
            clear=True,
        ):
            helper = LocalEnvHelper()
            # 没有场景匹配
            result = helper.is_local_dev(
                [TestScenario.SCENARIO1, TestScenario.SCENARIO2]
            )
            assert result is False
