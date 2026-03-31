"""单元测试 - logic/tool/get_agent_config 模块"""

import pytest
from unittest.mock import AsyncMock, patch


@pytest.mark.asyncio
class TestGetAgentConfig:
    """测试 get_agent_config 函数"""

    @patch("app.logic.tool.get_agent_config.agent_factory_service")
    async def test_get_by_agent_id(self, m_service):
        """测试通过 agent_id 获取配置"""
        from app.logic.tool.get_agent_config import get_agent_config

        expected_config = {"agent_id": "123", "name": "Test Agent"}
        m_service.get_agent_config = AsyncMock(return_value=expected_config)

        result = await get_agent_config(agent_id="123")

        assert result == expected_config
        m_service.get_agent_config.assert_called_once_with("123")

    @patch("app.logic.tool.get_agent_config.agent_factory_service")
    async def test_get_by_agent_key(self, m_service):
        """测试通过 agent_key 获取配置"""
        from app.logic.tool.get_agent_config import get_agent_config

        expected_config = {"agent_key": "key123", "name": "Test Agent"}
        m_service.get_agent_config_by_key = AsyncMock(return_value=expected_config)

        result = await get_agent_config(agent_key="key123")

        assert result == expected_config
        m_service.get_agent_config_by_key.assert_called_once_with("key123")

    @patch("app.logic.tool.get_agent_config.agent_factory_service")
    async def test_error_when_no_params(self, m_service):
        """测试当没有提供任何参数时抛出 ValueError"""
        from app.logic.tool.get_agent_config import get_agent_config

        with pytest.raises(ValueError) as exc_info:
            await get_agent_config()

        assert "必须提供agent_id或agent_key参数" in str(exc_info.value)

    @patch("app.logic.tool.get_agent_config.agent_factory_service")
    async def test_error_when_both_params(self, m_service):
        """测试当同时提供 agent_id 和 agent_key 时抛出 ValueError"""
        from app.logic.tool.get_agent_config import get_agent_config

        with pytest.raises(ValueError) as exc_info:
            await get_agent_config(agent_id="123", agent_key="key123")

        assert "agent_id和agent_key不能同时提供" in str(exc_info.value)

    @patch("app.logic.tool.get_agent_config.agent_factory_service")
    async def test_service_error_propagation(self, m_service):
        """测试服务错误传播"""
        from app.logic.tool.get_agent_config import get_agent_config

        m_service.get_agent_config = AsyncMock(
            side_effect=ConnectionError("Service unavailable")
        )

        with pytest.raises(ConnectionError) as exc_info:
            await get_agent_config(agent_id="123")

        assert "Service unavailable" in str(exc_info.value)
