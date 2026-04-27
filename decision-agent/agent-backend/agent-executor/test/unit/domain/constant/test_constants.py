"""单元测试 - domain/constant 模块"""

from unittest import TestCase

from app.domain.constant.agent_version import AGENT_VERSION_V0, AGENT_VERSION_LATEST
from app.domain.constant.agent_cache_constants import (
    AGENT_CACHE_TTL,
    AGENT_CACHE_DATA_UPDATE_PASS_SECOND,
)


class TestAgentVersion(TestCase):
    """测试 agent_version 常量"""

    def test_agent_version_v0(self):
        """测试AGENT_VERSION_V0常量"""
        self.assertEqual(AGENT_VERSION_V0, "v0")

    def test_agent_version_latest(self):
        """测试AGENT_VERSION_LATEST常量"""
        self.assertEqual(AGENT_VERSION_LATEST, "latest")

    def test_version_values_are_strings(self):
        """测试版本值都是字符串"""
        self.assertIsInstance(AGENT_VERSION_V0, str)
        self.assertIsInstance(AGENT_VERSION_LATEST, str)


class TestAgentCacheConstants(TestCase):
    """测试 agent_cache_constants 常量"""

    def test_agent_cache_ttl(self):
        """测试AGENT_CACHE_TTL常量"""
        self.assertEqual(AGENT_CACHE_TTL, 60)

    def test_agent_cache_data_update_pass_second(self):
        """测试AGENT_CACHE_DATA_UPDATE_PASS_SECOND常量"""
        self.assertEqual(AGENT_CACHE_DATA_UPDATE_PASS_SECOND, 10)

    def test_ttl_type(self):
        """测试TTL是整数"""
        self.assertIsInstance(AGENT_CACHE_TTL, int)

    def test_update_pass_type(self):
        """测试更新间隔是整数"""
        self.assertIsInstance(AGENT_CACHE_DATA_UPDATE_PASS_SECOND, int)

    def test_ttl_positive(self):
        """测试TTL为正数"""
        self.assertGreater(AGENT_CACHE_TTL, 0)

    def test_update_pass_positive(self):
        """测试更新间隔为正数"""
        self.assertGreater(AGENT_CACHE_DATA_UPDATE_PASS_SECOND, 0)

    def test_ttl_greater_than_update_pass(self):
        """测试TTL大于更新间隔"""
        self.assertGreater(AGENT_CACHE_TTL, AGENT_CACHE_DATA_UPDATE_PASS_SECOND)
