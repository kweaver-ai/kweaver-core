from typing import Any, Dict, Optional, TYPE_CHECKING

from app.common.config import Config
from app.utils.observability.trace_wrapper import internal_span
from opentelemetry.trace import Span

from ..trace import span_set_attrs
from app.domain.enum.common.user_account_header_key import (
    set_user_account_id,
    set_user_account_type,
)


if TYPE_CHECKING:
    from ..agent_core_v2 import AgentCoreV2


def _configure_local_dev_llm(llm: Dict[str, Any], llm_config: Dict[str, Any]) -> None:
    """配置本地开发环境的LLM设置

    使用 llm_config.name 作为 model_list 的查找 key，
    不修改 llm["llm_config"]["name"]，只更新 model_name/api/api_key，
    避免下次调用时 key 变化导致匹配失败。
    """
    # 用当前 name 查找 model_list（不会被 mutation 影响）
    original_name = llm["llm_config"]["name"]
    model_config = Config.get_local_dev_model_config(original_name)

    if model_config:
        # 使用配置映射表中的模型配置
        llm["llm_config"]["api"] = model_config["api"]
        llm["llm_config"]["api_key"] = model_config["api_key"]
        llm["llm_config"]["model_name"] = model_config["model"]
    else:
        # 使用默认配置
        llm["llm_config"]["api"] = Config.outer_llm.api
        llm["llm_config"]["api_key"] = Config.outer_llm.api_key
        llm["llm_config"]["model_name"] = Config.outer_llm.model

    from dolphin.core.config.global_config import TypeAPI

    llm["llm_config"]["type_api"] = TypeAPI.OPENAI.value
    # 注意：不再修改 llm["llm_config"]["name"]，保持原始值（如 "deepseek-v3"），
    # 以确保下次调用时仍能正确匹配 model_list。


@internal_span()
async def build_llm_config(
    ac: "AgentCoreV2",
    user_id: str = "",
    visitor_type: str = "",
    span: Optional[Span] = None,
) -> Dict[str, Any]:
    config = ac.agent_config

    span_set_attrs(
        span=span,
        agent_run_id=config.agent_run_id or "",
        agent_id=config.agent_id or "",
        user_id=user_id,
    )

    llm_config = {"default": "", "llms": {}}

    for llm in config.llms or []:
        if llm.get("is_default"):
            llm_config["default"] = llm["llm_config"]["name"]

        llm["llm_config"]["model_name"] = llm["llm_config"]["name"]

        llm["llm_config"]["api"] = (
            f"http://{Config.services.mf_model_api.host}:{Config.services.mf_model_api.port}/api/private/mf-model-api/v1/chat/completions"
        )

        if Config.local_dev.is_use_outer_llm:
            _configure_local_dev_llm(llm, llm_config)

        llm_headers = {}
        set_user_account_id(llm_headers, user_id)
        set_user_account_type(llm_headers, visitor_type)
        llm["llm_config"]["headers"] = llm_headers

        llm_config["llms"][llm["llm_config"]["name"]] = llm["llm_config"]

    

    return llm_config


def get_llm_config_from_cache(ac: "AgentCoreV2", llm_id: str) -> Dict[str, Any]:
    return ac.cache_handler.get_llm_config(llm_id)
