"""测试 - prepare 准备函数 - 简化版本"""

import pytest
from unittest.mock import AsyncMock, MagicMock, patch
from fastapi import Request


@pytest.mark.asyncio
class TestPrepare:
    """测试 prepare 函数"""

    async def test_with_agent_config(self):
        """测试提供agent_config"""
        request = MagicMock(spec=Request)
        request.headers = {}
        req = MagicMock()
        req.agent_config = MagicMock()
        req.agent_id = None
        req.agent_version = "1.0"
        req.agent_input = MagicMock()
        req.options = None

        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare
        from app.domain.vo.agentvo import AgentConfigVo

        agent_config_dict = {"agent_id": "test123", "name": "Test", "llms": []}
        req.agent_config = AgentConfigVo(**agent_config_dict)

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.set_headers = MagicMock()
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            result = await prepare(request, req, "acc123", "premium", "dom456")

        assert result[0] is req.agent_config

    async def test_fetches_agent_config(self):
        """测试获取agent配置"""
        request = MagicMock(spec=Request)
        request.headers = {}
        req = MagicMock()
        req.agent_config = None
        req.agent_id = "agent-456"
        req.agent_version = "1.0"
        req.agent_input = MagicMock()
        req.options = None

        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare

        config_json = '{"agent_id": "agent-456", "llms": []}'

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.set_headers = MagicMock()
            mock_service.get_agent_config_by_agent_id_and_version = AsyncMock(
                return_value={"config": config_json}
            )
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            result = await prepare(request, req, "acc123", "premium", "dom456")

        # Verify get_agent_config_by_agent_id_and_version was called
        mock_service.get_agent_config_by_agent_id_and_version.assert_called_once_with(
            "agent-456", "1.0"
        )

    async def test_processes_options(self):
        """测试处理选项"""
        request = MagicMock(spec=Request)
        request.headers = {}
        req = MagicMock()
        req.agent_config = MagicMock()
        req.agent_id = "agent-789"
        req.agent_version = "1.0"
        req.agent_input = MagicMock()
        req.options = MagicMock()

        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.set_headers = MagicMock()
            mock_service.get_agent_config_by_agent_id_and_version = AsyncMock(
                return_value={"config": '{"llms": []}'}
            )
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            result = await prepare(request, req, "acc123", "premium", "dom456")

        # Verify prepare completed successfully
        assert result is not None
        assert isinstance(result, tuple)

    async def test_checks_permission(self):
        """测试检查权限"""
        request = MagicMock(spec=Request)
        request.headers = {}
        req = MagicMock()
        req.agent_config = MagicMock()
        req.agent_config.agent_id = "perm-123"
        req.agent_id = "agent-789"
        req.agent_version = "1.0"
        req.agent_input = MagicMock()
        req.options = None

        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare
        from app.common.errors import AgentPermissionException

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.set_headers = MagicMock()
            mock_service.get_agent_config_by_agent_id_and_version = AsyncMock(
                return_value={"config": "{}"}
            )
            mock_service.check_agent_permission = AsyncMock(return_value=False)

            with pytest.raises(AgentPermissionException):
                await prepare(request, req, "acc123", "premium", "dom456")

    async def test_returns_tuple(self):
        """测试返回三元组"""
        request = MagicMock(spec=Request)
        request.headers = {"h1": "v1"}
        req = MagicMock()
        req.agent_config = MagicMock()
        req.agent_id = "agent-789"
        req.agent_version = "1.0"
        req.agent_input = MagicMock()
        req.options = None

        from app.router.agent_controller_pkg.run_agent_v2.prepare import prepare

        with patch(
            "app.router.agent_controller_pkg.run_agent_v2.prepare.agent_factory_service"
        ) as mock_service:
            mock_service.set_headers = MagicMock()
            mock_service.get_agent_config_by_agent_id_and_version = AsyncMock(
                return_value={"config": "{}"}
            )
            mock_service.check_agent_permission = AsyncMock(return_value=True)

            result = await prepare(request, req, "acc123", "premium", "dom456")

        assert isinstance(result, tuple)
        assert len(result) == 3
