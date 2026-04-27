# -*- coding: utf-8 -*-
"""单元测试 - resume_agent_v2 模块"""

import pytest
from unittest.mock import MagicMock, patch


class TestResumeAgent:
    """测试 resume_agent 函数"""

    @pytest.fixture
    def mock_request(self):
        """创建 mock Request"""
        return MagicMock()

    @pytest.fixture
    def mock_resume_req(self):
        """创建 mock ResumeAgentRequest"""
        req = MagicMock()
        req.agent_run_id = "test_run_id"
        req.resume_info = MagicMock()
        req.resume_info.action = "confirm"
        return req

    @pytest.mark.asyncio
    async def test_resume_agent_success(self, mock_request, mock_resume_req):
        """测试成功恢复Agent"""
        mock_agent = MagicMock()
        mock_agent_core = MagicMock()

        with patch(
            "app.router.agent_controller_pkg.resume_agent_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get.return_value = (mock_agent, mock_agent_core)

            with patch(
                "app.router.agent_controller_pkg.resume_agent_v2.create_resume_generator"
            ) as mock_create:

                async def mock_gen():
                    yield "test data"

                mock_create.return_value = mock_gen()

                from app.router.agent_controller_pkg.resume_agent_v2 import (
                    resume_agent,
                )

                result = await resume_agent(mock_request, mock_resume_req)

                assert result is not None
                mock_manager.get.assert_called_once_with("test_run_id")
                mock_create.assert_called_once()

    @pytest.mark.asyncio
    async def test_resume_agent_not_found(self, mock_request, mock_resume_req):
        """测试Agent实例不存在"""
        with patch(
            "app.router.agent_controller_pkg.resume_agent_v2.agent_instance_manager"
        ) as mock_manager:
            mock_manager.get.return_value = None

            from app.router.agent_controller_pkg.resume_agent_v2 import (
                resume_agent,
            )
            from fastapi.responses import JSONResponse

            result = await resume_agent(mock_request, mock_resume_req)

            assert isinstance(result, JSONResponse)
            assert result.status_code == 404
