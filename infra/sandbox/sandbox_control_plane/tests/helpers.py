"""
共享测试工具

提供通用的测试 fixture 和 mock 工具，减少测试代码重复。
"""
from unittest.mock import Mock, AsyncMock
from datetime import datetime

from src.domain.entities.session import Session
from src.domain.entities.template import Template
from src.domain.entities.execution import Execution
from src.domain.value_objects.resource_limit import ResourceLimit
from src.domain.value_objects.execution_status import SessionStatus, ExecutionStatus
from src.domain.services.scheduler import RuntimeNode


def create_mock_template(
    template_id: str = "python-test",
    name: str = "Python Test",
    image: str = "python:3.11",
    default_timeout_sec: int = 300
) -> Template:
    """
    创建测试用模板实体

    Args:
        template_id: 模板 ID
        name: 模板名称
        image: 镜像名称
        default_timeout_sec: 默认超时时间（秒）

    Returns:
        Template 实体
    """
    return Template(
        id=template_id,
        name=name,
        image=image,
        base_image=image,
        pre_installed_packages=[],
        default_resources=ResourceLimit(
            cpu="1",
            memory="512Mi",
            disk="1Gi",
            max_processes=128,
        ),
        default_timeout_sec=default_timeout_sec,
        security_context={},
    )


def create_mock_session(
    session_id: str = "test-session-123",
    template_id: str = "python-test",
    status: str = "running"
) -> Session:
    """
    创建测试用会话实体

    Args:
        session_id: 会话 ID
        template_id: 模板 ID
        status: 会话状态

    Returns:
        Session 实体
    """
    return Session(
        id=session_id,
        template_id=template_id,
        status=SessionStatus(status),
        resource_limit=ResourceLimit(
            cpu="1",
            memory="512Mi",
            disk="1Gi",
            max_processes=128,
        ),
        runtime_type="docker",
        runtime_node="node-1",
        container_id=f"container-{session_id}",
        workspace_path=f"sessions/{session_id}",
        timeout=300,
    )


def create_mock_execution(
    execution_id: str = "test-execution-123",
    session_id: str = "test-session-123",
    status: str = "pending"
) -> Execution:
    """
    创建测试用执行实体

    Args:
        execution_id: 执行 ID
        session_id: 会话 ID
        status: 执行状态

    Returns:
        Execution 实体
    """
    return Execution(
        id=execution_id,
        session_id=session_id,
        status=ExecutionStatus(status),
        code="print('Hello, World!')",
        language="python",
        timeout=30,
    )


def create_mock_runtime_node(
    node_id: str = "node-1",
    node_type: str = "docker"
) -> RuntimeNode:
    """
    创建测试用运行时节点

    Args:
        node_id: 节点 ID
        node_type: 节点类型

    Returns:
        RuntimeNode 对象
    """
    return RuntimeNode(
        id=node_id,
        type=node_type,
        url=f"{node_type}://{node_id}",
        status="healthy",
        cpu_usage=0.5,
        mem_usage=0.6,
        session_count=5,
        max_sessions=100,
        cached_templates=["python-test"],
    )


def create_mock_repository(
    find_by_id_return=None,
    save_return=None,
    find_all_return=None
) -> Mock:
    """
    创建模拟仓储

    Args:
        find_by_id_return: find_by_id 方法的返回值
        save_return: save 方法的返回值
        find_all_return: find_all 方法的返回值

    Returns:
        Mock 对象
    """
    repo = Mock()
    repo.save = AsyncMock(return_value=save_return)
    repo.find_by_id = AsyncMock(return_value=find_by_id_return)
    repo.find_all = AsyncMock(return_value=find_all_return or [])
    repo.delete = AsyncMock()
    return repo


def create_mock_scheduler(
    schedule_return=None,
    create_container_return="container-123",
    destroy_container_return=None
) -> Mock:
    """
    创建模拟调度器

    Args:
        schedule_return: schedule 方法的返回值
        create_container_return: create_container_for_session 方法的返回值
        destroy_container_return: destroy_container 方法的返回值

    Returns:
        Mock 对象
    """
    scheduler = Mock()
    scheduler.schedule = AsyncMock(return_value=schedule_return)
    scheduler.create_container_for_session = AsyncMock(
        return_value=create_container_return
    )
    scheduler.destroy_container = AsyncMock(return_value=destroy_container_return)
    return scheduler
