# -*- coding:utf-8 -*-
"""Agent 恢复执行接口"""

from fastapi import Request
from fastapi.responses import JSONResponse
from sse_starlette import EventSourceResponse

from .common_v2 import router_v2
from .rdto.v2.req.resume_agent import ResumeAgentRequest
from app.logic.agent_core_logic_v2.agent_instance_manager import agent_instance_manager
from app.logic.agent_core_logic_v2.resume_handler import create_resume_generator
from app.common.stand_log import StandLogger


@router_v2.post("/resume", summary="恢复Agent执行")
async def resume_agent(request: Request, req: ResumeAgentRequest):
    """恢复被中断的 Agent 执行

    Args:
        request: FastAPI Request 对象
        req: 恢复请求参数

    Returns:
        SSE 流式响应或错误响应
    """

    agent_run_id = req.agent_run_id
    resume_info = req.resume_info

    StandLogger.info(f"Resume agent: {agent_run_id}, action: {resume_info.action}")

    # 1. 获取 Agent 实例
    instance_data = agent_instance_manager.get(agent_run_id)
    if instance_data is None:
        StandLogger.error(f"agent_run_id {agent_run_id} not found")
        return JSONResponse(
            status_code=404,
            content={"error": f"agent_run_id {agent_run_id} not found"},
        )

    agent, agent_core = instance_data

    # 2. 创建恢复生成器（独立方法）
    generator = create_resume_generator(agent, agent_core, agent_run_id, resume_info)

    return EventSourceResponse(generator, ping=3600)
