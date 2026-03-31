"""
本地开发环境变量处理工具
参考 Go 版本的 env_local_dev.go 实现
"""

import os
from enum import Enum
from typing import Optional, List


class RunScenario(Enum):
    """运行场景枚举"""

    AARON_LOCAL_DEV = "aaron_local_dev"
    # 可以根据需要添加更多场景


class LocalEnvHelper:
    """环境变量处理助手"""

    def __init__(self):
        self._env_inited = False
        self._is_local_dev = None
        self._run_scenarios = []
        self._init_env()

    def _init_env(self) -> None:
        """初始化环境变量"""
        try:
            # 检查环境变量是否已设置
            self._is_local_dev = os.getenv("IS_LOCAL_DEV", "false").lower() == "true"

            # 解析运行场景（支持逗号分隔的多个场景）
            run_scenario_str = os.getenv("LOCAL_DEVRUN_SCENARIO", "")
            if run_scenario_str:
                self._run_scenarios = [
                    scenario.strip()
                    for scenario in run_scenario_str.split(",")
                    if scenario.strip()
                ]
            else:
                self._run_scenarios = []

            self._env_inited = True
        except Exception as e:
            raise RuntimeError(f"Failed to initialize environment: {e}")

    def is_local_dev(self, run_scenarios: Optional[List[RunScenario]] = None) -> bool:
        """
        判断是否为本地开发环境

        Args:
            run_scenarios: 运行场景列表
                - 如果为空，则判断是否为本地开发环境
                - 如果不为空，则判断是否为本地开发环境且运行场景在 run_scenarios 中

        Returns:
            bool: 是否为本地开发环境
        """
        if not self._env_inited:
            raise RuntimeError("Environment not initialized")

        # 第一层判断：是否为本地开发环境
        if not self._is_local_dev:
            return False

        # 第二层判断：运行场景
        if not run_scenarios:
            return True

        # 检查当前运行场景列表是否为空
        if not self._run_scenarios:
            return False

        # 检查运行场景是否在指定列表中
        for scenario in run_scenarios:
            if scenario.value in self._run_scenarios:
                return True

        return False

    def is_aaron_local_dev(self) -> bool:
        """判断是否为 Aaron 本地开发环境"""
        return self.is_local_dev([RunScenario.AARON_LOCAL_DEV])

    def get_current_scenarios(self) -> List[str]:
        """获取当前运行场景列表"""
        if not self._env_inited:
            raise RuntimeError("Environment not initialized")
        return self._run_scenarios.copy()

    def is_scenario(self, scenario: RunScenario) -> bool:
        """判断当前是否为指定场景"""
        return scenario.value in self._run_scenarios


# 全局实例
_env_helper = LocalEnvHelper()


def is_local_dev(run_scenarios: Optional[List[RunScenario]] = None) -> bool:
    """
    判断是否为本地开发环境（全局函数）

    Args:
        run_scenarios: 运行场景列表

    Returns:
        bool: 是否为本地开发环境
    """
    return _env_helper.is_local_dev(run_scenarios)


def is_aaron_local_dev() -> bool:
    """判断是否为 Aaron 本地开发环境（全局函数）"""
    return _env_helper.is_aaron_local_dev()


def get_current_scenarios() -> List[str]:
    """获取当前运行场景列表（全局函数）"""
    return _env_helper.get_current_scenarios()


def is_scenario(scenario: RunScenario) -> bool:
    """判断当前是否为指定场景（全局函数）"""
    return _env_helper.is_scenario(scenario)


# 便捷函数
# 可以根据需要添加更多场景判断函数
