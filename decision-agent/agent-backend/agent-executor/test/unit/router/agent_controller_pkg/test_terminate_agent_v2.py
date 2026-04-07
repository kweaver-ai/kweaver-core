"""单元测试 - terminate_agent_v2 模块"""

import pytest
from unittest.mock import MagicMock, AsyncMock, patch
from fastapi import Request


@pytest.mark.asyncio
class TestTerminateAgent:
    """测试 terminate_agent 函数"""

    async def test_terminate_agent_success(self):
        """测试成功终止 Agent"""
        from app.router.agent_controller_pkg.terminate_agent_v2 import terminate_agent
        from app.router.agent_controller_pkg.rdto.v2.req.terminate_agent import (
            TerminateAgentRequest,
        )

        request = MagicMock(spec=Request)
        req = TerminateAgentRequest(agent_run_id="test-run-123")

        mock_agent = AsyncMock()
        mock_agent_core = MagicMock()
        mock_agent_core.cleanup = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.terminate_agent_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get = MagicMock(return_value=(mock_agent, mock_agent_core))
            mock_manager.remove = MagicMock()

            with patch(
                "app.router.agent_controller_pkg.terminate_agent_v2.StandLogger"
            ) as mock_logger:
                response = await terminate_agent(request, req)

        assert response.status_code == 204
        mock_agent.terminate.assert_called_once()
        mock_manager.remove.assert_called_once_with("test-run-123")
        mock_agent_core.cleanup.assert_called_once()

    async def test_terminate_agent_not_found(self):
        """测试 Agent 实例不存在"""
        from app.router.agent_controller_pkg.terminate_agent_v2 import terminate_agent
        from app.router.agent_controller_pkg.rdto.v2.req.terminate_agent import (
            TerminateAgentRequest,
        )

        request = MagicMock(spec=Request)
        req = TerminateAgentRequest(agent_run_id="nonexistent-run")

        with patch(
            "app.router.agent_controller_pkg.terminate_agent_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get = MagicMock(return_value=None)

            with patch(
                "app.router.agent_controller_pkg.terminate_agent_v2.StandLogger"
            ):
                response = await terminate_agent(request, req)

        assert response.status_code == 404

    async def test_terminate_agent_with_exception(self):
        """测试终止 Agent 时发生异常"""
        from app.router.agent_controller_pkg.terminate_agent_v2 import terminate_agent
        from app.router.agent_controller_pkg.rdto.v2.req.terminate_agent import (
            TerminateAgentRequest,
        )

        request = MagicMock(spec=Request)
        req = TerminateAgentRequest(agent_run_id="test-run-123")

        mock_agent = AsyncMock()
        mock_agent.terminate = AsyncMock(side_effect=Exception("Termination error"))
        mock_agent_core = MagicMock()
        mock_agent_core.cleanup = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.terminate_agent_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get = MagicMock(return_value=(mock_agent, mock_agent_core))
            mock_manager.remove = MagicMock()

            with patch(
                "app.router.agent_controller_pkg.terminate_agent_v2.StandLogger"
            ):
                response = await terminate_agent(request, req)

        # Should still return 204 and clean up even if terminate fails
        assert response.status_code == 204
        mock_manager.remove.assert_called_once_with("test-run-123")
        mock_agent_core.cleanup.assert_called_once()
