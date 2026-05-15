"""
会话 REST API 路由

定义会话相关的 HTTP 端点。
"""

from fastapi import APIRouter, Depends, HTTPException, status, Response
from typing import List, Optional

from src.application.commands.install_session_dependencies import (
    InstallSessionDependenciesCommand,
)
from src.application.services.session_service import SessionService
from src.shared.errors.domain import ConflictError
from src.application.commands.create_session import CreateSessionCommand
from src.application.commands.execute_code import ExecuteCodeCommand
from src.application.dtos.session_dto import SessionDTO
from src.infrastructure.executors.errors import (
    ExecutorConnectionError,
    ExecutorResponseError,
    ExecutorTimeoutError,
    ExecutorUnavailableError,
    ExecutorValidationError,
)
from src.interfaces.rest.schemas.request import (
    CreateSessionRequest,
    ExecuteCodeRequest,
    InstallSessionDependenciesRequest,
)
from src.interfaces.rest.schemas.response import (
    DependencyResponse,
    InstalledDependencyResponse,
    SessionResponse,
    SessionListResponse,
    ExecuteCodeResponse,
    ErrorResponse,
)
from src.infrastructure.dependencies import get_session_service_db

router = APIRouter(prefix="/sessions", tags=["sessions"])


@router.post("", response_model=SessionResponse, status_code=status.HTTP_201_CREATED)
async def create_session(
    request: CreateSessionRequest, service: SessionService = Depends(get_session_service_db)
):
    """
    创建会话

    - **template_id**: 模板 ID；未传时使用 DEFAULT_TEMPLATE_ID 配置
    - **timeout**: 超时时间（秒），默认 300，最大 3600
    - **cpu**: CPU 核心数，如 "1", "2"
    - **memory**: 内存限制，如 "512Mi", "1Gi"
    - **disk**: 磁盘限制，如 "1Gi", "10Gi"
    - **env_vars**: 环境变量字典
    - **dependencies**: 会话级依赖包列表（新增）
    - **install_timeout**: 依赖安装超时时间（秒），默认 300（新增）
    - **fail_on_dependency_error**: 依赖安装失败时是否终止会话创建（新增）
    - **allow_version_conflicts**: 是否允许版本冲突（新增）
    """
    from src.domain.value_objects.resource_limit import ResourceLimit

    try:
        resource_limit = ResourceLimit(cpu=request.cpu, memory=request.memory, disk=request.disk)

        dependencies_pip_specs = [dep.to_pip_spec() for dep in request.dependencies]

        command = CreateSessionCommand(
            id=request.id,
            template_id=request.template_id,
            timeout=request.timeout,
            resource_limit=resource_limit,
            env_vars=request.env_vars,
            python_package_index_url=request.python_package_index_url,
            dependencies=dependencies_pip_specs,
            install_timeout=request.install_timeout,
            fail_on_dependency_error=request.fail_on_dependency_error,
            allow_version_conflicts=request.allow_version_conflicts,
        )

        session_dto = await service.create_session(command)
        return _map_dto_to_response(session_dto)

    except ConflictError as e:
        raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=str(e))
    except Exception as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.get("", response_model=SessionListResponse)
async def list_sessions(
    status: Optional[str] = None,
    template_id: Optional[str] = None,
    limit: int = 50,
    offset: int = 0,
    service: SessionService = Depends(get_session_service_db),
):
    """
    列出会话

    支持按 status 和 template_id 筛选，以及分页。

    - **status**: 会话状态筛选（可选），如 "running", "terminated"
    - **template_id**: 模板 ID 筛选（可选）
    - **limit**: 返回数量限制（1-200，默认 50）
    - **offset**: 偏移量（用于分页，默认 0）
    """
    result = await service.list_sessions(
        status=status, template_id=template_id, limit=limit, offset=offset
    )

    return SessionListResponse(
        items=[_map_dto_to_response(item) for item in result["items"]],
        total=result["total"],
        limit=result["limit"],
        offset=result["offset"],
        has_more=result["has_more"],
    )


@router.get("/{session_id}", response_model=SessionResponse)
async def get_session(session_id: str, service: SessionService = Depends(get_session_service_db)):
    """获取会话详情"""
    from src.application.queries.get_session import GetSessionQuery

    query = GetSessionQuery(session_id=session_id)
    session_dto = await service.get_session(query)
    return _map_dto_to_response(session_dto)


@router.post(
    "/{session_id}/dependencies/install",
    response_model=SessionResponse,
)
async def install_session_dependencies(
    session_id: str,
    request: InstallSessionDependenciesRequest,
    service: SessionService = Depends(get_session_service_db),
):
    """增量安装 Python 依赖。"""
    try:
        command = InstallSessionDependenciesCommand(
            session_id=session_id,
            python_package_index_url=request.python_package_index_url,
            dependencies=[dep.to_pip_spec() for dep in request.dependencies],
            install_timeout=request.install_timeout,
        )
        session_dto = await service.install_session_dependencies(command)
        return _map_dto_to_response(session_dto)
    except ConflictError as e:
        raise HTTPException(status_code=status.HTTP_409_CONFLICT, detail=str(e))
    except (ExecutorUnavailableError, ExecutorConnectionError, ExecutorTimeoutError) as e:
        raise HTTPException(status_code=status.HTTP_503_SERVICE_UNAVAILABLE, detail=str(e))
    except (ExecutorValidationError, ExecutorResponseError) as e:
        raise HTTPException(status_code=status.HTTP_400_BAD_REQUEST, detail=str(e))


@router.post("/{session_id}/terminate", response_model=SessionResponse)
async def terminate_session(
    session_id: str, service: SessionService = Depends(get_session_service_db)
):
    """
    终止会话（软终止，保留记录）

    行为：
    1. 销毁容器
    2. 删除 S3 工作区文件
    3. 更新会话状态为 terminated
    4. 保留数据库记录
    """
    session_dto = await service.terminate_session(session_id)
    return _map_dto_to_response(session_dto)


@router.delete("/{session_id}", status_code=status.HTTP_204_NO_CONTENT)
async def delete_session(
    session_id: str, service: SessionService = Depends(get_session_service_db)
):
    """
    删除会话（硬删除，级联删除执行记录）

    行为：
    1. 执行清理（销毁容器 + 删除 S3 文件）
    2. 硬删除 session 记录
    3. 级联删除所有关联的 execution 记录
    """
    await service.delete_session(session_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)


def _map_dto_to_response(dto: SessionDTO) -> SessionResponse:
    """将 SessionDTO 映射为 SessionResponse"""
    return SessionResponse(
        id=dto.id,
        template_id=dto.template_id,
        status=dto.status,
        resource_limit=dto.resource_limit,
        workspace_path=dto.workspace_path,
        language_runtime=dto.language_runtime,
        runtime_type=dto.runtime_type,
        runtime_node=dto.runtime_node,
        container_id=dto.container_id,
        pod_name=dto.pod_name,
        env_vars=dto.env_vars,
        timeout=dto.timeout,
        python_package_index_url=dto.python_package_index_url,
        requested_dependencies=[DependencyResponse(**dep) for dep in dto.requested_dependencies],
        installed_dependencies=[
            InstalledDependencyResponse(**dep) for dep in dto.installed_dependencies
        ],
        dependency_install_status=dto.dependency_install_status,
        dependency_install_error=dto.dependency_install_error,
        dependency_install_started_at=dto.dependency_install_started_at,
        dependency_install_completed_at=dto.dependency_install_completed_at,
        created_at=dto.created_at,
        updated_at=dto.updated_at,
        completed_at=dto.completed_at,
        last_activity_at=dto.last_activity_at,
    )
