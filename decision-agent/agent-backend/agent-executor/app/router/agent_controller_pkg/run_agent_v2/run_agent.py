import time
from fastapi import Request, Depends
from sse_starlette import EventSourceResponse

from app.logic.agent_core_logic_v2.agent_core_v2 import AgentCoreV2
from app.logic.agent_core_logic_v2.agent_cache_manage_logic import AgentCacheManager

from ..common_v2 import router_v2
from ..dependencies import get_account_id, get_account_type, get_biz_domain_id
from ..rdto.v2.req.run_agent import V2RunAgentReq
from ..rdto.v2.res.run_agent import V2RunAgentResponse

from .prepare import prepare
from .handle_cache import handle_cache
from .safe_output_generator import create_safe_output_generator

# 全局AgentCacheManager实例
cache_manager = AgentCacheManager()


@router_v2.post(
    "/run",
    summary="运行agentV2(由agent-app内部调用)",
    response_model=V2RunAgentResponse,
)
async def run_agent(
    request: Request,
    req: V2RunAgentReq,
    is_debug_run: bool = False,
    account_id: str = Depends(get_account_id),
    account_type: str = Depends(get_account_type),
    biz_domain_id: str = Depends(get_biz_domain_id),
) -> EventSourceResponse:
    """
    运行agentV2
    """

    # 记录开始时间（用于计算TTFT）
    start_time = time.time()

    # 1. prepare
    agent_config, agent_input, headers = await prepare(
        request, req, account_id, account_type, biz_domain_id
    )

    # 2. 初始化AgentCoreV2
    agent_core_v2 = AgentCoreV2(agent_config)
    agent_core_v2.set_run_options(req.options)

    # 3. 处理缓存
    cache_id_vo = None
    if req.options.enable_dependency_cache:
        cache_id_vo = await handle_cache(
            agent_id=req.agent_id,
            agent_core_v2=agent_core_v2,
            is_debug_run=is_debug_run,
            headers=headers,
            account_id=account_id,
            account_type=account_type,
        )

    return EventSourceResponse(
        create_safe_output_generator(
            agent_core_v2=agent_core_v2,
            agent_config=agent_config,
            agent_input=agent_input,
            headers=headers,
            is_debug_run=is_debug_run,
            start_time=start_time,
            cache_id_vo=cache_id_vo,
            account_id=account_id,
            account_type=account_type,
        ),
        ping=3600,
    )
