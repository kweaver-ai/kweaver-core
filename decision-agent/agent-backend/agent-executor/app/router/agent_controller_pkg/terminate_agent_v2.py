# -*- coding:utf-8 -*-
"""Agent 终止执行接口"""

from fastapi import Request
from fastapi.responses import Response, JSONResponse

from .common_v2 import router_v2
from .rdto.v2.req.terminate_agent import TerminateAgentRequest
from app.logic.agent_core_logic_v2.agent_instance_manager import agent_instance_manager
from app.common.stand_log import StandLogger


@router_v2.post("/terminate", summary="终止Agent执行", status_code=204)
async def terminate_agent(request: Request, req: TerminateAgentRequest):
    """终止正在运行的 Agent

    Args:
        request: FastAPI Request 对象
        req: 终止请求参数

    Returns:
        204 No Content 或错误响应
    """

    agent_run_id = req.agent_run_id

    StandLogger.info(f"Terminate agent: {agent_run_id}")

    # 1. 获取 Agent 实例
    instance_data = agent_instance_manager.get(agent_run_id)
    if instance_data is None:
        StandLogger.error(f"agent_run_id {agent_run_id} not found")
        return JSONResponse(
            status_code=404,
            content={"error": f"agent_run_id {agent_run_id} not found"},
        )

    agent, agent_core = instance_data

    # 2. 调用 SDK terminate
    try:
        await agent.terminate()
        StandLogger.info(f"Agent {agent_run_id} terminated successfully")
    except Exception as e:
        StandLogger.error(f"Terminate agent error: {e}")
    finally:
        # 3. 清理实例
        agent_instance_manager.remove(agent_run_id)
        agent_core.cleanup()

    return Response(status_code=204)
