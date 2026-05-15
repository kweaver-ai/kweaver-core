"""
内部 API 路由

定义由 Executor 调用的内部 API 端点。
这些端点仅在容器网络内可访问。
"""

import logging
from datetime import datetime
from typing import Dict

from fastapi import APIRouter, Depends, HTTPException, status
from fastapi.responses import JSONResponse

from src.domain.repositories.execution_repository import IExecutionRepository
from src.domain.value_objects.execution_status import ExecutionStatus
from src.domain.value_objects.artifact import Artifact, ArtifactType
from src.interfaces.rest.schemas.internal import (
    ContainerReadyRequest,
    ExecutionResultReport,
    InternalAPIResponse,
)
from src.infrastructure.dependencies import (
    USE_SQL_REPOSITORIES,
    get_execution_repository as get_sql_execution_repository,
    get_session_repository as get_sql_session_repository,
)

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/internal", tags=["internal"])


# 根据模式选择依赖注入函数
# SQL 模式：使用 get_execution_repository（带 Depends() 注入数据库会话）
# Mock 模式：使用从 app.state 获取仓储的函数
if USE_SQL_REPOSITORIES:
    _get_execution_repository = get_sql_execution_repository
    _get_session_repository = get_sql_session_repository
else:
    from src.infrastructure.dependencies import (
        get_execution_repository as get_mock_execution_repository,
    )
    from src.infrastructure.dependencies import (
        get_session_repository as get_mock_session_repository,
    )

    _get_execution_repository = get_mock_execution_repository
    _get_session_repository = get_mock_session_repository


@router.post("/containers/ready")
async def handle_container_ready(
    request: ContainerReadyRequest,
    session_repo=Depends(_get_session_repository),
):
    """
    处理容器就绪事件

    由 Executor 在启动完成后调用，通知控制平面容器已就绪。
    更新对应的会话状态为 RUNNING。
    """
    logger.info(f"Container ready event received: container_id={request.container_id}")

    # 查找对应的会话
    session = await session_repo.find_by_container_id(request.container_id)
    if session:
        # 更新会话状态为 RUNNING
        from src.domain.value_objects.execution_status import SessionStatus

        session.status = SessionStatus.RUNNING

        await session_repo.save(session)
        logger.info(f"Session {session.id} status updated to RUNNING")
        return InternalAPIResponse(message="Container ready acknowledged, session updated")
    else:
        logger.warning(f"No session found for container_id={request.container_id}")
        return InternalAPIResponse(message="Container ready acknowledged (no session found)")


@router.post("/containers/exited")
async def handle_container_exited():
    """
    处理容器退出事件

    由 Executor 在关闭前调用，通知控制平面容器即将退出。
    """
    logger.info("Container exited event received")
    # Currently just acknowledge - future: update container status in database
    return InternalAPIResponse(message="Container exited acknowledged")


@router.post("/executions/{execution_id}/heartbeat")
async def handle_execution_heartbeat(execution_id: str):
    """
    处理执行心跳

    由 Executor 在执行过程中定期调用，保持执行活跃状态。
    """
    logger.debug(f"Heartbeat received for execution {execution_id}")
    # Currently just acknowledge - future: update last_heartbeat timestamp
    return InternalAPIResponse(message="Heartbeat acknowledged")


@router.post(
    "/executions/{execution_id}/result",
    response_model=InternalAPIResponse,
    status_code=status.HTTP_200_OK,
)
async def report_execution_result(
    execution_id: str,
    report: ExecutionResultReport,
    execution_repo: IExecutionRepository = Depends(_get_execution_repository),
):
    """
    上报执行结果

    由 Executor 在执行完成后调用，上报执行结果到控制平面。

    ## 状态映射
    - API: `"success"` → Domain: `ExecutionStatus.COMPLETED`
    - API: `"failed"` → Domain: `ExecutionStatus.FAILED`
    - API: `"timeout"` → Domain: `ExecutionStatus.TIMEOUT`
    - API: `"crashed"` → Domain: `ExecutionStatus.CRASHED`

    ## 幂等性
    - 如果执行记录已经是终态，返回 200（重复上报）
    - 如果是首次上报，更新后返回 201
    """
    # 1. 查找执行记录
    execution = await execution_repo.find_by_id(execution_id)
    if not execution:
        raise HTTPException(
            status_code=status.HTTP_404_NOT_FOUND,
            detail=f"Execution not found: {execution_id}",
        )

    # 2. 检查是否已经是终态（幂等性）
    if execution.is_terminal():
        logger.info(f"Execution {execution_id} already in terminal state: {execution.state.status}")
        return InternalAPIResponse(message="Result already recorded")

    # 3. 映射 API 状态到域状态
    status_map: Dict[str, ExecutionStatus] = {
        "success": ExecutionStatus.COMPLETED,
        "failed": ExecutionStatus.FAILED,
        "timeout": ExecutionStatus.TIMEOUT,
        "crashed": ExecutionStatus.CRASHED,
    }

    domain_status = status_map.get(report.status)
    if not domain_status:
        raise HTTPException(
            status_code=status.HTTP_400_BAD_REQUEST,
            detail=f"Invalid status: {report.status}",
        )

    # 4. 自动转换 PENDING → RUNNING（如果需要）
    # 根据领域规则，必须是 PENDING → RUNNING → COMPLETED/FAILED/TIMEOUT/CRASHED
    # 但 executor 报告结果时可能已经完成了，所以自动处理这个转换
    if execution.state.status == ExecutionStatus.PENDING:
        execution.mark_running()

    # 5. 根据状态更新执行实体
    try:
        if domain_status == ExecutionStatus.COMPLETED:
            # 转换 artifacts 字符串列表为 Artifact 对象
            now = datetime.now()
            artifact_objects = [
                Artifact(
                    path=path, size=0, mime_type="", type=ArtifactType.ARTIFACT, created_at=now
                )
                for path in report.artifacts
            ]

            # 转换 metrics
            metrics_dict = None
            if report.metrics:
                metrics_dict = {
                    "duration_ms": report.metrics.duration_ms,
                    "cpu_time_ms": report.metrics.cpu_time_ms,
                    "peak_memory_mb": report.metrics.peak_memory_mb,
                    "io_read_bytes": report.metrics.io_read_bytes,
                    "io_write_bytes": report.metrics.io_write_bytes,
                }

            execution.mark_completed(
                stdout=report.stdout,
                stderr=report.stderr,
                exit_code=report.exit_code,
                execution_time=report.execution_time,
                artifacts=artifact_objects,
                return_value=report.return_value,
                metrics=metrics_dict,
            )

        elif domain_status == ExecutionStatus.FAILED:
            # 使用 stderr 作为错误消息，同时保存 stdout 和 stderr
            error_message = report.stderr if report.stderr else "Execution failed"
            execution.mark_failed(
                error_message=error_message,
                exit_code=report.exit_code,
                stdout=report.stdout,
                stderr=report.stderr,
            )

        elif domain_status == ExecutionStatus.TIMEOUT:
            execution.mark_timeout()

        elif domain_status == ExecutionStatus.CRASHED:
            execution.mark_crashed()

        # 6. 保存到仓储
        await execution_repo.save(execution)

        # 6.5. 提交事务，确保其他请求可以立即看到更新后的执行状态
        await execution_repo.commit()

        logger.info(
            f"Execution result recorded: {execution_id}, status={domain_status}, "
            f"exit_code={report.exit_code}"
        )

        # 6. 返回 201 表示首次创建
        return JSONResponse(
            status_code=status.HTTP_201_CREATED,
            content={"message": "Result recorded successfully"},
        )

    except ValueError as e:
        # 状态转换错误（例如从未完成状态直接尝试标记为完成）
        raise HTTPException(
            status_code=status.HTTP_409_CONFLICT,
            detail=f"State conflict: {str(e)}",
        )
    except Exception as e:
        logger.error(f"Failed to record execution result: {e}")
        raise HTTPException(
            status_code=status.HTTP_500_INTERNAL_SERVER_ERROR,
            detail=f"Internal error: {str(e)}",
        )
