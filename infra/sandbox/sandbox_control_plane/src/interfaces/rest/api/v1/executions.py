"""
执行 REST API 路由

定义执行相关的 HTTP 端点。
"""
import asyncio
import fastapi
from fastapi import APIRouter, Depends, HTTPException, Query, status
from typing import Optional

from src.application.services.session_service import SessionService
from src.application.commands.execute_code import ExecuteCodeCommand
from src.application.queries.get_execution import GetExecutionQuery
from src.application.dtos.execution_dto import ExecutionDTO
from src.domain.value_objects.execution_status import ExecutionStatus
from src.interfaces.rest.schemas.request import ExecuteCodeRequest
from src.interfaces.rest.schemas.response import (
    ExecutionResponse,
    ExecuteCodeResponse,
    ErrorResponse
)
from src.infrastructure.dependencies import (
    USE_SQL_REPOSITORIES,
    get_session_service_db,
    get_session_service as get_mock_session_service,
)
from src.infrastructure.persistence.database import db_manager

router = APIRouter(prefix="/executions", tags=["executions"])


# 根据模式选择依赖注入函数
# SQL 模式：使用 get_session_service_db（带 Depends() 注入仓储）
# Mock 模式：使用 get_mock_session_service（从 app.state 获取）
_get_session_service = get_session_service_db if USE_SQL_REPOSITORIES else get_mock_session_service


@router.post("/sessions/{session_id}/execute", response_model=ExecuteCodeResponse, status_code=status.HTTP_201_CREATED)
async def submit_execution(
    session_id: str,
    request: ExecuteCodeRequest,
    service: SessionService = Depends(_get_session_service)
):
    """
    提交代码执行

    - **code**: 要执行的代码
    - **language**: 编程语言 (python, javascript, shell)
    - **timeout**: 超时时间（秒），默认 30
    - **event**: 事件数据
    """
    command = ExecuteCodeCommand(
        session_id=session_id,
        code=request.code,
        language=request.language,
        timeout=request.timeout,
        event_data=request.event
    )

    execution_dto = await service.execute_code(command)

    return ExecuteCodeResponse(
        execution_id=execution_dto.id,
        session_id=execution_dto.session_id,
        status=execution_dto.status,
        created_at=execution_dto.created_at
    )


@router.post("/sessions/{session_id}/execute-sync", response_model=ExecutionResponse)
async def execute_code_sync(
    session_id: str,
    request: ExecuteCodeRequest,
    poll_interval: float = Query(default=0.5, ge=0.1, le=10.0, description="Polling interval in seconds"),
    sync_timeout: int = Query(default=300, ge=10, le=3600, description="Maximum wait time in seconds"),
    service: SessionService = Depends(_get_session_service)
):
    """
    Synchronous code execution endpoint

    Internally calls async execution and polls for result until:
    - Execution reaches terminal state (COMPLETED, FAILED, TIMEOUT, CRASHED)
    - sync_timeout is reached

    - **poll_interval**: Polling interval in seconds (default: 0.5, range: 0.1-10.0)
    - **sync_timeout**: Maximum wait time in seconds (default: 300, range: 10-3600)
    - **code**: Code to execute
    - **language**: Programming language (python, javascript, shell)
    - **timeout**: Execution timeout in seconds
    - **event**: Event data
    """
    # 1. Submit execution
    command = ExecuteCodeCommand(
        session_id=session_id,
        code=request.code,
        language=request.language,
        timeout=request.timeout,
        event_data=request.event
    )
    execution_dto = await service.execute_code(command)
    execution_id = execution_dto.id

    # 2. Poll for result with timeout
    loop = asyncio.get_event_loop()
    start_time = loop.time()
    # Terminal states for early exit
    terminal_states = {
        ExecutionStatus.COMPLETED.value,
        ExecutionStatus.FAILED.value,
        ExecutionStatus.TIMEOUT.value,
        ExecutionStatus.CRASHED.value
    }

    import logging
    logger = logging.getLogger(__name__)
    logger.info(f"Starting sync polling loop: execution_id={execution_id}, USE_SQL_REPOSITORIES={USE_SQL_REPOSITORIES}")

    while True:
        # Check timeout
        elapsed = loop.time() - start_time
        if elapsed >= sync_timeout:
            raise HTTPException(
                status_code=status.HTTP_408_REQUEST_TIMEOUT,
                detail=f"Synchronous execution timeout after {sync_timeout}s"
            )

        # Get current status - use a fresh database session for each poll
        # to avoid REPEATABLE-READ transaction isolation issues
        if USE_SQL_REPOSITORIES:
            execution_dto = await _get_execution_with_fresh_session(execution_id)
        else:
            # Mock mode - use the service directly
            query = GetExecutionQuery(execution_id=execution_id)
            execution_dto = await service.get_execution(query)

        # Check if terminal state
        if execution_dto.status in terminal_states:
            return _map_dto_to_response(execution_dto)

        # Wait before next poll
        await asyncio.sleep(poll_interval)


async def _get_execution_with_fresh_session(execution_id: str) -> ExecutionDTO:
    """
    Get execution using a fresh database session.

    This is required for the sync execution polling loop to work around
    MySQL's REPEATABLE-READ transaction isolation. Each poll needs to
    see the latest committed data from the executor callback.
    """
    from src.infrastructure.persistence.repositories.sql_execution_repository import SqlExecutionRepository

    async with db_manager.get_session() as session:
        repo = SqlExecutionRepository(session)
        execution = await repo.find_by_id(execution_id)
        if not execution:
            from src.shared.errors.domain import NotFoundError
            raise NotFoundError(f"Execution not found: {execution_id}")
        return ExecutionDTO.from_entity(execution)


@router.get("/{execution_id}/status", response_model=ExecutionResponse)
async def get_execution_status(
    execution_id: str,
    service: SessionService = Depends(_get_session_service)
):
    """获取执行状态"""
    query = GetExecutionQuery(execution_id=execution_id)
    execution_dto = await service.get_execution(query)
    return _map_dto_to_response(execution_dto)


@router.get("/{execution_id}/result", response_model=ExecutionResponse)
async def get_execution_result(
    execution_id: str,
    service: SessionService = Depends(_get_session_service)
):
    """获取执行结果"""
    query = GetExecutionQuery(execution_id=execution_id)
    execution_dto = await service.get_execution(query)
    return _map_dto_to_response(execution_dto)


@router.get("/sessions/{session_id}/executions")
async def list_executions(
    session_id: str,
    limit: int = 50,
    offset: int = 0,
    service: SessionService = Depends(_get_session_service)
):
    """列出会话的所有执行"""
    executions = await service.list_executions(session_id=session_id, limit=limit)

    return {
        "items": executions,
        "total": len(executions),
        "limit": limit,
        "offset": offset
    }


def _map_dto_to_response(dto: ExecutionDTO) -> ExecutionResponse:
    """将 ExecutionDTO 映射为 ExecutionResponse"""
    return ExecutionResponse(
        id=dto.id,
        session_id=dto.session_id,
        status=dto.status,
        code=dto.code,
        language=dto.language,
        timeout=dto.timeout,
        stdout=dto.stdout,
        stderr=dto.stderr,
        exit_code=dto.exit_code,
        return_value=dto.return_value,
        metrics=dto.metrics,
        created_at=dto.created_at,
        started_at=dto.started_at,
        completed_at=dto.completed_at
    )
