import json

from fastapi import APIRouter, Header, Request
from sse_starlette import EventSourceResponse

from app.common.config import Config
from app.models.tool_requests import (
    ZhipuSearchRequest,
    OnlineSearchCiteRequest,
    SkillLoadRequest,
    SkillReadFileRequest,
    SkillExecuteScriptRequest,
)
from app.models.tool_responses import (
    OnlineSearchCiteResponse,
    ZhipuSearchResponse,
    SkillLoadResponse,
    SkillReadFileResponse,
    SkillExecuteScriptResponse,
)


router = APIRouter(prefix=Config.app.host_prefix + "/tools", tags=["internal-tools"])


@router.post(
    "/zhipu_search_tool", response_model=ZhipuSearchResponse, summary="智谱搜索"
)
async def zhipu_search(
    request: ZhipuSearchRequest,
    api_key: str = Header(..., description="智谱API密钥", alias="api_key"),
) -> ZhipuSearchResponse:
    """
    执行智谱搜索

    - **query**: 搜索查询词

    返回搜索结果内容
    """
    from app.logic.tool.zhipu_search_tool import zhipu_search_tool

    param = request.model_dump()
    res = await zhipu_search_tool(param, {"api_key": api_key}, None, None, None)

    return ZhipuSearchResponse(**res)


@router.post(
    "/online_search_cite_tool",
    response_model=OnlineSearchCiteResponse,
    summary="联网搜索添加引用工具",
)
async def online_search_cite_tool(
    request: OnlineSearchCiteRequest,
) -> OnlineSearchCiteResponse:
    """
    执行联网搜索并添加引用

    - **query**: 搜索查询词
    - **model_name**: 模型名称
    - **search_tool**: 搜索工具
    - **api_key**: 搜索工具API密钥
    - **user_id**: 用户id
    返回搜索结果内容
    """
    headers = {"x-account-id": request.user_id}
    if not request.stream:
        from app.logic.tool.online_search_cite_tool import online_search_cite_tool

        param = request.model_dump()
        res = await online_search_cite_tool(request=param, headers=headers)

        return OnlineSearchCiteResponse(**res)
    else:
        param = request.model_dump()

        async def generate_stream():
            from app.logic.tool.online_search_cite_tool import (
                get_answer,
                get_completion_stream,
                get_search_results,
            )

            search_results = await get_search_results(param, headers)
            final_references = []
            ref_list = search_results["choices"][0]["message"]["tool_calls"][1][
                "search_result"
            ]
            for index, ref in enumerate(ref_list):
                ref_item = {
                    "title": ref.get("title", "未知标题"),
                    "content": ref.get("content", ""),
                    "link": ref.get("link", ""),
                    "index": index,
                }
                final_references.append(ref_item)

            full_answer = ""
            current_response = OnlineSearchCiteResponse(
                answer=full_answer,
                references=final_references,
            )
            yield f"{json.dumps(current_response.model_dump(), ensure_ascii=False)}"

            answer, _ = await get_answer(param, headers, search_results)

            async for chunk in get_completion_stream(
                param, headers, answer, final_references
            ):
                full_answer += chunk
                current_response = OnlineSearchCiteResponse(
                    answer=full_answer,
                    references=final_references,
                )
                yield f"{json.dumps(current_response.model_dump(), ensure_ascii=False)}"

        return EventSourceResponse(generate_stream(), ping=3600)


@router.post(
    "/builtin_skill_load",
    response_model=SkillLoadResponse,
    summary="内置工具：加载 Skill 包",
)
async def builtin_skill_load(
    request: Request,
    body: SkillLoadRequest,
) -> SkillLoadResponse:
    """
    加载指定 Skill 包的说明文档（SKILL.md）及文件列表。

    - **skill_id**: Skill 标识符

    返回 SKILL.md 内容、可用脚本列表和参考文件列表。
    """
    from app.driven.dip.agent_operator_integration_service import agent_operator_integration_service
    from app.logic.agent_core_logic_v2.factory_skill_runtime import load_skill

    headers = dict(request.headers)
    result = await load_skill(agent_operator_integration_service, body.skill_id, request_headers=headers)
    return SkillLoadResponse(**result.get("answer", result))


@router.post(
    "/builtin_skill_read_file",
    response_model=SkillReadFileResponse,
    summary="内置工具：读取 Skill 包内文件",
)
async def builtin_skill_read_file(
    request: Request,
    body: SkillReadFileRequest,
) -> SkillReadFileResponse:
    """
    读取指定 Skill 包内的单个文本文件内容。

    - **skill_id**: Skill 标识符
    - **file_path**: 文件在 Skill 包内的相对路径（如 references/guide.md）

    返回文件的完整文本内容。
    """
    from app.driven.dip.agent_operator_integration_service import agent_operator_integration_service
    from app.logic.agent_core_logic_v2.factory_skill_runtime import read_skill_file

    headers = dict(request.headers)
    result = await read_skill_file(
        agent_operator_integration_service, body.skill_id, body.file_path, request_headers=headers
    )
    return SkillReadFileResponse(**result.get("answer", result))


@router.post(
    "/builtin_skill_execute_script",
    response_model=SkillExecuteScriptResponse,
    summary="内置工具：执行 Skill 包脚本",
)
async def builtin_skill_execute_script(
    request: Request,
    body: SkillExecuteScriptRequest,
) -> SkillExecuteScriptResponse:
    """
    在执行工厂沙箱中执行指定 Skill 包的脚本。

    - **skill_id**: Skill 标识符
    - **entry_shell**: 来自 SKILL.md 的入口 Shell 命令（如 python scripts/analyze.py）

    返回执行结果，包含 stdout、stderr、exit_code 和 duration_ms。
    """
    from app.driven.dip.agent_operator_integration_service import agent_operator_integration_service
    from app.logic.agent_core_logic_v2.factory_skill_runtime import execute_skill_script

    headers = dict(request.headers)
    result = await execute_skill_script(
        agent_operator_integration_service, body.skill_id, body.entry_shell, request_headers=headers
    )
    return SkillExecuteScriptResponse(**result.get("answer", result))
