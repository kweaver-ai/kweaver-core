"""
会话 DTO 单元测试

测试 SessionDTO 的功能。
"""
import pytest
from datetime import datetime

from src.application.dtos.session_dto import SessionDTO
from src.domain.entities.session import Session
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus


class TestSessionDTO:
    """会话 DTO 测试"""

    def test_create_with_required_fields(self):
        """测试使用必填字段创建"""
        dto = SessionDTO(
            id="test-session",
            template_id="python-test",
            status="running",
            resource_limit={"cpu": "1", "memory": "512Mi", "disk": "1Gi"},
            workspace_path="/workspace/test",
            runtime_type="python3.11"
        )

        assert dto.id == "test-session"
        assert dto.template_id == "python-test"
        assert dto.status == "running"
        assert dto.resource_limit == {"cpu": "1", "memory": "512Mi", "disk": "1Gi"}
        assert dto.workspace_path == "/workspace/test"
        assert dto.runtime_type == "python3.11"
        assert dto.runtime_node is None
        assert dto.container_id is None
        assert dto.pod_name is None
        assert dto.env_vars == {}  # Default from __post_init__
        assert dto.timeout == 300
        assert dto.language_runtime == "python3.11"
        assert dto.python_package_index_url == "https://pypi.org/simple/"
        assert dto.requested_dependencies == []
        assert dto.installed_dependencies == []
        assert dto.dependency_install_status == "pending"
        assert dto.created_at is not None  # Default from __post_init__
        assert dto.updated_at is not None  # Default from __post_init__
        assert dto.last_activity_at is not None  # Default from __post_init__

    def test_create_with_all_fields(self):
        """测试使用所有字段创建"""
        now = datetime.now()
        dto = SessionDTO(
            id="test-session",
            template_id="python-test",
            status="completed",
            resource_limit={"cpu": "2", "memory": "1Gi", "disk": "10Gi"},
            workspace_path="/workspace/test",
            runtime_type="python3.11",
            runtime_node="node-1",
            container_id="container-123",
            pod_name="pod-456",
            env_vars={"DEBUG": "true"},
            timeout=600,
            language_runtime="python3.11",
            python_package_index_url="https://mirror.example/simple",
            requested_dependencies=[{"name": "requests", "version": "==2.31.0"}],
            installed_dependencies=[],
            dependency_install_status="completed",
            created_at=now,
            updated_at=now,
            completed_at=now,
            last_activity_at=now
        )

        assert dto.runtime_node == "node-1"
        assert dto.container_id == "container-123"
        assert dto.pod_name == "pod-456"
        assert dto.env_vars == {"DEBUG": "true"}
        assert dto.timeout == 600
        assert dto.python_package_index_url == "https://mirror.example/simple"
        assert dto.completed_at == now

    def test_post_init_defaults(self):
        """测试 __post_init__ 默认值"""
        dto = SessionDTO(
            id="test-session",
            template_id="python-test",
            status="running",
            resource_limit={"cpu": "1", "memory": "512Mi", "disk": "1Gi"},
            workspace_path="/workspace/test",
            runtime_type="python3.11",
            env_vars=None,
            created_at=None,
            updated_at=None,
            last_activity_at=None
        )

        # __post_init__ should set defaults
        assert dto.env_vars == {}
        assert dto.created_at is not None
        assert dto.updated_at is not None
        assert dto.last_activity_at is not None
        assert dto.requested_dependencies == []
        assert dto.installed_dependencies == []

    def test_from_entity(self):
        """测试从领域实体创建 DTO"""
        session = Session(
            id="test-session",
            template_id="python-test",
            status=SessionStatus.CREATING,
            resource_limit=ResourceLimit(
                cpu="1",
                memory="512Mi",
                disk="1Gi"
            ),
            workspace_path="/workspace/test",
            runtime_type="python3.11",
            requested_dependencies=["requests==2.31.0"],
        )

        dto = SessionDTO.from_entity(session)

        assert dto.id == "test-session"
        assert dto.template_id == "python-test"
        assert dto.status == SessionStatus.CREATING.value
        assert dto.resource_limit == {
            "cpu": "1",
            "memory": "512Mi",
            "disk": "1Gi",
            "max_processes": 128
        }
        assert dto.workspace_path == "/workspace/test"
        assert dto.runtime_type == "python3.11"
        assert dto.language_runtime == "python3.11"
        assert dto.requested_dependencies == [
            {"name": "requests", "version": "==2.31.0"}
        ]

    def test_from_entity_with_all_fields(self):
        """测试从领域实体创建 DTO（所有字段）"""
        now = datetime.now()
        session = Session(
            id="test-session",
            template_id="python-test",
            status=SessionStatus.RUNNING,
            resource_limit=ResourceLimit(
                cpu="2",
                memory="1Gi",
                disk="10Gi"
            ),
            workspace_path="/workspace/test",
            runtime_type="python3.11",
            runtime_node="node-1",
            container_id="container-123",
            pod_name="pod-456",
            env_vars={"DEBUG": "true"},
            timeout=600,
            python_package_index_url="https://mirror.example/simple",
            requested_dependencies=["requests==2.31.0"],
            created_at=now,
            updated_at=now,
            completed_at=now,
            last_activity_at=now
        )

        dto = SessionDTO.from_entity(session)

        assert dto.status == "running"
        assert dto.runtime_node == "node-1"
        assert dto.container_id == "container-123"
        assert dto.pod_name == "pod-456"
        assert dto.env_vars == {"DEBUG": "true"}
        assert dto.timeout == 600
        assert dto.python_package_index_url == "https://mirror.example/simple"
        assert dto.completed_at == now

    def test_is_dataclass(self):
        """测试是数据类"""
        from dataclasses import is_dataclass

        assert is_dataclass(SessionDTO)
