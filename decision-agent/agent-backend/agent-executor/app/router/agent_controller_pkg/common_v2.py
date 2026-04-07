from fastapi import APIRouter

from app.common.config import Config


router_v2 = APIRouter(
    prefix=Config.app.host_prefix_v2 + "/agent", tags=["agent-executor"]
)
