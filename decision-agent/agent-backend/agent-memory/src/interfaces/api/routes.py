from fastapi import APIRouter, Request, Response
from typing import Dict, Any
from .schemas import (
    BuildMemoryRequest,
    RetrievalMemoryRequest,
    UpdateMemoryRequest,
    GetAllMemoriesRequest,
    MemoryResponse,
    MemoryListResponse,
    MemoryHistoryResponse,
    SearchMemoryResponse,
)
from .exceptions import (
    MemoryNotFoundError,
    MemoryOperationError,
)
from src.application import (
    BuildMemoryUseCase,
    RetrievalMemoryUseCase,
    ManageMemoryUseCase,
)
from src.utils.logger import logger
from fastapi import status, Depends

internal_router = APIRouter(prefix="/api/agent-memory/internal/v1", tags=["memory"])
external_router = APIRouter(
    prefix="/api/agent-memory/v1", tags=["memory"]
)  # 预留外部接口扩展

# 初始化用例
build_memory_use_case = BuildMemoryUseCase()
retrieval_memory_use_case = RetrievalMemoryUseCase()
manage_memory_use_case = ManageMemoryUseCase()


def _convert_to_memory_response(memory_data: Dict[str, Any]) -> Dict[str, Any]:
    """将 mem0 返回的数据转换为 MemoryResponse 格式"""
    return {
        "id": memory_data["id"],
        "memory": memory_data["memory"],
        "hash": memory_data.get("hash"),
        "metadata": memory_data.get("metadata"),
        "score": memory_data.get("score"),
        "rerank_score": memory_data.get("rerank_score", 0),
        "created_at": memory_data["created_at"],
        "updated_at": memory_data.get("updated_at"),
        "user_id": memory_data.get("user_id"),
        "agent_id": memory_data.get("agent_id"),
        "run_id": memory_data.get("run_id"),
    }


@internal_router.post("/memory", response_model=MemoryResponse)
async def build_memory(request: Request, build_request: BuildMemoryRequest):
    """构建记忆"""

    # Extract headers for context
    context = {
        "user_id": request.headers.get("x-account-id"),
        "visitor_type": request.headers.get("x-account-type"),
    }

    logger.debugf(
        "Building memory, request params=%s, context=%s",
        build_request.model_dump(),
        context,
    )
    await build_memory_use_case.execute(
        messages=[msg.model_dump() for msg in build_request.messages],
        user_id=build_request.user_id,
        agent_id=build_request.agent_id,
        run_id=build_request.run_id,
        metadata=build_request.metadata,
        infer=build_request.infer,
        memory_type=build_request.memory_type,
        prompt=build_request.prompt,
        context=context,
    )

    logger.info("Memory built successfully")
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@internal_router.post("/search", response_model=SearchMemoryResponse)
async def retrieval_memory(request: Request, retrieval_request: RetrievalMemoryRequest):
    """召回记忆"""

    # Extract headers for context
    context = {
        "user_id": request.headers.get("x-account-id"),
        "visitor_type": request.headers.get("x-account-type"),
    }

    logger.infof(
        "Retrievaling memories, query:%s, user_id:%s, context:%s",
        retrieval_request.query,
        retrieval_request.user_id,
        context,
    )

    results = await retrieval_memory_use_case.execute(
        query=retrieval_request.query,
        user_id=retrieval_request.user_id,
        agent_id=retrieval_request.agent_id,
        run_id=retrieval_request.run_id,
        limit=retrieval_request.limit,
        filters=retrieval_request.filters,
        threshold=retrieval_request.threshold,
        rerank_threshold=retrieval_request.rerank_threshold,
        context=context,
    )

    # 转换结果格式
    memory_list = [
        _convert_to_memory_response(item) for item in results.get("results", results)
    ]
    logger.infof(
        "Memories retrievaled successfully, count=%s, user_id=%s",
        len(memory_list),
        retrieval_request.user_id,
    )

    return {"result": memory_list}


@external_router.get("/memory/{memory_id}", response_model=MemoryResponse)
async def get_memory(request: Request, memory_id: str):
    """获取记忆"""
    logger.info("Getting memory", memory_id=memory_id)

    result = await manage_memory_use_case.get_memory(memory_id)
    if not result:
        raise MemoryNotFoundError(memory_id)

    logger.info("Memory retrieved successfully", memory_id=memory_id)
    return result


@external_router.get("/memory", response_model=MemoryListResponse)
async def get_all_memories(request: Request, params: GetAllMemoriesRequest = Depends()):
    """获取所有记忆"""
    logger.infof("Getting all memories, request: %s", params.model_dump())

    results = await manage_memory_use_case.get_all_memories(
        user_id=params.user_id,
        agent_id=params.agent_id,
        run_id=params.run_id,
        limit=params.limit,
    )
    logger.info(
        "All memories retrieved successfully",
        count=len(results.get("results", results)),
        extra=results,
    )
    return results


@external_router.put("/memory/{memory_id}")
async def update_memory(
    request: Request, memory_id: str, update_request: UpdateMemoryRequest
):
    """更新记忆"""
    logger.info("Updating memory", memory_id=memory_id, data=update_request.data)

    result = await manage_memory_use_case.get_memory(memory_id)
    if not result:
        raise MemoryNotFoundError(memory_id)

    result = await manage_memory_use_case.update_memory(memory_id, update_request.data)
    if not result:
        raise MemoryOperationError(memory_id)

    logger.infof("Memory updated successfully, memory id:%s", memory_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@external_router.delete("/memory/{memory_id}")
async def delete_memory(request: Request, memory_id: str):
    """删除记忆"""
    logger.info("Deleting memory", memory_id=memory_id)

    result = await manage_memory_use_case.get_memory(memory_id)
    if not result:
        raise MemoryNotFoundError(memory_id)

    result = await manage_memory_use_case.delete_memory(memory_id)
    if not result:
        raise MemoryOperationError(memory_id)

    logger.info("Memory deleted successfully", memory_id=memory_id)
    return Response(status_code=status.HTTP_204_NO_CONTENT)


@external_router.get(
    "/memory/{memory_id}/history", response_model=MemoryHistoryResponse
)
async def get_memory_history(request: Request, memory_id: str):
    """获取记忆历史"""
    logger.info("Getting memory history", memory_id=memory_id)

    history = await manage_memory_use_case.get_memory_history(memory_id)

    logger.info(
        "Memory history retrieved successfully",
        memory_id=memory_id,
        history_count=len(history),
    )
    return {"history": history}
