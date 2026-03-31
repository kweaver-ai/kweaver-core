from fastapi import Request, Depends
from sse_starlette import EventSourceResponse

from .dependencies import get_account_id, get_account_type, get_biz_domain_id


from .common_v2 import router_v2

from .rdto.v2.req.run_agent import V2RunAgentReq
from .rdto.v2.res.run_agent import V2RunAgentResponse

from .run_agent_v2 import run_agent


@router_v2.post(
    "/debug",
    summary="运行agentV2(调试模式)",
    response_model=V2RunAgentResponse,
)
async def run_agent_debug(
    request: Request,
    req: V2RunAgentReq,
    account_id: str = Depends(get_account_id),
    account_type: str = Depends(get_account_type),
    biz_domain_id: str = Depends(get_biz_domain_id),
) -> EventSourceResponse:
    """
    运行agentV2(调试模式)
    """

    return await run_agent(
        request,
        req,
        is_debug_run=True,
        account_id=account_id,
        account_type=account_type,
        biz_domain_id=biz_domain_id,
    )
