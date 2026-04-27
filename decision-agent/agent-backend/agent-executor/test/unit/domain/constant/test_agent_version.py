"""单元测试 - domain/constant/agent_version 模块"""

from app.domain.constant.agent_version import AGENT_VERSION_V0, AGENT_VERSION_LATEST


class TestAgentVersionConstants:
    """测试 Agent 版本常量"""

    def test_agent_version_v0(self):
        """测试 AGENT_VERSION_V0 常量"""
        assert AGENT_VERSION_V0 == "v0"

    def test_agent_version_latest(self):
        """测试 AGENT_VERSION_LATEST 常量"""
        assert AGENT_VERSION_LATEST == "latest"
