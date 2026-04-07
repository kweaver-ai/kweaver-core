"""
执行器 DTO 单元测试

测试执行器数据传输对象。
"""
import pytest
from pydantic import ValidationError

from src.infrastructure.executors.dto import (
    ExecutorExecuteRequest,
    ExecutorExecuteResponse,
    ExecutorHealthResponse,
    ExecutorContainerInfo,
)


class TestExecutorExecuteRequest:
    """执行器执行请求测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        request = ExecutorExecuteRequest(
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python"
        )

        assert request.execution_id == "exec-123"
        assert request.session_id == "sess-456"
        assert request.code == "print('hello')"
        assert request.language == "python"
        assert request.event == {}
        assert request.timeout == 300  # default
        assert request.env_vars == {}  # default

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        request = ExecutorExecuteRequest(
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            event={"name": "World"},
            timeout=60,
            env_vars={"DEBUG": "true"}
        )

        assert request.execution_id == "exec-123"
        assert request.session_id == "sess-456"
        assert request.code == "print('hello')"
        assert request.language == "python"
        assert request.event == {"name": "World"}
        assert request.timeout == 60
        assert request.env_vars == {"DEBUG": "true"}

    def test_timeout_minimum(self):
        """测试超时最小值"""
        request = ExecutorExecuteRequest(
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            timeout=1
        )
        assert request.timeout == 1

    def test_timeout_maximum(self):
        """测试超时最大值"""
        request = ExecutorExecuteRequest(
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            timeout=3600
        )
        assert request.timeout == 3600

    def test_timeout_below_minimum(self):
        """测试超时低于最小值"""
        with pytest.raises(ValidationError):
            ExecutorExecuteRequest(
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                timeout=0
            )

    def test_timeout_above_maximum(self):
        """测试超时高于最大值"""
        with pytest.raises(ValidationError):
            ExecutorExecuteRequest(
                execution_id="exec-123",
                session_id="sess-456",
                code="print('hello')",
                language="python",
                timeout=3601
            )

    def test_missing_required_fields(self):
        """测试缺少必填字段"""
        with pytest.raises(ValidationError):
            ExecutorExecuteRequest(
                execution_id="exec-123",
                # missing session_id
                code="print('hello')",
                language="python"
            )

    def test_model_dump(self):
        """测试序列化为字典"""
        request = ExecutorExecuteRequest(
            execution_id="exec-123",
            session_id="sess-456",
            code="print('hello')",
            language="python",
            event={"name": "World"},
            timeout=60
        )

        data = request.model_dump()

        assert data["execution_id"] == "exec-123"
        assert data["session_id"] == "sess-456"
        assert data["code"] == "print('hello')"
        assert data["language"] == "python"
        assert data["event"] == {"name": "World"}
        assert data["timeout"] == 60

    def test_json_schema_examples(self):
        """测试 JSON schema 示例"""
        # Should have examples defined
        assert "examples" in ExecutorExecuteRequest.model_config.get("json_schema_extra", {})


class TestExecutorExecuteResponse:
    """执行器执行响应测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        response = ExecutorExecuteResponse(
            execution_id="exec-123",
            status="submitted"
        )

        assert response.execution_id == "exec-123"
        assert response.status == "submitted"
        assert response.message == ""  # default

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        response = ExecutorExecuteResponse(
            execution_id="exec-123",
            status="completed",
            message="Execution finished successfully"
        )

        assert response.execution_id == "exec-123"
        assert response.status == "completed"
        assert response.message == "Execution finished successfully"

    def test_missing_required_fields(self):
        """测试缺少必填字段"""
        with pytest.raises(ValidationError):
            ExecutorExecuteResponse(
                execution_id="exec-123"
                # missing status
            )

    def test_from_json(self):
        """测试从 JSON 创建"""
        response = ExecutorExecuteResponse(**{
            "execution_id": "exec-123",
            "status": "completed",
            "message": "Done"
        })

        assert response.execution_id == "exec-123"
        assert response.status == "completed"
        assert response.message == "Done"


class TestExecutorHealthResponse:
    """执行器健康检查响应测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        response = ExecutorHealthResponse(
            status="healthy"
        )

        assert response.status == "healthy"
        assert response.version == "1.0.0"  # default
        assert response.uptime_seconds is None
        assert response.active_executions is None

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        response = ExecutorHealthResponse(
            status="healthy",
            version="2.0.0",
            uptime_seconds=3600.5,
            active_executions=5
        )

        assert response.status == "healthy"
        assert response.version == "2.0.0"
        assert response.uptime_seconds == 3600.5
        assert response.active_executions == 5

    def test_unhealthy_status(self):
        """测试不健康状态"""
        response = ExecutorHealthResponse(
            status="unhealthy"
        )

        assert response.status == "unhealthy"

    def test_from_json(self):
        """测试从 JSON 创建"""
        response = ExecutorHealthResponse(**{
            "status": "healthy",
            "version": "1.5.0",
            "uptime_seconds": 100.0,
            "active_executions": 3
        })

        assert response.status == "healthy"
        assert response.version == "1.5.0"
        assert response.uptime_seconds == 100.0
        assert response.active_executions == 3

    def test_from_json_minimal(self):
        """测试从最小 JSON 创建"""
        response = ExecutorHealthResponse(**{
            "status": "healthy"
        })

        assert response.status == "healthy"
        assert response.version == "1.0.0"


class TestExecutorContainerInfo:
    """执行器容器信息测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        info = ExecutorContainerInfo(
            container_id="container-123",
            container_name="sandbox-sess-123"
        )

        assert info.container_id == "container-123"
        assert info.container_name == "sandbox-sess-123"
        assert info.executor_port == 8080  # default

    def test_create_with_custom_port(self):
        """测试使用自定义端口创建"""
        info = ExecutorContainerInfo(
            container_id="container-123",
            container_name="sandbox-sess-123",
            executor_port=9090
        )

        assert info.executor_port == 9090

    def test_executor_url_default_port(self):
        """测试执行器 URL（默认端口）"""
        info = ExecutorContainerInfo(
            container_id="container-123",
            container_name="sandbox-sess-123"
        )

        assert info.executor_url == "http://sandbox-sess-123:8080"

    def test_executor_url_custom_port(self):
        """测试执行器 URL（自定义端口）"""
        info = ExecutorContainerInfo(
            container_id="container-123",
            container_name="sandbox-sess-123",
            executor_port=9090
        )

        assert info.executor_url == "http://sandbox-sess-123:9090"

    def test_executor_url_with_underscored_name(self):
        """测试执行器 URL（带下划线的名称）"""
        info = ExecutorContainerInfo(
            container_id="container-123",
            container_name="sandbox_session_test"
        )

        assert info.executor_url == "http://sandbox_session_test:8080"

    def test_is_dataclass(self):
        """测试是数据类"""
        from dataclasses import is_dataclass

        assert is_dataclass(ExecutorContainerInfo)
